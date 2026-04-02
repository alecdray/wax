-- +goose Up
-- +goose StatementBegin
CREATE TABLE album_genres (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id     TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    genre_id     TEXT NOT NULL,
    genre_label  TEXT NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, genre_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE album_moods (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    mood       TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, mood)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE album_user_tags (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, tag)
);
-- +goose StatementEnd

-- Migrate Mood-group tags to album_moods
-- +goose StatementBegin
INSERT INTO album_moods (id, user_id, album_id, mood, created_at)
SELECT at.id, at.user_id, at.album_id, t.name, at.created_at
FROM album_tags at
JOIN tags t ON at.tag_id = t.id
JOIN tag_groups tg ON t.group_id = tg.id
WHERE tg.name = 'Mood'
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- Migrate ungrouped tags to album_user_tags
-- +goose StatementBegin
INSERT INTO album_user_tags (id, user_id, album_id, tag, created_at)
SELECT at.id, at.user_id, at.album_id, t.name, at.created_at
FROM album_tags at
JOIN tags t ON at.tag_id = t.id
WHERE t.group_id IS NULL
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE album_genres;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE album_moods;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE album_user_tags;
-- +goose StatementEnd
