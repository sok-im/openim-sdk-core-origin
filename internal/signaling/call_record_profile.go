package signaling

import (
	"context"
	"strings"

	"github.com/openimsdk/tools/log"
)

// UpdateCallRecordsUserProfile 同步更新本地通话记录中该用户的展示昵称、头像及检索字段。
func (s *Signaling) UpdateCallRecordsUserProfile(ctx context.Context, userID, nickname, faceURL string) error {
	if s.db == nil {
		return nil
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	if nickname == "" && faceURL == "" {
		return nil
	}
	if err := s.db.UpdateSignalCallRecordUserProfile(ctx, userID, nickname, faceURL); err != nil {
		return err
	}

	searchTokens := s.collectUserSearchTokens(ctx, userID, nickname)
	if len(searchTokens) == 0 {
		return nil
	}

	records, err := s.db.ListSignalCallRecordsByParticipant(ctx, userID)
	if err != nil {
		return err
	}
	for _, rec := range records {
		if rec == nil || rec.SID == "" {
			continue
		}
		newText := appendUniqueSearchTokens(rec.CalleeMatchText, searchTokens)
		if newText == rec.CalleeMatchText {
			continue
		}
		if err := s.db.UpdateSignalCallRecordCalleeMatchText(ctx, rec.SID, newText); err != nil {
			log.ZWarn(ctx, "UpdateSignalCallRecordCalleeMatchText failed", err, "sID", rec.SID, "userID", userID)
			continue
		}
		if s.detailCache != nil {
			s.detailCache.Delete(rec.SID)
		}
	}
	return nil
}

func (s *Signaling) collectUserSearchTokens(ctx context.Context, userID, fallbackNickname string) []string {
	if s.db == nil {
		if fallbackNickname != "" {
			return []string{fallbackNickname}
		}
		return nil
	}
	friends, err := s.db.GetFriendInfoList(ctx, []string{userID})
	if err == nil && len(friends) > 0 && friends[0] != nil {
		return searchableFriendTokens(friends[0])
	}
	user, err := s.db.GetLoginUser(ctx, userID)
	if err == nil && user != nil {
		return searchableUserTokens(user)
	}
	if fallbackNickname != "" {
		return []string{fallbackNickname}
	}
	return nil
}
