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

package model_struct

import (
	"database/sql/driver"
	"encoding/json"
	"strings"

	"github.com/openimsdk/tools/errs"
)

type LocalFriend struct {
	OwnerUserID    string `gorm:"column:owner_user_id;primary_key;type:varchar(64)" json:"ownerUserID"`
	FriendUserID   string `gorm:"column:friend_user_id;primary_key;type:varchar(64)" json:"userID"`
	Remark         string `gorm:"column:remark;type:varchar(255)" json:"remark"`
	CreateTime     int64  `gorm:"column:create_time" json:"createTime"`
	AddSource      int32  `gorm:"column:add_source" json:"addSource"`
	OperatorUserID string `gorm:"column:operator_user_id;type:varchar(64)" json:"operatorUserID"`
	Nickname       string `gorm:"column:name;type:varchar;type:varchar(255)" json:"nickname"`
	FirstName      string `gorm:"column:first_name;type:varchar(255)" json:"firstName"`
	LastName       string `gorm:"column:last_name;type:varchar(255)" json:"lastName"`
	FaceURL        string `gorm:"column:face_url;type:varchar;type:varchar(255)" json:"faceURL"`
	Ex             string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
	AttachedInfo   string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
	IsPinned       bool   `gorm:"column:is_pinned;" json:"isPinned"`
}

// UserDisplayName returns remark > firstName+lastName (trimmed) > nickname for non-friend users.
func UserDisplayName(firstName, lastName, nickname string) string {
	if firstName != "" || lastName != "" {
		return strings.TrimSpace(firstName + " " + lastName)
	}
	return nickname
}

// DisplayName returns firstName+lastName (trimmed) > nickname for the login user.
func (u *LocalUser) DisplayName() string {
	if u == nil {
		return ""
	}
	return UserDisplayName(u.FirstName, u.LastName, u.Nickname)
}

// ConversationShowName is the single-chat session list title for this friend:
// remark, else firstName+lastName (trimmed), else nickname.
func (f *LocalFriend) ConversationShowName() string {
	if f == nil {
		return ""
	}
	if f.Remark != "" {
		return f.Remark
	}
	return UserDisplayName(f.FirstName, f.LastName, f.Nickname)
}

func (LocalFriend) TableName() string {
	return "local_friends"
}

type LocalFriendRequest struct {
	FromUserID   string `gorm:"column:from_user_id;primary_key;type:varchar(64)" json:"fromUserID"`
	FromNickname string `gorm:"column:from_nickname;type:varchar;type:varchar(255)" json:"fromNickname"`
	FromFaceURL  string `gorm:"column:from_face_url;type:varchar;type:varchar(255)" json:"fromFaceURL"`
	// FromGender   int32  `gorm:"column:from_gender" json:"fromGender"`

	ToUserID   string `gorm:"column:to_user_id;primary_key;type:varchar(64)" json:"toUserID"`
	ToNickname string `gorm:"column:to_nickname;type:varchar;type:varchar(255)" json:"toNickname"`
	ToFaceURL  string `gorm:"column:to_face_url;type:varchar;type:varchar(255)" json:"toFaceURL"`
	// ToGender   int32  `gorm:"column:to_gender" json:"toGender"`

	HandleResult  int32  `gorm:"column:handle_result" json:"handleResult"`
	ReqMsg        string `gorm:"column:req_msg;type:varchar(255)" json:"reqMsg"`
	CreateTime    int64  `gorm:"column:create_time" json:"createTime"`
	HandlerUserID string `gorm:"column:handler_user_id;type:varchar(64)" json:"handlerUserID"`
	HandleMsg     string `gorm:"column:handle_msg;type:varchar(255)" json:"handleMsg"`
	HandleTime    int64  `gorm:"column:handle_time" json:"handleTime"`
	Ex            string `gorm:"column:ex;type:varchar(1024)" json:"ex"`

	AttachedInfo string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
}

