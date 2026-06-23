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

  Scenario: The discover page offers the Spotify radar inbox control
    Given a logged-in user on the discover page
    Then the radar inbox control is shown — either an enable button or a link to their radar playlist

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

  Scenario: Typing several characters quickly issues a single search after the debounce
    Given a logged-in user on the discover page
    When they type a multi-character query in rapid succession
    Then only one search request is issued for the burst, not one per keystroke

  Scenario: Submitting the search bypasses the debounce and fires immediately
    Given a logged-in user on the discover page who has just typed one character
    When they submit the search bar before the debounce delay would elapse
    Then a search request fires straight away rather than waiting for the debounce

  Scenario: A library change refreshes the discover results panel
    Given a logged-in user on the discover page
    When a library-change event occurs on the page
    Then the results panel re-fetches and re-renders without a full page reload

  Scenario: A new search result swaps into the same panel and leaves the rest of the page intact
    Given a logged-in user on the discover page with the radar visible above the search bar
    When they search for an album
    Then the new results appear inside the results panel only
    And the radar above the search bar remains in place

  Scenario: A new search result scrolls the results panel back to the top
    Given a logged-in user on the discover page who has searched and scrolled the results down
    When they run a new search
    Then the results panel is scrolled to the top after the new results swap in

  Scenario: Erasing the query back to empty restores the same empty-query message as a fresh visit
    Given a logged-in user who has just loaded the discover page and sees the empty-query message
    When they type a query and then erase it back to empty with the keyboard
    Then the empty-query message reappears in the results panel
