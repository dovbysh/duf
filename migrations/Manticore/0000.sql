DROP TABLE IF EXISTS files;

CREATE TABLE files(
                      path text,
                      name text,
                      size bigint,
                      mtime bigint,
                      ctime bigint,
                      sha256 string,
                      is_deleted int
) engine='rt' ;
