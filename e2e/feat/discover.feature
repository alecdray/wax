Feature: Discover

  The discover page lets a user search Spotify for albums to add to their
  collection and surfaces the albums they have flagged as worth listening to
  later on their radar.

  Scenario: Searching Spotify for an album returns matching results
    Given a logged-in user on the discover page
    When they type an album query into the search field
    Then a list of matching albums from Spotify appears

  Scenario: Visiting the discover page shows the radar
    Given a logged-in user
    When they navigate to the discover page
    Then they see their radar — either the albums currently on it or an empty-state message inviting them to add some
