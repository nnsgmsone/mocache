// Copyright 2020 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package manual

func New(n int) []byte {
	return make([]byte, n)
}

func Free(_ []byte) {
}
