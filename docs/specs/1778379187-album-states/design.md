# Album lifecycle states — design

Give albums in a user's library lifecycle states beyond "in collection / not in collection." Two grains, deliberately separate:

- **Album-level radar** — per-`(user, album)` intent: "I want to listen / on my radar, haven't picked a format."
- **Release-level lifecycle** — per-`(user, release)` enum on `user_releases`: `wishlist | owned | removed`.

Scope of this spec: data model, schema migration, state transitions, repo/service surface, tests. UI is out of scope and follows in a later spec.

## Conceptual model

```
album-level                    release-level
                               (per format: digital | vinyl | cd | cassette)

(no radar row)                 (no user_releases row)
       │                              │
       ▼ add to radar                 ▼ wishlist  /  acquire directly
user_album_radar               status = 'wishlist' ◄──► status = 'owned'
       │                              │                     │
       │                              │ decline             │ remove
       │                              ▼ (delete row)        ▼
       │                              ─                status = 'removed'
       │                              │                     │
       │                              ▼                     ▼ re-acquire
       └──── wiped on any release-level activity ──► back to 'owned'
```

Two cross-grain rules:

1. Adding any release to wishlist or owned **wipes the album's radar row**. Radar is strictly pre-decision.
2. The user's **collection** is exactly `user_releases.status = 'owned'`. `wishlist` and `removed` rows live in the same table but are excluded from collection queries.

"None of the above" is the absence of rows — no radar entry, no `user_releases` row.

## Schema

### `user_releases` — add lifecycle status, normalize timestamps

```sql
ALTER TABLE user_releases ADD COLUMN status TEXT NOT NULL DEFAULT 'owned'
    CHECK (status IN ('wishlist', 'owned', 'removed'));
ALTER TABLE user_releases ADD COLUMN created_at DATETIME;     -- nullable until backfilled
ALTER TABLE user_releases RENAME COLUMN added_at TO status_updated_at;
-- removed_at dropped after backfill (see Migration plan)
```

Column meanings after migration:

| Column              | Meaning                                                                                           |
| ------------------- | ------------------------------------------------------------------------------------------------- |
| `status`            | `wishlist` \| `owned` \| `removed`. Source of truth for current lifecycle state.                  |
| `created_at`        | Immutable. First time this `(user, release)` row was created (any state). Used for "since" UIs.   |
| `status_updated_at` | Bumped on every status transition. Tells you when the current state was entered.                 |

The pre-existing unused `deleted_at` column on `user_releases` (legacy from the initial migration; superseded by `removed_at` and never referenced) is left in place — separate cleanup.

Single-row state invariant is enforced by the check constraint: each `(user_id, release_id)` pair is in exactly one of the three states at any time.

### `user_album_radar` — new table

```sql
CREATE TABLE user_album_radar (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id)
);
```

Radar is presence-based — no enum, no soft-delete. Removing from radar deletes the row. Mirrors the per-`(user, album)` shape of `album_rating_state`.

## Migration plan

Three migrations, paired the way `album_rating_state` was paired with `backfill_rating_state`:

**Migration 1 — schema and backfill (single transaction):**

```sql
-- Add nullable / defaulted columns first
ALTER TABLE user_releases ADD COLUMN status TEXT NOT NULL DEFAULT 'owned'
    CHECK (status IN ('wishlist', 'owned', 'removed'));
ALTER TABLE user_releases ADD COLUMN created_at DATETIME;

-- Backfill from the legacy timestamp columns
UPDATE user_releases SET status     = 'removed' WHERE removed_at IS NOT NULL;
UPDATE user_releases SET created_at = added_at  WHERE created_at IS NULL;

-- Rename added_at to its new role
ALTER TABLE user_releases RENAME COLUMN added_at TO status_updated_at;

-- For removed rows, status_updated_at should reflect the removal time
UPDATE user_releases
   SET status_updated_at = removed_at
 WHERE status = 'removed' AND removed_at IS NOT NULL;

-- Drop the now-redundant column
ALTER TABLE user_releases DROP COLUMN removed_at;
```

**Migration 2 — enforce `created_at NOT NULL`:**

SQLite cannot add a NOT NULL constraint to an existing column without a table rebuild. Use the standard rebuild pattern (CREATE new, INSERT SELECT, DROP old, RENAME). Do this only after Migration 1 has run and `created_at` is fully populated.

**Migration 3 — create the radar table:**

```sql
CREATE TABLE user_album_radar (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id)
);
```

After all three: regenerate sqlc and update queries (`db/queries/user_releases.sql`, new `db/queries/user_album_radar.sql`).

## State transitions (canonical list)

