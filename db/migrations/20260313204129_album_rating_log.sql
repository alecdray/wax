-- +goose Up
-- +goose StatementBegin
CREATE TABLE album_rating_log (
    id         text primary key,
    user_id    text not null references users(id) on delete cascade,
    album_id   text not null references albums(id) on delete cascade,
    rating     float not null,
    note       text,
    created_at datetime not null default current_timestamp
);

INSERT INTO album_rating_log (id, user_id, album_id, rating, note, created_at)
SELECT id, user_id, album_id, rating, review, created_at
FROM album_ratings
WHERE rating IS NOT NULL;

DROP TABLE album_ratings;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'album_rating_log migration cannot be rolled back; data from album_ratings has been dropped';
-- +goose StatementEnd
