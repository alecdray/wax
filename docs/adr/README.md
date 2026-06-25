# Architecture Decision Records

Short entries that capture **why** a decision was made, when the rationale would otherwise be lost once the old approach is gone.

## Decision log

| # | Decision | Summary |
|---|---|---|
| [0001](0001-library-visual-list.md) | Library shifts from table view to visual list | The library dashboard is a cover-art-first visual list with chip-bar filtering, replacing a sortable table; dashboard Spotify outlinks are dropped to keep navigation rooted in Wax. |
| [0002](0002-loading-feedback-for-network-actions.md) | Loading feedback for network actions | User-triggered network actions gain layered feedback — an app-wide indeterminate progress bar on every request, a busy/non-resubmittable state on discrete actions, and dim-and-overlay on in-place data reloads (a trailing spinner for append-style loads). |
| [0003](0003-rating-lifecycle-determined-by-action.md) | Rating lifecycle state is determined by the save action | Saving always lands provisional, finalizing always lands finalized, both from any prior state; saving a finalized album demotes it — the only un-finalize path. |
| [0004](0004-spotify-radar-playlist-entry.md) | A dedicated Spotify playlist is the radar's Spotify-side entry point | Albums reach the radar from inside Spotify via an opt-in, Wax-managed playlist; a periodic sync derives albums from its tracks, adds them, and removes only the tracks it ingested. |
| [0005](0005-radar-eligibility-excludes-only-owned-wishlisted.md) | Radar eligibility excludes only owned and wishlisted albums | An album is radar-eligible unless owned or wishlisted; a `removed` album can return to the radar, aligning the implementation with the documented "not in the library" definition. |
| [0006](0006-spotify-rate-limit-guard.md) | Spotify calls flow through a shared rate-limit guard that honors Retry-After | One process-wide guard paces all Spotify calls and pauses them for the `Retry-After` window on a 429; failed syncs back off, user actions fail fast while paused, and the access token is cached until expiry. |
| [0007](0007-bottom-nav-all-viewports.md) | Navigation is a bottom bar on all viewports | A single fixed bottom bar replaces the top header's navigation on every viewport incl. desktop; it carries the top-level destinations plus an account menu and is a domain-free layout-owned primitive, while a slim library-owned top header keeps the wordmark and feed sync-status visible on every authenticated page. |
| [0008](0008-radar-destination-discover-search-naming.md) | The watchlist destination is named Radar; "discover" names only its album-search mechanic | The top-level destination is named for the radar (the watchlist) it is built around; "discover" is retained as the narrower name of the Spotify album-search affordance that feeds it, kept distinct from the established radar concept. |
| [0009](0009-primary-genres-curated-facet.md) | Primary genres are an app-curated facet derived from the Wikidata genre graph | Every album is auto-assigned zero or more broad, app-curated primary genres for coarse filtering; sub-genres map up via the genre graph's ancestors, keeping the most-specific bucket per path (metal over rock), and genres are additive to — not a replacement for — free-form user tags. |

## Format

Lead with the decision — 1–2 sentences naming what was decided — under an `# h1` title. Then add only the **minimal context needed to understand that decision and its implications**: the constraint or trade-off that forced it. Rejected alternatives and consequences earn space only when they carry real weight, and as a brief clause or short list — never an obligatory section. An ADR that runs past a few short paragraphs is usually restating things that belong elsewhere.

The **current** state is the codebase — don't restate it. Keep implementation details out: no file names, class names, function names, or exact UI strings that a routine refactor would invalidate. If a sentence would need to change after such a refactor, it doesn't belong here.

## Naming

`NNNN-short-slug.md` — four-digit zero-padded prefix, never renumbered. A decision that replaces an earlier one is a new ADR; reference the prior number in the body. Don't edit the old one.

## When to write one

Write an ADR when a change replaces a meaningful prior approach, or locks in a foundational choice, that a future reader would otherwise wonder *"why is it like this?"* about. Routine changes don't qualify.
