package database

import "context"

func (p *PostgresClient) UpsertImageAnalysis(ctx context.Context, fileID int64, analysisText, title, extractedText string) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO duf.image_analyses AS existing
		(file_id, analysis_text, title, extracted_text)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (file_id) DO UPDATE SET
			analysis_text = EXCLUDED.analysis_text,
			title = EXCLUDED.title,
			extracted_text = EXCLUDED.extracted_text,
			updated_at = now()`,
		fileID,
		analysisText,
		title,
		extractedText,
	)
	return err
}
