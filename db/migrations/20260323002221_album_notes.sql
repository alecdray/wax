-- +goose Up
CREATE TABLE IF NOT EXISTS album_notes (
    id         TEXT NOT NULL PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    content    TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL DEFAULT current_timestamp,
    UNIQUE(user_id, album_id)
);

-- +goose Down
DROP TABLE IF EXISTS album_notes;
