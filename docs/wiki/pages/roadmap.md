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
| **Ranklists** | User-curated ranked lists of albums |
| **Shelves** | Organize albums into named custom shelves |
| **Stats & Insights** | Analytics across library, listening history, ratings, ranklists, and shelves |
| **Notifications** | In-app notifications for events (sync, activity) |
| **Wishlist** | Track albums you want but don't own yet |
| **Sleeve Notes** | Attach free-form notes to any library entity |
| **Linked Albums** | Connect albums to each other, building a personal music graph |
| **[My Library v1.1](./my-library-v1.1.md)** | Rework the library table into a list view with chip-based faceted filtering and sorting; dashboard UX cleanup: remove all Spotify outlinks from the dashboard, make album image and artist name clicks open the album detail page |
| **Library Search** | Search/filter box on the dashboard to find albums in the library by title or artist |
| **Physical Media** | Support for vinyl, CD, and cassette ownership; manual add flow with Discogs/MusicBrainz lookup; format facet in library filters |
| **Hidden Albums** | Soft-remove albums from the main library view without deleting them (e.g., podcasts or junk synced from Spotify) |
| **Album Detail — Non-Library Albums** | Support a read-only detail view for albums not in the user's library; fixes Recently Spun carousel items that currently 404 or fall back to Spotify outlinks |
| **Auth Error Handling** | Graceful handling of JWT middleware failures and expired/invalid Spotify token failures |

## Ideas & Open Questions

- **Stats & Insights visualizations** — listening heatmap (GitHub-style activity grid by day/month), genre evolution timeline showing how tastes shifted year over year, top artists by decade, "record DNA" radar chart showing where a library skews across tempo/energy/mood/era
- **Progressive Web App (PWA)** — open question: whether to convert Wax to a PWA for offline support and installability; deferred until the mobile experience is more fully developed
- **Comparative ranking** — derive a rating by pitting an album against others the user has already rated; a series of "is this better than that?" questions produces a score grounded in relative preference rather than an abstract 0–10 pick
- **Dual-axis rating** — separate scores for objective quality vs. personal enjoyment ("it's a masterpiece but I never play it")
- **Timestamped reviews** — reviews as journal entries to track how opinions evolve
- **Linked Albums as a graph** — similar to Obsidian's graph view, surface connections between records
- **Tags → Ranklists** — each tag automatically generates a ranked list of tagged albums
- **Tag management** — a dedicated interface for managing tags at the user level; create, rename, merge, and delete tags without having to navigate through individual albums
- **Album Detail — external sources** — links to Pitchfork, Wikipedia, NPR, and YouTube per album; eventual goal is a rich album detail page that aggregates critical context, video, and background alongside the user's own library data; users should also be able to manually attach their own resource links (live performances, Tiny Desk concerts, interviews, articles, reviews) to any album
- **Influences** — surface what influenced an album and what it influenced
- **Shared by / shared with** — optional field to track who introduced you to a record
- **Social features** — Goodreads-style network, but secondary to personal library depth
- **Last.fm integration** — extended listening history, working around Spotify's 50-track limit (see [integrations](./integrations.md))
- **Multiple view modes** — grid/cover wall, compact text-only, and table alongside the default list view; mode switcher in the library header; persisted per user

