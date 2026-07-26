# Bugs

Known bugs, regressions, and operational tech debt. Include repro steps or a pointer to them when known.

## Active

### Mobile: filter popovers can render off-screen

On mobile, opening a filter popover (rating, format, genre, artist) in the unified search bar can cause the popover to extend beyond the right edge of the screen, making options unreachable.

### Library search bar: mid-request input loss

When typing in the search bar, a debounced HTMX request can fire and swap in the response between keystrokes, discarding characters typed after the request was sent. The input loses text the user is actively typing.

### Review modal: back and exit buttons overlap

On some screen sizes the back button (return to score entry from questionnaire) and the modal close/exit button render on top of each other.

### Album detail: track list is out of order

Tracks on the album detail page render in the wrong order.

## Tech debt

### Reusable pattern for migration-side ID generation

Backfill migrations need to generate row IDs in SQL, but the codebase has no consistent approach. `lower(hex(randomblob(16)))` (32-char hex) appears in two prior migrations; an inline UUID-v4 construction was used in `20260517202814`. Neither matches runtime, which uses `uuid.NewString()` (36-char dashed UUID v4).

Next: when a third backfill is on the horizon, pick a convention (probably a custom goose binary + `RegisterFunc("uuid_v4", ...)`) and document it.

### Migrate raw `HX-Trigger` header writes to `httpx.SetHXTrigger`

`library/adapters/http.go` still sets several `HX-Trigger` headers via raw `w.Header().Set(...)`. Consider migrating them to the new `httpx.SetHXTrigger` helper for consistency.

