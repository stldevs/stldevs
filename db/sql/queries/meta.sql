-- name: LastRun :one
SELECT created_at
FROM agg_meta
ORDER BY created_at DESC
LIMIT 1;

-- name: InsertRunLog :exec
INSERT INTO agg_meta (created_at)
VALUES ($1);
