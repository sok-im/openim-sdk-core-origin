package relation

import (
	"context"
	"sort"
	"strings"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
	"github.com/openimsdk/protocol/relation"
	"github.com/openimsdk/tools/utils/datautil"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/common"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/datafetcher"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	sdk "github.com/openimsdk/openim-sdk-core/v3/pkg/sdk_params_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdkerrs"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/server_api_params"

	"github.com/openimsdk/tools/log"
)

func (r *Relation) GetSpecifiedFriendsInfo(ctx context.Context, friendUserIDList []string, filterBlack bool) ([]*model_struct.LocalFriend, error) {
	dataFetcher := datafetcher.NewDataFetcher(
		r.db,
		r.friendListTableName(),
		r.loginUserID,
		func(localFriend *model_struct.LocalFriend) string {
			return localFriend.FriendUserID
		},
		func(ctx context.Context, values []*model_struct.LocalFriend) error {
			return r.db.BatchInsertFriend(ctx, values)
		},
		func(ctx context.Context, userIDs []string) ([]*model_struct.LocalFriend, bool, error) {
			localFriends, err := r.db.GetFriendInfoList(ctx, userIDs)
			return localFriends, true, err
		},
		func(ctx context.Context, userIDs []string) ([]*model_struct.LocalFriend, error) {
			serverFriend, err := r.getDesignatedFriends(ctx, userIDs)
			if err != nil {
				return nil, err
			}
			return datautil.Batch(ServerFriendToLocalFriend, serverFriend), nil
		},
	)

	localFriendList, err := dataFetcher.FetchMissingAndFillLocal(ctx, friendUserIDList)
	if err != nil {
		log.ZWarn(ctx, "GetDesignatedFriendsInfo", err)
		return nil, err
	}

	log.ZDebug(ctx, "GetDesignatedFriendsInfo", "localFriendList", localFriendList)

	if !filterBlack {
		log.ZDebug(ctx, "GetDesignatedFriendsInfo", "localFriendList", localFriendList)
		return localFriendList, nil
	}
	log.ZDebug(ctx, "GetDesignatedFriendsInfo", "localFriendList", localFriendList)
	blackList, err := r.db.GetBlackInfoList(ctx, friendUserIDList)
	if err != nil {
		return nil, err
	}
	if len(blackList) == 0 {
		return localFriendList, nil
	}

	log.ZDebug(ctx, "GetDesignatedFriendsInfo", "blackList", blackList)
	m := datautil.SliceSetAny(blackList, func(e *model_struct.LocalBlack) string {
		return e.BlockUserID
	})
	var res []*model_struct.LocalFriend
	for _, localFriend := range localFriendList {
		if _, ok := m[localFriend.FriendUserID]; !ok {
			res = append(res, localFriend)
		}
	}
	return res, nil
}

func (r *Relation) AddFriend(ctx context.Context, req *relation.ApplyToAddFriendReq) error {
	return r.AddOnewayFriend(ctx, req)
}

// AddOnewayFriend adds toUserID to the caller's friend list without requiring consent.
// Only the caller's friend list is updated; the target user's list remains unchanged.
// After the incremental friend sync it synchronously updates the conversation show_name
// so that GetConversationListSplit reflects the remark immediately.
func (r *Relation) AddOnewayFriend(ctx context.Context, req *relation.ApplyToAddFriendReq) error {
	if err := r.addOnewayFriend(ctx, req); err != nil {
		log.ZWarn(ctx, "AddOnewayFriend failed", err, "req", req)
		return err
	}

	r.relationSyncMutex.Lock()
	if err := r.IncrSyncFriends(ctx); err != nil {
		r.relationSyncMutex.Unlock()
		log.ZWarn(ctx, "IncrSyncFriends failed", err, "req", req)
		return err
	}
	r.relationSyncMutex.Unlock()

	// Fallback: if IncrSyncFriends returned early (server incremental version not yet
	// updated atomically), the new friend may not be in local DB. Fetch it directly
	// from the server and insert it so that GetSpecifiedFriendsInfo is never empty.
	if err := r.ensureFriendInLocalDB(ctx, req.ToUserID); err != nil {
		log.ZWarn(ctx, "ensureFriendInLocalDB failed", err, "toUserID", req.ToUserID)
	}

	r.syncConversationShowNameForFriend(ctx, req.ToUserID, req.Remark)
	return nil
}

