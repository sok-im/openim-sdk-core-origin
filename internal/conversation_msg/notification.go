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

package conversation_msg

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/common"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"

	pConstant "github.com/openimsdk/protocol/constant"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/log"
)

const (
	syncWait = iota
	asyncNoWait
	asyncWait
)

// InitSyncProgress is initialize Sync when reinstall.
const InitSyncProgress = 10

func (c *Conversation) Work(c2v common.Cmd2Value) {
	log.ZDebug(c2v.Ctx, "NotificationCmd start", "caller", c2v.Caller, "cmd", c2v.Cmd, "value", c2v.Value)
	defer log.ZDebug(c2v.Ctx, "NotificationCmd end", "caller", c2v.Caller, "cmd", c2v.Cmd, "value", c2v.Value)
	switch c2v.Cmd {
	case constant.CmdNewMsgCome:
		c.doMsgNew(c2v)
	case constant.CmdUpdateConversation:
		c.doUpdateConversation(c2v)
	case constant.CmdUpdateMessage:
		c.doUpdateMessage(c2v)
	case constant.CmdNotification:
		c.doNotificationManager(c2v)
	case constant.CmdSyncData:
		c.syncData(c2v)
	case constant.CmdSyncFlag:
		c.syncFlag(c2v)
	case constant.CmdMsgSyncInReinstall:
		c.doMsgSyncByReinstalled(c2v)
	default:
		log.ZError(c2v.Ctx, "unknown cmd", nil, "cmd", c2v.Cmd)
	}
}

func (c *Conversation) syncFlag(c2v common.Cmd2Value) {
	ctx := c2v.Ctx
	syncFlag := c2v.Value.(sdk_struct.CmdNewMsgComeToConversation).SyncFlag
	switch syncFlag {
	case constant.AppDataSyncStart:
		log.ZDebug(ctx, "AppDataSyncStart")
		c.startTime = time.Now()
		c.ConversationListener().OnSyncServerStart(true)
		c.ConversationListener().OnSyncServerProgress(1)
		asyncWaitFunctions := []func(c context.Context) error{
			c.group.SyncAllJoinedGroupsAndMembers,
			c.relation.IncrSyncFriends,
		}
		runSyncFunctions(ctx, asyncWaitFunctions, asyncWait)
		c.addInitProgress(InitSyncProgress * 4 / 10)              // add 40% of InitSyncProgress as progress
		c.ConversationListener().OnSyncServerProgress(c.progress) // notify server current Progress

		syncWaitFunctions := []func(c context.Context) error{
			c.IncrSyncConversations,
			c.SyncAllConversationHashReadSeqs,
		}
		runSyncFunctions(ctx, syncWaitFunctions, syncWait)
		log.ZWarn(ctx, "core data sync over", nil, "cost time", time.Since(c.startTime).Seconds())
		c.addInitProgress(InitSyncProgress * 6 / 10)              // add 60% of InitSyncProgress as progress
		c.ConversationListener().OnSyncServerProgress(c.progress) // notify server current Progress

		asyncNoWaitFunctions := []func(c context.Context) error{
			c.user.SyncLoginUserInfoWithoutNotice,
			c.relation.SyncAllBlackListWithoutNotice,
		}
		runSyncFunctions(ctx, asyncNoWaitFunctions, asyncNoWait)

	case constant.AppDataSyncFinish:
		log.ZDebug(ctx, "AppDataSyncFinish", "time", time.Since(c.startTime).Milliseconds())
		c.progress = 100
		c.ConversationListener().OnSyncServerProgress(c.progress)
		c.ConversationListener().OnSyncServerFinish(true)
	case constant.MsgSyncBegin:
		log.ZDebug(ctx, "MsgSyncBegin")
		c.ConversationListener().OnSyncServerStart(false)
		c.syncData(c2v)
	case constant.MsgSyncFailed:
		c.ConversationListener().OnSyncServerFailed(false)
	case constant.MsgSyncEnd:
		log.ZDebug(ctx, "MsgSyncEnd", "time", time.Since(c.startTime).Milliseconds())
		c.ConversationListener().OnSyncServerFinish(false)
	}
}

