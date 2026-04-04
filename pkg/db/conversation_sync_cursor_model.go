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

//go:build !js
// +build !js

package db

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/tools/errs"
)

// BatchUpsertConversationSyncedMaxSeqs inserts or updates the synced-max-seq cursor
// for each normal conversation. GORM's Save performs an upsert on the primary key.
func (d *DataBase) BatchUpsertConversationSyncedMaxSeqs(ctx context.Context, seqs []*model_struct.LocalConversationSyncedMaxSeq) error {
	if len(seqs) == 0 {
		return nil
	}
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	return errs.WrapMsg(d.conn.WithContext(ctx).Save(seqs).Error, "BatchUpsertConversationSyncedMaxSeqs failed")
}

// GetAllConversationSyncedMaxSeqs returns all persisted sync cursors.
func (d *DataBase) GetAllConversationSyncedMaxSeqs(ctx context.Context) ([]*model_struct.LocalConversationSyncedMaxSeq, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	var seqs []*model_struct.LocalConversationSyncedMaxSeq
	return seqs, errs.WrapMsg(d.conn.WithContext(ctx).Find(&seqs).Error, "GetAllConversationSyncedMaxSeqs failed")
}
