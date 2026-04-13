-- +goose Up
-- +goose StatementBegin
INSERT INTO album_rating_state (id, user_id, album_id, state, snooze_count, next_rerate_at, created_at, updated_at)
SELECT
    lower(hex(randomblob(16))),
    r.user_id,
    r.album_id,
    'provisional',
    0,
    datetime(max(r.created_at), '+30 days'),
    current_timestamp,
    current_timestamp
FROM album_rating_log r
LEFT JOIN album_rating_state s ON s.user_id = r.user_id AND s.album_id = r.album_id
WHERE s.id IS NULL
GROUP BY r.user_id, r.album_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Removes the backfilled rows. Any state rows created by the application
-- (not this migration) will also be removed for albums that had no state
-- before this migration ran — acceptable for a rollback scenario.
DELETE FROM album_rating_state
WHERE (user_id, album_id) IN (
    SELECT r.user_id, r.album_id
    FROM album_rating_log r
    GROUP BY r.user_id, r.album_id
);
-- +goose StatementEnd
