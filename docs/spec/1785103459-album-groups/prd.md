# Crates — Mini-PRD

## Target Repo
`wax`

## Problem / Need
Users want to organise albums into named personal collections — "jazz," "road trip music," "gifts for friends" — where the collection itself is browsable and meaningful. The library's current organisation primitives (format filters, sort, tags as inline labels) don't support this. Tags are too lightweight; ranked lists are ordered by design. There is no first-class "named crate with its own page" concept.

## Product-Level Solution Concept
A **crate** is a named, unordered collection of albums. Users can create crates, name them, add or remove albums from them, and navigate to a crate's own page to see its members. Crates get a dedicated tab in the bottom nav.

## MVP
The minimum to ship a working Crates feature. Everything lives within the Crates surface — no changes to album detail.

- Create a crate (from the Crates tab)
- Delete a crate
- Add albums to a crate (from within the crate detail page)
- Remove an album from a crate
- Crate detail page — member albums rendered as library rows (reused)
- Crates index — name + member count
- Bottom nav tab

## v1
Enhancements once the core works.

- Add to crate from album detail page (picker modal showing all crates)
- Rename a crate
- Cover art mosaic on index cards (2×2 grid of album covers)

## Out of Scope
- Ordering albums within a crate (that's Ranked Lists)
- Sharing crates with other users
- Crate descriptions or cover images
- Stats or analytics across crates

## Hard Constraints
- Crates are unordered — no position metadata
- Distinct from tags — crates are browsable objects, tags are inline labels; the two are not merged

## Soft Constraints
- An album may belong to multiple crates
- Crate names need not be unique
- Entry points should feel natural on mobile (primary surface)

## Decisions
- **Navigation:** dedicated tab in the bottom nav (alongside Library, Radar, etc.)
- **Product name:** "Crates" (fits the vinyl/analog aesthetic)
- **Crate detail rows:** reuse existing library album rows — no new row component needed
- **Modals for mutations:** create crate and add-album flows are modals triggered from within the Crates surface
- **Delete:** available only from the crate detail view; requires a confirmation step (not a modal — native confirm); not surfaced elsewhere
