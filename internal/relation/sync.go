package relation

import (
	"context"

	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/datautil"
)

func filterValidServerBlacks(serverData []*sdkws.BlackInfo) []*sdkws.BlackInfo {
	valid := make([]*sdkws.BlackInfo, 0, len(serverData))
	for _, info := range serverData {
		if isValidServerBlack(info) {
			valid = append(valid, info)
		}
	}
	return valid
}

func (r *Relation) SyncAllBlackList(ctx context.Context) error {
	serverData, err := r.getBlackList(ctx)
	if err != nil {
		return err
	}
	serverData = filterValidServerBlacks(serverData)
	log.ZDebug(ctx, "black from server", "data", serverData)
	localData, err := r.db.GetBlackListDB(ctx)
	if err != nil {
		return err
	}
	log.ZDebug(ctx, "black from local", "data", localData)
	return r.blackSyncer.Sync(ctx, datautil.Batch(ServerBlackToLocalBlack, serverData), localData, nil)
}

func (r *Relation) SyncAllBlackListWithoutNotice(ctx context.Context) error {
	serverData, err := r.getBlackList(ctx)
	if err != nil {
		return err
	}
	serverData = filterValidServerBlacks(serverData)
	log.ZDebug(ctx, "black from server", "data", serverData)
	localData, err := r.db.GetBlackListDB(ctx)
	if err != nil {
		return err
	}
	log.ZDebug(ctx, "black from local", "data", localData)
	return r.blackSyncer.Sync(ctx, datautil.Batch(ServerBlackToLocalBlack, serverData), localData, nil, false, true)
}
