# Bugs

Known bugs, regressions, and operational tech debt. Include repro steps or a pointer to them when known.

## Active

_(none)_

## Tech debt

### Reusable pattern for migration-side ID generation

Backfill migrations need to generate row IDs in SQL, but the codebase has no consistent approach. `lower(hex(randomblob(16)))` (32-char hex) appears in two prior migrations; an inline UUID-v4 construction was used in `20260517202814`. Neither matches runtime, which uses `uuid.NewString()` (36-char dashed UUID v4).

Next: when a third backfill is on the horizon, pick a convention (probably a custom goose binary + `RegisterFunc("uuid_v4", ...)`) and document it.

### Migrate raw `HX-Trigger` header writes to `httpx.SetHXTrigger`

`library/adapters/http.go` still sets several `HX-Trigger` headers via raw `w.Header().Set(...)`. Consider migrating them to the new `httpx.SetHXTrigger` helper for consistency.
