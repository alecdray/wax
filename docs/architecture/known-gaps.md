# Known Architectural Gaps

This doc tracks current architectural violations in the codebase — places where the rules in [`archetypes/`](archetypes/) (and module `CLAUDE.md` files) don't yet match reality. Each entry describes the gap, why it exists, and what closing it would require.

Gaps are tracked here (and not enumerated inside the archetype docs themselves) so the rule docs stay durable and conceptual, while the concrete list of divergences lives in one searchable place.

## Peer adapters import `library/adapters/views`

**Rule violated:** [`archetypes/domain-module.md`](archetypes/domain-module.md) — *Other domain modules' `adapters/` (and `adapters/views/`) packages may not be imported by peers.*

**Where:**

- `src/internal/review/adapters/http.go` — renders library view components (`AlbumScoreReadout`, `AlbumScoreBadge`, `AlbumRatingHistory`, `AlbumRowTagsSection`) when finalising a rating, to swap the affected library UI back in over HTMX.
- `src/internal/tags/adapters/http.go` — renders library view components (`AlbumTagsCell`, `AlbumRowTagsSection`) when saving tags, for the same reason.

**Why it exists:** these handlers mutate album state (rating, tags) and then need to update slices of the library UI in the response. Today they call library's templ components directly to do that.

**What closing it requires:** one of the following architectural moves:

- The mutating module returns its own fragment and library composes the final response — would require coordinating multi-module renders at the route boundary, or returning multiple swap targets via OOB from a library-owned handler.
- Library exposes a `*Service` method (or a small dedicated interface) that returns a renderer for *the parts of the album view affected by a state change*, and peer modules call that. Keeps the templ components private to library.
- The shared fragments move into a neutral location (a primitive in `core/templates/`, or a shared sub-package). Viable only if those fragments genuinely don't depend on library-specific domain types — the current ones do.

## Page templs don't wrap in `PageLayoutComponent`

**Rule violated:** [`../design/archetypes/page-templ.md`](../design/archetypes/page-templ.md) — *A page templ wraps its body in `@templates.PageLayoutComponent`.*

**Where:** every page templ in the codebase. Pages currently wrap directly in `@templates.RootComponent` and emit their own navbar/container chrome inline. `PageLayoutComponent` is declared in `core/templates/layout.templ` but has zero callers.

**Why it exists:** the layout primitive was introduced after the pages were written; the migration to actually thread it through hasn't been done.

**What closing it requires:** decide whether the page-layout primitive should compose `RootComponent` + navbar + content container as a single wrapper (in which case the existing inline chrome moves into the primitive and pages call only `PageLayoutComponent`), then update each page accordingly. Alternatively, if the chrome legitimately differs per page, the doc should be revised to describe what the actual page layout pattern is.
