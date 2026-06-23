# Testids

Every templ component's top-level root element carries a `data-testid`. Testids are how Playwright tests, HTMX `hx-target="closest [data-testid='...']"` selectors, and ad-hoc DOM tooling locate elements without depending on Tailwind classes that change with styling.

## Naming

```
data-testid="<component>[-<postfix>]"
```

- **`<component>`** — kebab-case of the templ function name, with the `Frag` suffix dropped if present. `AlbumScoreBadgeFrag` → `album-score-badge`. `albumListRow` (private templ) → `album-list-row`.
- **`<postfix>`** — added only when needed to disambiguate. Required when:
  - The single root is rendered by different `if`/`else`/`switch` branches and each branch is a meaningfully different variant (`-rated` vs `-unrated`, `-empty` vs the populated form).
  - The same conceptual component appears with materially different variants (`album-card-in-library` vs `album-card-new`).
- A component with one unambiguous root takes the base name alone — no `-main`, no `-root`, no filler.

The postfix describes the **variant** of the root (which branch, which state) — `-empty`, `-rated`, `-unrated`, `-in-library`, `-new`. It is not the role of a sibling; siblings live under a wrapper (see "One root per component" below).

## Non-root elements

Descendants that need their own testid follow the same pattern, prefixed by the containing component:

```
data-testid="<component>-<role>"
```

`AlbumDetailPage` has a title heading inside it → `album-detail-page-title`. A submit button in `BaseQuestionsFormFrag` → `base-questions-form-submit`. The role names what the element does within the component; it is not derived from a separate component name.

When a sub-fragment is composed into **exactly one parent** and exists to serve that parent, "containing component" means the parent: the fragment's testid takes the parent's prefix, not its own. A `FormatsReleasesFrag` that lives only inside `AlbumDetailPage` declares `album-detail-page-releases`, not `formats-releases`. The grep-the-codebase rule still applies — a testid's prefix doesn't always point to its declaring file.

## One root per component

A non-OOB templ component renders exactly one top-level root. If a component would emit several always-rendered siblings (a header next to a form, a heading next to a list), wrap them in a `<div>` and let the wrapper carry the testid — that wrapper is also the natural target for HTMX swaps, layout classes, and Alpine scopes, so the constraint pays for itself.

Two narrow exceptions:

- **Pure delegation** — a component that just calls into another templ (`@templates.Modal(...)`, `@templates.ForceCloseModal(...)`) doesn't get a testid; the testid belongs to whatever component actually owns the rendered root.
- **List emitters** — a component that emits a homogeneous list (e.g. a `for` loop of `<li>` items with no enclosing `<ul>`, where the caller supplies the wrapper) doesn't invent a wrapper just to host a testid; each item carries its own testid if needed.

Conditional branches (`if`/`else`/`switch` where exactly one root renders) are not multi-root — they are one root that varies by branch, and each branch gets its own variant postfix.

## Selected state

Active / selected state — the current nav tab, a chosen option — is marked semantically with `aria-current` (`page` for navigation, `true` otherwise), not with a `-active` testid variant. The element keeps one stable testid across both states and tests assert the state through the attribute. The branch-variant rule above is for genuinely different rendered roots, not for one element toggling selected.

## Out-of-band swap targets

OOB swap fragments don't define their own HTML — they compose a shared region templ. The testid lives on that region in exactly one place, and is inherited by both the initial render and the OOB swap. See [oob-swaps.md](oob-swaps.md).

## Testids are not runtime selectors

`hx-target`, Alpine `x-ref` lookups, and other runtime selectors target the DOM by `id`, not by `data-testid`. Ids are the source of truth for what a region is named (see the "DOM ids belong to the templ that owns the region" principle); testids are an orthogonal channel for tests and debugging. Coupling runtime behavior to testids couples those two concerns and breaks any test that renames its own selector.

If `hx-target="closest [data-testid='...']"` would be the natural expression, give the target element an `id` (via a helper next to the templ) and use `hx-target="closest #..."` or `hx-target="#..."` instead.

## Examples

```templ
// Single root, no postfix needed
templ AlbumScoreBadgeFrag(album library.AlbumDTO) {
  <div data-testid="album-score-badge" class="...">
    ...
  </div>
}

// Multiple roots via branching — each branch gets a postfix
templ AlbumScoreReadoutFrag(album library.AlbumDTO) {
  if album.Rating != nil {
    <div data-testid="album-score-readout-rated" class="...">...</div>
  } else {
    <div data-testid="album-score-readout-unrated" class="...">...</div>
  }
}

// Descendants carry the component prefix
templ BaseQuestionsFormFrag(...) {
  <form data-testid="base-questions-form" ...>
    ...
    <button data-testid="base-questions-form-submit" type="submit">Next</button>
  </form>
}
```
