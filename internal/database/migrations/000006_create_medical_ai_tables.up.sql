CREATE TABLE IF NOT EXISTS duf.medical_document_classifications (
    file_id bigint PRIMARY KEY REFERENCES duf.files(id) ON DELETE CASCADE,
    is_medical_document boolean NOT NULL DEFAULT false,
    document_type text,
    explanation text NOT NULL DEFAULT '',
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(document_type, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(explanation, '')), 'B')
    ) STORED,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS medical_document_classifications_search_vector_idx
    ON duf.medical_document_classifications
    USING GIN (search_vector);

CREATE TABLE IF NOT EXISTS duf.medical_lab_reports (
    file_id bigint PRIMARY KEY REFERENCES duf.files(id) ON DELETE CASCADE,
    metadata jsonb,
    results_table jsonb NOT NULL DEFAULT '[]'::jsonb,
    notes text NOT NULL DEFAULT '',
    footer jsonb,
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(metadata::text, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(results_table::text, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(notes, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(footer::text, '')), 'C')
    ) STORED,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS medical_lab_reports_search_vector_idx
    ON duf.medical_lab_reports
    USING GIN (search_vector);
