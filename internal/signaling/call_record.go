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
	action           string // constant.SignalCallAction*
	inviteMs         int64
	connectMs        int64
	endMs            int64
	// callDuration: 客户端主动上报的通话时长（毫秒）。>0 时直接使用，不再从时间戳推算，
	// 用于保证主/被叫双端记录时长一致。
	callDuration     int64
	calleeMatchText  string
	inviteeNickname  string
	inviteeUID       string
	inviteeFaceURL   string
	inviteeIDsJSON   string
	inviterNickname  string
	inviterFaceURL   string
	groupName        string
	ownerUserID      string
}

// newLocalSignalCallRecord 根据 callRecordConfig 生成本地通话记录，并计算拨打时长与通话时长。
func newLocalSignalCallRecord(cfg callRecordConfig) *model_struct.LocalSignalCallRecord {
	inv := cfg.inv
	if inv == nil || inv.RoomID == "" {
		return nil
	}

	// 未接通记录不应携带接通时间，避免误算 callDuration（如超时却被展示成通话时长）。
	if cfg.status == constant.SignalCallStatusNotConnected {
		cfg.connectMs = 0
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
		if cfg.callDuration > 0 {
			// 主动方挂断时上报的精确时长，优先于本地时间戳推算，确保双端一致。
			callDuration = cfg.callDuration
		} else {
			callDuration = cfg.endMs - cfg.connectMs
		}
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
	// 同一 room 一次通话使用 inviteMs 作为稳定主键，避免接听/挂断多次写入产生重复记录。
	sidKey := inviteMs
	if sidKey <= 0 {
		sidKey = cfg.endMs
	}
	sid := fmt.Sprintf("%s-%s-%d", prefix, inv.RoomID, sidKey)

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
		InviteeUID:          cfg.inviteeUID,
		InviteeUserFaceURL:  cfg.inviteeFaceURL,
		InviteeUserIDsJSON:  cfg.inviteeIDsJSON,
		CalleeMatchText:     strings.TrimSpace(cfg.calleeMatchText),
		Direction:           cfg.direction,
		Role:                constant.SignalCallRoleFromDirection(cfg.direction),
		Action:              cfg.action,
		OwnerUserID:         strings.TrimSpace(cfg.ownerUserID),
	}
}

// searchableFriendTokens 返回好友侧可用于通话记录模糊检索的展示名片段（含 first/last 单独字段）。
func searchableFriendTokens(f *model_struct.LocalFriend) []string {
	if f == nil {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	add(f.Remark)
	add(f.FirstName)
	add(f.LastName)
	add(model_struct.UserDisplayName(f.FirstName, f.LastName, f.Nickname))
	add(f.Nickname)
	return out
}

// searchableUserTokens 返回本地用户表可用于通话记录模糊检索的展示名片段。
func searchableUserTokens(u *model_struct.LocalUser) []string {
	if u == nil {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	add(u.FirstName)
	add(u.LastName)
	add(u.DisplayName())
	add(u.Nickname)
	return out
}

func appendUniqueSearchTokens(base string, tokens []string) string {
	if len(tokens) == 0 {
		return strings.TrimSpace(base)
	}
	seen := make(map[string]struct{})
	var extras []string
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		extras = append(extras, t)
	}
	if len(extras) == 0 {
		return strings.TrimSpace(base)
	}
	if base == "" {
		return strings.Join(extras, " ")
	}
	return strings.TrimSpace(base + " " + strings.Join(extras, " "))
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

func extractPrimaryInviteeUID(inv *rtc.InvitationInfo) string {
	if inv == nil {
		return ""
	}
	inviterID := strings.TrimSpace(inv.InviterUserID)
	for _, uid := range inv.InviteeUserIDList {
		if uid = strings.TrimSpace(uid); uid == "" || uid == inviterID {
			continue
		}
		return uid
	}
	// 兼容仅含主叫自身的异常列表，避免落库时 peer 为空。
	for _, uid := range inv.InviteeUserIDList {
		if uid = strings.TrimSpace(uid); uid != "" {
			return uid
		}
	}
	return ""
}

func faceURLFromParticipant(userID string, p *rtc.ParticipantMetaData) string {
	if p == nil || userID == "" {
		return ""
	}
	if p.UserInfo != nil && p.UserInfo.UserID == userID && p.UserInfo.FaceURL != "" {
		return p.UserInfo.FaceURL
	}
	if p.GroupMemberInfo != nil && p.GroupMemberInfo.UserID == userID && p.GroupMemberInfo.FaceURL != "" {
		return p.GroupMemberInfo.FaceURL
	}
	return ""
}

func extractInviteeFaceURL(inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	uid := extractPrimaryInviteeUID(inv)
	if uid == "" {
		return ""
	}
	return faceURLFromParticipant(uid, p)
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
	uid := extractPrimaryInviteeUID(inv)
	if uid == "" {
		return ""
	}
	return nicknameFromParticipant(uid, p)
}

func extractInviterNickname(inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if inv == nil {
		return ""
	}
	return nicknameFromParticipant(inv.InviterUserID, p)
}

func extractInviterFaceURL(inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if inv == nil {
		return ""
	}
	return faceURLFromParticipant(inv.InviterUserID, p)
}

func extractGroupName(p *rtc.ParticipantMetaData) string {
	if p != nil && p.GroupInfo != nil && p.GroupInfo.GroupName != "" {
		return p.GroupInfo.GroupName
	}
	return ""
}

func extractGroupFaceURL(p *rtc.ParticipantMetaData) string {
	if p != nil && p.GroupInfo != nil && p.GroupInfo.FaceURL != "" {
		return p.GroupInfo.FaceURL
	}
	return ""
}

// isGroupChatCall 判定是否为群音视频通话（此类记录不落本地通话表）。
func isGroupChatCall(inv *rtc.InvitationInfo) bool {
	if inv == nil {
		return false
	}
	if strings.TrimSpace(inv.GroupID) != "" {
		return true
	}
	return inv.SessionType == constant.WriteGroupChatType || inv.SessionType == constant.ReadGroupChatType
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
		InviteeUID:          l.InviteeUID,
		InviteeUserFaceURL:  l.InviteeUserFaceURL,
		Action:              l.Action,
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
