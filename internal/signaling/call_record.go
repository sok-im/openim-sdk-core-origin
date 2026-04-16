package signaling

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/protocol/rtc"
)

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

// newLocalSignalCallRecord 生成本地通话记录；calleeMatchText 为被叫侧模糊检索串（见 composeCalleeMatchText）。
func newLocalSignalCallRecord(inv *rtc.InvitationInfo, endMs int64, dialStatus int32, calleeMatchText string) *model_struct.LocalSignalCallRecord {
	if inv == nil || inv.RoomID == "" {
		return nil
	}
	prefix := "local-nc"
	if dialStatus == constant.SignalCallDialStatusConnected {
		prefix = "local"
	}
	sid := fmt.Sprintf("%s-%s-%d", prefix, inv.RoomID, endMs)
	createTime := inv.InitiateTime
	if createTime == 0 {
		createTime = endMs
	}
	return &model_struct.LocalSignalCallRecord{
		SID:             sid,
		RoomID:          inv.RoomID,
		MediaType:       inv.MediaType,
		SessionType:     inv.SessionType,
		InviterUserID:   inv.InviterUserID,
		GroupID:         inv.GroupID,
		CreateTime:      createTime,
		EndTime:         endMs,
		InviterUsersJSON: "",
		CalleeMatchText: strings.TrimSpace(calleeMatchText),
		DialStatus:      dialStatus,
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
