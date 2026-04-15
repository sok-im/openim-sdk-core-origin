//go:build !js
// +build !js

package signaling

import (
	"sync"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"
)

const detailCacheTTL = 5 * time.Minute
const detailCacheMaxSize = 1000

type cacheItem struct {
	value      *sdk_struct.SignalCallRecordWithDialStatus
	expiration time.Time
}

// detailCache 为 GetLocalSignalCallRecordDetail 提供简单 LRU + TTL 缓存（仅 native 使用）
type detailCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
	order []string // 简单 LRU 顺序（FIFO 近似）
}

func newDetailCache() *detailCache {
	return &detailCache{
		items: make(map[string]*cacheItem, detailCacheMaxSize),
		order: make([]string, 0, detailCacheMaxSize),
	}
}

func (c *detailCache) Get(sID string) (*sdk_struct.SignalCallRecordWithDialStatus, bool) {
	c.mu.RLock()
	item, ok := c.items[sID]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if time.Now().After(item.expiration) {
		c.mu.Lock()
		delete(c.items, sID)
		c.mu.Unlock()
		return nil, false
	}

	return item.value, true
}

func (c *detailCache) Set(sID string, value *sdk_struct.SignalCallRecordWithDialStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.items[sID] = &cacheItem{
		value:      value,
		expiration: now.Add(detailCacheTTL),
	}

	c.order = append(c.order, sID)
	if len(c.items) > detailCacheMaxSize {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.items, oldest)
	}
}

func (c *detailCache) Clear() {
	c.mu.Lock()
	c.items = make(map[string]*cacheItem)
	c.order = c.order[:0]
	c.mu.Unlock()
}
