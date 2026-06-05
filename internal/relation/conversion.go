package relation

import (
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/protocol/sdkws"
)

func ServerFriendRequestToLocalFriendRequest(info *sdkws.FriendRequest) *model_struct.LocalFriendRequest {
	return &model_struct.LocalFriendRequest{
		FromUserID:    info.FromUserID,
		FromNickname:  info.FromNickname,
		FromFaceURL:   info.FromFaceURL,
		ToUserID:      info.ToUserID,
		ToNickname:    info.ToNickname,
		ToFaceURL:     info.ToFaceURL,
		HandleResult:  info.HandleResult,
		ReqMsg:        info.ReqMsg,
		CreateTime:    info.CreateTime,
		HandlerUserID: info.HandlerUserID,
		HandleMsg:     info.HandleMsg,
		HandleTime:    info.HandleTime,
		Ex:            info.Ex,
	}
}

func ServerFriendToLocalFriend(info *sdkws.FriendInfo) *model_struct.LocalFriend {
	local := &model_struct.LocalFriend{
		OwnerUserID:    info.OwnerUserID,
		Remark:         info.Remark,
		CreateTime:     info.CreateTime,
		AddSource:      info.AddSource,
		OperatorUserID: info.OperatorUserID,
		Ex:             info.Ex,
		IsPinned:       info.IsPinned,
	}
	if info.FriendUser != nil {
		local.FriendUserID = info.FriendUser.UserID
		local.Nickname = info.FriendUser.Nickname
		local.FirstName = info.FriendUser.FirstName
		local.LastName = info.FriendUser.LastName
		local.FaceURL = info.FriendUser.FaceURL
	}
	return local
}

func ServerBlackToLocalBlack(info *sdkws.BlackInfo) *model_struct.LocalBlack {
	local := &model_struct.LocalBlack{
		OwnerUserID:    info.OwnerUserID,
		CreateTime:     info.CreateTime,
		AddSource:      info.AddSource,
		OperatorUserID: info.OperatorUserID,
		Ex:             info.Ex,
	}
	if info.BlackUserInfo != nil {
		local.BlockUserID = info.BlackUserInfo.UserID
		local.Nickname = info.BlackUserInfo.Nickname
		local.FaceURL = info.BlackUserInfo.FaceURL
	}
	return local
}

func isValidServerBlack(info *sdkws.BlackInfo) bool {
	return info != nil && info.BlackUserInfo != nil && info.BlackUserInfo.UserID != ""
}
