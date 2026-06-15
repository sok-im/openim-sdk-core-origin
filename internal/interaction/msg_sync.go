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

package interaction

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/common"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/db_interface"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"
	"github.com/openimsdk/protocol/msg"
	"github.com/openimsdk/tools/errs"

	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/log"
)

const (
	connectPullNums = 1
	defaultPullNums = 10
	SplitPullMsgNum = 100

	pullMsgGoroutineLimit = 10
)

// MsgSyncer is a central hub for message relay, responsible for sequential message gap pulling,
// handling network events, and managing app foreground and background events.
type MsgSyncer struct {
	loginUserID       string                // login user ID
	longConnMgr       *LongConnMgr          // long connection manager
	recvCh            chan common.Cmd2Value // channel for receiving push messages and the maximum SEQ number
	conversationCh    chan common.Cmd2Value // storage and session triggering
	syncedMaxSeqs     map[string]int64      // map of the maximum synced SEQ numbers for all group IDs
	syncedMaxSeqsLock sync.RWMutex          // syncedMaxSeqs map lock
	db                db_interface.DataBase // data store
	syncTimes         int                   // times of sync
	ctx               context.Context       // context
	reinstalled       bool                  // true if the app was uninstalled and reinstalled
	syncAllHistory    bool                  // pull all history gaps (legacy); when false only unread history is fetched
	isSyncing         bool                  // indicates whether data is being synced
	isSyncingLock     sync.Mutex            // lock for syncing state
}

// NewMsgSyncer creates a new instance of the message synchronizer.
// syncAllHistory mirrors IMConfig.SyncAllHistory: when false (the default) the
// syncer fetches the server hasReadSeq on every connect and only pulls messages
// with seq > hasReadSeq, skipping already-read history.  Set to true to restore
// the legacy behaviour of pulling all history gaps.
func NewMsgSyncer(ctx context.Context, conversationCh, recvCh chan common.Cmd2Value,
	loginUserID string, longConnMgr *LongConnMgr, db db_interface.DataBase, syncTimes int,
	syncAllHistory bool) (*MsgSyncer, error) {
	m := &MsgSyncer{
		loginUserID:    loginUserID,
		longConnMgr:    longConnMgr,
		recvCh:         recvCh,
		conversationCh: conversationCh,
		ctx:            ctx,
		syncedMaxSeqs:  make(map[string]int64),
		db:             db,
		syncTimes:      syncTimes,
		syncAllHistory: syncAllHistory,
	}
	if err := m.loadSeq(ctx); err != nil {
		log.ZError(ctx, "loadSeq err", err)
		return nil, err
	}
	log.ZInfo(ctx, "lintao NewMsgSyncer initialized",
		"syncAllHistory", syncAllHistory,
		"reinstalled", m.reinstalled,
		"syncedMaxSeqs", m.syncedMaxSeqs)
	return m, nil
}

