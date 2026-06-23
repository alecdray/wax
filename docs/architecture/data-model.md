# Data Model

Cross-cutting design decisions for how the Wax domain is shaped. Per-entity meaning lives in each owning module's `README.md`; the authoritative schema lives in [`db/schema.sql`](../../db/schema.sql) with versioned changes in [`db/migrations/`](../../db/migrations/).

Modules under `src/internal/` align with entity groups. A module that owns a table is the only one that writes to it; cross-module reads flow through the owning module's `*Service`, never raw SQL.

## Key design decisions

- **Album is the anchor.** Almost every user interaction is scoped to an album: ratings, tags, notes, ownership, radar, listening history all hang off albums rather than tracks or artists.
- **Releases model format.** The same record can exist in multiple formats under one album. The user's relationship with each format is independent; an album appears in the library when at least one of its formats is owned or wishlisted.
- **Radar is independent of ownership.** A user can watch an album they have no other relationship with. Radar eligibility turns only on library membership — owning or wishlisting an album clears its radar entry, while a `removed` album stays radar-eligible, so radar and a `removed` release can coexist ([ADR 0005](../adr/0005-radar-eligibility-excludes-only-owned-wishlisted.md)). Beyond in-app actions, albums also reach the radar from a Spotify-side inbox the user opts into ([ADR 0004](../adr/0004-spotify-radar-playlist-entry.md)); like any external source, that inbox is owned by the feed that syncs it rather than the user, per the external-client rule. The [library module](../../src/internal/library/README.md) owns how these rules play out.
- **Ratings are an append-only log plus a state machine.** Every rating event is recorded with its note; the album's current lifecycle position (provisional / finalized) is tracked separately, transitioned only by explicit user action. See the [review module](../../src/internal/review/README.md) for the rating philosophy.
- **Tags are user-defined.** No global taxonomy. Each user builds their own vocabulary inside their own tag groups.
- **Annotations are all optional.** The library is valuable without ratings, tags, notes, or radar — each layer is additive.
