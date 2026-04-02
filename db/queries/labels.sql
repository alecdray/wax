-- name: CreateAlbumGenre :one
INSERT INTO album_genres (id, user_id, album_id, genre_id, genre_label)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (user_id, album_id, genre_id) DO UPDATE SET genre_label = excluded.genre_label
RETURNING *;

-- name: DeleteAlbumGenresByAlbumId :exec
DELETE FROM album_genres WHERE user_id = ? AND album_id = ?;

-- name: GetAlbumGenresByAlbumId :many
SELECT * FROM album_genres WHERE user_id = ? AND album_id = ? ORDER BY genre_label;

-- name: GetAlbumGenresByAlbumIds :many
SELECT * FROM album_genres WHERE user_id = ? AND album_id IN (sqlc.slice('album_ids')) ORDER BY genre_label;

-- name: CreateAlbumMood :one
INSERT INTO album_moods (id, user_id, album_id, mood)
VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, album_id, mood) DO UPDATE SET mood = mood
RETURNING *;

-- name: DeleteAlbumMoodsByAlbumId :exec
DELETE FROM album_moods WHERE user_id = ? AND album_id = ?;

-- name: GetAlbumMoodsByAlbumId :many
SELECT * FROM album_moods WHERE user_id = ? AND album_id = ? ORDER BY mood;

-- name: GetAlbumMoodsByAlbumIds :many
SELECT * FROM album_moods WHERE user_id = ? AND album_id IN (sqlc.slice('album_ids')) ORDER BY mood;

-- name: GetDistinctUserMoods :many
SELECT DISTINCT mood FROM album_moods WHERE user_id = ? ORDER BY mood;

-- name: CreateAlbumUserTag :one
INSERT INTO album_user_tags (id, user_id, album_id, tag)
VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, album_id, tag) DO UPDATE SET tag = tag
RETURNING *;

-- name: DeleteAlbumUserTagsByAlbumId :exec
DELETE FROM album_user_tags WHERE user_id = ? AND album_id = ?;

-- name: GetAlbumUserTagsByAlbumId :many
SELECT * FROM album_user_tags WHERE user_id = ? AND album_id = ? ORDER BY tag;

-- name: GetAlbumUserTagsByAlbumIds :many
SELECT * FROM album_user_tags WHERE user_id = ? AND album_id IN (sqlc.slice('album_ids')) ORDER BY tag;

-- name: GetDistinctUserTags :many
SELECT DISTINCT tag FROM album_user_tags WHERE user_id = ? ORDER BY tag;