// seq The db reads the data to the memory,set syncedMaxSeqs
func (m *MsgSyncer) loadSeq(ctx context.Context) error {
	conversationIDList, err := m.db.GetAllConversationIDList(ctx)
	if err != nil {
		log.ZError(ctx, "get conversation id list failed", err)
		return err
	}

	if len(conversationIDList) == 0 {
		version, err := m.db.GetAppSDKVersion(ctx)
		if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
			return err
		}
		if version == nil || !version.Installed {
			m.reinstalled = true
			log.ZInfo(ctx, "lintao loadSeq: detected reinstall (empty conversation list)",
				"installed", version != nil && version.Installed)
		}
	}

	// TODO With a large number of sessions(10w), this could potentially cause blocking and needs optimization.

	type SyncedSeq struct {
		ConversationID string
		MaxSyncedSeq   int64
		Err            error
	}

	partSize := 20
	currency := (len(conversationIDList)-1)/partSize + 1
	if len(conversationIDList) == 0 {
		currency = 0
	}
	var wg sync.WaitGroup
	resultMaps := make([]map[string]SyncedSeq, currency)

	for i := 0; i < currency; i++ {
		wg.Add(1)
		start := i * partSize
		end := start + partSize
		if i == currency-1 {
			end = len(conversationIDList)
		}

		resultMaps[i] = make(map[string]SyncedSeq)

		go func(i, start, end int) {
			defer wg.Done()
			for _, v := range conversationIDList[start:end] {
				maxSyncedSeq, err := m.db.CheckConversationNormalMsgSeq(ctx, v)
				resultMaps[i][v] = SyncedSeq{
					ConversationID: v,
					MaxSyncedSeq:   maxSyncedSeq,
					Err:            err,
				}
			}
		}(i, start, end)
	}

	wg.Wait()

	// merge map
	for _, resultMap := range resultMaps {
		for k, v := range resultMap {
			if v.Err != nil {
				log.ZError(ctx, "get group normal seq failed", errs.Wrap(v.Err), "conversationID", k)
				continue
			}
			m.syncedMaxSeqs[k] = v.MaxSyncedSeq
		}
	}
	notificationSeqs, err := m.db.GetNotificationAllSeqs(ctx)
	if err != nil {
		log.ZError(ctx, "get notification seq failed", err)
		return err
	}
	for _, notificationSeq := range notificationSeqs {
		m.syncedMaxSeqs[notificationSeq.ConversationID] = notificationSeq.Seq
	}

	persistedCursors, err := m.db.GetAllConversationSyncedMaxSeqs(ctx)
	if err != nil {
		log.ZWarn(ctx, "get persisted sync cursors failed", err)
	} else {
		log.ZDebug(ctx, "loadSeq", "persistedCursors", persistedCursors)
		for _, cursor := range persistedCursors {
			if cur, ok := m.syncedMaxSeqs[cursor.ConversationID]; !ok || cur < cursor.SyncedMaxSeq {
				m.syncedMaxSeqs[cursor.ConversationID] = cursor.SyncedMaxSeq
			}
		}
	}

	log.ZDebug(ctx, "loadSeq", "syncedMaxSeqs", m.syncedMaxSeqs)
	return nil
}

// DoListener Listen to the message pipe of the message synchronizer
// and process received and pushed messages
func (m *MsgSyncer) DoListener(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Sprintf("panic: %+v\n%s", r, debug.Stack())

			log.ZWarn(ctx, "DoListener panic", nil, "panic info", err)
		}
	}()
	for {
		select {
		case cmd := <-m.recvCh:
			m.handlePushMsgAndEvent(cmd)
		case <-ctx.Done():
			log.ZInfo(m.ctx, "msg syncer done, sdk logout.....")
			return
		}
	}
}

// get seqs need sync interval
func (m *MsgSyncer) getSeqsNeedSync(syncedMaxSeq, maxSeq int64) []int64 {
	var seqs []int64
	for i := syncedMaxSeq + 1; i <= maxSeq; i++ {
		seqs = append(seqs, i)
	}
	return seqs
}

// recv msg from
func (m *MsgSyncer) handlePushMsgAndEvent(cmd common.Cmd2Value) {
	switch cmd.Cmd {
	case constant.CmdConnSuccesss:
		log.ZInfo(cmd.Ctx, "recv long conn mgr connected", "cmd", cmd.Cmd, "value", cmd.Value)
		if m.startSync() {
			m.doConnected(cmd.Ctx)
		} else {
			log.ZWarn(cmd.Ctx, "syncing, ignore connected event", nil, "cmd", cmd.Cmd, "value", cmd.Value)
		}
	case constant.CmdWakeUpDataSync:
		log.ZInfo(cmd.Ctx, "app wake up, start sync msgs", "cmd", cmd.Cmd, "value", cmd.Value)
		if m.startSync() {
			m.doWakeupDataSync(cmd.Ctx)
		} else {
			log.ZWarn(cmd.Ctx, "syncing, ignore wake up event", nil, "cmd", cmd.Cmd, "value", cmd.Value)

		}
	case constant.CmdIMMessageSync:
		log.ZInfo(cmd.Ctx, "manually trigger IM message synchronization", "cmd", cmd.Cmd, "value", cmd.Value)
		m.doIMMessageSync(cmd.Ctx)

	case constant.CmdPushMsg:
		m.doPushMsg(cmd.Ctx, cmd.Value.(*sdkws.PushMessages))
	}
}

