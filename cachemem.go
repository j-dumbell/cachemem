package cachemem

import "time"

type entry[V any] struct {
	value     V
	expiresAt time.Time
}

// Cache is a strongly typed, concurrency-safe, in-memory cache.
type Cache[K comparable, V any] struct {
	store asyncMap[K, entry[V]]
}

// New initializes a new Cache.
func New[K comparable, V any]() Cache[K, V] {
	return Cache[K, V]{
		store: newAsyncMap[K, entry[V]](),
	}
}

// Set writes a new entry to the cache with key Key and value Value and no expiry time.
// If an entry with the same key already exists, it will be overwritten.
func (cache *Cache[K, V]) Set(key K, value V) {
	e := entry[V]{
		value:     value,
		expiresAt: time.Time{},
	}

	cache.store.Set(key, e)
}

// SetWithExpiry writes a new entry to the cache with key Key and value Value and expiry duration expiresIn.
// If an entry with the same key already exists, it will be overwritten.
// After expiresIn has elapsed, the entry will be deleted from the cache.
func (cache *Cache[K, V]) SetWithExpiry(key K, value V, expiresIn time.Duration) {
	e := entry[V]{
		value:     value,
		expiresAt: time.Now().Add(expiresIn),
	}
	cache.store.Set(key, e)

	if expiresIn > 0 {
		time.AfterFunc(expiresIn, func() {
			cache.store.Delete(key)
		})
	}
}

// Get retrieves a record with key Key from the cache.  If the record exists and has not expired, it's value
// and true are returned, otherwise false.
func (cache *Cache[K, V]) Get(key K) (V, bool) {
	record, exists := cache.store.Get(key)
	if !exists || (exists && !record.expiresAt.IsZero() && !record.expiresAt.After(time.Now())) {
		return record.value, false
	}

	return record.value, true
}

// Delete deletes an entry with key Key from the cache.
func (cache *Cache[K, V]) Delete(key K) {
	cache.store.Delete(key)
}

// Clear deletes all entries in the cache.
func (cache *Cache[K, V]) Clear() {
	cache.store.Clear()
}

// Len returns the number of entries in the cache.
func (cache *Cache[K, V]) Len() int {
	return cache.store.Len()
}
