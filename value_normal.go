// Copyright 2020 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

//go:build (!invariants && !tracing) || race
// +build !invariants,!tracing race

package mocache

import (
	"unsafe"
)

const valueSize = int(unsafe.Sizeof(Value{}))

func newValue(n int) *Value {
	if n == 0 {
		return nil
	}

	v := &Value{buf: make([]byte, n)}
	v.ref.init(1)
	return v
}

func (v *Value) free() {
}
