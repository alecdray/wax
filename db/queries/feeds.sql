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

-- name: GetSyncableRadarFeeds :many
-- Radar inbox feeds eligible to sync: any with a playlist handle. Unlike
-- saved-album feeds there is no staleness window. The inbox is polled each cron
-- tick (skipping in-flight ones in the task) so added albums land promptly, and
-- never-synced feeds are picked up immediately. Least-recently-synced first.
SELECT * FROM feeds
WHERE kind = 'spotify_radar' AND source_ref IS NOT NULL
ORDER BY last_sync_completed_at ASC
LIMIT 10;
