// Copyright © 2023 OpenIM SDK. All rights reserved.

//go:build !js
// +build !js

package db

import (
	"context"
	"errors"
	"strings"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
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
		r.OwnerUserID = d.loginUserID
		if err := d.signalDB().WithContext(ctx).Save(r).Error; err != nil {
			return errs.WrapMsg(err, "BatchUpsertSignalCallRecords Save failed")
		}
	}
	return nil
}

func applySignalCallOwnerFilter(tx *gorm.DB, ownerUserID string) *gorm.DB {
	ownerUserID = strings.TrimSpace(ownerUserID)
	if ownerUserID == "" {
		return tx
	}
	return tx.Where("owner_user_id = ?", ownerUserID)
}

func (d *DataBase) SearchSignalCallRecords(ctx context.Context, offset, count int, sessionType int32, status int32, direction int32, startTime, endTime int64, keyword, userName, inviteeNickname, inviterUserID, peerUserID string) ([]*model_struct.LocalSignalCallRecord, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	tx := d.signalDB().WithContext(ctx).Model(&model_struct.LocalSignalCallRecord{})
	tx = applySignalCallOwnerFilter(tx, d.loginUserID)
	tx = applySignalCallFilters(tx, sessionType, status, direction, startTime, endTime, keyword, userName, inviteeNickname, inviterUserID, peerUserID)
	var list []*model_struct.LocalSignalCallRecord
	err := tx.Order("create_time DESC").Offset(offset).Limit(count).Find(&list).Error
	return list, errs.WrapMsg(err, "SearchSignalCallRecords failed")
}

func (d *DataBase) CountSignalCallRecords(ctx context.Context, sessionType int32, status int32, direction int32, startTime, endTime int64, keyword, userName, inviteeNickname, inviterUserID, peerUserID string) (int64, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	tx := d.signalDB().WithContext(ctx).Model(&model_struct.LocalSignalCallRecord{})
	tx = applySignalCallOwnerFilter(tx, d.loginUserID)
	tx = applySignalCallFilters(tx, sessionType, status, direction, startTime, endTime, keyword, userName, inviteeNickname, inviterUserID, peerUserID)
	var n int64
	err := tx.Count(&n).Error
	return n, errs.WrapMsg(err, "CountSignalCallRecords failed")
}

func applySignalCallFilters(tx *gorm.DB, sessionType int32, status int32, direction int32, startTime, endTime int64, keyword, userName, inviteeNickname, inviterUserID, peerUserID string) *gorm.DB {
	if sessionType != 0 {
		tx = tx.Where("session_type = ?", sessionType)
	} else {
		tx = applyExcludeGroupCallFilter(tx)
	}
	if status != 0 {
		tx = tx.Where("status = ?", status)
	}
	if direction != 0 {
		tx = tx.Where("direction = ?", direction)
	}
	if uid := strings.TrimSpace(inviterUserID); uid != "" {
		tx = tx.Where("inviter_user_id = ?", uid)
	}
	if peer := strings.TrimSpace(peerUserID); peer != "" {
		pattern := "%\"" + peer + "\"%"
		tx = tx.Where("inviter_user_id = ? OR invitee_user_ids LIKE ?", peer, pattern)
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
		tx = tx.Where("room_id LIKE ? OR group_name LIKE ? OR inviter_user_nickname LIKE ? OR invitee_user_nickname LIKE ? OR group_id LIKE ? OR inviter_user_id LIKE ? OR inviter_user_face_url LIKE ?",
			pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	}
	tx = applySignalCallUserNameFilter(tx, userName)
	in := strings.TrimSpace(inviteeNickname)
	if in != "" {
		p := "%" + in + "%"
		tx = tx.Where("invitee_user_nickname LIKE ?", p)
	}
	return tx
}

// applyExcludeGroupCallFilter 排除群通话（与 signaling.isGroupChatCall 对齐：group_id 非空或群聊 session_type）。
func applyExcludeGroupCallFilter(tx *gorm.DB) *gorm.DB {
	return tx.Where(
		"(group_id = '' OR group_id IS NULL) AND session_type NOT IN (?, ?)",
		constant.WriteGroupChatType, constant.ReadGroupChatType,
	)
}

// applySignalCallUserNameFilter 按用户名模糊匹配；多词（如 firstName + lastName）时要求每个词均命中。
func applySignalCallUserNameFilter(tx *gorm.DB, userName string) *gorm.DB {
	un := strings.TrimSpace(userName)
	if un == "" {
		return tx
	}
	tokens := strings.Fields(un)
	if len(tokens) == 0 {
		return tx
	}
	for _, token := range tokens {
		p := "%" + token + "%"
		tx = tx.Where(
			"inviter_user_nickname LIKE ? OR invitee_user_nickname LIKE ? OR callee_match_text LIKE ?",
			p, p, p,
		)
	}
	return tx
}

func (d *DataBase) SearchSignalCallRecordsByUser(ctx context.Context, userID string, status int32, offset, count int, startTime, endTime int64) ([]*model_struct.LocalSignalCallRecord, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	tx := d.signalDB().WithContext(ctx).Model(&model_struct.LocalSignalCallRecord{})
	tx = applySignalCallOwnerFilter(tx, d.loginUserID)
	tx = applySignalCallUserFilters(tx, userID, status, startTime, endTime)
	var list []*model_struct.LocalSignalCallRecord
	err := tx.Order("create_time DESC").Offset(offset).Limit(count).Find(&list).Error
	return list, errs.WrapMsg(err, "SearchSignalCallRecordsByUser failed")
}

func (d *DataBase) CountSignalCallRecordsByUser(ctx context.Context, userID string, status int32, startTime, endTime int64) (int64, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	tx := d.signalDB().WithContext(ctx).Model(&model_struct.LocalSignalCallRecord{})
	tx = applySignalCallOwnerFilter(tx, d.loginUserID)
	tx = applySignalCallUserFilters(tx, userID, status, startTime, endTime)
	var n int64
	err := tx.Count(&n).Error
	return n, errs.WrapMsg(err, "CountSignalCallRecordsByUser failed")
}

func applySignalCallUserFilters(tx *gorm.DB, userID string, status int32, startTime, endTime int64) *gorm.DB {
	tx = applyExcludeGroupCallFilter(tx)
	uid := strings.TrimSpace(userID)
	if uid != "" {
		pattern := "%\"" + uid + "\"%"
		tx = tx.Where("inviter_user_id = ? OR invitee_user_ids LIKE ?", uid, pattern)
	}
	if status != 0 {
		tx = tx.Where("status = ?", status)
	}
	if startTime > 0 {
		tx = tx.Where("create_time >= ?", startTime)
	}
	if endTime > 0 {
		tx = tx.Where("create_time <= ?", endTime)
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
	err := applySignalCallOwnerFilter(
		d.signalDB().WithContext(ctx).Where("s_id = ?", sID),
		d.loginUserID,
	).Take(&rec).Error
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
		applySignalCallOwnerFilter(
			d.signalDB().WithContext(ctx).Where("s_id IN ?", sIDs),
			d.loginUserID,
		).Delete(&model_struct.LocalSignalCallRecord{}).Error,
		"DeleteSignalCallRecords failed",
	)
}

func (d *DataBase) ClearAllSignalCallRecords(ctx context.Context) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	return errs.WrapMsg(
		applySignalCallOwnerFilter(
			d.signalDB().WithContext(ctx),
			d.loginUserID,
		).Delete(&model_struct.LocalSignalCallRecord{}).Error,
		"ClearAllSignalCallRecords failed",
	)
}

