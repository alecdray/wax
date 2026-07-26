# Modal Header Bar

Restructure the `Modal` primitive's close affordance from a floating absolute-positioned button to an in-flow header bar. Add a `HeaderActions` slot so callers can inject contextual controls (e.g. a back button) to the left of the close button. Migrate the rating modal's multi-step back button from inline content into the header.

## Motivation

The current close button (`absolute right-2 top-2`) floats over modal content. `BaseQuestionsFormFrag` already has a back button with `pr-8` on its wrapper as an explicit workaround for the overlap. Moving to an in-flow header bar fixes the layering and gives the back button a proper structural home.

## Goals

- Close button is in-flow, not floating
- Header bar is a flex row: `[HeaderActions — flex-1]  [✕]`
- Header is sticky — it does not scroll with modal content
- `HeaderActions` is an optional `templ.Component` prop; when nil the left side is empty, close button stays right-aligned
- When provided, `HeaderActions` takes the full remaining width — callers control internal alignment
- All existing single-step callers work unchanged (`HeaderActions` omitted)
- Rating modal's back button moves from inline content to the header, via full modal re-renders on step transitions

## Out of scope

- Changing the modal's width, height, or bottom-sheet vs. centred behaviour
- Per-modal header titles as a built-in field
- Any structural change to `ForceCloseModal`

## Design

```
┌─ modal-box ────────────────────────────────────────┐
│ ┌─ header bar (flex items-center) ───────────────┐ │
│ │  [HeaderActions — flex-1, caller-controlled]   │ │
│ │                                          [✕]   │ │
│ └────────────────────────────────────────────────┘ │
│                                                     │
│  modal content                                      │
│                                                     │
└─────────────────────────────────────────────────────┘
```

Header bar markup sketch:

```html
<!-- modal-box: p-0 so header sits flush; content area owns its own padding -->
<div class="modal-box flex flex-col p-0 sm:max-h-[95vh]">
  <!-- sticky header: bg matches modal-box so content doesn't bleed through -->
  <div class="sticky top-0 z-10 flex items-center px-3 pt-3 pb-2 bg-base-100">
    <div class="flex-1">
      <!-- HeaderActions slot, or empty -->
    </div>
    <form method="dialog">
      <button class="btn btn-sm btn-ghost btn-square" aria-label="Close">
        <!-- x-lg icon -->
      </button>
    </form>
  </div>
  <!-- scrollable content -->
  <div class="flex flex-col px-3 pb-3">
    <!-- content -->
  </div>
</div>
```

The `modal-box` provides `overflow-y: auto` (DaisyUI default), making it the scroll container. `sticky top-0` on the header sticks to the top of that container. `bg-base-100` on the header matches the box background so scrolling content doesn't bleed through.

## Rating modal step-navigation change

Currently step transitions swap only the content area (`hx-target="closest form"` / `outerHTML`), leaving the modal shell untouched. To put the back button in the header, step transitions must instead re-render the full modal shell via OOB `ModalContainer` swap.

**Step → header mapping:**

| Step | HeaderActions |
|---|---|
| Score entry (`RatingConfirmFormFrag`) | nil (no back) |
| Base questions (`BaseQuestionsFormFrag`) | back button → `GET /app/review/rating-recommender/confirm` |
| Confidence interstitial (`ConfidenceInterstitialFrag`) | back button → `GET /app/review/rating-recommender/questions` |

Each step gets a dedicated `*ModalFrag` wrapper that calls `templates.Modal(RatingModalId, ModalProps{HeaderActions: …, ModalContent: …})`. The route handlers that currently return bare fragment components are updated to return these wrappers instead.

The inline back button div (with `pr-8`) in `BaseQuestionsFormFrag` is removed.

## Change surface

| File | Change |
|---|---|
| `src/internal/core/templates/modal.templ` | Add `HeaderActions templ.Component` to `ModalProps`; replace `absolute` close button with header bar |
| `src/internal/core/templates/modal_templ.go` | Regenerated (`task build/templ`) |
| `src/internal/review/adapters/views/rating_modal_frag.templ` | Per-step `*ModalFrag` components with appropriate `HeaderActions` |
| `src/internal/review/adapters/views/base_questions_form_frag.templ` | Remove inline back button div and `pr-8` hack |
| `src/internal/review/adapters/views/confidence_interstitial_frag.templ` | Add back button (→ questions step) via new modal wrapper |
| Review route handlers | Return per-step modal frags instead of bare fragment components |
| All other callers (`album_actions`, `formats`, `tags`) | No change — `HeaderActions` omitted |

## Testing

- Existing e2e for modal close button continues to pass unchanged
- Rating modal step navigation (questionnaire open, back, confidence interstitial back) should be covered — check existing e2e before deciding if new cases are needed
