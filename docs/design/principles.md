# Design Principles

Cross-cutting rules that apply to every archetype. These describe what is already true of the codebase — a principle lands here when it lands in the code, not before.

## Server renders HTML; HTMX drives interaction

The interaction model is server-rendered HTML over HTMX, not client-rendered components. Forms submit with `hx-post` / `hx-put` and the server responds with an HTML fragment. Page navigation uses `hx-boost` on the body so links morph the relevant region instead of full-page reloading. JavaScript is reserved for genuinely client-only state (Alpine `x-data` for ephemeral UI state); it is not the medium for fetching, validating, or transforming domain data.

## Fragments over pages

When only a slice of the UI needs to change, the handler returns a fragment, not a full page. A fragment swaps into the existing DOM; a full page re-renders everything. The fragment is the reusable unit (see [archetypes/fragment-templ.md](archetypes/fragment-templ.md)); a page templ wraps a fragment in the shared layout when the same content is also reachable directly by URL.

## Errors render inline, in place

When a request fails in a way the user can recover from, the server returns an error component scoped to the relevant region — not a redirect, not a global alert, not a banner on the next page. The error appears where the action was taken. The mechanism is `httpx.HandleErrorResponse` plus an error templ component sized to the swap target.

## DOM ids belong to the templ that owns the region

If a templ defines a region that HTMX targets by id, the id-generating helper lives next to that templ (same file or its `.go` sibling). Callers obtain the id by calling the helper, not by hard-coding the string. The templ stays the single source of truth for what its swap target is named.

## Theme tokens, not raw colors

Styling uses the DaisyUI theme tokens defined in `static/src/main.css` (`bg-base-100`, `text-primary-content`, `border-accent`, etc.), not hex literals or one-off CSS variables in markup. When a new color is needed, it is added to the theme as a semantic token, not embedded inline at the call site.
