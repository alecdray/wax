-- name: UpsertTrackPlay :exec
INSERT INTO track_plays (id, user_id, track_id, album_id, played_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (user_id, track_id, played_at) DO NOTHING;

-- name: GetLastPlayedAtByAlbumIds :many
SELECT album_id, MAX(played_at) as last_played_at
FROM track_plays
WHERE user_id = ? AND album_id IN (sqlc.slice('album_ids'))
GROUP BY album_id;

-- name: GetRecentlyPlayedAlbums :many
-- rating_state and latest_rating come from the album's latest album_rating_log
-- entry (same row, so score and state are coherent). State is read from the log
-- rather than album_rating_state because the log entry is written on every
-- rating action and is the populated source. rating and state are NOT NULL in
-- schema, so a LEFT JOIN miss would make sqlc scan NULL into a non-nullable Go
-- type; COALESCE keeps them off the wire as NULL (0 rating, empty state) and
-- has_rating carries log-row presence to the app. The latest entry is the row
-- with the greatest created_at, tie-broken by id, matching
-- GetLatestUserAlbumRatings.
SELECT albums.*, MAX(track_plays.played_at) as last_played_at,
    COALESCE((
        SELECT GROUP_CONCAT(a.name, ', ')
        FROM (SELECT DISTINCT ar.id, ar.name FROM album_artists aa JOIN artists ar ON ar.id = aa.artist_id WHERE aa.album_id = albums.id) AS a
    ), '') as artist_names,
    EXISTS (
        SELECT 1 FROM user_releases
        JOIN releases ON releases.id = user_releases.release_id
        WHERE releases.album_id = albums.id AND user_releases.user_id = track_plays.user_id AND user_releases.status = 'owned'
    ) as in_library,
    EXISTS (
        SELECT 1 FROM user_album_radar
        WHERE user_album_radar.album_id = albums.id AND user_album_radar.user_id = track_plays.user_id
    ) as on_radar,
    COALESCE(arl.state, '') AS rating_state,
    COALESCE(arl.rating, 0) AS latest_rating,
    CASE WHEN arl.rating IS NULL THEN 0 ELSE 1 END AS has_rating
FROM track_plays
JOIN albums ON albums.id = track_plays.album_id
LEFT JOIN (
    SELECT arl.album_id, arl.rating, arl.state
    FROM album_rating_log arl
    JOIN (
        SELECT arl2.album_id, MAX(arl2.id) AS max_id
        FROM album_rating_log arl2
        WHERE arl2.user_id = ?
          AND (arl2.album_id, arl2.created_at) IN (
              SELECT arl3.album_id, MAX(arl3.created_at)
              FROM album_rating_log arl3
              WHERE arl3.user_id = ?
              GROUP BY arl3.album_id
          )
        GROUP BY arl2.album_id
    ) latest ON arl.album_id = latest.album_id AND arl.id = latest.max_id
) arl ON arl.album_id = albums.id
WHERE track_plays.user_id = ?
GROUP BY albums.id
ORDER BY last_played_at DESC
LIMIT 20;

