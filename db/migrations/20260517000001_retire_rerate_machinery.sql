-- +goose Up
-- +goose StatementBegin
-- Retire the time-based rerate lifecycle on album_rating_state:
--   * drop snooze_count and next_rerate_at
--   * narrow the state CHECK from {provisional, finalized, stalled} to
--     {provisional, finalized}
--   * backfill any pre-existing state='stalled' rows to 'provisional' so they
--     satisfy the new CHECK
--
-- SQLite cannot DROP COLUMN with a CHECK that references the dropped columns,
-- and the original CHECK on the state column blocks a plain UPDATE to map
-- 'stalled' to 'provisional', so we rebuild the table and map state inline
-- inside the INSERT.
--
-- Idempotent via goose's applied-version tracking — the rebuild runs only on
-- the first apply.
--
-- This migration does not touch album_rating_log: log entries are immutable
-- history and the log's state CHECK still admits 'stalled' for entries written
-- under the earlier lifecycle. No new album_rating_log rows are written by
-- this migration.
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
