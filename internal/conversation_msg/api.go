package conversation_msg

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/openimsdk/tools/errs"

	"github.com/openimsdk/openim-sdk-core/v3/internal/third/file"
	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/common"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/content_type"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdk_params_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdkerrs"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/server_api_params"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"

	"github.com/openimsdk/tools/log"

	pbConversation "github.com/openimsdk/protocol/conversation"
	pbMsg "github.com/openimsdk/protocol/msg"
	"github.com/openimsdk/protocol/sdkws"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func (c *Conversation) GetAllConversationList(ctx context.Context) ([]*model_struct.LocalConversation, error) {
	return c.db.GetAllConversationListDB(ctx)
}

func (c *Conversation) GetConversationListSplit(ctx context.Context, offset, count int) ([]*model_struct.LocalConversation, error) {
	return c.db.GetConversationListSplitDB(ctx, offset, count)
}

func (c *Conversation) HideConversation(ctx context.Context, conversationID string) error {
	err := c.db.ResetConversation(ctx, conversationID)
	if err != nil {
		return err
	}
	return nil
}

func (c *Conversation) GetAtAllTag(_ context.Context) string {
	return constant.AtAllString
}

func (c *Conversation) GetOneConversation(ctx context.Context, sessionType int32, sourceID string) (*model_struct.LocalConversation, error) {
	conversationID := c.getConversationIDBySessionType(sourceID, int(sessionType))
	lc, err := c.db.GetConversation(ctx, conversationID)
	if err == nil {
		if sessionType == constant.SingleChatType && lc.LatestMsgSendTime == 0 && !lc.IsPrivateChat && lc.BurnDuration == 0 {
			c.applyLoginUserBurnToSingleChatConversation(ctx, lc)
			if lc.IsPrivateChat && lc.BurnDuration > 0 {
				_ = c.db.UpdateColumnsConversation(ctx, lc.ConversationID, map[string]interface{}{
					"is_private_chat": true,
					"burn_duration":   lc.BurnDuration,
				})
			}
		}
		return lc, nil
	} else {
		var newConversation model_struct.LocalConversation
		newConversation.ConversationID = conversationID
		newConversation.ConversationType = sessionType
		switch sessionType {
		case constant.SingleChatType:
			newConversation.UserID = sourceID
			faceUrl, name, err := c.getUserNameAndFaceURL(ctx, sourceID)
			if err != nil {
				return nil, err
			}
			newConversation.ShowName = name
			newConversation.FaceURL = faceUrl
			if friendInfo, err := c.relation.Db().GetFriendInfoByFriendUserID(ctx, sourceID); err == nil {
				newConversation.IsPinned = friendInfo.IsPinned
			}
			c.applyLoginUserBurnToSingleChatConversation(ctx, &newConversation)
			log.ZInfo(ctx, " GetOneConversation", "conversation", newConversation)

		case constant.WriteGroupChatType, constant.ReadGroupChatType:
			newConversation.GroupID = sourceID
			g, err := c.group.FetchGroupOrError(ctx, sourceID)
			if err != nil {
				return nil, err
			}
			newConversation.ShowName = g.GroupName
			newConversation.FaceURL = g.FaceURL
			log.ZInfo(ctx, " GetOneConversation", "conversation", newConversation)

		}
		//double check if the conversation exists
		lc, err := c.db.GetConversation(ctx, conversationID)
		if err == nil {
			return lc, nil
		}
		return &newConversation, nil
	}
}

func (c *Conversation) GetMultipleConversation(ctx context.Context, conversationIDList []string) ([]*model_struct.LocalConversation, error) {
	conversations, err := c.db.GetMultipleConversationDB(ctx, conversationIDList)
	if err != nil {
		return nil, err
	}
	return conversations, nil

}

