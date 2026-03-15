---
description: >
  Significant decisions that changed the direction of the product or architecture. Records
  what the old approach was, what replaced it, and why. Does not belong here: current
  architecture or feature descriptions (see architecture.md, features.md); planned changes
  (see roadmap.md).
links:
  - wiki
  - architecture
  - features
---

[Parent: wiki](../wiki.md)

# Decisions

A log of significant changes to product or architecture — what we did before, what we changed to, and why. Ordered most-recent first.

Entries belong here when a destructive wiki edit would otherwise lose context that is useful for understanding *why* the current state looks the way it does.

---

<!-- Add entries below in this format:

## <Short title of the decision>
**Date:** YYYY-MM-DD
**Was:** Brief description of the old approach.
**Now:** Brief description of the new approach.
**Why:** The reason for the change.

-->

## My Library: table view replaced by visual list with chip-based filtering
**Date:** 2026-03-14
**Was:** The library was a table with sortable column headers (title, artist, rating, date added, last played), a rating badge/button per row, and tags accessible via an ellipsis (⋯) dropdown. Album titles linked to the detail page; Spotify was accessible via icon links on each row.
**Now:** The library is a visual list. Each row leads with a format icon column and album art, followed by title and artist, with a large numeric rating at the right. Sorting and filtering (rating range, format, artist, rated/unrated) are controlled by a chip bar above the list. No Spotify links appear in the list; Spotify access moves exclusively to the album detail page. Tags are shown as read-only badges in the row footer.
**Why:** The table format was dense and spreadsheet-like. A visual list leads with the cover art, evoking physical media. The chip bar enables simultaneous multi-facet filtering, which column-header sort could not support. Spotify outlinks were removed from the dashboard to keep navigation rooted in Wax.

