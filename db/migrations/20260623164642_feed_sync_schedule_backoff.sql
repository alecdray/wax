-- +goose Up
-- +goose StatementBegin
-- next_sync_at: when a feed is next eligible to sync. NULL = due immediately
-- (never scheduled). consecutive_failures drives the failure backoff.
alter table feeds add column next_sync_at datetime;
alter table feeds add column consecutive_failures integer not null default 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table feeds drop column consecutive_failures;
alter table feeds drop column next_sync_at;
-- +goose StatementEnd
