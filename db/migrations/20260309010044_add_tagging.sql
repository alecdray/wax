-- +goose Up
-- +goose StatementBegin
CREATE TABLE tag_groups (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE tags (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    group_id   TEXT REFERENCES tag_groups(id) ON DELETE SET NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE album_tags (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    tag_id     TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, tag_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE album_tags;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE tags;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE tag_groups;
-- +goose StatementEnd
