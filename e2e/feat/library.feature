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
