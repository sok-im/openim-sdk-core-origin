package open_im_sdk

import "github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"

// 音视频信令导出接口说明
//
// 约定：
//   - 除 roomID/groupID/sID 等裸字符串参数外，其余 req 均为 JSON 字符串，反序列化为 protocol/rtc 对应消息。
//   - 成功时 callback.OnSuccess 回传 JSON；失败时 OnError(errCode, errMsg)。
//   - SDK 会覆盖 userID=loginUserID、部分 platformID；App 无需也不应伪造他人身份。
//   - E2EE：仅当 invitation.customData 内 e2ee.required=true 时服务端强制门禁；能力不足返回 1830–1834。
//
// 公共结构 InvitationInfo（嵌于多数请求）：
//
//	字段                  必填     含义
//	inviterUserID         否*     主叫用户 ID；Invite 时 SDK 覆盖为登录用户
//	inviteeUserIDList     条件    被叫列表；1v1 至少 1 个；群邀请为被邀成员
//	customData            否      JSON 字符串；E2EE 时内嵌 e2ee 描述符（见下）
//	groupID               条件    群通话必填；1v1 为空
//	roomID                条件    邀请时可空（服务端生成）；Accept/Cancel/HungUp 等必填服务端 roomID
//	timeout               否      振铃超时秒；0 时 SDK 默认 30
//	mediaType             是      "audio" | "video"
//	platformID            否*     平台；Invite 时 SDK 覆盖
//	sessionType           是      会话类型（单聊/群聊，与 OpenIM 一致）
//	initiateTime          否      发起时间 Unix ms；0 时服务端/SDK 补齐
//	busyLineUserIDList    否      忙线用户（响应侧常见）
//	callerRingtoneURL     否      主叫铃声 URL（多由服务端填充）
//	notAllowUserIDList    否      因接听设置不可邀请的用户（响应侧）
//	conversationID        否      规范化会话 ID；以服务端回传为准（单聊 si_a_b，群聊=groupID）
//
// customData.e2ee（E2EE 时建议写入；服务端只解析 required/callID，其余透传）：
//
//	required, scheme, version, conversationID, callID, roomID, roundID,
//	targetEpoch, generation, keyIndex, contextHash, expiresAt
//	禁止写入 K_media / exporter secret / 控制明文。
//
// 公共结构 e2eeCapability（Invite/Accept/Join/GetToken 在强制 E2EE 房间必填）：
//
//	schemes[]        是   支持的方案，如 ["mls-exporter-livekit-v1"]
//	frameCryptor     是   必须为 true，否则 1830
//	maxKeyRingSize   否   密钥环容量提示
//	platform         否   ios|android|…
//	clientVersion    条件 可解析版本号（如 "1"/"1.2.0"）；服务端 minVersion≥1 时必填，否则 1833

// SignalingInvite 发起 1v1 音视频邀请。
//
// 请求 JSON（SignalInviteReq）：
//
//	invitation         是   InvitationInfo；inviteeUserIDList 至少 1 人；roomID 可空
//	offlinePushInfo    否   离线推送 {title,desc,ex,…}
//	participant        否   参与者元数据（userInfo/groupInfo 等）
//	userID             否*  SDK 覆盖为登录用户
//	e2eeCapability     条件 强制 E2EE 时必填（见上）
//
// 成功响应 JSON（SignalInviteResp）：
//
//	token                 是  LiveKit JWT（E2EE 时短 TTL，attributes 含会话绑定）
//	roomID                是  服务端房间 ID，后续一律以此为准
//	liveURL               是  LiveKit 地址
//	busyLineUserIDList    否  忙线用户
//	calleeRingtoneURL     否  被叫铃声
//	notAllowUserIDList    否  不可邀请用户
//	callerRingtoneURL     否  主叫铃声
//	conversationID        是  规范化会话 ID（E2EE/多端同步以本字段为准）
func SignalingInvite(callback open_im_sdk_callback.Base, operationID string, signalInviteReq string) {
	call(callback, operationID, UserForSDK.Signaling().Invite, signalInviteReq)
}

