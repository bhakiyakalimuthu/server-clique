package server

import (
	"context"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

type MemStoreOptimised struct {
	logger *zap.Logger
	mu     *sync.RWMutex
	items  []item
	cache  map[string]int
}

func NewMemStoreOptimised(logger *zap.Logger) *MemStoreOptimised {
	return &MemStoreOptimised{
		logger: logger,
		mu:     new(sync.RWMutex),
		items:  make([]item, 0),
		// index of items used as value for map instead of item to reduce the size of cache
		cache: make(map[string]int),
	}
}

func (m *MemStoreOptimised) Add(ctx context.Context, key, value string, timestamp time.Time) {
	defer m.mu.Unlock()
	m.mu.Lock()
	val := item{
		key:       key,
		value:     value,
		timestamp: timestamp.UnixNano(),
	}
	if index, ok := m.cache[key]; ok {
		// key exist already, update the value
		m.items[index] = val
		return
	}
	m.items = append(m.items, val)  // append the value
	m.cache[key] = len(m.items) - 1 // update the index
}

func (m *MemStoreOptimised) Remove(ctx context.Context, key string) bool {
	defer m.mu.Unlock()
	m.mu.Lock()
	index, ok := m.cache[key]
	if ok {
		m.items = append(m.items[:index], m.items[index+1:]...)
		delete(m.cache, key)

		for i := index; i < len(m.items); i++ {
			m.cache[m.items[i].key] = i
		}
	}
	return ok
}

func (m *MemStoreOptimised) Get(ctx context.Context, key string) (string, bool) {
	defer m.mu.RUnlock()
	m.mu.RLock()
	index, ok := m.cache[key]
	if ok {
		return m.items[index].value, ok
	}
	return "", ok
}

func (m *MemStoreOptimised) GetAll(ctx context.Context) []item {
	defer m.mu.RUnlock()
	m.mu.RLock()
	sorted := make([]item, len(m.items))
	copy(sorted, m.items)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].timestamp < sorted[j].timestamp
	})
	return sorted
}