// ensureFriendInLocalDB checks whether toUserID is already present in the local
// friend DB (written by IncrSyncFriends). If not, it fetches the friend record
// directly from the server and inserts it, preventing GetSpecifiedFriendsInfo
// from returning empty when the server's incremental version lags behind.
func (r *Relation) ensureFriendInLocalDB(ctx context.Context, toUserID string) error {
	existing, err := r.db.GetFriendInfoList(ctx, []string{toUserID})
	if err != nil {
		log.ZWarn(ctx, "ensureFriendInLocalDB failed", err, "toUserID", toUserID)
		return err
	}
	if len(existing) > 0 {
		log.ZInfo(ctx, "ensureFriendInLocalDB success", "toUserID", toUserID)
		return nil
	}
	serverFriends, err := r.getDesignatedFriends(ctx, []string{toUserID})
	if err != nil {
		log.ZWarn(ctx, "ensureFriendInLocalDB failed", err, "toUserID", toUserID)
		return err
	}
	for _, sf := range serverFriends {
		local := ServerFriendToLocalFriend(sf)
		if err := r.db.InsertFriend(ctx, local); err != nil {
			log.ZWarn(ctx, "ensureFriendInLocalDB failed", err, "toUserID", toUserID)
			return err
		}
	}
	log.ZInfo(ctx, "ensureFriendInLocalDB success", "toUserID", toUserID)
	return nil
}

// syncConversationShowNameForFriend directly writes show_name into the local
// conversation DB for the single chat between the caller and friendUserID,
// then fires ConChange so the conversation list listener is notified immediately.
//
// This is called synchronously after IncrSyncFriends so the update is visible
// to GetConversationListSplit without waiting for the async channel pipeline
// (TriggerCmdUpdateConversation → UpdateConFaceUrlAndNickName).
//
// If no conversation row exists yet the DB update is a no-op (RowsAffected==0)
// and ConChange is not fired; the async pipeline will handle it when the first
// message is sent.
func (r *Relation) syncConversationShowNameForFriend(ctx context.Context, friendUserID, remark string) {
	log.ZInfo(ctx, "lintao syncConversationShowNameForFriend", "friendUserID", friendUserID, "remark", remark)
	showName := remark
	if showName == "" {
		friend, err := r.db.GetFriendInfoByFriendUserID(ctx, friendUserID)
		if err != nil || friend == nil {
			log.ZWarn(ctx, "lintao syncConversationShowNameForFriend failed", err, "friendUserID", friendUserID)
			return
		}
		showName = friend.ConversationShowName()
	}
	if showName == "" {
		log.ZInfo(ctx, "lintao syncConversationShowNameForFriend failed", "friendUserID", friendUserID)
		return
	}
	ids := []string{r.loginUserID, friendUserID}
	sort.Strings(ids)
	conversationID := "si_" + strings.Join(ids, "_")
	if err := r.db.UpdateConversation(ctx, &model_struct.LocalConversation{
		ConversationID:   conversationID,
		ConversationType: constant.SingleChatType,
		ShowName:         showName,
	}); err != nil {
		// RowsAffected==0 means no conversation row exists yet; skip silently.
		log.ZWarn(ctx, "lintao syncConversationShowNameForFriend update skipped", err,
			"conversationID", conversationID, "friendUserID", friendUserID)
		return
	}
	log.ZInfo(ctx, "lintao syncConversationShowNameForFriend success", "friendUserID", friendUserID, "remark", remark)
	// Notify the conversation list listener so the UI refreshes immediately.
	_ = common.TriggerCmdUpdateConversation(ctx, common.UpdateConNode{
		ConID:  conversationID,
		Action: constant.ConChange,
		Args:   []string{conversationID},
	}, r.conversationCh)
	log.ZInfo(ctx, "lintao syncConversationShowNameForFriend trigger conversation change", "friendUserID", friendUserID, "remark", remark)
}

