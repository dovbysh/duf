CREATE TABLE IF NOT EXISTS duf.document_classifications (
    file_id bigint PRIMARY KEY REFERENCES duf.files(id) ON DELETE CASCADE,
    document_status text NOT NULL CHECK (document_status IN ('Документ', 'Не документ')),
    explanation_document text NOT NULL DEFAULT '',
    text_present boolean NOT NULL DEFAULT false,
    explanation_text text NOT NULL DEFAULT '',
    summary text NOT NULL DEFAULT '',
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(document_status, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(explanation_document, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(explanation_text, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(summary, '')), 'C')
    ) STORED,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS document_classifications_search_vector_idx
    ON duf.document_classifications
    USING GIN (search_vector);
