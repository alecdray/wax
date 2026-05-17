Feature: Feed Sync

  Users can trigger an on-demand sync of an external feed (e.g. their Spotify
  collection) from the feeds dropdown in the dashboard header, and see
  immediate confirmation that the sync has been accepted.

  Scenario: Triggering a feed sync from the dashboard shows a syncing indicator
    Given a logged-in user on the dashboard with a configured feed
    When they open the feeds dropdown in the header
    And they click the sync button for the feed
    Then the feed entry shows that a sync is in progress
