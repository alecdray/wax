-- name: AddAlbumToRadar :one
-- Idempotent: if a radar row exists, it's returned unchanged.
INSERT INTO user_album_radar (id, user_id, album_id)
VALUES (?, ?, ?)
ON CONFLICT (user_id, album_id) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING *;

-- name: RemoveAlbumFromRadar :exec
DELETE FROM user_album_radar
WHERE user_id = ? AND album_id = ?;

-- name: GetRadarAlbums :many
-- Returns radar rows whose album the user does not currently own or wishlist.
-- A `removed` release does not disqualify the album, so a discarded album can
-- sit on the radar (ADR 0005); owning or wishlisting it filters it out here.
SELECT sqlc.embed(user_album_radar), sqlc.embed(albums)
FROM user_album_radar
JOIN albums ON user_album_radar.album_id = albums.id
WHERE user_album_radar.user_id = ?
  AND NOT EXISTS (
      SELECT 1 FROM user_releases ur
      JOIN releases r ON r.id = ur.release_id
      WHERE ur.user_id = user_album_radar.user_id
        AND r.album_id = user_album_radar.album_id
        AND ur.status IN ('owned', 'wishlist')
  )
ORDER BY user_album_radar.created_at ASC;

-- name: IsAlbumOnRadar :one
SELECT EXISTS (
    SELECT 1 FROM user_album_radar
    WHERE user_id = ? AND album_id = ?
) AS on_radar;