// SignalingInviteInGroup 发起群音视频邀请。
//
// 请求 JSON（SignalInviteInGroupReq）：字段同 Invite；invitation.groupID 必填。
//
// 成功响应 JSON（SignalInviteInGroupResp）：
//
//	token, roomID, liveURL           是
//	busyLineUserIDList               否
//	calleeRingtoneURL                否
//	notAllowUserIDList               否
//	conversationID                   是  群通话等于 groupID
func SignalingInviteInGroup(callback open_im_sdk_callback.Base, operationID string, signalInviteInGroupReq string) {
	call(callback, operationID, UserForSDK.Signaling().InviteInGroup, signalInviteInGroupReq)
}

// SignalingAccept 被叫接听。
//
// 请求 JSON（SignalAcceptReq）：
//
//	invitation         是   至少含服务端 roomID；其余与邀请对齐
//	offlinePushInfo    否
//	participant        否
//	opUserPlatformID   否*  SDK 覆盖为本端 platformID
//	userID             否*  SDK 覆盖为登录用户
//	e2eeCapability     条件 强制 E2EE 房间必填
//
// 成功响应 JSON（SignalAcceptResp）：
//
//	token              是  LiveKit JWT
//	roomID             是
//	liveURL            是
//	conversationID     是  规范化会话 ID
//	e2ee               否  邀请时保存的 E2EE 描述符 JSON 字符串（原样回传；非 E2EE 可空）
func SignalingAccept(callback open_im_sdk_callback.Base, operationID string, signalAcceptReq string) {
	call(callback, operationID, UserForSDK.Signaling().Accept, signalAcceptReq)
}

// SignalingJoin 群成员加入进行中的群通话（无需事先被邀请）。
//
// 请求 JSON（SignalJoinReq）：
//
//	invitation         是   至少 roomID、groupID、mediaType、sessionType
//	participant        否
//	opUserPlatformID   否*  SDK 覆盖
//	userID             否*  SDK 覆盖
//	e2eeCapability     条件 强制 E2EE 时必填
//
// 成功响应 JSON（SignalJoinResp）：
//
//	token              是
//	roomID             是
//	liveURL            是
//	participant[]      否  当前通话成员元数据
//	inCall             是  房间内是否有人在通话（含本次 join）
//	conversationID     是  通常等于 groupID
//	e2ee               否  E2EE 描述符原样字符串
func SignalingJoin(callback open_im_sdk_callback.Base, operationID string, signalJoinReq string) {
	call(callback, operationID, UserForSDK.Signaling().Join, signalJoinReq)
}

// SignalingReject 被叫拒接。
//
// 请求 JSON（SignalRejectReq）：
//
//	invitation         是   含 roomID
//	offlinePushInfo    否
//	participant        否
//	opUserPlatformID   否*  SDK 覆盖
//	userID             否*  SDK 覆盖
//
// 成功响应：空对象 {}。
func SignalingReject(callback open_im_sdk_callback.Base, operationID string, signalRejectReq string) {
	call(callback, operationID, UserForSDK.Signaling().Reject, signalRejectReq)
}

// SignalingCancel 主叫取消邀请。
//
// 请求 JSON（SignalCancelReq）：
//
//	invitation         是   含 roomID
//	offlinePushInfo    否
//	participant        否
//	userID             否*  SDK 覆盖
//
// 成功响应：空对象 {}。
func SignalingCancel(callback open_im_sdk_callback.Base, operationID string, signalCancelReq string) {
	call(callback, operationID, UserForSDK.Signaling().Cancel, signalCancelReq)
}

// SignalingTimeout 主叫侧超时未接通（通常由 SDK 定时器触发，App 也可主动调）。
//
// 请求 JSON（SignalTimeoutReq）：
//
//	invitation         是   含 roomID
//	offlinePushInfo    否
//	userID             否*  SDK 覆盖
//
// 成功响应：空对象 {}。
func SignalingTimeout(callback open_im_sdk_callback.Base, operationID string, signalTimeoutReq string) {
	call(callback, operationID, UserForSDK.Signaling().Timeout, signalTimeoutReq)
}

// SignalingHungUp 任意一方挂断。
//
// 请求 JSON（SignalHungUpReq）：
//
//	invitation         是   含 roomID
//	offlinePushInfo    否
//	userID             否*  SDK 覆盖
//	callDuration       否*  App 传入会被忽略；SDK 按接通→挂断秒数写入并同步对端
//
// 成功响应：空对象 {}。
func SignalingHungUp(callback open_im_sdk_callback.Base, operationID string, signalHungUpReq string) {
	call(callback, operationID, UserForSDK.Signaling().HungUp, signalHungUpReq)
}

