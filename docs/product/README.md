# Product

The model of the **experience** — organized by feature / user journey. Each doc describes one feature as it exists today: what a user can do, what happens when they do it, and why the feature exists. Present-state only — no roadmap, no history.

Start with [vision.md](vision.md) for the overarching product philosophy.

**Product is the entry point; [domain](../domain/) is the deep reference.** A reader meets a feature here — the why, the what, and its headline business rules at a level a newcomer or stakeholder follows — then follows a link into domain for the complete model. Product is the front door; domain is the reference behind it.

Features and the domain model are **orthogonal decompositions of the same system**: one feature may compose several domain concepts, and one domain concept may surface across many features. This doc is the only home for that composed, cross-module, user-facing view.

**Two boundaries keep this doc from rotting into others:**
- **Link, don't redefine.** For any term or rule with a specific meaning, link to [domain](../domain/) — never restate the full definition here.
- **What and why, not how.** Technical structure belongs in [architecture](../architecture/) and the module READMEs.

## Features

| Feature | Doc | What |
|---|---|---|
| Library | [library.md](library.md) | Collect albums: own, wishlist, or remove. Browse by format, sort, and filter. |
| Radar | [radar.md](radar.md) | Watch albums for later. Inbox from Spotify, manual add, clear when owned. |
| Ratings | [ratings.md](ratings.md) | Score albums 0–10 with a weighted questionnaire. Save provisionally or finalize. |
| Tags | [tags.md](tags.md) | User-defined tag groups and vocabulary for organizing albums. |
| Notes | [notes.md](notes.md) | Freeform album annotations. |
| Genres | [genres.md](genres.md) | Curated, auto-assigned genre badges. |
| Listening History | [listening-history.md](listening-history.md) | Play records over time. |
| Discover | [discover.md](discover.md) | Search for albums to add to the radar. |