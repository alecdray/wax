Feature: Library Dashboard

  The main hub of the app where users browse their music collection. Shows
  library statistics, a carousel of recent or unrated albums, and a visual
  list of all albums narrowed by a single unified search bar that combines
  free-text query with inline filter and sort controls.

  Scenario: Viewing the library dashboard shows list view
    Given a logged-in user with a library
    When they navigate to the dashboard
    Then they see their library statistics
    And they see the album carousel
    And they see the albums list

  Scenario: Album rows show art, title, and rating — no Spotify outlinks
    Given a logged-in user on the dashboard with at least one album
    Then each album row shows album art, title link, and rating
    And no Spotify outlinks appear in the list

  # Retired: "No ellipsis menu on album rows" — the original assertion targeted
  # data-testid values (album-row-menu, album-row-tags-button) that have never
  # been declared in any templ, so the check passed vacuously. The row's
  # closed affordance set (title link, tags section, score readout) is what
  # the templ enforces; testing absence of never-declared testids proves nothing.

  Scenario: Default carousel shows Recently Spun
    Given a logged-in user on the dashboard
    Then the Recently Spun carousel tab is active

  Scenario: Switching the carousel to Unrated
    Given a logged-in user on the dashboard
    When they click the Unrated carousel tab
    Then the carousel reloads showing the Unrated view

  Scenario: Unified search bar replaces the chip-modal bar
    Given a logged-in user on the dashboard
    Then the unified search bar is visible above the album list
    And the legacy chip-modal bar is not present anywhere on the page

  Scenario: Typing in the search bar narrows the album list live
    Given a logged-in user on the dashboard
    When they type a query that matches at least one album in the search bar
    Then the album list is replaced via HTMX without a page reload
    And every visible row matches the query in title or artist

  Scenario: Clearing the search bar restores the full library
    Given a logged-in user on the dashboard with a non-empty search query active
    When they clear the search bar
    Then the album list shows the full library again

  Scenario: Sort control on the unified bar reorders the list
    Given a logged-in user on the dashboard
    When they open the sort popover and choose Artist
    Then the albums list reloads and the sort control shows "Artist"

  Scenario: Rating control on the unified bar narrows by minimum rating
    Given a logged-in user on the dashboard
    When they open the rating popover and set a minimum rating of 7
    Then the albums list reloads
    And the rating control shows the active filter

  Scenario: Filtering to unrated from the unified bar
    Given a logged-in user on the dashboard
    When they open the rating popover and choose Unrated only
    Then the albums list reloads showing only unrated albums
    And the rating control shows "Unrated"

  Scenario: Format control on the unified bar supports multi-select
    Given a logged-in user on the dashboard
    When they open the format popover and select Vinyl
    Then the albums list reloads
    And the format control shows "vinyl"

  Scenario: Artist control on the unified bar opens a searchable list when artists exist
    Given a logged-in user on the dashboard with artists in their library
    When they open the artist popover
    Then a searchable artist checkbox list is visible inside the popover

  Scenario: Opening the rating modal from an album row
    Given a logged-in user on the dashboard with at least one album
    When they click the rating control on an album row
    Then the rating modal opens

  # --- Task 1.4 — Active non-default state visible on the bar at rest ---

  Scenario: Bar shows no badges at full defaults
    Given a logged-in user on the dashboard
    Then the unified search bar shows no active-state badges

  Scenario: Bar surfaces a badge for a non-default filter at rest
    Given a logged-in user on the dashboard
    When they apply a non-default filter from the unified bar
    Then a compact badge naming that filter dimension is visible on the bar without opening any popover

  Scenario: Bar surfaces a sort badge for a non-default sort at rest
    Given a logged-in user on the dashboard
    When they apply a non-default sort from the unified bar
    Then a compact badge naming the active sort is visible on the bar without opening any popover

  # --- Task 1.5 — URL reflects full view state; reloading reproduces it ---

  Scenario: Bare dashboard URL stays bare in the address bar
    Given a logged-in user on the dashboard
    Then the URL contains no q, sortBy, dir, minRating, maxRating, rated, format, or artist params

  Scenario: Applying state writes the expected params to the URL
    Given a logged-in user on the dashboard
    When they apply a non-default sort and a text query
    Then the URL contains the corresponding params under their canonical names

  Scenario: Reloading the URL reproduces the same view
    Given a logged-in user on the dashboard with applied filter and sort state
    When they reload the page
    Then the same params remain in the URL and the bar reflects the applied state

  Scenario: A fresh browser context renders the same view for a deep URL
    Given a deep dashboard URL with filter and sort params
    When a fresh authenticated context visits that URL
    Then the bar reflects the same state and the album list matches

  Scenario: Setting a value back to its default removes the param from the URL
    Given a logged-in user on the dashboard with a non-default filter active
    When they activate reset on the unified bar
    Then the URL no longer contains the dropped param

  # --- Task 1.6 — Infinite-scroll pagination preserves all state ---

  Scenario: Pagination request carries every active filter and sort param
    Given a logged-in user on the dashboard with filter and sort state applied
    When they scroll to the pagination sentinel
    Then the next-page request URL includes every active param
    And the appended rows respect the narrowed view with no duplicates across the page boundary

  # --- Task 1.7 — Zero-result state with one-click reset ---

  Scenario: Zero-result view shows a non-judgemental message and a single reset control
    Given a logged-in user on the dashboard with a query that matches no albums
    Then the list region shows a short empty-state message
    And a single visible reset control is present in the empty state

  Scenario: Activating reset from the empty state restores the full library and bare URL
    Given a logged-in user on the dashboard with a query that matches no albums
    When they activate the reset control
    Then the URL is bare, the full library re-renders, and no active-state badges are shown

  # --- Product criteria (PC) — whole-system invariants of the assembled feature ---
  #
  # These scenarios assert the cross-cutting product criteria PC1, PC2, PC3 as
  # emergent properties, distinct from the per-task scenarios above. They are
  # the gates the unified-search build must hold even after all tasks pass
  # individually.

  Scenario: PC1 — combined q + filter + sort produces the predicted set in the predicted order
    Given a logged-in user on the dashboard
    When they apply a text query, a non-default filter, and a non-default sort together
    Then the rendered album list equals the AND of all three dimensions in the active sort order

  Scenario: PC2 — a captured URL reproduces the exact DOM order in a fresh browser context
    Given a logged-in user who composes a non-trivial view state via the unified bar
    When the resulting URL is opened in a fresh authenticated browser context
    Then the rendered album list matches the original DOM order row-for-row

  Scenario: PC2 — every reachable param combination round-trips via the URL
    Given a logged-in user on the dashboard
    When they reach each of several combined view states via the bar and the resulting URL is opened fresh
    Then each fresh-context render matches the corresponding originally-rendered list