func (m *MsgSyncer) compareSeqsAndBatchSync(ctx context.Context, maxSeqToSync map[string]int64, pullNums int64) {
	needSyncSeqMap := make(map[string][2]int64)

	// skipNewNotifications is true when the SDK should NOT pull notification
	// history for conversations it has never tracked locally.  This covers:
	//   1. App reinstall / fresh install (m.reinstalled == true)
	//   2. syncAllHistory == false (default) – skip read notification history
	//      on every connect, not just after reinstall.
	//
	// For already-tracked notification conversations (syncedMaxSeqs entry
	// exists) we still perform normal incremental sync regardless of this flag.
	skipNewNotifications := m.reinstalled || !m.syncAllHistory
	log.ZInfo(ctx, "lintao compareSeqsAndBatchSync start",
		"syncAllHistory", m.syncAllHistory,
		"reinstalled", m.reinstalled,
		"skipNewNotifications", skipNewNotifications,
		"conversationMaxSeqs", maxSeqToSync,
		"pullNums", pullNums)

	if skipNewNotifications {
		var newNotificationSeqs []*model_struct.NotificationSeqs
		skippedNotificationCount := 0

		for conversationID, maxSeq := range maxSeqToSync {
			if IsNotification(conversationID) {
				if _, tracked := m.syncedMaxSeqs[conversationID]; !tracked {
					// New (untracked) notification conversation: record the
					// server's current max seq without pulling any history.
					if maxSeq != 0 {
						newNotificationSeqs = append(newNotificationSeqs, &model_struct.NotificationSeqs{
							ConversationID: conversationID,
							Seq:            maxSeq,
						})
						m.syncedMaxSeqs[conversationID] = maxSeq
						skippedNotificationCount++
						log.ZDebug(ctx, "lintao compareSeqsAndBatchSync: skip new notification history",
							"conversationID", conversationID, "maxSeq", maxSeq)
					}
					continue
				}
				// Already tracked – fall through to normal gap computation.
			}

			// Normal gap computation for message conversations and
			// already-tracked notification conversations.
			if syncedMaxSeq, ok := m.syncedMaxSeqs[conversationID]; ok {
				if maxSeq > syncedMaxSeq {
					needSyncSeqMap[conversationID] = [2]int64{syncedMaxSeq + 1, maxSeq}
				}
			} else {
				if maxSeq != 0 {
					needSyncSeqMap[conversationID] = [2]int64{0, maxSeq}
				}
			}
		}

		if len(newNotificationSeqs) > 0 {
			if err := m.db.BatchInsertNotificationSeq(ctx, newNotificationSeqs); err != nil {
				log.ZWarn(ctx, "BatchInsertNotificationSeq err", err)
			}
		}
		log.ZInfo(ctx, "lintao compareSeqsAndBatchSync unread-only path",
			"needSyncConversationSeqs", needSyncSeqMap,
			"skippedNewNotificationCount", skippedNotificationCount,
			"recordedNotificationSeqs", newNotificationSeqs)

		if m.reinstalled {
			defer func() {
				if err := m.db.SetAppSDKVersion(ctx, &model_struct.LocalAppSDKVersion{
					Installed: true,
				}); err != nil {
					log.ZError(ctx, "SetAppSDKVersion err", err)
				}
				m.reinstalled = false
			}()
			_ = m.syncAndTriggerReinstallMsgs(m.ctx, needSyncSeqMap, pullNums)
		} else {
			_ = m.syncAndTriggerMsgs(m.ctx, needSyncSeqMap, pullNums)
		}
	} else {
		log.ZInfo(ctx, "lintao compareSeqsAndBatchSync: syncAllHistory enabled, pull all history gaps")
		for conversationID, maxSeq := range maxSeqToSync {
			if syncedMaxSeq, ok := m.syncedMaxSeqs[conversationID]; ok {
				if maxSeq > syncedMaxSeq {
					needSyncSeqMap[conversationID] = [2]int64{syncedMaxSeq + 1, maxSeq}
				}
			} else {
				if maxSeq != 0 {
					needSyncSeqMap[conversationID] = [2]int64{0, maxSeq}
				}
			}
		}
		log.ZInfo(ctx, "lintao compareSeqsAndBatchSync legacy path",
			"needSyncConversationSeqs", needSyncSeqMap,
			"pullNums", pullNums)
		_ = m.syncAndTriggerMsgs(m.ctx, needSyncSeqMap, pullNums)
	}
}

