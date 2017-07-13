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
