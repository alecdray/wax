-- name: CreateFeed :one
insert into feeds (user_id, kind)
values (?, ?)
returning *;

-- name: UpsertFeed :one
INSERT INTO feeds (id, user_id, kind)
VALUES (?, ?, ?)
ON CONFLICT (user_id, kind) DO UPDATE SET
    user_id = excluded.user_id,
    kind = excluded.kind
RETURNING *;

-- name: UpdateFeed :one
UPDATE feeds
SET last_sync_completed_at = COALESCE(?, last_sync_completed_at),
    last_sync_started_at = COALESCE(?, last_sync_started_at),
    last_sync_status = COALESCE(?, last_sync_status)
WHERE id = ?
RETURNING *;

-- name: GetFeedsByUserId :many
select * from feeds where user_id = ?;

-- name: GetFeedByID :one
select * from feeds where id = ? and user_id = ?;

-- name: GetStaleFeedsBatch :many
SELECT * FROM feeds
WHERE last_sync_completed_at IS NOT NULL
AND last_sync_completed_at < datetime('now', ?)
AND kind = ?
ORDER BY last_sync_completed_at ASC
LIMIT 10;
