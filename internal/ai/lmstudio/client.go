package lmstudio

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif" // Обязательно импортируйте нужные форматы с нижним подчеркиванием
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	authToken string
	apiURL    string
	deleteURL string
	modelName string

	client *http.Client
}

type Config struct {
	AuthToken string
	APIURL    string
	DeleteURL string
	ModelName string
}

func New(
	authToken,
	apiURL,
	deleteURL,
	modelName string,
) *Client {
	return &Client{
		authToken: authToken,
		apiURL:    apiURL,
		deleteURL: deleteURL,
		modelName: modelName,
		client:    &http.Client{},
	}
}

func NewDefault() *Client {
	return NewFromConfig(Config{})
}

func NewFromConfig(cfg Config) *Client {
	if cfg.APIURL == "" {
		cfg.APIURL = "http://localhost:1234/api/v1/chat"
	}
	if cfg.DeleteURL == "" {
		cfg.DeleteURL = strings.TrimRight(cfg.APIURL, "/") + "/{response_id}"
	}
	if cfg.ModelName == "" {
		cfg.ModelName = "google/gemma-4-e4b"
	}

	return New(
		cfg.AuthToken,
		cfg.APIURL,
		cfg.DeleteURL,
		cfg.ModelName,
	)
}

func (c Client) Req(ctx context.Context, prompt string, img []byte) (*ChatResponse, error) {
	return c.ReqImgs(ctx, prompt, [][]byte{img})
}

func (c Client) ReqImgs(ctx context.Context, prompt string, imgs [][]byte) (*ChatResponse, error) {
	return c.ReqImgsWithOptions(ctx, prompt, imgs, RequestOptions{
		Store: false,
	})
}

type RequestOptions struct {
	Store              bool
	PreviousResponseID string
}

func (c Client) ReqImgsWithOptions(ctx context.Context, prompt string, imgs [][]byte, opts RequestOptions) (*ChatResponse, error) {
	requestPayload := ChatRequest{
		Model: c.modelName,
		Input: []MessagePart{
			{
				Type:    "text",
				Content: prompt,
			},
		},
		ContextLength:      131072,
		Temperature:        0.0,
		Reasoning:          "on",
		Store:              opts.Store,
		PreviousResponseID: opts.PreviousResponseID,
	}
	for _, img := range imgs {
		if len(img) > 0 {
			mimeType := http.DetectContentType(img)
			t := "image"
			_, _, err := image.DecodeConfig(bytes.NewReader(img))
			if err != nil {
				t = "application"
			}
			requestPayload.Input = append(requestPayload.Input, MessagePart{
				Type:    t,
				DataURL: "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(img),
			})
		}
	}
	// 3. Маршалинг (преобразование) структуры Go в JSON-байт
	jsonBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при маршалинге payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("Ошибка при создании запроса: %w", err)
	}

	// 5. Установка заголовков (Headers), как в curl
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	req.Header.Set("Content-Type", "application/json")

	// 6. Выполнение запроса с помощью HTTP клиента
	client := c.client
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при выполнении запроса (Проверьте запущен ли API на localhost:1234): %w", err)
	}
	defer resp.Body.Close()

	// 7. Обработка ответа
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API вернул ошибку. Статус: %d. Тело ошибки: %s", resp.StatusCode, string(bodyBytes))
	}

	response := ChatResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("Не удалось распарсить тело ответа: %w", err)
	}

	return &response, nil
}

func (c Client) StartChat(ctx context.Context, prompt string, img []byte) (*ChatResponse, error) {
	return c.ReqImgsWithOptions(ctx, prompt, [][]byte{img}, RequestOptions{
		Store: true,
	})
}

func (c Client) ContinueChat(ctx context.Context, prompt string, previousResponseID string) (*ChatResponse, error) {
	return c.ReqImgsWithOptions(ctx, prompt, nil, RequestOptions{
		Store:              true,
		PreviousResponseID: previousResponseID,
	})
}

func (c Client) DeleteChat(ctx context.Context, responseID string) error {
	if responseID == "" {
		return nil
	}

	url := strings.ReplaceAll(c.deleteURL, "{response_id}", responseID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("Ошибка при создании запроса удаления чата: %w", err)
	}
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении чата: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API вернул ошибку удаления чата. Статус: %d. Тело ошибки: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c Client) GetMessage(ctx context.Context, prompt string, img []byte) (string, error) {
	items, err := c.Req(ctx, prompt, img)
	if err != nil {
		return "", err
	}
	for _, item := range items.Output {
		if item.Type == "message" {
			return item.Content, nil
		}
	}
	return "", fmt.Errorf("no message found: %v", items)
}
