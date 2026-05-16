# Design System

The visual vocabulary used across pages, fragments, and primitives. This doc is the **conceptual** layer over `static/src/main.css` — it describes the categories, conventions, and intent. `main.css` is the source of truth for exact values; this doc explains how they're used.

The visual direction is analog warmth: dark, warm-toned surfaces (browns rather than grays), amber-orange accents (the wax of a sealed record sleeve, the glow of warm light), light text on dark surfaces by default, and intentional motion where motion serves meaning. The tokens and utilities below are how that direction is enacted.

## Foundation

The styling stack is **Tailwind CSS + DaisyUI**, with a custom DaisyUI theme named `wax` declared in `static/src/main.css`. The theme defines semantic color tokens (base / primary / secondary / accent / neutral / info / success / warning / error, each paired with a `*-content` variant for legible text on that background), corner radii (box / button / badge), and the dark color scheme.

Use semantic token names in markup — never hex literals, never `--color-*` variables directly. Tokens are referenced through Tailwind utility classes (`bg-base-100`, `text-primary-content`, `border-accent`) and DaisyUI component classes (`btn`, `card`, `badge`).

## Icons

Wax uses **Bootstrap Icons** (MIT, ~2000 icons) as its single icon source. The vendored CSS and font live under `static/`; the layout primitive loads them. All icons in the app are emitted by the single `Icon` primitive in `core/templates/icons.templ` — call sites pass the BI catalog name (without the `bi-` prefix) and an optional `IconStyle` (Outline or Fill). Sizing comes from the parent's `text-{size}`; color comes from the parent's text color (BI inherits `currentColor`).

**Outline / fill convention.** Outline is the default presentation; Fill marks the current page or selected state. Used wherever a UI surface has a paired notion of "this one vs the others" (today: the nav header). Most icons are decorative or single-meaning — for those, leave `Style` at its default.

**Single-variant icons.** Some BI icons exist in only one style (`vinyl`, `cassette`, `arrow-repeat`, etc.). Check BI's catalog before passing `Fill` for an icon name; if no `-fill` variant exists, omit the prop.

See `core/templates/icons.templ` for the primitive's signature.

## Colors

The wax theme defines three groups of tokens, each with a distinct role.

### Surfaces (backgrounds, borders, dividers)

Stratified by elevation:
- `bg-base-100` — page background. The default canvas.
- `bg-base-200` — raised surface (cards, panels, dropdowns, modal bodies).
- `bg-base-300` — highest elevation (hovered rows, pressed states, the topmost layer); also the default border color (`border-base-300`).

### Brand tones (emphasis, identity, decorative chrome)

Each token has a defined role; reach for the role, not the color that "looks right" in isolation.

- **`primary`** — interactive emphasis. Links, brand wordmark, selected/active text states. Reserved; scarcity is the point.
- **`accent`** — decorative highlights and "glow" moments. Brand flourishes, animated chrome.
- **`secondary`** — supporting actions and tags-domain affordances. Weighted-but-not-primary.
- **`neutral`** — chrome that isn't a surface and isn't a brand expression: tooltips, kbd hints, neutral badges.

Each tone has a paired `-content` token for legible text **on** that color. Always use the pair; never put `text-base-content` on a brand background.

### Semantic (status only)

- `info` — neutral informational status.
- `success` — completed actions, positive validation.
- `warning` — recoverable problems.
- `error` — failed actions, validation errors, **and destructive actions** (delete buttons, remove-tag, irreversible CTAs).

### Text emphasis scale

Four named utilities express the text hierarchy. Use the named utility, not raw `text-base-content/NN`.

| Utility | Role |
|---|---|
| `text-default` | Body copy, headings, primary values — the voice of the page. |
| `text-muted` | Section labels, captions, supporting meta-context. |
| `text-subtle` | Timestamps, helper text, low-priority metadata. |
| `text-ghost` | Placeholders, empty-state hints, dimmed icons. |

Brand-colored text (`text-primary`, `text-error`, etc.) is a separate mechanism — emphasis by color, not by hierarchy. Brand colors don't get the four-stop scale.

### Element opacity

Two narrow roles, each wrapped as a utility class. Raw `opacity-NN` on a whole element should not appear in templ markup outside these.

- `.is-disabled` — disabled state (whole element non-interactive). Pairs the visual dim with a not-allowed cursor and suppressed pointer events so they can't drift apart.
- `.hover-fade-out` — hover affordance on a whole-element block (cards, link-wrapped media). Visible → subtly dimmed on hover. Don't layer onto buttons or controls where DaisyUI handles the hover.
- `.hover-fade-in` — reveal-on-hover for secondary affordances (small buttons, ✕ controls on chips, row-scoped actions). Dim → fully visible on hover. Use when an action should be present but de-emphasized at rest.

See `static/CLAUDE.md` for the verbatim CSS definitions.

## Typography

Body and UI text use DaisyUI's default stack. The brand mark uses a custom `.font-brand` utility (Instrument Sans) defined in `main.css`. Add a font utility only when a new typographic role recurs across multiple surfaces; one-off styling stays inline at the call site.

## Animations

Custom keyframes live in `main.css` and are applied via Tailwind `animate-*` utilities or inline `style`. Animations exist for purposeful motion (the ticker, transitions); the codebase does not animate for its own sake.

## When to add to `main.css`

`main.css` is small on purpose. Add to it when:

- **A new theme token is needed.** Extend the `[data-theme="wax"]` block. Use a semantic name (`--color-success-muted`), not a literal one (`--color-green-light`).
- **A reusable utility is needed.** Define it at file scope, like `.font-brand`. Do not add utilities that wrap a single Tailwind class — use the Tailwind class directly.
- **A new keyframe is needed.** Define it at file scope; consume it via Tailwind's `animate-*` or inline.
- **A named-role wrapper around theme-semantic atoms is needed.** E.g. text emphasis utilities, state utilities like `.is-disabled`. Bare single-class wraps (`text-amber` for `text-primary`) are still not added.

Do **not** add per-page or per-component styling to `main.css`. Templ files use Tailwind utilities directly.

## Client-side libraries

The root primitive loads three external libraries: **HTMX** for interaction, **idiomorph** for smart DOM swaps, and **Alpine.js** for ephemeral client state. Use them in that order of preference — HTMX first, idiomorph as a swap strategy when needed, Alpine only for state that genuinely lives on the client (open/closed UI, focus rings, hover affordances).

A fourth library is added only when none of the three can do the job. Adding one means updating the root primitive and documenting the addition here.

The layout also loads **Bootstrap Icons** as a third-party stylesheet for icon rendering — see the Icons section above.
