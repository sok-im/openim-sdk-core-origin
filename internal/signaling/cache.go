//go:build !js
// +build !js

package signaling

import (
	"container/list"
	"sync"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"
)

const detailCacheTTL = 5 * time.Minute
const detailCacheMaxSize = 1000

type cacheItem struct {
	key        string
	value      *sdk_struct.SignalCallRecordWithDialStatus
	expiration time.Time
	element    *list.Element
}

// detailCache 提供 LRU + TTL 缓存（仅 native 使用）。
// 使用 container/list 实现精确 LRU 驱逐，避免 FIFO slice 的重复条目问题。
type detailCache struct {
	mu    sync.Mutex
	items map[string]*cacheItem
	order *list.List
}

func newDetailCache() *detailCache {
	return &detailCache{
		items: make(map[string]*cacheItem, detailCacheMaxSize),
		order: list.New(),
	}
}

func (c *detailCache) Get(sID string) (*sdk_struct.SignalCallRecordWithDialStatus, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[sID]
	if !ok {
		return nil, false
	}

	if time.Now().After(item.expiration) {
		c.removeLocked(item)
		return nil, false
	}

	c.order.MoveToFront(item.element)
	return item.value, true
}

func (c *detailCache) Set(sID string, value *sdk_struct.SignalCallRecordWithDialStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.items[sID]; ok {
		existing.value = value
		existing.expiration = time.Now().Add(detailCacheTTL)
		c.order.MoveToFront(existing.element)
		return
	}

	item := &cacheItem{
		key:        sID,
		value:      value,
		expiration: time.Now().Add(detailCacheTTL),
	}
	item.element = c.order.PushFront(item)
	c.items[sID] = item

	for len(c.items) > detailCacheMaxSize {
		c.removeOldestLocked()
	}
}

func (c *detailCache) Delete(sID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[sID]; ok {
		c.removeLocked(item)
	}
}

func (c *detailCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem, detailCacheMaxSize)
	c.order.Init()
}

func (c *detailCache) removeLocked(item *cacheItem) {
	c.order.Remove(item.element)
	delete(c.items, item.key)
}

func (c *detailCache) removeOldestLocked() {
	back := c.order.Back()
	if back == nil {
		return
	}
	item := back.Value.(*cacheItem)
	c.removeLocked(item)
}
