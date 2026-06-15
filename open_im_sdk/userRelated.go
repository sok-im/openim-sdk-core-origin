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
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/openimsdk/openim-sdk-core/v3/internal/flagconst"
	"github.com/openimsdk/openim-sdk-core/v3/internal/third/file"

	icrypto "github.com/openimsdk/openim-sdk-core/v3/internal/crypto"
	"github.com/openimsdk/openim-sdk-core/v3/internal/openmls"
	"github.com/openimsdk/openim-sdk-core/v3/internal/redpacket"
	"github.com/openimsdk/openim-sdk-core/v3/internal/relation"
	sig "github.com/openimsdk/openim-sdk-core/v3/internal/signaling"
	ivirgil "github.com/openimsdk/openim-sdk-core/v3/internal/virgilsecurity"

	conv "github.com/openimsdk/openim-sdk-core/v3/internal/conversation_msg"
	"github.com/openimsdk/openim-sdk-core/v3/internal/group"
	"github.com/openimsdk/openim-sdk-core/v3/internal/interaction"
	"github.com/openimsdk/openim-sdk-core/v3/internal/third"
	"github.com/openimsdk/openim-sdk-core/v3/internal/user"
	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/ccontext"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/common"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/db_interface"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdkerrs"
	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"
	"github.com/openimsdk/protocol/push"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/jsonutil"
)

const (
	LogoutStatus = iota + 1
	Logging
	Logged
	LoggingOut
)

const (
	LogoutTips = "js sdk socket close"
)

func loginStatusString(status int) string {
	switch status {
	case LogoutStatus:
		return "LogoutStatus"
	case Logging:
		return "Logging"
	case Logged:
		return "Logged"
	case LoggingOut:
		return "LoggingOut"
	default:
		return fmt.Sprintf("Unknown(%d)", status)
	}
}

var (
	// UserForSDK Client-independent user class
	UserForSDK *LoginMgr
)

func shortFuncName(funcName string) string {
	parts := strings.Split(funcName, ".")
	return parts[len(parts)-1]
}

// isCaptchaFunc reports pre-login captcha APIs that must survive SDK logout context cancellation.
func isCaptchaFunc(shortFuncName string) bool {
	switch shortFuncName {
	case "GenerateCaptcha-fm", "VerifyCaptcha-fm",
		"GenerateClickCaptcha-fm", "VerifyClickCaptcha-fm":
		return true
	default:
		return false
	}
}

// isSessionLifecycleFunc reports Login/Logout which may overlap logout context cancellation.
func isSessionLifecycleFunc(shortFuncName string) bool {
	switch shortFuncName {
	case "Login-fm", "Logout-fm":
		return true
	default:
		return false
	}
}

// CheckResourceLoad checks the SDK is resource load status.
func CheckResourceLoad(uSDK *LoginMgr, funcName string) error {
	if uSDK == nil {
		log.ZError(context.Background(), "CheckResourceLoad", errs.New("SDK not initialized,userForSDK is nil", "funcName", funcName).Wrap(), "funcName", funcName, "uSDK", uSDK)
		return errs.New("SDK not initialized,userForSDK is nil", "funcName", funcName).Wrap()
	}

	if funcName == "" {
		return nil
	}

	short := shortFuncName(funcName)
	if short == "Login-fm" || short == "Log-fm" {
		return nil
	}
	if isCaptchaFunc(short) {
		return nil
	}
	if short == "PhoneGetSNInfo-fm" {
		return nil
	}

	if uSDK.getLoginStatus(context.Background()) != Logged {
		log.ZError(context.Background(), "CheckResourceLoad", errs.New("SDK not logged in", "funcName", funcName).Wrap(), "funcName", funcName, "uSDK", uSDK)
		return errs.New("SDK not logged in", "funcName", funcName).Wrap()
	}

	return nil
}

