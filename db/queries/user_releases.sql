-- name: UpsertUserRelease :one
INSERT INTO user_releases (id, user_id, release_id, status_updated_at) VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, release_id)
DO UPDATE SET
    status_updated_at = EXCLUDED.status_updated_at,
    status = 'owned'
RETURNING *;

-- name: GetUserReleases :many
SELECT sqlc.embed(user_releases), sqlc.embed(releases) FROM user_releases
JOIN releases ON user_releases.release_id = releases.id
WHERE user_id = ? AND status != 'removed';

-- name: GetUserReleasesByAlbumId :many
SELECT sqlc.embed(user_releases), sqlc.embed(releases) FROM user_releases
JOIN releases ON user_releases.release_id = releases.id
WHERE user_id = ?
AND album_id = ?
AND status != 'removed';

-- name: SoftDeleteUserReleasesByAlbumId :exec
UPDATE user_releases
SET status = 'removed', status_updated_at = current_timestamp
WHERE user_id = ? AND release_id IN (
    SELECT id FROM releases WHERE album_id = ?
);

-- name: SoftDeleteUserRelease :exec
UPDATE user_releases
SET status = 'removed', status_updated_at = current_timestamp
WHERE user_id = ? AND release_id = ? AND status != 'removed';
