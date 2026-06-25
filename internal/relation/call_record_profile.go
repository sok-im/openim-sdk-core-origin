package relation

import (
	"context"
	"strings"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdkerrs"
	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/log"
)

// SyncCallRecordsUserProfile 在好友资料变更后，同步更新本地通话记录中的展示昵称与头像。
func (r *Relation) SyncCallRecordsUserProfile(ctx context.Context, friendUserIDs ...string) {
	r.syncCallRecordsUserProfile(ctx, friendUserIDs...)
}

// syncCallRecordsUserProfile 在好友资料变更后，同步更新本地通话记录中的展示昵称与头像。
func (r *Relation) syncCallRecordsUserProfile(ctx context.Context, friendUserIDs ...string) {
	log.ZDebug(ctx, "lintao syncCallRecordsUserProfile", "friendUserIDs", friendUserIDs)
	r.syncCallRecordsUserProfileWithMode(ctx, false, friendUserIDs...)
}

// syncCallRecordsUserProfileForRemovedFriend 在好友关系被删除（含对方注销账号）后刷新通话记录展示。
func (r *Relation) syncCallRecordsUserProfileForRemovedFriend(ctx context.Context, friendUserIDs ...string) {
	log.ZDebug(ctx, "lintao syncCallRecordsUserProfileForRemovedFriend", "friendUserIDs", friendUserIDs)
	r.syncCallRecordsUserProfileWithMode(ctx, true, friendUserIDs...)
}

func (r *Relation) syncCallRecordsUserProfileWithMode(ctx context.Context, friendRemoved bool, friendUserIDs ...string) {
	if r.db == nil || len(friendUserIDs) == 0 {
		log.ZDebug(ctx, "lintao syncCallRecordsUserProfileWithMode db is nil or friendUserIDs is empty", "friendUserIDs", friendUserIDs)
		return
	}
	ids := make([]string, 0, len(friendUserIDs))
	for _, id := range friendUserIDs {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		log.ZDebug(ctx, "lintao syncCallRecordsUserProfileWithMode ids is empty", "friendUserIDs", friendUserIDs)
		return
	}

	for _, userID := range ids {
		if friendRemoved {
			log.ZDebug(ctx, "lintao syncCallRecordsUserProfileForRemovedFriend", "userID", userID)
			r.syncCallRecordUserProfileForRemovedFriend(ctx, userID)
		} else {
			log.ZDebug(ctx, "lintao syncCallRecordsUserProfile", "userID", userID)
			r.syncCallRecordUserProfile(ctx, userID)
		}
	}
}

func (r *Relation) syncCallRecordUserProfileForRemovedFriend(ctx context.Context, userID string) {
	if serverUser, err := r.user.GetSingleUserFromServer(ctx, userID); err == nil && serverUser != nil {
		showName := serverUser.DisplayName()
		if showName == constant.DeactivatedUserNickname || serverUser.FaceURL == constant.DeactivatedUserFaceURL {
			log.ZDebug(ctx, "lintao syncCallRecordUserProfileForRemovedFriend serverUser is deactivated", "userID", userID)
			r.applyCallRecordUserProfile(ctx, userID, showName, serverUser.FaceURL, "server-user")
			return
		}
	}
	log.ZDebug(ctx, "lintao syncCallRecordUserProfileForRemovedFriend serverUser is not deactivated", "userID", userID)
	r.applyCallRecordUserProfile(ctx, userID, constant.DeactivatedUserNickname, constant.DeactivatedUserFaceURL, "deactivated")
}

func (r *Relation) syncCallRecordUserProfile(ctx context.Context, userID string) {
	if serverUser, err := r.user.GetSingleUserFromServer(ctx, userID); err == nil && serverUser != nil {
		r.applyCallRecordUserProfile(ctx, userID, serverUser.DisplayName(), serverUser.FaceURL, "server-user")
		return
	} else if err != nil && !sdkerrs.ErrUserIDNotFound.Is(errs.Unwrap(err)) {
		log.ZWarn(ctx, "lintao syncCallRecordUserProfile GetSingleUserFromServer failed", err,
			"loginUserID", r.loginUserID, "friendUserID", userID)
	}

	if serverFriends, err := r.getDesignatedFriends(ctx, []string{userID}); err == nil {
		for _, sf := range serverFriends {
			local := ServerFriendToLocalFriend(sf)
			if local != nil && local.FriendUserID == userID {
				log.ZDebug(ctx, "lintao syncCallRecordUserProfile server-friend", "userID", userID)
				r.applyCallRecordUserProfile(ctx, userID, local.ConversationShowName(), local.FaceURL, "server-friend")
				return
			}
		}
	}

	friends, err := r.db.GetFriendInfoList(ctx, []string{userID})
	if err == nil {
		for _, f := range friends {
			if f != nil && f.FriendUserID == userID {
				log.ZDebug(ctx, "lintao syncCallRecordUserProfile local-friend", "userID", userID)
				r.applyCallRecordUserProfile(ctx, userID, f.ConversationShowName(), f.FaceURL, "local-friend")
				return
			}
		}
	}

	log.ZDebug(ctx, "lintao syncCallRecordUserProfile no friend found", "userID", userID)
	r.applyCallRecordUserProfile(ctx, userID, constant.DeactivatedUserNickname, constant.DeactivatedUserFaceURL, "deactivated")
}

func (r *Relation) applyCallRecordUserProfile(ctx context.Context, userID, showName, faceURL, source string) {
	showName = strings.TrimSpace(showName)
	if showName == "" && faceURL == "" {
		return
	}
	var err error
	if r.updateCallRecordsUserProfile != nil {
		err = r.updateCallRecordsUserProfile(ctx, userID, showName, faceURL)
		log.ZDebug(ctx, "lintao applyCallRecordUserProfile updateCallRecordsUserProfile", "userID", userID, "showName", showName, "faceURL", faceURL)
	} else {
		err = r.db.UpdateSignalCallRecordUserProfile(ctx, userID, showName, faceURL)
		log.ZDebug(ctx, "lintao applyCallRecordUserProfile updateSignalCallRecordUserProfile", "userID", userID, "showName", showName, "faceURL", faceURL)
	}
	if err != nil {
		log.ZWarn(ctx, "lintao syncCallRecordsUserProfile update call record profile failed", err,
			"loginUserID", r.loginUserID, "friendUserID", userID,
			"source", source, "showName", showName, "faceURL", faceURL)
		return
	}
	log.ZDebug(ctx, "lintao syncCallRecordsUserProfile update call record profile ok",
		"loginUserID", r.loginUserID, "friendUserID", userID,
		"source", source, "showName", showName, "faceURL", faceURL)
}
