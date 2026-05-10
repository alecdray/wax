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
-- Returns radar rows whose album has no user_releases entries (not on wishlist, not owned, not removed).
-- This enforces the "any release activity wipes radar" invariant at query time.
SELECT sqlc.embed(user_album_radar), sqlc.embed(albums)
FROM user_album_radar
JOIN albums ON user_album_radar.album_id = albums.id
WHERE user_album_radar.user_id = ?
  AND NOT EXISTS (
      SELECT 1 FROM user_releases ur
      JOIN releases r ON r.id = ur.release_id
      WHERE ur.user_id = user_album_radar.user_id
        AND r.album_id = user_album_radar.album_id
  );

-- name: IsAlbumOnRadar :one
SELECT EXISTS (
    SELECT 1 FROM user_album_radar
    WHERE user_id = ? AND album_id = ?
) AS on_radar;
