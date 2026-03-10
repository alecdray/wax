-- name: GetOrCreateTagGroup :one
INSERT INTO tag_groups (id, user_id, name) VALUES (?, ?, ?)
ON CONFLICT (user_id, name) DO UPDATE SET name = name
RETURNING *;

-- name: GetTagGroupsByUserId :many
SELECT * FROM tag_groups WHERE user_id = ?;

-- name: GetOrCreateTag :one
INSERT INTO tags (id, user_id, name, group_id) VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, name) DO UPDATE SET name = name
RETURNING *;

-- name: GetTagsByUserId :many
SELECT sqlc.embed(tags),
    COALESCE(tag_groups.id, '') as group_id_value,
    COALESCE(tag_groups.name, '') as group_name
FROM tags
LEFT JOIN tag_groups ON tags.group_id = tag_groups.id
WHERE tags.user_id = ?
ORDER BY tags.name;

-- name: GetAlbumTagsByAlbumId :many
SELECT album_tags.album_id, sqlc.embed(tags),
    COALESCE(tag_groups.id, '') as group_id_value,
    COALESCE(tag_groups.name, '') as group_name
FROM album_tags
JOIN tags ON album_tags.tag_id = tags.id
LEFT JOIN tag_groups ON tags.group_id = tag_groups.id
WHERE album_tags.user_id = ? AND album_tags.album_id = ?
ORDER BY tags.name;

-- name: GetAlbumTagsByAlbumIds :many
SELECT album_tags.album_id, sqlc.embed(tags),
    COALESCE(tag_groups.id, '') as group_id_value,
    COALESCE(tag_groups.name, '') as group_name
FROM album_tags
JOIN tags ON album_tags.tag_id = tags.id
LEFT JOIN tag_groups ON tags.group_id = tag_groups.id
WHERE album_tags.user_id = ? AND album_tags.album_id IN (sqlc.slice('album_ids'))
ORDER BY tags.name;

-- name: CreateAlbumTag :one
INSERT INTO album_tags (id, user_id, album_id, tag_id) VALUES (?, ?, ?, ?)
ON CONFLICT (user_id, album_id, tag_id) DO UPDATE SET tag_id = tag_id
RETURNING *;

-- name: DeleteAlbumTag :exec
DELETE FROM album_tags WHERE user_id = ? AND album_id = ? AND tag_id = ?;

-- name: DeleteAlbumTagsByAlbumId :exec
DELETE FROM album_tags WHERE user_id = ? AND album_id = ?;
