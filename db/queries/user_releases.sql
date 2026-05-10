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

-- name: HasAnyUserReleaseForAlbum :one
-- Reports whether the user has any user_release row for the album, regardless
-- of status. Used to gate radar adds: radar is strictly pre-decision, so any
-- existing decision (owned, wishlist, removed) disqualifies the album.
SELECT EXISTS (
    SELECT 1 FROM user_releases ur
    JOIN releases r ON r.id = ur.release_id
    WHERE ur.user_id = ? AND r.album_id = ?
) AS has_release;

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

-- name: GetUserAlbumStateBySpotifyIds :many
-- For each input Spotify ID, returns the album's wax ID plus the user's state
-- for that album: 'in_library', 'removed', or 'on_radar'. An album may appear in
-- more than one branch when invariants drift (e.g., a stray radar row alongside
-- a wishlist or removed user_release); the service collapses to one state per
-- album with precedence in_library > on_radar > removed.
--
-- The Spotify-ID filter is expressed once via a CTE so sqlc's slice expansion
-- happens at exactly one site (the generator only replaces the first
-- /*SLICE*/ marker per call).
WITH target_albums AS (
    SELECT id, spotify_id
    FROM albums
    WHERE spotify_id IN (sqlc.slice('spotify_ids'))
)
SELECT ta.id        AS album_id,
       ta.spotify_id AS spotify_id,
       'in_library' AS state
FROM target_albums ta
JOIN user_releases ON user_releases.release_id IN (
    SELECT id FROM releases WHERE album_id = ta.id
)
WHERE user_releases.user_id = ?
  AND user_releases.status = 'owned'

UNION ALL

SELECT ta.id        AS album_id,
       ta.spotify_id AS spotify_id,
       'removed'    AS state
FROM target_albums ta
JOIN user_releases ON user_releases.release_id IN (
    SELECT id FROM releases WHERE album_id = ta.id
)
WHERE user_releases.user_id = ?
  AND user_releases.status = 'removed'
  AND NOT EXISTS (
      SELECT 1 FROM user_releases ur2
      JOIN releases r2 ON r2.id = ur2.release_id
      WHERE ur2.user_id = user_releases.user_id
        AND r2.album_id = ta.id
        AND ur2.status = 'owned'
  )

UNION ALL

SELECT ta.id        AS album_id,
       ta.spotify_id AS spotify_id,
       'on_radar'   AS state
FROM target_albums ta
JOIN user_album_radar ON user_album_radar.album_id = ta.id
WHERE user_album_radar.user_id = ?;
