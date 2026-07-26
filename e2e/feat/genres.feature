Feature: Primary genres

  Albums are auto-assigned a small set of app-curated primary genres, shown as
  read-only badges and usable as a coarse library filter.

  Scenario: An album's primary genres show as badges on its detail page
    Given an album with resolved genres in the library
    When the user opens the album detail page
    Then the album's primary genres are listed as badges

  Scenario: Filtering the library by a primary genre keeps matching albums
    Given an album with a pop genre in the library
    When the user filters the dashboard by the pop genre
    Then the album is shown with its primary-genre badge

  Scenario: Filtering by a genre the fixture album lacks excludes it from results
    Given an album without a reggae genre in the library
    When the user filters the dashboard by the reggae genre
    Then the album is absent from the filtered results