type LocalGroup struct {
	GroupID                string `gorm:"column:group_id;primary_key;type:varchar(64)" json:"groupID" binding:"required"`
	GroupName              string `gorm:"column:name;size:255" json:"groupName"`
	Notification           string `gorm:"column:notification;type:varchar(255)" json:"notification"`
	Introduction           string `gorm:"column:introduction;type:varchar(255)" json:"introduction"`
	FaceURL                string `gorm:"column:face_url;type:varchar(255)" json:"faceURL"`
	CreateTime             int64  `gorm:"column:create_time" json:"createTime"`
	Status                 int32  `gorm:"column:status" json:"status"`
	CreatorUserID          string `gorm:"column:creator_user_id;type:varchar(64)" json:"creatorUserID"`
	GroupType              int32  `gorm:"column:group_type" json:"groupType"`
	OwnerUserID            string `gorm:"column:owner_user_id;type:varchar(64)" json:"ownerUserID"`
	MemberCount            int32  `gorm:"column:member_count" json:"memberCount"`
	Ex                     string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
	AttachedInfo           string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
	NeedVerification       int32  `gorm:"column:need_verification"  json:"needVerification"`
	LookMemberInfo         int32  `gorm:"column:look_member_info" json:"lookMemberInfo"`
	ApplyMemberFriend      int32  `gorm:"column:apply_member_friend" json:"applyMemberFriend"`
	NotificationUpdateTime int64  `gorm:"column:notification_update_time" json:"notificationUpdateTime"`
	NotificationUserID     string `gorm:"column:notification_user_id;size:64" json:"notificationUserID"`
	AllowSendMsg           int32  `gorm:"column:allow_send_msg" json:"allowSendMsg"`
	AllowPinMsg            int32  `gorm:"column:allow_pin_msg" json:"allowPinMsg"`
	AllowAddMember         int32  `gorm:"column:allow_add_member" json:"allowAddMember"`
	AllowEditGroupInfo     int32  `gorm:"column:allow_edit_group_info" json:"allowEditGroupInfo"`
	MsgBurnDuration        int32  `gorm:"column:msg_burn_duration" json:"msgBurnDuration"`
	AllowBurn              int32  `gorm:"column:allow_burn" json:"allowBurn"`
	EnableInviteLink       int32  `gorm:"column:enable_invite_link" json:"enableInviteLink"`
}

func (LocalGroup) TableName() string {
	return "local_groups"
}

type LocalGroupMember struct {
	GroupID        string `gorm:"column:group_id;primary_key;type:varchar(64)" json:"groupID"`
	UserID         string `gorm:"column:user_id;primary_key;type:varchar(64)" json:"userID"`
	Nickname       string `gorm:"column:nickname;type:varchar(255)" json:"nickname"`
	FaceURL        string `gorm:"column:user_group_face_url;type:varchar(255)" json:"faceURL"`
	RoleLevel      int32  `gorm:"column:role_level;index:index_role_level;" json:"roleLevel"`
	JoinTime       int64  `gorm:"column:join_time;index:index_join_time;" json:"joinTime"`
	JoinSource     int32  `gorm:"column:join_source" json:"joinSource"`
	InviterUserID  string `gorm:"column:inviter_user_id;size:64"  json:"inviterUserID"`
	MuteEndTime    int64  `gorm:"column:mute_end_time;default:0" json:"muteEndTime"`
	OperatorUserID string `gorm:"column:operator_user_id;type:varchar(64)" json:"operatorUserID"`
	Ex             string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
	AttachedInfo   string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
}

func (LocalGroupMember) TableName() string {
	return "local_group_members"
}

