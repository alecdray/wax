-- +goose Up
-- +goose StatementBegin
alter table users add column spotify_access_token text;
alter table users add column spotify_access_token_expires_at datetime;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users drop column spotify_access_token_expires_at;
alter table users drop column spotify_access_token;
-- +goose StatementEnd
