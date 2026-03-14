Feature: Library Dashboard

  The main hub of the app where users browse their music collection. Shows
  library statistics, a carousel of recent or unrated albums, and a sortable
  table of all albums.

  Scenario: Viewing the library dashboard
    Given a logged-in user with a library
    When they navigate to the dashboard
    Then they see their library statistics
    And they see the album carousel
    And they see the albums table

  Scenario: Default carousel shows Recently Spun
    Given a logged-in user on the dashboard
    Then the Recently Spun carousel tab is active

  Scenario: Switching the carousel to Unrated
    Given a logged-in user on the dashboard
    When they click the Unrated carousel tab
    Then the carousel reloads showing the Unrated view

  Scenario: Sorting albums by artist
    Given a logged-in user on the dashboard
    When they click the Artists column header
    Then the albums table reloads sorted by artist

  Scenario: Opening the rating modal from an album row
    Given a logged-in user on the dashboard with at least one album
    When they click the rating control on an album row
    Then the rating modal opens


  Scenario: Opening the tags modal from an album row
    Given a logged-in user on the dashboard with at least one album
    When they open the action menu on an album row and click Tags
    Then the tags modal opens