type LocalGroupRequest struct {
	GroupID       string `gorm:"column:group_id;primary_key;type:varchar(64)" json:"groupID"`
	GroupName     string `gorm:"column:group_name;size:255" json:"groupName"`
	Notification  string `gorm:"column:notification;type:varchar(255)" json:"notification"`
	Introduction  string `gorm:"column:introduction;type:varchar(255)" json:"introduction"`
	GroupFaceURL  string `gorm:"column:face_url;type:varchar(255)" json:"groupFaceURL"`
	CreateTime    int64  `gorm:"column:create_time" json:"createTime"`
	Status        int32  `gorm:"column:status" json:"status"`
	CreatorUserID string `gorm:"column:creator_user_id;type:varchar(64)" json:"creatorUserID"`
	GroupType     int32  `gorm:"column:group_type" json:"groupType"`
	OwnerUserID   string `gorm:"column:owner_user_id;type:varchar(64)" json:"ownerUserID"`
	MemberCount   int32  `gorm:"column:member_count" json:"memberCount"`

	UserID      string `gorm:"column:user_id;primary_key;type:varchar(64)" json:"userID"`
	Nickname    string `gorm:"column:nickname;type:varchar(255)" json:"nickname"`
	UserFaceURL string `gorm:"column:user_face_url;type:varchar(255)" json:"userFaceURL"`
	// Gender      int32  `gorm:"column:gender" json:"gender"`

	HandleResult  int32  `gorm:"column:handle_result" json:"handleResult"`
	ReqMsg        string `gorm:"column:req_msg;type:varchar(255)" json:"reqMsg"`
	HandledMsg    string `gorm:"column:handle_msg;type:varchar(255)" json:"handledMsg"`
	ReqTime       int64  `gorm:"column:req_time" json:"reqTime"`
	HandleUserID  string `gorm:"column:handle_user_id;type:varchar(64)" json:"handleUserID"`
	HandledTime   int64  `gorm:"column:handle_time" json:"handledTime"`
	Ex            string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
	AttachedInfo  string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
	JoinSource    int32  `gorm:"column:join_source" json:"joinSource"`
	InviterUserID string `gorm:"column:inviter_user_id;size:64"  json:"inviterUserID"`
}

type LocalUser struct {
	UserID           string `gorm:"column:user_id;primary_key;type:varchar(64)" json:"userID"`
	Nickname         string `gorm:"column:name;type:varchar(255)" json:"nickname"`
	FirstName        string `gorm:"column:first_name;type:varchar(255)" json:"firstName"`
	LastName         string `gorm:"column:last_name;type:varchar(255)" json:"lastName"`
	FaceURL          string `gorm:"column:face_url;type:varchar(255)" json:"faceURL"`
	CreateTime       int64  `gorm:"column:create_time" json:"createTime"`
	AppMangerLevel   int32  `gorm:"column:app_manger_level" json:"-"`
	Ex               string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
	AttachedInfo     string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
	GlobalRecvMsgOpt int32  `gorm:"column:global_recv_msg_opt" json:"globalRecvMsgOpt"`
	// Phone 对应服务端 UserInfo.Phone（有权限时下发）；DB 列名保持 phone_name 以兼容已有库。
	Phone    string `gorm:"column:phone_name;type:varchar(64)" json:"phone"`
	AreaCode string `gorm:"column:area_code;type:varchar(32)" json:"areaCode"`
	// CallRingtone* 对应服务端 UserInfo 来电铃声字段。
	CallRingtoneURL    string `gorm:"column:call_ringtone_url;type:varchar(512)" json:"callRingtoneURL"`
	CallRingtoneName   string `gorm:"column:call_ringtone_name;type:varchar(255)" json:"callRingtoneName"`
	CallRingtoneCover  string `gorm:"column:call_ringtone_cover;type:varchar(512)" json:"callRingtoneCover"`
	CallRingtoneAuthor string `gorm:"column:call_ringtone_author;type:varchar(255)" json:"callRingtoneAuthor"`
}

type LocalBlack struct {
	OwnerUserID string `gorm:"column:owner_user_id;primary_key;type:varchar(64)" json:"ownerUserID"`
	BlockUserID string `gorm:"column:block_user_id;primary_key;type:varchar(64)" json:"userID"`
	Nickname    string `gorm:"column:nickname;type:varchar(255)" json:"nickname"`
	FaceURL     string `gorm:"column:face_url;type:varchar(255)" json:"faceURL"`
	// Gender         int32  `gorm:"column:gender" json:"gender"`
	CreateTime     int64  `gorm:"column:create_time" json:"createTime"`
	AddSource      int32  `gorm:"column:add_source" json:"addSource"`
	OperatorUserID string `gorm:"column:operator_user_id;type:varchar(64)" json:"operatorUserID"`
	Ex             string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
	AttachedInfo   string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
}

type LocalSeqData struct {
	UserID string `gorm:"column:user_id;primary_key;type:varchar(64)"`
	Seq    uint32 `gorm:"column:seq"`
}

type LocalSeq struct {
	ID     string `gorm:"column:id;primary_key;type:varchar(64)"`
	MinSeq uint32 `gorm:"column:min_seq"`
}

