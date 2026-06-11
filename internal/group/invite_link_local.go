package group

import (
	"context"

	"github.com/google/go-cmp/cmp"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/protocol/group"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/datautil"
)

func groupInviteLinkInfoToSDKWS(link *group.GroupInviteLinkInfo) *sdkws.GroupInviteLinkInfo {
	if link == nil {
		return nil
	}
	return &sdkws.GroupInviteLinkInfo{
		LinkID:      link.LinkID,
		GroupID:     link.GroupID,
		CreatorID:   link.CreatorID,
		ExpireAt:    link.ExpireAt,
		MaxUseCount: link.MaxUseCount,
		UsedCount:   link.UsedCount,
		Revoked:     link.Revoked,
		CreatedAt:   link.CreatedAt,
		ShareURL:    link.ShareURL,
	}
}

func localGroupEqual(a, b *model_struct.LocalGroup) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	aCopy, bCopy := *a, *b
	aCopy.InviteLink, bCopy.InviteLink = nil, nil
	if !cmp.Equal(&aCopy, &bCopy) {
		return false
	}
	return inviteLinkEqual(a.InviteLink, b.InviteLink)
}

func inviteLinkEqual(a, b *sdkws.GroupInviteLinkInfo) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.LinkID == b.LinkID &&
		a.GroupID == b.GroupID &&
		a.CreatorID == b.CreatorID &&
		a.ExpireAt == b.ExpireAt &&
		a.MaxUseCount == b.MaxUseCount &&
		a.UsedCount == b.UsedCount &&
		a.Revoked == b.Revoked &&
		a.CreatedAt == b.CreatedAt &&
		a.ShareURL == b.ShareURL
}

func firstInviteLinkFromList(links []*group.GroupInviteLinkInfo) *sdkws.GroupInviteLinkInfo {
	if len(links) == 0 {
		return nil
	}
	return groupInviteLinkInfoToSDKWS(links[0])
}

// fetchGroupInviteLinkFromServer 通过 list_invite_links 从服务端拉取群内唯一邀请链接。
func (g *Group) fetchGroupInviteLinkFromServer(ctx context.Context, groupID string) *sdkws.GroupInviteLinkInfo {
	resp, err := g.listGroupInviteLinks(ctx, &group.ListGroupInviteLinksReq{
		GroupID: groupID,
		Pagination: &sdkws.RequestPagination{
			PageNumber: 1,
			ShowNumber: 1,
		},
	})
	if err != nil {
		log.ZDebug(ctx, "fetch group invite link failed", "groupID", groupID, "err", err)
		return nil
	}
	return firstInviteLinkFromList(resp.GetLinks())
}

func (g *Group) resolveGroupInviteLink(ctx context.Context, enableInviteLink int32, groupID string) *sdkws.GroupInviteLinkInfo {
	if enableInviteLink != 1 {
		return nil
	}
	return g.fetchGroupInviteLinkFromServer(ctx, groupID)
}

func (g *Group) applyInviteLinkToServerGroupInfo(ctx context.Context, info *sdkws.GroupInfo) {
	if info == nil {
		return
	}
	info.InviteLink = g.resolveGroupInviteLink(ctx, info.EnableInviteLink, info.GroupID)
}

func (g *Group) applyInviteLinkToLocalGroup(ctx context.Context, local *model_struct.LocalGroup) {
	if local == nil {
		return
	}
	local.InviteLink = g.resolveGroupInviteLink(ctx, local.EnableInviteLink, local.GroupID)
}

func (g *Group) enrichServerGroupInviteLinks(ctx context.Context, groups []*sdkws.GroupInfo) {
	for _, info := range groups {
		g.applyInviteLinkToServerGroupInfo(ctx, info)
	}
}

func (g *Group) enrichLocalGroupInviteLinks(ctx context.Context, local *model_struct.LocalGroup) {
	g.applyInviteLinkToLocalGroup(ctx, local)
}

func (g *Group) enrichLocalGroupsInviteLinks(ctx context.Context, groups []*model_struct.LocalGroup) {
	for _, local := range groups {
		g.applyInviteLinkToLocalGroup(ctx, local)
	}
}

// serverGroupInfosToLocalGroups 将服务端群资料转为本地模型，并同步邀请链接。
func (g *Group) serverGroupInfosToLocalGroups(ctx context.Context, infos []*sdkws.GroupInfo) []*model_struct.LocalGroup {
	g.enrichServerGroupInviteLinks(ctx, infos)
	return datautil.Batch(ServerGroupToLocalGroup, infos)
}

// updateLocalGroupInviteLink 将邀请链接写入本地群资料。
func (g *Group) updateLocalGroupInviteLink(ctx context.Context, groupID string, link *sdkws.GroupInviteLinkInfo) error {
	local, err := g.db.GetGroupInfoByGroupID(ctx, groupID)
	if err != nil {
		return err
	}
	local.InviteLink = link
	return g.db.UpdateGroup(ctx, local)
}

// refreshLocalGroupInviteLink 刷新本地库中指定群的邀请链接（从服务端拉取）。
func (g *Group) refreshLocalGroupInviteLink(ctx context.Context, groupID string) error {
	local, err := g.db.GetGroupInfoByGroupID(ctx, groupID)
	if err != nil {
		return err
	}
	g.applyInviteLinkToLocalGroup(ctx, local)
	return g.db.UpdateGroup(ctx, local)
}
