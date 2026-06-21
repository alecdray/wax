# Loading feedback for network actions

User-triggered network actions had no loading feedback — a slow or failing request left the interface looking idle, with no signal that anything was happening. Feedback is now layered: an app-wide indeterminate progress indicator on every request, a busy and non-resubmittable state on discrete one-shot actions, and a dim-and-overlay treatment on regions that reload data in place.

The global indicator is indeterminate because no action reports real progress; it gives a baseline "working" cue at near-zero cost and covers new actions automatically as they are added. Discrete actions add a local busy state so the user sees the specific control was registered and cannot double-fire it. In-place reloads dim and overlay their existing content rather than swapping in placeholders, because a reload refines content already on screen and preserving it keeps context; append-style loads (pagination) instead show a trailing spinner, since dimming content that stays on screen would be wrong.

Rejected: skeleton placeholders for the reloading regions — they suit the first paint of an empty region, not the refinement of content already displayed.