type LoginMgr struct {
	relation       *relation.Relation
	group          *group.Group
	conversation   *conv.Conversation
	user           *user.User
	file           *file.File
	signaling      *sig.Signaling
	crypto         *icrypto.Crypto
	redpacket      *redpacket.RedPacket
	virgilSecurity *ivirgil.VirgilSecurity
	openMLS        *openmls.OpenMLS

	db           db_interface.DataBase
	longConnMgr  *interaction.LongConnMgr
	msgSyncer    *interaction.MsgSyncer
	third        *third.Third
	token        string
	loginUserID  string
	connListener open_im_sdk_callback.OnConnListener

	justOnceFlag bool

	w           sync.Mutex
	loginStatus int

	groupListener        open_im_sdk_callback.OnGroupListener
	friendshipListener   open_im_sdk_callback.OnFriendshipListener
	conversationListener open_im_sdk_callback.OnConversationListener
	advancedMsgListener  open_im_sdk_callback.OnAdvancedMsgListener
	batchMsgListener     open_im_sdk_callback.OnBatchMsgListener
	userListener         open_im_sdk_callback.OnUserListener
	signalingListener    open_im_sdk_callback.OnSignalingListener
	businessListener     open_im_sdk_callback.OnCustomBusinessListener
	msgKvListener        open_im_sdk_callback.OnMessageKvInfoListener

	conversationCh chan common.Cmd2Value
	cmdWsCh        chan common.Cmd2Value
	msgSyncerCh    chan common.Cmd2Value
	loginMgrCh     chan common.Cmd2Value

	ctx       context.Context
	cancel    context.CancelFunc
	info      *ccontext.GlobalConfig
	id2MinSeq map[string]int64

	// wg tracks goroutines started in run() (long-conn pumps, listeners) so
	// logout() can wait for them to fully exit before closing the database.
	wg sync.WaitGroup
}

func (u *LoginMgr) GroupListener() open_im_sdk_callback.OnGroupListener {
	return u.groupListener
}

func (u *LoginMgr) FriendshipListener() open_im_sdk_callback.OnFriendshipListener {
	return u.friendshipListener
}

func (u *LoginMgr) ConversationListener() open_im_sdk_callback.OnConversationListener {
	return u.conversationListener
}

func (u *LoginMgr) AdvancedMsgListener() open_im_sdk_callback.OnAdvancedMsgListener {
	return u.advancedMsgListener
}

func (u *LoginMgr) BatchMsgListener() open_im_sdk_callback.OnBatchMsgListener {
	return u.batchMsgListener
}

func (u *LoginMgr) UserListener() open_im_sdk_callback.OnUserListener {
	return u.userListener
}

func (u *LoginMgr) SignalingListener() open_im_sdk_callback.OnSignalingListener {
	return u.signalingListener
}

func (u *LoginMgr) BusinessListener() open_im_sdk_callback.OnCustomBusinessListener {
	return u.businessListener
}

func (u *LoginMgr) MsgKvListener() open_im_sdk_callback.OnMessageKvInfoListener {
	return u.msgKvListener
}

func (u *LoginMgr) Exit() {
	u.cancel()
}

func (u *LoginMgr) GetToken() string {
	return u.token
}

func (u *LoginMgr) Third() *third.Third {
	return u.third
}

func (u *LoginMgr) ImConfig() sdk_struct.IMConfig {
	return sdk_struct.IMConfig{
		PlatformID:           u.info.PlatformID,
		ApiAddr:              u.info.ApiAddr,
		WsAddr:               u.info.WsAddr,
		DataDir:              u.info.DataDir,
		LogLevel:             u.info.LogLevel,
		IsExternalExtensions: u.info.IsExternalExtensions,
	}
}

func (u *LoginMgr) Conversation() *conv.Conversation {
	return u.conversation
}

func (u *LoginMgr) User() *user.User {
	return u.user
}

func (u *LoginMgr) File() *file.File {
	return u.file
}

func (u *LoginMgr) Group() *group.Group {
	return u.group
}

func (u *LoginMgr) Relation() *relation.Relation {
	return u.relation
}

func (u *LoginMgr) Signaling() *sig.Signaling {
	return u.signaling
}

func (u *LoginMgr) Crypto() *icrypto.Crypto {
	return u.crypto
}

func (u *LoginMgr) RedPacket() *redpacket.RedPacket {
	return u.redpacket
}

func (u *LoginMgr) VirgilSecurity() *ivirgil.VirgilSecurity {
	return u.virgilSecurity
}

func (u *LoginMgr) OpenMLS() *openmls.OpenMLS {
	return u.openMLS
}

func (u *LoginMgr) SetConversationListener(conversationListener open_im_sdk_callback.OnConversationListener) {
	u.conversationListener = conversationListener
}

func (u *LoginMgr) SetAdvancedMsgListener(advancedMsgListener open_im_sdk_callback.OnAdvancedMsgListener) {
	u.advancedMsgListener = advancedMsgListener
}

func (u *LoginMgr) SetMessageKvInfoListener(messageKvInfoListener open_im_sdk_callback.OnMessageKvInfoListener) {
	u.msgKvListener = messageKvInfoListener
}

