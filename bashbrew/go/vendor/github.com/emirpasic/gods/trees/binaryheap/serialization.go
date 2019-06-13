// Copyright (c) 2015, Emir Pasic. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binaryheap

import "github.com/emirpasic/gods/containers"

func assertSerializationImplementation() {
	var _ containers.JSONSerializer = (*Heap)(nil)
	var _ containers.JSONDeserializer = (*Heap)(nil)
}

// ToJSON outputs the JSON representation of list's elements.
func (heap *Heap) ToJSON() ([]byte, error) {
	return heap.list.ToJSON()
}

// FromJSON populates list's elements from the input JSON representation.
func (heap *Heap) FromJSON(data []byte) error {
	return heap.list.FromJSON(data)
}
