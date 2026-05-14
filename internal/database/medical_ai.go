package database

import (
	"context"
	"encoding/json"

	"github.com/dovbysh/duf.git/internal/ai/is_medical"
	"github.com/dovbysh/duf.git/internal/ai/medical_lab_report"
)

func (p *PostgresClient) UpsertMedicalDocumentClassification(ctx context.Context, fileID int64, doc is_medical.DocumentClassification) error {
	var documentType any
	if doc.DocumentType != nil {
		documentType = string(*doc.DocumentType)
	}

	_, err := p.db.ExecContext(ctx, `INSERT INTO duf.medical_document_classifications AS existing
		(file_id, is_medical_document, document_type, explanation)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (file_id) DO UPDATE SET
			is_medical_document = EXCLUDED.is_medical_document,
			document_type = EXCLUDED.document_type,
			explanation = EXCLUDED.explanation,
			updated_at = now()`,
		fileID,
		doc.IsMedicalDocument,
		documentType,
		doc.Explanation,
	)
	return err
}

func (p *PostgresClient) UpsertMedicalLabReport(ctx context.Context, fileID int64, report medical_lab_report.LabReport) error {
	metadata, err := json.Marshal(report.Metadata)
	if err != nil {
		return err
	}
	resultsTable, err := json.Marshal(report.ResultsTable)
	if err != nil {
		return err
	}
	footer, err := json.Marshal(report.Footer)
	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, `INSERT INTO duf.medical_lab_reports AS existing
		(file_id, metadata, results_table, notes, footer)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (file_id) DO UPDATE SET
			metadata = EXCLUDED.metadata,
			results_table = EXCLUDED.results_table,
			notes = EXCLUDED.notes,
			footer = EXCLUDED.footer,
			updated_at = now()`,
		fileID,
		metadata,
		resultsTable,
		report.Notes,
		footer,
	)
	return err
}
