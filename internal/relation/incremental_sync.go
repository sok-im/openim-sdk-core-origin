package relation

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/syncer"
	"github.com/openimsdk/protocol/relation"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/datautil"
)

func (r *Relation) IncrSyncFriends(ctx context.Context) error {
	log.ZDebug(ctx, "lintao IncrSyncFriends", "loginUserID", r.loginUserID)
	friendSyncer := syncer.VersionSynchronizer[*model_struct.LocalFriend, *relation.GetIncrementalFriendsResp]{
		Ctx:       ctx,
		DB:        r.db,
		TableName: r.friendListTableName(),
		EntityID:  r.loginUserID,
		Key: func(localFriend *model_struct.LocalFriend) string {
			return localFriend.FriendUserID
		},
		Local: func() ([]*model_struct.LocalFriend, error) {
			return r.db.GetAllFriendList(ctx)
		},
		Server: func(version *model_struct.LocalVersionSync) (*relation.GetIncrementalFriendsResp, error) {
			localVersion := uint64(0)
			localVersionID := ""
			if version != nil {
				localVersion = version.Version
				localVersionID = version.VersionID
			}
			log.ZInfo(ctx, "lintao IncrSyncFriends request",
				"loginUserID", r.loginUserID,
				"localVersion", localVersion,
				"localVersionID", localVersionID)
			resp, err := r.getIncrementalFriends(ctx, &relation.GetIncrementalFriendsReq{
				UserID:    r.loginUserID,
				Version:   localVersion,
				VersionID: localVersionID,
			})
			if err != nil {
				log.ZError(ctx, "lintao IncrSyncFriends request failed", err,
					"loginUserID", r.loginUserID)
				return nil, err
			}
			updateUserIDs := make([]string, 0, len(resp.Update))
			updateNicknames := make(map[string]string, len(resp.Update))
			for _, item := range resp.Update {
				if item.FriendUser == nil {
					continue
				}
				updateUserIDs = append(updateUserIDs, item.FriendUser.UserID)
				updateNicknames[item.FriendUser.UserID] = item.FriendUser.Nickname
			}
			insertUserIDs := make([]string, 0, len(resp.Insert))
			for _, item := range resp.Insert {
				if item.FriendUser == nil {
					continue
				}
				insertUserIDs = append(insertUserIDs, item.FriendUser.UserID)
			}
			log.ZInfo(ctx, "lintao IncrSyncFriends response",
				"loginUserID", r.loginUserID,
				"serverVersion", resp.Version,
				"serverVersionID", resp.VersionID,
				"full", resp.Full,
				"deleteCount", len(resp.Delete),
				"deleteUserIDs", resp.Delete,
				"updateCount", len(resp.Update),
				"updateUserIDs", updateUserIDs,
				"updateNicknames", updateNicknames,
				"insertCount", len(resp.Insert),
				"insertUserIDs", insertUserIDs)
			return resp, nil
		},
		Full: func(resp *relation.GetIncrementalFriendsResp) bool {
			return resp.Full
		},
		Version: func(resp *relation.GetIncrementalFriendsResp) (string, uint64) {
			return resp.VersionID, resp.Version
		},
		Delete: func(resp *relation.GetIncrementalFriendsResp) []string {
			return resp.Delete
		},
		Update: func(resp *relation.GetIncrementalFriendsResp) []*model_struct.LocalFriend {
			return datautil.Batch(ServerFriendToLocalFriend, resp.Update)
		},
		Insert: func(resp *relation.GetIncrementalFriendsResp) []*model_struct.LocalFriend {
			return datautil.Batch(ServerFriendToLocalFriend, resp.Insert)
		},
		Syncer: func(server, local []*model_struct.LocalFriend) error {
			return r.friendSyncer.Sync(ctx, server, local, nil)
		},
		FullSyncer: func(ctx context.Context) error {
			return r.friendSyncer.FullSync(ctx, r.loginUserID)
		},
		FullID: func(ctx context.Context) ([]string, error) {
			resp, err := r.getFullFriendUserIDs(ctx, &relation.GetFullFriendUserIDsReq{
				UserID: r.loginUserID,
			})
			if err != nil {
				return nil, err
			}
			return resp.UserIDs, nil
		},
		IDOrderChanged: func(resp *relation.GetIncrementalFriendsResp) bool {
			return resp.SortVersion > 0
		},
	}
	return friendSyncer.IncrementalSync()
}

func (r *Relation) friendListTableName() string {
	return model_struct.LocalFriend{}.TableName()
}
func (r *Relation) IncrSyncFriendsWithLock(ctx context.Context) error {
	r.relationSyncMutex.Lock()
	defer r.relationSyncMutex.Unlock()
	return r.IncrSyncFriends(ctx)
}
