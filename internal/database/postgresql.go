package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dovbysh/duf.git/internal/models"
	_ "github.com/lib/pq"
)

type PostgresClient struct {
	db  *sql.DB
	dsn string
}

func NewClient(dsn string) (*PostgresClient, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return &PostgresClient{
		db:  db,
		dsn: dsn,
	}, nil
}

func (p *PostgresClient) Close() error {
	return p.db.Close()
}

func (p *PostgresClient) BatchReplace(ctx context.Context, files []models.FileRecord) error {
	if len(files) == 0 {
		return nil
	}

	const columnsPerRow = 7
	args := make([]any, 0, len(files)*columnsPerRow)
	values := make([]string, 0, len(files))

	for i, f := range files {
		base := i*columnsPerRow + 1
		values = append(values, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base,
			base+1,
			base+2,
			base+3,
			base+4,
			base+5,
			base+6,
		))
		args = append(args, f.Path, f.Name, f.Size, f.MTime, f.CTime, f.SHA256, f.IsDeleted)
	}

	query := fmt.Sprintf(`INSERT INTO duf.files AS existing
		(path, name, size, mtime, ctime, sha256, is_deleted)
		VALUES %s
		ON CONFLICT (path) DO UPDATE SET
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
		strings.Join(values, ","),
	)

	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

func (p *PostgresClient) MarkAllAsDeleted(ctx context.Context) error {
	_, err := p.db.ExecContext(ctx, "UPDATE duf.files SET is_deleted = 1, updated_at = now() WHERE is_deleted = 0")
	return err
}

func (p *PostgresClient) GetFilesWithoutHash(ctx context.Context, limit int32) ([]models.FileRecord, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id, path, name, size, mtime, ctime, sha256, is_deleted
		FROM duf.files
		WHERE is_deleted = 0 AND sha256 = ''
		ORDER BY id
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.FileRecord
	for rows.Next() {
		var file models.FileRecord
		if err := rows.Scan(
			&file.ID,
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
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (p *PostgresClient) UpdateHash(ctx context.Context, id int64, hash string) error {
	_, err := p.db.ExecContext(ctx, "UPDATE duf.files SET sha256 = $1, updated_at = now() WHERE id = $2", hash, id)
	return err
}
