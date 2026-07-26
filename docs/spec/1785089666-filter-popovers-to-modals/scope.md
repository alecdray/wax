# Filter popovers to modals

Convert the unified search bar's 5 filter/sort inline popovers (sort, rating, format, genre, artist) from absolute-positioned Alpine.js dropdowns to native `<dialog>` modals using DaisyUI's `modal-bottom sm:modal-middle` pattern. Each chip button opens a dialog that slides up from the bottom on mobile and centers on desktop — consistent with the rating and tags modals elsewhere in the app.

---

## Current implementation

Each filter chip is wrapped in `<div x-data="{ open: false }" class="relative">`. The button toggles `open`, an absolute-positioned `<div x-show="open" @click.outside="open = false">` drops down from the button, and each form inside adds `@submit="open = false"` to close on Apply.

**Problems this change solves:**
- Absolute dropdowns overflow the viewport on mobile even with `left-0` anchoring (tall content, small screens)
- Alpine.js `@click.outside` is fragile on scroll/touch
- The pattern is inconsistent with the rest of the app, which uses DaisyUI `<dialog>` modals for user-interactive overlays

---

## Proposed implementation

### Structure per filter

Replace each `<div x-data ...>` wrapper + absolute popover with a chip button + `<dialog>`:

```templ
// Chip button — opens the dialog
<button
    type="button"
    class="btn btn-sm btn-outline"
    onclick="document.getElementById('filter-sort-modal').showModal()"
    data-testid="unified-search-bar-sort-toggle"
>
    Sort: ...
</button>

// Dialog — inline in unifiedSearchBar, not OOB
<dialog id="filter-sort-modal" class="modal modal-bottom sm:modal-middle">
    <div class="modal-box">
        // ✕ close button (native form method=dialog)
        <form method="dialog">
            <button class="btn btn-sm btn-ghost absolute right-2 top-2">✕</button>
        </form>
        // Filter form — same HTMX attributes as before
        <form
            hx-get="/app/library/dashboard/albums-table"
            hx-target="#album-list"
            hx-indicator="#album-list-results"
            hx-swap="outerHTML"
            hx-push-url="true"
            hx-on:htmx:after-swap="this.closest('dialog').close()"
        >
            ... hidden state inputs, filter controls, Apply button ...
        </form>
    </div>
    // Backdrop click to close
    <form method="dialog" class="modal-backdrop"><button>close</button></form>
</dialog>
```

### What changes

| Before | After |
|---|---|
| `<div x-data="{ open: false }" class="relative">` wrapper | Removed — no wrapper needed |
| `<div x-show="open" x-cloak @click.outside="open = false" class="absolute z-50 ...">` | `<dialog id="filter-X-modal" class="modal modal-bottom sm:modal-middle">` |
| `@click="open = !open"` on button | `onclick="document.getElementById('filter-X-modal').showModal()"` |
| `@submit="open = false"` on forms | `hx-on:htmx:after-swap="this.closest('dialog').close()"` on forms |
| Backdrop: `@click.outside="open = false"` | `<form method="dialog" class="modal-backdrop"><button>close</button></form>` |

### What stays the same

- All HTMX form attributes, hidden state inputs, and filter controls — unchanged
- Chip button labels and active-state styling (`btn-soft btn-primary` / `btn-outline`)
- `unifiedSearchBarBadges` row
- Artist popover's inner `x-data="{ search: '' }"` for artist-name filtering — preserved inside the dialog

### Dialog IDs

`filter-sort-modal`, `filter-rating-modal`, `filter-format-modal`, `filter-genre-modal`, `filter-artist-modal`

Plain HTML IDs; no `FormatCallableID` needed (that utility is for the HTMX-driven OOB modal system).

---

## Scope

**In:** `src/internal/library/adapters/views/albums_list_frag.templ` only. No HTTP handler changes, no new routes, no new primitives.

**Out:** The existing `templates.Modal` primitive is HTMX-driven (OOB swap into `#global-modal-container`) and not a fit here — filter dialogs are static page content. Do not use it.

**Regression note:** The `left-0` popover positioning fix on this branch becomes moot once the absolute popovers are removed; the commit stays in history but has no net effect after this change lands.

---

## Open question

The filter dialogs will render inline in `AlbumsListFrag`. Each page render includes 5 `<dialog>` elements in the DOM even when closed. This is fine for the app's scale (private, single-user) but worth noting for future reference.
