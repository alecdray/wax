import { execSync } from 'node:child_process';

// The reviews suite needs to set up specific RatingState values that no public
// API offers (delete-state, demote-finalized-to-provisional). Direct sqlite3
// shell-out keeps the helper dependency-free; the suite's "real backend" rule
// permits direct DB inserts when no API surface covers the needed shape.

const DB_PATH = process.env.GOOSE_DBSTRING ?? './tmp/db.sql';

function execSql(sql: string): string {
  return execSync(`sqlite3 ${DB_PATH} ${JSON.stringify(sql)}`, { encoding: 'utf8' });
}

// resetAlbumRating wipes every rating-log entry and the rating-state row for
// one user/album, leaving the album in the "never been rated" shape (no log
// entries, no state row).
export function resetAlbumRating(userId: string, albumId: string) {
  execSql(
    `DELETE FROM album_rating_log WHERE user_id = '${userId}' AND album_id = '${albumId}';` +
      ` DELETE FROM album_rating_state WHERE user_id = '${userId}' AND album_id = '${albumId}';`,
  );
}

// setRatingStateValue forces the album's rating-state row to the given
// lifecycle value, creating the row if it does not exist. Used to position an
// album as provisional or finalized for tests that need a specific state
// without exercising the full save flow.
export function setRatingStateValue(userId: string, albumId: string, state: 'provisional' | 'finalized') {
  const stateId = `${userId.slice(0, 8)}-${albumId.slice(0, 8)}-state`;
  execSql(
    `INSERT INTO album_rating_state (id, user_id, album_id, state, created_at, updated_at)` +
      ` VALUES ('${stateId}', '${userId}', '${albumId}', '${state}', current_timestamp, current_timestamp)` +
      ` ON CONFLICT(user_id, album_id) DO UPDATE SET state = excluded.state, updated_at = current_timestamp;`,
  );
}

// seedRatingLogEntry inserts a rating-log row with an explicitly-set
// created_at, dated several seconds in the past so that any subsequent save
// during the test wins the latest-rating comparison unambiguously. Sidesteps
// the second-resolution CURRENT_TIMESTAMP pitfall.
export function seedRatingLogEntry(userId: string, albumId: string, score: number) {
  const id = `seeded-${userId.slice(0, 8)}-${albumId.slice(0, 8)}-${Date.now()}`;
  execSql(
    `INSERT INTO album_rating_log (id, user_id, album_id, rating, created_at)` +
      ` VALUES ('${id}', '${userId}', '${albumId}', ${score}, datetime('now', '-10 seconds'));`,
  );
}

// clearAllRatingStates wipes every rating-state row for the user, leaving
// every album in the user's library in the "no state row" shape. Used by the
// dashboard carousel suite to construct a clean baseline before positioning
// a small number of albums into known states.
export function clearAllRatingStates(userId: string) {
  execSql(`DELETE FROM album_rating_state WHERE user_id = '${userId}';`);
}

// getLibraryAlbumIds returns the IDs of every album the user currently owns
// (status = 'owned' on at least one user_release row). Sorted for determinism.
export function getLibraryAlbumIds(userId: string): string[] {
  const out = execSql(
    `SELECT DISTINCT album_id FROM user_releases` +
      ` WHERE user_id = '${userId}' AND status = 'owned'` +
      ` ORDER BY album_id;`,
  );
  return out
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.length > 0);
}

// getAlbumTitle returns the title of a wax album row by ID. Lets specs that
// want to scope a list-row search to a known album do so via the unified-bar
// `q=` filter, which collapses the list to rows whose title or artist
// contains the substring — the cleanest way to pin a deterministic single-row
// view that respects the suite's testid-only selector rule.
export function getAlbumTitle(albumId: string): string {
  return execSql(`SELECT title FROM albums WHERE id = '${albumId}';`).trim();
}
