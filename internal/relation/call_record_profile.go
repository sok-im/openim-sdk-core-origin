package relation

import (
	"context"
	"strings"

	"github.com/openimsdk/tools/log"
)

// syncCallRecordsUserProfile updates stored nicknames/avatars in local call history
// after a friend's display profile changes (e.g. account deactivated).
func (r *Relation) syncCallRecordsUserProfile(ctx context.Context, friendUserIDs ...string) {
	if r.db == nil || len(friendUserIDs) == 0 {
		return
	}
	ids := make([]string, 0, len(friendUserIDs))
	for _, id := range friendUserIDs {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return
	}

	friends, err := r.db.GetFriendInfoList(ctx, ids)
	if err != nil {
		log.ZWarn(ctx, "syncCallRecordsUserProfile GetFriendInfoList failed", err, "friendUserIDs", ids)
		return
	}
	for _, f := range friends {
		if f == nil || f.FriendUserID == "" {
			continue
		}
		showName := f.ConversationShowName()
		if showName == "" {
			continue
		}
		if err := r.db.UpdateSignalCallRecordUserProfile(ctx, f.FriendUserID, showName, f.FaceURL); err != nil {
			log.ZWarn(ctx, "syncCallRecordsUserProfile UpdateSignalCallRecordUserProfile failed", err,
				"friendUserID", f.FriendUserID, "showName", showName)
		}
	}
}
