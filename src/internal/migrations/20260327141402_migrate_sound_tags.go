package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/alecdray/wax/src/internal/genres"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upMigrateSoundTags, downMigrateSoundTags)
}

func upMigrateSoundTags(ctx context.Context, tx *sql.Tx) error {
	dag, err := genres.Load()
	if err != nil {
		slog.Warn("Genre DAG unavailable; skipping sound tag migration", "error", err)
		return nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT at.user_id, at.album_id, t.name
		FROM album_tags at
		JOIN tags t ON at.tag_id = t.id
		JOIN tag_groups tg ON t.group_id = tg.id
		WHERE tg.name = 'Sound'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type soundTag struct {
		userID  string
		albumID string
		name    string
	}
	var soundTags []soundTag
	for rows.Next() {
		var st soundTag
		if err := rows.Scan(&st.userID, &st.albumID, &st.name); err != nil {
			return err
		}
		soundTags = append(soundTags, st)
	}
	if err := rows.Close(); err != nil {
		return err
	}

	matched, unmatched := 0, 0
	for _, st := range soundTags {
		results := dag.Search(st.name)
		if len(results) == 0 {
			slog.Warn("No DAG match for sound tag", "tag", st.name, "album_id", st.albumID)
			unmatched++
			continue
		}
		best := results[0]
		_, err := tx.ExecContext(ctx,
			`INSERT INTO album_genres (id, user_id, album_id, genre_id, genre_label)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT (user_id, album_id, genre_id) DO UPDATE SET genre_label = excluded.genre_label`,
			uuid.NewString(), st.userID, st.albumID, best.ID, best.Label,
		)
		if err != nil {
			return err
		}
		slog.Info("Migrated sound tag", "tag", st.name, "genre_id", best.ID, "label", best.Label)
		matched++
	}

	slog.Info("Sound tag migration complete", "matched", matched, "unmatched", unmatched)
	return nil
}

func downMigrateSoundTags(ctx context.Context, tx *sql.Tx) error {
	return nil
}
