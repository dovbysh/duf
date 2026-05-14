CREATE SCHEMA IF NOT EXISTS duf;

CREATE TABLE IF NOT EXISTS duf.files (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    path text NOT NULL UNIQUE,
    name text NOT NULL,
    size bigint NOT NULL,
    mtime bigint NOT NULL,
    ctime bigint NOT NULL,
    sha256 text NOT NULL DEFAULT '',
    is_deleted integer NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now()
);
