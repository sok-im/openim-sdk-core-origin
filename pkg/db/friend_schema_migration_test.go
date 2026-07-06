package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRunFriendSchemaMigrationsAddsNoteColumn(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "legacy_friends.db")

	legacyDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := legacyDB.Exec(`CREATE TABLE local_friends (
		owner_user_id varchar(64),
		friend_user_id varchar(64),
		remark varchar(255),
		create_time integer,
		add_source integer,
		operator_user_id varchar(64),
		name varchar(255),
		first_name varchar(255),
		last_name varchar(255),
		face_url varchar(255),
		ex varchar(1024),
		attached_info varchar(1024),
		is_pinned numeric,
		PRIMARY KEY (owner_user_id, friend_user_id)
	)`).Error; err != nil {
		t.Fatal(err)
	}
	sqlDB, err := legacyDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.Close()

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	if err := runFriendSchemaMigrations(ctx, db); err != nil {
		t.Fatal(err)
	}

	var noteCount int64
	if err := db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM pragma_table_info('local_friends') WHERE name = 'note'`,
	).Scan(&noteCount).Error; err != nil {
		t.Fatal(err)
	}
	if noteCount == 0 {
		t.Fatal("expected note column to be added")
	}

	friend := &model_struct.LocalFriend{
		OwnerUserID:  "user1",
		FriendUserID: "user2",
		Remark:       "hello",
		Note:         "private note",
	}
	if err := db.WithContext(ctx).Create(friend).Error; err != nil {
		t.Fatal(err)
	}
}

func TestNewDataBaseMigratesLegacyFriendTable(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	loginUserID := "1695766238"

	legacyPath := filepath.Join(dir, "OpenIM_v3_"+loginUserID+".db")
	legacyDB, err := gorm.Open(sqlite.Open(legacyPath), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := legacyDB.Exec(`CREATE TABLE local_friends (
		owner_user_id varchar(64),
		friend_user_id varchar(64),
		remark varchar(255),
		create_time integer,
		add_source integer,
		operator_user_id varchar(64),
		name varchar(255),
		first_name varchar(255),
		last_name varchar(255),
		face_url varchar(255),
		ex varchar(1024),
		attached_info varchar(1024),
		is_pinned numeric,
		PRIMARY KEY (owner_user_id, friend_user_id)
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := legacyDB.Exec(`CREATE TABLE local_app_sdk_version (
		version varchar(64) PRIMARY KEY
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := legacyDB.Exec(`INSERT INTO local_app_sdk_version(version) VALUES ('3.8.0')`).Error; err != nil {
		t.Fatal(err)
	}
	sqlDB, err := legacyDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.Close()

	db, err := NewDataBase(ctx, loginUserID, dir, 6)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close(ctx)
		_ = os.Remove(legacyPath)
	}()

	if err := db.BatchInsertFriend(ctx, []*model_struct.LocalFriend{{
		OwnerUserID:  loginUserID,
		FriendUserID: "friend1",
		Remark:       "hi",
		Note:         "note",
	}}); err != nil {
		t.Fatal(err)
	}
}
