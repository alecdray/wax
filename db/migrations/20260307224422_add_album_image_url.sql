-- +goose Up
ALTER TABLE albums ADD COLUMN image_url TEXT;

-- +goose Down
-- SQLite does not support DROP COLUMN; no-op