func (r *Relation) GetFriendApplicationListAsRecipient(ctx context.Context, req *sdk.GetFriendApplicationListAsRecipientReq) ([]*model_struct.LocalFriendRequest, error) {
	friendRequests, err := r.getRecvFriendApplicationList(ctx, req.HandleResults, utils.GetPageNumber(req.Offset, req.Count), req.Count)
	if err != nil {
		return nil, err
	}
	return datautil.Batch(ServerFriendRequestToLocalFriendRequest, friendRequests), nil
}

func (r *Relation) GetFriendApplicationListAsApplicant(ctx context.Context, req *sdk.GetFriendApplicationListAsApplicantReq) ([]*model_struct.LocalFriendRequest, error) {
	friendRequests, err := r.getSelfFriendApplicationList(ctx, utils.GetPageNumber(req.Offset, req.Count), req.Count)
	if err != nil {
		return nil, err
	}
	return datautil.Batch(ServerFriendRequestToLocalFriendRequest, friendRequests), nil
}

func (r *Relation) AcceptFriendApplication(ctx context.Context, userIDHandleMsg *sdk.ProcessFriendApplicationParams) error {
	return r.RespondFriendApply(ctx, &relation.RespondFriendApplyReq{FromUserID: userIDHandleMsg.ToUserID, ToUserID: r.loginUserID, HandleResult: constant.FriendResponseAgree, HandleMsg: userIDHandleMsg.HandleMsg})
}

func (r *Relation) RefuseFriendApplication(ctx context.Context, userIDHandleMsg *sdk.ProcessFriendApplicationParams) error {
	return r.RespondFriendApply(ctx, &relation.RespondFriendApplyReq{FromUserID: userIDHandleMsg.ToUserID, ToUserID: r.loginUserID, HandleResult: constant.FriendResponseRefuse, HandleMsg: userIDHandleMsg.HandleMsg})
}

func (r *Relation) RespondFriendApply(ctx context.Context, req *relation.RespondFriendApplyReq) error {
	if err := r.addFriendResponse(ctx, req); err != nil {
		return err
	}
	r.relationSyncMutex.Lock()
	defer r.relationSyncMutex.Unlock()

	if req.HandleResult == constant.FriendResponseAgree {
		_ = r.IncrSyncFriends(ctx)
	}
	return nil
}

func (r *Relation) CheckFriend(ctx context.Context, friendUserIDList []string) ([]*server_api_params.UserIDResult, error) {
	friendList, err := r.db.GetFriendInfoList(ctx, friendUserIDList)
	if err != nil {
		return nil, err
	}
	blackList, err := r.db.GetBlackInfoList(ctx, friendUserIDList)
	if err != nil {
		return nil, err
	}
	res := make([]*server_api_params.UserIDResult, 0, len(friendUserIDList))
	for _, v := range friendUserIDList {
		var r server_api_params.UserIDResult
		isBlack := false
		isFriend := false
		for _, b := range blackList {
			if v == b.BlockUserID {
				isBlack = true
				break
			}
		}
		for _, r := range friendList {
			if v == r.FriendUserID {
				isFriend = true
				break
			}
		}
		r.UserID = v
		if isFriend && !isBlack {
			r.Result = 1
		} else {
			r.Result = 0
		}
		res = append(res, &r)
	}
	return res, nil
}

func (r *Relation) DeleteFriend(ctx context.Context, friendUserID string) error {
	if err := r.deleteFriend(ctx, friendUserID); err != nil {
		return err
	}

	r.relationSyncMutex.Lock()
	defer r.relationSyncMutex.Unlock()

	return r.IncrSyncFriends(ctx)
}

