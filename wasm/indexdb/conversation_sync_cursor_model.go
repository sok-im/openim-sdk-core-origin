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

//go:build js && wasm
// +build js,wasm

package indexdb

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
	"github.com/openimsdk/openim-sdk-core/v3/wasm/exec"
)

type LocalConversationSyncCursor struct{}

func NewLocalConversationSyncCursor() *LocalConversationSyncCursor {
	return &LocalConversationSyncCursor{}
}

func (i *LocalConversationSyncCursor) BatchUpsertConversationSyncedMaxSeqs(ctx context.Context, seqs []*model_struct.LocalConversationSyncedMaxSeq) error {
	_, err := exec.Exec(utils.StructToJsonString(seqs))
	return err
}

func (i *LocalConversationSyncCursor) GetAllConversationSyncedMaxSeqs(ctx context.Context) ([]*model_struct.LocalConversationSyncedMaxSeq, error) {
	gList, err := exec.Exec()
	if err != nil {
		return nil, err
	}
	if v, ok := gList.(string); ok {
		var result []*model_struct.LocalConversationSyncedMaxSeq
		if err := utils.JsonStringToStruct(v, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, exec.ErrType
}
