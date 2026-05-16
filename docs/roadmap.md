# Roadmap

Planned features in rough priority order. See each module's README for what's already shipped, and [`docs/backlog.md`](./backlog.md) for operational follow-ups.

## In queue

| Feature | Summary |
|---|---|
| **Ranklists** | User-curated ranked lists of albums |
| **Shelves** | Organize albums into named custom shelves |
| **Stats & Insights** | Analytics across library, listening history, ratings, ranklists, and shelves |
| **Notifications** | In-app notifications for events (sync, activity) |
| **Wishlist surfaces** | Wishlist status exists in the data model; dedicated UI surfaces (a wishlist view, add-to-wishlist affordances on discover) need to land |
| **Linked Albums** | Connect albums to each other, building a personal music graph |
| **Library Search** | Search/filter box on the dashboard to find albums in the library by title or artist |
| **Filter/Sort UX polish** | The chip-based filter and sort UI is functional but visually rough — dialog styling, chip bar layout, and interaction patterns need iteration |
| **Hidden Albums** | Soft-remove albums from the main library view without deleting them (e.g., podcasts or junk synced from Spotify) |
| **Album Detail — Non-Library Albums** | A read-only detail view for albums not in the user's library; fixes carousel items and discover results that currently fall back to Spotify outlinks |
| **Auth Error Handling** | Graceful handling of JWT middleware failures and expired/invalid Spotify token failures |
| **Rating Label Sync (Album List)** | When a rating is changed via the rating modal from the album list, the rating label badge on the row should update without a page reload |
| **Saved Tracks Sync** | Sync the user's saved tracks from Spotify so loved tracks can be highlighted within album views |

## Ideas

Directionally clear, not yet queued.

- **Stats & Insights visualizations** — listening heatmap (GitHub-style activity grid), genre evolution timeline, top artists by decade, "record DNA" radar chart of tempo/energy/mood/era.
- **Comparative ranking** — derive a rating by pitting an album against others already rated; a series of "is this better than that?" questions produces a score grounded in relative preference.
- **Dual-axis rating** — separate scores for objective quality vs. personal enjoyment ("it's a masterpiece but I never play it").
- **Timestamped reviews** — reviews as journal entries to track how opinions evolve.
- **Linked Albums as a graph** — Obsidian-style graph view of connections between records.
- **Tags → Ranklists** — each tag automatically generates a ranked list of tagged albums.
- **Tag management** — dedicated interface for managing tags at the user level (create, rename, merge, delete) without going through individual albums.
- **Album Detail — external sources** — links to Pitchfork, Wikipedia, NPR, YouTube per album; users should be able to attach their own resource links (live performances, Tiny Desk concerts, interviews, articles).
- **Influences** — surface what influenced an album and what it influenced.
- **Shared by / shared with** — optional field to track who introduced you to a record.
- **Last.fm integration** — extended listening history, working around Spotify's 50-track recently-played limit.
- **Multiple view modes** — grid/cover wall, compact text-only, table — mode switcher in the library header, persisted per user.

## Open questions

Genuinely undecided — direction or approach unresolved.

- **Progressive Web App (PWA)** — whether to convert Wax to a PWA for offline support and installability; deferred until the mobile experience is more fully developed.
- **Social features** — Goodreads-style network is a natural long-term direction, but it's unclear whether this belongs in the core product or as a separate surface; secondary to personal library depth.
