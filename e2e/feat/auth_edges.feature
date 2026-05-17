Feature: Auth edges

  Covers the auth lifecycle edges outside the login page itself:
  logging out from an authenticated session, and being shown the
  unauthorized page when an unauthenticated user attempts to reach
  a protected route.

  Scenario: Logged-in user logs out and returns to the login page
    Given a logged-in user is on the library dashboard
    When they open the user menu and click "Logout"
    Then they are returned to the login page

  Scenario: Unauthenticated user is shown the unauthorized page on a protected route
    Given an unauthenticated user navigates directly to a protected route
    Then the unauthorized page is shown
