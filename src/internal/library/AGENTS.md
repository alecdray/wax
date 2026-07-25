# library — domain module

Rules: ../../../docs/architecture/archetypes/domain-module.md

Module-specific notes:
- `library.Library` is the central aggregate of the package: a user's albums plus derived artist/track sets used for dashboard filters. The package-name repetition is intentional — same pattern as `time.Time`, `context.Context`.
- Library owns the album view UI. Inline content from peer modules (e.g. sleeve notes from `notes`) is rendered by library's adapters using the peer module's `*Service`. Peer adapters never import `library/adapters` and library's adapters never import peer adapters.

## Product docs

| Doc | What | When |
|---|---|---|
| [library](../../../docs/product/library.md) | collection: own, wishlist, remove, browse | deciding library behaviour or how ownership should work |
| [radar](../../../docs/product/radar.md) | watch albums; eligibility rules; Spotify-inbox entry | deciding what radar surfaces to render or how eligibility interacts with ownership |
| [discover](../../../docs/product/discover.md) | search for albums to add to radar | deciding search behaviour or radar-add flows |
