# Known Architectural Gaps

This doc tracks current architectural violations in the codebase — places where the rules in [`archetypes/`](archetypes/) (and module `CLAUDE.md` files) don't yet match reality. Each entry describes the gap, why it exists, and what closing it would require.

Gaps are tracked here (and not enumerated inside the archetype docs themselves) so the rule docs stay durable and conceptual, while the concrete list of divergences lives in one searchable place.

## Peer adapters import `library/adapters/views`

**Rule violated:** [`archetypes/domain-module.md`](archetypes/domain-module.md) — *Other domain modules' `adapters/` (and `adapters/views/`) packages may not be imported by peers.*

**Where:**

- `src/internal/review/adapters/http.go` — renders library view components (`AlbumScoreReadout`, `AlbumScoreBadge`, `AlbumRatingHistory`, `AlbumRowTagsSection`) when finalising a rating, to swap the affected library UI back in over HTMX.
- `src/internal/tags/adapters/http.go` — renders library view components (`AlbumTagsFrag`, `AlbumRowTagsSection`) when saving tags, for the same reason.

**Why it exists:** these handlers mutate album state (rating, tags) and then need to update slices of the library UI in the response. Today they call library's templ components directly to do that.

**What closing it requires:** one of the following architectural moves:

- The mutating module returns its own fragment and library composes the final response — would require coordinating multi-module renders at the route boundary, or returning multiple swap targets via OOB from a library-owned handler.
- Library exposes a `*Service` method (or a small dedicated interface) that returns a renderer for *the parts of the album view affected by a state change*, and peer modules call that. Keeps the templ components private to library.
- The shared fragments move into a neutral location (a primitive in `core/templates/`, or a shared sub-package). Viable only if those fragments genuinely don't depend on library-specific domain types — the current ones do.