// SignalingHeartbeat 通话进行中周期性刷新本端忙线/通话状态 TTL。
//
// 参数：
//
//	roomID  是  当前通话房间 ID
//
// 建议间隔 20–30s（小于服务端约 1 分钟 TTL）。成功响应空。
func SignalingHeartbeat(callback open_im_sdk_callback.Base, operationID string, roomID string) {
	call(callback, operationID, UserForSDK.Signaling().Heartbeat, roomID)
}

// SignalingGetTokenByRoomID 按房间获取/续期 LiveKit Token。
//
// 请求 JSON（SignalGetTokenByRoomIDReq）：
//
//	roomID             是
//	participant        否
//	userID             否*  SDK 覆盖为登录用户（须与鉴权身份一致，否则 E2EE 返回 1834）
//	e2eeCapability     条件 强制 E2EE 房间续期时必填
//
// 成功响应 JSON（SignalGetTokenByRoomIDResp）：
//
//	token              是  LiveKit JWT；E2EE TTL≤5min，attributes 含 e2eeRoomID/ConversationID/CallID/Scheme/Version
//	liveURL            是
//	conversationID     是
//	e2ee               否  E2EE 描述符原样字符串
func SignalingGetTokenByRoomID(callback open_im_sdk_callback.Base, operationID string, signalGetTokenByRoomIDReq string) {
	call(callback, operationID, UserForSDK.Signaling().GetTokenByRoomID, signalGetTokenByRoomIDReq)
}

// SignalingGetRoomByGroupID 查询群是否有进行中的通话。
//
// 参数：
//
//	groupID  是  群 ID
//
// 成功响应 JSON（SignalGetRoomByGroupIDResp）：
//
//	invitation         否  进行中邀请（含 customData）
//	participant[]      否  LiveKit 已连接成员
//	roomID             否
//	inCall             是  false 表示无进行中通话
//	conversationID     否  有通话时通常等于 groupID
//	e2ee               否  E2EE 描述符原样字符串
func SignalingGetRoomByGroupID(callback open_im_sdk_callback.Base, operationID string, groupID string) {
	call(callback, operationID, UserForSDK.Signaling().GetRoomByGroupID, groupID)
}

// SignalingIsCallEndedByRoomID 查询指定房间通话是否已结束。
//
// 参数：roomID 必填。
// 成功响应：{"isEnded": true|false}
func SignalingIsCallEndedByRoomID(callback open_im_sdk_callback.Base, operationID string, roomID string) {
	call(callback, operationID, UserForSDK.Signaling().IsCallEndedByRoomID, roomID)
}

// SignalingGetInvitationInfoStartApp App 冷启动恢复待接听邀请。
//
// 请求 JSON（GetSignalInvitationInfoStartAppReq）：
//
//	userID  否*  SDK 覆盖为登录用户
//
// 成功响应 JSON（GetSignalInvitationInfoStartAppResp）：
//
//	invitation         否  待接听邀请；无则空
//	offlinePushInfo    否
//	conversationID     否  有邀请时回传规范化会话 ID
func SignalingGetInvitationInfoStartApp(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetSignalInvitationInfoStartApp, req)
}

// SignalingGetInvitationRecords 拉取服务端历史通话记录（与本地记录独立）。
//
// 请求 JSON（GetSignalInvitationRecordsReq）：
//
//	pagination     是  {pageNumber, showNumber}
//	sessionType    否
//	sendID/recvID  否  过滤
//	startTime/endTime 否  Unix 时间窗
//	joinedUsers    否
//
// 成功响应：{"total":N,"signalRecords":[...]}，记录含 roomID/sID/mediaType/sessionType/时长/文件等。
func SignalingGetInvitationRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetSignalInvitationRecords, req)
}

// SignalingGetLocalCallRecords 查询本地通话记录。
// 请求 JSON：userID 必填；status 0全部/1已接听/2未接通；offset/count；startTime/endTime 可选。
// 响应：{"total":N,"records":[...]}
func SignalingGetLocalCallRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalCallRecords, req)
}

// SignalingGetLocalCallRecordsWithUser 查询与指定用户的本地单聊通话记录（排除群）。
// 请求：userID 必填；offset/count/startTime/endTime 可选。
func SignalingGetLocalCallRecordsWithUser(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalCallRecordsWithUser, req)
}

