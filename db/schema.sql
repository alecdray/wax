CREATE TABLE goose_db_version (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version_id INTEGER NOT NULL,
		is_applied INTEGER NOT NULL,
		tstamp TIMESTAMP DEFAULT (datetime('now'))
	);
CREATE TABLE sqlite_sequence(name,seq);
CREATE TABLE users (
    id text primary key,
    spotify_id text not null unique,
    created_at datetime not null default current_timestamp,
    deleted_at datetime
, spotify_refresh_token text);
CREATE TABLE artists (
    id text primary key,
    spotify_id text not null unique,
    name text not null,
    created_at datetime not null default current_timestamp,
    deleted_at datetime
);
CREATE TABLE albums (
    id text primary key,
    spotify_id text not null unique,
    title text not null,
    created_at datetime not null default current_timestamp,
    deleted_at datetime
, image_url TEXT);
CREATE TABLE tracks (
    id text primary key,
    spotify_id text not null unique,
    title text not null,
    created_at datetime not null default current_timestamp,
    deleted_at datetime
);
CREATE TABLE releases (
    id text primary key,
    album_id text not null references albums(id) on delete cascade,
    format text not null check(format in ('digital', 'vinyl', 'cd', 'cassette')),
    created_at datetime not null default current_timestamp,
    deleted_at datetime, discogs_id TEXT, label TEXT, released_at DATETIME,
    unique(album_id, format)
);
CREATE TABLE user_releases (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    release_id text not null references releases(id) on delete cascade,
    added_at datetime not null default current_timestamp,
    deleted_at datetime, removed_at datetime,
    unique(user_id, release_id)
);
CREATE TABLE user_tracks (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    track_id text not null references tracks(id) on delete cascade,
    added_at datetime not null default current_timestamp,
    deleted_at datetime,
    unique(user_id, track_id)
);
CREATE TABLE user_artists (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    artist_id text not null references artists(id) on delete cascade,
    added_at datetime not null default current_timestamp,
    deleted_at datetime,
    unique(user_id, artist_id)
);
CREATE TABLE feeds (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    kind text not null check(kind in ('spotify')),
    created_at datetime not null default current_timestamp, last_sync_completed_at datetime, last_sync_started_at datetime, last_sync_status text default 'none' check(last_sync_status in ('none', 'success', 'failure', 'pending')),
    unique(user_id, kind)
);
CREATE TABLE IF NOT EXISTS "album_artists" (
    album_id text not null references albums(id) on delete cascade,
    artist_id text not null references artists(id) on delete cascade,
    unique(album_id, artist_id)
);
CREATE TABLE IF NOT EXISTS "album_tracks" (
    album_id text not null references albums(id) on delete cascade,
    track_id text not null references tracks(id) on delete cascade,
    unique(album_id, track_id)
);
CREATE TABLE track_plays (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    track_id text not null references tracks(id) on delete cascade,
    album_id text not null references albums(id) on delete cascade,
    played_at datetime not null,
    unique(user_id, track_id, played_at)
);
CREATE TABLE album_rating_log (
    id         text primary key,
    user_id    text not null references users(id) on delete cascade,
    album_id   text not null references albums(id) on delete cascade,
    rating     float not null,
    note       text,
    created_at datetime not null default current_timestamp
, state TEXT CHECK(state IN ('provisional', 'finalized', 'stalled')));
CREATE TABLE album_notes (
    id         TEXT NOT NULL PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    content    TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL DEFAULT current_timestamp,
    UNIQUE(user_id, album_id)
);
CREATE TABLE album_genres (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id     TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    genre_id     TEXT NOT NULL,
    genre_label  TEXT NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, genre_id)
);
CREATE TABLE album_moods (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    mood       TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, mood)
);
CREATE TABLE album_user_tags (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, tag)
);
CREATE TABLE tag_groups (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);
CREATE TABLE tags (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    group_id   TEXT REFERENCES tag_groups(id) ON DELETE SET NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);
CREATE TABLE album_tags (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    tag_id     TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, tag_id)
);
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
