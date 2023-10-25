// Copyright 2020 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.
//
//go:build (!invariants && !tracing) || race
// +build !invariants,!tracing race

package mocache

import (
	"sync"
	"unsafe"

	"github.com/nnsgmsone/mocache/internal/invariants"
)

const (
	entrySize            = int(unsafe.Sizeof(entry{}))
	entryAllocCacheLimit = 128
	entriesGoAllocated   = invariants.RaceEnabled
)

var entryAllocPool = sync.Pool{
	New: func() interface{} {
		return newEntryAllocCache()
	},
}

func entryAllocNew() *entry {
	a := entryAllocPool.Get().(*entryAllocCache)
	e := a.alloc()
	entryAllocPool.Put(a)
	return e
}

func entryAllocFree(e *entry) {
	a := entryAllocPool.Get().(*entryAllocCache)
	a.free(e)
	entryAllocPool.Put(a)
}

type entryAllocCache struct {
	entries []*entry
}

func newEntryAllocCache() *entryAllocCache {
	c := &entryAllocCache{}
	return c
}

func freeEntryAllocCache(obj interface{}) {
	c := obj.(*entryAllocCache)
	for i, e := range c.entries {
		c.dealloc(e)
		c.entries[i] = nil
	}
}

func (c *entryAllocCache) alloc() *entry {
	n := len(c.entries)
	if n == 0 {
		return &entry{}
	}
	e := c.entries[n-1]
	c.entries = c.entries[:n-1]
	return e
}

func (c *entryAllocCache) dealloc(e *entry) {
}

func (c *entryAllocCache) free(e *entry) {
	if len(c.entries) == entryAllocCacheLimit {
		c.dealloc(e)
		return
	}
	c.entries = append(c.entries, e)
}
