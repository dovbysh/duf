package medical_lab_report

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed 01.md
var Prompt01 []byte

// --- 1. Главная структура отчета ---
type LabReport struct {
	Metadata     *Metadata    `json:"metadata"`      // Указатель, так как объект может быть null или отсутствует
	ResultsTable []TestResult `json:"results_table"` // Слайс (массив) структур результатов
	Notes        string       `json:"notes"`
	Footer       *Footer      `json:"footer"` // Указатель для элемента футера
}

// --- 2. Структура метаданных пациента/анализа ---
type Metadata struct {
	PatientID              *string `json:"patient_id"` // Использование *string позволяет корректно обрабатывать JSON null
	AnalysisType           string  `json:"analysis_type"`
	MaterialCollectionDate *string `json:"material_collection_date"`
	LaboratoryStaff        *string `json:"laboratory_staff"`
}

// --- 3. Структура одного результата теста (ячейка таблицы) ---
type RR struct {
	Lln float64 `json:"lln"`
	Uln float64 `json:"uln"`
}
type TestResult struct {
	SrcTestName             string   `json:"src_test_name"`
	TestName                string   `json:"test_name"`
	ActualValue             *float64 `json:"actual_value,omitempty"` // Числовое значение
	ActualValueNonNumeric   *string  `json:"actual_value_non_numeric,omitempty"`
	Unit                    string   `json:"unit"`
	ReferenceRange          *RR      `json:"reference_range"`
	DeviationFlag           string   `json:"deviation_flag,omitempty"` // 'N', 'H' или 'L'
	DeviationFlagNonNumeric *bool    `json:"deviation_flag_non_numeric"`
	DeviationText           string   `json:"deviation_text"`
	SrcDeviationText        string   `json:"src_deviation_text"`
	IsNumerical             bool     `json:"is_numerical"`
}

// --- 4. Структура подвала (Footer) ---
type Footer struct {
	PrintDate string `json:"print_date"`
	PageCount string `json:"page_count"`
}

func GetLabReport(s string) (*LabReport, error) {
	var doc LabReport
	err := json.Unmarshal(
		[]byte(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(s, "```json"), "```"))),
		&doc,
	)
	return &doc, err
}