// startSync checks if the sync is already in progress.
// If syncing is in progress, it returns false. Otherwise, it starts syncing and returns true.
func (m *MsgSyncer) startSync() bool {
	m.isSyncingLock.Lock()
	defer m.isSyncingLock.Unlock()

	if m.isSyncing {
		// If already syncing, return false
		return false
	}

	// Set syncing to true and start the sync
	m.isSyncing = true

	// Create a goroutine that waits for 5 seconds and then sets isSyncing to false
	go func() {
		time.Sleep(5 * time.Second)
		m.isSyncingLock.Lock()
		m.isSyncing = false
		m.isSyncingLock.Unlock()
	}()

	return true
}

func (m *MsgSyncer) doPushMsg(ctx context.Context, push *sdkws.PushMessages) {
	log.ZDebug(ctx, "push msgs", "push", push, "syncedMaxSeqs", m.syncedMaxSeqs)
	m.pushTriggerAndSync(ctx, push.Msgs, m.triggerConversation)
	m.pushTriggerAndSync(ctx, push.NotificationMsgs, m.triggerNotification)
}

func (m *MsgSyncer) pushTriggerAndSync(ctx context.Context, pushMessages map[string]*sdkws.PullMsgs, triggerFunc func(ctx context.Context, msgs map[string]*sdkws.PullMsgs) error) {
	if len(pushMessages) == 0 {
		return
	}
	needSyncSeqMap := make(map[string][2]int64)
	var lastSeq int64
	var storageMsgs []*sdkws.MsgData
	for conversationID, msgs := range pushMessages {
		for _, msg := range msgs.Msgs {
			if msg.Seq == 0 {
				_ = triggerFunc(ctx, map[string]*sdkws.PullMsgs{conversationID: {Msgs: []*sdkws.MsgData{msg}}})
				continue
			}
			lastSeq = msg.Seq
			storageMsgs = append(storageMsgs, msg)
		}
		if lastSeq == m.syncedMaxSeqs[conversationID]+int64(len(storageMsgs)) && lastSeq != 0 {
			log.ZDebug(ctx, "trigger msgs", "msgs", storageMsgs)
			_ = triggerFunc(ctx, map[string]*sdkws.PullMsgs{conversationID: {Msgs: storageMsgs}})
			m.syncedMaxSeqs[conversationID] = lastSeq
		} else if lastSeq != 0 && lastSeq > m.syncedMaxSeqs[conversationID] {
			//must pull message when message type is notification
			needSyncSeqMap[conversationID] = [2]int64{m.syncedMaxSeqs[conversationID] + 1, lastSeq}
		}
	}
	m.syncAndTriggerMsgs(ctx, needSyncSeqMap, defaultPullNums)
}

