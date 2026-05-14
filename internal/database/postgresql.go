package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dovbysh/duf.git/internal/models"
	"github.com/lib/pq"
)

type PostgresClient struct {
	db        *sql.DB
	tableName string
}

func NewClient(dsn string, tableName string) (*PostgresClient, error) {
	quotedTableName, err := quoteQualifiedIdentifier(tableName)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return &PostgresClient{
		db:        db,
		tableName: quotedTableName,
	}, nil
}

func (p *PostgresClient) Close() error {
	return p.db.Close()
}

func (p *PostgresClient) BatchReplace(ctx context.Context, files []models.FileRecord) error {
	if len(files) == 0 {
		return nil
	}

	const columnsPerRow = 8
	args := make([]any, 0, len(files)*columnsPerRow)
	values := make([]string, 0, len(files))

	for i, f := range files {
		base := i*columnsPerRow + 1
		values = append(values, fmt.Sprintf(
			"($%d::numeric, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base,
			base+1,
			base+2,
			base+3,
			base+4,
			base+5,
			base+6,
			base+7,
		))
		args = append(args, f.ID.String(), f.Path, f.Name, f.Size, f.MTime, f.CTime, f.SHA256, f.IsDeleted)
	}

	query := fmt.Sprintf(`INSERT INTO %s AS existing
		(id, path, name, size, mtime, ctime, sha256, is_deleted)
		VALUES %s
		ON CONFLICT (id) DO UPDATE SET
			path = EXCLUDED.path,
			name = EXCLUDED.name,
			size = EXCLUDED.size,
			mtime = EXCLUDED.mtime,
			ctime = EXCLUDED.ctime,
			is_deleted = EXCLUDED.is_deleted,
			sha256 = CASE
				WHEN existing.size IS DISTINCT FROM EXCLUDED.size
					OR existing.mtime IS DISTINCT FROM EXCLUDED.mtime
				THEN ''
				ELSE existing.sha256
			END,
			updated_at = now()`,
		p.tableName,
		strings.Join(values, ","),
	)

	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

func (p *PostgresClient) MarkAllAsDeleted(ctx context.Context) error {
	query := fmt.Sprintf("UPDATE %s SET is_deleted = 1, updated_at = now() WHERE is_deleted = 0", p.tableName)
	_, err := p.db.ExecContext(ctx, query)
	return err
}

func (p *PostgresClient) GetFilesWithoutHash(ctx context.Context, limit int32) ([]models.FileRecord, error) {
	query := fmt.Sprintf(`SELECT id::text, path, name, size, mtime, ctime, sha256, is_deleted
		FROM %s
		WHERE is_deleted = 0 AND sha256 = ''
		ORDER BY id
		LIMIT $1`, p.tableName)

	rows, err := p.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.FileRecord
	for rows.Next() {
		var id string
		var file models.FileRecord
		if err := rows.Scan(
			&id,
			&file.Path,
			&file.Name,
			&file.Size,
			&file.MTime,
			&file.CTime,
			&file.SHA256,
			&file.IsDeleted,
		); err != nil {
			return nil, err
		}
		file.ID = json.Number(id)
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (p *PostgresClient) UpdateHash(ctx context.Context, id json.Number, hash string) error {
	query := fmt.Sprintf("UPDATE %s SET sha256 = $1, updated_at = now() WHERE id = $2::numeric", p.tableName)
	_, err := p.db.ExecContext(ctx, query, hash, id.String())
	return err
}

func quoteQualifiedIdentifier(name string) (string, error) {
	parts := strings.Split(name, ".")
	quoted := make([]string, 0, len(parts))

	for _, part := range parts {
		if !isIdentifier(part) {
			return "", fmt.Errorf("invalid database table name: %q", name)
		}
		quoted = append(quoted, pq.QuoteIdentifier(part))
	}

	return strings.Join(quoted, "."), nil
}

func isIdentifier(value string) bool {
	if value == "" {
		return false
	}

	for i, r := range value {
		if i == 0 {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != '_' {
				return false
			}
			continue
		}

		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}

	return true
}
