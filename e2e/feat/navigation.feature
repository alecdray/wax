Feature: Bottom navigation

  The app's primary navigation is a fixed bottom bar present on every
  authenticated page. It marks the current top-level destination and moves
  between the library and radar surfaces.

  Scenario: The bottom nav marks the current destination on the dashboard
    Given a logged-in user is on the library dashboard
    Then the bottom nav marks "Library" as the current destination

  Scenario: Selecting Radar from the bottom nav navigates to the radar page
    Given a logged-in user is on the library dashboard
    When they select "Radar" in the bottom nav
    Then the radar page is shown
    And the bottom nav marks "Radar" as the current destination
