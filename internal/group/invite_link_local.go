package group

import (
	"context"

	"github.com/google/go-cmp/cmp"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/protocol/group"
	"github.com/openimsdk/protocol/sdkws"
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

func (g *Group) enrichServerGroupInviteLinks(ctx context.Context, groups []*sdkws.GroupInfo) {
	for _, info := range groups {
		if info == nil {
			continue
		}
		if info.EnableInviteLink != 1 {
			info.InviteLink = nil
			continue
		}
		if info.InviteLink != nil {
			continue
		}
		resp, err := g.getGroupInviteLinkByGroupID(ctx, &group.GetGroupInviteLinkByGroupIDReq{
			GroupID: info.GroupID,
		})
		if err != nil {
			continue
		}
		info.InviteLink = groupInviteLinkInfoToSDKWS(resp.GetLink())
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
	if local.InviteLink != nil {
		return
	}
	resp, err := g.getGroupInviteLinkByGroupID(ctx, &group.GetGroupInviteLinkByGroupIDReq{
		GroupID: local.GroupID,
	})
	if err != nil {
		return
	}
	local.InviteLink = groupInviteLinkInfoToSDKWS(resp.GetLink())
}

func (g *Group) enrichLocalGroupsInviteLinks(ctx context.Context, groups []*model_struct.LocalGroup) {
	for _, local := range groups {
		g.enrichLocalGroupInviteLinks(ctx, local)
	}
}