func (d *DataBase) UpdateSignalCallRecordUserProfile(ctx context.Context, userID, nickname, faceURL string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	tx := applySignalCallOwnerFilter(d.signalDB().WithContext(ctx), d.loginUserID)
	if nickname != "" {
		if err := tx.Model(&model_struct.LocalSignalCallRecord{}).Where("inviter_user_id = ?", userID).
			Update("inviter_user_nickname", nickname).Error; err != nil {
			return errs.WrapMsg(err, "UpdateSignalCallRecordUserProfile inviter nickname failed")
		}
		if err := tx.Model(&model_struct.LocalSignalCallRecord{}).Where("invitee_uid = ?", userID).
			Update("invitee_user_nickname", nickname).Error; err != nil {
			return errs.WrapMsg(err, "UpdateSignalCallRecordUserProfile invitee nickname failed")
		}
	}
	// Always sync faceURL (including empty string) so a stale avatar is cleared when
	// the user's account is deleted and their profile becomes empty.
	if err := tx.Model(&model_struct.LocalSignalCallRecord{}).Where("inviter_user_id = ?", userID).
		Update("inviter_user_face_url", faceURL).Error; err != nil {
		return errs.WrapMsg(err, "UpdateSignalCallRecordUserProfile inviter faceURL failed")
	}
	if err := tx.Model(&model_struct.LocalSignalCallRecord{}).Where("invitee_uid = ?", userID).
		Update("invitee_user_face_url", faceURL).Error; err != nil {
		return errs.WrapMsg(err, "UpdateSignalCallRecordUserProfile invitee faceURL failed")
	}
	return nil
}

func (d *DataBase) ListSignalCallRecordsByParticipant(ctx context.Context, userID string) ([]*model_struct.LocalSignalCallRecord, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil
	}
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	pattern := "%\"" + userID + "\"%"
	var list []*model_struct.LocalSignalCallRecord
	err := applySignalCallOwnerFilter(
		d.signalDB().WithContext(ctx).
			Where("inviter_user_id = ? OR invitee_uid = ? OR invitee_user_ids LIKE ?", userID, userID, pattern),
		d.loginUserID,
	).Find(&list).Error
	return list, errs.WrapMsg(err, "ListSignalCallRecordsByParticipant failed")
}

func (d *DataBase) UpdateSignalCallRecordCalleeMatchText(ctx context.Context, sID, calleeMatchText string) error {
	sID = strings.TrimSpace(sID)
	if sID == "" {
		return nil
	}
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	return errs.WrapMsg(
		applySignalCallOwnerFilter(
			d.signalDB().WithContext(ctx).Model(&model_struct.LocalSignalCallRecord{}).
				Where("s_id = ?", sID),
			d.loginUserID,
		).Update("callee_match_text", calleeMatchText).Error,
		"UpdateSignalCallRecordCalleeMatchText failed",
	)
}
