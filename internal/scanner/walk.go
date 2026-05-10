package scanner

import (
	"io/fs"

	"path/filepath"

	"github.com/dovbysh/duf.git/internal/models"
)

type Scanner struct {
	filter *Filter
}

func NewScanner(excludePatterns []string) *Scanner {
	return &Scanner{
		filter: NewFilter(excludePatterns),
	}
}

// Scan запускает обход директории и отправляет результаты в канал
func (s *Scanner) Scan(root string, results chan<- models.FileRecord) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Пропускаем ошибки доступа к файлам
		}

		if s.filter.IsExcluded(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil
			}

			// Получаем ctime (зависит от ОС, в Linux через syscall)
			// Для простоты здесь используем mtime, но в полной версии можно вытянуть ctime через stat_t
			stat := info.Sys()
			_ = stat

			results <- models.FileRecord{
				ID:        GenerateID(path),
				Path:      path,
				Name:      d.Name(),
				Size:      info.Size(),
				MTime:     info.ModTime().Unix(),
				CTime:     info.ModTime().Unix(), // В Linux честный ctime требует syscall.Stat_t
				IsDeleted: 0,
			}
		}
		return nil
	})
}
