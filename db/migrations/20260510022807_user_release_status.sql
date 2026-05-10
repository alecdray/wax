-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_releases ADD COLUMN status TEXT NOT NULL DEFAULT 'owned'
    CHECK (status IN ('wishlist', 'owned', 'removed'));
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_releases ADD COLUMN created_at DATETIME;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE user_releases SET status = 'removed' WHERE removed_at IS NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE user_releases SET created_at = added_at WHERE created_at IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_releases RENAME COLUMN added_at TO status_updated_at;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE user_releases
   SET status_updated_at = removed_at
 WHERE status = 'removed' AND removed_at IS NOT NULL;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_releases DROP COLUMN removed_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_releases ADD COLUMN removed_at DATETIME;
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE user_releases SET removed_at = status_updated_at WHERE status = 'removed';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_releases RENAME COLUMN status_updated_at TO added_at;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_releases DROP COLUMN created_at;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_releases DROP COLUMN status;
-- +goose StatementEnd
