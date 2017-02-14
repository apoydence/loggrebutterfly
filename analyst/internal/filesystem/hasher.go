package filesystem

import "hash/fnv"

type StringHasher struct {
}

func NewHasher() *StringHasher {
	return &StringHasher{}
}

func (h *StringHasher) HashString(s string) (hash uint64) {
	f := fnv.New64a()
	f.Write([]byte(s))

	return f.Sum64()
}
