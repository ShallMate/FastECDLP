package btcec

// Hash is the internal hash type. Any change in its definition will require overall changes in this file.
type Hash uint32

const hashBits = 32 // # of bits in hash type, at most unsafe.Sizeof(Key)*8.

const (
	murmur3_c1_32 uint32 = 0xcc9e2d51
	murmur3_c2_32 uint32 = 0x1b873593
)

const (
	xx_prime32_1 uint32 = 2654435761
	xx_prime32_2 uint32 = 2246822519
	xx_prime32_3 uint32 = 3266489917
	xx_prime32_4 uint32 = 668265263
	xx_prime32_5 uint32 = 374761393
)

const (
	mem_c0 = 2860486313
	mem_c1 = 3267000013
)

func murmur3_32(k uint32, seed uint32) uint32 {
	k *= murmur3_c1_32
	k = (k << 15) | (k >> (32 - 15))
	k *= murmur3_c2_32

	h := seed
	h ^= k
	h = (h << 13) | (h >> (32 - 13))
	h = (h<<2 + h) + 0xe6546b64

	return h
}

func xx_32(k uint32, seed uint32) uint32 {
	h := seed + xx_prime32_5
	h += k * xx_prime32_3
	h = ((h << 17) | (h >> (32 - 17))) * xx_prime32_4
	h ^= h >> 15
	h *= xx_prime32_2
	h ^= h >> 13
	h *= xx_prime32_3
	h ^= h >> 16

	return h
}

func mem_32(k uint32, seed uint32) uint32 {
	h := k ^ mem_c0
	h ^= (k & 0xff) * mem_c1
	h ^= (k >> 8 & 0xff) * mem_c1
	h ^= (k >> 16 & 0xff) * mem_c1
	h ^= (k >> 24 & 0xff) * mem_c1

	return h
}