// SignalingGetLocalMissedCallRecords 本地未接来电（被叫未接通）。
func SignalingGetLocalMissedCallRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalMissedCallRecords, req)
}

// SignalingGetLocalAnsweredCallRecords 本地已接通通话。
func SignalingGetLocalAnsweredCallRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalAnsweredCallRecords, req)
}

// SignalingGetLocalAllCallRecords 本地全部通话记录。
func SignalingGetLocalAllCallRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalAllCallRecords, req)
}

// SignalingGetLocalCallRecordsByUserName 按用户名模糊查本地通话记录。
// 请求：userName 必填；status/direction/offset/count/时间窗可选。
func SignalingGetLocalCallRecordsByUserName(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalCallRecordsByUserName, req)
}

// SignalingGetLocalCallRecordsByDate 按自然日查本地通话记录。
// 请求：date 必填（"2006-01-02"）；status/direction/offset/count 可选（count=-1 返回当天全部）。
func SignalingGetLocalCallRecordsByDate(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalCallRecordsByDate, req)
}

// SignalingSearchLocalCallRecords 搜索本地通话记录（不含群）。
func SignalingSearchLocalCallRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().SearchLocalSignalCallRecords, req)
}

// SignalingGetLocalCallRecordDetail 按 sID 查本地单条详情。参数 sID 必填。
func SignalingGetLocalCallRecordDetail(callback open_im_sdk_callback.Base, operationID string, sID string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalSignalCallRecordDetail, sID)
}

// SignalingDeleteLocalCallRecords 删除本地通话记录。参数 sIDs 为 JSON 字符串数组，如 ["sid1","sid2"]。
func SignalingDeleteLocalCallRecords(callback open_im_sdk_callback.Base, operationID string, sIDs string) {
	call(callback, operationID, UserForSDK.Signaling().DeleteLocalSignalCallRecords, sIDs)
}

// SignalingDeleteSignalRecords 删除服务端通话记录并同步删本地。
// 请求 JSON：{"sIDs":["..."]}，sIDs 必填。成功响应空。
func SignalingDeleteSignalRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().DeleteSignalRecords, req)
}

// SignalingClearAllLocalCallRecords 清空本地全部通话记录及缓存（不影响服务端）。
func SignalingClearAllLocalCallRecords(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.Signaling().ClearAllLocalSignalCallRecords)
}

// SignalingClearAllCallRecords 同 ClearAllLocalCallRecords（仅本地）。
func SignalingClearAllCallRecords(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.Signaling().ClearAllSignalCallRecords)
}

// SignalingSendCustomSignal 发送自定义信令（E2EE 换钥控制面等）。
//
// 请求 JSON（SignalSendCustomSignalReq）：
//
//	roomID       是  当前通话房间
//	customInfo   是  不透明 JSON 字符串，≤16KB；建议含 messageID 供服务端去重
//	                 例：{"messageID":"...","kind":"call_e2ee_control","mlsMessage":"<base64>"}
//	                 服务端不解析/不解密内容；发送者须为房间参与者
//
// 成功响应：空对象 {}。重复 messageID 在同房间 6h 内静默跳过仍返回成功。
//
// 对端回调 OnReceiveCustomSignal，载荷 JSON：
//
//	roomID             是
//	senderUserID       是
//	senderPlatformID   是
//	serverSeq          是  房间内单调序号
//	messageID          否  顶层去重键（发送方未带时可空）
//	customInfo         是  恒为字符串（三端类型一致），原样转发
func SignalingSendCustomSignal(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().SendCustomSignal, req)
}

// SignalingNotifyGroupCallEnded 通知群成员群通话已结束（触发 1523 / OnGroupCallStatusChanged 等）。
//
// 请求 JSON（SignalNotifyGroupCallEndedReq）：
//
//	groupID          是
//	roomID           否  有则用于解析 inviter/mediaType
//	inviterUserID    否  展示为结束方；默认可取邀请主叫
//	mediaType        否
//	durationSecs     否  通话秒数；0 表示未接通/取消
//	endReason        否  hungup|cancel|reject|timeout
//
// 成功响应：空对象 {}。
func SignalingNotifyGroupCallEnded(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().NotifyGroupCallEnded, req)
}
