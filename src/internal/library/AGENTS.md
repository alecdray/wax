# library — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- `library.Library` is the central aggregate of the package: a user's albums plus derived artist/track sets used for dashboard filters. The package-name repetition is intentional — same pattern as `time.Time`, `context.Context`.
- Library owns the album view UI. Inline content from peer modules (e.g. sleeve notes from `notes`) is rendered by library's adapters using the peer module's `*Service`. Peer adapters never import `library/adapters` and library's adapters never import peer adapters.

## Domain docs

| Doc | What | When |
|---|---|---|
| [library](../../../docs/domain/library.md) | ownership state machine (owned/wishlisted/removed), release model, radar eligibility rules and auto-clear | working on ownership transitions, release records, or anything that affects radar eligibility |

## Product docs

| Doc | What | When |
|---|---|---|
| [library](../../../docs/product/library.md) | collection: own, wishlist, remove, browse | deciding library behaviour or how ownership should work |
| [radar](../../../docs/product/radar.md) | watch albums; eligibility rules; Spotify-inbox entry | deciding what radar surfaces to render or how eligibility interacts with ownership |
| [discover](../../../docs/product/discover.md) | search for albums to add to radar | deciding search behaviour or radar-add flows |

## Relevant ADRs

| ADR | Decision | When |
|---|---|---|
| [ADR 0005](../../../docs/adr/0005-radar-eligibility-excludes-only-owned-wishlisted.md) | `removed` albums are radar-eligible — eligibility excludes only `owned` and `wishlisted` | working on radar entry points or ownership transitions; questioning why a removed album can be re-radared |
| [ADR 0008](../../../docs/adr/0008-radar-destination-discover-search-naming.md) | the watchlist is named Radar; "discover" names only the album-search mechanic within it | naming routes, labels, or search UI elements in library or radar flows |
