// Copyright © 2026 OpenIM SDK. All rights reserved.
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

	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdkerrs"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
	pbmsg "github.com/openimsdk/protocol/msg"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/log"
)

// SetMessageReaction 设置/切换/取消对某条消息的表情反应（POST /msg/set_reactions）。
func (c *Conversation) SetMessageReaction(ctx context.Context, req *pbmsg.SetMessageReactionReq) (*pbmsg.SetMessageReactionResp, error) {
	if req == nil {
		return nil, sdkerrs.ErrArgs.WrapMsg("req is nil")
	}
	if err := req.Check(); err != nil {
		return nil, sdkerrs.ErrArgs.WrapMsg(err.Error())
	}
	return c.setMessageReactionToServer(ctx, req)
}

// GetMessageReactions 查询单条消息的反应聚合（POST /msg/get_reactions）。
func (c *Conversation) GetMessageReactions(ctx context.Context, req *pbmsg.GetMessageReactionsReq) (*pbmsg.GetMessageReactionsResp, error) {
	if req == nil {
		return nil, sdkerrs.ErrArgs.WrapMsg("req is nil")
	}
	if err := req.Check(); err != nil {
		return nil, sdkerrs.ErrArgs.WrapMsg(err.Error())
	}
	return c.getMessageReactionsFromServer(ctx, req)
}

// BatchGetMessageReactions 批量拉取会话内多条消息的反应聚合（POST /msg/batch_get_reactions）。
func (c *Conversation) BatchGetMessageReactions(ctx context.Context, req *pbmsg.BatchGetMessageReactionsReq) (*pbmsg.BatchGetMessageReactionsResp, error) {
	if req == nil {
		return nil, sdkerrs.ErrArgs.WrapMsg("req is nil")
	}
	if err := req.Check(); err != nil {
		return nil, sdkerrs.ErrArgs.WrapMsg(err.Error())
	}
	return c.batchGetMessageReactionsFromServer(ctx, req)
}

// doReactionUpdated 处理 MsgReactionUpdatedNotification(2103)：解析聚合快照并通过 AdvancedMsgListener.OnReactionUpdated 回调。
func (c *Conversation) doReactionUpdated(ctx context.Context, msg *sdkws.MsgData) error {
	var tips pbmsg.MessageReactionUpdatedTips
	if err := utils.UnmarshalNotificationElem(msg.Content, &tips); err != nil {
		log.ZWarn(ctx, "unmarshal MessageReactionUpdatedTips failed", err, "msg", msg)
		return errs.Wrap(err)
	}
	log.ZDebug(ctx, "do reaction updated", "tips", &tips)
	if c.msgListener() != nil {
		c.msgListener().OnReactionUpdated(utils.StructToJsonString(&tips))
	}
	return nil
}