// Called after successful reconnection to synchronize the latest message
func (m *MsgSyncer) doConnected(ctx context.Context) {
	reinstalled := m.reinstalled
	log.ZInfo(ctx, "lintao doConnected start",
		"reinstalled", reinstalled,
		"syncAllHistory", m.syncAllHistory)
	if reinstalled {
		common.TriggerCmdSyncFlag(m.ctx, constant.AppDataSyncStart, m.conversationCh)
	} else {
		common.TriggerCmdSyncFlag(m.ctx, constant.MsgSyncBegin, m.conversationCh)
	}
	var resp sdkws.GetMaxSeqResp
	if err := m.longConnMgr.SendReqWaitResp(m.ctx, &sdkws.GetMaxSeqReq{UserID: m.loginUserID}, constant.GetNewestSeq, &resp); err != nil {
		log.ZError(m.ctx, "get max seq error", err)
		common.TriggerCmdSyncFlag(m.ctx, constant.MsgSyncFailed, m.conversationCh)
		return
	} else {
		log.ZDebug(m.ctx, "get max seq success", "resp", resp.MaxSeqs)
	}

	// Fetch the server-side hasReadSeq cursors when:
	//   (a) app was reinstalled / fresh-install, OR
	//   (b) syncAllHistory is false (default) – only pull unread history.
	// The fetched values advance syncedMaxSeqs so that seq ≤ hasReadSeq is
	// never re-pulled on this connect cycle.
	if reinstalled || !m.syncAllHistory {
		log.ZInfo(ctx, "lintao doConnected: prefetch hasReadSeq cursors",
			"reinstalled", reinstalled, "syncAllHistory", m.syncAllHistory)
		if hasReadSeqs, hErr := m.getHasReadSeqs(ctx); hErr == nil {
			bumpedCount := 0
			m.syncedMaxSeqsLock.Lock()
			for convID, hasReadSeq := range hasReadSeqs {
				if cur, ok := m.syncedMaxSeqs[convID]; !ok || cur < hasReadSeq {
					log.ZDebug(ctx, "lintao doConnected: advance syncedMaxSeq by hasReadSeq",
						"conversationID", convID, "prevSyncedMaxSeq", cur, "hasReadSeq", hasReadSeq)
					m.syncedMaxSeqs[convID] = hasReadSeq
					bumpedCount++
				}
			}
			m.syncedMaxSeqsLock.Unlock()
			log.ZInfo(ctx, "lintao doConnected: applied hasReadSeq cursors",
				"reinstalled", reinstalled,
				"syncAllHistory", m.syncAllHistory,
				"hasReadSeqs", hasReadSeqs,
				"bumpedSyncedMaxSeqCount", bumpedCount)

			// Persist normal-conversation cursors so that subsequent logins can
			// still skip already-read history even when no local messages were
			// stored (hasReadSeq == maxSeq at reinstall time).
			m.persistHasReadSeqs(ctx, hasReadSeqs)
		} else {
			log.ZWarn(ctx, "getHasReadSeqs failed", hErr,
				"reinstalled", reinstalled, "syncAllHistory", m.syncAllHistory)
		}
	} else {
		log.ZInfo(ctx, "lintao doConnected: skip hasReadSeq prefetch (syncAllHistory enabled)",
			"reinstalled", reinstalled)
	}

	m.compareSeqsAndBatchSync(ctx, resp.MaxSeqs, connectPullNums)
	if reinstalled {
		common.TriggerCmdSyncFlag(m.ctx, constant.AppDataSyncFinish, m.conversationCh)
	} else {
		common.TriggerCmdSyncFlag(m.ctx, constant.MsgSyncEnd, m.conversationCh)
	}
}

func (m *MsgSyncer) doWakeupDataSync(ctx context.Context) {
	common.TriggerCmdSyncData(ctx, m.conversationCh)
	var resp sdkws.GetMaxSeqResp
	if err := m.longConnMgr.SendReqWaitResp(m.ctx, &sdkws.GetMaxSeqReq{UserID: m.loginUserID}, constant.GetNewestSeq, &resp); err != nil {
		log.ZError(m.ctx, "get max seq error", err)
		return
	} else {
		log.ZDebug(m.ctx, "get max seq success", "resp", resp.MaxSeqs)
	}
	m.compareSeqsAndBatchSync(ctx, resp.MaxSeqs, defaultPullNums)
}

func (m *MsgSyncer) doIMMessageSync(ctx context.Context) {
	var resp sdkws.GetMaxSeqResp
	if err := m.longConnMgr.SendReqWaitResp(m.ctx, &sdkws.GetMaxSeqReq{UserID: m.loginUserID}, constant.GetNewestSeq, &resp); err != nil {
		log.ZError(m.ctx, "get max seq error", err)
		return
	} else {
		log.ZDebug(m.ctx, "get max seq success", "resp", resp.MaxSeqs)
	}
	m.compareSeqsAndBatchSync(ctx, resp.MaxSeqs, defaultPullNums)
}

func IsNotification(conversationID string) bool {
	return strings.HasPrefix(conversationID, "n_")
}

