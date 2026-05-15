# core/templates — UI primitives (singleton)

This directory is the home of UI **primitives** — reusable visual building blocks used by 2+ modules. Full rules: [`docs/design/archetypes/primitive.md`](../../../../docs/design/archetypes/primitive.md).

## Rules

- Files here are primitives: domain-free, parameterized by plain values, consumed by any number of pages and fragments.
- Do **not** import any domain module (`library`, `review`, `tags`, `notes`, etc.). A primitive that needs a domain type isn't a primitive — it belongs in the owning module's `adapters/`.
- Do **not** import `core/db/sqlc`.
- The root and shared-layout primitives are loaded by every page templ via `templates.PageLayoutComponent`. Anything that should appear on every page — chrome, fonts, scripts, modal container — lives here and is pulled in through the layout, not duplicated in pages.

## After editing

- Run `task build/templ` after modifying any `.templ` file.