func (c *Conversation) HideAllConversations(ctx context.Context) error {
	err := c.db.ResetAllConversation(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *Conversation) SetConversationDraft(ctx context.Context, conversationID, draftText string) error {
	if draftText != "" {
		err := c.db.SetConversationDraftDB(ctx, conversationID, draftText)
		if err != nil {
			return err
		}
	} else {
		err := c.db.RemoveConversationDraft(ctx, conversationID, draftText)
		if err != nil {
			return err
		}
	}
	_ = common.TriggerCmdUpdateConversation(ctx, common.UpdateConNode{Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
	return nil
}

func (c *Conversation) SetConversation(ctx context.Context, conversationID string, req *pbConversation.ConversationReq) error {
	c.conversationSyncMutex.Lock()
	defer c.conversationSyncMutex.Unlock()

	existsInLocal := true
	lc, err := c.db.GetConversation(ctx, conversationID)
	if err != nil {
		if !errs.ErrRecordNotFound.Is(errs.Unwrap(err)) && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		existsInLocal = false
		sessionType, sourceID, parseErr := c.parseConversationID(conversationID)
		if parseErr != nil {
			return err
		}
		lc, err = c.GetOneConversation(ctx, sessionType, sourceID)
		if err != nil {
			return err
		}
	}
	apiReq := &pbConversation.SetConversationsReq{Conversation: req}
	if err = c.setConversation(ctx, apiReq, lc); err != nil {
		return err
	}
	applyConversationReqToLocal(lc, req)
	if existsInLocal {
		if updateMap := conversationReqToUpdateMap(req); len(updateMap) > 0 {
			if err = c.db.UpdateColumnsConversation(ctx, conversationID, updateMap); err != nil {
				return err
			}
		}
	} else if err = c.db.InsertConversation(ctx, lc); err != nil {
		return err
	}
	_ = common.TriggerCmdUpdateConversation(ctx, common.UpdateConNode{Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
	return c.IncrSyncConversations(ctx)
}

func (c *Conversation) SetConversationMute(ctx context.Context, req *sdk_params_callback.SetConversationMuteReq) error {
	if err := c.setConversationMute(ctx, &pbConversation.SetConversationMuteReq{
		OwnerUserID:    c.loginUserID,
		ConversationID: req.ConversationID,
		Duration:       req.Duration,
	}); err != nil {
		return err
	}
	c.conversationSyncMutex.Lock()
	defer c.conversationSyncMutex.Unlock()
	return c.IncrSyncConversations(ctx)
}

func (c *Conversation) SetConversationBurn(ctx context.Context, req *sdk_params_callback.SetConversationBurnReq) error {
	if err := c.setConversationBurn(ctx, &pbConversation.SetConversationBurnReq{
		OwnerUserID:    c.loginUserID,
		ConversationID: req.ConversationID,
		BurnDuration:   req.BurnDuration,
	}); err != nil {
		return err
	}
	c.conversationSyncMutex.Lock()
	defer c.conversationSyncMutex.Unlock()
	return c.IncrSyncConversations(ctx)
}

func (c *Conversation) GetTotalUnreadMsgCount(ctx context.Context) (totalUnreadCount int32, err error) {
	return c.db.GetTotalUnreadMsgCountDB(ctx)
}

func (c *Conversation) SetConversationListener(listener func() open_im_sdk_callback.OnConversationListener) {
	c.ConversationListener = listener
}

func (c *Conversation) updateMsgStatusAndTriggerConversation(ctx context.Context, clientMsgID, serverMsgID string, sendTime int64, status int32, s *sdk_struct.MsgStruct,
	lc *model_struct.LocalConversation, isOnlineOnly bool) {
	log.ZDebug(ctx, "this is test send message ", "sendTime", sendTime, "status", status, "clientMsgID", clientMsgID, "serverMsgID", serverMsgID)
	if isOnlineOnly {
		return
	}
	s.SendTime = sendTime
	s.Status = status
	s.ServerMsgID = serverMsgID
	err := c.db.UpdateMessageTimeAndStatus(ctx, lc.ConversationID, clientMsgID, serverMsgID, sendTime, status)
	if err != nil {
		log.ZWarn(ctx, "send message update message status error", err,
			"sendTime", sendTime, "status", status, "clientMsgID", clientMsgID, "serverMsgID", serverMsgID)
	}
	err = c.db.DeleteSendingMessage(ctx, lc.ConversationID, clientMsgID)
	if err != nil {
		log.ZWarn(ctx, "send message delete sending message error", err)
	}
	lc.LatestMsg = c.latestMsgJSON(ctx, s)
	lc.LatestMsgSendTime = sendTime
	_ = common.TriggerCmdUpdateConversation(ctx, common.UpdateConNode{ConID: lc.ConversationID, Action: constant.AddConOrUpLatMsg, Args: *lc}, c.GetCh())
}

func (c *Conversation) fileName(ftype string, id string) string {
	return fmt.Sprintf("msg_%s_%s", ftype, id)
}

func (c *Conversation) checkID(ctx context.Context, s *sdk_struct.MsgStruct,
	recvID, groupID string, options map[string]bool) (*model_struct.LocalConversation, error) {
	if recvID == "" && groupID == "" {
		return nil, sdkerrs.ErrArgs
	}
	s.SendID = c.loginUserID
	s.SenderPlatformID = c.platformID
	lc := &model_struct.LocalConversation{LatestMsgSendTime: s.CreateTime}
	//assemble messages and conversations based on single or group chat types
	if recvID == "" {
		g, err := c.group.FetchGroupOrError(ctx, groupID)
		if err != nil {
			return nil, err
		}
		lc.ShowName = g.GroupName
		lc.FaceURL = g.FaceURL
		switch g.GroupType {
		case constant.NormalGroup:
			s.SessionType = constant.WriteGroupChatType
			lc.ConversationType = constant.WriteGroupChatType
			lc.ConversationID = c.getConversationIDBySessionType(groupID, constant.WriteGroupChatType)
		case constant.SuperGroup, constant.WorkingGroup:
			s.SessionType = constant.ReadGroupChatType
			lc.ConversationID = c.getConversationIDBySessionType(groupID, constant.ReadGroupChatType)
			lc.ConversationType = constant.ReadGroupChatType
		}
		s.GroupID = groupID
		lc.GroupID = groupID
		var attachedInfo sdk_struct.AttachedInfoElem
		attachedInfo.GroupHasReadInfo.GroupMemberCount = g.MemberCount
		s.AttachedInfoElem = &attachedInfo
	} else {
		s.SessionType = constant.SingleChatType
		s.RecvID = recvID
		lc.ConversationID = utils.GetConversationIDByMsg(s)
		lc.UserID = recvID
		lc.ConversationType = constant.SingleChatType
		oldLc, err := c.db.GetConversation(ctx, lc.ConversationID)
		if err == nil && oldLc.IsPrivateChat {
			options[constant.IsNotPrivate] = false
			var attachedInfo sdk_struct.AttachedInfoElem
			attachedInfo.IsPrivateChat = true
			attachedInfo.BurnDuration = oldLc.BurnDuration
			s.AttachedInfoElem = &attachedInfo
		}
		if err != nil {
			//t := time.Now()
			faceUrl, name, err := c.getUserNameAndFaceURL(ctx, recvID)
			log.ZDebug(ctx, " GetUserNameAndFaceURL", "conversation", lc)
			if err != nil {
				return nil, err
			}
			lc.FaceURL = faceUrl
			lc.ShowName = name
		}

	}
	return lc, nil
}
func (c *Conversation) getConversationIDBySessionType(sourceID string, sessionType int) string {
	switch sessionType {
	case constant.SingleChatType:
		l := []string{c.loginUserID, sourceID}
		sort.Strings(l)
		return "si_" + strings.Join(l, "_") // single chat
	case constant.WriteGroupChatType:
		return "g_" + sourceID // group chat
	case constant.ReadGroupChatType:
		return "sg_" + sourceID // super group chat
	case constant.NotificationChatType:
		return "sn_" + sourceID + "_" + c.loginUserID // server notification chat
	}
	return ""
}

func (c *Conversation) parseConversationID(conversationID string) (sessionType int32, sourceID string, err error) {
	switch {
	case strings.HasPrefix(conversationID, "si_"):
		ids := strings.Split(strings.TrimPrefix(conversationID, "si_"), "_")
		if len(ids) != 2 {
			return 0, "", sdkerrs.ErrArgs.WrapMsg("invalid single chat conversation id")
		}
		for _, id := range ids {
			if id != c.loginUserID {
				return constant.SingleChatType, id, nil
			}
		}
		return 0, "", sdkerrs.ErrArgs.WrapMsg("invalid single chat conversation id")
	case strings.HasPrefix(conversationID, "sg_"):
		return constant.ReadGroupChatType, strings.TrimPrefix(conversationID, "sg_"), nil
	case strings.HasPrefix(conversationID, "g_"):
		return constant.WriteGroupChatType, strings.TrimPrefix(conversationID, "g_"), nil
	case strings.HasPrefix(conversationID, "sn_"):
		rest := strings.TrimPrefix(conversationID, "sn_")
		suffix := "_" + c.loginUserID
		if !strings.HasSuffix(rest, suffix) {
			return 0, "", sdkerrs.ErrArgs.WrapMsg("invalid notification conversation id")
		}
		return constant.NotificationChatType, strings.TrimSuffix(rest, suffix), nil
	default:
		return 0, "", sdkerrs.ErrArgs.WrapMsg("unknown conversation id format")
	}
}

func (c *Conversation) GetConversationIDBySessionType(_ context.Context, sourceID string, sessionType int) string {
	return c.getConversationIDBySessionType(sourceID, sessionType)
}

func (c *Conversation) SendMessage(ctx context.Context, s *sdk_struct.MsgStruct, recvID, groupID string, p *sdkws.OfflinePushInfo, isOnlineOnly bool) (*sdk_struct.MsgStruct, error) {
	filepathExt := func(name ...string) string {
		for _, path := range name {
			if ext := filepath.Ext(path); ext != "" {
				return ext
			}
		}
		return ""
	}
	options := make(map[string]bool, 2)
	lc, err := c.checkID(ctx, s, recvID, groupID, options)
	if err != nil {
		return nil, err
	}
	callback, _ := ctx.Value("callback").(open_im_sdk_callback.SendMsgCallBack)
	log.ZDebug(ctx, "before insert message is", "message", *s)
	if !isOnlineOnly {
		oldMessage, err := c.db.GetMessage(ctx, lc.ConversationID, s.ClientMsgID)
		if err != nil {
			localMessage := MsgStructToLocalChatLog(s)
			err := c.db.InsertMessage(ctx, lc.ConversationID, localMessage)
			if err != nil {
				return nil, err
			}
			err = c.db.InsertSendingMessage(ctx, &model_struct.LocalSendingMessages{
				ConversationID: lc.ConversationID,
				ClientMsgID:    localMessage.ClientMsgID,
			})
			if err != nil {
				return nil, err
			}
		} else {
			if oldMessage.Status != constant.MsgStatusSendFailed {
				return nil, sdkerrs.ErrMsgRepeated
			} else {
				s.Status = constant.MsgStatusSending
				err = c.db.InsertSendingMessage(ctx, &model_struct.LocalSendingMessages{
					ConversationID: lc.ConversationID,
					ClientMsgID:    s.ClientMsgID,
				})
				if err != nil {
					return nil, err
				}
			}
		}
		lc.LatestMsg = c.latestMsgJSON(ctx, s)
		log.ZDebug(ctx, "send message come here", "conversion", *lc)
		_ = common.TriggerCmdUpdateConversation(ctx, common.UpdateConNode{ConID: lc.ConversationID, Action: constant.AddConOrUpLatMsg, Args: *lc}, c.GetCh())
	}

	var delFile []string
	//media file handle
	switch s.ContentType {
	case constant.Picture:
		if s.Status == constant.MsgStatusSendSuccess {
			s.Content = utils.StructToJsonString(s.PictureElem)
			break
		}
		var sourcePath string
		if utils.FileExist(s.PictureElem.SourcePath) {
			sourcePath = s.PictureElem.SourcePath
			delFile = append(delFile, utils.FileTmpPath(s.PictureElem.SourcePath, c.DataDir))
		} else {
			sourcePath = utils.FileTmpPath(s.PictureElem.SourcePath, c.DataDir)
			delFile = append(delFile, sourcePath)
		}
		log.ZDebug(ctx, "send picture", "path", sourcePath)

		res, err := c.file.UploadFile(ctx, &file.UploadFileReq{
			ContentType: s.PictureElem.SourcePicture.Type,
			Filepath:    sourcePath,
			Uuid:        s.PictureElem.SourcePicture.UUID,
			Name:        c.fileName("picture", s.ClientMsgID) + filepathExt(s.PictureElem.SourcePicture.UUID, sourcePath),
			Cause:       "msg-picture",
		}, NewUploadFileCallback(ctx, callback.OnProgress, s, lc.ConversationID, c.db))
		if err != nil {
			c.updateMsgStatusAndTriggerConversation(ctx, s.ClientMsgID, "", s.CreateTime, constant.MsgStatusSendFailed, s, lc, isOnlineOnly)
			return nil, err
		}
		s.PictureElem.SourcePicture.Url = res.URL
		s.PictureElem.BigPicture = s.PictureElem.SourcePicture
		u, err := url.Parse(res.URL)
		if err == nil {
			snapshot := u.Query()
			snapshot.Set("type", "image")
			snapshot.Set("width", "640")
			snapshot.Set("height", "640")
			u.RawQuery = snapshot.Encode()
			s.PictureElem.SnapshotPicture = &sdk_struct.PictureBaseInfo{
				Width:  640,
				Height: 640,
				Url:    u.String(),
			}
		} else {
			log.ZError(ctx, "parse url failed", err, "url", res.URL, "err", err)
			s.PictureElem.SnapshotPicture = s.PictureElem.SourcePicture
		}
		s.Content = utils.StructToJsonString(s.PictureElem)
	case constant.Sound:
		if s.Status == constant.MsgStatusSendSuccess {
			s.Content = utils.StructToJsonString(s.SoundElem)
			break
		}
		var sourcePath string
		if utils.FileExist(s.SoundElem.SoundPath) {
			sourcePath = s.SoundElem.SoundPath
			delFile = append(delFile, utils.FileTmpPath(s.SoundElem.SoundPath, c.DataDir))
		} else {
			sourcePath = utils.FileTmpPath(s.SoundElem.SoundPath, c.DataDir)
			delFile = append(delFile, sourcePath)
		}
		// log.Info("", "file", sourcePath, delFile)

		res, err := c.file.UploadFile(ctx, &file.UploadFileReq{
			ContentType: s.SoundElem.SoundType,
			Filepath:    sourcePath,
			Uuid:        s.SoundElem.UUID,
			Name:        c.fileName("voice", s.ClientMsgID) + filepathExt(s.SoundElem.UUID, sourcePath),
			Cause:       "msg-voice",
		}, NewUploadFileCallback(ctx, callback.OnProgress, s, lc.ConversationID, c.db))
		if err != nil {
			c.updateMsgStatusAndTriggerConversation(ctx, s.ClientMsgID, "", s.CreateTime, constant.MsgStatusSendFailed, s, lc, isOnlineOnly)
			return nil, err
		}
		s.SoundElem.SourceURL = res.URL
		s.Content = utils.StructToJsonString(s.SoundElem)
	case constant.Video:
		if s.Status == constant.MsgStatusSendSuccess {
			s.Content = utils.StructToJsonString(s.VideoElem)
			break
		}
		var videoPath string
		var snapPath string
		if utils.FileExist(s.VideoElem.VideoPath) {
			videoPath = s.VideoElem.VideoPath
			snapPath = s.VideoElem.SnapshotPath
			delFile = append(delFile, utils.FileTmpPath(s.VideoElem.VideoPath, c.DataDir))
			delFile = append(delFile, utils.FileTmpPath(s.VideoElem.SnapshotPath, c.DataDir))
		} else {
			videoPath = utils.FileTmpPath(s.VideoElem.VideoPath, c.DataDir)
			snapPath = utils.FileTmpPath(s.VideoElem.SnapshotPath, c.DataDir)
			delFile = append(delFile, videoPath)
			delFile = append(delFile, snapPath)
		}
		log.ZDebug(ctx, "file", "videoPath", videoPath, "snapPath", snapPath, "delFile", delFile)

		var wg sync.WaitGroup
		wg.Add(2)
		var putErrs error
		go func() {
			defer wg.Done()
			snapRes, err := c.file.UploadFile(ctx, &file.UploadFileReq{
				ContentType: s.VideoElem.SnapshotType,
				Filepath:    snapPath,
				Uuid:        s.VideoElem.SnapshotUUID,
				Name:        c.fileName("videoSnapshot", s.ClientMsgID) + filepathExt(s.VideoElem.SnapshotUUID, snapPath),
				Cause:       "msg-video-snapshot",
			}, nil)
			if err != nil {
				log.ZWarn(ctx, "upload video snapshot failed", err)
				return
			}
			s.VideoElem.SnapshotURL = snapRes.URL
		}()

		go func() {
			defer wg.Done()
			res, err := c.file.UploadFile(ctx, &file.UploadFileReq{
				ContentType: content_type.GetType(s.VideoElem.VideoType, filepath.Ext(s.VideoElem.VideoPath)),
				Filepath:    videoPath,
				Uuid:        s.VideoElem.VideoUUID,
				Name:        c.fileName("video", s.ClientMsgID) + filepathExt(s.VideoElem.VideoUUID, videoPath),
				Cause:       "msg-video",
			}, NewUploadFileCallback(ctx, callback.OnProgress, s, lc.ConversationID, c.db))
			if err != nil {
				c.updateMsgStatusAndTriggerConversation(ctx, s.ClientMsgID, "", s.CreateTime, constant.MsgStatusSendFailed, s, lc, isOnlineOnly)
				putErrs = err
				return
			}
			if res != nil {
				s.VideoElem.VideoURL = res.URL
			}
		}()
		wg.Wait()
		if err := putErrs; err != nil {
			return nil, err
		}
		s.Content = utils.StructToJsonString(s.VideoElem)
	case constant.File:
		if s.Status == constant.MsgStatusSendSuccess {
			s.Content = utils.StructToJsonString(s.FileElem)
			break
		}
		name := s.FileElem.FileName

		if name == "" {
			name = s.FileElem.FilePath
		}
		if name == "" {
			name = fmt.Sprintf("msg_file_%s.unknown", s.ClientMsgID)
		}

		var sourcePath string
		if utils.FileExist(s.FileElem.FilePath) {
			sourcePath = s.FileElem.FilePath
			delFile = append(delFile, utils.FileTmpPath(s.FileElem.FilePath, c.DataDir))
		} else {
			sourcePath = utils.FileTmpPath(s.FileElem.FilePath, c.DataDir)
			delFile = append(delFile, sourcePath)
		}

		res, err := c.file.UploadFile(ctx, &file.UploadFileReq{
			ContentType: content_type.GetType(s.FileElem.FileType, filepath.Ext(s.FileElem.FilePath), filepath.Ext(s.FileElem.FileName)),
			Filepath:    sourcePath,
			Uuid:        s.FileElem.UUID,
			Name:        c.fileName("file", s.ClientMsgID) + "/" + filepath.Base(name),
			Cause:       "msg-file",
		}, NewUploadFileCallback(ctx, callback.OnProgress, s, lc.ConversationID, c.db))
		if err != nil {
			c.updateMsgStatusAndTriggerConversation(ctx, s.ClientMsgID, "", s.CreateTime, constant.MsgStatusSendFailed, s, lc, isOnlineOnly)
			return nil, err
		}
		s.FileElem.SourceURL = res.URL
		s.Content = utils.StructToJsonString(s.FileElem)
	case constant.Text:
		s.Content = utils.StructToJsonString(s.TextElem)
	case constant.AtText:
		s.Content = utils.StructToJsonString(s.AtTextElem)
	case constant.Location:
		s.Content = utils.StructToJsonString(s.LocationElem)
	case constant.Custom:
		s.Content = utils.StructToJsonString(s.CustomElem)
	case constant.Merger:
		s.Content = utils.StructToJsonString(s.MergeElem)
	case constant.Quote:
		quoteMessage, err := c.db.GetMessage(ctx, lc.ConversationID, s.QuoteElem.QuoteMessage.ClientMsgID)
		if err != nil {
			log.ZWarn(ctx, "quote message not found", err)
		}
		s.QuoteElem.QuoteMessage.Seq = quoteMessage.Seq
		s.Content = utils.StructToJsonString(s.QuoteElem)
	case constant.Card:
		s.Content = utils.StructToJsonString(s.CardElem)
	case constant.Face:
		s.Content = utils.StructToJsonString(s.FaceElem)
	case constant.AdvancedText:
		s.Content = utils.StructToJsonString(s.AdvancedTextElem)
	default:
		return nil, sdkerrs.ErrMsgContentTypeNotSupport
	}
	if utils.IsContainInt(int(s.ContentType), []int{constant.Picture, constant.Sound, constant.Video, constant.File}) {
		if !isOnlineOnly {
			localMessage := MsgStructToLocalChatLog(s)
			log.ZDebug(ctx, "update message is ", "localMessage", localMessage)
			err = c.db.UpdateMessage(ctx, lc.ConversationID, localMessage)
			if err != nil {
				return nil, err
			}
		}
	}

	return c.sendMessageToServer(ctx, s, lc, callback, delFile, p, options, isOnlineOnly)
}

func (c *Conversation) SendMessageNotOss(ctx context.Context, s *sdk_struct.MsgStruct, recvID, groupID string,
	p *sdkws.OfflinePushInfo, isOnlineOnly bool) (*sdk_struct.MsgStruct, error) {
	options := make(map[string]bool, 2)
	lc, err := c.checkID(ctx, s, recvID, groupID, options)
	if err != nil {
		return nil, err
	}
	callback, _ := ctx.Value("callback").(open_im_sdk_callback.SendMsgCallBack)
	if !isOnlineOnly {
		oldMessage, err := c.db.GetMessage(ctx, lc.ConversationID, s.ClientMsgID)
		if err != nil {
			localMessage := MsgStructToLocalChatLog(s)
			err := c.db.InsertMessage(ctx, lc.ConversationID, localMessage)
			if err != nil {
				return nil, err
			}
			err = c.db.InsertSendingMessage(ctx, &model_struct.LocalSendingMessages{
				ConversationID: lc.ConversationID,
				ClientMsgID:    localMessage.ClientMsgID,
			})
			if err != nil {
				return nil, err
			}
		} else {
			if oldMessage.Status != constant.MsgStatusSendFailed {
				return nil, sdkerrs.ErrMsgRepeated
			} else {
				s.Status = constant.MsgStatusSending
				err = c.db.InsertSendingMessage(ctx, &model_struct.LocalSendingMessages{
					ConversationID: lc.ConversationID,
					ClientMsgID:    s.ClientMsgID,
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}
	lc.LatestMsg = c.latestMsgJSON(ctx, s)
	var delFile []string
	switch s.ContentType {
	case constant.Picture:
		s.Content = utils.StructToJsonString(s.PictureElem)
	case constant.Sound:
		s.Content = utils.StructToJsonString(s.SoundElem)
	case constant.Video:
		s.Content = utils.StructToJsonString(s.VideoElem)
	case constant.File:
		s.Content = utils.StructToJsonString(s.FileElem)
	case constant.Text:
		s.Content = utils.StructToJsonString(s.TextElem)
	case constant.AtText:
		s.Content = utils.StructToJsonString(s.AtTextElem)
	case constant.Location:
		s.Content = utils.StructToJsonString(s.LocationElem)
	case constant.Custom:
		s.Content = utils.StructToJsonString(s.CustomElem)
	case constant.Merger:
		s.Content = utils.StructToJsonString(s.MergeElem)
	case constant.Quote:
		s.Content = utils.StructToJsonString(s.QuoteElem)
	case constant.Card:
		s.Content = utils.StructToJsonString(s.CardElem)
	case constant.Face:
		s.Content = utils.StructToJsonString(s.FaceElem)
	case constant.AdvancedText:
		s.Content = utils.StructToJsonString(s.AdvancedTextElem)
	default:
		return nil, sdkerrs.ErrMsgContentTypeNotSupport
	}
	if utils.IsContainInt(int(s.ContentType), []int{constant.Picture, constant.Sound, constant.Video, constant.File}) {
		if isOnlineOnly {
			localMessage := MsgStructToLocalChatLog(s)
			err = c.db.UpdateMessage(ctx, lc.ConversationID, localMessage)
			if err != nil {
				return nil, err
			}
		}
	}
	return c.sendMessageToServer(ctx, s, lc, callback, delFile, p, options, isOnlineOnly)
}

func (c *Conversation) sendMessageToServer(ctx context.Context, s *sdk_struct.MsgStruct, lc *model_struct.LocalConversation, callback open_im_sdk_callback.SendMsgCallBack,
	delFiles []string, offlinePushInfo *sdkws.OfflinePushInfo, options map[string]bool, isOnlineOnly bool) (*sdk_struct.MsgStruct, error) {
	if isOnlineOnly {
		utils.SetSwitchFromOptions(options, constant.IsHistory, false)
		utils.SetSwitchFromOptions(options, constant.IsPersistent, false)
		utils.SetSwitchFromOptions(options, constant.IsSenderSync, false)
		utils.SetSwitchFromOptions(options, constant.IsConversationUpdate, false)
		utils.SetSwitchFromOptions(options, constant.IsSenderConversationUpdate, false)
		utils.SetSwitchFromOptions(options, constant.IsUnreadCount, false)
		utils.SetSwitchFromOptions(options, constant.IsOfflinePush, false)
	}
	//Protocol conversion
	var wsMsgData sdkws.MsgData
	copier.Copy(&wsMsgData, s)
	wsMsgData.AttachedInfo = utils.StructToJsonString(s.AttachedInfoElem)
	wsMsgData.Content = []byte(s.Content)
	wsMsgData.CreateTime = s.CreateTime
	wsMsgData.SendTime = 0
	wsMsgData.Options = options
	if wsMsgData.ContentType == constant.AtText {
		wsMsgData.AtUserIDList = s.AtTextElem.AtUserList
	}
	wsMsgData.OfflinePushInfo = offlinePushInfo
	s.Content = ""
	var sendMsgResp sdkws.UserSendMsgResp

	err := c.LongConnMgr.SendReqWaitResp(ctx, &wsMsgData, constant.SendMsg, &sendMsgResp)
	if err != nil {
		//if send message network timeout need to double-check message has received by db.
		if sdkerrs.ErrNetworkTimeOut.Is(err) && !isOnlineOnly {
			oldMessage, _ := c.db.GetMessage(ctx, lc.ConversationID, s.ClientMsgID)
			if oldMessage.Status == constant.MsgStatusSendSuccess {
				sendMsgResp.SendTime = oldMessage.SendTime
				sendMsgResp.ClientMsgID = oldMessage.ClientMsgID
				sendMsgResp.ServerMsgID = oldMessage.ServerMsgID
			} else {
				log.ZError(ctx, "send msg to server failed", err, "message", s)
				c.updateMsgStatusAndTriggerConversation(ctx, s.ClientMsgID, "", s.CreateTime,
					constant.MsgStatusSendFailed, s, lc, isOnlineOnly)
				return s, err
			}
		} else {
			log.ZError(ctx, "send msg to server failed", err, "message", s)
			c.updateMsgStatusAndTriggerConversation(ctx, s.ClientMsgID, "", s.CreateTime,
				constant.MsgStatusSendFailed, s, lc, isOnlineOnly)
			return s, err
		}
	}
	s.SendTime = sendMsgResp.SendTime
	s.Status = constant.MsgStatusSendSuccess
	s.ServerMsgID = sendMsgResp.ServerMsgID
	go func() {
		//remove media cache file
		for _, file := range delFiles {
			err := os.Remove(file)
			if err != nil {
				log.ZError(ctx, "delete temp File is failed", err, "filePath", file)
			}
			// log.ZDebug(ctx, "remove temp file:", "file", file)
		}

		c.updateMsgStatusAndTriggerConversation(ctx, sendMsgResp.ClientMsgID, sendMsgResp.ServerMsgID, sendMsgResp.SendTime, constant.MsgStatusSendSuccess, s, lc, isOnlineOnly)
	}()
	return s, nil

}

func (c *Conversation) FindMessageList(ctx context.Context, req []*sdk_params_callback.ConversationArgs) (*sdk_params_callback.FindMessageListCallback, error) {
	var r sdk_params_callback.FindMessageListCallback
	type tempConversationAndMessageList struct {
		conversation *model_struct.LocalConversation
		msgIDList    []string
	}
	var s []*tempConversationAndMessageList
	for _, conversationsArgs := range req {
		localConversation, err := c.db.GetConversation(ctx, conversationsArgs.ConversationID)
		if err != nil {
			log.ZError(ctx, "GetConversation err", err, "conversationsArgs", conversationsArgs)
		} else {
			t := new(tempConversationAndMessageList)
			t.conversation = localConversation
			t.msgIDList = conversationsArgs.ClientMsgIDList
			s = append(s, t)
		}
	}
	for _, v := range s {
		messages, err := c.db.GetMessagesByClientMsgIDs(ctx, v.conversation.ConversationID, v.msgIDList)
		if err == nil {
			var tempMessageList []*sdk_struct.MsgStruct
			for _, message := range messages {
				temp := LocalChatLogToMsgStruct(message)
				tempMessageList = append(tempMessageList, temp)
			}
			findResultItem := sdk_params_callback.SearchByConversationResult{}
			findResultItem.ConversationID = v.conversation.ConversationID
			findResultItem.FaceURL = v.conversation.FaceURL
			findResultItem.ShowName = v.conversation.ShowName
			findResultItem.ConversationType = v.conversation.ConversationType
			findResultItem.MessageList = tempMessageList
			findResultItem.MessageCount = len(findResultItem.MessageList)
			r.FindResultItems = append(r.FindResultItems, &findResultItem)
			r.TotalCount += findResultItem.MessageCount
		} else {
			log.ZError(ctx, "GetMessagesByClientMsgIDs err", err, "conversationID", v.conversation.ConversationID, "msgIDList", v.msgIDList)
		}
	}
	return &r, nil

}

func (c *Conversation) GetAdvancedHistoryMessageList(ctx context.Context, req sdk_params_callback.GetAdvancedHistoryMessageListParams) (*sdk_params_callback.GetAdvancedHistoryMessageListCallback, error) {
	result, err := c.getAdvancedHistoryMessageList(ctx, req, false)
	if err != nil {
		return nil, err
	}
	if len(result.MessageList) == 0 {
		s := make([]*sdk_struct.MsgStruct, 0)
		result.MessageList = s
	}
	return result, nil
}

func (c *Conversation) GetAdvancedHistoryMessageListReverse(ctx context.Context, req sdk_params_callback.GetAdvancedHistoryMessageListParams) (*sdk_params_callback.GetAdvancedHistoryMessageListCallback, error) {
	result, err := c.getAdvancedHistoryMessageList(ctx, req, true)
	if err != nil {
		return nil, err
	}
	if len(result.MessageList) == 0 {
		s := make([]*sdk_struct.MsgStruct, 0)
		result.MessageList = s
	}
	return result, nil
}

func (c *Conversation) RevokeMessage(ctx context.Context, conversationID, clientMsgID string) error {
	return c.revokeOneMessage(ctx, conversationID, clientMsgID)
}

// ReportSpam 向服务端提交垃圾消息/用户举报；reporter 由网关从 token 解析，无需在 req 中填写。
func (c *Conversation) ReportSpam(ctx context.Context, req *pbMsg.ReportSpamReq) (*pbMsg.ReportSpamResp, error) {
	if req == nil {
		return nil, sdkerrs.ErrArgs.WrapMsg("req is nil")
	}
	if err := req.Check(); err != nil {
		return nil, sdkerrs.ErrArgs.WrapMsg(err.Error())
	}
	return c.reportSpamToServer(ctx, req)
}

// SendPaymentNotification 向单个用户发送钱包/支付通知（POST /msg/send_payment_notification）。
// sendUserID 须为已注册的通知账号；使用当前登录用户 Token 即可。
func (c *Conversation) SendPaymentNotification(ctx context.Context, req *api.SendPaymentNotificationReq) (*pbMsg.SendMsgResp, error) {
	if req == nil {
		return nil, sdkerrs.ErrArgs.WrapMsg("req is nil")
	}
	if req.SendUserID == "" || req.RecvUserID == "" {
		return nil, sdkerrs.ErrArgs.WrapMsg("sendUserID and recvUserID are required")
	}
	content := req.Content
	if content.Title == "" || content.Amount == "" || content.TransactionType == "" || content.TransactionTime == "" || content.Currency == "" {
		return nil, sdkerrs.ErrArgs.WrapMsg("content title, amount, transactionType, transactionTime and currency are required")
	}
	return c.sendPaymentNotificationToServer(ctx, req)
}

func (c *Conversation) NotifyRedPacketClaimed(ctx context.Context, req *api.WalletDualPartyNotifyReq) (*api.WalletNotifyResp, error) {
	if err := validateWalletDualPartyNotifyReq(req); err != nil {
		return nil, err
	}
	return api.NotifyRedPacketClaimed.Invoke(ctx, req)
}

func (c *Conversation) NotifyTransferReceived(ctx context.Context, req *api.WalletDualPartyNotifyReq) (*api.WalletNotifyResp, error) {
	if err := validateWalletDualPartyNotifyReq(req); err != nil {
		return nil, err
	}
	return api.NotifyTransferReceived.Invoke(ctx, req)
}

func (c *Conversation) NotifyRedPacketExpired(ctx context.Context, req *api.WalletExpiredNotifyReq) (*api.WalletNotifyResp, error) {
	if err := validateWalletExpiredNotifyReq(req); err != nil {
		return nil, err
	}
	return api.NotifyRedPacketExpired.Invoke(ctx, req)
}

func (c *Conversation) NotifyTransferExpired(ctx context.Context, req *api.WalletExpiredNotifyReq) (*api.WalletNotifyResp, error) {
	if err := validateWalletExpiredNotifyReq(req); err != nil {
		return nil, err
	}
	return api.NotifyTransferExpired.Invoke(ctx, req)
}

// GetLinkPreview fetches Open Graph / meta preview for a URL via POST /link/preview.
func (c *Conversation) GetLinkPreview(ctx context.Context, req *api.LinkPreviewReq) (*api.LinkPreviewResp, error) {
	if req == nil {
		return nil, sdkerrs.ErrArgs.WrapMsg("req is nil")
	}
	if strings.TrimSpace(req.URL) == "" {
		return nil, sdkerrs.ErrArgs.WrapMsg("url is required")
	}
	return api.LinkPreview.Invoke(ctx, req)
}

func validateWalletDualPartyNotifyReq(req *api.WalletDualPartyNotifyReq) error {
	if req == nil {
		return sdkerrs.ErrArgs.WrapMsg("req is nil")
	}
	if req.SenderUserID == "" || req.ReceiverUserID == "" || req.BizID == "" {
		return sdkerrs.ErrArgs.WrapMsg("senderUserID, receiverUserID and bizID are required")
	}
	if req.ReceiverText == "" || req.SenderText == "" {
		return sdkerrs.ErrArgs.WrapMsg("receiverText and senderText are required")
	}
	return nil
}

func validateWalletExpiredNotifyReq(req *api.WalletExpiredNotifyReq) error {
	if req == nil {
		return sdkerrs.ErrArgs.WrapMsg("req is nil")
	}
	if req.SenderUserID == "" || req.BizID == "" {
		return sdkerrs.ErrArgs.WrapMsg("senderUserID and bizID are required")
	}
	if req.Text == "" {
		return sdkerrs.ErrArgs.WrapMsg("text is required")
	}
	return nil
}

func (c *Conversation) TypingStatusUpdate(ctx context.Context, recvID, msgTip string) error {
	return c.typingStatusUpdate(ctx, recvID, msgTip)
}

func (c *Conversation) MarkConversationMessageAsRead(ctx context.Context, conversationID string) error {
	return c.markConversationMessageAsRead(ctx, conversationID)
}

func (c *Conversation) MarkAllConversationMessageAsRead(ctx context.Context) error {
	conversationIDs, err := c.db.FindAllUnreadConversationConversationID(ctx)
	if err != nil {
		return err
	}
	for _, conversationID := range conversationIDs {
		if err = c.markConversationMessageAsRead(ctx, conversationID); err != nil {
			return err
		}
	}
	return nil
}

// deprecated
func (c *Conversation) MarkMessagesAsReadByMsgID(ctx context.Context, conversationID string, clientMsgIDs []string) error {
	return c.markMessagesAsReadByMsgID(ctx, conversationID, clientMsgIDs)
}

func (c *Conversation) MarkGroupMsgsAsRead(ctx context.Context, conversationID string, seqs []int64) error {
	if conversationID == "" {
		return sdkerrs.ErrArgs.WrapMsg("conversationID is empty")
	}
	if len(seqs) == 0 {
		return sdkerrs.ErrArgs.WrapMsg("seqs is empty")
	}
	return c.markGroupMsgsAsRead2Server(ctx, conversationID, seqs)
}

func (c *Conversation) DeleteMessageFromLocalStorage(ctx context.Context, conversationID string, clientMsgID string) error {
	return c.deleteMessageFromLocal(ctx, conversationID, clientMsgID)
}

func (c *Conversation) DeleteMessage(ctx context.Context, conversationID string, clientMsgID string) error {
	return c.deleteMessage(ctx, conversationID, clientMsgID, nil)
}

func (c *Conversation) DeleteMessageWithOptions(ctx context.Context, req *sdk_params_callback.DeleteMessageReq) error {
	if req == nil {
		return sdkerrs.ErrArgs.WrapMsg("DeleteMessageReq is nil")
	}
	var opt *pbMsg.DeleteSyncOpt
	if req.DeleteSyncOpt != nil {
		opt = &pbMsg.DeleteSyncOpt{
			IsSyncSelf:  req.DeleteSyncOpt.IsSyncSelf,
			IsSyncOther: req.DeleteSyncOpt.IsSyncOther,
		}
	}
	return c.deleteMessage(ctx, req.ConversationID, req.ClientMsgID, opt)
}

func (c *Conversation) DeleteAllMsgFromLocalAndServer(ctx context.Context) error {
	return c.deleteAllMsgFromLocalAndServer(ctx)
}

func (c *Conversation) DeleteAllMessageFromLocalStorage(ctx context.Context) error {
	return c.deleteAllMsgFromLocal(ctx, true)
}

func (c *Conversation) ClearConversationAndDeleteAllMsg(ctx context.Context, conversationID string) error {
	return c.clearConversationFromLocalAndServer(ctx, conversationID, c.db.ClearConversation)
}

func (c *Conversation) DeleteConversationAndDeleteAllMsg(ctx context.Context, conversationID string) error {
	return c.clearConversationFromLocalAndServer(ctx, conversationID, c.db.ResetConversation)
}

func (c *Conversation) InsertSingleMessageToLocalStorage(ctx context.Context, s *sdk_struct.MsgStruct, recvID, sendID string) (*sdk_struct.MsgStruct, error) {
	if recvID == "" || sendID == "" {
		return nil, sdkerrs.ErrArgs
	}
	var conversation model_struct.LocalConversation
	if sendID != c.loginUserID {
		faceUrl, name, err := c.getUserNameAndFaceURL(ctx, sendID)
		if err != nil {
			//log.Error(operationID, "GetUserNameAndFaceURL err", err.Error(), sendID)
		}
		s.SenderFaceURL = faceUrl
		s.SenderNickname = name
		conversation.FaceURL = faceUrl
		conversation.ShowName = name
		conversation.UserID = sendID
		conversation.ConversationID = c.getConversationIDBySessionType(sendID, constant.SingleChatType)
		log.ZInfo(ctx, " InsertSingleMessageToLocalStorage", "conversation", conversation)
	} else {
		conversation.UserID = recvID
		conversation.ConversationID = c.getConversationIDBySessionType(recvID, constant.SingleChatType)
		_, err := c.db.GetConversation(ctx, conversation.ConversationID)
		if err != nil {
			faceUrl, name, err := c.getUserNameAndFaceURL(ctx, recvID)
			if err != nil {
				return nil, err
			}
			conversation.FaceURL = faceUrl
			conversation.ShowName = name
			log.ZInfo(ctx, " InsertSingleMessageToLocalStorage", "conversation", conversation)

		}
	}

	s.SendID = sendID
	s.RecvID = recvID
	s.ClientMsgID = utils.GetMsgID(s.SendID)
	s.SendTime = utils.GetCurrentTimestampByMill()
	s.SessionType = constant.SingleChatType
	s.Status = constant.MsgStatusSendSuccess
	localMessage := MsgStructToLocalChatLog(s)
	conversation.LatestMsg = c.latestMsgJSON(ctx, s)
	conversation.ConversationType = constant.SingleChatType
	conversation.LatestMsgSendTime = s.SendTime
	err := c.insertMessageToLocalStorage(ctx, conversation.ConversationID, localMessage)
	if err != nil {
		return nil, err
	}
	_ = common.TriggerCmdUpdateConversation(ctx, common.UpdateConNode{ConID: conversation.ConversationID, Action: constant.AddConOrUpLatMsg, Args: conversation}, c.GetCh())
	return s, nil

}

func (c *Conversation) InsertGroupMessageToLocalStorage(ctx context.Context, s *sdk_struct.MsgStruct, groupID, sendID string) (*sdk_struct.MsgStruct, error) {
	if groupID == "" || sendID == "" {
		return nil, sdkerrs.ErrArgs
	}
	var conversation model_struct.LocalConversation
	var err error
	_, conversation.ConversationType, err = c.getConversationTypeByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	conversation.ConversationID = c.getConversationIDBySessionType(groupID, int(conversation.ConversationType))
	if sendID != c.loginUserID {
		faceUrl, name, err := c.getUserNameAndFaceURL(ctx, sendID)
		if err != nil {
			// log.Error("", "getUserNameAndFaceUrlByUid err", err.Error(), sendID)
		}
		s.SenderFaceURL = faceUrl
		s.SenderNickname = name
	}
	s.SendID = sendID
	s.RecvID = groupID
	s.GroupID = groupID
	s.ClientMsgID = utils.GetMsgID(s.SendID)
	s.SendTime = utils.GetCurrentTimestampByMill()
	s.SessionType = conversation.ConversationType
	s.Status = constant.MsgStatusSendSuccess
	localMessage := MsgStructToLocalChatLog(s)
	conversation.LatestMsg = c.latestMsgJSON(ctx, s)
	conversation.LatestMsgSendTime = s.SendTime
	conversation.FaceURL = s.SenderFaceURL
	conversation.ShowName = s.SenderNickname
	log.ZInfo(ctx, " InsertGroupMessageToLocalStorage", "conversation", conversation)
	err = c.insertMessageToLocalStorage(ctx, conversation.ConversationID, localMessage)
	if err != nil {
		return nil, err
	}
	_ = common.TriggerCmdUpdateConversation(ctx, common.UpdateConNode{ConID: conversation.ConversationID, Action: constant.AddConOrUpLatMsg, Args: conversation}, c.GetCh())
	return s, nil

}

func (c *Conversation) SearchLocalMessages(ctx context.Context, searchParam *sdk_params_callback.SearchLocalMessagesParams) (*sdk_params_callback.SearchLocalMessagesCallback, error) {
	searchParam.KeywordList = utils.TrimStringList(searchParam.KeywordList)
	return c.searchLocalMessages(ctx, searchParam)

}
func (c *Conversation) SetMessageLocalEx(ctx context.Context, conversationID string, clientMsgID string, localEx string) error {
	err := c.db.UpdateColumnsMessage(ctx, conversationID, clientMsgID, map[string]interface{}{"local_ex": localEx})
	if err != nil {
		return err
	}
	conversation, err := c.db.GetConversation(ctx, conversationID)
	if err != nil {
		return err
	}
	var latestMsg sdk_struct.MsgStruct
	utils.JsonStringToStruct(conversation.LatestMsg, &latestMsg)
	if latestMsg.ClientMsgID == clientMsgID {
		log.ZDebug(ctx, " latestMsg local ex changed", "seq", latestMsg.Seq, "clientMsgID", latestMsg.ClientMsgID)
		latestMsg.LocalEx = localEx
		latestMsgStr := c.latestMsgJSON(ctx, &latestMsg)
		if err = c.db.UpdateColumnsConversation(ctx, conversationID, map[string]interface{}{"latest_msg": latestMsgStr, "latest_msg_send_time": latestMsg.SendTime}); err != nil {
			return err
		}
		c.doUpdateConversation(common.Cmd2Value{Value: common.UpdateConNode{Action: constant.ConChange, Args: []string{conversationID}}})
	}
	return nil
}

func (c *Conversation) initBasicInfo(ctx context.Context, message *sdk_struct.MsgStruct, msgFrom, contentType int32) error {
	message.CreateTime = utils.GetCurrentTimestampByMill()
	message.SendTime = message.CreateTime
	message.IsRead = false
	message.Status = constant.MsgStatusSending
	message.SendID = c.loginUserID
	userInfo, err := c.user.GetUserInfoWithCache(ctx, c.loginUserID)
	if err != nil {
		return err
	}
	message.SenderFaceURL = userInfo.FaceURL
	message.SenderNickname = model_struct.UserDisplayName(userInfo.FirstName, userInfo.LastName, userInfo.Nickname)
	ClientMsgID := utils.GetMsgID(message.SendID)
	message.ClientMsgID = ClientMsgID
	message.MsgFrom = msgFrom
	message.ContentType = contentType
	message.SenderPlatformID = c.platformID
	return nil
}

func (c *Conversation) getConversationTypeByGroupID(ctx context.Context, groupID string) (conversationID string, conversationType int32, err error) {
	g, err := c.group.FetchGroupOrError(ctx, groupID)
	if err != nil {
		return "", 0, errs.WrapMsg(err, "get group info error")
	}
	switch g.GroupType {
	case constant.NormalGroup:
		return c.getConversationIDBySessionType(groupID, constant.WriteGroupChatType), constant.WriteGroupChatType, nil
	case constant.SuperGroup, constant.WorkingGroup:
		return c.getConversationIDBySessionType(groupID, constant.ReadGroupChatType), constant.ReadGroupChatType, nil
	default:
		return "", 0, sdkerrs.ErrGroupType
	}
}

func conversationReqToUpdateMap(req *pbConversation.ConversationReq) map[string]interface{} {
	if req == nil {
		return nil
	}
	m := make(map[string]interface{})
	if req.RecvMsgOpt != nil {
		m["recv_msg_opt"] = req.RecvMsgOpt.Value
	}
	if req.IsPinned != nil {
		m["is_pinned"] = req.IsPinned.Value
	}
	if req.IsPrivateChat != nil {
		m["is_private_chat"] = req.IsPrivateChat.Value
	}
	if req.BurnDuration != nil {
		m["burn_duration"] = req.BurnDuration.Value
	}
	if req.GroupAtType != nil {
		m["group_at_type"] = req.GroupAtType.Value
	}
	if req.AttachedInfo != nil {
		m["attached_info"] = req.AttachedInfo.Value
	}
	if req.Ex != nil {
		m["ex"] = req.Ex.Value
	}
	if req.MsgDestructTime != nil {
		m["msg_destruct_time"] = req.MsgDestructTime.Value
	}
	if req.IsMsgDestruct != nil {
		m["is_msg_destruct"] = req.IsMsgDestruct.Value
	}
	if req.MuteDuration != nil {
		m["mute_duration"] = req.MuteDuration.Value
	}
	if req.MuteEndTime != nil {
		m["mute_end_time"] = req.MuteEndTime.Value
	}
	if req.MinSeq != nil {
		m["min_seq"] = req.MinSeq.Value
	}
	if req.MaxSeq != nil {
		m["max_seq"] = req.MaxSeq.Value
	}
	return m
}

func applyConversationReqToLocal(lc *model_struct.LocalConversation, req *pbConversation.ConversationReq) {
	if req == nil || lc == nil {
		return
	}
	if req.RecvMsgOpt != nil {
		lc.RecvMsgOpt = req.RecvMsgOpt.Value
	}
	if req.IsPinned != nil {
		lc.IsPinned = req.IsPinned.Value
	}
	if req.IsPrivateChat != nil {
		lc.IsPrivateChat = req.IsPrivateChat.Value
	}
	if req.BurnDuration != nil {
		lc.BurnDuration = req.BurnDuration.Value
	}
	if req.GroupAtType != nil {
		lc.GroupAtType = req.GroupAtType.Value
	}
	if req.AttachedInfo != nil {
		lc.AttachedInfo = req.AttachedInfo.Value
	}
	if req.Ex != nil {
		lc.Ex = req.Ex.Value
	}
	if req.MsgDestructTime != nil {
		lc.MsgDestructTime = req.MsgDestructTime.Value
	}
	if req.IsMsgDestruct != nil {
		lc.IsMsgDestruct = req.IsMsgDestruct.Value
	}
	if req.MuteDuration != nil {
		lc.MuteDuration = req.MuteDuration.Value
	}
	if req.MuteEndTime != nil {
		lc.MuteEndTime = req.MuteEndTime.Value
	}
	if req.MinSeq != nil {
		lc.MinSeq = req.MinSeq.Value
	}
	if req.MaxSeq != nil {
		lc.MaxSeq = req.MaxSeq.Value
	}
}

func (c *Conversation) SearchConversation(ctx context.Context, searchParam string) ([]*server_api_params.Conversation, error) {
	// Check if search parameter is empty
	if searchParam == "" {
		return nil, sdkerrs.ErrArgs.WrapMsg("search parameter cannot be empty")
	}

	// Perform the search in your database or data source
	// This is a placeholder for the actual database call
	conversations, err := c.db.SearchConversations(ctx, searchParam)
	if err != nil {
		// Handle any errors that occurred during the search
		return nil, err
	}
	apiConversations := make([]*server_api_params.Conversation, len(conversations))
	for i, localConv := range conversations {
		// Create new server_api_params.Conversation and map fields from localConv
		apiConv := &server_api_params.Conversation{
			ConversationID:        localConv.ConversationID,
			ConversationType:      localConv.ConversationType,
			UserID:                localConv.UserID,
			GroupID:               localConv.GroupID,
			RecvMsgOpt:            localConv.RecvMsgOpt,
			UnreadCount:           localConv.UnreadCount,
			DraftTextTime:         localConv.DraftTextTime,
			IsPinned:              localConv.IsPinned,
			IsPrivateChat:         localConv.IsPrivateChat,
			BurnDuration:          localConv.BurnDuration,
			GroupAtType:           localConv.GroupAtType,
			IsNotInGroup:          localConv.IsNotInGroup,
			UpdateUnreadCountTime: localConv.UpdateUnreadCountTime,
			AttachedInfo:          localConv.AttachedInfo,
			Ex:                    localConv.Ex,
		}
		apiConversations[i] = apiConv
	}
	// Return the list of conversations
	return apiConversations, nil
}
func (c *Conversation) GetInputStates(ctx context.Context, conversationID string, userID string) ([]int32, error) {
	return c.typing.GetInputStates(conversationID, userID), nil
}

func (c *Conversation) ChangeInputStates(ctx context.Context, conversationID string, focus bool) error {
	return c.typing.ChangeInputStates(ctx, conversationID, focus)
}
