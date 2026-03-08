-- +goose Up

CREATE TABLE track_plays (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    track_id text not null references tracks(id) on delete cascade,
    album_id text not null references albums(id) on delete cascade,
    played_at datetime not null,
    unique(user_id, track_id, played_at)
);

CREATE TABLE listening_history_syncs (
    user_id text primary key references users(id) on delete cascade,
    last_synced_at datetime not null
);

-- +goose Down

DROP TABLE track_plays;
DROP TABLE listening_history_syncs;
