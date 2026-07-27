# Crates — Technical Spec (MVP)

## Module

New domain module: `src/internal/crates/` — domain module archetype.

```
src/internal/crates/
├── service.go
├── repo.go
├── crates.go          # domain types + DTOs
├── README.md
├── AGENTS.md
└── adapters/
    ├── http.go
    ├── routes.go
    └── views/
        ├── crates_page.templ
        ├── crate_detail_page.templ
        ├── crate_members_frag.templ       # member list; initial load + crateUpdated re-fetch
        ├── edit_crate_modal_frag.templ    # HTMX-loaded from crate detail
        └── create_crate_modal_frag.templ  # HTMX-loaded from crates index
```

---

## Schema

Two migrations:

```sql
-- crates
CREATE TABLE crates (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- crate_albums
CREATE TABLE crate_albums (
    id         TEXT PRIMARY KEY,
    crate_id   TEXT NOT NULL REFERENCES crates(id) ON DELETE CASCADE,
    album_id   TEXT NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    added_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(crate_id, album_id)
);
```

No position column — crates are unordered by design.

---

## Domain Types (`crates.go`)

```go
type CrateDTO struct {
    ID          string
    Name        string
    AlbumCount  int
    CreatedAt   time.Time
}

type CrateDetailDTO struct {
    CrateDTO
    Albums []library.AlbumDTO
}
```

`library.AlbumDTO` is injected from `library.Service` — crates does not own album data.

---

## Service API (`service.go`)

| Method | Signature | Notes |
|---|---|---|
| `ListCrates` | `(ctx) ([]CrateDTO, error)` | Ordered by `created_at DESC` |
| `GetCrate` | `(ctx, id) (CrateDetailDTO, error)` | Includes member albums |
| `CreateCrate` | `(ctx, name) (CrateDTO, error)` | Trims name; rejects empty |
| `DeleteCrate` | `(ctx, id) error` | Cascades via FK |
| `AddAlbum` | `(ctx, crateID, albumID) error` | Idempotent (ignore duplicate) |
| `RemoveAlbum` | `(ctx, crateID, albumID) error` | |
| `SearchNonMembers` | `(ctx, crateID, q string) ([]library.AlbumDTO, error)` | Calls `library.Service.GetAlbumsInLibrary`, excludes current member IDs, filters in Go by `q` against title + artist |
Constructor takes `*Repo` and `*library.Service` (for hydrating `AlbumDTO` in `GetCrate`).

**Library service addition:** `GetAlbumsByIDs(ctx contextx.ContextX, albumIDs []string) ([]AlbumDTO, error)` — new method on `library.Service`; reuses the existing `repo.GetAlbumsByIDs` with full hydration (ratings, tags, genres, notes). Called by `crates.Service.GetCrate`.

---

## Routes

```
GET    /app/crates                         → crates index page
GET    /app/crates/new-modal               → create crate modal fragment (HTMX-loaded)
GET    /app/crates/{id}                    → crate detail page
POST   /app/crates                         → create crate → OOB refresh index
DELETE /app/crates/{id}                    → delete crate → redirect /app/crates
GET    /app/crates/{id}/members            → member list fragment (initial render + crateUpdated re-fetch)
GET    /app/crates/{id}/edit-modal         → edit crate modal fragment (HTMX-loaded)
GET    /app/crates/{id}/edit-modal/search  → filtered non-member album results fragment (?q=)
POST   /app/crates/{id}/albums/{albumId}   → add album → OOB refresh modal content
DELETE /app/crates/{id}/albums/{albumId}   → remove album → OOB refresh modal content
```

All routes are authenticated (app sub-mux).

---

## Views

### `crates_page.templ` — Crates index

- Page templ; `ActiveNav: "crates"`
- Renders a list of `CrateDTO` cards: **name + count**
- Each card links to `/app/crates/{id}`
- "New Crate" button → `hx-get="/app/crates/new-modal"` `hx-swap="none"` → HTMX-loads `create_crate_modal_frag` into `#global-modal-container`

### `crate_detail_page.templ` — Crate detail