// DeleteFriendOneway calls /friend/delete_friend_oneway on the server.
// Only the caller's friend row is removed; the peer keeps the caller in their list.
// After the server call it runs IncrSyncFriends so the local DB reflects the
// removal immediately without waiting for an async notification.
func (r *Relation) DeleteFriendOneway(ctx context.Context, friendUserID string) error {
	if err := r.deleteFriendOneway(ctx, friendUserID); err != nil {
		return err
	}

	r.relationSyncMutex.Lock()
	defer r.relationSyncMutex.Unlock()

	return r.IncrSyncFriends(ctx)
}

func (r *Relation) GetFriendList(ctx context.Context, filterBlack bool) ([]*model_struct.LocalFriend, error) {
	localFriendList, err := r.db.GetAllFriendList(ctx)
	if err != nil {
		return nil, err
	}
	if len(localFriendList) == 0 || !filterBlack {
		return localFriendList, nil
	}
	localBlackList, err := r.db.GetBlackListDB(ctx)
	if err != nil {
		return nil, err
	}
	if len(localBlackList) == 0 {
		return localFriendList, nil
	}
	blackSet := datautil.SliceSetAny(localBlackList, func(e *model_struct.LocalBlack) string {
		return e.BlockUserID
	})
	var res []*model_struct.LocalFriend
	for _, friend := range localFriendList {
		if _, ok := blackSet[friend.FriendUserID]; !ok {
			res = append(res, friend)
		}
	}
	return res, nil
}

func (r *Relation) GetFriendListPage(ctx context.Context, offset, count int32, filterBlack bool) ([]*model_struct.LocalFriend, error) {
	dataFetcher := datafetcher.NewDataFetcher(
		r.db,
		r.friendListTableName(),
		r.loginUserID,
		func(localFriend *model_struct.LocalFriend) string {
			return localFriend.FriendUserID
		},
		func(ctx context.Context, values []*model_struct.LocalFriend) error {
			return r.db.BatchInsertFriend(ctx, values)
		},
		func(ctx context.Context, userIDs []string) ([]*model_struct.LocalFriend, bool, error) {
			localFriendList, err := r.db.GetFriendInfoList(ctx, userIDs)
			return localFriendList, true, err
		},
		func(ctx context.Context, userIDs []string) ([]*model_struct.LocalFriend, error) {
			serverFriend, err := r.getDesignatedFriends(ctx, userIDs)
			if err != nil {
				return nil, err
			}
			return datautil.Batch(ServerFriendToLocalFriend, serverFriend), nil
		},
	)
	localBlackList, err := r.db.GetBlackListDB(ctx)
	if err != nil {
		return nil, err
	}
	if (!filterBlack) || len(localBlackList) == 0 {
		return dataFetcher.FetchWithPagination(ctx, int(offset), int(count))
	}
	localFriendList, err := dataFetcher.FetchWithPagination(ctx, int(offset), int(count*2))
	if err != nil {
		return nil, err
	}
	blackUserIDs := datautil.SliceSetAny(localBlackList, func(e *model_struct.LocalBlack) string {
		return e.BlockUserID
	})
	res := localFriendList[:0]
	for _, friend := range localFriendList {
		if _, ok := blackUserIDs[friend.FriendUserID]; !ok {
			res = append(res, friend)
		}
		if len(res) == int(count) {
			break
		}
	}
	return res, nil
}

