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

package constant

import "strings"

const (
	CmdSyncData           = "syncData"
	CmdSyncFlag           = "syncFlag"
	CmdNotification       = "notification"
	CmdMsgSyncInReinstall = "msgSyncInReinstall"
	CmdNewMsgCome         = "newMsgCome"
	CmdUpdateConversation = "updateConversation"
	CmdUpdateMessage      = "updateMessage"

	CmdPushMsg        = "pushMsg"
	CmdConnSuccesss   = "connSuccess"
	CmdWakeUpDataSync = "wakeUpDataSync"
	CmdIMMessageSync  = "imMessageSync"
	CmdLogOut         = "loginOut"
)

const (
	//ContentType
	Text                            = 101
	Picture                         = 102
	Sound                           = 103
	Video                           = 104
	File                            = 105
	AtText                          = 106
	Merger                          = 107
	Card                            = 108
	Location                        = 109
	Custom                          = 110
	Typing                          = 113
	Quote                           = 114
	Face                            = 115
	AdvancedText                    = 117
	CustomMsgNotTriggerConversation = 119
	CustomMsgOnlineOnly             = 120

	NotificationBegin = 1000

	FriendNotificationBegin = 1200

	FriendApplicationApprovedNotification = 1201 //add_friend_response
	FriendApplicationRejectedNotification = 1202 //add_friend_response
	FriendApplicationNotification         = 1203 //add_friend
	FriendAddedNotification               = 1204
	FriendDeletedNotification             = 1205 //delete_friend
	FriendRemarkSetNotification           = 1206 //set_friend_remark?
	BlackAddedNotification                = 1207 //add_black
	BlackDeletedNotification              = 1208 //remove_black
	FriendInfoUpdatedNotification         = 1209
	FriendsInfoUpdateNotification         = 1210
	FriendNotificationEnd                 = 1299
	ConversationChangeNotification        = 1300

	UserNotificationBegin         = 1301
	UserInfoUpdatedNotification   = 1303 //SetSelfInfoTip             = 204
	UserStatusChangeNotification  = 1304
	UserCommandAddNotification    = 1305
	UserCommandDeleteNotification = 1306
	UserCommandUpdateNotification = 1307

	UserNotificationEnd = 1399
	OANotification      = 1400

	GroupNotificationBegin = 1500

	GroupCreatedNotification                 = 1501
	GroupInfoSetNotification                 = 1502
	JoinGroupApplicationNotification         = 1503
	MemberQuitNotification                   = 1504
	GroupApplicationAcceptedNotification     = 1505
	GroupApplicationRejectedNotification     = 1506
	GroupOwnerTransferredNotification        = 1507
	MemberKickedNotification                 = 1508
	MemberInvitedNotification                = 1509
	MemberEnterNotification                  = 1510
	GroupDismissedNotification               = 1511
	GroupMemberMutedNotification             = 1512
	GroupMemberCancelMutedNotification       = 1513
	GroupMutedNotification                   = 1514
	GroupCancelMutedNotification             = 1515
	GroupMemberInfoSetNotification           = 1516
	GroupMemberSetToAdminNotification        = 1517
	GroupMemberSetToOrdinaryUserNotification = 1518
	GroupInfoSetAnnouncementNotification     = 1519
	GroupInfoSetNameNotification             = 1520
	GroupMessagePinnedNotification           = 1521
	GroupCallStartedNotification                    = 1522
	GroupCallEndedNotification                      = 1523
	GroupCallParticipantCountUpdatedNotification    = 1527
	GroupCallParticipantDeclinedNotification        = 1528
	GroupBurnDurationSetNotification                = 1524
	GroupFaceURLSetNotification              = 1525
	GroupNeedVerificationSetNotification     = 1526
	GroupNotificationEnd                     = 1599

	ConversationPrivateChatNotification = 1701
	ClearConversationNotification       = 1703

	// 钱包动作通知（红包/转账）
	RedPacketClaimNotification   = 1801
	TransferReceiveNotification  = 1802
	RedPacketExpiredNotification = 1803
	TransferExpiredNotification  = 1804

	BusinessNotification = 2001
	ServiceNotification  = 2002
	PaymentNotification  = 2003

	RevokeNotification = 2101

	DeleteMsgsNotification = 2102

	HasReadReceipt = 2200

	NotificationEnd = 5000
	////////////////////////////////////////

	//MsgFrom
	UserMsgType = 100
	SysMsgType  = 200

	/////////////////////////////////////
	//SessionType
	SingleChatType = 1
	// WriteGroupChatType Not enabled temporarily
	WriteGroupChatType   = 2
	ReadGroupChatType    = 3
	NotificationChatType = 4

	MsgStatusSending     = 1
	MsgStatusSendSuccess = 2
	MsgStatusSendFailed  = 3
	MsgStatusHasDeleted  = 4
	MsgStatusFiltered    = 5

	//OptionsKey
	IsHistory                  = "history"
	IsPersistent               = "persistent"
	IsUnreadCount              = "unreadCount"
	IsConversationUpdate       = "conversationUpdate"
	IsOfflinePush              = "offlinePush"
	IsSenderSync               = "senderSync"
	IsNotPrivate               = "notPrivate"
	IsSenderConversationUpdate = "senderConversationUpdate"

	//GroupStatus
	GroupOk              = 0
	GroupBanChat         = 1
	GroupStatusDismissed = 2
	GroupStatusMuted     = 3
)

