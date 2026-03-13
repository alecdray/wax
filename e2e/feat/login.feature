Feature: Login

  The login page is the entry point for all unauthenticated users.
  It presents a Spotify OAuth login button that initiates the auth flow.

  Scenario: Unauthenticated user sees the login button
    Given an unauthenticated user visits the app
    Then the login button is visible

  Scenario: Login button links to Spotify OAuth
    Given an unauthenticated user is on the login page
    Then the login button href points to accounts.spotify.com

  Scenario: Login page has the app title
    Given an unauthenticated user visits the app
    Then the page title contains "wax"

  Scenario: Authenticated user is redirected to the library
    Given a user with a valid session token visits the root
    Then they are redirected to the library dashboard