type LocalChatLog struct {
	ClientMsgID      string `gorm:"column:client_msg_id;primary_key;type:char(64)" json:"clientMsgID"`
	ServerMsgID      string `gorm:"column:server_msg_id;type:char(64)" json:"serverMsgID"`
	SendID           string `gorm:"column:send_id;type:char(64)" json:"sendID"`
	RecvID           string `gorm:"column:recv_id;index:index_recv_id;type:char(64)" json:"recvID"`
	SenderPlatformID int32  `gorm:"column:sender_platform_id" json:"senderPlatformID"`
	SenderNickname   string `gorm:"column:sender_nick_name;type:varchar(255)" json:"senderNickname"`
	SenderFaceURL    string `gorm:"column:sender_face_url;type:varchar(255)" json:"senderFaceURL"`
	SessionType      int32  `gorm:"column:session_type" json:"sessionType"`
	MsgFrom          int32  `gorm:"column:msg_from" json:"msgFrom"`
	ContentType      int32  `gorm:"column:content_type;index:content_type_alone" json:"contentType"`
	Content          string `gorm:"column:content;type:varchar(1000)" json:"content"`
	IsRead           bool   `gorm:"column:is_read" json:"isRead"`
	Status           int32  `gorm:"column:status" json:"status"`
	Seq              int64  `gorm:"column:seq;index:index_seq;default:0" json:"seq"`
	SendTime         int64  `gorm:"column:send_time;index:index_send_time;" json:"sendTime"`
	CreateTime       int64  `gorm:"column:create_time" json:"createTime"`
	AttachedInfo     string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
	Ex               string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
	LocalEx          string `gorm:"column:local_ex;type:varchar(1024)" json:"localEx"`
}

type LocalConversation struct {
	ConversationID        string `gorm:"column:conversation_id;primary_key;type:char(128)" json:"conversationID"`
	ConversationType      int32  `gorm:"column:conversation_type" json:"conversationType"`
	UserID                string `gorm:"column:user_id;type:char(64)" json:"userID"`
	GroupID               string `gorm:"column:group_id;type:char(128)" json:"groupID"`
	ShowName              string `gorm:"column:show_name;type:varchar(255)" json:"showName"`
	FaceURL               string `gorm:"column:face_url;type:varchar(255)" json:"faceURL"`
	RecvMsgOpt            int32  `gorm:"column:recv_msg_opt" json:"recvMsgOpt"`
	UnreadCount           int32  `gorm:"column:unread_count" json:"unreadCount"`
	GroupAtType           int32  `gorm:"column:group_at_type" json:"groupAtType"`
	LatestMsg             string `gorm:"column:latest_msg;type:varchar(1000)" json:"latestMsg"`
	LatestMsgSendTime     int64  `gorm:"column:latest_msg_send_time;index:index_latest_msg_send_time" json:"latestMsgSendTime"`
	DraftText             string `gorm:"column:draft_text" json:"draftText"`
	DraftTextTime         int64  `gorm:"column:draft_text_time" json:"draftTextTime"`
	IsPinned              bool   `gorm:"column:is_pinned" json:"isPinned"`
	IsPrivateChat         bool   `gorm:"column:is_private_chat" json:"isPrivateChat"`
	BurnDuration          int32  `gorm:"column:burn_duration;default:30" json:"burnDuration"`
	IsNotInGroup          bool   `gorm:"column:is_not_in_group" json:"isNotInGroup"`
	UpdateUnreadCountTime int64  `gorm:"column:update_unread_count_time" json:"updateUnreadCountTime"`
	AttachedInfo          string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
	Ex                    string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
	MaxSeq                int64  `gorm:"column:max_seq" json:"maxSeq"`
	MinSeq                int64  `gorm:"column:min_seq" json:"minSeq"`
	MsgDestructTime       int64  `gorm:"column:msg_destruct_time;default:604800" json:"msgDestructTime"`
	IsMsgDestruct         bool   `gorm:"column:is_msg_destruct;default:false" json:"isMsgDestruct"`
	IsMuted               bool   `gorm:"column:is_muted;default:false" json:"isMuted"`
	MuteDuration          int32  `gorm:"column:mute_duration;default:0" json:"muteDuration"`
	MuteEndTime           int64  `gorm:"column:mute_end_time;default:0" json:"muteEndTime"`
}