const (
	BlackRelationship  = 0
	FriendRelationship = 1
)

const (
	NormalGroup                       = 0
	SuperGroup                        = 1
	WorkingGroup                      = 2
	SuperGroupErrChatLogsTableNamePre = "local_sg_err_chat_logs_"
	ChatLogsTableNamePre              = "chat_logs_"
)

const (
	AddConOrUpLatMsg                      = 1
	TotalUnreadMessageChanged             = 2
	UpdateConFaceUrlAndNickName           = 3
	UpdateLatestMessageReadState          = 4
	UpdateLatestMessageFaceUrlAndNickName = 5
	ConChange                             = 6
	NewCon                                = 7
	ConChangeDirect                       = 8
	NewConDirect                          = 9
	UpdateMsgFaceUrlAndNickName           = 10

	HasRead = 1
	NotRead = 0
)

const (
	GetNewestSeq          = 1001
	PullMsgByRange        = 1002
	SendMsg               = 1003
	SendSignalMsg         = 1004
	PullMsgBySeqList      = 1005
	GetConvMaxReadSeq     = 1006
	PullConvLastMessage   = 1007
	PushMsg               = 2001
	KickOnlineMsg         = 2002
	LogoutMsg             = 2003
	SetBackgroundStatus   = 2004
	WsSubUserOnlineStatus = 2005
)

// conversation
const (
	//MsgReceiveOpt
	ReceiveMessage = 0
	// NotReceiveMessage This option is currently disabled
	NotReceiveMessage       = 1
	ReceiveNotNotifyMessage = 2

	Online  = 1
	Offline = 0
)

const (
	GroupOwner         = 100 // Group member type: owner
	GroupAdmin         = 60  // Group member type: administrator
	GroupOrdinaryUsers = 20  // Group member type: ordinary user

	GroupFilterAll                   = 0
	GroupFilterOwner                 = 1
	GroupFilterAdmin                 = 2
	GroupFilterOrdinaryUsers         = 3
	GroupFilterAdminAndOrdinaryUsers = 4
	GroupFilterOwnerAndAdmin         = 5

	GroupResponseAgree  = 1  // Response to group application: agree
	GroupResponseRefuse = -1 // Response to group application: refuse

	FriendResponseAgree   = 1  // Response to friend request: agree
	FriendResponseRefuse  = -1 // Response to friend request: refuse
	FriendResponseDefault = 0
)
const (
	AtAllString = "AtAllTag" // String for 'all people' mention tag
	AtNormal    = 0          // Mention mode: normal
	AtMe        = 1          // Mention mode: mention sender only
	AtAll       = 2          // Mention mode: mention all people
	AtAllAtMe   = 3          // Mention mode: mention all people and sender

)

