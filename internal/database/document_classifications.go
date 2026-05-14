package database

import (
	"context"

	"github.com/dovbysh/duf.git/internal/ai/is_document"
)

func (p *PostgresClient) UpsertDocumentClassification(ctx context.Context, fileID int64, doc is_document.DocumentClassification) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO duf.document_classifications AS existing
		(file_id, document_status, explanation_document, text_present, explanation_text, summary)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (file_id) DO UPDATE SET
			document_status = EXCLUDED.document_status,
			explanation_document = EXCLUDED.explanation_document,
			text_present = EXCLUDED.text_present,
			explanation_text = EXCLUDED.explanation_text,
			summary = EXCLUDED.summary,
			updated_at = now()`,
		fileID,
		doc.DocumentStatus,
		doc.ExplanationDocument,
		doc.TextPresent,
		doc.ExplanationText,
		doc.Summary,
	)
	return err
}
