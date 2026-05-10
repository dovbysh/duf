package database

import (
	"fmt"

	"github.com/dovbysh/duf.git/internal/models"
)

// GetFilesWithoutHash выбирает записи, где хеш еще не рассчитан
func (m *ManticoreClient) GetFilesWithoutHash(limit int) ([]models.FileRecord, error) {
	query := fmt.Sprintf("SELECT id, path FROM %s WHERE sha256 = '' AND is_deleted = 0 LIMIT %d", m.tableName, limit)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.FileRecord
	for rows.Next() {
		var f models.FileRecord
		if err := rows.Scan(&f.ID, &f.Path); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

// UpdateHash обновляет только поле sha256
func (m *ManticoreClient) UpdateHash(id uint64, hash string) error {
	query := fmt.Sprintf("UPDATE %s SET sha256 = ? WHERE id = ?", m.tableName)
	_, err := m.db.Exec(query, hash, id)
	return err
}

// MarkAllAsDeleted сбрасывает флаг перед проверкой
func (m *ManticoreClient) MarkAllAsDeleted() error {
	query := fmt.Sprintf("UPDATE %s SET is_deleted = 1", m.tableName)
	_, err := m.db.Exec(query)
	return err
}
