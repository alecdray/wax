-- name: InsertAlbumRatingState :one
INSERT INTO album_rating_state (id, user_id, album_id, state, snooze_count, next_rerate_at, created_at, updated_at)
VALUES (?, ?, ?, ?, 0, ?, current_timestamp, current_timestamp)
RETURNING *;

-- name: UpdateAlbumRatingState :one
UPDATE album_rating_state
SET state = ?, snooze_count = ?, next_rerate_at = ?, updated_at = current_timestamp
WHERE user_id = ? AND album_id = ?
RETURNING *;

-- name: GetAlbumRatingState :one
SELECT * FROM album_rating_state
WHERE user_id = ? AND album_id = ?;

-- name: GetAllAlbumRatingStates :many
SELECT * FROM album_rating_state
WHERE user_id = ?;

-- name: GetRerateQueueAlbums :many
SELECT
    albums.id,
    albums.spotify_id,
    albums.title,
    albums.image_url,
    COALESCE((
        SELECT GROUP_CONCAT(ar.name, ', ')
        FROM (SELECT DISTINCT ar2.id, ar2.name FROM album_artists aa JOIN artists ar2 ON ar2.id = aa.artist_id WHERE aa.album_id = albums.id) AS ar
    ), '') AS artist_names,
    ars.state,
    arl.rating
FROM album_rating_state ars
JOIN albums ON albums.id = ars.album_id
LEFT JOIN (
    SELECT arl2.album_id, arl2.rating
    FROM album_rating_log arl2
    JOIN (
        SELECT arl3.album_id, MAX(arl3.created_at) AS max_created_at
        FROM album_rating_log arl3
        WHERE arl3.user_id = ?
        GROUP BY arl3.album_id
    ) latest ON arl2.album_id = latest.album_id AND arl2.created_at = latest.max_created_at
    WHERE arl2.user_id = ?
) arl ON arl.album_id = ars.album_id
WHERE ars.user_id = ?
  AND (
    (ars.state = 'provisional' AND ars.next_rerate_at <= current_timestamp)
    OR ars.state = 'stalled'
  )
ORDER BY CASE ars.state WHEN 'stalled' THEN 0 WHEN 'provisional' THEN 1 ELSE 2 END,
         ars.next_rerate_at ASC;
