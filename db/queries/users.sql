-- name: CreateUser :one
INSERT INTO users (id, spotify_id) VALUES (?, ?)
RETURNING *;

-- name: UpsertSpotifyUser :one
INSERT INTO users (id, spotify_id, spotify_refresh_token) VALUES (?, ?, ?)
ON CONFLICT (spotify_id)
DO UPDATE SET spotify_id = EXCLUDED.spotify_id, spotify_refresh_token = coalesce(EXCLUDED.spotify_refresh_token, spotify_refresh_token)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserBySpotifyId :one
SELECT * FROM users WHERE spotify_id = ?;

-- name: GetUsersWithSpotifyToken :many
SELECT * FROM users
WHERE spotify_refresh_token IS NOT NULL AND deleted_at IS NULL;
