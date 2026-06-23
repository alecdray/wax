-- +goose Up
-- Widen feeds.kind to allow the radar inbox feed and add a per-feed external
-- source handle (the radar inbox playlist's id). SQLite can't alter a CHECK in
-- place, so rebuild the table. Nothing references feeds, so the rebuild is safe.
-- +goose StatementBegin
CREATE TABLE feeds_new (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    kind text not null check(kind in ('spotify', 'spotify_radar')),
    created_at datetime not null default current_timestamp,
    last_sync_completed_at datetime,
    last_sync_started_at datetime,
    last_sync_status text default 'none' check(last_sync_status in ('none', 'success', 'failure', 'pending')),
    source_ref text,
    unique(user_id, kind)
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO feeds_new (id, user_id, kind, created_at, last_sync_completed_at, last_sync_started_at, last_sync_status)
SELECT id, user_id, kind, created_at, last_sync_completed_at, last_sync_started_at, last_sync_status
FROM feeds;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE feeds;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE feeds_new RENAME TO feeds;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE feeds_old (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    kind text not null check(kind in ('spotify')),
    created_at datetime not null default current_timestamp,
    last_sync_completed_at datetime,
    last_sync_started_at datetime,
    last_sync_status text default 'none' check(last_sync_status in ('none', 'success', 'failure', 'pending')),
    unique(user_id, kind)
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO feeds_old (id, user_id, kind, created_at, last_sync_completed_at, last_sync_started_at, last_sync_status)
SELECT id, user_id, kind, created_at, last_sync_completed_at, last_sync_started_at, last_sync_status
FROM feeds
WHERE kind = 'spotify';
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE feeds;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE feeds_old RENAME TO feeds;
-- +goose StatementEnd
