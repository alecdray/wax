-- name: UpsertAlbumNote :one
INSERT INTO album_notes (id, user_id, album_id, content, updated_at)
VALUES (?, ?, ?, ?, current_timestamp)
ON CONFLICT(user_id, album_id) DO UPDATE SET
    content = excluded.content,
    updated_at = current_timestamp
RETURNING *;

-- name: GetAlbumNote :one
SELECT * FROM album_notes WHERE user_id = ? AND album_id = ?;

-- name: GetAlbumNotesByAlbumIds :many
SELECT * FROM album_notes
WHERE user_id = ? AND album_id IN (sqlc.slice('album_ids'));
