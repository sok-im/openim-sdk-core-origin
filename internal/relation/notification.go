package relation

import (
	"context"
	"fmt"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/log"
)

func (r *Relation) DoNotification(ctx context.Context, msg *sdkws.MsgData) {
	if err := r.doNotification(ctx, msg); err != nil {
		log.ZError(ctx, "doNotification error", err, "msg", msg)
	}
}

func (r *Relation) doNotification(ctx context.Context, msg *sdkws.MsgData) error {
	r.relationSyncMutex.Lock()
	defer r.relationSyncMutex.Unlock()

	switch msg.ContentType {
	case constant.FriendApplicationNotification:
		tips := sdkws.FriendApplicationTips{}
		if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
			return err
		}
		r.friendshipListener.OnFriendApplicationAdded(*ServerFriendRequestToLocalFriendRequest(tips.Request))
	case constant.FriendApplicationApprovedNotification:
		var tips sdkws.FriendApplicationApprovedTips
		err := utils.UnmarshalNotificationElem(msg.Content, &tips)
		if err != nil {
			return err
		}
		if tips.Request != nil {
			r.friendshipListener.OnFriendApplicationAccepted(*ServerFriendRequestToLocalFriendRequest(tips.Request))
		}
		return r.IncrSyncFriends(ctx)
	case constant.FriendApplicationRejectedNotification:
		var tips sdkws.FriendApplicationRejectedTips
		if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
			return err
		}
		r.friendshipListener.OnFriendApplicationRejected(*ServerFriendRequestToLocalFriendRequest(tips.Request))
	case constant.FriendAddedNotification:
		var tips sdkws.FriendAddedTips
		if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
			return err
		}
		if tips.Friend != nil && tips.Friend.FriendUser != nil {
			if tips.Friend.FriendUser.UserID == r.loginUserID {
				return r.IncrSyncFriends(ctx)
			} else if tips.Friend.OwnerUserID == r.loginUserID {
				return r.IncrSyncFriends(ctx)
			}
		}
	case constant.FriendDeletedNotification:
		var tips sdkws.FriendDeletedTips
		if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
			return err
		}
		if tips.FromToUserID != nil {
			// When the login user removed a friend (FromUserID == self), the friend's
			// UserCache entry must be evicted immediately. IncrSyncFriends only updates
			// the friend DB; it does not touch UserCache, leaving stale data that
			// GetUsersInfo would serve even after the friend's account is deleted.
			if tips.FromToUserID.FromUserID == r.loginUserID && tips.FromToUserID.ToUserID != r.loginUserID {
				r.user.UserCache.Delete(tips.FromToUserID.ToUserID)
				// 账号注销等场景下好友关系被删除时，先刷新通话记录展示，再同步好友列表。
				log.ZDebug(ctx, "lintao FriendDeletedNotification syncCallRecordsUserProfileForRemovedFriend", "userID", tips.FromToUserID.ToUserID)
				r.syncCallRecordsUserProfileForRemovedFriend(ctx, tips.FromToUserID.ToUserID)
			}
			if tips.FromToUserID.FromUserID == r.loginUserID {
				return r.IncrSyncFriends(ctx)
			}
		}
	case constant.FriendRemarkSetNotification:
		var tips sdkws.FriendInfoChangedTips
		if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
			return err
		}
		if tips.FromToUserID != nil {
			if tips.FromToUserID.FromUserID == r.loginUserID {
				return r.IncrSyncFriends(ctx)
			}
		}
	case constant.FriendInfoUpdatedNotification:
		var tips sdkws.UserInfoUpdatedTips
		if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
			return err
		}
		if tips.UserID != r.loginUserID {
			// Actively sync the changed user's profile from the server.
			// This covers the one-way case (A added B but B didn't add A): the server
			// does NOT bump B's friend-list version, so IncrSyncFriends returns nothing
			// and the friendSyncer never calls UserCache.Delete — leaving stale data.
			// SyncUserInfo fetches fresh data, overwrites the cache, and fires
			// conversation/message update events when name or face URL changed.
			if err := r.user.SyncUserInfo(ctx, tips.UserID); err != nil {
				log.ZWarn(ctx, "FriendInfoUpdatedNotification SyncUserInfo failed", err, "userID", tips.UserID)
			}
			if err := r.IncrSyncFriends(ctx); err != nil {
				return err
			}
			log.ZDebug(ctx, "lintao FriendInfoUpdatedNotification syncCallRecordsUserProfile", "userID", tips.UserID)
			r.syncCallRecordsUserProfile(ctx, tips.UserID)
			return nil
		}
	case constant.BlackAddedNotification:
		var tips sdkws.BlackAddedTips
		if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
			return err
		}
		if tips.FromToUserID.FromUserID == r.loginUserID {
			return r.SyncAllBlackList(ctx)
		}
	case constant.BlackDeletedNotification:
		var tips sdkws.BlackDeletedTips
		if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
			return err
		}
		if tips.FromToUserID.FromUserID == r.loginUserID {
			return r.SyncAllBlackList(ctx)
		}
	case constant.FriendsInfoUpdateNotification:
		var tips sdkws.FriendsInfoUpdateTips
		if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
			return err
		}
		if tips.FromToUserID.ToUserID == r.loginUserID {
			// Eagerly evict the named friend IDs from UserCache so that the next
			// GetUsersInfo call fetches fresh data from the server rather than
			// returning stale cached data. This is critical for the account-deletion
			// case: DeleteFriend (step 3b) sends this notification before the user
			// record is hard-deleted (step 6), so without eviction there is a window
			// where GetUsersInfo still returns the deleted user's profile.
			for _, friendID := range tips.FriendIDs {
				r.user.UserCache.Delete(friendID)
			}
			log.ZDebug(ctx, "lintao FriendsInfoUpdateNotification syncCallRecordsUserProfileForRemovedFriend delete userCache", "friendIDs", tips.FriendIDs)
			r.syncCallRecordsUserProfileForRemovedFriend(ctx, tips.FriendIDs...)
			return r.IncrSyncFriends(ctx)
		}
	default:
		return fmt.Errorf("type failed %d", msg.ContentType)
	}
	return nil
}
