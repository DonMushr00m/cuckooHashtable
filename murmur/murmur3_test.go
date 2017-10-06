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

package murmur3

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

var data = []struct {
	hash uint32
	val  uint32
}{
	{0x2362f9de, 0},
	{0xc8ce92e3, 1024},
	{0x9ef181ca, 675431},
	{0x76293b50, 4294967295},
	{0x156c5f38, 987654321},
}

func TestReference(t *testing.T) {
	for i := 0; i < len(data); i++ {
		res := Murmur_32(data[i].val, 0)
		if res != data[i].hash {
			t.Errorf("murmur3 failed for: Got 0x%x, expected 0x%x\n", res, data[i].hash)
		}
	}
}

func getSeed() uint32 {
	s := make([]byte, 4)
	_, err := rand.Read(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	return binary.LittleEndian.Uint32(s)
}

func BenchmarkMurmur3_32(b *testing.B) {
	seed := getSeed()
	input := generateInput(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = Murmur_32(input[i], seed)
	}
}

func generateInput(length int) []uint32 {
	result := make([]uint32, length)
	source := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(source)

	for i := 0; i < length; i++ {
		result[i] = rnd.Uint32()
	}

	return result
}
