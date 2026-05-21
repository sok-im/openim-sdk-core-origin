package signaling

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"
	"github.com/openimsdk/protocol/rtc"
)

// callRecordConfig 构建本地通话记录所需的全部参数。
type callRecordConfig struct {
	inv              *rtc.InvitationInfo
	status           int32 // constant.SignalCallStatus*
	direction        int32 // constant.SignalCallDirection*
	inviteMs         int64
	connectMs        int64
	endMs            int64
	calleeMatchText  string
	inviteeNickname  string
	inviteeIDsJSON   string
	inviterNickname  string
	inviterFaceURL   string
	groupName        string
}

// newLocalSignalCallRecord 根据 callRecordConfig 生成本地通话记录，并计算拨打时长与通话时长。
func newLocalSignalCallRecord(cfg callRecordConfig) *model_struct.LocalSignalCallRecord {
	inv := cfg.inv
	if inv == nil || inv.RoomID == "" {
		return nil
	}

	inviteMs := cfg.inviteMs
	if inviteMs <= 0 {
		inviteMs = inv.InitiateTime
	}
	if inviteMs <= 0 {
		inviteMs = cfg.endMs
	}

	var dialDuration, callDuration int64
	if cfg.connectMs > 0 {
		dialDuration = cfg.connectMs - inviteMs
		callDuration = cfg.endMs - cfg.connectMs
	} else {
		dialDuration = cfg.endMs - inviteMs
	}
	if dialDuration < 0 {
		dialDuration = 0
	}
	if callDuration < 0 {
		callDuration = 0
	}

	prefix := "local-nc"
	if cfg.status == constant.SignalCallStatusAnswered {
		prefix = "local"
	}
	sid := fmt.Sprintf("%s-%s-%d", prefix, inv.RoomID, cfg.endMs)

	return &model_struct.LocalSignalCallRecord{
		SID:                 sid,
		RoomID:              inv.RoomID,
		Status:              cfg.status,
		MediaType:           inv.MediaType,
		SessionType:         inv.SessionType,
		InviterUserID:       inv.InviterUserID,
		InviterUserNickname: cfg.inviterNickname,
		InviterUserFaceURL:  cfg.inviterFaceURL,
		GroupID:             inv.GroupID,
		GroupName:           cfg.groupName,
		CreateTime:          inviteMs,
		EndTime:             cfg.endMs,
		ConnectTime:         cfg.connectMs,
		DialDuration:        dialDuration,
		CallDuration:        callDuration,
		InviteeUserNickname: cfg.inviteeNickname,
		InviteeUserIDsJSON:  cfg.inviteeIDsJSON,
		CalleeMatchText:     strings.TrimSpace(cfg.calleeMatchText),
		Direction:           cfg.direction,
		Role:                constant.SignalCallRoleFromDirection(cfg.direction),
	}
}

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

func nicknameFromParticipant(userID string, p *rtc.ParticipantMetaData) string {
	if p == nil || userID == "" {
		return ""
	}
	if p.UserInfo != nil && p.UserInfo.UserID == userID {
		return p.UserInfo.Nickname
	}
	if p.GroupMemberInfo != nil && p.GroupMemberInfo.UserID == userID {
		return p.GroupMemberInfo.Nickname
	}
	return ""
}

func extractInviteeNickname(inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if inv == nil || len(inv.InviteeUserIDList) == 0 {
		return ""
	}
	return nicknameFromParticipant(inv.InviteeUserIDList[0], p)
}

func extractInviterNickname(inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if inv == nil {
		return ""
	}
	return nicknameFromParticipant(inv.InviterUserID, p)
}

func extractInviterFaceURL(inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if inv == nil || p == nil {
		return ""
	}
	if p.UserInfo != nil && p.UserInfo.UserID == inv.InviterUserID && p.UserInfo.FaceURL != "" {
		return p.UserInfo.FaceURL
	}
	if p.GroupMemberInfo != nil && p.GroupMemberInfo.UserID == inv.InviterUserID && p.GroupMemberInfo.FaceURL != "" {
		return p.GroupMemberInfo.FaceURL
	}
	return ""
}

func extractGroupName(p *rtc.ParticipantMetaData) string {
	if p != nil && p.GroupInfo != nil && p.GroupInfo.GroupName != "" {
		return p.GroupInfo.GroupName
	}
	return ""
}

func callRecordRole(l *model_struct.LocalSignalCallRecord) int32 {
	if l == nil {
		return constant.SignalCallRoleUnknown
	}
	if l.Role != 0 {
		return l.Role
	}
	return constant.SignalCallRoleFromDirection(l.Direction)
}

func localRecordToSDK(l *model_struct.LocalSignalCallRecord) *sdk_struct.SignalCallRecordWithDialStatus {
	if l == nil {
		return nil
	}
	return &sdk_struct.SignalCallRecordWithDialStatus{
		SID:                 l.SID,
		RoomID:              l.RoomID,
		Status:              l.Status,
		CreateTime:          l.CreateTime,
		DialDuration:        l.DialDuration,
		CallDuration:        l.CallDuration,
		MediaType:           l.MediaType,
		SessionType:         l.SessionType,
		InviterUserID:       l.InviterUserID,
		InviterUserNickname: l.InviterUserNickname,
		InviterUserFaceURL:  l.InviterUserFaceURL,
		GroupID:             l.GroupID,
		GroupName:           l.GroupName,
		Direction:           l.Direction,
		Role:                callRecordRole(l),
		ConnectTime:         l.ConnectTime,
		InviteeUserNickname: l.InviteeUserNickname,
	}
}

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
