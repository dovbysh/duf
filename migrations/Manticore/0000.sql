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

drop table if exists tc;
create table tc(
    description text
) morphology='lemmatize_ru';
