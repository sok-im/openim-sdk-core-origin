package group

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/protocol/group"
	"github.com/openimsdk/protocol/sdkws"
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

func (g *Group) enrichServerGroupInviteLinks(ctx context.Context, groups []*sdkws.GroupInfo) {
	for _, info := range groups {
		if info == nil {
			continue
		}
		if info.EnableInviteLink != 1 {
			info.InviteLink = nil
			continue
		}
		if len(info.InviteLink) > 0 {
			continue
		}
		resp, err := g.listGroupInviteLinks(ctx, &group.ListGroupInviteLinksReq{
			GroupID: info.GroupID,
			Pagination: &sdkws.RequestPagination{
				PageNumber: 1,
				ShowNumber: 100,
			},
		})
		if err != nil {
			continue
		}
		info.InviteLink = datautil.Batch(groupInviteLinkInfoToSDKWS, resp.GetLinks())
	}
}

func (g *Group) enrichLocalGroupInviteLinks(ctx context.Context, local *model_struct.LocalGroup) {
	if local == nil {
		return
	}
	if local.EnableInviteLink != 1 {
		local.InviteLink = nil
		return
	}
	if len(local.InviteLink) > 0 {
		return
	}
	resp, err := g.listGroupInviteLinks(ctx, &group.ListGroupInviteLinksReq{
		GroupID: local.GroupID,
		Pagination: &sdkws.RequestPagination{
			PageNumber: 1,
			ShowNumber: 100,
		},
	})
	if err != nil {
		return
	}
	local.InviteLink = datautil.Batch(groupInviteLinkInfoToSDKWS, resp.GetLinks())
}

func (g *Group) enrichLocalGroupsInviteLinks(ctx context.Context, groups []*model_struct.LocalGroup) {
	for _, local := range groups {
		g.enrichLocalGroupInviteLinks(ctx, local)
	}
}
