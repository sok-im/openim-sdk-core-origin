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
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/tools/errs"
)

const friendFullNameMatchSQL = "first_name LIKE ? OR last_name LIKE ? OR (first_name || ' ' || last_name) LIKE ?"

func applyFriendProfileSearch(tx *gorm.DB, keyword string) *gorm.DB {
	pattern := "%" + keyword + "%"
	profileCond := "remark LIKE ? OR name LIKE ? OR friend_first_name LIKE ? OR friend_last_name LIKE ? OR (friend_first_name || ' ' || friend_last_name) LIKE ? OR first_name LIKE ? OR last_name LIKE ? OR (first_name || ' ' || last_name) LIKE ?"
	args := []interface{}{pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern}

	tokens := strings.Fields(keyword)
	if len(tokens) > 1 {
		var multiParts []string
		for _, t := range tokens {
			p := "%" + t + "%"
			multiParts = append(multiParts, "("+friendFullNameMatchSQL+")")
			args = append(args, p, p, p)
		}
		profileCond = "(" + profileCond + " OR (" + strings.Join(multiParts, " AND ") + "))"
	}
	return tx.Where(profileCond, args...)
}

func (d *DataBase) InsertFriend(ctx context.Context, friend *model_struct.LocalFriend) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	return errs.WrapMsg(d.conn.WithContext(ctx).Create(friend).Error, "InsertFriend failed")
}

func (d *DataBase) DeleteFriendDB(ctx context.Context, friendUserID string) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	return errs.WrapMsg(d.conn.WithContext(ctx).Where("owner_user_id=? and friend_user_id=?", d.loginUserID, friendUserID).Delete(&model_struct.LocalFriend{}).Error, "DeleteFriend failed")
}

func (d *DataBase) GetFriendListCount(ctx context.Context) (int64, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	var count int64
	return count, errs.WrapMsg(d.conn.WithContext(ctx).Model(&model_struct.LocalFriend{}).Count(&count).Error, "GetFriendListCount failed")
}

func (d *DataBase) UpdateFriend(ctx context.Context, friend *model_struct.LocalFriend) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()

	t := d.conn.WithContext(ctx).Model(friend).Select("*").Updates(*friend)
	if t.RowsAffected == 0 {
		return errs.WrapMsg(errors.New("RowsAffected == 0"), "no update")
	}
	return errs.Wrap(t.Error)

}
func (d *DataBase) GetAllFriendList(ctx context.Context) ([]*model_struct.LocalFriend, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	var friendList []*model_struct.LocalFriend
	return friendList, errs.WrapMsg(d.conn.WithContext(ctx).Where("owner_user_id = ?", d.loginUserID).Find(&friendList).Error,
		"GetFriendList failed")
}

func (d *DataBase) GetPageFriendList(ctx context.Context, offset, count int) ([]*model_struct.LocalFriend, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	var friendList []*model_struct.LocalFriend
	err := errs.WrapMsg(d.conn.WithContext(ctx).Where("owner_user_id = ?", d.loginUserID).Offset(offset).Limit(count).Order("name").Find(&friendList).Error,
		"GetFriendList failed")
	return friendList, err
}

func (d *DataBase) BatchInsertFriend(ctx context.Context, friendList []*model_struct.LocalFriend) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	if friendList == nil {
		return errs.New("nil").Wrap()
	}
	return errs.WrapMsg(d.conn.WithContext(ctx).Create(friendList).Error, "BatchInsertFriendList failed")
}

func (d *DataBase) DeleteAllFriend(ctx context.Context) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	return errs.WrapMsg(d.conn.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model_struct.LocalFriend{}).Error, "DeleteAllFriend failed")
}

func (d *DataBase) SearchFriendList(ctx context.Context, keyword string, isSearchUserID, isSearchNickname, isSearchRemark, isSearchFullName bool) ([]*model_struct.LocalFriend, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, nil
	}
	var count int
	var friendList []*model_struct.LocalFriend
	var condition string
	if isSearchUserID {
		condition = fmt.Sprintf("friend_user_id like %q ", "%"+keyword+"%")
		count++
	}
	if isSearchNickname {
		if count > 0 {
			condition += "or "
		}
		condition += fmt.Sprintf("name like %q ", "%"+keyword+"%")
		count++
	}
	if isSearchRemark {
		if count > 0 {
			condition += "or "
		}
		condition += fmt.Sprintf("remark like %q ", "%"+keyword+"%")
		count++
	}
	if isSearchFullName {
		if count > 0 {
			condition += "or "
		}
		condition += fmt.Sprintf("(first_name like %q or last_name like %q or (first_name || ' ' || last_name) like %q) ", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if count == 0 {
		return nil, nil
	}
	err := d.conn.WithContext(ctx).Where("owner_user_id = ?", d.loginUserID).Where(condition).Order("create_time DESC").Find(&friendList).Error
	return friendList, errs.WrapMsg(err, "SearchFriendList failed")
}

func (d *DataBase) SearchFriendListByProfile(ctx context.Context, keyword string) ([]*model_struct.LocalFriend, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, nil
	}
	var friendList []*model_struct.LocalFriend
	tx := d.conn.WithContext(ctx).Where("owner_user_id = ?", d.loginUserID)
	tx = applyFriendProfileSearch(tx, keyword)
	err := tx.Order("create_time DESC").Find(&friendList).Error
	return friendList, errs.WrapMsg(err, "SearchFriendListByProfile failed")
}

func (d *DataBase) GetFriendInfoByFriendUserID(ctx context.Context, FriendUserID string) (*model_struct.LocalFriend, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	var friend model_struct.LocalFriend
	return &friend, errs.WrapMsg(d.conn.WithContext(ctx).Where("owner_user_id = ? AND friend_user_id = ?",
		d.loginUserID, FriendUserID).Take(&friend).Error, "GetFriendInfoByFriendUserID failed")
}

func (d *DataBase) GetFriendInfoList(ctx context.Context, friendUserIDList []string) ([]*model_struct.LocalFriend, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	var friendList []*model_struct.LocalFriend
	err := errs.WrapMsg(d.conn.WithContext(ctx).Where("owner_user_id = ? AND friend_user_id IN ?", d.loginUserID, friendUserIDList).Find(&friendList).Error, "GetFriendInfoListByFriendUserID failed")
	return friendList, err
}
func (d *DataBase) UpdateColumnsFriend(ctx context.Context, friendIDs []string, args map[string]interface{}) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	return errs.WrapMsg(d.conn.WithContext(ctx).Model(&model_struct.LocalFriend{}).Where("friend_user_id IN ?", friendIDs).Updates(args).Error, "UpdateColumnsFriend failed")
}
