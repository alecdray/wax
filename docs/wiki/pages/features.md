---
description: >
  What Wax can do today — shipped, production features and their scope. Belongs here: feature
  descriptions, user-facing behaviour, and constraints of live functionality. Does not belong here:
  planned or in-queue work (→ roadmap), implementation internals (→ architecture), or data
  structure definitions (→ data-model).
links:
  - roadmap
  - vision
  - data-model
  - frontend
  - integrations
---

[wiki](../wiki.md)

# Features

Shipped features in production today.

## My Library

The core of the app. A user's library is their collection of music — albums, artists, tracks, and releases (format variants: digital, vinyl, CD, cassette).

A stats bar at the top of the dashboard shows the user's total artist, album, and track counts at a glance.

- Digital media is automatically synced from Spotify on a recurring schedule
- Users can browse and sort their library by title, artist, rating, date added, and last played — clicking a column header reloads the table with the new sort applied
- Album titles link to the album's detail page; Spotify remains accessible via an icon on both the table row and the detail page
- Albums load in batches of 20 using infinite scroll — additional albums load automatically as the user scrolls to the bottom of the table

Each table row exposes per-row actions:
- A rating control — shown as a score badge if the album is already rated, or a "Rate" button if not — opens the [rating modal](#rankings--reviews) directly
- Notes and Tags are accessible via an ellipsis (⋯) dropdown menu on each row

A carousel above the library table offers two togglable views for surfacing albums worth acting on:

| View | What it shows |
|---|---|
| **Recently Spun** | Albums from [listening history](#listening-history), in reverse-chronological order; default view on load |
| **Unrated** | Albums in the library with no [rating](#rankings--reviews) yet — a prompt to rate what you've been playing |

Each carousel item shows album art, title, and artist, and links directly to the album in Spotify. Switching tabs swaps the carousel content without a full page reload; only the inactive tab is clickable at any time.

---

## Album Detail

Each album in the library has a dedicated detail page showing all information Wax holds for it:

- Cover art, title, and artists (with Spotify links)
- Release formats in the user's library with the date each was added
- Rating, review notes, and tags — all editable from the page via the same modals used on the dashboard
- Last played date (when listening history is available)
- Track list

The page is designed mobile-first with a stacked layout. Albums not in the user's library return a 404.

---

## Listening History

Wax records what you've been playing by polling Spotify's recently played tracks every hour in the background.

- Each track play is stored with a timestamp and linked to its album and artist
- Last played time per album is derived from play history and surfaces as a sort option in the library
- Sync runs hourly for all users with an active Spotify connection; token failures are handled gracefully (that user is skipped, the job continues)
- Because Spotify's API returns only the last 50 recently played tracks, syncing frequently is important — gaps can occur during long sessions or if the sync falls behind (see [integrations](./integrations.md) for the full constraint)

---

## Rankings & Reviews

Users can rate and review albums in their library.

- 0–10 score per album, manually set or guided by a 3-question questionnaire (scoring approach inspired by [Pitchfork](https://pitchfork.com/news/how-to-rate-albums-using-pitchfork-scores/))
- Free-text review notes attached to the rating
- Rating and review are independently updatable

### Rating Modal

The rating modal is the primary entry point for scoring an album. It always opens to the **confirm form**, regardless of whether the album already has a rating — there is no "questionnaire first" path.

The confirm form contains:
- A numeric score input (0–10, step 0.1) with a live label (e.g. "Heavy Rotation") that updates as the score changes
- A "Lock in" button to save the rating
- A **?** button that navigates to the questionnaire within the modal
- A trash button that deletes an existing rating and closes the modal

After a successful save, the modal closes automatically.

The **review notes** modal is separate — it contains a textarea (up to 2,000 characters) and a "Save Notes" button. It also closes automatically after a successful save.

### Rating Questionnaire

The questionnaire produces a recommended score from three questions, each capturing a distinct dimension of the listening experience:

| Question | What it measures |
|---|---|
| **Track consistency** — "How would you describe the album track by track?" | Whether the album holds up across its runtime; skips and weak tracks pull this down |
| **Emotional impact** — "How did this album make you feel while listening?" | The primary differentiator between good and great; measures absorption, not just enjoyment |
| **Gut check** — "When the album ended, what was your immediate reaction?" | The overall impression beyond individual tracks — relief, satisfaction, or the urge to restart |

Emotional impact carries the most weight because an album full of consistent, pleasant tracks can still leave you cold — while a flawed record with genuine emotional pull tends to be the one you remember.

### Rating Scale

Every score maps to a label. The labels are intentionally opinionated — they describe a relationship with the album, not a letter grade.

| Score | Label | What it means |
|---|---|---|
| 0–2.9 | **DOA** | Active rejection; nearly impossible via the questionnaire (floor is 2.0), so usually a manual override |
| 3.0–3.9 | **Nope** | Genuinely disliked — not just indifferent, a real negative reaction |
| 4.0–5.9 | **Not For Me** | The widest band; the album may be objectively fine but doesn't resonate — could be genre, mood, or just not clicking |
| 6.0–6.4 | **Lukewarm** | Faint praise — you finished it, maybe found some moments, but it didn't stick |
| 6.5–6.9 | **Solid** | A clear positive without being a standout; good enough to recommend with a caveat |
| 7.0–7.9 | **Staff Pick** | Actively good; worth telling someone about |
| 8.0–8.9 | **Heavy Rotation** | Something you return to — essential listening |
| 9.0–9.9 | **Instant Classic** | Near-perfect; exceptional across the board |
| 10.0 | **Masterpiece** | The absolute ceiling; reserved for records that are genuinely transformative |

The narrow bands at 6.x (Lukewarm / Solid) reflect the reality that many albums cluster in the "decent but not memorable" zone — splitting that range gives more resolution where most scores land.

---

## Tagging

Users can apply custom tags to albums for flexible organization and discovery.

- Tags belong to tag groups or stand alone
- No limit on tags per album or tags per group
- Two built-in tag group concepts: **Sound** (genre, style, influences) and **Mood** (context, feeling, occasion)
- Users define their own tags within these groups

### Tagging Modal

The tags modal is accessible from the ellipsis dropdown on each library row and from the tags button on the album detail page.

- Tags are entered in a text input; pressing **Enter** or **comma** converts the current text into a chip shown above the input
- **Backspace** on an empty input removes the last chip
- An autocomplete dropdown appears while typing, suggesting existing tags (up to 8 results, excluding already-selected tags)
- A tag can optionally be assigned to a group by clicking a group button before or after the chip is created
- Clicking **Save Tags** submits all chips and closes the modal

