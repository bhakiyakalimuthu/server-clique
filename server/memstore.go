package server

import (
	"context"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

type MemStore struct {
	logger *zap.Logger
	mu     *sync.RWMutex
	cache  map[string]item
}

type item struct {
	key, value string
	timestamp  int64
}

var _ Store = (*MemStore)(nil)

func NewMemStore(logger *zap.Logger) *MemStore {
	return &MemStore{
		logger: logger,
		mu:     new(sync.RWMutex),
		cache:  make(map[string]item),
	}
}

func (m *MemStore) Add(ctx context.Context, key, value string, timestamp time.Time) {
	m.mu.Lock()
	m.cache[key] = item{
		key:       key,
		value:     value,
		timestamp: timestamp.UnixNano(),
	}
	m.mu.Unlock()
}

func (m *MemStore) Remove(ctx context.Context, key string) bool {
	defer m.mu.Unlock()
	m.mu.Lock()
	_, ok := m.cache[key]
	if ok {
		delete(m.cache, key)
	}
	return ok
}

func (m *MemStore) Get(ctx context.Context, key string) (string, bool) {
	defer m.mu.RUnlock()
	m.mu.RLock()
	_item, ok := m.cache[key]
	if ok {
		return _item.value, ok
	}
	return "", ok
}

func (m *MemStore) GetAll(ctx context.Context) []item {
	defer m.mu.RUnlock()
	m.mu.RLock()
	sorted := make([]item, 0, len(m.cache))
	for _, _item := range m.cache {
		sorted = append(sorted, _item)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].timestamp < sorted[j].timestamp
	})
	return sorted
}
