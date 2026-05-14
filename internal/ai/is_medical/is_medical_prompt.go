package is_medical

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed 01.md
var Prompt01 []byte

// DocumentType представляет собой перечисление возможных типов медицинских документов.
// Использование типа позволяет избежать ошибок при присвоении строки.
type DocumentType string

const (
	// Типы, соответствующие требованиям промпта и JSON-схеме
	DocProtocol           DocumentType = "Doctor's Appointment Protocol"
	DocLabAnalysisResults DocumentType = "Lab Analysis Results"
	DocOtherMedical       DocumentType = "Other Medical Document"
)

// ClassificationStatus используется для обозначения случая, когда документ НЕ медицинский.
const NonMedicalType DocumentType = "" // Пустая строка или null для JSON

// DocumentClassification - Основная структура, которая соответствует требуемому JSON-формату.
type DocumentClassification struct {
	IsMedicalDocument bool          `json:"is_medical_document"`
	DocumentType      *DocumentType `json:"document_type"` // Использование указателя позволяет корректно вернуть null в JSON
	Explanation       string        `json:"explanation"`
}

func GetDocumentClassification(s string) (*DocumentClassification, error) {
	var doc DocumentClassification
	err := json.Unmarshal(
		[]byte(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(s, "```json"), "```"))),
		&doc,
	)
	return &doc, err
}
