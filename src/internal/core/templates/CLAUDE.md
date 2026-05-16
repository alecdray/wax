# core/templates — UI primitives (singleton)

This directory is the home of UI **primitives** — reusable visual building blocks used by 2+ modules. Full rules: [`docs/design/archetypes/primitive.md`](../../../../docs/design/archetypes/primitive.md).

## Rules

- Files here are primitives: domain-free, parameterized by plain values, consumed by any number of pages and fragments.
- Do **not** import any domain module (`library`, `review`, `tags`, `notes`, etc.). A primitive that needs a domain type isn't a primitive — it belongs in the owning module's `adapters/`.
- Do **not** import `core/db/sqlc`.
- The root and shared-layout primitives are loaded by every page templ via `templates.PageLayoutComponent`. Anything that should appear on every page — chrome, fonts, scripts, modal container — lives here and is pulled in through the layout, not duplicated in pages.

## The Icon primitive

`icons.templ` defines the single `Icon` primitive, which wraps Bootstrap Icons. Pass a BI catalog name (without the `bi-` prefix) and an optional `IconStyle` (Outline | Fill). Sizing and color come from the parent (`text-{size}` for size, parent text color for color). The CSS that powers it is vendored under `static/public/`; the BI catalog lives at https://icons.getbootstrap.com/.

## After editing

- Run `task build/templ` after modifying any `.templ` file.