func (c *Conversation) doNotificationManager(c2v common.Cmd2Value) {
	ctx := c2v.Ctx
	allMsg := c2v.Value.(sdk_struct.CmdNewMsgComeToConversation).Msgs

	for conversationID, msgs := range allMsg {
		log.ZDebug(ctx, "notification handling", "conversationID", conversationID, "msgs", msgs)

		// First, process all the notifications
		for _, msg := range msgs.Msgs {
			if msg.ContentType > constant.FriendNotificationBegin && msg.ContentType < constant.FriendNotificationEnd {
				c.relation.DoNotification(ctx, msg)
			} else if msg.ContentType > constant.UserNotificationBegin && msg.ContentType < constant.UserNotificationEnd {
				c.user.DoNotification(ctx, msg)
			} else if msg.ContentType > constant.GroupNotificationBegin && msg.ContentType < constant.GroupNotificationEnd {
				c.group.DoNotification(ctx, msg)
			} else if msg.ContentType > pConstant.SignalingNotificationBegin && msg.ContentType < pConstant.SignalingNotificationEnd {
				if c.signaling == nil {
					log.ZError(ctx, "signaling is nil", nil, "msg", msg)
					continue
				} else {
					c.signaling.DoNotification(ctx, msg)
				}
			} else {
				c.DoNotification(ctx, msg)
			}
		}

		// After all notifications are processed, update the sequence number
		if len(msgs.Msgs) != 0 {
			lastMsg := msgs.Msgs[len(msgs.Msgs)-1]
			log.ZDebug(ctx, "SetNotificationSeq", "conversationID", conversationID, "seq", lastMsg.Seq)
			if lastMsg.Seq != 0 {
				if err := c.db.SetNotificationSeq(ctx, conversationID, lastMsg.Seq); err != nil {
					// Log an error if setting the sequence number fails
					log.ZError(ctx, "SetNotificationSeq err", err, "conversationID", conversationID, "lastMsg", lastMsg)
				}
			}
		}
	}

}

func (c *Conversation) DoNotification(ctx context.Context, msg *sdkws.MsgData) {
	if err := c.doNotification(ctx, msg); err != nil {
		log.ZWarn(ctx, "DoConversationNotification failed", err)
	}
}

func (c *Conversation) doNotification(ctx context.Context, msg *sdkws.MsgData) error {
	switch msg.ContentType {
	case constant.ConversationChangeNotification:
		return c.DoConversationChangedNotification(ctx, msg)
	case constant.ConversationPrivateChatNotification: // 1701
		return c.DoConversationIsPrivateChangedNotification(ctx, msg)
	case constant.BusinessNotification:
		return c.doBusinessNotification(ctx, msg)
	case constant.RevokeNotification: // 2101
		return c.doRevokeMsg(ctx, msg)
	case constant.ClearConversationNotification: // 1703
		return c.doClearConversations(ctx, msg)
	case constant.DeleteMsgsNotification:
		return c.doDeleteMsgs(ctx, msg)
	case constant.HasReadReceipt: // 2200
		return c.doReadDrawing(ctx, msg)
	}
	return errs.New("unknown tips type", "contentType", msg.ContentType).Wrap()
}

func (c *Conversation) getConversationLatestMsgClientID(latestMsg string) string {
	msg := &sdk_struct.MsgStruct{}
	if err := json.Unmarshal([]byte(latestMsg), msg); err != nil {
		log.ZError(context.Background(), "getConversationLatestMsgClientID", err, "latestMsg", latestMsg)
	}
	return msg.ClientMsgID
}

