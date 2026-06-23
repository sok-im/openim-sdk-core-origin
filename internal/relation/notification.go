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
			log.ZInfo(ctx, "lintao FriendDeletedNotification received",
				"loginUserID", r.loginUserID,
				"fromUserID", tips.FromToUserID.FromUserID,
				"toUserID", tips.FromToUserID.ToUserID)
			if tips.FromToUserID.ToUserID == r.loginUserID && tips.FromToUserID.FromUserID != r.loginUserID {
				log.ZInfo(ctx, "lintao FriendDeletedNotification SyncUserInfo for deleted friend profile",
					"loginUserID", r.loginUserID, "changedUserID", tips.FromToUserID.FromUserID)
				if err := r.user.SyncUserInfo(ctx, tips.FromToUserID.FromUserID); err != nil {
					log.ZWarn(ctx, "lintao FriendDeletedNotification SyncUserInfo failed", err, "userID", tips.FromToUserID.FromUserID)
				}
			}
			if tips.FromToUserID.FromUserID == r.loginUserID && tips.FromToUserID.ToUserID != r.loginUserID {
				log.ZInfo(ctx, "lintao FriendDeletedNotification evict UserCache",
					"loginUserID", r.loginUserID, "friendUserID", tips.FromToUserID.ToUserID)
				r.user.UserCache.Delete(tips.FromToUserID.ToUserID)
			}
			if tips.FromToUserID.FromUserID == r.loginUserID || tips.FromToUserID.ToUserID == r.loginUserID {
				if err := r.IncrSyncFriends(ctx); err != nil {
					log.ZWarn(ctx, "lintao FriendDeletedNotification IncrSyncFriends failed", err,
						"loginUserID", r.loginUserID, "fromUserID", tips.FromToUserID.FromUserID,
						"toUserID", tips.FromToUserID.ToUserID)
					return err
				}
				log.ZInfo(ctx, "lintao FriendDeletedNotification IncrSyncFriends ok",
					"loginUserID", r.loginUserID, "fromUserID", tips.FromToUserID.FromUserID,
					"toUserID", tips.FromToUserID.ToUserID)
				return nil
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
		log.ZInfo(ctx, "lintao FriendInfoUpdatedNotification received",
			"loginUserID", r.loginUserID, "changedUserID", tips.UserID)
		if tips.UserID != r.loginUserID {
			if err := r.user.SyncUserInfo(ctx, tips.UserID); err != nil {
				log.ZWarn(ctx, "lintao FriendInfoUpdatedNotification SyncUserInfo failed", err, "userID", tips.UserID)
			} else {
				log.ZInfo(ctx, "lintao FriendInfoUpdatedNotification SyncUserInfo ok", "userID", tips.UserID)
			}
			if err := r.IncrSyncFriends(ctx); err != nil {
				log.ZWarn(ctx, "lintao FriendInfoUpdatedNotification IncrSyncFriends failed", err, "userID", tips.UserID)
				return err
			}
			r.syncCallRecordsUserProfile(ctx, tips.UserID)
			log.ZInfo(ctx, "lintao FriendInfoUpdatedNotification IncrSyncFriends ok", "userID", tips.UserID)
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
		log.ZInfo(ctx, "lintao FriendsInfoUpdateNotification received",
			"loginUserID", r.loginUserID, "toUserID", tips.FromToUserID.GetToUserID(), "friendIDs", tips.FriendIDs)
		if tips.FromToUserID.ToUserID == r.loginUserID {
			for _, friendID := range tips.FriendIDs {
				_, hadCache := r.user.UserCache.Load(friendID)
				r.user.UserCache.Delete(friendID)
				log.ZInfo(ctx, "lintao FriendsInfoUpdateNotification evict UserCache",
					"loginUserID", r.loginUserID, "friendUserID", friendID, "hadCache", hadCache)
			}
			if err := r.IncrSyncFriends(ctx); err != nil {
				log.ZWarn(ctx, "lintao FriendsInfoUpdateNotification IncrSyncFriends failed", err,
					"loginUserID", r.loginUserID, "friendIDs", tips.FriendIDs)
				return err
			}
			r.syncCallRecordsUserProfile(ctx, tips.FriendIDs...)
			log.ZInfo(ctx, "lintao FriendsInfoUpdateNotification IncrSyncFriends ok",
				"loginUserID", r.loginUserID, "friendIDs", tips.FriendIDs)
			return nil
		}
	default:
		return fmt.Errorf("type failed %d", msg.ContentType)
	}
	return nil
}