func (m *MsgSyncer) syncAndTriggerMsgs(ctx context.Context, seqMap map[string][2]int64, syncMsgNum int64) error {
	if len(seqMap) == 0 {
		log.ZDebug(ctx, "nothing to sync", "syncMsgNum", syncMsgNum)
		return nil
	}

	log.ZDebug(ctx, "current sync seqMap", "seqMap", seqMap)
	var (
		tempSeqMap = make(map[string][2]int64, 50)
		msgNum     = 0
	)

	for k, v := range seqMap {
		oneConversationSyncNum := v[1] - v[0] + 1
		tempSeqMap[k] = v
		// For notification conversations, use oneConversationSyncNum directly
		if IsNotification(k) {
			msgNum += int(oneConversationSyncNum)
		} else {
			// For regular conversations, ensure msgNum is the minimum of oneConversationSyncNum and syncMsgNum
			currentSyncMsgNum := int64(0)
			if oneConversationSyncNum > syncMsgNum {
				currentSyncMsgNum = syncMsgNum
			} else {
				currentSyncMsgNum = oneConversationSyncNum
			}
			msgNum += int(currentSyncMsgNum)
		}

		// If accumulated msgNum reaches SplitPullMsgNum, trigger a batch pull
		if msgNum >= SplitPullMsgNum {
			resp, err := m.pullMsgBySeqRange(ctx, tempSeqMap, syncMsgNum)
			if err != nil {
				log.ZError(ctx, "syncMsgFromServer error", err, "tempSeqMap", tempSeqMap)
				return err
			}
			_ = m.triggerConversation(ctx, resp.Msgs)
			_ = m.triggerNotification(ctx, resp.NotificationMsgs)
			for conversationID, seqs := range tempSeqMap {
				m.syncedMaxSeqs[conversationID] = seqs[1]
			}
			// Reset tempSeqMap and msgNum to handle the next batch
			tempSeqMap = make(map[string][2]int64, 50)
			msgNum = 0
		}
	}

	// Handle remaining messages to ensure all are synced
	if len(tempSeqMap) > 0 {
		resp, err := m.pullMsgBySeqRange(ctx, tempSeqMap, syncMsgNum)
		if err != nil {
			log.ZError(ctx, "syncMsgFromServer error", err, "tempSeqMap", tempSeqMap)
			return err
		}
		_ = m.triggerConversation(ctx, resp.Msgs)
		_ = m.triggerNotification(ctx, resp.NotificationMsgs)
		for conversationID, seqs := range tempSeqMap {
			m.syncedMaxSeqs[conversationID] = seqs[1]
		}
	}

	return nil
}

// Fragment synchronization message, seq refresh after successful trigger
func (m *MsgSyncer) syncAndTriggerReinstallMsgs(ctx context.Context, seqMap map[string][2]int64, syncMsgNum int64) error {
	if len(seqMap) > 0 {
		log.ZDebug(ctx, "current sync seqMap", "seqMap", seqMap)
		var (
			tempSeqMap = make(map[string][2]int64, 50)
			msgNum     = 0
			total      = len(seqMap)
		)

		for k, v := range seqMap {
			oneConversationSyncNum := min(v[1]-v[0]+1, syncMsgNum)
			tempSeqMap[k] = v
			if oneConversationSyncNum > 0 {
				// For regular conversations, ensure msgNum is the minimum of oneConversationSyncNum and syncMsgNum
				msgNum += int(min(oneConversationSyncNum, syncMsgNum))
			}
			if msgNum >= SplitPullMsgNum {
				resp, err := m.pullMsgBySeqRange(ctx, tempSeqMap, syncMsgNum)
				if err != nil {
					log.ZError(ctx, "syncMsgFromServer err", err, "tempSeqMap", tempSeqMap)
					return err
				}
				m.checkMessagesAndGetLastMessage(ctx, resp.Msgs)
				_ = m.triggerReinstallConversation(ctx, resp.Msgs, total)
				_ = m.triggerNotification(ctx, resp.NotificationMsgs)
				for conversationID, seqs := range tempSeqMap {
					m.syncedMaxSeqs[conversationID] = seqs[1]
				}

				// renew
				tempSeqMap = make(map[string][2]int64, 50)
				msgNum = 0
			}
		}

		if len(tempSeqMap) > 0 && msgNum > 0 {
			resp, err := m.pullMsgBySeqRange(ctx, tempSeqMap, syncMsgNum)
			if err != nil {
				log.ZError(ctx, "syncMsgFromServer err", err, "seqMap", seqMap)
				return err
			}

			m.checkMessagesAndGetLastMessage(ctx, resp.Msgs)
			_ = m.triggerReinstallConversation(ctx, resp.Msgs, total)
			_ = m.triggerNotification(ctx, resp.NotificationMsgs)
			for conversationID, seqs := range tempSeqMap {
				m.syncedMaxSeqs[conversationID] = seqs[1]
			}
		}
	} else {
		log.ZDebug(ctx, "noting conversation to sync", "syncMsgNum", syncMsgNum)
	}

	return nil
}
func (m *MsgSyncer) checkMessagesAndGetLastMessage(ctx context.Context, messages map[string]*sdkws.PullMsgs) {
	var conversationIDs []string

	for conversationID, message := range messages {
		allInValid := true
		for _, data := range message.Msgs {
			if data.Status < constant.MsgStatusHasDeleted {
				allInValid = false
				break
			}
		}
		if allInValid {
			conversationIDs = append(conversationIDs, conversationID)
		}
	}
	if len(conversationIDs) > 0 {
		resp, err := m.fetchLatestValidMessages(ctx, conversationIDs)
		if err != nil {
			log.ZError(ctx, "fetchLatestValidMessages", err, "conversationIDs", conversationIDs)
			return
		}
		for conversationID, message := range resp.Msgs {
			messages[conversationID] = &sdkws.PullMsgs{Msgs: []*sdkws.MsgData{message}}
		}
	}

}

