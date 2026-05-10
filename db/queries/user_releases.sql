-- name: UpsertOwnedRelease :one
-- Upserts a user_release in the 'owned' state. On insert, created_at = sqlc.arg(created_at).
-- On conflict, status flips to 'owned' and status_updated_at is bumped; created_at is preserved.
INSERT INTO user_releases (id, user_id, release_id, status, created_at, status_updated_at)
VALUES (?, ?, ?, 'owned', ?, ?)
ON CONFLICT (user_id, release_id)
DO UPDATE SET
    status            = 'owned',
    status_updated_at = EXCLUDED.status_updated_at
RETURNING *;

-- name: UpsertWishlistRelease :one
-- Upserts a user_release in the 'wishlist' state. On insert, created_at = sqlc.arg(created_at).
-- On conflict, status flips to 'wishlist' and status_updated_at is bumped; created_at is preserved.
INSERT INTO user_releases (id, user_id, release_id, status, created_at, status_updated_at)
VALUES (?, ?, ?, 'wishlist', ?, ?)
ON CONFLICT (user_id, release_id)
DO UPDATE SET
    status            = 'wishlist',
    status_updated_at = EXCLUDED.status_updated_at
RETURNING *;

-- name: GetUserReleases :many
-- Returns the user's owned releases.
SELECT sqlc.embed(user_releases), sqlc.embed(releases)
FROM user_releases
JOIN releases ON user_releases.release_id = releases.id
WHERE user_id = ? AND status = 'owned';

-- name: GetUserReleasesByAlbumId :many
-- Returns the user's owned releases for one album.
SELECT sqlc.embed(user_releases), sqlc.embed(releases)
FROM user_releases
JOIN releases ON user_releases.release_id = releases.id
WHERE user_id = ?
  AND album_id = ?
  AND status = 'owned';

-- name: GetWishlistReleases :many
-- Returns the user's wishlist releases.
SELECT sqlc.embed(user_releases), sqlc.embed(releases)
FROM user_releases
JOIN releases ON user_releases.release_id = releases.id
WHERE user_id = ? AND status = 'wishlist';

-- name: MarkReleaseRemoved :exec
UPDATE user_releases
   SET status = 'removed', status_updated_at = current_timestamp
 WHERE user_id = ? AND release_id = ? AND status = 'owned';

-- name: MarkReleasesRemovedByAlbumId :exec
UPDATE user_releases
   SET status = 'removed', status_updated_at = current_timestamp
 WHERE user_id = ?
   AND status = 'owned'
   AND release_id IN (SELECT id FROM releases WHERE album_id = ?);

-- name: DeleteWishlistRelease :exec
DELETE FROM user_releases
 WHERE user_id = ? AND release_id = ? AND status = 'wishlist';
