package main

const (
	fnvOffset64 = uint64(14695981039346656037)
	fnvPrime64  = uint64(1099511628211)
)

// RolloutHash computes a stable, deterministic 64-bit FNV-1a hash over the
// flag key and the user identifier, separated by a single null byte.
func RolloutHash(key, user string) uint64 {
	h := fnvOffset64
	h = fnv1a(h, key)
	h = fnv1aByte(h, 0)
	h = fnv1a(h, user)
	return h
}

func fnv1a(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h = fnv1aByte(h, s[i])
	}
	return h
}

func fnv1aByte(h uint64, b byte) uint64 {
	h ^= uint64(b)
	h *= fnvPrime64
	return h
}
