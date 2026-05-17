Feature: Rating Lifecycle

  Cross-surface invariants the rating-modal rework must hold once all the
  individual pieces are assembled. Each scenario walks through the running
  system the way a user does (and the way HTMX swaps it), exercising
  invariants that no single per-task scenario covers on its own.

  Scenario: Modal opens directly to the score-entry form from every UI affordance
    Given a logged-in user with a finalized rating on one album and a provisional rating on another
    When they open the rating modal from the score-readout on the dashboard for each album
    Then the modal always opens on the score-entry form regardless of the album's rating state

  Scenario: Score-entry pre-fill resolves the tie between two same-timestamp log entries by greater id
    Given a logged-in user with two rating-log entries on the same album that share a created_at
    When they open the rating modal for that album
    Then the score input is pre-filled with the score from the entry with the greater id

  Scenario: Dashboard carousel HTMX swap target id is stable
    Given a logged-in user on the dashboard
    Then the carousel section's DOM element id is carousel-section
    And HTMX swaps that target carousel-section land on the same element
