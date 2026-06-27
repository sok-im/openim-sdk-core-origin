package conversation_msg

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
)

// applyLoginUserBurnToSingleChatConversation 新发起单聊占位会话时，带入当前登录用户的全局阅后即焚设置。
func (c *Conversation) applyLoginUserBurnToSingleChatConversation(ctx context.Context, conv *model_struct.LocalConversation) {
	if conv == nil || conv.ConversationType != constant.SingleChatType {
		return
	}
	if conv.IsPrivateChat && conv.BurnDuration > 0 {
		return
	}
	loginUser, err := c.user.GetSelfUserInfo(ctx)
	if err != nil || loginUser == nil || loginUser.MsgBurnDuration <= 0 {
		return
	}
	conv.BurnDuration = loginUser.MsgBurnDuration
	conv.IsPrivateChat = true
}
