# crates

Named, unordered album collections. A user creates crates, adds albums to them, and browses them on their own pages.

## Entities

- **Crate** — has an ID, user ID, name, and created timestamp. Owned by one user.
- **CrateMembership** — join between a crate and an album (`crate_albums`). An album can belong to multiple crates.

## Key behaviour

- Creating a crate trims and validates the name (rejects empty).
- Deleting a crate cascades to all its memberships; member albums are unaffected.
- Adding an album is idempotent — duplicate inserts are silently ignored.
- Member album data (title, artists, art, ratings, formats) is read through `library.Service`, not owned here.
- `SearchNonMembers` filters the user's full library in Go — no cross-module SQL.

## HTTP surface

| Route | Purpose |
|---|---|
| `GET /app/crates` | Crates index page |
| `GET /app/crates/new-modal` | Create crate modal fragment |
| `GET /app/crates/{id}` | Crate detail page |
| `GET /app/crates/{id}/members` | Member list fragment (re-fetched on `crateUpdated`) |
| `GET /app/crates/{id}/edit-modal` | Edit crate modal fragment |
| `GET /app/crates/{id}/edit-modal/search` | Non-member album search results fragment |
| `POST /app/crates` | Create crate |
| `DELETE /app/crates/{id}` | Delete crate |
| `POST /app/crates/{id}/albums/{albumId}` | Add album to crate |
| `DELETE /app/crates/{id}/albums/{albumId}` | Remove album from crate |
