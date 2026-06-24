package relation

import (
	"context"
	"strings"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/tools/log"
)

// syncCallRecordsUserProfile updates stored nicknames/avatars in local call history
// after a friend's display profile changes (e.g. account deactivated).
// It fetches the latest profile from the server so deactivated accounts receive the
// configured placeholder nickname and faceURL instead of stale local data.
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

	log.ZInfo(ctx, "lintao syncCallRecordsUserProfile start",
		"loginUserID", r.loginUserID, "friendUserIDs", ids)

	serverFriends, err := r.getDesignatedFriends(ctx, ids)
	if err != nil {
		log.ZWarn(ctx, "lintao syncCallRecordsUserProfile getDesignatedFriends failed, fallback to local",
			err, "loginUserID", r.loginUserID, "friendUserIDs", ids)
		r.syncCallRecordsUserProfileFromLocal(ctx, ids)
		return
	}

	log.ZInfo(ctx, "lintao syncCallRecordsUserProfile getDesignatedFriends ok",
		"loginUserID", r.loginUserID, "friendUserIDs", ids, "serverCount", len(serverFriends))

	synced := make(map[string]struct{}, len(serverFriends))
	for _, sf := range serverFriends {
		local := ServerFriendToLocalFriend(sf)
		if local == nil || local.FriendUserID == "" {
			log.ZWarn(ctx, "lintao syncCallRecordsUserProfile skip invalid server friend", nil,
				"loginUserID", r.loginUserID)
			continue
		}
		synced[local.FriendUserID] = struct{}{}
		r.applyCallRecordUserProfile(ctx, local, "server")
	}

	var missing []string
	for _, id := range ids {
		if _, ok := synced[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		log.ZInfo(ctx, "lintao syncCallRecordsUserProfile fallback to local for missing IDs",
			"loginUserID", r.loginUserID, "missingFriendUserIDs", missing)
		r.syncCallRecordsUserProfileFromLocal(ctx, missing)
	}

	log.ZInfo(ctx, "lintao syncCallRecordsUserProfile done",
		"loginUserID", r.loginUserID, "friendUserIDs", ids,
		"syncedFromServer", len(synced), "fallbackLocal", len(missing))
}

func (r *Relation) syncCallRecordsUserProfileFromLocal(ctx context.Context, ids []string) {
	log.ZInfo(ctx, "lintao syncCallRecordsUserProfileFromLocal start",
		"loginUserID", r.loginUserID, "friendUserIDs", ids)

	friends, err := r.db.GetFriendInfoList(ctx, ids)
	if err != nil {
		log.ZWarn(ctx, "lintao syncCallRecordsUserProfileFromLocal GetFriendInfoList failed", err,
			"loginUserID", r.loginUserID, "friendUserIDs", ids)
		return
	}

	log.ZInfo(ctx, "lintao syncCallRecordsUserProfileFromLocal GetFriendInfoList ok",
		"loginUserID", r.loginUserID, "friendUserIDs", ids, "localCount", len(friends))

	for _, f := range friends {
		if f == nil || f.FriendUserID == "" {
			continue
		}
		r.applyCallRecordUserProfile(ctx, f, "local")
	}
}

func (r *Relation) applyCallRecordUserProfile(ctx context.Context, f *model_struct.LocalFriend, source string) {
	showName := f.ConversationShowName()
	if showName == "" {
		log.ZInfo(ctx, "lintao syncCallRecordsUserProfile skip empty showName",
			"loginUserID", r.loginUserID, "friendUserID", f.FriendUserID,
			"source", source, "nickname", f.Nickname, "firstName", f.FirstName,
			"lastName", f.LastName, "remark", f.Remark)
		return
	}
	if err := r.db.UpdateSignalCallRecordUserProfile(ctx, f.FriendUserID, showName, f.FaceURL); err != nil {
		log.ZWarn(ctx, "lintao syncCallRecordsUserProfile UpdateSignalCallRecordUserProfile failed", err,
			"loginUserID", r.loginUserID, "friendUserID", f.FriendUserID,
			"source", source, "showName", showName, "faceURL", f.FaceURL)
		return
	}
	log.ZInfo(ctx, "lintao syncCallRecordsUserProfile UpdateSignalCallRecordUserProfile ok",
		"loginUserID", r.loginUserID, "friendUserID", f.FriendUserID,
		"source", source, "showName", showName, "faceURL", f.FaceURL)
	if r.invalidateCallRecordDetailCache != nil {
		r.invalidateCallRecordDetailCache(ctx, f.FriendUserID)
	}
}