func (u *LoginMgr) SetBatchMsgListener(batchMsgListener open_im_sdk_callback.OnBatchMsgListener) {
	u.batchMsgListener = batchMsgListener
}

func (u *LoginMgr) SetFriendshipListener(friendshipListener open_im_sdk_callback.OnFriendshipListener) {
	u.friendshipListener = friendshipListener
}

func (u *LoginMgr) SetGroupListener(groupListener open_im_sdk_callback.OnGroupListener) {
	u.groupListener = groupListener
}

func (u *LoginMgr) SetUserListener(userListener open_im_sdk_callback.OnUserListener) {
	u.userListener = userListener
}

func (u *LoginMgr) SetSignalingListener(listener open_im_sdk_callback.OnSignalingListener) {
	u.signalingListener = listener
}

func (u *LoginMgr) SetCustomBusinessListener(listener open_im_sdk_callback.OnCustomBusinessListener) {
	u.businessListener = listener
}
func (u *LoginMgr) GetLoginUserID() string {
	return u.loginUserID
}
func (u *LoginMgr) logoutListener(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Sprintf("panic: %+v\n%s", r, debug.Stack())

			log.ZWarn(ctx, "logoutListener panic", nil, "panic info", err)
		}
	}()

	for {
		select {
		case cmd := <-u.loginMgrCh:
			currentStatus := u.getLoginStatus(ctx)
			log.ZInfo(ctx, "lintao logoutListener triggered",
				"cmd", cmd.Cmd,
				"loginStatus", loginStatusString(currentStatus),
				"loginUserID", u.loginUserID)
			// Skip if already logging out or already logged out to prevent
			// concurrent logout goroutines from corrupting shared state.
			if currentStatus != Logged {
				log.ZWarn(ctx, "logoutListener: skip duplicate logout trigger",
					nil, "loginStatus", loginStatusString(currentStatus))
				continue
			}
			// Advance status immediately so any subsequent message in loginMgrCh
			// is discarded by the guard above before the goroutine runs.
			u.setLoginStatus(LoggingOut)
			// Run logout asynchronously: logout() waits on u.wg which includes
			// this goroutine; a synchronous call deadlocks in LoggingOut forever.
			logoutCtx := ctx
			if cmd.Ctx != nil {
				logoutCtx = cmd.Ctx
			}
			go func() {
				log.ZInfo(logoutCtx, "lintao logoutListener triggered asynchronously",
					"cmd", cmd.Cmd,
					"loginUserID", u.loginUserID)
				if err := u.logout(logoutCtx, true); err != nil {
					log.ZError(logoutCtx, "logout error", err)
				}
			}()
		case <-ctx.Done():
			log.ZInfo(ctx, "logoutListener done sdk logout.....")
			return
		}
	}

}

func NewLoginMgr() *LoginMgr {
	return &LoginMgr{
		info: &ccontext.GlobalConfig{},
	}
}
func (u *LoginMgr) getLoginStatus(_ context.Context) int {
	u.w.Lock()
	defer u.w.Unlock()
	return u.loginStatus
}
func (u *LoginMgr) setLoginStatus(status int) {
	u.w.Lock()
	prev := u.loginStatus
	u.loginStatus = status
	u.w.Unlock()
	if prev != status {
		log.ZInfo(context.Background(), "lintao login status changed",
			"from", loginStatusString(prev),
			"to", loginStatusString(status),
			"loginUserID", u.loginUserID)
	}
}
func (u *LoginMgr) checkSendingMessage(ctx context.Context) {
	sendingMessages, err := u.db.GetAllSendingMessages(ctx)
	if err != nil {
		log.ZError(ctx, "GetAllSendingMessages failed", err)
	}
	for _, message := range sendingMessages {
		if err := u.handlerSendingMsg(ctx, message); err != nil {
			log.ZError(ctx, "handlerSendingMsg failed", err, "message", message)
		}
		if err := u.db.DeleteSendingMessage(ctx, message.ConversationID, message.ClientMsgID); err != nil {
			log.ZError(ctx, "DeleteSendingMessage failed", err, "conversationID", message.ConversationID, "clientMsgID", message.ClientMsgID)
		}
	}
}

