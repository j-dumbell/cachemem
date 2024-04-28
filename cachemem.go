package cachemem

import (
	"sync"
	"time"
)

// Fetcher fetches records by their key.
type Fetcher[K comparable, V any] interface {
	FetchOne(K) (V, error)
	FetchMany(arrK []K) ([]V, error)
}

type entry[V any] struct {
	value     V
	expiresAt time.Time
}

func (e *entry[V]) hasExpired() bool {
	return time.Now().After(e.expiresAt)
}

// Cache is a strongly typed, concurrency-safe, in-memory cache.
type Cache[K comparable, V any] struct {
	fetcher         Fetcher[K, V]
	getKey          func(V) K
	mutex           sync.Mutex
	store           map[K]entry[V]
	cleanFreq       time.Duration
	signalStopClean chan struct{}
	isCleaning      bool
}

// New initializes a new, empty Cache.
func New[K comparable, V any](fetcher Fetcher[K, V], getKey func(V) K, cleanFreq time.Duration) Cache[K, V] {
	return Cache[K, V]{
		fetcher:         fetcher,
		getKey:          getKey,
		mutex:           sync.Mutex{},
		store:           map[K]entry[V]{},
		cleanFreq:       cleanFreq,
		signalStopClean: make(chan struct{}),
		isCleaning:      false,
	}
}

// StartCleaning begins removing expired records from the cache at the configured frequency.
// It blocks until StopCleaning is called.
func (cache *Cache[K, V]) StartCleaning() {
	if cache.isCleaning {
		return
	}

	cache.isCleaning = true
	ticker := time.NewTicker(cache.cleanFreq)
	for {
		select {
		case <-ticker.C:
			cache.clean()

		case <-cache.signalStopClean:
			ticker.Stop()
			cache.isCleaning = false
			return
		}
	}
}

// StopCleaning stops removing expired records from the cache.
func (cache *Cache[K, V]) StopCleaning() {
	if !cache.isCleaning {
		return
	}
	cache.signalStopClean <- struct{}{}
}

func (cache *Cache[K, V]) clean() {
	for k, v := range cache.store {
		if v.hasExpired() {
			cache.Delete(k)
		}
	}
}

func (cache *Cache[K, V]) set(e entry[V]) {
	cache.mutex.Lock()
	cache.store[cache.getKey(e.value)] = e
	cache.mutex.Unlock()
}

// Get retrieves a record with key Key from the cache if it exists and
// has not expired.
func (cache *Cache[K, V]) Get(key K) (V, bool) {
	e, exists := cache.store[key]
	if !exists || e.hasExpired() {
		return e.value, false
	}

	return e.value, true
}

// GetMany retrieves the subset of the provided records from the cache that exist and have not expired.
func (cache *Cache[K, V]) GetMany(keys []K) []V {
	var cachedRecords []V

	for _, key := range keys {
		value, ok := cache.Get(key)
		if ok {
			cachedRecords = append(cachedRecords, value)
		}
	}

	return cachedRecords
}

// Set writes a new entry to the cache with expiry duration expiresIn.
// If an entry with the same key already exists, it will be overwritten.
// After expiresIn has elapsed, the entry will be deleted from the cache.
func (cache *Cache[K, V]) Set(value V, expiresIn time.Duration) {
	e := entry[V]{
		value:     value,
		expiresAt: time.Now().Add(expiresIn),
	}
	cache.set(e)
}

// GetOrFetch retrieves a record by key from the cache if it exists and
// has not expired, otherwise it fetches and caches it with the provided expiry.
func (cache *Cache[K, V]) GetOrFetch(key K, expiresIn time.Duration) (V, error) {
	cachedValue, ok := cache.Get(key)
	if ok {
		return cachedValue, nil
	}

	fetchedValue, err := cache.fetcher.FetchOne(key)
	if err != nil {
		var v V
		return v, err
	}

	cache.Set(fetchedValue, expiresIn)
	return fetchedValue, nil
}

// Delete deletes an record by key from the cache.
func (cache *Cache[K, V]) Delete(key K) {
	cache.mutex.Lock()
	delete(cache.store, key)
	cache.mutex.Unlock()
}

// Clear deletes all entries in the cache.
func (cache *Cache[K, V]) Clear() {
	cache.mutex.Lock()
	cache.store = map[K]entry[V]{}
	cache.mutex.Unlock()
}

// Len returns the number of records in the cache, including
// expired records.
func (cache *Cache[K, V]) Len() int {
	return len(cache.store)
}

// FetchMany fetches and caches the subset of the provided records that have
// not been cached and have not expired.
func (cache *Cache[K, V]) FetchMany(arrK []K, expiresIn time.Duration) error {
	expiresAt := time.Now().Add(expiresIn)

	var keysToFetch []K
	for _, key := range arrK {
		_, ok := cache.Get(key)
		if !ok {
			keysToFetch = append(keysToFetch, key)
		}
	}

	values, err := cache.fetcher.FetchMany(keysToFetch)
	if err != nil {
		return err
	}

	for _, value := range values {
		e := entry[V]{
			value:     value,
			expiresAt: expiresAt,
		}
		cache.set(e)
	}

	return nil
}
