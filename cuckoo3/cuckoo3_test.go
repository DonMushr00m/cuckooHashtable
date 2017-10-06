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

package cuckoo3

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
)

var nOfItems = 60000

var (
	keys      []uint32
	values    []uint32
	keyValMap map[uint32]uint32
)

var (
	mapBytes uint64
	mBench   map[uint32]uint32
	cBench   CuckooTable3
)

func readAlloc() uint64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.Alloc
}

func init() {
	keys = make([]uint32, nOfItems)
	values = make([]uint32, nOfItems)

	runtime.GC()
	before := readAlloc()

	keyValMap = make(map[uint32]uint32)
	for i := 0; i < nOfItems; i++ {
		k := rand.Uint32()
		v := rand.Uint32()
		keyValMap[k] = v
		keys[i] = k
		values[i] = v
	}

	after := readAlloc()
	mapBytes = after - before
}

func TestInsertAndSearch(t *testing.T) {
	c := NewCuckoo()

	for key, value := range keyValMap {
		c.Insert(key, value)
	}

	for key, value := range keyValMap {
		elem, ok := c.LookUp(key)

		if !ok || elem != value {
			t.Error("search failed")
		}
	}
}

func TestDelete(t *testing.T) {
	c := NewCuckoo()

	for i := 0; i < 2000; i++ {
		c.Insert(keys[i], values[i])
	}

	for i := 0; i < 2000; i++ {
		c.Delete(keys[i])
	}

	for i := 0; i < 2000; i++ {
		_, ok := c.LookUp(keys[i])

		if ok {
			t.Error("Delete failed")
		}
	}
}

func TestMemory(t *testing.T) {
	runtime.GC()
	before := readAlloc()
	c := NewCuckoo()

	for key, value := range keyValMap {
		c.Insert(key, value)
	}

	after := readAlloc()
	cBytes := after - before

	fmt.Print("MEMTEST")
	t.Log("Built-in map mem usage (MiB):", float64(mapBytes)/float64(1<<20))
	t.Log("Cuckoo hash mem usage (MiB):", float64(cBytes)/float64(1<<20))
	t.Log("Cuckoo hash LoadFactor:", c.LoadFactor())
}

func BenchmarkCuckooInsert(b *testing.B) {
	cBench = *NewCuckoo()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cBench.Insert(keys[i%nOfItems], values[i%nOfItems])
	}
}

func BenchmarkCuckooLookUp(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cBench.LookUp(keys[i%nOfItems])
	}
}

func BenchmarkCuckooTable_Delete(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cBench.Delete(keys[i%nOfItems])
	}
}

func BenchmarkMapInsert(b *testing.B) {
	mBench = make(map[uint32]uint32)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := keys[i%nOfItems]
		value := values[i%nOfItems]
		mBench[key] = value
	}
}

func BenchmarkMapSearch(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = mBench[keys[i%nOfItems]]
	}
}

func BenchmarkMapDelete(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		delete(mBench, keys[i%nOfItems])
	}
}
