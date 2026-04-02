# Labels Redesign — Implementation Plan

## Context

The generic `tags` system (tags table + tag_groups + album_tags) used "Sound" and "Mood" as loosely-typed groups with no semantic distinction. This prevented genre-hierarchy filtering (e.g., "show all Electronic albums") because genres had no DAG node IDs attached.

## Approach

Replace the generic tags abstraction with three first-class entities:

- **Genres** — strict DAG-backed (Wikidata node ID stored); enables hierarchical filtering
- **Moods** — freeform text; autocomplete from user's existing mood assignments
- **User Tags** — freeform text; general labels, same mechanics as moods

Key decisions:

- **No separate library tables for moods/user_tags** — autocomplete derived via `SELECT DISTINCT` queries
- **Genre search is DAG-only** — no DB roundtrip; client calls `/app/labels/genre-search?q=...`
- **Hierarchical genre filtering** — `FilterParams.ExpandGenreDescendants(dag)` expands selected genre IDs to include all DAG descendants before the filter loop
- **Sound tags dropped in migration** — Sound-group tags cannot be reliably mapped to Wikidata QIDs automatically; `cmd/migrate-sound-tags` is provided to run against a backup before migration
- **Mood and user-tag tags migrated** — Mood-group tags → `album_moods`; ungrouped tags → `album_user_tags`

## Architecture

### DB Schema (3 new tables, replace tags/tag_groups/album_tags)

```sql
CREATE TABLE album_genres (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id     TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    genre_id     TEXT NOT NULL,   -- Wikidata QID e.g. "Q11399"
    genre_label  TEXT NOT NULL,   -- snapshot label for display
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, genre_id)
);

CREATE TABLE album_moods (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    mood       TEXT NOT NULL,   -- normalized lowercase
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, mood)
);

CREATE TABLE album_user_tags (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,   -- normalized lowercase
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, album_id, tag)
);
```

### Package: `src/internal/labels/`

Single package replacing `src/internal/tags/`. Key types:

```go
type GenreDTO struct { ID, Label string }
type GenreSuggestion struct { ID, Label, ParentLabel string }
type AlbumLabels struct { Genres []GenreDTO; Moods []string; UserTags []string }
```

Key service methods:
- `SearchGenres(query string) []GenreSuggestion` — DAG only, no DB
- `SetAlbumGenres / GetAlbumGenresByAlbumIds`
- `SetAlbumMoods / GetAlbumMoodsByAlbumIds / GetDistinctUserMoods`
- `SetAlbumUserTags / GetAlbumUserTagsByAlbumIds / GetDistinctUserTags`
- `GetAlbumLabelsByAlbumIds / GetAlbumLabels`

### Library Service Changes

`AlbumDTO` fields replaced:
- `Tags []tags.TagDTO` → `Genres []labels.GenreDTO`, `Moods []string`, `UserTags []string`

`FilterParams` additions:
- `GenreIDs []string` — expanded to include descendants before filtering
- `Moods []string` — exact match (any)
- `UserTags []string` — exact match (any)

`NewService` takes `labelsService *labels.Service` instead of `tagsService *tags.Service`.

### HTTP Layer

Routes (replacing `/app/tags/*`):
```
GET  /app/labels/album          → GetLabelsModal
POST /app/labels/album          → SubmitAlbumLabels
GET  /app/labels/genre-search   → SearchGenres (JSON for Alpine.js)
```

Library HTTP handler now takes `labelsService *labels.Service` and `genreDAG *genres.DAG` to resolve genre labels and expand descendants at query time.

### Tagging Modal UI (`src/internal/labels/adapters/labels.templ`)

Three Alpine.js sections:
- **Genres**: search input → debounced fetch → dropdown with parent breadcrumb → chips with hidden `genre[]` inputs
- **Moods**: freeform input + autocomplete from `allMoods` → `mood[]` hidden inputs
- **User Tags**: same as moods → `tag[]` hidden inputs

### Dashboard Filter Chips

New chips after Artist: Genre, Mood, Tag.
- Genre chip: Alpine.js search dialog (same DAG search); active state shows "N Genres"
- Mood/Tag chips: checkbox list from user's distinct values; active state shows value or count

`buildAlbumsTableURL` / `buildAlbumsPageURL` now include `genre`, `mood`, `tag` query params.

`DashboardPageProps` additions: `AllMoods []string`, `AllUserTags []string`.

### Data Migration (`db/migrations/20260327141401_redesign_tagging.sql`)

1. Create 3 new tables
2. Migrate Mood-group tags → `album_moods`
3. Migrate ungrouped tags → `album_user_tags`
4. Drop `album_tags`, `tags`, `tag_groups`

Sound tags cannot be migrated via SQL — use `cmd/migrate-sound-tags` against a backup first.
