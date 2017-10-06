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

// Implementation of the MurmurHash3 optimized to handle only uint32 integers.
// A documentation about MurmurHash3 can be found at
// https://github.com/aappleby/smhasher/wiki/MurmurHash3

package murmur3

func Murmur_32(key uint32, seed uint32) uint32 {
	hash := seed
	key *= 0xcc9e2d51
	key = (key << 15) | (key >> 17)
	key *= 0x1b873593

	hash ^= key
	hash = (hash << 13) | (hash >> 19)
	hash *= 5
	hash += 0xe6546b64

	hash ^= 4

	hash ^= hash >> 16
	hash *= 0x85ebca6b
	hash ^= hash >> 13
	hash *= 0xc2b2ae35
	hash ^= hash >> 16

	return hash
}