func (c *Conversation) doUpdateConversation(c2v common.Cmd2Value) {
	// 记录调用来源，便于排查由哪个调用链触发会话更新。
	if c2v.Caller == "" {
		c2v.Caller = common.GetCaller(2)
	}
	ctx := c2v.Ctx
	// 统一的会话更新节点，Action 决定具体处理分支。
	node := c2v.Value.(common.UpdateConNode)
	log.ZInfo(ctx, "doUpdateConversation", "node", node, "cmd", c2v.Cmd, "caller", c2v.Caller)
	switch node.Action {
	case constant.AddConOrUpLatMsg:
		// 场景：新增会话，或更新会话最新消息（latest_msg）。
		var list []*model_struct.LocalConversation
		lc := node.Args.(model_struct.LocalConversation)
		// 先查本地是否已有该会话。
		oc, err := c.db.GetConversation(ctx, lc.ConversationID)
		if err == nil {
			// 异步消息更新以“发送时间较新”或“同一条消息(clientMsgID相同)”为准，避免旧消息覆盖新状态。
			if lc.LatestMsgSendTime >= oc.LatestMsgSendTime || c.getConversationLatestMsgClientID(lc.LatestMsg) == c.getConversationLatestMsgClientID(oc.LatestMsg) { // The session update of asynchronous messages is subject to the latest sending time
				err := c.db.UpdateColumnsConversation(ctx, node.ConID, map[string]interface{}{"latest_msg_send_time": lc.LatestMsgSendTime, "latest_msg": lc.LatestMsg})
				if err != nil {
					log.ZError(ctx, "updateConversationLatestMsgModel", err, "conversationID", node.ConID)
				} else {
					// 回写内存对象后，触发会话变更监听。
					oc.LatestMsgSendTime = lc.LatestMsgSendTime
					oc.LatestMsg = lc.LatestMsg
					list = append(list, oc)
					data := utils.StructToJsonString(list)
					log.ZInfo(ctx, "OnConversationChanged", "data", data)
					c.ConversationListener().OnConversationChanged(data)
				}
			}
		} else {
			// 本地不存在会话：插入新会话并触发 OnNewConversation。
			log.ZDebug(ctx, "new conversation", "lc", lc)
			err4 := c.db.InsertConversation(ctx, &lc)
			if err4 != nil {
				log.ZWarn(ctx, "insert new conversation err", err4)
			} else {
				list = append(list, &lc)
				c.ConversationListener().OnNewConversation(utils.StructToJsonString(list))
			}
		}

	case constant.TotalUnreadMessageChanged:
		// 场景：总未读数变化，查询 DB 后通知 UI。
		totalUnreadCount, err := c.db.GetTotalUnreadMsgCountDB(ctx)
		if err != nil {
			log.ZWarn(ctx, "GetTotalUnreadMsgCountDB err", err)
		} else {
			c.ConversationListener().OnTotalUnreadMessageCountChanged(totalUnreadCount)
		}
	case constant.UpdateConFaceUrlAndNickName:
		// 场景：会话展示信息（昵称/头像）更新。
		var lc model_struct.LocalConversation
		st := node.Args.(common.SourceIDAndSessionType)
		log.ZInfo(ctx, "UpdateConFaceUrlAndNickName", "st", st)
		switch st.SessionType {
		case constant.SingleChatType:
			// 单聊：SourceID 即对端 userID。
			lc.UserID = st.SourceID
			lc.ConversationID = c.getConversationIDBySessionType(st.SourceID, constant.SingleChatType)
			lc.ConversationType = constant.SingleChatType
		case constant.ReadGroupChatType:
			// 群聊：先根据群 ID 反查实际会话类型与会话 ID。
			conversationID, conversationType, err := c.getConversationTypeByGroupID(ctx, st.SourceID)
			if err != nil {
				return
			}
			lc.GroupID = st.SourceID
			lc.ConversationID = conversationID
			lc.ConversationType = conversationType
		case constant.NotificationChatType:
			// 通知会话：按通知会话规则构造 conversationID。
			lc.UserID = st.SourceID
			lc.ConversationID = c.getConversationIDBySessionType(st.SourceID, constant.NotificationChatType)
			lc.ConversationType = constant.NotificationChatType
		default:
			// 未知会话类型直接返回，避免写入错误数据。
			log.ZError(ctx, "not support sessionType", nil, "sessionType", st.SessionType)
			return
		}
		lc.ShowName = st.Nickname
		lc.FaceURL = st.FaceURL
		err := c.db.UpdateConversation(ctx, &lc)
		if err != nil {
			// log.Error("internal", "setConversationFaceUrlAndNickName database err:", err.Error())
			return
		}
		log.ZInfo(ctx, " doUpdateConversation", "lc", lc, "node", node, "cmd", c2v.Cmd, "caller", c2v.Caller)
		// 更新 DB 后，复用 ConChange 分支触发标准会话变更回调。
		c.doUpdateConversation(common.Cmd2Value{Value: common.UpdateConNode{ConID: lc.ConversationID, Action: constant.ConChange, Args: []string{lc.ConversationID}}})

	case constant.UpdateLatestMessageReadState:
		// 场景：将会话 latest_msg 标记为已读（例如点击会话后同步状态）。
		conversationID := node.ConID
		var latestMsg sdk_struct.MsgStruct
		l, err := c.db.GetConversation(ctx, conversationID)
		if err != nil {
			log.ZError(ctx, "getConversationLatestMsgModel err", err, "conversationID", conversationID)
		} else {
			// latest_msg 以 JSON 存储，反序列化后修改 IsRead 再写回。
			err := json.Unmarshal([]byte(l.LatestMsg), &latestMsg)
			if err != nil {
				log.ZError(ctx, "latestMsg,Unmarshal err", err)
			} else {
				latestMsg.IsRead = true
				newLatestMessage := utils.StructToJsonString(latestMsg)
				// 同步 latest_msg_send_time，保持会话排序字段一致。
				err = c.db.UpdateColumnsConversation(ctx, node.ConID, map[string]interface{}{"latest_msg_send_time": latestMsg.SendTime, "latest_msg": newLatestMessage})
				if err != nil {
					log.ZError(ctx, "updateConversationLatestMsgModel err", err)
				}
			}
		}
	case constant.UpdateLatestMessageFaceUrlAndNickName:
		// 场景：群成员资料变化时，必要时同步会话 latest_msg 中的发送者信息。
		args := node.Args.(common.UpdateMessageInfo)
		switch args.SessionType {
		case constant.ReadGroupChatType:
			conversationID := c.getConversationIDBySessionType(args.GroupID, constant.ReadGroupChatType)
			lc, err := c.db.GetConversation(ctx, conversationID)
			if err != nil {
				log.ZWarn(ctx, "getConversation err", err)
				return
			}
			var latestMsg sdk_struct.MsgStruct
			err = json.Unmarshal([]byte(lc.LatestMsg), &latestMsg)
			if err != nil {
				log.ZError(ctx, "latestMsg,Unmarshal err", err)
			} else {
				// 只有当“会话最后一条消息发送者”恰好是本次变更成员时，才需要更新 latest_msg 的头像昵称。
				if latestMsg.SendID == args.UserID {
					faceURL, name, _ := c.getUserNameAndFaceURL(ctx, args.UserID)
					if faceURL != "" {
						latestMsg.SenderFaceURL = faceURL
					} else if args.FaceURL != "" {
						latestMsg.SenderFaceURL = args.FaceURL
					}
					if name != "" {
						latestMsg.SenderNickname = name
					} else if args.Nickname != "" {
						latestMsg.SenderNickname = args.Nickname
					}
					newLatestMessage := utils.StructToJsonString(latestMsg)
					lc.LatestMsg = newLatestMessage
					err = c.db.UpdateColumnsConversation(ctx, conversationID, map[string]interface{}{"latest_msg": newLatestMessage})
					if err != nil {
						log.ZError(ctx, "updateConversationLatestMsgModel err", err)
					} else {
						// 仅回调当前受影响会话，避免全量刷新。
						var cList []*model_struct.LocalConversation
						cList = append(cList, lc)
						data := utils.StructToJsonStringDefault(cList)
						c.ConversationListener().OnConversationChanged(data)
					}

				}
			}
		}

	case constant.ConChange:
		// 场景：按会话 ID 列表触发“会话变更”通知（先从 DB 取最新数据）。
		conversationIDs := node.Args.([]string)
		conversations, err := c.db.GetMultipleConversationDB(ctx, conversationIDs)
		if err != nil {
			log.ZError(ctx, "getMultipleConversationModel err", err)
		} else {
			var newCList []*model_struct.LocalConversation
			for _, v := range conversations {
				// 过滤未初始化会话（latest_msg_send_time=0），避免对外暴露空壳会话。
				if v.LatestMsgSendTime != 0 {
					newCList = append(newCList, v)
				}
			}
			data := utils.StructToJsonStringDefault(newCList)
			log.ZInfo(ctx, "OnConversationChanged", "data", data)
			c.ConversationListener().OnConversationChanged(data)
		}
	case constant.NewCon:
		// 场景：按会话 ID 列表触发“新增会话”通知（由 DB 返回完整会话对象）。
		cidList := node.Args.([]string)
		cLists, err := c.db.GetMultipleConversationDB(ctx, cidList)
		if err != nil {
			log.ZWarn(ctx, "getMultipleConversationModel err", err)
		} else {
			if cLists != nil {
				log.ZDebug(ctx, "getMultipleConversationModel success", "cLists", cLists)
				c.ConversationListener().OnNewConversation(utils.StructToJsonString(cLists))
			}
		}
	case constant.ConChangeDirect:
		// 直传模式：Args 已是序列化后的会话列表 JSON，直接回调上层。
		cidList := node.Args.(string)
		log.ZInfo(ctx, "ConversationChanged", "cidList", cidList)
		c.ConversationListener().OnConversationChanged(cidList)

	case constant.NewConDirect:
		// 直传模式：Args 已是序列化后的新增会话 JSON，直接回调上层。
		cidList := node.Args.(string)
		log.ZDebug(ctx, "NewConversation", "cidList", cidList)
		c.ConversationListener().OnNewConversation(cidList)

	}
}