func (u *LoginMgr) handlerSendingMsg(ctx context.Context, sendingMsg *model_struct.LocalSendingMessages) error {
	tableMessage, err := u.db.GetMessage(ctx, sendingMsg.ConversationID, sendingMsg.ClientMsgID)
	if err != nil {
		return err
	}
	if tableMessage.Status != constant.MsgStatusSending {
		return nil
	}
	err = u.db.UpdateMessage(ctx, sendingMsg.ConversationID, &model_struct.LocalChatLog{ClientMsgID: sendingMsg.ClientMsgID, Status: constant.MsgStatusSendFailed})
	if err != nil {
		return err
	}
	conversation, err := u.db.GetConversation(ctx, sendingMsg.ConversationID)
	if err != nil {
		return err
	}
	latestMsg := &sdk_struct.MsgStruct{}
	if err := json.Unmarshal([]byte(conversation.LatestMsg), &latestMsg); err != nil {
		return err
	}
	if latestMsg.ClientMsgID == sendingMsg.ClientMsgID {
		latestMsg.Status = constant.MsgStatusSendFailed
		conversation.LatestMsg = jsonutil.StructToJsonString(latestMsg)
		return u.db.UpdateConversation(ctx, conversation)
	}
	return nil
}

func (u *LoginMgr) waitLogoutComplete() error {
	const maxWait = 10 * time.Second
	deadline := time.Now().Add(maxWait)
	status := u.getLoginStatus(context.Background())
	if status == LoggingOut {
		log.ZInfo(context.Background(), "lintao login wait for logout complete",
			"loginStatus", loginStatusString(status),
			"loginUserID", u.loginUserID,
			"maxWait", maxWait.String())
	}
	for status == LoggingOut {
		if time.Now().After(deadline) {
			log.ZError(context.Background(), "lintao login wait logout complete timeout", nil,
				"loginStatus", loginStatusString(u.getLoginStatus(context.Background())),
				"loginUserID", u.loginUserID,
				"maxWait", maxWait.String())
			return sdkerrs.ErrSdkInternal.WrapMsg("wait logout complete timeout")
		}
		time.Sleep(time.Millisecond * 100)
		status = u.getLoginStatus(context.Background())
	}
	if status != LogoutStatus {
		log.ZInfo(context.Background(), "lintao login proceed after waitLogoutComplete",
			"loginStatus", loginStatusString(status),
			"loginUserID", u.loginUserID)
	}
	return nil
}

func (u *LoginMgr) login(ctx context.Context, userID, token string) error {
	operationID := ccontext.Info(ctx).OperationID()
	log.ZInfo(ctx, "login enter",
		"userID", userID,
		"loginStatus", loginStatusString(u.getLoginStatus(ctx)),
		"prevLoginUserID", u.loginUserID,
		"ctxErr", ctx.Err())
	if err := u.waitLogoutComplete(); err != nil {
		return err
	}
	sessionCtx := u.Context()
	if sessionCtx.Err() != nil {
		log.ZError(ctx, "login session ctx canceled after waitLogoutComplete", sessionCtx.Err(),
			"userID", userID,
			"loginStatus", loginStatusString(u.getLoginStatus(ctx)),
			"prevLoginUserID", u.loginUserID)
	}
	ctx = ccontext.WithOperationID(sessionCtx, operationID)
	if u.getLoginStatus(ctx) == Logged {
		return sdkerrs.ErrLoginRepeat
	}
	u.setLoginStatus(Logging)
	log.ZDebug(ctx, "login start... ", "userID", userID, "token", token)
	t1 := time.Now()

	u.info.UserID = userID
	u.info.Token = token
	u.token = token
	u.loginUserID = userID
	var err error
	u.db, err = db.NewDataBase(ctx, userID, u.info.DataDir, int(u.info.LogLevel))
	if err != nil {
		return sdkerrs.ErrSdkInternal.WrapMsg("init database " + err.Error())
	}
	u.checkSendingMessage(ctx)
	log.ZDebug(ctx, "NewDataBase ok", "userID", userID, "dataDir", u.info.DataDir, "login cost time", time.Since(t1))
	u.user = user.NewUser(u.db, u.loginUserID, u.conversationCh)
	u.file = file.NewFile(u.db, u.loginUserID)
	u.relation = relation.NewFriend(u.loginUserID, u.db, u.user, u.conversationCh)

	u.group = group.NewGroup(u.loginUserID, u.db, u.conversationCh)
	u.third = third.NewThird(u.info.PlatformID, u.loginUserID, u.info.SystemType, u.info.LogFilePath, u.file)
	u.signaling = sig.NewSignaling(u.longConnMgr, u.loginUserID, u.info.PlatformID, u.db)
	u.crypto = icrypto.NewCrypto(u.loginUserID)
	u.redpacket = redpacket.NewRedPacket()
	u.virgilSecurity = ivirgil.NewVirgilSecurity()
	u.openMLS = openmls.NewOpenMLS()
	log.ZDebug(ctx, "forcedSynchronization success...", "login cost time: ", time.Since(t1))
	log.ZInfo(ctx, "lintao login: history sync strategy",
		"syncAllHistory", u.info.SyncAllHistory,
		"userID", userID)

	u.msgSyncer, _ = interaction.NewMsgSyncer(ctx, u.conversationCh, u.msgSyncerCh, u.loginUserID, u.longConnMgr, u.db, 0, u.info.SyncAllHistory)
	u.conversation = conv.NewConversation(ctx, u.longConnMgr, u.db, u.conversationCh, u.msgSyncerCh,
		u.relation, u.group, u.user, u.file, u.signaling, u.info.SyncAllHistory)
	u.relation.SetIncrSyncConversations(u.conversation.IncrSyncConversationsWithLock)
	u.group.SetIncrSyncConversations(u.conversation.IncrSyncConversationsWithLock)
	u.setListener(ctx)

	u.run(ctx)
	u.setLoginStatus(Logged)
	log.ZDebug(ctx, "login success...", "login cost time: ", time.Since(t1))
	return nil
}

