Feature: Loading indicators

  The app renders a global progress bar and region-level overlays to signal
  in-flight HTMX requests, and disables action buttons while their request
  is in flight.

  Scenario: Global progress bar is present in the layout
    Given I am logged in
    When I visit the dashboard
    Then a global progress bar element is attached to the page

  Scenario: Discover search results region carries a loading overlay
    Given I am logged in
    When I visit the discover page
    Then the discover results region contains a region overlay element

  Scenario: Add-to-library button declares the disable-on-request contract
    Given I am logged in and on the discover page
    When I search for an album not already in my library
    And I open the album actions modal for that album
    Then the primary action button has the hx-disabled-elt="this" attribute
    And the primary action button carries the btn-busy class
