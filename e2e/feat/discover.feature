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

  Scenario: A clear control is offered only while the search has a value
    Given a logged-in user on the discover page
    When they have not yet typed anything
    Then no clear control is offered in the search bar
    When they type a value into the search
    Then a clear control is offered in the search bar
    When they erase the value with the keyboard
    Then the clear control is no longer offered

  Scenario: Activating the clear control resets the search to its empty state
    Given a logged-in user on the discover page with a search value entered and results showing
    When they activate the clear control
    Then the search input is empty
    And the results panel shows the same empty-query message as a fresh visit
    And the search input is focused, ready for the next keystroke
