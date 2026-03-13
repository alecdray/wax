---
description: >
  Where the product is going — queued features, future versions of shipped features, and open
  ideas. Belongs here: unbuilt features, enhancement plans, open product questions, and
  speculative directions. Does not belong here: descriptions of what already works (→ features),
  or technical implementation plans (→ architecture).
links:
  - vision
  - features
  - integrations
---

[wiki](../wiki.md)

# Roadmap

Planned features in rough priority order. See [features](./features.md) for what's already shipped.

## In Queue

| Feature | Summary |
|---|---|
| **Album Detail Page** | Dedicated per-album page showing all library info, Spotify link, rating and review, and notes; replaces the table row as the primary way to interact with an album and improves mobile usability |
| **Ranklists** | User-curated ranked lists of albums |
| **Shelves** | Organize albums into named custom shelves |
| **Stats & Insights** | Analytics across library, listening history, ratings, ranklists, and shelves |
| **Notifications** | In-app notifications for events (sync, activity) |
| **Wishlist** | Track albums you want but don't own yet |
| **Sleeve Notes** | Attach free-form notes to any library entity |
| **Linked Albums** | Connect albums to each other, building a personal music graph |
| **My Library v1.1** | Faceted search, physical media, multiple view modes |

## Ideas & Open Questions

- **Stats & Insights visualizations** — listening heatmap (GitHub-style activity grid by day/month), genre evolution timeline showing how tastes shifted year over year, top artists by decade, "record DNA" radar chart showing where a library skews across tempo/energy/mood/era
- **Progressive Web App (PWA)** — open question: whether to convert Wax to a PWA for offline support and installability; deferred until the mobile experience is more fully developed
- **Rating history** — instead of a single current score, append each rating update as a timestamped entry so the full arc of how an opinion changed is preserved; each entry can carry a note explaining why the rating changed
- **Comparative ranking** — derive a rating by pitting an album against others the user has already rated; a series of "is this better than that?" questions produces a score grounded in relative preference rather than an abstract 0–10 pick
- **Dual-axis rating** — separate scores for objective quality vs. personal enjoyment ("it's a masterpiece but I never play it")
- **Timestamped reviews** — reviews as journal entries to track how opinions evolve
- **Linked Albums as a graph** — similar to Obsidian's graph view, surface connections between records
- **Tags → Ranklists** — each tag automatically generates a ranked list of tagged albums
- **Album Detail — external sources** — links to Pitchfork, Wikipedia, NPR, and YouTube per album; eventual goal is a rich album detail page that aggregates critical context, video, and background alongside the user's own library data
- **Influences** — surface what influenced an album and what it influenced
- **Shared by / shared with** — optional field to track who introduced you to a record
- **Social features** — Goodreads-style network, but secondary to personal library depth
- **Last.fm integration** — extended listening history, working around Spotify's 50-track limit (see [integrations](./integrations.md))

