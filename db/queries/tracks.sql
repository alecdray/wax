-- name: CreateTrack :exec
INSERT INTO tracks (id, spotify_id, title, disc_number, track_number) VALUES (?, ?, ?, ?, ?);

-- name: GetOrCreateTrack :one
INSERT INTO tracks (id, spotify_id, title, disc_number, track_number) VALUES (?, ?, ?, ?, ?)
ON CONFLICT (spotify_id)
DO UPDATE SET disc_number = excluded.disc_number, track_number = excluded.track_number
RETURNING *;

-- name: GetTrack :one
SELECT * FROM tracks WHERE id = ?;

-- name: GetTrackBySpotifyId :one
SELECT * FROM tracks WHERE spotify_id = ?;
