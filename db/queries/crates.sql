-- name: ListCrates :many
SELECT c.id, c.name, COUNT(ca.id) AS album_count, c.created_at
FROM crates c
LEFT JOIN crate_albums ca ON ca.crate_id = c.id
WHERE c.user_id = ?
GROUP BY c.id
ORDER BY c.name COLLATE NOCASE ASC;

-- name: GetCrate :one
SELECT c.id, c.name, c.created_at
FROM crates c
WHERE c.id = ? AND c.user_id = ?;

-- name: GetCrateAlbumIDs :many
SELECT album_id FROM crate_albums WHERE crate_id = ? ORDER BY added_at DESC;

-- name: InsertCrate :one
INSERT INTO crates (id, user_id, name) VALUES (?, ?, ?) RETURNING id, user_id, name, created_at;

-- name: DeleteCrate :exec
DELETE FROM crates WHERE id = ? AND user_id = ?;

-- name: InsertCrateAlbum :exec
INSERT INTO crate_albums (id, crate_id, album_id) VALUES (?, ?, ?) ON CONFLICT(crate_id, album_id) DO NOTHING;

-- name: DeleteCrateAlbum :exec
DELETE FROM crate_albums WHERE crate_id = ? AND album_id = ?;
