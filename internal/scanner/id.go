package scanner

import (
	"hash/fnv"
)

// GenerateID создает 64-битный хеш из полного пути файла
func GenerateID(path string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(path))
	return h.Sum64()
}
