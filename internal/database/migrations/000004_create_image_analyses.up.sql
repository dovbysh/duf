CREATE TABLE IF NOT EXISTS duf.image_analyses (
    file_id bigint PRIMARY KEY REFERENCES duf.files(id) ON DELETE CASCADE,
    analysis_text text NOT NULL DEFAULT '',
    search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('simple', coalesce(analysis_text, ''))
    ) STORED,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS image_analyses_search_vector_idx
    ON duf.image_analyses
    USING GIN (search_vector);
