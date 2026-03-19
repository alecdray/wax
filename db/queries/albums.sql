-- name: CreateAlbum :exec
INSERT INTO albums (id, spotify_id, title, image_url) VALUES (?, ?, ?, ?);

-- name: GetOrCreateAlbum :one
INSERT INTO albums (id, spotify_id, title, image_url) VALUES (?, ?, ?, ?)
ON CONFLICT (spotify_id)
DO UPDATE SET image_url = excluded.image_url
RETURNING *;

-- name: GetAlbum :one
SELECT * FROM albums WHERE id = ?;

-- name: GetAlbumsByIDs :many
SELECT * FROM albums WHERE id IN (sqlc.slice('ids'));

-- name: GetAlbumBySpotifyId :one
SELECT * FROM albums WHERE spotify_id = ?;

-- name: ListAlbums :many
SELECT * FROM albums WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT ?;
