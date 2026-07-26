-- +goose Up
CREATE TABLE crate_albums (
    id         TEXT PRIMARY KEY,
    crate_id   TEXT NOT NULL REFERENCES crates(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    added_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(crate_id, album_id)
);

-- +goose Down
DROP TABLE crate_albums;
