-- name: UpsertAlbumGenre :exec
INSERT INTO album_genres (id, album_id, genre_id, genre_label) VALUES (?, ?, ?, ?)
ON CONFLICT (album_id, genre_id) DO UPDATE SET genre_label = excluded.genre_label;

-- name: DeleteAlbumGenresByAlbumId :exec
DELETE FROM album_genres WHERE album_id = ?;

-- name: GetAlbumGenresByAlbumIds :many
SELECT album_id, genre_id, genre_label FROM album_genres
WHERE album_id IN (sqlc.slice('album_ids'))
ORDER BY genre_label;

-- name: MarkAlbumGenreEnriched :exec
INSERT INTO album_genre_enrichment (album_id) VALUES (?)
ON CONFLICT (album_id) DO UPDATE SET enriched_at = CURRENT_TIMESTAMP;

-- name: GetEnrichedAlbumIds :many
SELECT album_id FROM album_genre_enrichment
WHERE album_id IN (sqlc.slice('album_ids'));
