-- +goose Up
-- +goose StatementBegin
ALTER TABLE album_ratings ADD COLUMN review text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'SQLite does not support DROP COLUMN; rollback is a no-op';
-- +goose StatementEnd
