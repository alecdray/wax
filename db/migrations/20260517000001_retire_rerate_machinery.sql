-- +goose Up
-- +goose StatementBegin
-- Rebuild album_rating_state without snooze_count / next_rerate_at and with a
-- two-value CHECK constraint. SQLite cannot DROP COLUMN with a CHECK that
-- references the dropped columns, so we rebuild the table.
--
-- Idempotent: goose records the applied version, so the rebuild only runs on
-- the first apply. The INSERT pulls every existing row and maps any historical
-- 'stalled' state to 'provisional' in the same step — the previous CHECK
-- constraint on the live state column prevents a separate UPDATE pass.
CREATE TABLE __new_album_rating_state (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    state      TEXT NOT NULL CHECK(state IN ('provisional', 'finalized')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO __new_album_rating_state (id, user_id, album_id, state, created_at, updated_at)
SELECT
    id,
    user_id,
    album_id,
    CASE WHEN state = 'stalled' THEN 'provisional' ELSE state END,
    created_at,
    updated_at
FROM album_rating_state;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE album_rating_state;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE __new_album_rating_state RENAME TO album_rating_state;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Recreate the original table shape; rows previously 'stalled' cannot be
-- distinguished from genuine provisional rows and remain 'provisional'.
CREATE TABLE __old_album_rating_state (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id       TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    state          TEXT NOT NULL CHECK(state IN ('provisional', 'finalized', 'stalled')),
    snooze_count   INTEGER NOT NULL DEFAULT 0,
    next_rerate_at DATETIME CHECK(
        (state = 'stalled' AND next_rerate_at IS NULL)
        OR
        (state != 'stalled' AND next_rerate_at IS NOT NULL)
    ),
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO __old_album_rating_state (id, user_id, album_id, state, snooze_count, next_rerate_at, created_at, updated_at)
SELECT id, user_id, album_id, state, 0, datetime(created_at, '+30 days'), created_at, updated_at FROM album_rating_state;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE album_rating_state;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE __old_album_rating_state RENAME TO album_rating_state;
-- +goose StatementEnd
