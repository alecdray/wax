-- name: GetOrCreateAlbumTrack :one
INSERT INTO album_tracks (album_id, track_id) VALUES (?, ?)
ON CONFLICT (album_id, track_id)
DO UPDATE SET album_id = album_id
RETURNING *;

-- name: GetAlbumTracksByAlbumId :many
SELECT album_tracks.album_id, sqlc.embed(tracks) FROM album_tracks
JOIN tracks ON album_tracks.track_id = tracks.id
WHERE album_id = ?
ORDER BY tracks.disc_number ASC, tracks.track_number ASC;

-- name: GetAlbumTracksByAlbumIds :many
SELECT album_tracks.album_id, sqlc.embed(tracks) FROM album_tracks
JOIN tracks ON album_tracks.track_id = tracks.id
WHERE album_id IN (sqlc.slice('album_ids'))
ORDER BY tracks.disc_number ASC, tracks.track_number ASC;
