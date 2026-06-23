-- name: CreateFeed :one
insert into feeds (user_id, kind)
values (?, ?)
returning *;

-- name: SetFeedSourceRef :exec
-- Sets (or, with NULL, clears) a feed's external source handle.
UPDATE feeds SET source_ref = ? WHERE id = ?;

-- name: DeleteFeed :exec
DELETE FROM feeds WHERE id = ?;

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
    last_sync_status = COALESCE(?, last_sync_status),
    next_sync_at = ?,
    consecutive_failures = ?
WHERE id = ?
RETURNING *;

-- name: GetFeedsByUserId :many
select * from feeds where user_id = ?;

-- name: GetFeedByID :one
select * from feeds where id = ? and user_id = ?;

-- name: GetDueFeedsBatch :many
-- Feeds of a kind that are due to sync: next_sync_at has passed, or is NULL
-- (never scheduled, so sync immediately). A just-failed feed has next_sync_at
-- pushed into the future by its backoff, so it is not re-picked until then.
-- Soonest-due first; NULL (never scheduled) sorts ahead of the rest.
SELECT * FROM feeds
WHERE kind = ?
AND (next_sync_at IS NULL OR next_sync_at <= datetime('now'))
ORDER BY next_sync_at ASC
LIMIT 10;

-- name: GetSyncableRadarFeeds :many
-- Radar inbox feeds due to sync: those with a playlist handle whose next_sync_at
-- has passed (or is NULL). Same due-ness rule as GetDueFeedsBatch, so a failed
-- radar feed backs off instead of being re-read every tick.
SELECT * FROM feeds
WHERE kind = 'spotify_radar' AND source_ref IS NOT NULL
AND (next_sync_at IS NULL OR next_sync_at <= datetime('now'))
ORDER BY next_sync_at ASC
LIMIT 10;
