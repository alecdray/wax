-- name: UpsertAlbumRating :one
INSERT INTO album_ratings (id, user_id, album_id, rating) VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, album_id)
DO UPDATE SET rating = COALESCE(EXCLUDED.rating, rating), review = COALESCE(review, review), updated_at = current_timestamp
RETURNING *;

-- name: UpsertAlbumReview :one
INSERT INTO album_ratings (id, user_id, album_id, review) VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, album_id)
DO UPDATE SET review = EXCLUDED.review, updated_at = current_timestamp
RETURNING *;

-- name: GetUserAlbumRatings :many
select * from album_ratings
where user_id = ?;

-- name: GetUserAlbumRating :one
select * from album_ratings
where user_id = ?
and album_id = ?;

-- name: GetUserAlbumRatingById :one
select * from album_ratings
where id = ?;

-- name: ClearAlbumRating :exec
UPDATE album_ratings SET rating = NULL, updated_at = current_timestamp
WHERE user_id = ? AND album_id = ?;

-- name: GetUnratedAlbums :many
SELECT albums.*,
    COALESCE((
        SELECT GROUP_CONCAT(a.name, ', ')
        FROM (SELECT DISTINCT ar.id, ar.name FROM album_artists aa JOIN artists ar ON ar.id = aa.artist_id WHERE aa.album_id = albums.id) AS a
    ), '') as artist_names
FROM user_releases
JOIN releases ON releases.id = user_releases.release_id
JOIN albums ON albums.id = releases.album_id
LEFT JOIN album_ratings ON album_ratings.album_id = albums.id AND album_ratings.user_id = user_releases.user_id
WHERE user_releases.user_id = ?
  AND (album_ratings.id IS NULL OR album_ratings.rating IS NULL)
GROUP BY albums.id
ORDER BY MAX(user_releases.added_at) DESC
LIMIT 20;
