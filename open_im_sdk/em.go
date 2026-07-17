// Copyright © 2023 OpenIM SDK. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package open_im_sdk

import (
	"context"
	"errors"

	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
	"github.com/openimsdk/tools/log"
)

var ErrNotImplemented = errors.New("not set listener")

type emptyGroupListener struct {
	ctx context.Context
}

func newEmptyGroupListener(ctx context.Context) open_im_sdk_callback.OnGroupListener {
	return &emptyGroupListener{ctx}
}

func (e *emptyGroupListener) OnJoinedGroupAdded(groupInfo string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupInfo", groupInfo)
}

func (e *emptyGroupListener) OnJoinedGroupDeleted(groupInfo string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupInfo", groupInfo)

}

func (e *emptyGroupListener) OnGroupMemberAdded(groupMemberInfo string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupMemberInfo", groupMemberInfo)

}

func (e *emptyGroupListener) OnGroupMemberDeleted(groupMemberInfo string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupMemberInfo", groupMemberInfo)

}

func (e *emptyGroupListener) OnGroupApplicationAdded(groupApplication string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupApplication", groupApplication)

}

func (e *emptyGroupListener) OnGroupApplicationDeleted(groupApplication string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupApplication", groupApplication)

}

func (e *emptyGroupListener) OnGroupInfoChanged(groupInfo string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupInfo", groupInfo)

}

func (e *emptyGroupListener) OnGroupDismissed(groupInfo string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupInfo", groupInfo)
}

func (e *emptyGroupListener) OnGroupMemberInfoChanged(groupMemberInfo string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupMemberInfo", groupMemberInfo)
}

func (e *emptyGroupListener) OnGroupApplicationAccepted(groupApplication string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupApplication", groupApplication)
}

func (e *emptyGroupListener) OnGroupApplicationRejected(groupApplication string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupApplication", groupApplication)
}

func (e *emptyGroupListener) OnGroupMessagePinned(groupMessagePinnedTips string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupMessagePinnedTips", groupMessagePinnedTips)
}

func (e *emptyGroupListener) OnGroupCallStarted(groupCallStartedTips string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupCallStartedTips", groupCallStartedTips)
}

func (e *emptyGroupListener) OnGroupCallEnded(groupCallEndedTips string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupCallEndedTips", groupCallEndedTips)
}

func (e *emptyGroupListener) OnGroupCallParticipantCountUpdated(groupCallParticipantCountUpdatedTips string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupCallParticipantCountUpdatedTips", groupCallParticipantCountUpdatedTips)
}

func (e *emptyGroupListener) OnGroupCallParticipantDeclined(groupCallParticipantDeclinedTips string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupCallParticipantDeclinedTips", groupCallParticipantDeclinedTips)
}

func (e *emptyGroupListener) OnGroupBurnDurationSet(groupBurnDurationSetTips string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupBurnDurationSetTips", groupBurnDurationSetTips)
}

func (e *emptyGroupListener) OnGroupFaceURLSet(groupFaceURLSetTips string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupFaceURLSetTips", groupFaceURLSetTips)
}

func (e *emptyGroupListener) OnGroupNeedVerificationSet(groupNeedVerificationSetTips string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupNeedVerificationSetTips", groupNeedVerificationSetTips)
}

func (e *emptyGroupListener) OnGroupPermissionChanged(groupPermissionChangedTips string) {
	log.ZWarn(e.ctx, "GroupListener is not implemented", nil, "groupPermissionChangedTips", groupPermissionChangedTips)
}

type emptyFriendshipListener struct {
	ctx context.Context
}

func newEmptyFriendshipListener(ctx context.Context) open_im_sdk_callback.OnFriendshipListener {
	return &emptyFriendshipListener{ctx}
}

func (e *emptyFriendshipListener) OnFriendApplicationAdded(friendApplication string) {
	log.ZWarn(e.ctx, "FriendshipListener is not implemented", nil,
		"friendApplication", friendApplication)
}

func (e *emptyFriendshipListener) OnFriendApplicationDeleted(friendApplication string) {
	log.ZWarn(e.ctx, "FriendshipListener is not implemented", nil,
		"friendApplication", friendApplication)
}

func (e *emptyFriendshipListener) OnFriendApplicationAccepted(friendApplication string) {
	log.ZWarn(e.ctx, "FriendshipListener is not implemented", nil,
		"friendApplication", friendApplication)
}

func (e *emptyFriendshipListener) OnFriendApplicationRejected(friendApplication string) {
	log.ZWarn(e.ctx, "FriendshipListener is not implemented", nil,
		"friendApplication", friendApplication)
}

