package scanner

import (
	"encoding/json"
	"hash/fnv"
	"strconv"
)

// GenerateID создает 64-битный хеш из полного пути файла
func GenerateID(path string) json.Number {
	h := fnv.New64a()
	h.Write([]byte(path))
	return json.Number(strconv.FormatUint(h.Sum64(), 10))
}
