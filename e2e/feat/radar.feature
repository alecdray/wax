Feature: Radar

  The radar page lets a user search Spotify for albums to add to their
  collection and surfaces the albums they have flagged as worth listening to
  later on their radar.

  Scenario: Searching Spotify for an album returns matching results
    Given a logged-in user on the radar page
    When they type an album query into the search field
    Then a list of matching albums from Spotify appears

  Scenario: Visiting the radar page shows the radar
    Given a logged-in user
    When they navigate to the radar page
    Then they see their radar — either the albums currently on it or an empty-state message inviting them to add some

  Scenario: The radar page offers the Spotify radar inbox control
    Given a logged-in user on the radar page
    Then the radar inbox control is shown — either an enable button or a link to their radar playlist

  Scenario: A clear control is offered only while the search has a value
    Given a logged-in user on the radar page
    When they have not yet typed anything
    Then no clear control is offered in the search bar
    When they type a value into the search
    Then a clear control is offered in the search bar
    When they erase the value with the keyboard
    Then the clear control is no longer offered

  Scenario: Activating the clear control resets the search to its empty state
    Given a logged-in user on the radar page with a search value entered and results showing
    When they activate the clear control
    Then the search input is empty
    And the radar grid returns as the main region, as on a fresh visit
    And the results panel is hidden
    And the search input is focused, ready for the next keystroke

  Scenario: Typing several characters quickly issues a single search after the debounce
    Given a logged-in user on the radar page
    When they type a multi-character query in rapid succession
    Then only one search request is issued for the burst, not one per keystroke

  Scenario: Submitting the search bypasses the debounce and fires immediately
    Given a logged-in user on the radar page who has just typed one character
    When they submit the search bar before the debounce delay would elapse
    Then a search request fires straight away rather than waiting for the debounce

  Scenario: A library change refreshes the search results panel
    Given a logged-in user on the radar page who has searched and sees results
    When a library-change event occurs on the page
    Then the results panel re-fetches and re-renders without a full page reload

  Scenario: Searching swaps the radar grid out for the results panel
    Given a logged-in user on the radar page seeing their radar grid
    When they search for an album
    Then the new results appear inside the results panel
    And the radar grid is hidden while the search is active

  Scenario: A new search result scrolls the results panel back to the top
    Given a logged-in user on the radar page who has searched and scrolled the results down
    When they run a new search
    Then the results panel is scrolled to the top after the new results swap in

  Scenario: Erasing the query back to empty restores the radar grid
    Given a logged-in user who has just loaded the radar page and sees their radar grid
    When they type a query and then erase it back to empty with the keyboard
    Then the radar grid returns as the main region
    And the results panel is hidden
