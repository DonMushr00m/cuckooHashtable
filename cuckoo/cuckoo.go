// Copyright 2017 Nicolas Forster
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file implements a cuckoo hash table using two different hashes of the
// key to find an appropriate place for the element to be inserted.
// The maximum table size is 2^16 elements.

package cuckoo

import (
	"crypto/rand"
	"cuckooHash/murmur"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"bitbucket/cuckoohash/murmur"
)

const (
	maxLen      = 16
	minIdxBytes = 10
	maxLoadFact = 0.5
)

// The key has to be uint32. But one can adjust the value to whatever type one wants.
type entry struct {
	key   uint32
	value uint32
}

type CuckooTable struct {
	entries   []*entry
	seed      uint32
	idxBytes  uint32
	nEntries  uint32
	nRehashes uint32
}

// resetSeed() resets the current seed. Used during a rehash of the table.
func (c *CuckooTable) resetSeed() {
	s := make([]byte, 4)
	_, err := rand.Read(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	c.seed = binary.LittleEndian.Uint32(s)
}

func NewCuckoo() *CuckooTable {
	initLen := 1 << minIdxBytes
	entries := make([]*entry, initLen)
	c := &CuckooTable{
		entries:  entries,
		idxBytes: minIdxBytes,
	}
	c.resetSeed()

	return c
}

// getHashedKeys() generates the two hashed keys. Important to note is that
// only one hash is generated. This hash is then split up into the two
// hashed key values used for inserting/finding an object.
func (c *CuckooTable) getHashedKeys(key uint32) (uint32, uint32) {
	hash := murmur3.Murmur_32(key, c.seed)
	h1 := hash >> (32 - c.idxBytes)
	h2 := hash & uint32((1<<c.idxBytes)-1)
	return h1, h2
}

// LookUp() looks an element up in the table.
func (c *CuckooTable) LookUp(key uint32) (uint32, bool) {
	h1, h2 := c.getHashedKeys(key)
	if entry := c.entries[h1]; entry != nil && entry.key == key {
		return entry.value, true
	}

	if entry := c.entries[h2]; entry != nil && entry.key == key {
		return entry.value, true
	}

	return 0, false
}

// Insert() inserts an element at the appropriate position in the table.
func (c *CuckooTable) Insert(key uint32, value uint32) bool {
	if _, exists := c.LookUp(key); exists {
		return false
	}

	h1, h2 := c.getHashedKeys(key)

	newEntry := &entry{key, value}
	index := h1
	tLen := 1 << c.idxBytes

	// reorder the elements in the table until all elements found a place,
	// or tLen reordering steps have been done (to avoid an infinite loop).
	for count := 0; count < tLen; count++ {
		oldEntry := c.entries[index]
		c.entries[index] = newEntry

		if oldEntry == nil {
			c.nEntries += 1
			return true
		}

		h1, h2 = c.getHashedKeys(oldEntry.key)

		if index == h1 {
			index = h2
		} else {
			index = h1
		}

		newEntry = oldEntry
	}

	// If no stable table configuration can be found, first try to rehash the table
	// if that does not help, grow the table.
	if c.nRehashes < 3 {
		c.rehash()
	} else {
		c.grow()
	}

	return c.Insert(newEntry.key, newEntry.value)
}

// Delete() deletes an element.
func (c *CuckooTable) Delete(key uint32) {
	h1, h2 := c.getHashedKeys(key)
	if entry := c.entries[h1]; entry != nil && entry.key == key {
		c.entries[h1] = nil
		c.nEntries -= 1
	}

	if entry := c.entries[h2]; entry != nil && entry.key == key {
		c.entries[h2] = nil
		c.nEntries -= 1
	}

	// If the load factor of the table is too low, shrink the table.
	if c.LoadFactor() < maxLoadFact/2 {
		c.shrink()
	}

}

func (c *CuckooTable) rehash() {
	c.nEntries = 0
	c.nRehashes += 1
	c.reorganize()
}

func (c *CuckooTable) grow() {
	c.idxBytes += 1
	c.nEntries = 0
	c.nRehashes = 0

	if c.idxBytes > maxLen {
		panic("Too many elements")
	}

	c.reorganize()
}

func (c *CuckooTable) shrink() {
	if c.idxBytes <= minIdxBytes {
		return
	}
	c.idxBytes -= 1
	c.nEntries = 0
	c.nRehashes = 0

	c.reorganize()
}

func (c *CuckooTable) reorganize() {
	tempTab := &CuckooTable{}
	*tempTab = *c
	c.resetSeed()

	c.entries = make([]*entry, 1<<c.idxBytes)

	for _, val := range tempTab.entries {
		if val != nil {
			c.Insert(val.key, val.value)
		}
	}

	defer func() {
		tempTab = nil
		runtime.GC()
	}()
}

func (c *CuckooTable) LoadFactor() float64 {
	tLen := 1 << c.idxBytes
	return float64(c.nEntries) / float64(tLen)
}
