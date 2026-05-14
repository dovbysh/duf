DROP INDEX IF EXISTS duf.image_analyses_search_vector_idx;

ALTER TABLE duf.image_analyses
    DROP COLUMN IF EXISTS search_vector,
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS extracted_text;

ALTER TABLE duf.image_analyses
    ADD COLUMN search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('simple', coalesce(analysis_text, ''))
    ) STORED;

CREATE INDEX IF NOT EXISTS image_analyses_search_vector_idx
    ON duf.image_analyses
    USING GIN (search_vector);
