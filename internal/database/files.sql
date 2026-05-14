-- ### files without sha256
SELECT id, path
FROM duf.files
WHERE is_deleted = 0 AND sha256 = ''
ORDER BY id
LIMIT 100;

-- ### update hash
UPDATE duf.files
SET sha256 = 'example'
WHERE id = 0;

--
SELECT
    count(*) AS t,
    count(*) FILTER (WHERE sha256 = '' OR sha256 IS NULL) AS ns
FROM duf.files;
