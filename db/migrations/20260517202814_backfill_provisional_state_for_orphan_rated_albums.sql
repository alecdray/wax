-- Backfill: any (user, album) that has at least one album_rating_log entry but
-- no album_rating_state row gets a provisional state row. The post-rework
-- lifecycle holds the invariant "rated → has a state row" — the first save
-- implicitly creates the row as provisional. This migration repairs legacy
-- orphans from before that invariant was enforced. Idempotent: the same WHERE
-- clause filters out everything created by a prior run.

-- +goose Up
-- +goose StatementBegin
INSERT INTO album_rating_state (id, user_id, album_id, state, created_at, updated_at)
SELECT
    'backfill-' || arl.user_id || '-' || arl.album_id AS id,
    arl.user_id,
    arl.album_id,
    'provisional' AS state,
    MIN(arl.created_at) AS created_at,
    CURRENT_TIMESTAMP AS updated_at
FROM album_rating_log arl
LEFT JOIN album_rating_state ars
    ON ars.user_id = arl.user_id AND ars.album_id = arl.album_id
WHERE ars.id IS NULL
GROUP BY arl.user_id, arl.album_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM album_rating_state WHERE id LIKE 'backfill-%';
-- +goose StatementEnd
