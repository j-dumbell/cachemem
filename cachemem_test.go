package cachemem

import (
	"reflect"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	cache := New[int, string]()
	key := 1
	value := "hello"
	cache.Set(key, value)

	actual, ok := cache.Get(key)
	if ok != true {
		t.Fatalf("expected ok to be true")
	}

	if actual != value {
		t.Fatalf("actual value != expected.  Actual = %v; expected = %v", actual, value)
	}
}

func TestSetWithExpiry_expired(t *testing.T) {
	cache := New[int, string]()
	cache.SetWithExpiry(1, "blah", time.Millisecond*10)
	time.Sleep(time.Millisecond * 20)

	actualLen := cache.Len()
	if actualLen != 0 {
		t.Fatalf("expired record not deleted")
	}

	_, ok := cache.Get(1)
	if ok != false {
		t.Fatalf("expected ok to be false")
	}
}

func TestSetWithExpiry_notExpired(t *testing.T) {
	type Value struct {
		A string
	}
	value := Value{"a"}

	cache := New[int, Value]()
	cache.SetWithExpiry(1, value, time.Second*10)

	actualLen := cache.Len()
	if actualLen != 1 {
		t.Fatalf("expected 1 record in cache")
	}

	actual, ok := cache.Get(1)
	if !reflect.DeepEqual(actual, value) {
		t.Fatalf("actual value != expected.  Actual = %v; expected = %v", actual, value)
	}
	if ok != true {
		t.Fatalf("expected ok to be true")
	}

}

func TestGet_KeyNotExists(t *testing.T) {
	cache := New[uint32, struct{}]()

	_, ok := cache.Get(2)
	if ok != false {
		t.Fatalf("expected ok to be false")
	}
}

func TestTruncate(t *testing.T) {
	cache := New[int, uint]()
	cache.Set(1, 10)
	cache.Set(2, 20)

	cache.Clear()
	_, ok1 := cache.Get(1)
	_, ok2 := cache.Get(2)

	if ok1 != false || ok2 != false {
		t.Fatalf("expected cache to be empty")
	}
}

func TestLength(t *testing.T) {
	cache := New[string, int]()
	cache.Set("a", 1)
	cache.Set("b", 2)

	actual := cache.Len()
	expected := 2
	if actual != expected {
		t.Fatalf("expected length of %v; got %v", expected, actual)
	}
}
