-- name: CreateRelease :exec
INSERT INTO releases (id, album_id, format) VALUES (?, ?, ?);

-- name: GetOrCreateRelease :one
INSERT INTO releases (id, album_id, format) VALUES (?, ?, ?)
ON CONFLICT (album_id, format)
DO UPDATE SET album_id = album_id
RETURNING *;

-- name: GetRelease :one
SELECT * FROM releases WHERE id = ?;

-- name: GetReleases :many
SELECT * FROM releases WHERE album_id = ?;

-- name: UpdateRelease :exec
UPDATE releases
SET
    album_id = COALESCE(?, album_id),
    format = COALESCE(?, format)
WHERE id = ?;

-- name: UpdateReleaseDiscogsInfo :exec
UPDATE releases
SET
    discogs_id = ?,
    label = ?,
    released_at = ?
WHERE id = ?;
