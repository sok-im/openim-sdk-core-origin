package db

import (
	"context"
	"testing"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
)

func TestGetLatestValidateServerMessage(t *testing.T) {
	ctx := context.Background()
	db, err := NewDataBase(ctx, "1695766238", t.TempDir(), 6)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close(ctx) }()

	conversationID := "sg_93606743"
	startTime := int64(1710774468519)
	err = db.BatchInsertMessageList(ctx, conversationID, []*model_struct.LocalChatLog{
		{
			ClientMsgID: "msg1",
			ServerMsgID: "smsg1",
			Seq:         1,
			SendTime:    startTime - 1000,
		},
		{
			ClientMsgID: "msg2",
			ServerMsgID: "smsg2",
			Seq:         2,
			SendTime:    startTime - 100,
		},
		{
			ClientMsgID: "msg3",
			ServerMsgID: "smsg3",
			Seq:         0,
			SendTime:    startTime - 50,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	message, err := db.GetLatestValidServerMessage(ctx, conversationID, startTime, true)
	if err != nil {
		t.Fatal(err)
	}
	if message == nil {
		t.Fatal("expected a message, got nil")
	}
	if message.ClientMsgID != "msg2" {
		t.Fatalf("expected msg2, got %s", message.ClientMsgID)
	}
}
