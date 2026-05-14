package lmstudio

import "fmt"

// ChatRequest - Главная структура, соответствующая всему JSON-телу.
type ChatRequest struct {
	Model         string        `json:"model"`
	Input         []MessagePart `json:"input"`
	ContextLength int           `json:"context_length"`
	Temperature   float64       `json:"temperature"`
	Reasoning     string        `json:"reasoning"`
	Store         bool          `json:"store"`
}

// MessagePart - Структура, описывающая один элемент в массиве "input".
type MessagePart struct {
	Type    string `json:"type"`               // "text" или "image"
	Content string `json:"content,omitempty"`  // Используется только для типа "text"
	DataURL string `json:"data_url,omitempty"` // Используется только для типа "image"
}

// ChatResponse - Главная структура, представляющая весь JSON-объект.
type ChatResponse struct {
	ModelInstanceID string       `json:"model_instance_id"`
	Output          []OutputItem `json:"output"` // Это массив структур OutputItem
	Stats           Stats        `json:"stats"`  // Это структура статистики
	ResponseID      string       `json:"response_id"`
}

// OutputItem - Структура для элементов, содержащихся в массиве "output".
type OutputItem struct {
	Type    string `json:"type"` // Например, "message"
	Content string `json:"content"`
}

// Stats - Структура для статистики (метрики запроса).
type Stats struct {
	InputTokens             int     `json:"input_tokens"`
	TotalOutputTokens       int     `json:"total_output_tokens"`
	ReasoningOutputTokens   int     `json:"reasoning_output_tokens"`
	TokensPerSecond         float64 `json:"tokens_per_second"`
	TimeToFirstTokenSeconds float64 `json:"time_to_first_token_seconds"`
}

func (ch ChatResponse) GetMessage() (string, error) {
	for _, item := range ch.Output {
		if item.Type == "message" {
			return item.Content, nil
		}
	}
	return "", fmt.Errorf("no message found: %v", ch)
}

func (ch ChatResponse) GetReason() (string, error) {
	for _, item := range ch.Output {
		if item.Type == "reasoning" {
			return item.Content, nil
		}
	}
	return "", fmt.Errorf("no reason found: %v", ch)
}
