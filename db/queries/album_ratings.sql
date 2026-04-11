-- name: InsertAlbumRatingLogEntry :one
INSERT INTO album_rating_log (id, user_id, album_id, rating, note, state, created_at)
VALUES (?, ?, ?, ?, ?, ?, current_timestamp)
RETURNING *;

-- name: DeleteAlbumRatingLogEntry :exec
DELETE FROM album_rating_log
WHERE id = ? AND user_id = ?;

-- name: GetLatestUserAlbumRating :one
SELECT * FROM album_rating_log
WHERE user_id = ? AND album_id = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: GetLatestUserAlbumRatings :many
SELECT arl.* FROM album_rating_log arl
JOIN (
    SELECT arl2.album_id, MAX(arl2.created_at) AS max_created_at
    FROM album_rating_log arl2
    WHERE arl2.user_id = ?
    GROUP BY arl2.album_id
) latest ON arl.album_id = latest.album_id AND arl.created_at = latest.max_created_at
WHERE arl.user_id = ?;

-- name: GetUserAlbumRatingLog :many
SELECT * FROM album_rating_log
WHERE user_id = ? AND album_id = ?
ORDER BY created_at DESC;

-- name: GetUnratedAlbums :many
SELECT albums.*,
    COALESCE((
        SELECT GROUP_CONCAT(a.name, ', ')
        FROM (SELECT DISTINCT ar.id, ar.name FROM album_artists aa JOIN artists ar ON ar.id = aa.artist_id WHERE aa.album_id = albums.id) AS a
    ), '') as artist_names
FROM user_releases
JOIN releases ON releases.id = user_releases.release_id
JOIN albums ON albums.id = releases.album_id
LEFT JOIN (
    SELECT arl.album_id, arl.user_id
    FROM album_rating_log arl
    JOIN (
        SELECT arl2.album_id, MAX(arl2.created_at) AS max_created_at
        FROM album_rating_log arl2
        WHERE arl2.user_id = ?
        GROUP BY arl2.album_id
    ) latest ON arl.album_id = latest.album_id AND arl.created_at = latest.max_created_at
    WHERE arl.user_id = ?
) latest_rating ON latest_rating.album_id = albums.id AND latest_rating.user_id = user_releases.user_id
LEFT JOIN track_plays ON track_plays.album_id = albums.id AND track_plays.user_id = user_releases.user_id
WHERE user_releases.user_id = ?
  AND latest_rating.album_id IS NULL
  AND user_releases.removed_at IS NULL
GROUP BY albums.id
ORDER BY MAX(track_plays.played_at) DESC NULLS LAST, MAX(user_releases.added_at) DESC
LIMIT 20;
