package models

import "encoding/json"

type FileRecord struct {
	ID        json.Number // Хеш от полного пути
	Path      string
	Name      string
	Size      int64
	MTime     int64
	CTime     int64
	SHA256    string
	IsDeleted uint32 // 0 или 1
}
