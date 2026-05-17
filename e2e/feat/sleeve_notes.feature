Feature: Sleeve Notes

  Users can capture freeform notes about an album in their library. Notes are
  written in Markdown via an inline editor on the album detail page and the
  rendered note replaces the editor once saved.

  Scenario: Saving a sleeve note from the album detail page
    Given a logged-in user on an album detail page
    When they open the sleeve notes editor, type a note, and click Save
    Then the editor is replaced by the rendered sleeve note
    And the rendered note contains the text they entered