func (LocalConversation) TableName() string {
	return "local_conversations"
}

type LocalConversationUnreadMessage struct {
	ConversationID string `gorm:"column:conversation_id;primary_key;type:char(128)" json:"conversationID"`
	ClientMsgID    string `gorm:"column:client_msg_id;primary_key;type:char(64)" json:"clientMsgID"`
	SendTime       int64  `gorm:"column:send_time" json:"sendTime"`
	Ex             string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
}

type LocalAdminGroupRequest struct {
	LocalGroupRequest
}

type LocalChatLogReactionExtensions struct {
	ClientMsgID             string `gorm:"column:client_msg_id;primary_key;type:char(64)" json:"clientMsgID"`
	LocalReactionExtensions []byte `gorm:"column:local_reaction_extensions" json:"localReactionExtensions"`
}

type NotificationSeqs struct {
	ConversationID string `gorm:"column:conversation_id;primary_key;type:char(128)" json:"conversationID"`
	Seq            int64  `gorm:"column:seq" json:"seq"`
}

func (NotificationSeqs) TableName() string {
	return "local_notification_seqs"
}

type LocalUpload struct {
	PartHash   string `gorm:"column:part_hash;primary_key" json:"partHash"`
	UploadID   string `gorm:"column:upload_id;type:varchar(1000)" json:"uploadID"`
	UploadInfo string `gorm:"column:upload_info;type:varchar(2000)" json:"uploadInfo"`
	ExpireTime int64  `gorm:"column:expire_time" json:"expireTime"`
	CreateTime int64  `gorm:"column:create_time" json:"createTime"`
}

type LocalConversationSyncedMaxSeq struct {
	ConversationID string `gorm:"column:conversation_id;primary_key;type:char(128)" json:"conversationID"`
	SyncedMaxSeq   int64  `gorm:"column:synced_max_seq" json:"syncedMaxSeq"`
}

func (LocalConversationSyncedMaxSeq) TableName() string {
	return "local_conversation_synced_max_seqs"
}

func (LocalUpload) TableName() string {
	return "local_uploads"
}

type LocalStranger struct {
	UserID           string `gorm:"column:user_id;primary_key;type:varchar(64)" json:"userID"`
	Nickname         string `gorm:"column:name;type:varchar(255)" json:"nickname"`
	FaceURL          string `gorm:"column:face_url;type:varchar(255)" json:"faceURL"`
	CreateTime       int64  `gorm:"column:create_time" json:"createTime"`
	AppMangerLevel   int32  `gorm:"column:app_manger_level" json:"-"`
	Ex               string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
	AttachedInfo     string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
	GlobalRecvMsgOpt int32  `gorm:"column:global_recv_msg_opt" json:"globalRecvMsgOpt"`
}

func (LocalStranger) TableName() string {
	return "local_stranger"
}

type LocalSendingMessages struct {
	ConversationID string `gorm:"column:conversation_id;primary_key;type:char(128)" json:"conversationID"`
	ClientMsgID    string `gorm:"column:client_msg_id;primary_key;type:char(64)" json:"clientMsgID"`
	Ex             string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
}

func (LocalSendingMessages) TableName() string {
	return "local_sending_messages"
}

type LocalUserCommand struct {
	UserID     string `gorm:"column:user_id;type:char(128);primary_key" json:"userID"`
	Type       int32  `gorm:"column:type;primary_key" json:"type"`
	Uuid       string `gorm:"column:uuid;type:varchar(255);primary_key" json:"uuid"`
	CreateTime int64  `gorm:"column:create_time" json:"createTime"`
	Value      string `gorm:"column:value;type:varchar(255)" json:"value"`
	Ex         string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
}

func (LocalUserCommand) TableName() string {
	return "local_user_command"
}

type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errs.New("type assertion to []byte failed").Wrap()
	}
	return json.Unmarshal(b, &a)
}

