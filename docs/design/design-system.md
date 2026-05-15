# Design System

The visual vocabulary used across pages, fragments, and primitives. This doc is the **conceptual** layer over `static/src/main.css` — it describes the categories, conventions, and intent. `main.css` is the source of truth for exact values; this doc explains how they're used.

## Foundation

The styling stack is **Tailwind CSS + DaisyUI**, with a custom DaisyUI theme named `wax` declared in `static/src/main.css`. The theme defines semantic color tokens (base / primary / secondary / accent / neutral / info / success / warning / error, each paired with a `*-content` variant for legible text on that background), corner radii (box / button / badge), and the dark color scheme.

Use semantic token names in markup — never hex literals, never `--color-*` variables directly. Tokens are referenced through Tailwind utility classes (`bg-base-100`, `text-primary-content`, `border-accent`) and DaisyUI component classes (`btn`, `card`, `badge`).

## Typography

Body and UI text use DaisyUI's default stack. The brand mark uses a custom `.font-brand` utility (Instrument Sans) defined in `main.css`. Add a font utility only when a new typographic role recurs across multiple surfaces; one-off styling stays inline at the call site.

## Animations

Custom keyframes live in `main.css` and are applied via Tailwind `animate-*` utilities or inline `style`. Animations exist for purposeful motion (the ticker, transitions); the codebase does not animate for its own sake.

## When to add to `main.css`

`main.css` is small on purpose. Add to it when:

- **A new theme token is needed.** Extend the `[data-theme="wax"]` block. Use a semantic name (`--color-success-muted`), not a literal one (`--color-green-light`).
- **A reusable utility is needed.** Define it at file scope, like `.font-brand`. Do not add utilities that wrap a single Tailwind class — use the Tailwind class directly.
- **A new keyframe is needed.** Define it at file scope; consume it via Tailwind's `animate-*` or inline.

Do **not** add per-page or per-component styling to `main.css`. Templ files use Tailwind utilities directly.

## Client-side libraries

The root primitive loads three external libraries: **HTMX** for interaction, **idiomorph** for smart DOM swaps, and **Alpine.js** for ephemeral client state. Use them in that order of preference — HTMX first, idiomorph as a swap strategy when needed, Alpine only for state that genuinely lives on the client (open/closed UI, focus rings, hover affordances).

A fourth library is added only when none of the three can do the job. Adding one means updating the root primitive and documenting the addition here.
