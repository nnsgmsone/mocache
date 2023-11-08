// Copyright 2020 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package manual

import (
	"sync/atomic"
	"syscall"
	"unsafe"
)

var fd = -1
var AllocSize int64

// The go:linkname directives provides backdoor access to private functions in
// the runtime. Below we're accessing the throw function.

//go:linkname throw runtime.throw
func throw(s string)

// New allocates a slice of size n. The returned slice is from manually managed
// memory and MUST be released by calling Free. Failure to do so will result in
// a memory leak.
func New(n int) []byte {
	if n == 0 {
		return make([]byte, 0)
	}
	size := rollup(n)
	r0, _, e1 := syscall.Syscall6(syscall.SYS_MMAP, 0, uintptr(size), uintptr(syscall.PROT_READ|syscall.PROT_WRITE),
		uintptr(syscall.MAP_ANON|syscall.MAP_PRIVATE), uintptr(fd), uintptr(0))
	if e1 != 0 {
		throw("out of memory")
	}
	atomic.AddInt64(&AllocSize, int64(size))
	return unsafe.Slice((*byte)(unsafe.Pointer(r0)), n)
}

func Free(b []byte) {
	size := int64(rollup(cap(b)))
	atomic.AddInt64(&AllocSize, -size)
	syscall.Syscall(syscall.SYS_MUNMAP, uintptr(unsafe.Pointer(&b[0])), uintptr(size), 0)
}

func rollup(n int) int {
	return (n + 4095) & (^4095)
}
