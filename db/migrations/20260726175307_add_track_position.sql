-- +goose Up
ALTER TABLE tracks ADD COLUMN disc_number INTEGER NOT NULL DEFAULT 1;
ALTER TABLE tracks ADD COLUMN track_number INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE tracks DROP COLUMN disc_number;
ALTER TABLE tracks DROP COLUMN track_number;
