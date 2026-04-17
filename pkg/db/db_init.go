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
	"path/filepath"
	"sync"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/version"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/log"
)

type TableChecker struct {
	tableCache map[string]bool
	mu         sync.RWMutex
}

func NewTableChecker(tables []string) *TableChecker {
	tc := &TableChecker{
		tableCache: make(map[string]bool),
	}
	tc.InitTableCache(tables)
	return tc
}

func (tc *TableChecker) InitTableCache(tables []string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	for _, table := range tables {
		tc.tableCache[table] = true
	}
}

func (tc *TableChecker) HasTable(tableName string) bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.tableCache[tableName]
}
func (tc *TableChecker) UpdateTable(tableName string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tableCache[tableName] = true
}

type DataBase struct {
	loginUserID  string
	dbDir        string
	conn         *gorm.DB
	tableChecker *TableChecker
	mRWMutex     sync.RWMutex
}

func (d *DataBase) InitDB(ctx context.Context, userID string, dataDir string) error {
	panic("implement me")
}

func (d *DataBase) Close(ctx context.Context) error {
	dbConn, err := d.conn.WithContext(ctx).DB()
	if err != nil {
		return err
	} else {
		if dbConn != nil {
			err := dbConn.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func NewDataBase(ctx context.Context, loginUserID string, dbDir string, logLevel int) (*DataBase, error) {
	dataBase := &DataBase{loginUserID: loginUserID, dbDir: dbDir}
	err := dataBase.initDB(ctx, logLevel)
	if err != nil {
		return dataBase, errs.WrapMsg(err, "initDB failed "+dbDir)
	}
	tables, err := dataBase.GetExistTables(ctx)
	if err != nil {
		return dataBase, errs.Wrap(err)
	}
	dataBase.tableChecker = NewTableChecker(tables)

	return dataBase, nil
}

func (d *DataBase) initDB(ctx context.Context, logLevel int) error {
	var zLogLevel logger.LogLevel
	if d.loginUserID == "" {
		return errors.New("no uid")
	}
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()

	path := d.dbDir + "/OpenIM_" + constant.BigVersion + "_" + d.loginUserID + ".db"
	dbFileName, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	log.ZInfo(ctx, "sqlite", "path", dbFileName)
	// slowThreshold := 500
	// sqlLogger := log.NewSqlLogger(logger.LogLevel(sdk_struct.ServerConf.LogLevel), true, time.Duration(slowThreshold)*time.Millisecond)
	if logLevel > 5 {
		zLogLevel = logger.Info
	} else {
		zLogLevel = logger.Silent
	}
	var (
		db *gorm.DB
	)
	db, err = gorm.Open(sqlite.Open(dbFileName), &gorm.Config{Logger: log.NewSqlLogger(zLogLevel, false, time.Millisecond*200)})
	if err != nil {
		return errs.WrapMsg(err, "open db failed "+dbFileName)
	}

	log.ZDebug(ctx, "open db success", "dbFileName", dbFileName)
	sqlDB, err := db.DB()
	if err != nil {
		return errs.WrapMsg(err, "get sql db failed")
	}

	sqlDB.SetConnMaxLifetime(time.Hour * 1)
	sqlDB.SetMaxOpenConns(3)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxIdleTime(time.Minute * 10)
	d.conn = db

	// base
	if err = db.AutoMigrate(
		&model_struct.LocalAppSDKVersion{},
		&model_struct.LocalConversationSyncedMaxSeq{},
	); err != nil {
		return err
	}

	if err = d.versionDataMigrate(ctx); err != nil {
		return err
	}

	if err = d.conn.WithContext(ctx).AutoMigrate(&model_struct.LocalSignalCallRecord{}); err != nil {
		return err
	}

	// 旧数据无 dial_status 列时补为「已拨通」（历史记录均为挂断产生）
	if err = d.conn.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET dial_status = ? WHERE dial_status = 0 OR dial_status IS NULL`,
		constant.SignalCallDialStatusConnected,
	).Error; err != nil {
		return err
	}

	// 为 callee_match_text 增加索引（加速被叫用户名模糊查询）
	if err = d.conn.WithContext(ctx).Exec(
		`CREATE INDEX IF NOT EXISTS idx_callee_match_text ON local_signal_call_records(callee_match_text)`,
	).Error; err != nil {
		return err
	}

	// 旧数据无 direction 列时根据 inviter_user_id 推断：inviterUserID == loginUserID → outgoing(1)，否则 → incoming(2)
	if err = d.conn.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET direction = CASE WHEN inviter_user_id = ? THEN ? ELSE ? END WHERE direction = 0 OR direction IS NULL`,
		d.loginUserID, constant.SignalCallDirectionOutgoing, constant.SignalCallDirectionIncoming,
	).Error; err != nil {
		return err
	}

	// 旧数据无 connect_time 列时：已拨通(dial_status=2)的记录将 connect_time 设为 create_time（近似值）
	if err = d.conn.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET connect_time = create_time WHERE (connect_time = 0 OR connect_time IS NULL) AND dial_status = ?`,
		constant.SignalCallDialStatusConnected,
	).Error; err != nil {
		return err
	}

	// 旧数据无 call_duration 列时：已拨通的记录用 end_time - connect_time 近似通话时长
	if err = d.conn.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET call_duration = CASE WHEN end_time > connect_time THEN end_time - connect_time ELSE 0 END WHERE (call_duration = 0 OR call_duration IS NULL) AND dial_status = ? AND connect_time > 0`,
		constant.SignalCallDialStatusConnected,
	).Error; err != nil {
		return err
	}

	return nil
}

func (d *DataBase) versionDataMigrate(ctx context.Context) error {
	verModel, err := d.GetAppSDKVersion(ctx)
	if errs.Unwrap(err) == errs.ErrRecordNotFound {
		err = d.conn.AutoMigrate(
			&model_struct.LocalAppSDKVersion{},
			&model_struct.LocalFriend{},
			&model_struct.LocalGroup{},
			&model_struct.LocalGroupMember{},
			&model_struct.LocalUser{},
			&model_struct.LocalBlack{},
			&model_struct.LocalConversation{},
			&model_struct.NotificationSeqs{},
			&model_struct.LocalChatLog{},
			&model_struct.LocalConversationSyncedMaxSeq{},
			&model_struct.LocalChatLogReactionExtensions{},
			&model_struct.LocalUpload{},
			&model_struct.LocalStranger{},
			&model_struct.LocalSendingMessages{},
			&model_struct.LocalUserCommand{},
			&model_struct.LocalVersionSync{},
		)
		if err != nil {
			return err
		}
		err = d.SetAppSDKVersion(ctx, &model_struct.LocalAppSDKVersion{Version: version.Version})
		if err != nil {
			return err
		}

		return nil
	} else if err != nil {
		return err
	}
	if verModel.Version != version.Version {
		switch version.Version {
		case "3.8.0":
			d.conn.AutoMigrate(&model_struct.LocalAppSDKVersion{})
		}
		err = d.SetAppSDKVersion(ctx, &model_struct.LocalAppSDKVersion{Version: version.Version})
		if err != nil {
			return err
		}
	}

	return nil
}