func (u *LoginMgr) setListener(ctx context.Context) {
	setListener(ctx, &u.userListener, u.UserListener, u.user.SetListener, newEmptyUserListener)
	setListener(ctx, &u.friendshipListener, u.FriendshipListener, u.relation.SetListener, newEmptyFriendshipListener)
	setListener(ctx, &u.groupListener, u.GroupListener, u.group.SetGroupListener, newEmptyGroupListener)
	setListener(ctx, &u.conversationListener, u.ConversationListener, u.conversation.SetConversationListener, newEmptyConversationListener)
	setListener(ctx, &u.advancedMsgListener, u.AdvancedMsgListener, u.conversation.SetMsgListener, newEmptyAdvancedMsgListener)
	setListener(ctx, &u.batchMsgListener, u.BatchMsgListener, u.conversation.SetBatchMsgListener, nil)
	setListener(ctx, &u.businessListener, u.BusinessListener, u.conversation.SetBusinessListener, newEmptyCustomBusinessListener)
	setListener(ctx, &u.signalingListener, u.SignalingListener, u.signaling.SetListener, newEmptySignalingListener)
}

func setListener[T any](ctx context.Context, listener *T, getter func() T, setFunc func(listener func() T), newFunc func(context.Context) T) {
	if *(*unsafe.Pointer)(unsafe.Pointer(listener)) == nil && newFunc != nil {
		*listener = newFunc(ctx)
	}
	setFunc(getter)
}

func (u *LoginMgr) run(ctx context.Context) {
	u.longConnMgr.Run(ctx, &u.wg)
	u.wg.Add(3)
	go func() {
		defer u.wg.Done()
		u.msgSyncer.DoListener(ctx)
	}()
	go func() {
		defer u.wg.Done()
		common.DoListener(u.ctx, u.conversation)
	}()
	go func() {
		defer u.wg.Done()
		u.logoutListener(ctx)
	}()
}

func (u *LoginMgr) InitSDK(config sdk_struct.IMConfig, listener open_im_sdk_callback.OnConnListener) bool {
	if listener == nil {
		return false
	}
	u.info = &ccontext.GlobalConfig{}
	u.info.IMConfig = config
	u.connListener = listener
	u.initResources()
	return true
}

func (u *LoginMgr) Context() context.Context {
	return u.ctx
}

// preLoginCtx returns a context for APIs used before login (captcha, etc.).
// It is detached from SDK logout cancellation and API error callbacks.
func (u *LoginMgr) preLoginCtx() context.Context {
	if u.info == nil {
		return context.Background()
	}
	return ccontext.WithInfo(context.Background(), u.info)
}

func (u *LoginMgr) initResources() {
	log.ZInfo(context.Background(), "initResources",
		"prevLoginStatus", loginStatusString(u.getLoginStatus(context.Background())),
		"prevLoginUserID", u.loginUserID)
	ctx := ccontext.WithInfo(context.Background(), u.info)
	u.ctx, u.cancel = context.WithCancel(ctx)
	var convChanLen int
	if flagconst.TestMode {
		convChanLen = 100000
	} else {
		convChanLen = 1000
	}
	u.conversationCh = make(chan common.Cmd2Value, convChanLen)
	u.msgSyncerCh = make(chan common.Cmd2Value, 1000)
	u.loginMgrCh = make(chan common.Cmd2Value, 1)
	u.longConnMgr = interaction.NewLongConnMgr(u.ctx, u.connListener, u.userOnlineStatusChange, u.msgSyncerCh, u.loginMgrCh)
	u.ctx = ccontext.WithApiErrCode(u.ctx, &apiErrCallback{loginMgrCh: u.loginMgrCh, listener: u.connListener})
	u.setLoginStatus(LogoutStatus)
}

