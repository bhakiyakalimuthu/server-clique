package server

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type MemStore struct {
	logger *zap.Logger
	mu     *sync.RWMutex
	items  []pair
	cache  map[string]int
}

type pair struct {
	key, value string
}

var _ Store = (*MemStore)(nil)

func NewMemStore(logger *zap.Logger) *MemStore {
	return &MemStore{
		logger: logger,
		mu:     new(sync.RWMutex),
		items:  make([]pair, 0),
		// index of items used as value for map instead of items to reduce the size of cache
		cache: make(map[string]int),
	}
}

func (m *MemStore) Add(ctx context.Context, key, value string) {
	defer m.mu.Unlock()
	m.mu.Lock()
	val := pair{
		key:   key,
		value: value,
	}
	if index, ok := m.cache[key]; ok {
		// key exist already, update the value
		m.items[index] = val
		return
	}
	m.items = append(m.items, val)  // append the value
	m.cache[key] = len(m.items) - 1 // update the index
}

func (m *MemStore) Remove(ctx context.Context, key string) bool {
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

func (m *MemStore) Get(ctx context.Context, key string) (string, bool) {
	defer m.mu.RUnlock()
	m.mu.RLock()
	index, ok := m.cache[key]
	if ok {
		return m.items[index].value, ok
	}
	return "", ok
}

func (m *MemStore) GetAll(ctx context.Context) []pair {
	defer m.mu.RUnlock()
	m.mu.RLock()
	return m.items
}
