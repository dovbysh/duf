CREATE TABLE files(
                      path text,
                      name text,
                      size bigint,
                      mtime bigint,
                      ctime bigint,
                      sha256 string,
                      is_deleted int
) engine='rt' attr_uint='size', attr_uint='mtime', attr_uint='ctime', attr_uint='is_deleted';
