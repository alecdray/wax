Feature: Album Detail Page

  A dedicated page for a single album in the user's library, surfacing all
  album metadata, ratings, notes, tags, and release formats in a mobile-friendly layout.

  Scenario: Viewing an album in the library
    Given a logged-in user with at least one album in their library
    When they navigate to the album detail page
    Then they see the album cover art, title, and artist names
    And they see the release format badges
    And they see the album rating or a Rate button

  Scenario: Navigating to the detail page from the dashboard
    Given a logged-in user on the dashboard
    When they click an album title in the library table
    Then they are taken to that album's detail page

  Scenario: Rating an album from the detail page
    Given a logged-in user on an album detail page
    When they click the Rate button
    Then the rating modal opens

  Scenario: Editing notes from the detail page
    Given a logged-in user on an album detail page
    When they click the notes button
    Then the notes modal opens

  Scenario: Editing tags from the detail page
    Given a logged-in user on an album detail page
    When they click the tags button
    Then the tags modal opens

  Scenario: Accessing an album not in the library
    Given a logged-in user
    When they navigate to an album ID that is not in their library
    Then they receive a 404 response

  Scenario: Last played date is shown when available
    Given a logged-in user on an album detail page for an album with listening history
    Then the last played date is displayed

  Scenario: Last played date is absent when not available
    Given a logged-in user on an album detail page for an album with no listening history
    Then no last played date is shown

  Scenario: Back navigation to the dashboard
    Given a logged-in user on an album detail page
    When they click the back link in the header
    Then they are returned to the library dashboard