func (c *Conversation) doUpdateMessage(c2v common.Cmd2Value) {
	// 统一消息更新节点，当前函数主要处理“消息内发送者昵称/头像”同步。
	node := c2v.Value.(common.UpdateMessageNode)
	ctx := c2v.Ctx
	switch node.Action {
	case constant.UpdateMsgFaceUrlAndNickName:
		// 外部传入的资料变更信息（用户/群/会话类型）。
		args := node.Args.(common.UpdateMessageInfo)
		if faceURL, name, _ := c.getUserNameAndFaceURL(ctx, args.UserID); name != "" || faceURL != "" {
			if name != "" {
				args.Nickname = name
			}
			if faceURL != "" {
				args.FaceURL = faceURL
			}
		}
		switch args.SessionType {
		case constant.SingleChatType:
			// 单聊分两种场景：
			// 1) 更新的是“我自己”的资料：需要刷新所有单聊中我发送过的消息展示信息；
			// 2) 更新的是“对端用户”资料：只需要刷新对应单聊会话。
			if args.UserID == c.loginUserID {
				// 拿到全部单聊会话 ID，逐个更新消息表中的 sender 信息。
				conversationIDList, err := c.db.GetAllSingleConversationIDList(ctx)
				if err != nil {
					log.ZError(ctx, "GetAllSingleConversationIDList err", err)
					return
				} else {
					log.ZDebug(ctx, "get single conversationID list", "conversationIDList", conversationIDList)
					for _, conversationID := range conversationIDList {
						// 按会话更新指定发送者（args.UserID）的头像昵称。
						err := c.db.UpdateMsgSenderFaceURLAndSenderNickname(ctx, conversationID, args.UserID, args.FaceURL, args.Nickname)
						if err != nil {
							log.ZError(ctx, "UpdateMsgSenderFaceURLAndSenderNickname err", err)
							continue
						}
					}

				}
			} else {
				// 对端资料更新：只影响与该用户对应的单聊会话。
				conversationID := c.getConversationIDBySessionType(args.UserID, constant.SingleChatType)
				err := c.db.UpdateMsgSenderFaceURLAndSenderNickname(ctx, conversationID, args.UserID, args.FaceURL, args.Nickname)
				if err != nil {
					log.ZError(ctx, "UpdateMsgSenderFaceURLAndSenderNickname err", err)
				}

			}
		case constant.ReadGroupChatType:
			// 群聊：只更新目标群会话内，该成员发送过的消息展示信息。
			conversationID := c.getConversationIDBySessionType(args.GroupID, constant.ReadGroupChatType)
			err := c.db.UpdateMsgSenderFaceURLAndSenderNickname(ctx, conversationID, args.UserID, args.FaceURL, args.Nickname)
			if err != nil {
				log.ZError(ctx, "UpdateMsgSenderFaceURLAndSenderNickname err", err)
			}
		case constant.NotificationChatType:
			// 通知会话：按通知会话 ID 更新对应发送者展示信息。
			conversationID := c.getConversationIDBySessionType(args.UserID, constant.NotificationChatType)
			err := c.db.UpdateMsgSenderFaceURLAndSenderNickname(ctx, conversationID, args.UserID, args.FaceURL, args.Nickname)
			if err != nil {
				log.ZError(ctx, "UpdateMsgSenderFaceURLAndSenderNickname err", err)
			}
		default:
			// 未支持的会话类型直接返回，防止误更新。
			log.ZError(ctx, "not support sessionType", nil, "args", args, "node", node, "cmd", c2v.Cmd, "caller", c2v.Caller)
			return
		}
	}

}

