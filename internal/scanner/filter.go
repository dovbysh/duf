package scanner

import (
	"path/filepath"
)

type Filter struct {
	patterns []string
}

func NewFilter(patterns []string) *Filter {
	return &Filter{patterns: patterns}
}

// IsExcluded проверяет, попадает ли путь под шаблоны исключений
func (f *Filter) IsExcluded(path string) bool {
	for _, pattern := range f.patterns {
		// Проверяем как полное совпадение, так и совпадение по маске
		match, err := filepath.Match(pattern, path)
		if err == nil && match {
			return true
		}
		// Дополнительная проверка для вложенных папок через Base
		match, err = filepath.Match(pattern, filepath.Base(path))
		if err == nil && match {
			return true
		}
	}
	return false
}
