Feature: Bottom navigation

  The app's primary navigation is a fixed bottom bar present on every
  authenticated page. It marks the current top-level destination and moves
  between the library and discover surfaces.

  Scenario: The bottom nav marks the current destination on the dashboard
    Given a logged-in user is on the library dashboard
    Then the bottom nav marks "Library" as the current destination

  Scenario: Selecting Discover from the bottom nav navigates to the discover page
    Given a logged-in user is on the library dashboard
    When they select "Discover" in the bottom nav
    Then the discover page is shown
    And the bottom nav marks "Discover" as the current destination
