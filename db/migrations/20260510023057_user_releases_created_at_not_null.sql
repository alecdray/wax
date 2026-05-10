-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_releases_new (
    id                 TEXT PRIMARY KEY,
    user_id            TEXT NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    release_id         TEXT NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    status             TEXT NOT NULL DEFAULT 'owned'
                            CHECK (status IN ('wishlist', 'owned', 'removed')),
    created_at         DATETIME NOT NULL,
    status_updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at         DATETIME,
    UNIQUE(user_id, release_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO user_releases_new (id, user_id, release_id, status, created_at, status_updated_at, deleted_at)
SELECT id, user_id, release_id, status,
       COALESCE(created_at, status_updated_at, CURRENT_TIMESTAMP),
       status_updated_at,
       deleted_at
  FROM user_releases;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE user_releases;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_releases_new RENAME TO user_releases;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE user_releases_old (
    id                 TEXT PRIMARY KEY,
    user_id            TEXT NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    release_id         TEXT NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    status             TEXT NOT NULL DEFAULT 'owned'
                            CHECK (status IN ('wishlist', 'owned', 'removed')),
    created_at         DATETIME,
    status_updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at         DATETIME,
    UNIQUE(user_id, release_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO user_releases_old SELECT * FROM user_releases;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE user_releases;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_releases_old RENAME TO user_releases;
-- +goose StatementEnd