- Page templ; `ActiveNav: "crates"`
- Header: crate name + count
- "Edit" button → `hx-get="/app/crates/{id}/edit-modal"` `hx-swap="none"` → loads `edit_crate_modal_frag`
- Member albums: a stable-id region that loads via `hx-get="/app/crates/{id}/members"` and listens with `hx-trigger="crateUpdated from:body"` to re-fetch after the edit modal closes; rows rendered via `templates.AlbumRow`, read-only
- Delete crate: button at the bottom; `hx-confirm` (native confirm dialog); `hx-delete="/app/crates/{id}"` `hx-push-url="/app/crates"`

### `edit_crate_modal_frag.templ` — Edit crate (add + remove)

- Uses `templates.Modal` (HTMX OOB into `#global-modal-container`)
- Two sections:
  1. **Current members** — list of member albums with a remove button per row; `hx-delete="/app/crates/{id}/albums/{albumId}"` → OOB refreshes modal content
  2. **Add albums** — search input fires `hx-get="/app/crates/{id}/edit-modal/search?q="` with `keyup changed delay:300ms`; server returns a filtered fragment of non-member albums; each result has an "Add" button: `hx-post="/app/crates/{id}/albums/{albumId}"` → OOB refreshes modal content
- Each add/remove response includes an OOB `HXTrigger` dispatching `crateUpdated`; the member list region on the crate detail page listens with `hx-trigger="crateUpdated from:body"` and re-fetches itself — no close-event mechanism needed, member list is always current when the modal closes

### `create_crate_modal_frag.templ` — Create crate (from index)

- Simple name input + submit
- `POST /app/crates` → server responds with OOB swap refreshing the crates list + `ForceCloseModal`
- Uses `templates.Modal`

---

## New Primitive: `templates.AlbumRow`

New component in `core/templates/album_row.templ`.

```go
type AlbumRowProps struct {
    ID             string
    Title          string
    Artists        []string
    ImageURL       string
    Formats        []string
    Rating         *float64
    Finalized      bool
    OnRowClick     templ.Component  // wraps cover + title area; nil = inert
    OnFormatsClick templ.Component  // wraps format badges; nil = inert
    OnRatingClick  templ.Component  // wraps score readout; nil = inert
}

templ AlbumRow(props AlbumRowProps)
```

- Owns the full row: cover art, title, artists, format badges, rating
- Domain-free — plain values + `templ.Component` handlers; no domain type imports
- Callers decide behaviour: library passes HTMX-wired components for both handlers; crate detail passes `nil` for both (rows are inert — all interaction is in the edit modal)

**Library migration:** `albumListRow` in `library/adapters/views/albums_list_frag.templ` is refactored to call `templates.AlbumRow`, passing its existing formats button trigger as `OnFormatsClick`, detail link as `OnRowClick`, and rating recommender trigger as `OnRatingClick`. Behaviour and visual output unchanged.

---

## Bottom Nav

`core/templates/bottom_nav.templ` — add a 4th tab:

```diff
- <div class="max-w-md mx-auto grid grid-cols-3 h-14">
+ <div class="max-w-md mx-auto grid grid-cols-4 h-14">
  @bottomNavTab("library", ...)
  @bottomNavTab("radar", ...)
+ @bottomNavTab("crates", active == "crates", "/app/crates", "archive", "Crates", false)
  @bottomNavProfileTab()
```

Icon: `archive` (Bootstrap Icons — no fill variant, stays outline when active).

---

## Wire-up (`server/`)

1. Instantiate `crates.NewRepo(queries)`
2. Instantiate `crates.NewService(repo, libraryService)`
3. Instantiate `crates.adapters.NewHttpHandler(cratesService)`
4. Call `crates.adapters.RegisterRoutes(appMux, handler)`

---

## v1 Deferrals

| Feature | Notes |
|---|---|
| Add to crate from album detail | Picker modal on album detail page; new route `GET /app/crates/picker?albumId=`; no album detail changes in MVP |
| Rename crate | New endpoint + inline edit on detail page |
| Cover art mosaic on index cards | 2×2 grid of `image_url` from first 4 members; repo query change only |
