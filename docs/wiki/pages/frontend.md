---
description: >
  How the UI is built and rendered — the technology stack, interaction model, and rendering
  approach. Belongs here: templating, HTMX usage, CSS tooling, and the rationale for frontend
  decisions. Does not belong here: feature descriptions (→ features), server-side request
  handling (→ architecture), or product philosophy (→ vision).
links:
  - architecture
  - features
  - vision
  - roadmap
---

[wiki](../wiki.md)

# Frontend

How the UI is built and why.

## Approach

Wax uses **server-rendered HTML with targeted interactivity** via HTMX — a deliberate choice rooted in the [progressive enhancement](./vision.md) design philosophy. There is no JavaScript framework or client-side routing. The server renders complete HTML pages and partial fragments; HTMX swaps fragments in without full page reloads.

## Stack

| Tool | Role |
|---|---|
| **Templ** | Go-compiled HTML templates. Template files (`.templ`) compile to Go at build time — type-safe, no runtime parsing |
| **HTMX** | Declarative HTML attributes drive dynamic interactions (form submissions, partial updates) |
| **Tailwind CSS** | Utility-first styling |
| **DaisyUI** | Component layer on top of Tailwind (buttons, modals, cards, etc.) |

## Interaction Model

Forms and actions use HTMX attributes:
- `hx-post` / `hx-put` — submit data to the server
- `hx-swap` — control where the response HTML lands in the DOM
- Error states return HTML fragments rendered inline, not JSON error codes

This means most user interactions (rating an album, adding a tag, opening a review) result in a server round-trip that returns a small HTML fragment, not a full page reload and not a JSON payload that the client has to parse.

## Pages & Views

The primary view is the **library dashboard** — a sortable, browsable table/grid of the user's albums. From there users open individual album detail pages. See [features](./features.md) for the full breakdown of views and sort options.

## Known Limitations

The UI is not optimized for mobile web — layout, navigation, and interactions are designed primarily for desktop browsers. Mobile support is on the [roadmap](./roadmap.md).

## Static Assets

CSS is compiled at build time via the Tailwind CLI. JavaScript (HTMX + a small ticker script) is served as static files. No bundler or Node.js runtime is required at runtime.

