-- +goose Up
-- +goose StatementBegin
CREATE TABLE album_rating_state (
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
ALTER TABLE album_rating_log ADD COLUMN state TEXT CHECK(state IN ('provisional', 'finalized', 'stalled'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE album_rating_state;
-- +goose StatementEnd

-- +goose StatementBegin
-- SQLite does not support DROP COLUMN; album_rating_log.state cannot be removed on rollback.
SELECT 1;
-- +goose StatementEnd
