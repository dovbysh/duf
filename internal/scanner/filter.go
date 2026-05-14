package scanner

import (
	"path/filepath"
	"strings"
)

type Filter struct {
	patterns []string
}

func NewFilter(patterns []string) *Filter {
	return &Filter{patterns: patterns}
}

// IsExcluded проверяет, попадает ли путь под шаблоны исключений
func (f *Filter) IsExcluded(path string) bool {
	cleanPath := filepath.Clean(path)

	for _, pattern := range f.patterns {
		cleanPattern := filepath.Clean(pattern)

		// Проверяем как полное совпадение, так и совпадение по маске
		match, err := filepath.Match(cleanPattern, cleanPath)
		if err == nil && match {
			return true
		}
		// Дополнительная проверка для вложенных папок через Base
		match, err = filepath.Match(cleanPattern, filepath.Base(cleanPath))
		if err == nil && match {
			return true
		}
		if matchDoubleStar(cleanPattern, cleanPath) {
			return true
		}
	}
	return false
}

func matchDoubleStar(pattern string, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	if strings.HasPrefix(pattern, "**/") && strings.HasSuffix(pattern, "/**") {
		middle := strings.TrimSuffix(strings.TrimPrefix(pattern, "**/"), "/**")
		if middle == "" {
			return false
		}

		return path == middle ||
			strings.HasSuffix(path, "/"+middle) ||
			strings.Contains(path, "/"+middle+"/")
	}

	return false
}