func (c *Conversation) syncData(c2v common.Cmd2Value) {
	c.conversationSyncMutex.Lock()
	defer c.conversationSyncMutex.Unlock()
	defer func() {
		if r := recover(); r != nil {
			log.ZWarn(c2v.Ctx, "syncData panic recovered", nil,
				"panic", fmt.Sprintf("%+v", r),
				"stack", string(debug.Stack()))
		}
	}()

	ctx := c2v.Ctx
	c.startTime = time.Now()
	//clear SubscriptionStatusMap
	//c.user.OnlineStatusCache.DeleteAll()

	// Synchronous sync functions
	syncFuncs := []func(c context.Context) error{
		c.SyncAllConversationHashReadSeqs,
	}

	runSyncFunctions(ctx, syncFuncs, syncWait)

	// Asynchronous sync functions
	asyncFuncs := []func(c context.Context) error{
		c.user.SyncLoginUserInfo,
		c.relation.SyncAllBlackList,
		c.group.SyncAllJoinedGroupsAndMembersWithLock,
		c.relation.IncrSyncFriendsWithLock,
		c.IncrSyncConversationsWithLock,
	}

	runSyncFunctions(ctx, asyncFuncs, asyncNoWait)
}

func runSyncFunctions(ctx context.Context, funcs []func(c context.Context) error, mode int) {
	var wg sync.WaitGroup

	for _, fn := range funcs {
		switch mode {
		case asyncWait:
			wg.Add(1)
			go executeSyncFunction(ctx, fn, &wg)
		case asyncNoWait:
			go executeSyncFunction(ctx, fn, nil)
		case syncWait:
			executeSyncFunction(ctx, fn, nil)
		}
	}

	if mode == asyncWait {
		wg.Wait()
	}
}