| Trigger                                  | Effect                                                                                                                                                         |
| ---------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Add album to radar                       | Insert `user_album_radar` row. No-op if any `user_releases` row already exists for this `(user, album)` (radar is pre-decision; log the no-op).                |
| Remove album from radar                  | Delete `user_album_radar` row.                                                                                                                                 |
| Add release to wishlist                  | Upsert `user_releases` row, `status='wishlist'`, `status_updated_at=now`. On insert: `created_at=now`. **Delete the album's radar row if present.**            |
| Acquire (wishlist → owned)               | Update `status='owned'`, bump `status_updated_at`. `created_at` unchanged.                                                                                     |
| Add release directly to collection       | Upsert `user_releases`, `status='owned'`, `status_updated_at=now`. Delete radar row. (Replaces today's `UpsertUserRelease`.)                                   |
| Remove release from collection           | Update `status='removed'`, bump `status_updated_at`. (Replaces today's `SoftDeleteUserRelease[sByAlbumId]`.)                                                   |
| Re-acquire a removed release             | Update `status='owned'`, bump `status_updated_at`.                                                                                                             |
| Decline a wishlist item                  | Delete the `user_releases` row outright. **No `removed` tombstone for wishlist declines** — `removed` is reserved for formerly-owned releases.                 |

All transitions that touch both `user_album_radar` and `user_releases` must run inside a `db.WithTx` block.

## Repo / service surface

### `library.Repo` — additions

```go
// Album-level radar
AddAlbumToRadar(ctx, userID, albumID string) error
RemoveAlbumFromRadar(ctx, userID, albumID string) error
GetRadarAlbums(ctx, userID string) ([]AlbumDTO, error)
IsAlbumOnRadar(ctx, userID, albumID string) (bool, error)

// Release-level wishlist
AddReleaseToWishlist(ctx, userID, albumID string, format models.ReleaseFormat, releaseID string) (string, error)
RemoveReleaseFromWishlist(ctx, userID, releaseID string) error  // hard delete
GetWishlistReleases(ctx, userID string) ([]ReleaseDTO, error)
```

### `library.Repo` — modifications

- `GetUserReleases`, `GetUserReleasesByAlbumID`, `GetAlbumFormats`: add `WHERE status = 'owned'` so collection queries don't surface wishlist/removed rows.
- `AddAlbumToCollection` and `AddOwnedRelease`: write `status='owned'`, bump `status_updated_at`, and delete the matching `user_album_radar` row inside the same transaction.
- `SoftDeleteUserRelease` / `SoftDeleteUserReleasesByAlbumID` → renamed to `MarkReleaseRemoved` / `MarkReleasesRemovedByAlbumID`. Implementation switches from `SET removed_at = current_timestamp` to `SET status = 'removed', status_updated_at = current_timestamp`.

### DTOs

- `ReleaseDTO` gains `Status models.UserReleaseStatus` and `CreatedAt *time.Time`. `AddedAt` retains its current meaning ("when this row last entered the `owned` state") for now to avoid touching every caller; new code should prefer `Status` + `StatusUpdatedAt` once exposed.
- New small `RadarDTO { AlbumID string; CreatedAt time.Time }`.

### Service helpers

- `library.Service.AcquireFromWishlist(ctx, userID, releaseID)` — convenience wrapper over the wishlist→owned transition, runs in a single tx.
- Existing add-to-collection service paths get a "clear radar" side-effect; this is internal to repo, not a new service method.

## Testing

Extend `src/internal/library/service_test.go` and add repo-level tests as needed.

- **Per-transition tests**, one subtest each, named as plain behaviours per the project's testing conventions:
  - `"adding to wishlist clears radar"`
  - `"acquiring from wishlist preserves created_at and bumps status_updated_at"`
  - `"removing an owned release sets status='removed'"`
  - `"re-acquiring a removed release returns it to 'owned'"`
  - `"declining a wishlist item deletes the row"`
  - `"adding to radar is a no-op when any user_releases row exists for the album"`

- **Invariant tests**:
  - A release is in exactly one of `wishlist | owned | removed` (the check constraint guarantees this; the test asserts the application doesn't try to write disallowed states).
  - An album with any non-radar `user_releases` row has no `user_album_radar` row.

- **Query-scope tests**:
  - `GetUserReleases` / `GetUserReleasesByAlbumID` / `GetAlbumFormats` exclude wishlist and removed rows.
  - `GetWishlistReleases` returns only `status='wishlist'` rows.
  - `GetRadarAlbums` returns only albums with a radar row and excludes albums that have any `user_releases` row.

## Out of scope (deferred to follow-up specs)

- All UI changes (radar/wishlist views, badges on album rows, filter/sort surfacing).
- Spotify auto-sync interactions with the new states.
- The roadmap's "Hidden Albums" feature — likely a separate album-level concern, not folded into this spec.
- Cleanup of the legacy unused `deleted_at` column on `user_releases`.
