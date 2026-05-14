DROP INDEX IF EXISTS duf.image_analyses_search_vector_idx;

ALTER TABLE duf.image_analyses
    DROP COLUMN IF EXISTS search_vector,
    ADD COLUMN IF NOT EXISTS title text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS extracted_text text NOT NULL DEFAULT '';

ALTER TABLE duf.image_analyses
    ADD COLUMN search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(extracted_text, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(analysis_text, '')), 'B')
    ) STORED;

CREATE INDEX IF NOT EXISTS image_analyses_search_vector_idx
    ON duf.image_analyses
    USING GIN (search_vector);