type LocalVersionSync struct {
	Table      string      `gorm:"column:table_name;type:varchar(255);primary_key" json:"tableName"`
	EntityID   string      `gorm:"column:entity_id;type:varchar(255);primary_key" json:"entityID"`
	VersionID  string      `gorm:"column:version_id" json:"versionID"`
	Version    uint64      `gorm:"column:version" json:"version"`
	CreateTime int64       `gorm:"column:create_time" json:"createTime"`
	UIDList    StringArray `gorm:"column:id_list;type:text" json:"uidList"`
}

func (LocalVersionSync) TableName() string {
	return "local_sync_version"
}

type LocalAppSDKVersion struct {
	Version   string `gorm:"column:version;type:varchar(255);primary_key" json:"version"`
	Installed bool   `gorm:"column:installed" json:"installed"` // Mark whether it has already been loaded
}

func (LocalAppSDKVersion) TableName() string {
	return "local_app_sdk_version"
}

// LocalSignalCallRecord 音视频通话记录（与 rtc.CallRecordItem 对齐，存于本地 SQLite / IndexDB）。
type LocalSignalCallRecord struct {
	SID                 string `gorm:"column:s_id;primaryKey;type:varchar(128)" json:"sID"`
	RoomID              string `gorm:"column:room_id;type:varchar(128);index:idx_signal_room" json:"roomID"`
	// Status 见 constant.SignalCallStatus*：1=已接听 2=未接通
	Status     int32 `gorm:"column:status;index:idx_signal_status" json:"status"`
	CreateTime int64 `gorm:"column:create_time;index:idx_signal_create" json:"createTime"`
	// DialDuration 拨打时长（毫秒）
	DialDuration int64 `gorm:"column:dial_duration" json:"dialDuration"`
	// CallDuration 通话时长（毫秒），未接通为 0
	CallDuration int64 `gorm:"column:call_duration" json:"callDuration"`
	MediaType           string `gorm:"column:media_type;type:varchar(32)" json:"mediaType"`
	SessionType         int32  `gorm:"column:session_type;index:idx_signal_session" json:"sessionType"`
	InviterUserID       string `gorm:"column:inviter_user_id;type:varchar(64)" json:"inviterUserID"`
	InviterUserNickname string `gorm:"column:inviter_user_nickname;type:varchar(255)" json:"inviterUserNickname"`
	InviterUserFaceURL  string `gorm:"column:inviter_user_face_url;type:varchar(512)" json:"inviterUserFaceURL"`
	GroupID             string `gorm:"column:group_id;type:varchar(64)" json:"groupID"`
	GroupName           string `gorm:"column:group_name;type:varchar(255)" json:"groupName"`
	// CalleeMatchText 被叫侧模糊查询用：由邀请时 inviteeUserIDList 等拼接（仅用于检索）
	CalleeMatchText     string `gorm:"column:callee_match_text;type:text;index:idx_signal_callee_txt" json:"-"`
	InviteeUserNickname string `gorm:"column:invitee_user_nickname;type:varchar(255)" json:"inviteeUserNickname"`
	InviteeUID          string `gorm:"column:invitee_uid;type:varchar(64);index:idx_signal_invitee_uid" json:"inviteeUID"`
	InviteeUserFaceURL  string `gorm:"column:invitee_user_face_url;type:varchar(512)" json:"inviteeUserFaceURL"`
	InviteeUserIDsJSON  string `gorm:"column:invitee_user_ids;type:text" json:"-"`
	// Direction 见 constant.SignalCallDirection*：1=主叫 2=被叫(已接) 3=被叫(未接/错过)
	Direction int32 `gorm:"column:direction;index:idx_signal_dir" json:"direction"`
	// Role 见 constant.SignalCallRole*：1=主叫 2=被叫（由本端 direction 推导）
	Role int32 `gorm:"column:role;index:idx_signal_role" json:"role"`
	// ConnectTime 接通时间（毫秒时间戳），未接通时为 0
	ConnectTime int64 `gorm:"column:connect_time" json:"connectTime"`
	EndTime     int64 `gorm:"column:end_time" json:"endTime"`
	// Action 触发本条记录的信令动作：accept / reject / cancel / hungup / timeout
	Action string `gorm:"column:action;type:varchar(32);index:idx_signal_action" json:"action"`
}

func (LocalSignalCallRecord) TableName() string {
	return "local_signal_call_records"
}
