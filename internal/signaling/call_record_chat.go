package signaling

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"
	"github.com/openimsdk/protocol/rtc"
	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/log"
)

// rtcCallRecordCustomType 服务端 Custom(110) 通话记录消息的 customType 标识。
const rtcCallRecordCustomType = "rtcCallRecord"

// serverCallRecordData 与服务端 sendCallRecordChatMsg 发送的 CustomElem.data 字段对齐。
type serverCallRecordData struct {
	CustomType        string   `json:"customType"`
	MediaType         string   `json:"mediaType"`  // "audio" | "video"
	Status            string   `json:"status"`     // answered / cancelled / rejected / not_connected
	Duration          int64    `json:"duration"`   // 秒（服务端计算）
	InviterUserID     string   `json:"inviterUserID"`
	InviteeUserIDList []string `json:"inviteeUserIDList"`
	RoomID            string   `json:"roomID"`
}

// serverStatusToSDK 将服务端字符串状态映射为 SDK int32 常量及对应的 action 标识。
func serverStatusToSDK(s string) (status int32, action string) {
	switch s {
	case "answered":
		return constant.SignalCallStatusAnswered, constant.SignalCallActionHungUp
	case "cancelled":
		return constant.SignalCallStatusNotConnected, constant.SignalCallActionCancel
	case "rejected":
		return constant.SignalCallStatusNotConnected, constant.SignalCallActionReject
	default: // "not_connected"
		return constant.SignalCallStatusNotConnected, constant.SignalCallActionTimeout
	}
}

// resolveCallRecordDirectionByID 根据 inviterUserID 确定本端通话方向。
// 与 resolveCallRecordDirection 功能相同，但仅依赖 inviterUserID，无需完整 InvitationInfo。
func (s *Signaling) resolveCallRecordDirectionByID(inviterUserID string, status int32) int32 {
	if inviterUserID == s.loginUserID {
		return constant.SignalCallDirectionOutgoing
	}
	if status == constant.SignalCallStatusAnswered {
		return constant.SignalCallDirectionIncoming
	}
	return constant.SignalCallDirectionMissed
}

// MaybeUpsertCallRecordFromChatMsg 解析服务端发往单聊会话的 rtcCallRecord 通话记录消息，
// 并将其同步到本地通话记录表。
//
// 逻辑：
//   - 若本地已存在同一 roomID 的记录（正常信令流已写入），
//     则以服务端权威值更新 status / callDuration / action；
//   - 若本地不存在（多端同步、重装等场景），则构建完整记录并插入。
//
// 仅处理单聊（groupID 为空）的 Custom(110) 消息；群通话不使用此消息类型。
// 错误非致命，仅记录日志，不影响消息正常入库。
func (s *Signaling) MaybeUpsertCallRecordFromChatMsg(ctx context.Context, msg *sdk_struct.MsgStruct) {
	if s.db == nil || msg == nil || msg.ContentType != constant.Custom || msg.CustomElem == nil {
		return
	}

	var data serverCallRecordData
	if err := json.Unmarshal([]byte(msg.CustomElem.Data), &data); err != nil {
		return
	}
	if data.CustomType != rtcCallRecordCustomType || data.RoomID == "" {
		return
	}

	status, action := serverStatusToSDK(data.Status)
	// 服务端 duration 单位为秒，本地 CallDuration 单位为毫秒。
	callDurationMs := data.Duration * 1000

	existing, err := s.db.GetSignalCallRecordByRoomID(ctx, data.RoomID)
	if err == nil && existing != nil {
		// 已存在：仅更新服务端权威字段，保留本地写入的昵称/头像/时间戳等。
		existing.Status = status
		existing.CallDuration = callDurationMs
		existing.Action = action
		if err := s.db.BatchUpsertSignalCallRecords(ctx, []*model_struct.LocalSignalCallRecord{existing}); err != nil {
			log.ZWarn(ctx, "MaybeUpsertCallRecordFromChatMsg: update existing failed", err, "roomID", data.RoomID)
		} else {
			log.ZInfo(ctx, "MaybeUpsertCallRecordFromChatMsg: updated existing record",
				"roomID", data.RoomID, "status", status, "callDurationMs", callDurationMs)
		}
		return
	}
	if err != nil && !errs.ErrRecordNotFound.Is(err) {
		log.ZWarn(ctx, "MaybeUpsertCallRecordFromChatMsg: GetSignalCallRecordByRoomID failed", err, "roomID", data.RoomID)
	}

	// 不存在：构建新记录（多端同步 / 信令通知丢失场景）。
	inviteeUID := ""
	inviteeIDsJSON := ""
	if len(data.InviteeUserIDList) > 0 {
		inviteeUID = data.InviteeUserIDList[0]
		if b, e := json.Marshal(data.InviteeUserIDList); e == nil {
			inviteeIDsJSON = string(b)
		}
	}

	direction := s.resolveCallRecordDirectionByID(data.InviterUserID, status)

	inv := &rtc.InvitationInfo{
		RoomID:            data.RoomID,
		InviterUserID:     data.InviterUserID,
		InviteeUserIDList: data.InviteeUserIDList,
		MediaType:         data.MediaType,
		SessionType:       constant.SingleChatType,
	}
	inviterNickname := s.resolveInviterNickname(ctx, inv, nil)
	inviterFaceURL := s.resolveInviterFaceURL(ctx, inv, nil)
	inviteeNickname := s.resolveInviteeNickname(ctx, inv, nil)
	inviteeFaceURL := s.resolveInviteeFaceURL(ctx, inv, nil)
	calleeMatchText := s.composeCalleeMatchText(ctx, inv, nil)

	sendTime := msg.SendTime
	// 服务端消息发送时间约等于通话结束时刻；若已接通则反推接通时间。
	var connectMs int64
	if status == constant.SignalCallStatusAnswered && callDurationMs > 0 {
		connectMs = sendTime - callDurationMs
	}

	// SID 以 "server-" 前缀区别于本地信令流写入的记录，保持幂等。
	sid := fmt.Sprintf("server-%s", data.RoomID)

	rec := &model_struct.LocalSignalCallRecord{
		SID:                 sid,
		RoomID:              data.RoomID,
		Status:              status,
		MediaType:           data.MediaType,
		SessionType:         constant.SingleChatType,
		InviterUserID:       data.InviterUserID,
		InviterUserNickname: inviterNickname,
		InviterUserFaceURL:  inviterFaceURL,
		CreateTime:          sendTime,
		EndTime:             sendTime,
		ConnectTime:         connectMs,
		CallDuration:        callDurationMs,
		DialDuration:        0,
		InviteeUID:          inviteeUID,
		InviteeUserIDsJSON:  inviteeIDsJSON,
		InviteeUserNickname: inviteeNickname,
		InviteeUserFaceURL:  inviteeFaceURL,
		CalleeMatchText:     calleeMatchText,
		Direction:           direction,
		Role:                constant.SignalCallRoleFromDirection(direction),
		Action:              action,
	}
	if err := s.db.BatchUpsertSignalCallRecords(ctx, []*model_struct.LocalSignalCallRecord{rec}); err != nil {
		log.ZWarn(ctx, "MaybeUpsertCallRecordFromChatMsg: insert new record failed", err, "roomID", data.RoomID)
	} else {
		log.ZInfo(ctx, "MaybeUpsertCallRecordFromChatMsg: inserted new record",
			"roomID", data.RoomID, "sid", sid, "status", status)
	}
}
