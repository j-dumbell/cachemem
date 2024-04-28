package cachemem

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestFetcher struct {
	FetchManyCalls [][]int
}

func (fetcher *TestFetcher) FetchOne(i int) (string, error) {
	return strconv.Itoa(i), nil
}

func (fetcher *TestFetcher) FetchMany(arrI []int) ([]string, error) {
	fetcher.FetchManyCalls = append(fetcher.FetchManyCalls, arrI)

	var fetched []string
	for _, i := range arrI {
		fetched = append(fetched, strconv.Itoa(i))
	}
	return fetched, nil
}

var testFetcher = TestFetcher{}

func getKey(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func TestCache_Set(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Second)
	value := "10"
	cache.Set(value, time.Hour)

	actual, _ := cache.Get(10)
	assert.Equal(t, value, actual)
}

func TestCache_Get(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Second)
	value := "50"
	cache.Set(value, time.Hour)

	actual, ok := cache.Get(50)
	assert.Equal(t, value, actual)
	assert.True(t, ok)
}

func TestCache_Get_expired(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Second)
	value := "5"
	cache.Set(value, time.Nanosecond)

	time.Sleep(10 * time.Nanosecond)
	_, ok := cache.Get(5)

	assert.False(t, ok)
}

func TestCache_Get_keyNotExists(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Second)

	_, ok := cache.Get(2)
	assert.False(t, ok)
}

func TestCache_Delete(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Second)
	cache.Set("3", time.Hour)
	cache.Delete(3)

	_, ok := cache.Get(3)
	assert.False(t, ok)
}

func TestCache_Clear(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Second)
	cache.Set("1", time.Hour)
	cache.Set("2", time.Hour)

	cache.Clear()
	_, ok1 := cache.Get(1)
	_, ok2 := cache.Get(2)

	assert.False(t, ok1)
	assert.False(t, ok2)
}

func TestCache_Length(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Second)
	cache.Set("1", 1)
	cache.Set("2", 2)

	actual := cache.Len()
	assert.Equal(t, 2, actual)
}

func TestCache_GetOrFetch(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Second)
	actual, err := cache.GetOrFetch(2, time.Hour)
	assert.Equal(t, "2", actual)
	assert.NoError(t, err)

	cachedValue, ok := cache.Get(2)
	assert.Equal(t, "2", cachedValue)
	assert.True(t, ok)
}

func TestCache_FetchMany(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Second)
	cache.Set("1", time.Hour)
	cache.Set("3", time.Hour)
	err := cache.FetchMany([]int{1, 2, 3, 4}, time.Hour)

	value2, _ := cache.Get(2)
	value4, _ := cache.Get(4)

	assert.Equal(t, value2, "2")
	assert.Equal(t, value4, "4")

	assert.NoError(t, err)
	require.Len(t, testFetcher.FetchManyCalls, 1)
	assert.Len(t, testFetcher.FetchManyCalls[0], 2)
	assert.Subset(t, testFetcher.FetchManyCalls[0], []int{2, 4})
}

func TestCache_GetMany(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Second)
	cache.Set("1", time.Hour)
	cache.Set("2", time.Nanosecond)
	cache.Set("3", time.Hour)

	time.Sleep(10 * time.Nanosecond)
	actual := cache.GetMany([]int{1, 2, 3})

	assert.Len(t, actual, 2)
	assert.Subset(t, actual, []string{"1", "3"})
}

func TestCache_StartCleaning(t *testing.T) {
	cache := New[int, string](&testFetcher, getKey, time.Millisecond)
	cache.Set("100", time.Nanosecond)
	go cache.StartCleaning()
	time.Sleep(2 * time.Millisecond)
	cache.StopCleaning()
	assert.Equal(t, 0, cache.Len())
}
