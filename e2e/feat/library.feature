Feature: Library Dashboard

  The main hub of the app where users browse their music collection. Shows
  library statistics, a carousel of recent or unrated albums, and a visual
  list of all albums with chip-based sort and filter controls.

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

  Scenario: Sort chip is visible and shows default sort
    Given a logged-in user on the dashboard
    Then the sort chip is visible and shows "Date Added"

  Scenario: Sort chip opens a modal
    Given a logged-in user on the dashboard
    When they click the sort chip
    Then a sort modal opens with sort field and direction options

  Scenario: Sorting by artist via sort chip reloads the list
    Given a logged-in user on the dashboard
    When they click the sort chip and select Artist
    Then the albums list reloads and the sort chip shows "Artist"

  Scenario: Rating chip opens a modal with min/max inputs
    Given a logged-in user on the dashboard
    When they click the rating chip
    Then a rating modal opens with min/max number inputs and rated/unrated options

  Scenario: Rating chip becomes active after applying a min rating filter
    Given a logged-in user on the dashboard
    When they click the rating chip and set a minimum rating of 7
    Then the albums list reloads showing only albums rated 7 or above
    And the rating chip shows the active filter

  Scenario: Filtering to unrated only shows unrated chip label
    Given a logged-in user on the dashboard
    When they click the rating chip and select Unrated only
    Then the albums list reloads showing only unrated albums
    And the rating chip shows "Unrated"

  Scenario: Format chip opens a modal with format options
    Given a logged-in user on the dashboard
    When they click the format chip
    Then a format modal opens with digital, vinyl, CD, and cassette options

  Scenario: Format chip becomes active after selecting vinyl
    Given a logged-in user on the dashboard
    When they click the format chip and select Vinyl
    Then the albums list reloads showing only vinyl albums
    And the format chip shows "vinyl"

  Scenario: Artist chip opens a modal when artists exist
    Given a logged-in user on the dashboard with artists in their library
    When they click the artist chip
    Then an artist modal opens with a searchable checkbox list

  Scenario: Opening the rating modal from an album row
    Given a logged-in user on the dashboard with at least one album
    When they click the rating control on an album row
    Then the rating modal opens
