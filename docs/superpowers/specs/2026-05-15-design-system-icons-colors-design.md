# Design System: Icons and Colors

**Date:** 2026-05-15
**Status:** Spec — pending implementation

## Why

The wax design docs (`docs/design/`) describe the templ archetype system thoroughly but leave the visual vocabulary thin. Two specific gaps:

- **Icons.** `core/templates/icons.templ` is ~19 hand-inlined SVGs that mix Bootstrap Icons and Heroicons sources. The `IconStyleOutline | IconStyleFill` enum exists, but only ~6 icons actually implement both styles; several (`UserIcon`, `NotesIcon`) declare the prop branch with identical SVG bodies. New icons are added by copy-pasting SVG paths and writing a new templ component.
- **Colors and opacity.** The wax theme (`static/src/main.css`) defines surfaces, brand tones, semantic colors, and `-content` pairs, but `design-system.md` doesn't describe when to use which. Opacity has accumulated organically: `text-base-content/{20,30,40,50,60,70,80}` all appear in templ files (~65 call sites total), and raw `opacity-{20,30,50,70,80}` is used inconsistently for disabled, hover, and boolean dimming states. There is no documented scale.

The wiki carries design intent (analog warmth, vintage spirit, navigation-icon outline/fill convention) that the design docs don't yet capture. The CLAUDE.md governing `docs/design/` requires that principles and design-system docs reflect what exists in code — which is why the doc updates land alongside code changes, not before.

This spec covers the visual vocabulary for icons and colors. It explicitly does not cover typography (already documented and stable), animation (purposeful and already documented), or layout primitives (separate spec).

## Visual direction

Add a brief opening paragraph to `design-system.md`:

> The visual direction is analog warmth: dark, warm-toned surfaces (browns rather than grays), amber-orange accents (the wax of a sealed record sleeve, the glow of warm light), light text on dark surfaces by default, and intentional motion where motion serves meaning. The styling stack and tokens below are how that direction is enacted.

Two or three sentences, named as the existing direction (not aspiration), so future contributors have an anchor for whether a proposed addition fits the system.

## Icons

### Source

**Bootstrap Icons** (MIT, ~2000 icons), self-hosted from `static/`. Wax already uses BI for most existing icons via inline SVG; this migration switches to the BI CSS font system so call sites name an icon rather than copy its paths.

The non-BI icons currently in the file (`NotesIcon`, `TrashIcon`, `QuestionMarkIcon`, `EllipsisVerticalIcon` — Heroicons-derived SVGs) are replaced with their BI equivalents.

### API

`core/templates/icons.templ` is replaced with a single primitive:

```go
package templates

type IconStyle int

const (
    IconStyleOutline IconStyle = iota // bi-{name}        — BI's default is outline
    IconStyleFill                     // bi-{name}-fill
)

type IconProps struct {
    Name  string    // BI name without prefix, e.g. "collection", "compass", "house"
    Style IconStyle // defaults to Outline (zero value)
}

templ Icon(props IconProps) {
    <i class={ "bi bi-" + props.Name + iconStyleSuffix(props.Style) }></i>
}
```

Sizing comes from the parent's `text-{xs,sm,base,lg}` (BI inherits font-size). Color comes from the parent's text color (BI inherits `currentColor`). The `Icon` primitive carries no size or color props — those concerns belong to the call site, the same as text.

### Outline / fill convention

Outline = navigable destination or default presentation; fill = current page or selected state. Applies wherever a UI surface has a paired notion of "this one vs the others" (today: nav header). Most usages are decorative or single-meaning; for those, leave `Style` at its default (outline).

### Single-variant icons

A handful of BI icons exist in only one style (e.g. `arrow-repeat`, `vinyl`, `cassette`). Rule: check BI's catalog before passing `Fill` for an icon name; if no `-fill` variant exists, omit the prop. This is a docs-level rule, not enforced — the cost of code-level enforcement (a generated allowlist) outweighs the rare miss.

### Delivery

The layout primitive adds a `<link>` to `bootstrap-icons.css`. The CSS file and `bootstrap-icons.woff2` font live in `static/` (vendored, not CDN — same reasoning as other static assets).

## Colors

### Surfaces — for backgrounds, borders, dividers

Stratify by elevation:

- `bg-base-100` — page background. The default canvas.
- `bg-base-200` — raised surface (cards, panels, dropdowns, modal bodies).
- `bg-base-300` — highest elevation (hovered rows, pressed states, the topmost layer in a stack); also the default border color (`border-base-300`).

### Brand tones — for emphasis, identity, decorative chrome

Each token has a defined role; reach for the role, not for the color that "looks right" in isolation.

- `primary` (warm amber/orange) — interactive emphasis. Links, brand wordmark, selected/active text states. Reserved — fewer than ~5 call sites today; that scarcity is the point.
- `accent` (lighter amber) — decorative highlights and "glow" moments. Brand flourishes, animated chrome (the ticker), incidental emphasis where the goal is warmth, not action.
- `secondary` (deep red-brown) — supporting actions and tags-domain affordances. A weighted-but-not-primary feel for tag chips/badges where brand presence helps the eye but `primary` would be too loud. The tag-edit button is the prototype use.
- `neutral` (dark brown) — chrome that isn't a surface and isn't a brand expression: tooltips, kbd hints, neutral badges. The "quiet" brand color.

