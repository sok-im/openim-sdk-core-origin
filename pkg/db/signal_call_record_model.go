// Copyright © 2023 OpenIM SDK. All rights reserved.

//go:build !js
// +build !js

package db

import (
	"context"
	"errors"
	"strings"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/tools/errs"

	"gorm.io/gorm"
)

func (d *DataBase) BatchUpsertSignalCallRecords(ctx context.Context, records []*model_struct.LocalSignalCallRecord) error {
	if len(records) == 0 {
		return nil
	}
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	for _, r := range records {
		if r == nil || r.SID == "" {
			continue
		}
		if err := d.conn.WithContext(ctx).Save(r).Error; err != nil {
			return errs.WrapMsg(err, "BatchUpsertSignalCallRecords Save failed")
		}
	}
	return nil
}

func (d *DataBase) SearchSignalCallRecords(ctx context.Context, offset, count int, sessionType int32, dialStatus int32, direction int32, startTime, endTime int64, keyword, userName string) ([]*model_struct.LocalSignalCallRecord, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	tx := d.conn.WithContext(ctx).Model(&model_struct.LocalSignalCallRecord{})
	tx = applySignalCallFilters(tx, sessionType, dialStatus, direction, startTime, endTime, keyword, userName)
	var list []*model_struct.LocalSignalCallRecord
	err := tx.Order("create_time DESC").Offset(offset).Limit(count).Find(&list).Error
	return list, errs.WrapMsg(err, "SearchSignalCallRecords failed")
}

func (d *DataBase) CountSignalCallRecords(ctx context.Context, sessionType int32, dialStatus int32, direction int32, startTime, endTime int64, keyword, userName string) (int64, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	tx := d.conn.WithContext(ctx).Model(&model_struct.LocalSignalCallRecord{})
	tx = applySignalCallFilters(tx, sessionType, dialStatus, direction, startTime, endTime, keyword, userName)
	var n int64
	err := tx.Count(&n).Error
	return n, errs.WrapMsg(err, "CountSignalCallRecords failed")
}

func applySignalCallFilters(tx *gorm.DB, sessionType int32, dialStatus int32, direction int32, startTime, endTime int64, keyword, userName string) *gorm.DB {
	if sessionType != 0 {
		tx = tx.Where("session_type = ?", sessionType)
	}
	if dialStatus != 0 {
		tx = tx.Where("dial_status = ?", dialStatus)
	}
	if direction != 0 {
		tx = tx.Where("direction = ?", direction)
	}
	if startTime > 0 {
		tx = tx.Where("create_time >= ?", startTime)
	}
	if endTime > 0 {
		tx = tx.Where("create_time <= ?", endTime)
	}
	kw := strings.TrimSpace(keyword)
	if kw != "" {
		pattern := "%" + kw + "%"
		tx = tx.Where("room_id LIKE ? OR group_name LIKE ? OR inviter_user_nickname LIKE ? OR invitee_user_nickname LIKE ? OR group_id LIKE ? OR inviter_user_id LIKE ?",
			pattern, pattern, pattern, pattern, pattern, pattern)
	}
	un := strings.TrimSpace(userName)
	if un != "" {
		p := "%" + un + "%"
		tx = tx.Where("callee_match_text LIKE ?", p)
	}
	return tx
}

func (d *DataBase) GetSignalCallRecordBySID(ctx context.Context, sID string) (*model_struct.LocalSignalCallRecord, error) {
	if sID == "" {
		return nil, errs.ErrRecordNotFound.Wrap()
	}
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	var rec model_struct.LocalSignalCallRecord
	err := d.conn.WithContext(ctx).Where("s_id = ?", sID).Take(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrRecordNotFound.Wrap()
		}
		return nil, errs.WrapMsg(err, "GetSignalCallRecordBySID failed")
	}
	return &rec, nil
}

func (d *DataBase) DeleteSignalCallRecords(ctx context.Context, sIDs []string) error {
	if len(sIDs) == 0 {
		return nil
	}
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	return errs.WrapMsg(
		d.conn.WithContext(ctx).Where("s_id IN ?", sIDs).Delete(&model_struct.LocalSignalCallRecord{}).Error,
		"DeleteSignalCallRecords failed",
	)
}

func (d *DataBase) ClearAllSignalCallRecords(ctx context.Context) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	return errs.WrapMsg(
		d.conn.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model_struct.LocalSignalCallRecord{}).Error,
		"ClearAllSignalCallRecords failed",
	)
}
