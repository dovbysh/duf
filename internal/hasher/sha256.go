package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// CalculateSHA256 читает файл буферами и возвращает hex-строку
func CalculateSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	// Используем буфер 1MB для HDD (как указали в ТЗ)
	if _, err := io.CopyBuffer(h, f, make([]byte, 1024*1024)); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
