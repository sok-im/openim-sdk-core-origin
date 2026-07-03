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
	signalConn   *gorm.DB
	tableChecker *TableChecker
	mRWMutex     sync.RWMutex
}

func (d *DataBase) InitDB(ctx context.Context, userID string, dataDir string) error {
	panic("implement me")
}

func (d *DataBase) Close(ctx context.Context) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	var firstErr error
	closeGorm := func(db *gorm.DB) {
		if db == nil {
			return
		}
		sqlDB, err := db.WithContext(ctx).DB()
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return
		}
		if sqlDB != nil {
			if err := sqlDB.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	closeGorm(d.signalConn)
	closeGorm(d.conn)
	d.signalConn = nil
	d.conn = nil
	return firstErr
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
		// Run on every startup so additive column changes (e.g. the friend `note`
		// column) are applied to pre-existing tables. AutoMigrate only adds
		// missing columns and never drops data, so this is safe for upgrades.
		&model_struct.LocalFriend{},
	); err != nil {
		return err
	}

	if err = d.versionDataMigrate(ctx); err != nil {
		return err
	}

	if err = d.initSignalDB(ctx, zLogLevel); err != nil {
		return err
	}

	return nil
}

func (d *DataBase) signalDB() *gorm.DB {
	if d.signalConn != nil {
		return d.signalConn
	}
	return d.conn
}

func (d *DataBase) initSignalDB(ctx context.Context, zLogLevel logger.LogLevel) error {
	signalPath := d.dbDir + "/OpenIM_signal_" + constant.BigVersion + "_" + d.loginUserID + ".db"
	signalFileName, err := filepath.Abs(signalPath)
	if err != nil {
		return err
	}
	log.ZInfo(ctx, "sqlite signal call records", "path", signalFileName)

	signalDB, err := gorm.Open(sqlite.Open(signalFileName), &gorm.Config{Logger: log.NewSqlLogger(zLogLevel, false, time.Millisecond*200)})
	if err != nil {
		return errs.WrapMsg(err, "open signal db failed "+signalFileName)
	}
	signalSQL, err := signalDB.DB()
	if err != nil {
		return errs.WrapMsg(err, "get signal sql db failed")
	}
	signalSQL.SetConnMaxLifetime(time.Hour * 1)
	signalSQL.SetMaxOpenConns(3)
	signalSQL.SetMaxIdleConns(2)
	signalSQL.SetConnMaxIdleTime(time.Minute * 10)
	d.signalConn = signalDB

	if err = runSignalCallSchemaMigrations(ctx, d.signalConn, d.loginUserID); err != nil {
		return err
	}
	if err = migrateSignalCallRecordsFromMainDB(ctx, d.conn, d.signalConn, d.loginUserID); err != nil {
		return err
	}
	return backfillSignalCallRecordOwnerAndPurgeForeign(ctx, d.signalConn, d.loginUserID)
}