func (e *emptyFriendshipListener) OnFriendAdded(friendInfo string) {
	log.ZWarn(e.ctx, "FriendshipListener is not implemented", nil,
		"friendInfo", friendInfo)
}

func (e *emptyFriendshipListener) OnFriendDeleted(friendInfo string) {
	log.ZWarn(e.ctx, "FriendshipListener is not implemented", nil,
		"friendInfo", friendInfo)
}

func (e *emptyFriendshipListener) OnFriendInfoChanged(friendInfo string) {
	log.ZWarn(e.ctx, "FriendshipListener is not implemented", nil,
		"friendInfo", friendInfo)
}

func (e *emptyFriendshipListener) OnBlackAdded(blackInfo string) {
	log.ZWarn(e.ctx, "FriendshipListener is not implemented", nil,
		"blackInfo", blackInfo)
}

func (e *emptyFriendshipListener) OnBlackDeleted(blackInfo string) {
	log.ZWarn(e.ctx, "FriendshipListener is not implemented", nil,
		"blackInfo", blackInfo)
}

type emptyConversationListener struct {
	ctx context.Context
}

func newEmptyConversationListener(ctx context.Context) open_im_sdk_callback.OnConversationListener {
	return &emptyConversationListener{ctx: ctx}
}

func (e *emptyConversationListener) OnSyncServerStart(reinstalled bool) {
	log.ZWarn(e.ctx, "ConversationListener is not implemented", nil)
}

func (e *emptyConversationListener) OnSyncServerProgress(progress int) {
	log.ZWarn(e.ctx, "ConversationListener is not implemented", nil,
		"progress", progress)
}

func (e *emptyConversationListener) OnSyncServerFinish(reinstalled bool) {
	log.ZWarn(e.ctx, "ConversationListener is not implemented", nil)

}

func (e *emptyConversationListener) OnSyncServerFailed(reinstalled bool) {

	log.ZWarn(e.ctx, "ConversationListener is not implemented", nil)
}

func (e *emptyConversationListener) OnNewConversation(conversationList string) {
	log.ZWarn(e.ctx, "ConversationListener is not implemented", nil,
		"conversationList", conversationList)

}

func (e *emptyConversationListener) OnConversationChanged(conversationList string) {

	log.ZWarn(e.ctx, "ConversationListener is not implemented", nil,
		"conversationList", conversationList)
}

func (e *emptyConversationListener) OnTotalUnreadMessageCountChanged(totalUnreadCount int32) {
	log.ZWarn(e.ctx, "ConversationListener is not implemented", nil,
		"totalUnreadCount", totalUnreadCount)
}

func (e *emptyConversationListener) OnConversationUserInputStatusChanged(change string) {

}

type emptyAdvancedMsgListener struct {
	ctx context.Context
}

func newEmptyAdvancedMsgListener(ctx context.Context) open_im_sdk_callback.OnAdvancedMsgListener {
	return &emptyAdvancedMsgListener{ctx}
}

func (e *emptyAdvancedMsgListener) OnRecvOnlineOnlyMessage(message string) {

}

func (e *emptyAdvancedMsgListener) OnRecvNewMessage(message string) {
	log.ZWarn(e.ctx, "AdvancedMsgListener is not implemented", nil, "message", message)
}

func (e *emptyAdvancedMsgListener) OnRecvC2CReadReceipt(msgReceiptList string) {

	log.ZWarn(e.ctx, "AdvancedMsgListener is not implemented", nil,
		"msgReceiptList", msgReceiptList)
}

func (e *emptyAdvancedMsgListener) OnRecvGroupReadReceipt(groupMsgReceiptList string) {
	log.ZWarn(e.ctx, "AdvancedMsgListener is not implemented", nil,
		"groupMsgReceiptList", groupMsgReceiptList)
}

func (e *emptyAdvancedMsgListener) OnNewRecvMessageRevoked(messageRevoked string) {
	log.ZWarn(e.ctx, "AdvancedMsgListener is not implemented", nil, "messageRevoked", messageRevoked)
}

func (e *emptyAdvancedMsgListener) OnReactionUpdated(reactionUpdated string) {
	log.ZWarn(e.ctx, "AdvancedMsgListener is not implemented", nil, "reactionUpdated", reactionUpdated)
}

func (e *emptyAdvancedMsgListener) OnRecvMessageExtensionsChanged(msgID string, reactionExtensionList string) {
	log.ZWarn(e.ctx, "AdvancedMsgListener is not implemented", nil, "msgID", msgID,
		"reactionExtensionList", reactionExtensionList)
}

func (e *emptyAdvancedMsgListener) OnRecvMessageExtensionsDeleted(msgID string, reactionExtensionKeyList string) {
	log.ZWarn(e.ctx, "AdvancedMsgListener is not implemented", nil, "msgID", msgID,
		"reactionExtensionKeyList", reactionExtensionKeyList)
}

