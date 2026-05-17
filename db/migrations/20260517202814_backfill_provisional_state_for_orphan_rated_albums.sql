-- Backfill: any (user, album) that has at least one album_rating_log entry but
-- no album_rating_state row gets a provisional state row. The post-rework
-- lifecycle holds the invariant "rated → has a state row" — the first save
-- implicitly creates the row as provisional. This migration repairs legacy
-- orphans from before that invariant was enforced.
--
-- Forward-only: no Down section. Once these rows exist, users may interact
-- with them (save, finalize), at which point a rollback that deleted them
-- would silently destroy user-meaningful state.
--
-- Inline UUID v4 generation matches the runtime ID format produced by
-- `uuid.NewString()` in src/internal/review/repo.go. (See backlog: a reusable
-- pattern for SQL-side ID generation is a follow-up.)

-- +goose Up
-- +goose StatementBegin
INSERT INTO album_rating_state (id, user_id, album_id, state, created_at, updated_at)
SELECT
    lower(hex(randomblob(4))) || '-' ||
        lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))), 2) || '-' ||
        substr('89ab', 1 + abs(random() % 4), 1) || substr(lower(hex(randomblob(2))), 2) || '-' ||
        lower(hex(randomblob(6))) AS id,
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