Each tone has a paired `-content` token for legible text **on** that color (`text-primary-content` on `bg-primary`, etc.). Always use the pair; never put `text-base-content` on a brand background.

### Semantic — for status, only

- `info` — neutral informational status (sync state, "feed last updated").
- `success` — completed actions, positive validation.
- `warning` — recoverable problems, "this might surprise you" notices.
- `error` — failed actions, validation errors, **and destructive actions** (delete buttons, remove-tag, irreversible "are you sure" CTAs). Inline error components (per `principles.md`) lean on this.

### Text emphasis scale — four stops, named utilities

Every `*-content`-on-base color emphasis uses one of four named utilities. Each is defined in `main.css` as a Tailwind v4 `@utility` block so variants (`hover:text-muted`, `md:text-subtle`) work.

| Utility | Mechanism | Role | When to use |
|---|---|---|---|
| `text-default` | `var(--color-base-content)` | Default | Body copy, headings, primary values. The voice of the page. |
| `text-muted` | `base-content` at 70% | Muted | Section labels, captions, supporting meta-context. |
| `text-subtle` | `base-content` at 40% | Subtle | Timestamps, helper text, low-priority metadata. |
| `text-ghost` | `base-content` at 20% | Ghost | Placeholders, empty-state hints, dimmed icons (e.g. format-not-owned). |

Definitions:

```css
@utility text-default {
    color: var(--color-base-content);
}
@utility text-muted {
    color: color-mix(in oklab, var(--color-base-content) 70%, transparent);
}
@utility text-subtle {
    color: color-mix(in oklab, var(--color-base-content) 40%, transparent);
}
@utility text-ghost {
    color: color-mix(in oklab, var(--color-base-content) 20%, transparent);
}
```

Raw `text-base-content/NN` does not appear in templ markup outside these utilities.

**Brand-colored text** (`text-primary`, `text-error`, etc.) is a separate mechanism — it is *emphasis by color*, not emphasis by hierarchy. Brand colors do not get a four-stop scale. Existing patterns like `text-primary hover:text-primary/80` (a hover state using opacity-on-brand) and `text-error/50 hover:text-error/70` (a muted destructive look for delete affordances) are state transitions on a brand color, allowed but used sparingly. They are not affected by this section's rules and do not migrate to `.hover-fade` (which fades the whole element, not the color).

### Element opacity — two narrow roles, two utility classes

Raw `opacity-NN` on a whole element is reserved for two cases. Both are wrapped as utility classes in `main.css` so call sites name the role, not the mechanism.

**Disabled** — `.is-disabled`:

```css
.is-disabled {
    opacity: 0.5;
    cursor: not-allowed;
    pointer-events: none;
}
```

For non-form elements, use `class="is-disabled"`. For form controls (`<button>`, `<input>`), use the HTML `disabled` attribute and DaisyUI's existing disabled styles. Never strip one of the three properties — visually-disabled-but-clickable and non-clickable-but-full-opacity are both bugs.

**Hover affordance on a whole-element block** — `.hover-fade`:

```css
.hover-fade {
    transition: opacity 150ms;
}
.hover-fade:hover {
    opacity: 0.8;
}
```

For cards, link-wrapped media, decorative tiles where the entire surface is the click target. Don't layer `.hover-fade` onto buttons or interactive controls where DaisyUI provides a hover state.

**Not legitimate uses of raw opacity:**

- Text emphasis — use the four-stop text scale.
- Icon dimming — use the text scale (BI inherits `currentColor`, so `text-ghost` works).
- Boolean states (e.g. format owned vs not-owned) — use the text scale (`text-muted` vs `text-ghost`) so the dimming reads as content emphasis, not a half-disabled element.
- Hover on a button or interactive control where DaisyUI provides a hover state — let DaisyUI handle it.

## Migration

Three streams; A then B then C; three commits on one feature branch, reviewed as one PR.

### Stream A — Icons

1. Vendor `bootstrap-icons.css` and `fonts/bootstrap-icons.woff2` into `static/`. Add the `<link>` to the layout primitive.
2. Replace `core/templates/icons.templ` — delete the per-icon templ functions; add the single `Icon` primitive plus the `iconStyleSuffix` helper.
3. Walk the seven call-site files (`library/adapters/views/album_rating_history_frag.templ`, `library_header_bar_frag.templ`, `format_icon_frag.templ`, `feeds_dropdown_frag.templ`, `album_detail_page.templ`, `album_row_tags_section_frag.templ`, `review/adapters/views/rating_confirm_form_frag.templ`). Each `templates.CollectionIcon(...)` becomes `templates.Icon(IconProps{Name: "collection", Style: ...})`. Heroicons-derived icons (`NotesIcon`, `TrashIcon`, `QuestionMarkIcon`, `EllipsisVerticalIcon`) are replaced with their BI equivalents (`journal-text`, `trash`, `question-circle`, `three-dots-vertical`).
4. Run `task build/templ` after each templ edit batch.

