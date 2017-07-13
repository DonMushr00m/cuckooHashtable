package cuckoo

import (
	"crypto/rand"
	"cuckooHash/murmur"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
)

const (
	maxLen      = 16
	minIdxBytes = 10
	maxLoadFact = 0.5
)

type entry struct {
	key   uint32
	value uint32
}

type CuckooTable struct {
	entries  []*entry
	seed     uint32
	idxBytes uint32
	nEntries uint32
	nRehashs uint32
}

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

func (c *CuckooTable) getHashedKeys(key uint32) (uint32, uint32) {
	hash := murmur3.Murmur_32(key, c.seed)
	h1 := hash >> (32 - c.idxBytes)
	h2 := hash & uint32((1<<c.idxBytes)-1)
	return h1, h2
}

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

func (c *CuckooTable) Insert(key uint32, value uint32) bool {
	if _, exists := c.LookUp(key); exists {
		return false
	}

	h1, h2 := c.getHashedKeys(key)

	newEntry := &entry{key, value}
	index := h1
	tLen := 1 << c.idxBytes

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

	if c.nRehashs < 3 {
		c.rehash()
	} else {
		c.grow()
	}

	return c.Insert(newEntry.key, newEntry.value)
}

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

	if c.LoadFactor() < maxLoadFact/2 {
		c.shrink()
	}

}

func (c *CuckooTable) rehash() {
	c.nEntries = 0
	c.nRehashs += 1
	c.reorganize()
}

func (c *CuckooTable) grow() {
	c.idxBytes += 1
	c.nEntries = 0
	c.nRehashs = 0

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
	c.nRehashs = 0

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
