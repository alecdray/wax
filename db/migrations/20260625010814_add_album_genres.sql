-- +goose Up
-- Structured genre facet (ADR 0009). Genres are app-curated and album-intrinsic
-- (Discogs-derived), so they are global per album rather than per user — unlike
-- tags. album_genres holds the resolved leaf genre nodes (Wikidata Q-ids);
-- primaries are derived from these at read time via the genre graph.
--
-- An earlier, never-migrated per-user album_genres table drifted into some dev
-- databases; drop it so this migration is the authoritative definition.
-- +goose StatementBegin
DROP TABLE IF EXISTS album_genres;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE album_genres (
    id          TEXT PRIMARY KEY,
    album_id    TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    genre_id    TEXT NOT NULL,
    genre_label TEXT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(album_id, genre_id)
);
-- +goose StatementEnd

-- album_genre_enrichment records that an album's genres were resolved, so the
-- backfill task processes each album once and an album with no genres (queried,
-- nothing matched) is distinguishable from one not yet processed.
-- +goose StatementBegin
CREATE TABLE album_genre_enrichment (
    album_id    TEXT PRIMARY KEY REFERENCES albums(id) ON DELETE CASCADE,
    enriched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE album_genre_enrichment;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE album_genres;
-- +goose StatementEnd
