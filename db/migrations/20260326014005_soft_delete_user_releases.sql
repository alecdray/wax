-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_releases ADD COLUMN removed_at datetime;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_releases DROP COLUMN removed_at;
-- +goose StatementEnd
