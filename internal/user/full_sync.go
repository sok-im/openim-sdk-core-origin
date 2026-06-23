package user

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/common"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdkerrs"
	userPb "github.com/openimsdk/protocol/user"
	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/datautil"
)

func (u *User) SyncLoginUserInfo(ctx context.Context) error {
	remoteUser, err := u.GetSingleUserFromServer(ctx, u.loginUserID)
	if err != nil {
		return err
	}
	localUser, err := u.GetLoginUser(ctx, u.loginUserID)
	if err != nil && (!errs.ErrRecordNotFound.Is(errs.Unwrap(err))) {
		return err
	}
	var localUsers []*model_struct.LocalUser
	if err == nil {
		localUsers = []*model_struct.LocalUser{localUser}
	}
	log.ZDebug(ctx, "SyncLoginUserInfo", "remoteUser", remoteUser, "localUser", localUser)
	return u.userSyncer.Sync(ctx, []*model_struct.LocalUser{remoteUser}, localUsers, nil)
}

func (u *User) SyncLoginUserInfoWithoutNotice(ctx context.Context) error {
	remoteUser, err := u.GetSingleUserFromServer(ctx, u.loginUserID)
	if err != nil {
		return err
	}
	localUser, err := u.GetLoginUser(ctx, u.loginUserID)
	if err != nil && (!errs.ErrRecordNotFound.Is(errs.Unwrap(err))) {
		return err
	}
	var localUsers []*model_struct.LocalUser
	if err == nil {
		localUsers = []*model_struct.LocalUser{localUser}
	}
	log.ZDebug(ctx, "SyncLoginUserInfo", "remoteUser", remoteUser, "localUser", localUser)
	return u.userSyncer.Sync(ctx, []*model_struct.LocalUser{remoteUser}, localUsers, nil, false, true)
}

// SyncUserInfo fetches the latest profile for userID from the server, stores it
// in the in-memory UserCache, and fires conversation / message update events
// when the display name or face URL has changed.
//
// It is intended for non-login users (e.g. a one-way friend whose profile
// changed but whose friend-list version was not bumped on B's side).
func (u *User) SyncUserInfo(ctx context.Context, userID string) error {
	newUser, err := u.GetSingleUserFromServer(ctx, userID)
	if err != nil {
		if sdkerrs.ErrUserIDNotFound.Is(errs.Unwrap(err)) {
			u.UserCache.Delete(userID)
			log.ZInfo(ctx, "SyncUserInfo user not found, removed from cache", "userID", userID)
		} else {
			log.ZWarn(ctx, "SyncUserInfo GetSingleUserFromServer failed", err, "userID", userID)
		}
		return err
	}

	oldUser, hasOld := u.UserCache.Load(userID)
	u.UserCache.Store(userID, newUser)

	newShowName := newUser.DisplayName()
	changed := !hasOld ||
		oldUser.FaceURL != newUser.FaceURL ||
		oldUser.Nickname != newUser.Nickname ||
		oldUser.FirstName != newUser.FirstName ||
		oldUser.LastName != newUser.LastName

	if changed {
		log.ZInfo(ctx, "SyncUserInfo profile changed, triggering conversation update",
			"userID", userID, "faceURL", newUser.FaceURL, "showName", newShowName)
		_ = common.TriggerCmdUpdateConversation(ctx, common.UpdateConNode{
			Action: constant.UpdateConFaceUrlAndNickName,
			Args: common.SourceIDAndSessionType{
				SourceID:    newUser.UserID,
				SessionType: constant.SingleChatType,
				FaceURL:     newUser.FaceURL,
				Nickname:    newShowName,
			},
		}, u.conversationCh)
		_ = common.TriggerCmdUpdateMessage(ctx, common.UpdateMessageNode{
			Action: constant.UpdateMsgFaceUrlAndNickName,
			Args: common.UpdateMessageInfo{
				SessionType: constant.SingleChatType,
				UserID:      newUser.UserID,
				FaceURL:     newUser.FaceURL,
				Nickname:    newUser.Nickname,
			},
		}, u.conversationCh)
	}
	return nil
}

func (u *User) SyncAllCommand(ctx context.Context) error {
	return u.syncAllCommand(ctx, true)
}

func (u *User) SyncAllCommandWithoutNotice(ctx context.Context) error {
	return u.syncAllCommand(ctx, false)
}

func (u *User) syncAllCommand(ctx context.Context, withNotice bool) error {
	resp, err := u.processUserCommandGetAll(ctx, &userPb.ProcessUserCommandGetAllReq{UserID: u.loginUserID})
	if err != nil {
		return err
	}
	localData, err := u.DataBase.ProcessUserCommandGetAll(ctx)
	if err != nil {
		return err
	}
	log.ZDebug(ctx, "sync command", "data from server", resp, "data from local", localData)
	if withNotice {
		return u.commandSyncer.Sync(ctx, datautil.Batch(ServerCommandToLocalCommand, resp.CommandResp), localData, nil)
	} else {
		return u.commandSyncer.Sync(ctx, datautil.Batch(ServerCommandToLocalCommand, resp.CommandResp), localData, nil, false, true)
	}
}
