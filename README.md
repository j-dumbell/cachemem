# cache-mem
![example workflow](https://github.com/j-dumbell/cache-mem/actions/workflows/test-build.yml/badge.svg?branch=main)

A concurrency-safe, strongly typed, in-memory cache in Golang.  Cache records
can be configured to expire and automatically deleted.

## Installation
```shell
go get github.com/j-dumbell/cache-mem
```

## Example
```go
import (
    "fmt"
    "time"
    
    "github.com/j-dumbell/cachemem"
)

// implements Fetcher
type DummyFetcher struct {
}

func (f *DummyFetcher) FetchOne(i int) (string, error) {
    return "", nil
}

func (f *DummyFetcher) FetchMany(arrI []int) ([]string, error) {
    return []string{}, nil
}

func getKey(v string) int {
    return 0
}

func main() {
    fetcher := DummyFetcher{}
    
    // initialize a new cache with int keys and string values
    cache := cachemem.New[int, string](&fetcher, getKey, time.Minute)
    
    // Set a new record with an expiry of 1 hour
    cache.Set("123", time.Hour)
    
    // Get a record from the cache
    record, ok := cache.Get(1)
    
    // Get a record from the cache if it exists, otherwise fetch it.
    record, err := cache.GetOrFetch(2, time.Minute)
    
    // The number of records in the cache
    cacheLength := cache.Len()
    
    // Delete a cache record by key
    cache.Delete(1)
    
    // Delete all entries from the cache
    cache.Clear()
    
    // Start deleting expired records
    go cache.StartCleaning()
    
    // Stop deleting expired records
    cache.StopCleaning()
}
```