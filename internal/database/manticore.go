package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dovbysh/duf.git/internal/models"
	manticore "github.com/manticoresoftware/manticoresearch-go"
)

type ManticoreClient struct {
	httpAddr   string
	httpClient *http.Client
	client     *manticore.APIClient
	tableName  string
}

func NewClient(host string, tableName string) (*ManticoreClient, error) {
	config := manticore.NewConfiguration()
	config.Servers = manticore.ServerConfigurations{{URL: host}}
	client := manticore.NewAPIClient(config)

	return &ManticoreClient{
		client:     client,
		tableName:  tableName,
		httpAddr:   host,
		httpClient: &http.Client{},
	}, nil
}

func (m *ManticoreClient) InitSchema(ctx context.Context) error {
	// Убрана запятая после ")" и точка с запятой в конце
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
        path text, 
        name text, 
        size bigint, 
        mtime bigint, 
        ctime bigint, 
        sha256 string, 
        is_deleted int
    ) morphology='stem_en'`, m.tableName)

	_, _, err := m.client.UtilsAPI.Sql(ctx).Body(query).Execute()
	return err
}

func (m *ManticoreClient) BatchReplace(ctx context.Context, files []models.FileRecord) error {
	if len(files) == 0 {
		return nil
	}

	var bulkLines []string
	for _, f := range files {
		// Подготавливаем данные документа
		doc := map[string]interface{}{
			"path":       f.Path,
			"name":       f.Name,
			"size":       f.Size,
			"mtime":      f.MTime,
			"ctime":      f.CTime,
			"is_deleted": f.IsDeleted,
			"sha256":     f.SHA256,
		}

		// Создаем запрос Replace (аналог Insert в примере, но с перезаписью по ID)
		// Используем Marshal, чтобы SDK корректно экранировало спецсимволы в путях
		meta := map[string]interface{}{
			"replace": map[string]interface{}{
				"index": m.tableName,
				"id":    f.ID,
				"doc":   doc,
			},
		}

		b, err := json.Marshal(meta)
		if err != nil {
			continue
		}
		bulkLines = append(bulkLines, string(b))
	}

	// Соединяем строки через перенос (формат NDJSON для Bulk API)
	bulkBody := strings.Join(bulkLines, "\n") + "\n"

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
	// Construct the search request
	requestBody := map[string]interface{}{
		"table": m.tableName,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"equals": map[string]interface{}{"is_deleted": 0}},
					{"equals": map[string]interface{}{"sha256": ""}},
				},
			},
		},
		"limit": limit,
	}

	bodyBytes, _ := json.Marshal(requestBody)

	req, err := http.NewRequestWithContext(ctx, "POST", m.httpAddr+"/search", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// CRITICAL: Use json.Decoder with UseNumber to preserve uint64 precision
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	var result struct {
		Hits struct {
			Hits []struct {
				ID     json.Number            `json:"_id"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}

	var files []models.FileRecord
	for _, hit := range result.Hits.Hits {
		// Parse ID from json.Number (string) to uint64

		path, _ := hit.Source["path"].(string)
		if path == "" {
			continue
		}
		files = append(files, models.FileRecord{
			ID:   hit.ID,
			Path: path,
		})
	}

	return files, nil
}

func (m *ManticoreClient) UpdateHash(ctx context.Context, id json.Number, hash string) error {
	// Construct the update request body
	requestBody := map[string]interface{}{
		"table": m.tableName,
		"id":    id,
		"doc": map[string]interface{}{
			"sha256": hash,
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("marshal update error: %w", err)
	}

	// Send POST request to /update
	req, err := http.NewRequestWithContext(ctx, "POST", m.httpAddr+"/update", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request error: %w", err)
	}
	defer resp.Body.Close()

	// Check for Manticore errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("manticore update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
