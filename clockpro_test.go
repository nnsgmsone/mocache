// Copyright 2018. All rights reserved. Use of this source code is governed by
// an MIT-style license that can be found in the LICENSE file.

package mocache

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
)

func testValue(cache *Cache, s string, repeat int) *Value {
	b := bytes.Repeat([]byte(s), repeat)
	v := Alloc(len(b))
	copy(v.Buf(), b)
	return v
}

func TestCacheDelete(t *testing.T) {
	cache := newShards(100, 1)
	defer cache.Unref()

	cache.Set("10", 0, testValue(cache, "a", 5)).Release()
	cache.Set("11", 0, testValue(cache, "a", 5)).Release()
	cache.Set("12", 0, testValue(cache, "a", 5)).Release()
	if expected, size := int64(15), cache.Size(); expected != size {
		t.Fatalf("expected cache size %d, but found %d", expected, size)
	}
	cache.Delete("11", 0)
	if expected, size := int64(10), cache.Size(); expected != size {
		t.Fatalf("expected cache size %d, but found %d", expected, size)
	}
	if h := cache.Get("10", 0); h.Get() == nil {
		t.Fatalf("expected to find block 0/0")
	} else {
		h.Release()
	}
	if h := cache.Get("11", 0); h.Get() != nil {
		t.Fatalf("expected to not find block 1/0")
	} else {
		h.Release()
	}
	// Deleting a non-existing block does nothing.
	cache.Delete("11", 0)
	if expected, size := int64(10), cache.Size(); expected != size {
		t.Fatalf("expected cache size %d, but found %d", expected, size)
	}
}

func TestEvictFile(t *testing.T) {
	cache := newShards(100, 1)
	defer cache.Unref()

	cache.Set("10", 0, testValue(cache, "a", 5)).Release()
	cache.Set("11", 0, testValue(cache, "a", 5)).Release()
	cache.Set("12", 0, testValue(cache, "a", 5)).Release()
	cache.Set("12", 1, testValue(cache, "a", 5)).Release()
	cache.Set("12", 2, testValue(cache, "a", 5)).Release()
	if expected, size := int64(25), cache.Size(); expected != size {
		t.Fatalf("expected cache size %d, but found %d", expected, size)
	}
	cache.EvictFile("10")
	if expected, size := int64(20), cache.Size(); expected != size {
		t.Fatalf("expected cache size %d, but found %d", expected, size)
	}
	cache.EvictFile("11")
	if expected, size := int64(15), cache.Size(); expected != size {
		t.Fatalf("expected cache size %d, but found %d", expected, size)
	}
	cache.EvictFile("12")
	if expected, size := int64(0), cache.Size(); expected != size {
		t.Fatalf("expected cache size %d, but found %d", expected, size)
	}
}

func TestEvictAll(t *testing.T) {
	// Verify that it is okay to evict all of the data from a cache. Previously
	// this would trigger a nil-pointer dereference.
	cache := newShards(100, 1)
	defer cache.Unref()

	cache.Set("1", 0, testValue(cache, "a", 101)).Release()
	cache.Set("1", 0, testValue(cache, "a", 101)).Release()
}

func TestMultipleDBs(t *testing.T) {
	cache := newShards(100, 1)
	defer cache.Unref()

	cache.Set("1", 0, testValue(cache, "a", 5)).Release()
	cache.Set("2", 0, testValue(cache, "b", 5)).Release()
	if expected, size := int64(10), cache.Size(); expected != size {
		t.Fatalf("expected cache size %d, but found %d", expected, size)
	}
	cache.EvictFile("1")
	if expected, size := int64(5), cache.Size(); expected != size {
		t.Fatalf("expected cache size %d, but found %d", expected, size)
	}
	h := cache.Get("1", 0)
	if v := h.Get(); v != nil {
		t.Fatalf("expected not present, but found %s", v)
	}
	h = cache.Get("2", 0)
	if v := h.Get(); string(v) != "bbbbb" {
		t.Fatalf("expected bbbbb, but found %s", v)
	} else {
		h.Release()
	}
}

func TestZeroSize(t *testing.T) {
	cache := newShards(0, 1)
	defer cache.Unref()

	cache.Set("1", 0, testValue(cache, "a", 5)).Release()
}

func TestReserve(t *testing.T) {
	cache := newShards(4, 2)
	defer cache.Unref()

	cache.Set("1", 0, testValue(cache, "a", 1)).Release()
	cache.Set("2", 0, testValue(cache, "a", 1)).Release()
	require.EqualValues(t, 2, cache.Size())
	r := cache.Reserve(1)
	require.EqualValues(t, 0, cache.Size())
	cache.Set("1", 0, testValue(cache, "a", 1)).Release()
	cache.Set("2", 0, testValue(cache, "a", 1)).Release()
	cache.Set("3", 0, testValue(cache, "a", 1)).Release()
	cache.Set("4", 0, testValue(cache, "a", 1)).Release()
	require.EqualValues(t, 2, cache.Size())
	r()
	require.EqualValues(t, 2, cache.Size())
	cache.Set("1", 0, testValue(cache, "a", 1)).Release()
	cache.Set("2", 0, testValue(cache, "a", 1)).Release()
	require.EqualValues(t, 4, cache.Size())
}

func TestReserveDoubleRelease(t *testing.T) {
	cache := newShards(100, 1)
	defer cache.Unref()

	r := cache.Reserve(10)
	r()

	result := func() (result string) {
		defer func() {
			if v := recover(); v != nil {
				result = fmt.Sprint(v)
			}
		}()
		r()
		return ""
	}()
	const expected = "pebble: cache reservation already released"
	if expected != result {
		t.Fatalf("expected %q, but found %q", expected, result)
	}
}

func TestCacheStressSetExisting(t *testing.T) {
	cache := newShards(1, 1)
	defer cache.Unref()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				cache.Set("1", uint64(i), testValue(cache, "a", 1)).Release()
				runtime.Gosched()
			}
		}(i)
	}
	wg.Wait()
}

func BenchmarkCacheGet(b *testing.B) {
	const size = 100000

	cache := newShards(size, 1)
	defer cache.Unref()

	for i := 0; i < size; i++ {
		v := testValue(cache, "a", 1)
		cache.Set("1", uint64(i), v).Release()
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

		for pb.Next() {
			h := cache.Get("1", uint64(rng.Intn(size)))
			if h.Get() == nil {
				b.Fatal("failed to lookup value")
			}
			h.Release()
		}
	})
}

func TestReserveColdTarget(t *testing.T) {
	// If coldTarget isn't updated when we call shard.Reserve,
	// then we unnecessarily remove nodes from the
	// cache.

	cache := newShards(100, 1)
	defer cache.Unref()

	for i := 0; i < 50; i++ {
		cache.Set(fmt.Sprintf("%v", i+1), 0, testValue(cache, "a", 1)).Release()
	}

	if cache.Size() != 50 {
		require.Equal(t, 50, cache.Size(), "nodes were unnecessarily evicted from the cache")
	}

	// There won't be enough space left for 50 nodes in the cache after
	// we call shard.Reserve. This should trigger a call to evict.
	cache.Reserve(51)

	// If we don't update coldTarget in Reserve then the cache gets emptied to
	// size 0. In shard.Evict, we loop until shard.Size() < shard.targetSize().
	// Therefore, 100 - 51 = 49, but we evict one more node.
	if cache.Size() != 48 {
		t.Fatalf("expected positive cache size %d, but found %d", 48, cache.Size())
	}
}
