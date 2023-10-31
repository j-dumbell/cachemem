# cache-mem
A concurrency-safe, strongly typed, in-memory cache in Golang.  Cache entries
may optionally be given an expiry time, and are automatically deleted from the
cache after expiry.

```go
// initialize a new cache with int keys and string values
cache := New[int, string]()

// Set a new entry with key 1 and value 'hello'
cache.Set(1, "hello")

// Set a new entry with an expiry time of 10 seconds
cache.SetWithExpiry(2, "world", time.Seconds * 10)

// Get a cache entry with a key of 1
value, ok := cache.Get(1)

// The number of entries in the cache
cacheLength := cache.Len()

// Deletes the cache entry with key 1
cache.Delete(1)

// Delete all entries from the cache
cache.Clear()
```