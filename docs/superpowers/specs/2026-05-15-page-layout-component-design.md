# PageLayoutComponent Migration

## Goals & Non-Goals

### Goals
- Close Gap 2 in [`docs/architecture/known-gaps.md`](../../architecture/known-gaps.md): every page templ wraps in `PageLayoutComponent`.
- Align the page-templ archetype doc with how page chrome actually composes in the codebase.

### Non-Goals
- Closing Gap 1 (peer adapters importing `library/adapters/views`). Separate change.
- Per-page migration mechanics, commit boundaries, and validation steps. Implementation-plan territory.
- Introducing additional page-level chrome (footers, banners, auth-state utility bars) before a real second caller asks for it.

## Background

`PageLayoutComponent` is declared in `src/internal/core/templates/layout.templ` but has zero callers. Every page templ wraps directly in `RootComponent` and emits its navbar and content container inline. The [page-templ archetype doc](../../design/archetypes/page-templ.md) currently declares pages must wrap in `PageLayoutComponent` and that the layout supplies the navbar — neither matches reality.

Two underlying tensions explain the drift:

- **The navbar is module-specific.** Library pages use `LibraryHeaderBarFrag` with module-specific tab state; login uses no navbar at all. `PageLayoutComponent` sits in `core/templates/`, so it cannot import a domain module's fragment to render the navbar itself.
- **The content-container shape varies per page.** Dashboard is full-width, album detail is a `max-w-2xl` card, discover is `h-dvh` viewport-locked for nested scroll, login is a centred splash. No single container belongs in the shared primitive.

## Decision

Every page templ wraps in `PageLayoutComponent`. The primitive owns cross-cutting page chrome — the root chrome (DOCTYPE, head, fonts, HTMX/Alpine wiring, body wrapper, modal container) — and exposes a slot for the navbar. Each page supplies its own navbar (or omits it) and owns its content-container shape inside the layout.

The exact slot mechanism (prop name, signature, nil semantics) lives in the code. The design records the pattern, not the function signature.

## Doc updates

- [`docs/design/archetypes/page-templ.md`](../../design/archetypes/page-templ.md) — *Shape* section revised: pages wrap in `PageLayoutComponent`; the layout supplies root chrome; the navbar is page-provided through the layout's slot; the content-container shape stays with the page.
- [`docs/architecture/known-gaps.md`](../../architecture/known-gaps.md) — Gap 2 entry removed when the migration lands.

## Cleanup

`src/internal/core/templates/navbar.templ` deleted. The `NavBarComponent` primitive has zero callers and represents an abandoned alternative that clutters the primitive surface.
