Feature: Rating Log

  Users can rate albums on a 0–10 scale. Each rating submission creates a new
  entry in the rating log. The most recent entry is the current rating. The album
  detail page shows the full rating history. Notes are attached to a rating entry
  at submission time and are not separately editable.

  Scenario: Rating modal opens to the confirmation form
    Given a logged-in user with an album in their library
    When they open the rating modal for that album
    Then the rating confirmation form is shown

  Scenario: Navigating to the questionnaire from the confirmation form
    Given a logged-in user with the rating confirmation form open
    When they click the questionnaire button
    Then the questionnaire form is shown

  Scenario: Completing the questionnaire produces a score
    Given a logged-in user with the rating questionnaire open
    When they answer all questions and click Calculate Rating
    Then the rating confirmation form is shown with a computed score

  Scenario: Saving a rating
    Given a logged-in user on the rating confirmation form
    When they enter a score between 0 and 10 and click Lock in
    Then the modal closes

  Scenario: Saving a rating with a note
    Given a logged-in user on the rating confirmation form
    When they enter a score and a note and click Lock in
    Then the modal closes
    And the album detail page shows the note on the rating history entry

  Scenario: No delete button in the rating modal
    Given a logged-in user opening the rating modal for a rated album
    When the confirm form is shown
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

  Scenario: No notes on dashboard
    Given a logged-in user viewing the library dashboard
    When they look at an album row
    Then no notes icon or notes editing button is present
