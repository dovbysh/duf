package models

type FileRecord struct {
	ID        int64
	Path      string
	Name      string
	Size      int64
	MTime     int64
	CTime     int64
	SHA256    string
	IsDeleted uint32 // 0 или 1
}