const (
	KeywordMatchOr  = 0 // Keyword match mode: match any keyword
	KeywordMatchAnd = 1 // Keyword match mode: match all keywords
)

const BigVersion = "v3"

// SignalCallStatus 本地音视频通话记录状态（与 rtc.CallRecordItem.status 对齐）。
const (
	SignalCallStatusUnknown      int32 = 0 // 查询：不筛选
	SignalCallStatusAnswered     int32 = 1 // 已接听
	SignalCallStatusNotConnected int32 = 2 // 未接通
)

// SignalCallDirection 本地通话记录：通话方向（主叫 / 被叫 / 未接来电）。
const (
	SignalCallDirectionUnknown  int32 = 0 // 查询：不筛选
	SignalCallDirectionOutgoing int32 = 1 // 主叫（本端发起邀请）
	SignalCallDirectionIncoming int32 = 2 // 被叫（已接听，通话已建立）
	SignalCallDirectionMissed   int32 = 3 // 被叫未接（对方取消、超时或本端被拒）
)

// SignalCallRole 本端在通话中的角色（主叫 / 被叫）。
const (
	SignalCallRoleUnknown  int32 = 0 // 查询：不筛选
	SignalCallRoleOutgoing int32 = 1 // 主叫
	SignalCallRoleIncoming int32 = 2 // 被叫
)

// DeactivatedUserNickname / DeactivatedUserFaceURL 为已注销账号在本地历史记录中的占位展示。
const (
	DeactivatedUserNickname   = "Deactivated user"
	DeactivatedUserNicknameZh = "已注销的用户"
	DeactivatedUserFaceURL    = "http://13.215.203.29:10002/object/6794065114/mmexport1782219627453.jpg"
)

// DeactivatedUserNicknameForLanguage 按用户语言返回已注销账号展示昵称，默认英文。
func DeactivatedUserNicknameForLanguage(language string) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	if strings.HasPrefix(lang, "zh") {
		return DeactivatedUserNicknameZh
	}
	return DeactivatedUserNickname
}

// IsDeactivatedUserNickname 判断昵称是否为已注销账号占位（含中英文）。
func IsDeactivatedUserNickname(name string) bool {
	name = strings.TrimSpace(name)
	return name == DeactivatedUserNickname || name == DeactivatedUserNicknameZh
}

// SignalCallAction 本地通话记录：触发落库的信令动作。
const (
	SignalCallActionAccept  = "accept"
	SignalCallActionReject  = "reject"
	SignalCallActionCancel  = "cancel"
	SignalCallActionHungUp  = "hungup"
	SignalCallActionTimeout = "timeout"
)

// SignalCallRoleFromDirection 由 direction 推导 role（incoming/missed 均视为被叫）。
func SignalCallRoleFromDirection(direction int32) int32 {
	switch direction {
	case SignalCallDirectionOutgoing:
		return SignalCallRoleOutgoing
	case SignalCallDirectionIncoming, SignalCallDirectionMissed:
		return SignalCallRoleIncoming
	default:
		return SignalCallRoleUnknown
	}
}

const (
	MsgSyncBegin      = 1001 //
	MsgSyncProcessing = 1002 //
	MsgSyncEnd        = 1003 //
	MsgSyncFailed     = 1004
	AppDataSyncStart  = 1005
	AppDataSyncFinish = 1006
)

const (
	SplitPullMsgNum            = 100
	PullMsgNumForReadDiffusion = 50
)

const (
	Uninitialized = -1001
)
