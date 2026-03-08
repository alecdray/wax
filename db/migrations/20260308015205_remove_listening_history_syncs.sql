-- +goose Up
DROP TABLE listening_history_syncs;

-- +goose Down
CREATE TABLE listening_history_syncs (
    user_id text primary key references users(id) on delete cascade,
    last_synced_at datetime not null
);
