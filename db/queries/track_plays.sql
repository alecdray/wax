-- name: UpsertTrackPlay :exec
INSERT INTO track_plays (id, user_id, track_id, album_id, played_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (user_id, track_id, played_at) DO NOTHING;

-- name: GetLastPlayedAtByAlbumIds :many
SELECT album_id, MAX(played_at) as last_played_at
FROM track_plays
WHERE user_id = ? AND album_id IN (sqlc.slice('album_ids'))
GROUP BY album_id;

-- name: GetRecentlyPlayedAlbums :many
SELECT albums.*, MAX(track_plays.played_at) as last_played_at,
    COALESCE((
        SELECT GROUP_CONCAT(a.name, ', ')
        FROM (SELECT DISTINCT ar.id, ar.name FROM album_artists aa JOIN artists ar ON ar.id = aa.artist_id WHERE aa.album_id = albums.id) AS a
    ), '') as artist_names
FROM track_plays
JOIN albums ON albums.id = track_plays.album_id
WHERE track_plays.user_id = ?
GROUP BY albums.id
ORDER BY last_played_at DESC
LIMIT 20;

