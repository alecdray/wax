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
  - testing
---

[Parent: wiki](../wiki.md)

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

The primary view is the **library dashboard** — a sortable, browsable table/grid of the user's albums. Clicking an album title navigates to the **album detail page**, a dedicated per-album view. See [features](./features.md) for the full breakdown of views and behaviour.

Navigation icons use a filled variant to indicate the current page and an outline variant for navigable destinations.

## Responsive Design

The UI supports both desktop and mobile. On mobile, the dashboard uses natural document scroll with a sticky header. The album table collapses to its most essential columns, with secondary information accessible via the per-row actions menu. Desktop retains a fixed-height inner-scroll layout.

The library dashboard table is desktop-optimised. The album detail page uses a mobile-first stacked layout and is the primary interaction surface on small screens.

## Visual Design

### Theme

Wax uses a custom DaisyUI theme called **"wax"** — a dark colour scheme built around warm, earthy tones. The palette is organised around:

- **Base surfaces**: Three increasingly lighter near-black browns, creating depth between background layers
- **Content**: A warm off-white for default text
- **Primary**: A warm amber/orange — the main brand colour for interactive elements
- **Secondary**: A deep red-brown for supporting accents
- **Accent**: A lighter amber, used for highlights and decorative elements
- **Semantic colours**: Standard info/success/warning/error slots, all warm-tinted to stay coherent with the palette

Light text on dark surfaces is the default. Elements that use the neutral slot (e.g. tooltips) use a dark brown background with light cream text.

### Typography

The brand wordmark uses **Instrument Sans** — a bold, condensed sans-serif. Body text and UI elements use the system/Tailwind default sans-serif stack.

## Static Assets

CSS is compiled at build time via the Tailwind CLI. JavaScript (HTMX + a small ticker script) is served as static files. No bundler or Node.js runtime is required at runtime.