### Stream B — Colors

1. Add `.is-disabled`, `.hover-fade`, and the four `@utility` text-emphasis blocks to `main.css`.
2. Walk every `text-base-content/{NN}` call site. Map to the nearest of the four named utilities. **Where the new role label disagrees with the element's purpose** (e.g. an `<h1>` collapsing to `text-subtle`), flag the site for human review rather than applying mechanically — those are probable misuses, not migration mismatches.
3. Walk every raw `opacity-{NN}` and `hover:opacity-NN` site:
    - The active-nav `opacity-30 + pointer-events-none` collapses to `class="is-disabled"`.
    - The format-icon `opacity-70`/`opacity-20` boolean moves to `text-muted`/`text-ghost`.
    - Existing `hover:opacity-80 transition-opacity` sites collapse to `class="hover-fade"`.
    - Sites where DaisyUI already handles the hover state lose the redundant opacity.
4. Land first uses of `accent` (likely the brand wordmark or ticker animation — pick one site in this PR so the token isn't documented-but-unused). Confirm `secondary`'s tag-edit usage is consistent with the documented role.

### Stream C — Documentation

Updates land after the code, in a single commit. Documentation splits along the project's existing rule (per `docs/design/CLAUDE.md`): **conceptual content lives in `docs/design/`; verbatim implementation lives in a `CLAUDE.md` near the source.**

**`docs/design/design-system.md` — conceptual additions only:**

- Opening paragraph (visual direction).
- New "Icons" section: source (Bootstrap Icons), the outline/fill convention as a design rule, the single-variant rule. Names the `Icon` primitive and points to `core/templates/` for the signature; does not reproduce Go code.
- New "Colors" sections: surfaces (with role descriptions), brand tones (with role descriptions), semantic with destructive-actions note. Names the four text-emphasis utilities (`text-default`, `text-muted`, `text-subtle`, `text-ghost`) and the two element-state utilities (`.is-disabled`, `.hover-fade`) and points to `static/src/` for the definitions; does not reproduce CSS.
- The "When to add to `main.css`" rule gets a fourth bullet: *"A named-role wrapper around theme-semantic atoms is needed (e.g. text emphasis utilities, state utilities like `.is-disabled`). Bare single-class wraps (`text-amber` for `text-primary`) are still not added."*

**`docs/design/principles.md` — cross-cutting rule addendum:**

The "Theme tokens, not raw colors" principle gets the line: *"The text emphasis scale is `text-default`, `text-muted`, `text-subtle`, `text-ghost`; the element-state utilities are `.is-disabled` and `.hover-fade`. Raw `text-base-content/NN` and raw `opacity-NN` should not appear in templ markup outside these utilities."* No verbatim CSS — principles describe rules, not definitions.

**`static/src/CLAUDE.md` — new file, verbatim catalog of wax-specific additions to the Tailwind+DaisyUI stack:**

This is the implementation-detail home for everything in `main.css` that isn't a stock DaisyUI token. Contents:

- A short header (one paragraph) explaining what `main.css` is and the directory's role.
- A "Theme tokens" section listing the wax theme block (link to or reproduce the `[data-theme="wax"]` definition).
- A "Text emphasis utilities" section with the four `@utility` blocks (verbatim CSS).
- An "Element-state utilities" section with `.is-disabled` and `.hover-fade` (verbatim CSS).
- A "Bootstrap Icons" section noting that `bootstrap-icons.css` and `fonts/bootstrap-icons.woff2` are vendored here, loaded by the layout primitive.
- A "Custom keyframes" section noting `ticker-scroll` (already there) and the rule that new keyframes get a brief note when added.
- The hard rule: *"per-page or per-component styling does not go in `main.css` — templ files use Tailwind utilities directly."*

**`src/internal/core/templates/CLAUDE.md` — update existing file:**

Add a short note: *"The `Icon` primitive (in `icons.templ`) wraps Bootstrap Icons. Pass a BI catalog name (without the `bi-` prefix) and an `IconStyle`; see `icons.templ` for the signature. The CSS that powers it is vendored under `static/src/`."*

This split keeps `design-system.md` doing what its CLAUDE.md says it should — describing intent, naming conventions — and pushes the verbatim CSS, Go signatures, and class-by-class catalog to documents that live next to the source they describe.

## Out of scope

- Typography expansions (the existing `font-brand` rule is sufficient for current needs).
- Animation guidelines beyond the existing "purposeful motion" rule.
- DaisyUI component-class conventions (`btn`, `card`, `badge`) — DaisyUI's own docs cover these; wax doesn't override.
- Light-mode theme. Wax is dark-only by deliberate choice.
- Icon size standardization beyond "use the parent's font-size." A future spec can add named size utilities if the codebase converges on a small set.
