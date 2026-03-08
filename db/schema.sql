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
    deleted_at datetime,
    unique(album_id, format)
);
CREATE TABLE user_releases (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    release_id text not null references releases(id) on delete cascade,
    added_at datetime not null default current_timestamp,
    deleted_at datetime,
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
CREATE TABLE album_ratings (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    album_id text not null references albums(id) on delete cascade,
    rating float,
    created_at datetime not null default current_timestamp,
    updated_at datetime, review text,
    unique(user_id, album_id)
);
CREATE TABLE track_plays (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    track_id text not null references tracks(id) on delete cascade,
    album_id text not null references albums(id) on delete cascade,
    played_at datetime not null,
    unique(user_id, track_id, played_at)
);
