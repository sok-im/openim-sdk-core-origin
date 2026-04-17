package signaling

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/protocol/rtc"
)

// callRecordConfig 构建本地通话记录所需的全部参数。
type callRecordConfig struct {
	inv             *rtc.InvitationInfo
	dialStatus      int32  // constant.SignalCallDialStatus*
	direction       int32  // constant.SignalCallDirection*
	inviteMs        int64  // 本端发起/收到邀请的毫秒时间戳（0 则回退到 inv.InitiateTime）
	connectMs       int64  // 接通时的毫秒时间戳（0 = 未接通）
	endMs           int64  // 通话结束时的毫秒时间戳
	calleeMatchText string // 被叫检索串（用于模糊查询）
	inviteeNickname string // 被叫昵称（单聊展示用）
	inviteeIDsJSON  string // 被叫 userID 列表 JSON
	inviterNickname string // 主叫昵称（被叫侧展示用）
}

func localToSignalRecord(l *model_struct.LocalSignalCallRecord) *rtc.SignalRecord {
	if l == nil {
		return nil
	}
	r := &rtc.SignalRecord{
		RoomID:              l.RoomID,
		SID:                 l.SID,
		FileName:            l.FileName,
		MediaType:           l.MediaType,
		SessionType:         l.SessionType,
		InviterUserID:       l.InviterUserID,
		InviterUserNickname: l.InviterUserNickname,
		GroupID:             l.GroupID,
		GroupName:           l.GroupName,
		CreateTime:          l.CreateTime,
		EndTime:             l.EndTime,
		Size:                l.Size,
		FileURL:             l.FileURL,
	}
	if l.InviterUsersJSON != "" {
		var users []*rtc.SignalUser
		if err := json.Unmarshal([]byte(l.InviterUsersJSON), &users); err == nil {
			r.InviterUsers = users
		}
	}
	return r
}

// newLocalSignalCallRecord 根据 callRecordConfig 生成本地通话记录，并计算拨打时长与通话时长。
func newLocalSignalCallRecord(cfg callRecordConfig) *model_struct.LocalSignalCallRecord {
	inv := cfg.inv
	if inv == nil || inv.RoomID == "" {
		return nil
	}

	// 确定通话开始时间（主叫侧拨出时刻 / 被叫侧收到邀请时刻）
	inviteMs := cfg.inviteMs
	if inviteMs <= 0 {
		inviteMs = inv.InitiateTime
	}
	if inviteMs <= 0 {
		inviteMs = cfg.endMs
	}

	// 计算拨打时长与通话时长
	var dialDuration, callDuration int64
	if cfg.connectMs > 0 {
		// 已接通：振铃阶段 = connectMs - inviteMs；通话阶段 = endMs - connectMs
		dialDuration = cfg.connectMs - inviteMs
		callDuration = cfg.endMs - cfg.connectMs
	} else {
		// 未接通：全程为振铃阶段
		dialDuration = cfg.endMs - inviteMs
	}
	if dialDuration < 0 {
		dialDuration = 0
	}
	if callDuration < 0 {
		callDuration = 0
	}

	// SID 前缀区分已接通 / 未接通，便于离线排查
	prefix := "local-nc"
	if cfg.dialStatus == constant.SignalCallDialStatusConnected {
		prefix = "local"
	}
	sid := fmt.Sprintf("%s-%s-%d", prefix, inv.RoomID, cfg.endMs)

	return &model_struct.LocalSignalCallRecord{
		SID:                 sid,
		RoomID:              inv.RoomID,
		MediaType:           inv.MediaType,
		SessionType:         inv.SessionType,
		InviterUserID:       inv.InviterUserID,
		InviterUserNickname: cfg.inviterNickname,
		GroupID:             inv.GroupID,
		CreateTime:          inviteMs,
		EndTime:             cfg.endMs,
		ConnectTime:         cfg.connectMs,
		DialDuration:        dialDuration,
		CallDuration:        callDuration,
		InviteeUserNickname: cfg.inviteeNickname,
		InviteeUserIDsJSON:  cfg.inviteeIDsJSON,
		CalleeMatchText:     strings.TrimSpace(cfg.calleeMatchText),
		DialStatus:          cfg.dialStatus,
		Direction:           cfg.direction,
	}
}

// buildCalleeMatchTextFromProto 从邀请与单次 participant 元数据提取被叫 userID + 昵称（信令里可能带的被叫 PublicUserInfo / 群成员昵称）。
func buildCalleeMatchTextFromProto(inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if inv == nil {
		return ""
	}
	var parts []string
	for _, uid := range inv.InviteeUserIDList {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		parts = append(parts, uid)
		if p == nil {
			continue
		}
		if p.UserInfo != nil && p.UserInfo.UserID == uid && p.UserInfo.Nickname != "" {
			parts = append(parts, p.UserInfo.Nickname)
		}
		if p.GroupMemberInfo != nil && p.GroupMemberInfo.UserID == uid && p.GroupMemberInfo.Nickname != "" {
			parts = append(parts, p.GroupMemberInfo.Nickname)
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

// extractInviteeNickname 从 participant 元数据或 inviteeUserIDList 提取首位被叫的展示昵称（单聊场景）。
func extractInviteeNickname(inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if inv == nil || len(inv.InviteeUserIDList) == 0 {
		return ""
	}
	firstUID := inv.InviteeUserIDList[0]
	if p != nil {
		if p.UserInfo != nil && p.UserInfo.UserID == firstUID && p.UserInfo.Nickname != "" {
			return p.UserInfo.Nickname
		}
		if p.GroupMemberInfo != nil && p.GroupMemberInfo.UserID == firstUID && p.GroupMemberInfo.Nickname != "" {
			return p.GroupMemberInfo.Nickname
		}
	}
	return ""
}

// extractInviterNickname 从 participant 元数据提取主叫的展示昵称。
func extractInviterNickname(inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if inv == nil || p == nil {
		return ""
	}
	if p.UserInfo != nil && p.UserInfo.UserID == inv.InviterUserID && p.UserInfo.Nickname != "" {
		return p.UserInfo.Nickname
	}
	if p.GroupMemberInfo != nil && p.GroupMemberInfo.UserID == inv.InviterUserID && p.GroupMemberInfo.Nickname != "" {
		return p.GroupMemberInfo.Nickname
	}
	return ""
}

// buildInviteeIDsJSON 将被叫 userID 列表序列化为 JSON 字符串（空列表返回空串）。
func buildInviteeIDsJSON(inv *rtc.InvitationInfo) string {
	if inv == nil || len(inv.InviteeUserIDList) == 0 {
		return ""
	}
	b, err := json.Marshal(inv.InviteeUserIDList)
	if err != nil {
		return ""
	}
	return string(b)
}
