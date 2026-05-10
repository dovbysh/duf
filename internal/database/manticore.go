package database

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dovbysh/duf.git/internal/models"
	manticore "github.com/manticoresoftware/manticoresearch-go"
)

type ManticoreClient struct {
	client    *manticore.APIClient
	tableName string
}

func NewClient(host string, tableName string) (*ManticoreClient, error) {
	config := manticore.NewConfiguration()
	config.Servers = manticore.ServerConfigurations{{URL: host}}
	client := manticore.NewAPIClient(config)

	return &ManticoreClient{
		client:    client,
		tableName: tableName,
	}, nil
}

func (m *ManticoreClient) InitSchema(ctx context.Context) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
        path text, 
        name text, 
        size bigint, 
        mtime bigint, 
        ctime bigint, 
        sha256 string, 
        is_deleted int
    ) attr_uint='size', attr_uint='mtime', attr_uint='ctime', attr_uint='is_deleted'`, m.tableName)

	_, _, err := m.client.UtilsAPI.Sql(ctx).Body(query).Execute()
	return err
}

func (m *ManticoreClient) BatchReplace(ctx context.Context, files []models.FileRecord) error {
	if len(files) == 0 {
		return nil
	}

	var bulkBody string
	for _, f := range files {
		insertLine := fmt.Sprintf(`{"replace": {"index": "%s", "id": %d, "doc": {"path": "%s", "name": "%s", "size": %d, "mtime": %d, "ctime": %d, "is_deleted": %d, "sha256": "%s"}}}`,
			m.tableName, f.ID, f.Path, f.Name, f.Size, f.MTime, f.CTime, f.IsDeleted, f.SHA256)
		bulkBody += insertLine + "\n"
	}

	_, _, err := m.client.IndexAPI.Bulk(ctx).Body(bulkBody).Execute()
	return err
}

func (m *ManticoreClient) MarkAllAsDeleted(ctx context.Context) error {
	query := fmt.Sprintf("UPDATE %s SET is_deleted = 1 WHERE is_deleted = 0", m.tableName)
	_, _, err := m.client.UtilsAPI.Sql(ctx).Body(query).Execute()
	return err
}

// GetFilesWithoutHash переписан на SQL для обхода ограничений типов SDK
func (m *ManticoreClient) GetFilesWithoutHash(ctx context.Context, limit int32) ([]models.FileRecord, error) {
	query := fmt.Sprintf("SELECT id, path FROM %s WHERE is_deleted = 0 AND sha256 = '' LIMIT %d", m.tableName, limit)

	resp, _, err := m.client.UtilsAPI.Sql(ctx).Body(query).Execute()
	if err != nil {
		return nil, err
	}

	// Ответ SQL API в Manticore Go SDK возвращается в специфическом формате (вложенные слайсы)
	if len(resp) == 0 {
		return nil, nil
	}

	var results []models.FileRecord

	// В Manticore SQL API ответ — это []map[string]interface{}
	for _, row := range resp {
		data := row["data"].([]interface{})
		for _, item := range data {
			rowMap := item.(map[string]interface{})

			// Извлекаем ID (он может прийти как float64 или string)
			idVal := fmt.Sprintf("%v", rowMap["id"])
			id, _ := strconv.ParseUint(idVal, 10, 64)

			results = append(results, models.FileRecord{
				ID:   id,
				Path: rowMap["path"].(string),
			})
		}
	}
	return results, nil
}

func (m *ManticoreClient) UpdateHash(ctx context.Context, id uint64, hash string) error {
	// Прямой SQL UPDATE — самый надежный способ
	query := fmt.Sprintf("UPDATE %s SET sha256 = '%s' WHERE id = %d", m.tableName, hash, id)
	_, _, err := m.client.UtilsAPI.Sql(ctx).Body(query).Execute()
	return err
}
