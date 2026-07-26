# Features

Candidate features and improvements in rough priority order. Each ships through the [development process](../process.md).

## In queue

| Feature | Summary |
|---|---|
| **Ranklists** | User-curated ranked lists of albums |
| **Shelves** | Organize albums into named custom shelves |
| **Stats & Insights** | Analytics across library, listening history, ratings, ranklists, and shelves |
| **Notifications** | In-app notifications for events (sync, activity) |
| **Wishlist surfaces** | Wishlist status exists in the data model; dedicated UI surfaces (a wishlist view, add-to-wishlist affordances on the radar page) need to land |
| **Linked Albums** | Connect albums to each other, building a personal music graph |
| **Library Search** | Search/filter box on the dashboard to find albums in the library by title or artist |
| **Filter/Sort UX polish** | The chip-based filter and sort UI is functional but visually rough — dialog styling, chip bar layout, and interaction patterns need iteration |
| **Hidden Albums** | Soft-remove albums from the main library view without deleting them (e.g., podcasts or junk synced from Spotify) |
| **Album Detail — Non-Library Albums** | A read-only detail view for albums not in the user's library; fixes carousel items and discover results that currently fall back to Spotify outlinks |
| **Auth Error Handling** | Graceful handling of JWT middleware failures and expired/invalid Spotify token failures |
| **Rating Label Sync (Album List)** | When a rating is changed via the rating modal from the album list, the rating label badge on the row should update without a page reload |
| **Saved Tracks Sync** | Sync the user's saved tracks from Spotify so loved tracks can be highlighted within album views |
| **Simplify review flow** | Drop the time-based aspects of the review flow and let the user manually move an item from unrated → provisional → final. Simplify the rating flow so the rating modal opens directly rather than forcing a Q-and-A step first. |
| **Rethink tagging system** | The current tagging system isn't pulling its weight and needs a rethink. No replacement design yet. |
| **Auth-aware feed dormancy** | Classify feed failures — keep exponential backoff for transient errors, but give feeds a dormant/needs-reconnect state for auth/permission failures that stops auto-polling and surfaces a "reconnect Spotify" prompt in the UI. |
| **Visible toast for rate-limited user actions** | When the shared Spotify guard is paused, a user-initiated action fails fast with 429 + Retry-After, but HTMX doesn't swap 4xx responses. Decide the app-wide transient-error display mechanism and render a "Spotify is rate-limiting us, try again shortly" toast. |
| **Primary genres — user controls** | Two gaps: (1) no way to turn the genre badges off; (2) genres are entirely app-derived from Discogs with no manual override. Spec per-album genre editing and a display-preference toggle. |

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

- **Progressive Web App (PWA)** — whether to convert Wax to a PWA for offline support and installability; deferred until the mobile experience is more fully developed.
- **Social features** — Goodreads-style network is a natural long-term direction, but it's unclear whether this belongs in the core product or as a separate surface; secondary to personal library depth.