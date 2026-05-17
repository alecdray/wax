Feature: Rating Log

  Users can rate albums on a 0-10 scale. The rating modal opens directly to a
  score-entry form for every album, with the questionnaire reachable as an
  unobtrusive opt-in tool. Each save creates a new entry in the rating log;
  the most recent entry is the current rating. The album detail page shows
  the full rating history. Notes are attached to a rating entry at submission
  time and are not separately editable.

  Scenario: Modal opens to the score-entry form for an unrated album
    Given a logged-in user opening the rating modal for an unrated album
    When the modal is shown
    Then the score-entry form is the first view
    And the score input is empty

  Scenario: Score-entry form pre-fills with the latest rating for a provisional album
    Given a logged-in user who has saved a provisional rating for an album
    When they open the rating modal for that album
    Then the score-entry form is the first view
    And the score input is pre-filled with the most recent saved score

  Scenario: Score-entry form pre-fills with the latest rating for a finalized album
    Given a logged-in user who has finalized a rating for an album
    When they open the rating modal for that album
    Then the score-entry form is the first view
    And the score input is pre-filled with the most recent saved score

  Scenario: Opening the questionnaire from the score-entry form pre-fills the score after submit
    Given a logged-in user on the score-entry form
    When they open the questionnaire, answer all questions, and submit
    Then the score-entry form is shown with the questionnaire's computed score pre-filled

  Scenario: Dismissing the questionnaire preserves the prior pre-fill
    Given a logged-in user on the score-entry form for a previously-rated album
    When they open the questionnaire and dismiss it without submitting
    Then the score-entry form is shown with the prior pre-filled score unchanged

  Scenario: Saving a rating
    Given a logged-in user on the score-entry form
    When they enter a score between 0 and 10 and click Lock in
    Then the modal closes

  Scenario: Saving a rating with a note
    Given a logged-in user on the score-entry form
    When they enter a score and a note and click Lock in
    Then the modal closes
    And the album detail page shows the note on the rating history entry

  Scenario: Saving on a finalized album keeps it finalized in one submission
    Given a logged-in user who has finalized a rating for an album
    When they open the modal, enter a new score, and click Lock in
    Then the modal closes with no confirmation prompt
    And the album remains finalized with the new score recorded

  Scenario: Finalize button is visible on the score-entry form for a provisional album
    Given a logged-in user with a provisional rating for an album
    When they open the rating modal for that album
    Then the score-entry form shows the Finalize button

  Scenario: Finalize button is hidden on the score-entry form for an unrated album
    Given a logged-in user opening the rating modal for an unrated album
    When the modal is shown
    Then the score-entry form does not show the Finalize button

  Scenario: Finalize button is hidden on the score-entry form for a finalized album
    Given a logged-in user who has finalized a rating for an album
    When they open the rating modal for that album
    Then the score-entry form does not show the Finalize button

  Scenario: Clicking Finalize promotes the album in one submission
    Given a logged-in user on the score-entry form for a provisional album
    When they enter a score and click Finalize
    Then the modal closes with no confirmation prompt
    And a new rating-log entry is recorded with the entered score
    And the album is shown as finalized

  Scenario: No delete button in the rating modal
    Given a logged-in user opening the rating modal for an album
    When the modal is shown
    Then no delete button is visible in the modal

  Scenario: Rating history is shown on the album detail page
    Given a logged-in user who has submitted multiple ratings for an album
    When they navigate to the album detail page
    Then a Rating History section is visible listing all past entries in reverse-chronological order

  Scenario: Submitting a second rating creates a new history entry
    Given a logged-in user who has already rated an album
    When they submit a new rating for the same album
    Then the album detail page shows two entries in the rating history
    And the most recent entry is displayed as the current rating

  Scenario: Deleting the current rating from history rolls back to the previous one
    Given a logged-in user who has submitted two ratings for an album
    When they navigate to the album detail page and click delete on the most recent entry
    Then that entry is removed from the history
    And the previous rating is now shown as the current rating

  Scenario: Deleting the only rating entry clears the album rating
    Given a logged-in user who has rated an album exactly once
    When they navigate to the album detail page and click delete on the only entry
    Then the history section is empty
    And the album shows no rating

  # Retired: "No notes on dashboard" - the original assertion targeted
  # data-testid values (album-row-notes, album-row-notes-button) that have
  # never been declared in any templ, so the check passed vacuously. Notes
  # live on the album detail page (SleeveNotesSectionFrag), never on rows.
