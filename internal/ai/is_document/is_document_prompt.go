package is_document

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed 01.md
var Prompt01 []byte

type DocumentStatus string

const (
	DocumentStatusDocument    DocumentStatus = "Документ"
	DocumentStatusNotDocument DocumentStatus = "Не документ"
)

type DocumentClassification struct {
	DocumentStatus      DocumentStatus `json:"document_status"`
	ExplanationDocument string         `json:"explanation_document"`
	TextPresent         bool           `json:"text_present"`
	ExplanationText     string         `json:"explanation_text"`
	Summary             string         `json:"summary"`
}

func GetDocumentClassification(s string) (*DocumentClassification, error) {
	var doc DocumentClassification
	err := json.Unmarshal(
		[]byte(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(s, "```json"), "```"))),
		&doc,
	)
	return &doc, err
}