func executeSyncFunction(ctx context.Context, fn func(c context.Context) error, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	defer func() {
		if r := recover(); r != nil {
			log.ZWarn(ctx, "executeSyncFunction panic recovered", nil,
				"panic", fmt.Sprintf("%+v", r),
				"stack", string(debug.Stack()))
		}
	}()
	// If the session context is already cancelled (e.g. during logout), skip
	// the sync entirely.  asyncNoWait goroutines are spawned fire-and-forget;
	// they can outlive the DoListener goroutine and run after db.Close() has
	// set the gorm connection to nil, causing a nil-pointer panic.
	if ctx.Err() != nil {
		log.ZWarn(ctx, "executeSyncFunction ctx err", ctx.Err())
		return
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	startTime := time.Now()
	err := fn(ctx)
	duration := time.Since(startTime)
	if err != nil {
		log.ZWarn(ctx, fmt.Sprintf("%s sync error", funcName), err, "duration", duration.Seconds())
	} else {
		log.ZDebug(ctx, fmt.Sprintf("%s completed successfully", funcName), "duration", duration.Seconds())
	}
}

func (c *Conversation) DoConversationChangedNotification(ctx context.Context, msg *sdkws.MsgData) error {
	c.conversationSyncMutex.Lock()
	defer c.conversationSyncMutex.Unlock()

	//var notification sdkws.ConversationChangedNotification
	tips := &sdkws.ConversationUpdateTips{}
	if err := utils.UnmarshalNotificationElem(msg.Content, tips); err != nil {
		log.ZWarn(ctx, "UnmarshalNotificationElem err", err, "msg", msg)
		return err
	}

	err := c.IncrSyncConversations(ctx)
	if err != nil {
		log.ZWarn(ctx, "IncrSyncConversations err", err)
		return err
	}
	return nil
}

func (c *Conversation) DoConversationIsPrivateChangedNotification(ctx context.Context, msg *sdkws.MsgData) error {
	c.conversationSyncMutex.Lock()
	defer c.conversationSyncMutex.Unlock()

	tips := &sdkws.ConversationSetPrivateTips{}
	if err := utils.UnmarshalNotificationElem(msg.Content, tips); err != nil {
		log.ZWarn(ctx, "UnmarshalNotificationElem err", err, "msg", msg)
		return err
	}

	err := c.IncrSyncConversations(ctx)
	if err != nil {
		log.ZWarn(ctx, "IncrSyncConversations err", err)
		return err
	}
	return nil
}

func (c *Conversation) doBusinessNotification(ctx context.Context, msg *sdkws.MsgData) error {
	var n sdk_struct.NotificationElem
	err := utils.JsonStringToStruct(string(msg.Content), &n)
	if err != nil {
		log.ZError(ctx, "unmarshal failed", err, "msg", msg)
		return err

	}
	c.businessListener().OnRecvCustomBusinessMessage(n.Detail)
	return nil
}
