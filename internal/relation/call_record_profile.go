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
	deactivatedNickname := r.deactivatedUserNickname(ctx)
	if serverUser, err := r.user.GetSingleUserFromServer(ctx, userID); err == nil && serverUser != nil {
		showName := serverUser.DisplayName()
		if constant.IsDeactivatedUserNickname(showName) || serverUser.FaceURL == constant.DeactivatedUserFaceURL {
			log.ZDebug(ctx, "lintao syncCallRecordUserProfileForRemovedFriend serverUser is deactivated", "userID", userID)
			r.applyCallRecordUserProfile(ctx, userID, deactivatedNickname, serverUser.FaceURL, "server-user")
			return
		}
	}
	log.ZDebug(ctx, "lintao syncCallRecordUserProfileForRemovedFriend serverUser is not deactivated", "userID", userID)
	r.applyCallRecordUserProfile(ctx, userID, deactivatedNickname, constant.DeactivatedUserFaceURL, "deactivated")
}

func (r *Relation) syncCallRecordUserProfile(ctx context.Context, userID string) {
	if userID == r.loginUserID {
		r.syncCallRecordLoginUserProfile(ctx)
		return
	}

	serverFriends, serverFriendsErr := r.getDesignatedFriends(ctx, []string{userID})
	if serverFriendsErr == nil {
		for _, sf := range serverFriends {
			local := ServerFriendToLocalFriend(sf)
			if local != nil && local.FriendUserID == userID {
				log.ZDebug(ctx, "lintao syncCallRecordUserProfile server-friend", "userID", userID)
				r.applyCallRecordUserProfile(ctx, userID, local.ConversationShowName(), local.FaceURL, "server-friend")
				return
			}
		}
		// 服务端好友已删除但本地仍残留时，按注销处理，避免 server-user 旧资料覆盖。
		if r.hasLocalFriend(ctx, userID) {
			log.ZDebug(ctx, "lintao syncCallRecordUserProfile former-friend", "userID", userID)
			r.syncCallRecordUserProfileForRemovedFriend(ctx, userID)
			return
		}
	}

	if serverUser, err := r.user.GetSingleUserFromServer(ctx, userID); err == nil && serverUser != nil {
		showName := serverUser.DisplayName()
		faceURL := serverUser.FaceURL
		if constant.IsDeactivatedUserNickname(showName) || faceURL == constant.DeactivatedUserFaceURL {
			r.applyCallRecordUserProfile(ctx, userID, r.deactivatedUserNickname(ctx), faceURL, "server-user-deactivated")
			return
		}
		r.applyCallRecordUserProfile(ctx, userID, showName, faceURL, "server-user")
		return
	} else if err != nil {
		if sdkerrs.ErrUserIDNotFound.Is(errs.Unwrap(err)) {
			r.syncCallRecordUserProfileForRemovedFriend(ctx, userID)
			return
		}
		log.ZWarn(ctx, "lintao syncCallRecordUserProfile GetSingleUserFromServer failed", err,
			"loginUserID", r.loginUserID, "friendUserID", userID)
	}

	if serverFriendsErr == nil {
		r.syncCallRecordUserProfileForRemovedFriend(ctx, userID)
		return
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
	r.applyCallRecordUserProfile(ctx, userID, r.deactivatedUserNickname(ctx), constant.DeactivatedUserFaceURL, "deactivated")
}

func (r *Relation) deactivatedUserNickname(ctx context.Context) string {
	if r.db == nil {
		return constant.DeactivatedUserNicknameForLanguage("")
	}
	user, err := r.db.GetLoginUser(ctx, r.loginUserID)
	if err != nil || user == nil {
		return constant.DeactivatedUserNicknameForLanguage("")
	}
	return constant.DeactivatedUserNicknameForLanguage(user.Language)
}

func (r *Relation) hasLocalFriend(ctx context.Context, userID string) bool {
	if r.db == nil {
		return false
	}
	friends, err := r.db.GetFriendInfoList(ctx, []string{userID})
	if err != nil {
		return false
	}
	for _, f := range friends {
		if f != nil && f.FriendUserID == userID {
			return true
		}
	}
	return false
}

func (r *Relation) syncCallRecordLoginUserProfile(ctx context.Context) {
	user, err := r.db.GetLoginUser(ctx, r.loginUserID)
	if err != nil || user == nil {
		return
	}
	r.applyCallRecordUserProfile(ctx, r.loginUserID, user.DisplayName(), user.FaceURL, "login-user")
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
