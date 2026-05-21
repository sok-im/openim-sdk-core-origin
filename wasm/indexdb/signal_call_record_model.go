// Copyright © 2023 OpenIM SDK. All rights reserved.

//go:build js && wasm
// +build js,wasm

package indexdb

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
	"github.com/openimsdk/openim-sdk-core/v3/wasm/exec"
	"github.com/openimsdk/tools/errs"
)

type SignalCallRecords struct {
	loginUserID string
}

func NewSignalCallRecords(loginUserID string) *SignalCallRecords {
	return &SignalCallRecords{loginUserID: loginUserID}
}

func (i *SignalCallRecords) BatchUpsertSignalCallRecords(ctx context.Context, records []*model_struct.LocalSignalCallRecord) error {
	if len(records) == 0 {
		return nil
	}
	_, err := exec.Exec(utils.StructToJsonString(records))
	return err
}

func (i *SignalCallRecords) SearchSignalCallRecords(ctx context.Context, offset, count int, sessionType int32, status int32, direction int32, startTime, endTime int64, keyword, userName, inviteeNickname string) ([]*model_struct.LocalSignalCallRecord, error) {
	gList, err := exec.Exec(offset, count, sessionType, status, direction, startTime, endTime, keyword, userName, inviteeNickname, i.loginUserID)
	if err != nil {
		return nil, err
	}
	if v, ok := gList.(string); ok {
		var result []*model_struct.LocalSignalCallRecord
		if err := utils.JsonStringToStruct(v, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, exec.ErrType
}

func (i *SignalCallRecords) CountSignalCallRecords(ctx context.Context, sessionType int32, status int32, direction int32, startTime, endTime int64, keyword, userName, inviteeNickname string) (int64, error) {
	n, err := exec.Exec(sessionType, status, direction, startTime, endTime, keyword, userName, inviteeNickname, i.loginUserID)
	if err != nil {
		return 0, err
	}
	if v, ok := n.(float64); ok {
		return int64(v), nil
	}
	return 0, exec.ErrType
}

func (i *SignalCallRecords) SearchSignalCallRecordsByUser(ctx context.Context, userID string, status int32, offset, count int, startTime, endTime int64) ([]*model_struct.LocalSignalCallRecord, error) {
	gList, err := exec.Exec(userID, status, offset, count, startTime, endTime, i.loginUserID)
	if err != nil {
		return nil, err
	}
	if v, ok := gList.(string); ok {
		var result []*model_struct.LocalSignalCallRecord
		if err := utils.JsonStringToStruct(v, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, exec.ErrType
}

func (i *SignalCallRecords) CountSignalCallRecordsByUser(ctx context.Context, userID string, status int32, startTime, endTime int64) (int64, error) {
	n, err := exec.Exec(userID, status, startTime, endTime, i.loginUserID)
	if err != nil {
		return 0, err
	}
	if v, ok := n.(float64); ok {
		return int64(v), nil
	}
	return 0, exec.ErrType
}

func (i *SignalCallRecords) GetSignalCallRecordBySID(ctx context.Context, sID string) (*model_struct.LocalSignalCallRecord, error) {
	msg, err := exec.Exec(sID, i.loginUserID)
	if err != nil {
		return nil, err
	}
	if v, ok := msg.(string); ok {
		var rec model_struct.LocalSignalCallRecord
		if err := utils.JsonStringToStruct(v, &rec); err != nil {
			return nil, err
		}
		if rec.SID == "" {
			return nil, errs.ErrRecordNotFound.Wrap()
		}
		return &rec, nil
	}
	return nil, exec.ErrType
}

func (i *SignalCallRecords) DeleteSignalCallRecords(ctx context.Context, sIDs []string) error {
	if len(sIDs) == 0 {
		return nil
	}
	_, err := exec.Exec(utils.StructToJsonString(sIDs))
	return err
}

func (i *SignalCallRecords) ClearAllSignalCallRecords(ctx context.Context) error {
	_, err := exec.Exec()
	return err
}
