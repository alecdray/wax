-- +goose Up
ALTER TABLE releases ADD COLUMN discogs_id TEXT;
ALTER TABLE releases ADD COLUMN label TEXT;
ALTER TABLE releases ADD COLUMN released_at DATETIME;

-- +goose Down
ALTER TABLE releases DROP COLUMN discogs_id;
ALTER TABLE releases DROP COLUMN label;
ALTER TABLE releases DROP COLUMN released_at;