func (e *emptyAdvancedMsgListener) OnRecvMessageExtensionsAdded(msgID string, reactionExtensionList string) {
	log.ZWarn(e.ctx, "AdvancedMsgListener is not implemented", nil, "msgID", msgID,
		"reactionExtensionList", reactionExtensionList)
}

func (e *emptyAdvancedMsgListener) OnRecvOfflineNewMessage(message string) {
	log.ZWarn(e.ctx, "AdvancedMsgListener is not implemented", nil, "message", message)
}

func (e *emptyAdvancedMsgListener) OnMsgDeleted(message string) {
	log.ZWarn(e.ctx, "AdvancedMsgListener is not implemented", nil, "message", message)
}

type emptyBatchMsgListener struct{}

func newEmptyBatchMsgListener() *emptyBatchMsgListener {
	return &emptyBatchMsgListener{}
}

func (e *emptyBatchMsgListener) OnRecvNewMessages(messageList string) {

}

func (e *emptyBatchMsgListener) OnRecvOfflineNewMessages(messageList string) {

}

type emptyUserListener struct {
	ctx context.Context
}

func (e *emptyUserListener) OnUserCommandAdd(userCommand string) {
	log.ZWarn(e.ctx, "UserListener is not implemented", nil, "userCommand", userCommand)
}

func (e *emptyUserListener) OnUserCommandDelete(userCommand string) {
	log.ZWarn(e.ctx, "UserListener is not implemented", nil, "userCommand", userCommand)
}

func (e *emptyUserListener) OnUserCommandUpdate(userCommand string) {
	log.ZWarn(e.ctx, "UserListener is not implemented", nil, "userCommand", userCommand)
}

func newEmptyUserListener(ctx context.Context) open_im_sdk_callback.OnUserListener {
	return &emptyUserListener{ctx: ctx}
}

func (e *emptyUserListener) OnSelfInfoUpdated(userInfo string) {
	log.ZWarn(e.ctx, "UserListener is not implemented", nil, "userInfo", userInfo)
}

func (e *emptyUserListener) OnUserStatusChanged(statusMap string) {
	log.ZWarn(e.ctx, "UserListener is not implemented", nil, "statusMap", statusMap)
}

type emptyCustomBusinessListener struct {
	ctx context.Context
}

func newEmptyCustomBusinessListener(ctx context.Context) open_im_sdk_callback.OnCustomBusinessListener {
	return &emptyCustomBusinessListener{ctx: ctx}
}

func (e *emptyCustomBusinessListener) OnRecvCustomBusinessMessage(businessMessage string) {
	log.ZWarn(e.ctx, "CustomBusinessListener is not implemented", nil,
		"businessMessage", businessMessage)

}

type emptySignalingListener struct {
	ctx context.Context
}

func newEmptySignalingListener(ctx context.Context) open_im_sdk_callback.OnSignalingListener {
	return &emptySignalingListener{ctx: ctx}
}

func (e *emptySignalingListener) OnReceiveNewInvitation(receiveNewInvitationCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnReceiveNewInvitation")
}

func (e *emptySignalingListener) OnInviteeAccepted(inviteeAcceptedCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnInviteeAccepted")
}

func (e *emptySignalingListener) OnInviteeAcceptedByOtherDevice(inviteeAcceptedCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnInviteeAcceptedByOtherDevice")
}

func (e *emptySignalingListener) OnInviteeRejected(inviteeRejectedCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnInviteeRejected")
}

func (e *emptySignalingListener) OnInviteeRejectedByOtherDevice(inviteeRejectedCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnInviteeRejectedByOtherDevice")
}

func (e *emptySignalingListener) OnInvitationCancelled(invitationCancelledCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnInvitationCancelled")
}

func (e *emptySignalingListener) OnInvitationTimeout(invitationTimeoutCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnInvitationTimeout")
}

func (e *emptySignalingListener) OnHangUp(hangUpCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnHangUp")
}

func (e *emptySignalingListener) OnRoomParticipantConnected(onRoomParticipantConnectedCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnRoomParticipantConnected")
}

func (e *emptySignalingListener) OnRoomParticipantDisconnected(onRoomParticipantDisconnectedCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnRoomParticipantDisconnected")
}

func (e *emptySignalingListener) OnGroupCallStatusChanged(groupCallStatusCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnGroupCallStatusChanged")
}

func (e *emptySignalingListener) OnReceiveCustomSignal(customSignalCallback string) {
	log.ZWarn(e.ctx, "SignalingListener is not implemented", nil, "callback", "OnReceiveCustomSignal")
}
