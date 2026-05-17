Feature: Formats

  Each album in a user's library is available in four formats: digital,
  vinyl, CD, and cassette. The formats modal, opened from the album
  detail page, shows one row per format so the user can record which
  physical copies they own.

  Scenario: Formats modal opens from the album detail page
    Given a logged-in user on an album detail page
    When they click the formats button
    Then the formats modal opens with one row per format
