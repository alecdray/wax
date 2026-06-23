Feature: App header

  Every authenticated library page carries a slim top header with the wax
  wordmark and the feed sync-status control, giving brand identity and feed
  status a single consistent home across pages.

  Scenario: The app header appears on every authenticated page
    Given a logged-in user
    Then the header with the wordmark and feeds control is present on the dashboard, discover, and album-detail pages
