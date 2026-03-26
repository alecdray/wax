-- name: UpsertUserRelease :one
INSERT INTO user_releases (id, user_id, release_id, added_at) VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, release_id)
DO UPDATE SET
    added_at = COALESCE(EXCLUDED.added_at, added_at),
    removed_at = CASE
        WHEN removed_at IS NOT NULL AND EXCLUDED.added_at > removed_at THEN NULL
        ELSE removed_at
    END
RETURNING *;

-- name: GetUserReleases :many
SELECT sqlc.embed(user_releases), sqlc.embed(releases) FROM user_releases
JOIN releases ON user_releases.release_id = releases.id
WHERE user_id = ? AND removed_at IS NULL;

-- name: GetUserReleasesByAlbumId :many
SELECT sqlc.embed(user_releases), sqlc.embed(releases) FROM user_releases
JOIN releases ON user_releases.release_id = releases.id
WHERE user_id = ?
AND album_id = ?
AND removed_at IS NULL;

-- name: SoftDeleteUserReleasesByAlbumId :exec
UPDATE user_releases
SET removed_at = current_timestamp
WHERE user_id = ? AND release_id IN (
    SELECT id FROM releases WHERE album_id = ?
);
