//go:build js && wasm
// +build js,wasm

package signaling

import "github.com/openimsdk/openim-sdk-core/v3/sdk_struct"

// WASM 下不使用缓存（IndexedDB 本身有缓存机制）
type detailCache struct{}

func newDetailCache() *detailCache {
	return &detailCache{}
}

func (c *detailCache) Get(sID string) (*sdk_struct.SignalCallRecordWithDialStatus, bool) {
	return nil, false
}

func (c *detailCache) Set(sID string, value *sdk_struct.SignalCallRecordWithDialStatus) {}

func (c *detailCache) Clear() {}