func (r *Relation) SearchFriends(ctx context.Context, param *sdk.SearchFriendsParam) ([]*sdk.SearchFriendItem, error) {
	if len(param.KeywordList) == 0 || (!param.IsSearchNickname && !param.IsSearchUserID && !param.IsSearchRemark) {
		return nil, sdkerrs.ErrArgs.WrapMsg("keyword is null or search field all false")
	}
	localFriendList, err := r.db.SearchFriendList(ctx, param.KeywordList[0], param.IsSearchUserID, param.IsSearchNickname, param.IsSearchRemark)
	if err != nil {
		return nil, err
	}
	localBlackList, err := r.db.GetBlackListDB(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]struct{})
	for _, black := range localBlackList {
		m[black.BlockUserID] = struct{}{}
	}
	res := make([]*sdk.SearchFriendItem, 0, len(localFriendList))
	for i, localFriend := range localFriendList {
		var relationship int
		if _, ok := m[localFriend.FriendUserID]; ok {
			relationship = constant.BlackRelationship
		} else {
			relationship = constant.FriendRelationship
		}
		res = append(res, &sdk.SearchFriendItem{
			LocalFriend:  *localFriendList[i],
			Relationship: relationship,
		})
	}
	return res, nil
}

func (r *Relation) AddBlack(ctx context.Context, blackUserID string, ex string) error {
	if err := r.addBlack(ctx, &relation.AddBlackReq{BlackUserID: blackUserID, Ex: ex}); err != nil {
		return err
	}
	return r.SyncAllBlackList(ctx)
}

func (r *Relation) RemoveBlack(ctx context.Context, blackUserID string) error {
	if err := r.removeBlack(ctx, blackUserID); err != nil {
		return err
	}

	r.relationSyncMutex.Lock()
	defer r.relationSyncMutex.Unlock()

	return r.SyncAllBlackList(ctx)
}

func (r *Relation) GetBlackList(ctx context.Context) ([]*model_struct.LocalBlack, error) {
	return r.db.GetBlackListDB(ctx)
}

func (r *Relation) UpdateFriends(ctx context.Context, req *relation.UpdateFriendsReq) error {
	req.OwnerUserID = r.loginUserID
	if err := r.updateFriends(ctx, req); err != nil {
		return err
	}

	r.relationSyncMutex.Lock()
	defer r.relationSyncMutex.Unlock()

	return r.IncrSyncFriends(ctx)
}

func (r *Relation) GetFriendApplicationUnhandledCount(ctx context.Context, req *sdk.GetSelfUnhandledApplyCountReq) (int32, error) {
	return r.getSelfUnhandledApplyCount(ctx, req.Time)
}

func (r *Relation) SetFriendMute(ctx context.Context, req *sdk.SetFriendMuteReq) error {
	if err := r.setFriendMute(ctx, &relation.SetMuteReq{
		OwnerUserID:  r.loginUserID,
		TargetUserID: req.TargetUserID,
		Duration:     req.Duration,
	}); err != nil {
		return err
	}
	r.relationSyncMutex.Lock()
	defer r.relationSyncMutex.Unlock()
	return r.IncrSyncFriends(ctx)
}

func (r *Relation) GetFriendMute(ctx context.Context, req *sdk.GetFriendMuteReq) (*relation.GetMuteResp, error) {
	return r.getFriendMute(ctx, &relation.GetMuteReq{
		OwnerUserID:  r.loginUserID,
		TargetUserID: req.TargetUserID,
	})
}

func (r *Relation) PinFriend(ctx context.Context, req *sdk.PinFriendReq) error {
	if err := r.pinFriend(ctx, &relation.PinFriendReq{
		OwnerUserID:  r.loginUserID,
		FriendUserID: req.FriendUserID,
	}); err != nil {
		return err
	}
	r.relationSyncMutex.Lock()
	defer r.relationSyncMutex.Unlock()
	return r.IncrSyncFriends(ctx)
}

func (r *Relation) UnpinFriend(ctx context.Context, req *sdk.UnpinFriendReq) error {
	if err := r.unpinFriend(ctx, &relation.UnpinFriendReq{
		OwnerUserID:  r.loginUserID,
		FriendUserID: req.FriendUserID,
	}); err != nil {
		return err
	}
	r.relationSyncMutex.Lock()
	defer r.relationSyncMutex.Unlock()
	return r.IncrSyncFriends(ctx)
}
