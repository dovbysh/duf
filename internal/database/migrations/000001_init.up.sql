CREATE TABLE IF NOT EXISTS files (
    id numeric(20, 0) PRIMARY KEY,
    path text NOT NULL,
    name text NOT NULL,
    size bigint NOT NULL,
    mtime bigint NOT NULL,
    ctime bigint NOT NULL,
    sha256 text NOT NULL DEFAULT '',
    is_deleted integer NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS files_hash_queue_idx ON files (is_deleted, sha256);
