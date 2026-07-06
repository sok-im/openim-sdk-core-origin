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
			fromID := tips.FromToUserID.FromUserID
			toID := tips.FromToUserID.ToUserID
			// When the login user removed a friend (FromUserID == self), the friend's
			// UserCache entry must be evicted immediately. IncrSyncFriends only updates
			// the friend DB; it does not touch UserCache, leaving stale data that
			// GetUsersInfo would serve even after the friend's account is deleted.
			if fromID == r.loginUserID && toID != r.loginUserID {
				r.user.UserCache.Delete(toID)
				log.ZDebug(ctx, "FriendDeletedNotification syncCallRecordsUserProfileForRemovedFriend", "userID", toID)
				r.syncCallRecordsUserProfileForRemovedFriend(ctx, toID)
				return r.IncrSyncFriends(ctx)
			}
			// When the peer removed the friendship (e.g. account deletion), refresh
			// call-record display. Do not IncrSyncFriends: local friend row may still
			// exist until server-side reversal cleanup completes.
			if toID == r.loginUserID && fromID != r.loginUserID {
				r.user.UserCache.Delete(fromID)
				log.ZDebug(ctx, "FriendDeletedNotification peer removed syncCallRecordsUserProfileForRemovedFriend", "userID", fromID)
				r.syncCallRecordsUserProfileForRemovedFriend(ctx, fromID)
				return nil
			}
		}
	case constant.FriendRemarkSetNotification:
		var tips sdkws.FriendInfoChangedTips
		if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
			return err
		}
		log.ZDebug(ctx, "FriendRemarkSetNotification", "tips", tips, "loginUserID", r.loginUserID)	
		if tips.FromToUserID != nil {
			if tips.FromToUserID.FromUserID == r.loginUserID {
				friendUserID := tips.FromToUserID.ToUserID
				if err := r.IncrSyncFriends(ctx); err != nil {
					return err
				}
				// Remark/friendFirstName/friendLastName changes must refresh the single-chat
				// show_name immediately. IncrSyncFriends alone may skip the syncer Update
				// notice when local data already matches (e.g. cross-device race), leaving
				// the conversation list stale until the next message arrives.
				r.syncConversationShowNameForFriend(ctx, friendUserID, "")
				return nil
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
			log.ZDebug(ctx, "FriendInfoUpdatedNotification syncCallRecordsUserProfile", "userID", tips.UserID)
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
			log.ZDebug(ctx, "FriendsInfoUpdateNotification syncCallRecordsUserProfileForRemovedFriend delete userCache", "friendIDs", tips.FriendIDs)
			r.syncCallRecordsUserProfileForRemovedFriend(ctx, tips.FriendIDs...)
			return r.IncrSyncFriends(ctx)
		}
	default:
		return fmt.Errorf("type failed %d", msg.ContentType)
	}
	return nil
}
