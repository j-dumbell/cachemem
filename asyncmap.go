package cachemem

import (
	"sync"
)

type asyncMap[K comparable, V any] struct {
	store map[K]V
	mutex sync.RWMutex
}

func newAsyncMap[K comparable, V any]() asyncMap[K, V] {
	return asyncMap[K, V]{
		store: map[K]V{},
	}
}

func (m *asyncMap[K, V]) Set(key K, value V) {
	m.mutex.Lock()
	m.store[key] = value
	m.mutex.Unlock()
}

func (m *asyncMap[K, V]) Get(key K) (V, bool) {
	m.mutex.RLock()
	record, exists := m.store[key]
	m.mutex.RUnlock()
	return record, exists
}

func (m *asyncMap[K, V]) Delete(key K) {
	m.mutex.Lock()
	delete(m.store, key)
	m.mutex.Unlock()
}

func (m *asyncMap[K, V]) Clear() {
	m.mutex.Lock()
	m.store = map[K]V{}
	m.mutex.Unlock()
}

func (m *asyncMap[K, V]) Len() int {
	m.mutex.RLock()
	cacheLength := len(m.store)
	m.mutex.RUnlock()
	return cacheLength
}
