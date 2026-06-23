Feature: Album actions

  The album-actions modal lets a user act on an album they encounter outside
  their library — typically a radar-page search hit — without first adding
  it. Its contents depend on the album's current relationship to the user:
  for an album that is not yet in the library, the user can open it in
  Spotify and add it to their radar.

  Scenario: Opening the actions modal for a non-library album from a discover search result
    Given a logged-in user on the radar page
    When they search for an album and click a result that is not in their library
    Then the album-actions modal opens showing the album's title, an Open in Spotify link, and an Add to radar action