func (u *LoginMgr) userOnlineStatusChange(users map[string][]int32) {
	u.User().UserOnlineStatusChange(users)
}

func (u *LoginMgr) UnInitSDK() {
	if u.getLoginStatus(context.Background()) == Logged {
		fmt.Println("sdk not logout, please logout first")
		return
	}
	u.info = nil
	u.setLoginStatus(0)
}

// token error recycle recourse, kicked not recycle
func (u *LoginMgr) logout(ctx context.Context, isTokenValid bool) error {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Sprintf("panic: %+v\n%s", r, debug.Stack())
			log.ZWarn(ctx, "logout panic", nil, "panic info", err)
		}
	}()

	// Mark logging out before canceling the session context so new API calls are
	// rejected while in-flight requests are drained.
	log.ZInfo(ctx, "logout start",
		"isTokenValid", isTokenValid,
		"loginUserID", u.loginUserID,
		"loginStatus", loginStatusString(u.getLoginStatus(ctx)),
		"operationID", ccontext.Info(ctx).OperationID())
	u.setLoginStatus(LoggingOut)

	if ccontext.Info(ctx).OperationID() == LogoutTips {
		isTokenValid = true
	}
	if !isTokenValid {
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		err := u.longConnMgr.SendReqWaitResp(ctx, &push.DelUserPushTokenReq{UserID: u.info.UserID, PlatformID: u.info.PlatformID}, constant.LogoutMsg, &push.DelUserPushTokenResp{})
		if err != nil {
			log.ZWarn(ctx, "TriggerCmdLogout server recycle resources failed...", err)
		} else {
			log.ZDebug(ctx, "TriggerCmdLogout server recycle resources success...")
		}
	}
	log.ZInfo(ctx, "lintao logout cancel session ctx", "loginUserID", u.loginUserID)
	u.Exit()
	// Wait for all goroutines started in run() to finish before touching the
	// database or re-initialising channels.  Without this wait, doConnected
	// (called synchronously inside handlePushMsgAndEvent) can still be
	// executing a db call after db.Close() returns, causing a nil-pointer panic.
	u.wg.Wait()
	log.ZInfo(ctx, "lintao logout goroutines exited", "loginUserID", u.loginUserID)
	if u.signaling != nil {
		u.signaling.Close()
	}
	if u.db != nil {
		log.ZWarn(ctx, "TriggerCmdLogout db recycle resources", nil, "db", u.db)
		if err := u.db.Close(u.ctx); err != nil {
			log.ZWarn(ctx, "TriggerCmdLogout db recycle resources failed...", err)
		}
	} else {
		log.ZWarn(ctx, "TriggerCmdLogout db is nil", nil)
	}
	// Clear session credentials so pre-login APIs (captcha) do not send a stale token.
	if u.info != nil {
		u.info.UserID = ""
		u.info.Token = ""
	}
	u.token = ""
	u.loginUserID = ""
	// user object must be rest  when user logout
	u.initResources()
	log.ZInfo(ctx, "lintao logout complete, resources reinitialized",
		"isTokenValid", isTokenValid,
		"loginStatus", loginStatusString(u.getLoginStatus(ctx)),
		"sessionCtxErr", u.Context().Err())
	return nil
}

func (u *LoginMgr) setAppBackgroundStatus(ctx context.Context, isBackground bool) error {
	u.longConnMgr.SetBackground(isBackground)

	var resp sdkws.SetAppBackgroundStatusResp
	err := u.longConnMgr.SendReqWaitResp(ctx, &sdkws.SetAppBackgroundStatusReq{UserID: u.loginUserID, IsBackground: isBackground}, constant.SetBackgroundStatus, &resp)
	if err != nil {
		return err
	} else {
		if !isBackground {
			_ = common.TriggerCmdWakeUpDataSync(ctx, u.msgSyncerCh)
		}

		return nil
	}
}

func (u *LoginMgr) LongConnMgr() *interaction.LongConnMgr {
	return u.longConnMgr
}