func (m *MsgSyncer) splitSeqs(split int, seqsNeedSync []int64) (splitSeqs [][]int64) {
	if len(seqsNeedSync) <= split {
		splitSeqs = append(splitSeqs, seqsNeedSync)
		return
	}
	for i := 0; i < len(seqsNeedSync); i += split {
		end := i + split
		if end > len(seqsNeedSync) {
			end = len(seqsNeedSync)
		}
		splitSeqs = append(splitSeqs, seqsNeedSync[i:end])
	}
	return
}

func (m *MsgSyncer) pullMsgBySeqRange(ctx context.Context, seqMap map[string][2]int64, syncMsgNum int64) (resp *sdkws.PullMessageBySeqsResp, err error) {
	log.ZDebug(ctx, "pullMsgBySeqRange", "seqMap", seqMap, "syncMsgNum", syncMsgNum)

	req := sdkws.PullMessageBySeqsReq{UserID: m.loginUserID}
	for conversationID, seqs := range seqMap {
		req.SeqRanges = append(req.SeqRanges, &sdkws.SeqRange{
			ConversationID: conversationID,
			Begin:          seqs[0],
			End:            seqs[1],
			Num:            syncMsgNum,
		})
	}
	resp = &sdkws.PullMessageBySeqsResp{}
	if err := m.longConnMgr.SendReqWaitResp(ctx, &req, constant.PullMsgByRange, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MsgSyncer) fetchLatestValidMessages(ctx context.Context, conversationID []string) (resp *msg.GetLastMessageResp, err error) {
	log.ZDebug(ctx, "fetchLatestValidMessages", "conversationID", conversationID)

	req := msg.GetLastMessageReq{
		UserID:          m.loginUserID,
		ConversationIDs: conversationID,
	}
	resp = &msg.GetLastMessageResp{}
	if err := m.longConnMgr.SendReqWaitResp(ctx, &req, constant.PullConvLastMessage, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// synchronizes messages by SEQs.
func (m *MsgSyncer) syncMsgBySeqs(ctx context.Context, conversationID string, seqsNeedSync []int64) (allMsgs []*sdkws.MsgData, err error) {
	pullMsgReq := sdkws.PullMessageBySeqsReq{}
	pullMsgReq.UserID = m.loginUserID
	split := constant.SplitPullMsgNum
	seqsList := m.splitSeqs(split, seqsNeedSync)
	for i := 0; i < len(seqsList); {
		var pullMsgResp sdkws.PullMessageBySeqsResp
		err := m.longConnMgr.SendReqWaitResp(ctx, &pullMsgReq, constant.PullMsgByRange, &pullMsgResp)
		if err != nil {
			log.ZError(ctx, "syncMsgFromServerSplit err", err, "pullMsgReq", pullMsgReq)
			continue
		}
		i++
		allMsgs = append(allMsgs, pullMsgResp.Msgs[conversationID].Msgs...)
	}
	return allMsgs, nil
}

// triggers a conversation with a new message.
func (m *MsgSyncer) triggerConversation(ctx context.Context, msgs map[string]*sdkws.PullMsgs) error {
	if len(msgs) > 0 {
		err := common.TriggerCmdNewMsgCome(ctx, sdk_struct.CmdNewMsgComeToConversation{Msgs: msgs}, m.conversationCh)
		if err != nil {
			log.ZError(ctx, "triggerCmdNewMsgCome err", err, "msgs", msgs)
		}
		log.ZDebug(ctx, "triggerConversation", "msgs", msgs)
		return err
	} else {
		log.ZDebug(ctx, "triggerConversation is nil", "msgs", msgs)
	}
	return nil
}

// triggers a conversation with a new message.
func (m *MsgSyncer) triggerReinstallConversation(ctx context.Context, msgs map[string]*sdkws.PullMsgs, total int) (err error) {
	if len(msgs) > 0 {
		err = common.TriggerCmdMsgSyncInReinstall(ctx, sdk_struct.CmdMsgSyncInReinstall{
			Msgs:  msgs,
			Total: total,
		}, m.conversationCh)
		if err != nil {
			log.ZError(ctx, "triggerCmdNewMsgCome err", err, "msgs", msgs)
		}
		log.ZDebug(ctx, "triggerConversation", "msgs", msgs)
		return err
	} else {
		log.ZDebug(ctx, "triggerConversation is nil", "msgs", msgs)
	}
	return nil
}

func (m *MsgSyncer) triggerNotification(ctx context.Context, msgs map[string]*sdkws.PullMsgs) error {
	if len(msgs) > 0 {
		common.TriggerCmdNotification(ctx, sdk_struct.CmdNewMsgComeToConversation{Msgs: msgs}, m.conversationCh)
	} else {
		log.ZDebug(ctx, "triggerNotification is nil", "notifications", msgs)
	}
	return nil

}

// getHasReadSeqs queries the server for the per-conversation hasReadSeq of the
// current user.  The returned map is keyed by conversationID.
func (m *MsgSyncer) getHasReadSeqs(ctx context.Context) (map[string]int64, error) {
	req := msg.GetConversationsHasReadAndMaxSeqReq{UserID: m.loginUserID}
	var resp msg.GetConversationsHasReadAndMaxSeqResp
	if err := m.longConnMgr.SendReqWaitResp(ctx, &req, constant.GetConvMaxReadSeq, &resp); err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(resp.Seqs))
	for convID, seqs := range resp.Seqs {
		result[convID] = seqs.HasReadSeq
	}
	log.ZDebug(ctx, "lintao getHasReadSeqs success", "getHasReadSeqs", result)
	return result, nil
}

// persistHasReadSeqs writes the server-returned hasReadSeq values for normal
// (non-notification) conversations into SQLite.  The stored values act as a
// lower-bound sync cursor across logins, ensuring that conversations whose
// hasReadSeq equals maxSeq at reinstall time are never re-pulled from seq 0.
func (m *MsgSyncer) persistHasReadSeqs(ctx context.Context, hasReadSeqs map[string]int64) {
	var toUpsert []*model_struct.LocalConversationSyncedMaxSeq
	for convID, seq := range hasReadSeqs {
		if !IsNotification(convID) && seq > 0 {
			toUpsert = append(toUpsert, &model_struct.LocalConversationSyncedMaxSeq{
				ConversationID: convID,
				SyncedMaxSeq:   seq,
			})
		}
	}
	if len(toUpsert) == 0 {
		log.ZDebug(ctx, "lintao persistHasReadSeqs: nothing to persist")
		return
	}
	if err := m.db.BatchUpsertConversationSyncedMaxSeqs(ctx, toUpsert); err != nil {
		log.ZWarn(ctx, "persistHasReadSeqs: BatchUpsertConversationSyncedMaxSeqs failed", err)
	} else {
		log.ZDebug(ctx, "persistHasReadSeqs: saved sync cursors", "count", len(toUpsert))
	}
}
