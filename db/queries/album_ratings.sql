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