func runSignalCallSchemaMigrations(ctx context.Context, db *gorm.DB, loginUserID string) error {
	if err := db.WithContext(ctx).AutoMigrate(&model_struct.LocalSignalCallRecord{}); err != nil {
		return err
	}

	var dialStatusColCount int64
	if err := db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM pragma_table_info('local_signal_call_records') WHERE name = 'dial_status'`,
	).Scan(&dialStatusColCount).Error; err != nil {
		return err
	}
	if dialStatusColCount > 0 {
		if err := db.WithContext(ctx).Exec(
			`UPDATE local_signal_call_records SET status = CASE
				WHEN dial_status = 1 THEN ?
				WHEN dial_status = 2 THEN ?
				ELSE ?
			END WHERE status = 0 OR status IS NULL`,
			constant.SignalCallStatusNotConnected,
			constant.SignalCallStatusAnswered,
			constant.SignalCallStatusAnswered,
		).Error; err != nil {
			return err
		}
	}

	if err := db.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET status = ? WHERE status = 0 OR status IS NULL`,
		constant.SignalCallStatusAnswered,
	).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(
		`CREATE INDEX IF NOT EXISTS idx_callee_match_text ON local_signal_call_records(callee_match_text)`,
	).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET direction = CASE WHEN inviter_user_id = ? THEN ? ELSE ? END WHERE direction = 0 OR direction IS NULL`,
		loginUserID, constant.SignalCallDirectionOutgoing, constant.SignalCallDirectionIncoming,
	).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET role = CASE
			WHEN direction = ? THEN ?
			WHEN direction IN (?, ?) THEN ?
			ELSE ?
		END WHERE role = 0 OR role IS NULL`,
		constant.SignalCallDirectionOutgoing, constant.SignalCallRoleOutgoing,
		constant.SignalCallDirectionIncoming, constant.SignalCallDirectionMissed, constant.SignalCallRoleIncoming,
		constant.SignalCallRoleUnknown,
	).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET connect_time = create_time WHERE (connect_time = 0 OR connect_time IS NULL) AND status = ?`,
		constant.SignalCallStatusAnswered,
	).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET call_duration = CASE WHEN end_time > connect_time THEN end_time - connect_time ELSE 0 END WHERE (call_duration = 0 OR call_duration IS NULL) AND status = ? AND connect_time > 0`,
		constant.SignalCallStatusAnswered,
	).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET dial_duration = CASE
			WHEN connect_time > 0 AND connect_time > create_time THEN connect_time - create_time
			WHEN end_time > create_time THEN end_time - create_time
			ELSE 0
		END WHERE dial_duration = 0 OR dial_duration IS NULL`,
	).Error; err != nil {
		return err
	}

	return nil
}

// migrateSignalCallRecordsFromMainDB 将旧版主库中的通话记录一次性迁入独立库（按 loginUserID 隔离）。
func migrateSignalCallRecordsFromMainDB(ctx context.Context, mainDB, signalDB *gorm.DB, loginUserID string) error {
	if mainDB == nil || signalDB == nil {
		return nil
	}
	var mainTableCount int64
	if err := mainDB.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='local_signal_call_records'`,
	).Scan(&mainTableCount).Error; err != nil || mainTableCount == 0 {
		return err
	}

	var signalCount int64
	if err := signalDB.WithContext(ctx).Model(&model_struct.LocalSignalCallRecord{}).Count(&signalCount).Error; err != nil {
		return err
	}
	if signalCount > 0 {
		return nil
	}

	var records []*model_struct.LocalSignalCallRecord
	if err := mainDB.WithContext(ctx).Find(&records).Error; err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	log.ZInfo(ctx, "migrate signal call records to per-user signal db", "count", len(records))
	for _, r := range records {
		if r == nil || r.SID == "" {
			continue
		}
		r.OwnerUserID = loginUserID
		if err := signalDB.WithContext(ctx).Save(r).Error; err != nil {
			return errs.WrapMsg(err, "migrateSignalCallRecordsFromMainDB Save failed")
		}
	}
	return nil
}

// backfillSignalCallRecordOwnerAndPurgeForeign 为旧数据补全 owner_user_id，并清除非当前账号的记录。
func backfillSignalCallRecordOwnerAndPurgeForeign(ctx context.Context, db *gorm.DB, loginUserID string) error {
	if db == nil || loginUserID == "" {
		return nil
	}
	inviteePattern := "%\"" + loginUserID + "\"%"
	if err := db.WithContext(ctx).Exec(
		`UPDATE local_signal_call_records SET owner_user_id = ?
		 WHERE (owner_user_id = '' OR owner_user_id IS NULL)
		   AND (inviter_user_id = ? OR invitee_uid = ? OR invitee_user_ids LIKE ?)`,
		loginUserID, loginUserID, loginUserID, inviteePattern,
	).Error; err != nil {
		return err
	}
	if err := db.WithContext(ctx).Exec(
		`DELETE FROM local_signal_call_records
		 WHERE owner_user_id != ? OR owner_user_id IS NULL OR owner_user_id = ''`,
		loginUserID,
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
