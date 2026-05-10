package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dovbysh/duf.git/internal/models"
)

type ManticoreClient struct {
	db        *sql.DB
	tableName string
}

func NewClient(dsn, tableName string) (*ManticoreClient, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &ManticoreClient{db: db, tableName: tableName}, nil
}

// BatchReplace вставляет или обновляет пачку файлов
func (m *ManticoreClient) BatchReplace(files []models.FileRecord) error {
	if len(files) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(files))
	valueArgs := make([]interface{}, 0, len(files)*7)

	for _, f := range files {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, f.ID, f.Path, f.Name, f.Size, f.MTime, f.CTime, f.IsDeleted)
	}

	query := fmt.Sprintf("REPLACE INTO %s (id, path, name, size, mtime, ctime, is_deleted) VALUES %s",
		m.tableName, strings.Join(valueStrings, ","))

	_, err := m.db.Exec(query, valueArgs...)
	return err
}
