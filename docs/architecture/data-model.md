# Data Model

Cross-cutting design decisions for how the Wax domain is shaped. Per-entity meaning lives in each owning module's `README.md`; the authoritative schema lives in [`db/schema.sql`](../../db/schema.sql) with versioned changes in [`db/migrations/`](../../db/migrations/).

Modules under `src/internal/` align with entity groups. A module that owns a table is the only one that writes to it; cross-module reads flow through the owning module's `*Service`, never raw SQL.

## Key design decisions

- **Album is the anchor.** Almost every user interaction is scoped to an album: ratings, tags, notes, ownership, radar, listening history all hang off albums rather than tracks or artists.
- **Releases model format.** The same record can exist in multiple formats under one album. The user's relationship with each format is independent; an album appears in the library when at least one of its formats is owned or wishlisted.
- **Radar is independent of ownership.** A user can watch an album they have no other relationship with. An album is radar-eligible unless it is currently owned or wishlisted — a `removed` album can still be on the radar, so the two are not mutually exclusive ([ADR 0005](../adr/0005-radar-eligibility-excludes-only-owned-wishlisted.md)). Bringing the album into the library (owning or wishlisting it) clears the radar entry. The radar also has a Spotify-side inbox: a user can enqueue albums from within Spotify via an opt-in, Wax-managed playlist, which a periodic sync ingests and clears. The inbox is modelled as a feed kind, with the playlist handle stored on the feed row — not on the user — per the external-client rule. See [ADR 0004](../adr/0004-spotify-radar-playlist-entry.md).
- **Ratings are an append-only log plus a state machine.** Every rating event is recorded with its note; the album's current lifecycle position (provisional / finalized) is tracked separately, transitioned only by explicit user action. See the [review module](../../src/internal/review/README.md) for the rating philosophy.
- **Tags are user-defined.** No global taxonomy. Each user builds their own vocabulary inside their own tag groups.
- **Annotations are all optional.** The library is valuable without ratings, tags, notes, or radar — each layer is additive.
