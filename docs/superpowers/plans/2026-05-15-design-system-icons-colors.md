# Design System: Icons and Colors — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate wax onto a single `Icon` primitive backed by Bootstrap Icons, replace ad-hoc opacity usage with a four-stop text emphasis scale plus `.is-disabled` / `.hover-fade` element-state utilities, and split documentation along the project's existing rule (conceptual content in `docs/design/`, verbatim implementation in CLAUDE.md files near the source).

**Architecture:** Three sequential streams on one feature branch. Stream A swaps the icons subsystem (vendor Bootstrap Icons, build a single parameterized `Icon` primitive, walk every call site, delete the per-icon templs). Stream B adds CSS utility classes via Tailwind v4 `@utility` blocks and migrates every raw `opacity-NN` and `text-base-content/NN` site to the new vocabulary, flagging mismatches for human review. Stream C lands the documentation: conceptual updates in `docs/design/`, verbatim CSS catalog in a new `static/CLAUDE.md`. Final history squashed to three commits per the spec.

**Tech Stack:** Templ (Go HTML templates), Tailwind CSS v4, DaisyUI, Bootstrap Icons (web font), Go, taskfile-based build (`task build/templ`, `task build/tailwind`, `task test/unit`, `task test/e2e`).

**Spec:** `docs/superpowers/specs/2026-05-15-design-system-icons-colors-design.md` (commit `1f5ae2b` on branch `spec/design-system-icons-colors`).

---

## File Structure

**Stream A — Icons:**

| Path | Action | Responsibility |
|---|---|---|
| `static/public/bootstrap-icons.css` | Create | Vendored Bootstrap Icons CSS, served at `/static/bootstrap-icons.css` |
| `static/public/fonts/bootstrap-icons.woff2` | Create | Vendored Bootstrap Icons font, referenced relatively from the CSS |
| `src/internal/core/templates/root.templ` | Modify (add `<link>`) | Loads BI CSS in the document head |
| `src/internal/core/templates/icons.templ` | Replace contents | The `Icon` primitive plus `IconStyle`/`IconProps` types |
| `src/internal/core/templates/icons.go` | Create | The `iconStyleSuffix(IconStyle) string` helper |
| `src/internal/core/templates/icons_test.go` | Create | Unit test for `iconStyleSuffix` |
| `src/internal/library/adapters/views/album_rating_history_frag.templ` | Modify | TrashIcon → Icon |
| `src/internal/library/adapters/views/album_detail_page.templ` | Modify | TagIcon → Icon |
| `src/internal/library/adapters/views/library_header_bar_frag.templ` | Modify | Collection/Compass/User → Icon (preserve outline/fill convention for nav) |
| `src/internal/library/adapters/views/feeds_dropdown_frag.templ` | Modify | Database/Warning/Spinner/XMark/Check → Icon (Spinner needs `animate-spin` wrapper) |
| `src/internal/library/adapters/views/format_icon_frag.templ` | Modify | Vinyl/CD/Cassette/Digital → Icon |
| `src/internal/review/adapters/views/rating_confirm_form_frag.templ` | Modify | QuestionMark → Icon |
| `src/internal/library/adapters/views/album_row_tags_section_frag.templ` | Modify | Pen/Tag → Icon |

**Stream B — Colors:**

| Path | Action | Responsibility |
|---|---|---|
| `static/src/main.css` | Modify (add utilities) | `.is-disabled`, `.hover-fade`, four `@utility` blocks for text emphasis |
| `src/**/*.templ` (per-site walk) | Modify | Map every raw `opacity-NN` to `is-disabled`/`hover-fade`/text scale; map every `text-base-content/NN` to one of the four named utilities |

**Stream C — Documentation:**

| Path | Action | Responsibility |
|---|---|---|
| `docs/design/design-system.md` | Modify | Visual direction prelude; new Icons section; new Colors sections; updated `main.css` rule. Names utilities; no verbatim CSS or Go. |
| `docs/design/principles.md` | Modify | Theme-tokens-not-raw-colors principle gets the named-utility addendum. |
| `static/CLAUDE.md` | Create | Verbatim catalog of wax-specific additions to the Tailwind+DaisyUI stack, plus BI vendoring note. |
| `src/internal/core/templates/CLAUDE.md` | Modify | Add a short note on the `Icon` primitive. |

---

## Bootstrap Icons name reference

Used while migrating call sites. Prefix is `bi-`; `Style: IconStyleFill` appends `-fill` to the name.

| Current templ | BI name | Has `-fill` variant |
|---|---|---|
| `CollectionIcon` | `collection` | yes |
| `CompassIcon` | `compass` | yes |
| `UserIcon` | `person-circle` | no (single style) |
| `DatabaseIcon` | `database` | no |
| `WarningIcon` | `exclamation-triangle` | yes |
| `SpinnerIcon` | `arrow-repeat` | no (and needs `animate-spin` wrapper at call site) |
| `XMarkIcon` | `x-circle` | yes |
| `CheckIcon` | `check-circle` | yes |
| `VinylIcon` | `vinyl` | no |
| `CDIcon` | `disc` | no |
| `CassetteIcon` | `cassette` | no |
| `DigitalIcon` | `file-music` | yes |
| `QuestionMarkIcon` | `question-circle` | yes (visual change: was Heroicons-style) |
| `PenIcon` | `pen` | no |
| `TagIcon` | `tag` | yes |
| `TrashIcon` | `trash` | no (visual change: was Heroicons-style) |
| `HomeIcon`, `NotesIcon`, `EllipsisVerticalIcon` | — | unused, deleted not migrated |

---

# Stream A — Icons

### Task 1: Vendor Bootstrap Icons assets

**Files:**
- Create: `static/public/bootstrap-icons.css`
- Create: `static/public/fonts/bootstrap-icons.woff2`

- [ ] **Step 1: Download Bootstrap Icons CSS**

```bash
curl -fsSL https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.css -o static/public/bootstrap-icons.css
```

- [ ] **Step 2: Download Bootstrap Icons font (woff2 only — we don't need woff)**

```bash
mkdir -p static/public/fonts
curl -fsSL https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/fonts/bootstrap-icons.woff2 -o static/public/fonts/bootstrap-icons.woff2
```

- [ ] **Step 3: Verify the CSS references the font path correctly**

Open `static/public/bootstrap-icons.css` and confirm the `@font-face` `src` includes `url("./fonts/bootstrap-icons.woff2?...")`. The relative path `./fonts/...` resolves correctly when the CSS is served from `/static/` because the font ends up at `/static/fonts/bootstrap-icons.woff2`. If the CSS references additional formats (woff), they will 404 but the woff2 will load — acceptable. If you want a clean console, edit the `src:` declaration to keep only the `woff2` entry.

- [ ] **Step 4: Smoke-check files exist**

```bash
ls -la static/public/bootstrap-icons.css static/public/fonts/bootstrap-icons.woff2
```
Expected: both files present, CSS ~80KB, woff2 ~120KB.

- [ ] **Step 5: Commit**

```bash
git add static/public/bootstrap-icons.css static/public/fonts/bootstrap-icons.woff2
git commit -m "feat(icons): vendor Bootstrap Icons 1.11.3"
```

---

### Task 2: Load Bootstrap Icons in the document head

**Files:**
- Modify: `src/internal/core/templates/root.templ:17`

- [ ] **Step 1: Add the `<link>` to root.templ**

Insert one line after the existing `main.css` link (line 17). The exact change:

```diff
       <link rel="stylesheet" href={ "/static/main.css?v=" + cssVersion }/>
+      <link rel="stylesheet" href="/static/bootstrap-icons.css"/>
       <script src="/static/htmx.min.js"></script>
```

No cache-busting query — BI is vendored at a pinned version.

- [ ] **Step 2: Regenerate templ output**

Run: `task build/templ`
Expected: clean exit, `root_templ.go` regenerated.

- [ ] **Step 3: Smoke verify by running the server briefly and curling**

In one terminal: `task dev` (this runs server + tailwind + templ in watch mode).
In another:
```bash
curl -sI http://localhost:$PORT/static/bootstrap-icons.css | head -3
```
Expected: `HTTP/1.1 200 OK`. Stop `task dev` (Ctrl-C) once verified.

- [ ] **Step 4: Commit**

```bash
git add src/internal/core/templates/root.templ src/internal/core/templates/root_templ.go
git commit -m "feat(icons): load Bootstrap Icons CSS in root layout"
```

---

### Task 3: Add `iconStyleSuffix` helper with unit test (TDD)

**Files:**
- Create: `src/internal/core/templates/icons.go`
- Create: `src/internal/core/templates/icons_test.go`

- [ ] **Step 1: Write the failing test first**

Create `src/internal/core/templates/icons_test.go`:

```go
package templates

import "testing"

func TestIconStyleSuffix(t *testing.T) {
	t.Run("returns empty string for outline style", func(t *testing.T) {
		got := iconStyleSuffix(IconStyleOutline)
		if got != "" {
			t.Errorf("iconStyleSuffix(IconStyleOutline) = %q; want %q", got, "")
		}
	})

	t.Run("returns -fill suffix for fill style", func(t *testing.T) {
		got := iconStyleSuffix(IconStyleFill)
		if got != "-fill" {
			t.Errorf("iconStyleSuffix(IconStyleFill) = %q; want %q", got, "-fill")
		}
	})
}
```

- [ ] **Step 2: Run the test and confirm it fails**

Run: `go test ./src/internal/core/templates/ -run TestIconStyleSuffix -v`
Expected: compile error (`iconStyleSuffix` is undefined). That's the red bar.

- [ ] **Step 3: Implement the helper**

Create `src/internal/core/templates/icons.go`:

```go
package templates

// iconStyleSuffix returns the Bootstrap Icons class-name suffix for a given
// icon style. Outline maps to BI's default (no suffix); Fill maps to "-fill",
// matching BI's `bi-{name}-fill` filled-variant convention.
func iconStyleSuffix(style IconStyle) string {
	if style == IconStyleFill {
		return "-fill"
	}
	return ""
}
```

Note: `IconStyle`, `IconStyleOutline`, and `IconStyleFill` are still defined in `icons.templ` (will stay there in Task 4) — `icons.go` and `icons.templ` share the `templates` package so the helper sees the constants.

- [ ] **Step 4: Run the test and confirm it passes**

Run: `go test ./src/internal/core/templates/ -run TestIconStyleSuffix -v`
Expected: PASS for both subtests.

- [ ] **Step 5: Commit**

```bash
git add src/internal/core/templates/icons.go src/internal/core/templates/icons_test.go
git commit -m "feat(icons): add iconStyleSuffix helper"
```

---

### Task 4: Add the `Icon` primitive alongside the existing per-icon templs

**Files:**
- Modify: `src/internal/core/templates/icons.templ` (add new templ + `Name` field on IconProps; do NOT delete per-icon templs yet)

- [ ] **Step 1: Add `Name` to `IconProps` and add the `Icon` templ at the top of the file**

Edit `src/internal/core/templates/icons.templ`. Replace the existing `IconProps` struct and add the `Icon` templ immediately after it:

```go
type IconProps struct {
  Name  string    // Bootstrap Icons name without the "bi-" prefix, e.g. "collection", "compass", "house"
  Style IconStyle // defaults to IconStyleOutline (zero value)
}

templ Icon(props IconProps) {
  <i class={ "bi bi-" + props.Name + iconStyleSuffix(props.Style) }></i>
}
```

The existing per-icon templ functions (`WarningIcon`, `CheckIcon`, etc.) stay in place for now. Adding `Name` is backwards-compatible — existing call sites use `IconProps{Style: ...}` without `Name`, which is just `Name: ""` and ignored by the per-icon templs.

- [ ] **Step 2: Regenerate templ output**

Run: `task build/templ`
Expected: clean exit, `icons_templ.go` regenerated. The build still passes because per-icon templs are unchanged.

- [ ] **Step 3: Run the unit test again to confirm still passing**

Run: `go test ./src/internal/core/templates/ -v`
Expected: `TestIconStyleSuffix` PASS.

- [ ] **Step 4: Verify the Go build**

Run: `go build ./src/...`
Expected: clean exit.

- [ ] **Step 5: Commit**

```bash
git add src/internal/core/templates/icons.templ src/internal/core/templates/icons_templ.go
git commit -m "feat(icons): add Icon primitive (per-icon templs still in place)"
```

---

### Task 5: Migrate the five small call-site files

Five files with one or two call sites each. Each file is one self-contained edit. Group as one task because the changes are mechanical.

**Files:**
- Modify: `src/internal/library/adapters/views/album_rating_history_frag.templ:48`
- Modify: `src/internal/library/adapters/views/album_detail_page.templ:112`
- Modify: `src/internal/library/adapters/views/format_icon_frag.templ:12-18`
- Modify: `src/internal/review/adapters/views/rating_confirm_form_frag.templ:49`
- Modify: `src/internal/library/adapters/views/album_row_tags_section_frag.templ:38,55`

- [ ] **Step 1: Migrate `album_rating_history_frag.templ` line 48**

Replace:
```go
@templates.TrashIcon(templates.IconProps{})
```
with:
```go
@templates.Icon(templates.IconProps{Name: "trash"})
```

(BI's `bi-trash` is visually different from the Heroicons-style trash currently rendered. This is an accepted visual change per the spec.)

- [ ] **Step 2: Migrate `album_detail_page.templ` line 112**

Replace:
```go
@templates.TagIcon(templates.IconProps{})
```
with:
```go
@templates.Icon(templates.IconProps{Name: "tag"})
```

- [ ] **Step 3: Migrate `format_icon_frag.templ` lines 12–18**

Replace each of the four lines:
```go
@templates.VinylIcon(templates.IconProps{})    → @templates.Icon(templates.IconProps{Name: "vinyl"})
@templates.CDIcon(templates.IconProps{})       → @templates.Icon(templates.IconProps{Name: "disc"})
@templates.CassetteIcon(templates.IconProps{}) → @templates.Icon(templates.IconProps{Name: "cassette"})
@templates.DigitalIcon(templates.IconProps{})  → @templates.Icon(templates.IconProps{Name: "file-music"})
```

- [ ] **Step 4: Migrate `rating_confirm_form_frag.templ` line 49**

Replace:
```go
@templates.QuestionMarkIcon(templates.IconProps{})
```
with:
```go
@templates.Icon(templates.IconProps{Name: "question-circle"})
```

(Visual change: BI's `question-circle` replaces the Heroicons-style question mark.)

- [ ] **Step 5: Migrate `album_row_tags_section_frag.templ` lines 38 and 55**

Replace:
```go
@templates.PenIcon(templates.IconProps{})  → @templates.Icon(templates.IconProps{Name: "pen"})
@templates.TagIcon(templates.IconProps{})  → @templates.Icon(templates.IconProps{Name: "tag"})
```

- [ ] **Step 6: Regenerate templ and rebuild**

```bash
task build/templ
go build ./src/...
```
Expected: clean exit on both.

- [ ] **Step 7: Visual smoke**

Run `task dev`, browse to a page that exercises each icon (album row → trash + tag + pen icons, format icon dropdown, rating confirm modal). Confirm icons render. Stop the server.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat(icons): migrate small call sites to Icon primitive"
```

---

### Task 6: Migrate `library_header_bar_frag.templ` (preserves outline/fill nav convention)

**Files:**
- Modify: `src/internal/library/adapters/views/library_header_bar_frag.templ:24,28,35,39,52`

This file carries the outline/fill convention for the nav (`Style: IconStyleFill` for active, `Style: IconStyleOutline` for inactive). Preserve that pattern through the migration.

- [ ] **Step 1: Replace each call site, preserving Style**

Five replacements:

```go
// Line 24 (active library)
@templates.CollectionIcon(templates.IconProps{Style: templates.IconStyleFill})
→ @templates.Icon(templates.IconProps{Name: "collection", Style: templates.IconStyleFill})

// Line 28 (inactive library)
@templates.CollectionIcon(templates.IconProps{Style: templates.IconStyleOutline})
→ @templates.Icon(templates.IconProps{Name: "collection", Style: templates.IconStyleOutline})

// Line 35 (active discover)
@templates.CompassIcon(templates.IconProps{Style: templates.IconStyleFill})
→ @templates.Icon(templates.IconProps{Name: "compass", Style: templates.IconStyleFill})

// Line 39 (inactive discover)
@templates.CompassIcon(templates.IconProps{Style: templates.IconStyleOutline})
→ @templates.Icon(templates.IconProps{Name: "compass", Style: templates.IconStyleOutline})

// Line 52 (user dropdown)
@templates.UserIcon(templates.IconProps{Style: templates.IconStyleOutline})
→ @templates.Icon(templates.IconProps{Name: "person-circle"})
```

Note on the user line: `person-circle` doesn't have a fill variant in BI; passing `IconStyleOutline` would render `bi-person-circle` correctly anyway, but since this is an "icon with a single style" (per the spec's single-variant rule), omit the `Style` field rather than passing it.

- [ ] **Step 2: Regenerate templ**

```bash
task build/templ
```
Expected: clean exit.

- [ ] **Step 3: Visual verify the active/inactive state still renders correctly**

Run `task dev`, browse to `/app/library/dashboard` (library nav active) and `/app/library/discover` (discover active). Confirm: active icon renders filled (subtly via the existing `opacity-30` on the wrapper); inactive icon renders outline. The user dropdown icon renders.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat(icons): migrate header bar to Icon primitive (preserves outline/fill nav convention)"
```

---

### Task 7: Migrate `feeds_dropdown_frag.templ` (Spinner needs special handling)

**Files:**
- Modify: `src/internal/library/adapters/views/feeds_dropdown_frag.templ:43,101,105,109,113`

The `SpinnerIcon` currently bakes `class="bi bi-arrow-repeat animate-spin"` directly. The new `Icon` primitive doesn't accept extra classes. Wrap it in a span at the call site.

- [ ] **Step 1: Replace the four straightforward call sites**

```go
// Line 43
@templates.DatabaseIcon(templates.IconProps{})
→ @templates.Icon(templates.IconProps{Name: "database"})

// Line 101
@templates.WarningIcon(templates.IconProps{Style: templates.IconStyleOutline})
→ @templates.Icon(templates.IconProps{Name: "exclamation-triangle"})
// Style is the zero value (Outline), so omit it.

// Line 109
@templates.XMarkIcon(templates.IconProps{Style: templates.IconStyleOutline})
→ @templates.Icon(templates.IconProps{Name: "x-circle"})

// Line 113
@templates.CheckIcon(templates.IconProps{Style: templates.IconStyleOutline})
→ @templates.Icon(templates.IconProps{Name: "check-circle"})
```

- [ ] **Step 2: Replace the spinner with a wrapped Icon**

Line 105:
```go
@templates.SpinnerIcon(templates.IconProps{})
```
becomes:
```go
<span class="animate-spin inline-block">
  @templates.Icon(templates.IconProps{Name: "arrow-repeat"})
</span>
```

(`inline-block` ensures the span has dimensions for the rotation; without it the rotation may not be visible on inline content.)

- [ ] **Step 3: Regenerate templ**

```bash
task build/templ
```

- [ ] **Step 4: Visual verify**

Run `task dev`. Trigger a feed sync to see the spinner animate. Verify warning/x-mark/check icons render in feed status states. Verify the database icon renders next to "Feed Status".

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat(icons): migrate feeds dropdown to Icon primitive (spinner uses animate-spin wrapper)"
```

---

### Task 8: Delete per-icon templs and dead-code icon definitions

**Files:**
- Modify: `src/internal/core/templates/icons.templ` (strip down to types + Icon templ)

After Task 7, no call site references any of the per-icon templ functions. Delete them all.

- [ ] **Step 1: Confirm no remaining references**

```bash
grep -rn --include="*.templ" -E "templates\.(Warning|Check|XMark|Spinner|Home|User|Database|Vinyl|CD|Cassette|Notes|Pen|Trash|QuestionMark|EllipsisVertical|Digital|Collection|Tag|Compass)Icon\(" src
```
Expected: no output (no matches).

- [ ] **Step 2: Replace `icons.templ` with the minimal version**

The full new file:

```go
package templates

type IconStyle int

const (
  IconStyleOutline IconStyle = iota // bi-{name} — BI's default is outline
  IconStyleFill                     // bi-{name}-fill
)

type IconProps struct {
  Name  string    // Bootstrap Icons name without the "bi-" prefix, e.g. "collection", "compass", "house"
  Style IconStyle // defaults to IconStyleOutline (zero value)
}

templ Icon(props IconProps) {
  <i class={ "bi bi-" + props.Name + iconStyleSuffix(props.Style) }></i>
}
```

- [ ] **Step 3: Regenerate templ and rebuild**

```bash
task build/templ
go build ./src/...
```
Expected: clean exit.

- [ ] **Step 4: Run unit tests**

```bash
task test/unit
```
Expected: all PASS, including `TestIconStyleSuffix`.

- [ ] **Step 5: Run e2e tests**

Start the server: `task dev` in one terminal. In another:
```bash
task test/e2e
```
Expected: all PASS. (If e2e tests assert on icon DOM specifically, expect updates — flag any failures and fix before continuing.)

- [ ] **Step 6: Commit**

```bash
git add src/internal/core/templates/icons.templ src/internal/core/templates/icons_templ.go
git commit -m "feat(icons): remove per-icon templs, leaving only Icon primitive"
```

---

# Stream B — Colors

### Task 9: Add element-state utility classes to main.css

**Files:**
- Modify: `static/src/main.css` (append after the `[data-theme="wax"]` block, before the existing `.font-brand` definition)

- [ ] **Step 1: Add the two utility classes**

Insert into `static/src/main.css`:

```css
.is-disabled {
    opacity: 0.5;
    cursor: not-allowed;
    pointer-events: none;
}

.hover-fade {
    transition: opacity 150ms;
}
.hover-fade:hover {
    opacity: 0.8;
}
```

Place them after the existing `[x-cloak]` and `[data-theme="wax"]` blocks. Order in the file: theme block → element-state utilities → font / animation utilities (existing).

- [ ] **Step 2: Recompile Tailwind**

```bash
task build/tailwind
```
Expected: clean exit, `static/public/main.css` regenerated.

- [ ] **Step 3: Verify the classes appear in the compiled CSS**

```bash
grep -E "^\.is-disabled|^\.hover-fade" static/public/main.css
```
Expected: matches for both `.is-disabled` and `.hover-fade`.

- [ ] **Step 4: Commit**

```bash
git add static/src/main.css static/public/main.css
git commit -m "feat(css): add .is-disabled and .hover-fade element-state utilities"
```

---

### Task 10: Add text emphasis `@utility` blocks to main.css

**Files:**
- Modify: `static/src/main.css` (append after element-state utilities)

- [ ] **Step 1: Add the four `@utility` blocks**

Insert into `static/src/main.css`, after the `.hover-fade` block:

```css
@utility text-default {
    color: var(--color-base-content);
}

@utility text-muted {
    color: color-mix(in oklab, var(--color-base-content) 70%, transparent);
}

@utility text-subtle {
    color: color-mix(in oklab, var(--color-base-content) 40%, transparent);
}

@utility text-ghost {
    color: color-mix(in oklab, var(--color-base-content) 20%, transparent);
}
```

These register as Tailwind v4 first-class utilities, so variants like `hover:text-muted`, `md:text-subtle`, etc. work automatically.

- [ ] **Step 2: Recompile Tailwind**

```bash
task build/tailwind
```
Expected: clean exit.

- [ ] **Step 3: Verify by writing a sanity-check in any templ and recompiling**

Pick any templ file currently using `text-base-content/60`. Temporarily replace one usage with `text-muted` and rebuild:
```bash
task build/templ && go build ./src/...
```
Then run `task dev` and visually compare — `text-muted` should render at the same emphasis as the previous `text-base-content/60`. Revert the temporary edit before committing.

- [ ] **Step 4: Commit**

```bash
git add static/src/main.css static/public/main.css
git commit -m "feat(css): add text-default/muted/subtle/ghost @utility blocks"
```

---

### Task 11: Migrate raw `opacity-NN` and `hover:opacity-NN` sites

**Files (per the earlier grep):**
- Modify: `src/internal/library/adapters/views/library_header_bar_frag.templ:23,34`
- Modify: `src/internal/library/adapters/views/feeds_dropdown_frag.templ:80,83,118`
- Modify: `src/internal/library/adapters/views/radar_carousel_frag.templ:47`
- Modify: `src/internal/library/adapters/views/format_icon_frag.templ:25,31`
- Modify: `src/internal/library/adapters/views/formats_modal_frag.templ:40,56`
- Modify: `src/internal/library/adapters/views/album_score_badge_frag.templ:58`
- Modify: `src/internal/library/adapters/views/carousel_section_frag.templ:32,42,95`
- Modify: `src/internal/library/adapters/views/albums_list_frag.templ:84,137,281,296`
- Modify: `src/internal/tags/adapters/views/tags_form_frag.templ:133,134,155`
- Modify: `src/internal/library/adapters/views/album_detail_page.templ:64`

**Mapping rule:**

| Current | Becomes | Reason |
|---|---|---|
| `opacity-30 pointer-events-none` (active nav state) | `is-disabled` | Disabled-style appearance |
| `opacity-50 cursor-not-allowed` (formats modal disabled item) | `is-disabled` | Same |
| `hover:opacity-80 transition-opacity` (cards/links) | `hover-fade` | Hover affordance |
| `opacity-70` / `opacity-20` (format-owned vs not-owned, on a styling div around an icon) | `text-muted` / `text-ghost` | Boolean dim — content emphasis, not element state |
| `opacity-60` / `opacity-50` (text labels) | text scale (`text-muted` / `text-subtle`) | Text emphasis |
| `opacity-20 hover:opacity-100` (album_score_badge — visible on hover only) | leave the bare opacity for now and **flag for human review** — this is a "reveal-on-hover" pattern not covered by the new vocabulary | The two-narrow-roles rule doesn't admit this case. Flag in a brief comment for post-merge follow-up. |

- [ ] **Step 1: Walk each file and apply the mapping table**

For each file in the list above, open it and rewrite each `opacity-NN`/`hover:opacity-NN` to the mapped utility. Where the original element has more than one of these classes (e.g. `opacity-50 cursor-not-allowed` plus other classes), drop the redundant atoms when the utility class supersedes them — `is-disabled` already includes `cursor: not-allowed` and `pointer-events: none`, so don't keep those alongside it. **Where the migration mismatches the table — e.g. an `opacity-50` that's actually for "this is dimmed on purpose, not disabled" — flag the site in a code comment (`// TODO(design-system): review — what is this dimming for?`) and pick the closest utility rather than mechanically applying.**

- [ ] **Step 2: Regenerate templ**

```bash
task build/templ
```
Expected: clean exit.

- [ ] **Step 3: Visual verify**

Run `task dev`. Browse pages that exercise the migrated sites: header bar (active nav looks disabled-styled), formats modal (disabled item dimmed), albums list (rows hover-fade), tags form (chip remove buttons hover-revealed). Note any visual regressions and fix before committing.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor(css): replace raw opacity-NN with .is-disabled / .hover-fade utilities"
```

---

### Task 12: Migrate `text-base-content/NN` sites to the four named utilities

**Files (per the earlier grep — 19 files):**

```
src/internal/library/adapters/views/album_actions_modal_frag.templ
src/internal/library/adapters/views/album_detail_page.templ
src/internal/library/adapters/views/album_rating_history_frag.templ
src/internal/library/adapters/views/album_row_tags_section_frag.templ
src/internal/library/adapters/views/album_score_readout_frag.templ
src/internal/library/adapters/views/album_tags_cell_frag.templ
src/internal/library/adapters/views/albums_list_frag.templ
src/internal/library/adapters/views/carousel_section_frag.templ
src/internal/library/adapters/views/discogs_release_details_frag.templ
src/internal/library/adapters/views/discogs_search_results_frag.templ
src/internal/library/adapters/views/discover_search_results_frag.templ
src/internal/library/adapters/views/formats_modal_frag.templ
src/internal/library/adapters/views/radar_carousel_frag.templ
src/internal/library/adapters/views/sleeve_notes_editor_frag.templ
src/internal/library/adapters/views/sleeve_notes_section_frag.templ
src/internal/review/adapters/views/base_questions_form_frag.templ
src/internal/review/adapters/views/rating_confirm_form_frag.templ
src/internal/review/adapters/views/rerate_prompt_frag.templ
src/internal/tags/adapters/views/tags_form_frag.templ
```

**Mapping rule:**

| Current | Becomes | Notes |
|---|---|---|
| `text-base-content` (no opacity) | `text-default` | The full-emphasis voice of the page |
| `text-base-content/80` | `text-muted` | Round up to /70 (closer to default) |
| `text-base-content/70` | `text-muted` | Direct |
| `text-base-content/60` | `text-muted` | Round up to /70 |
| `text-base-content/50` | `text-subtle` | Round down to /40 |
| `text-base-content/40` | `text-subtle` | Direct |
| `text-base-content/30` | `text-subtle` or `text-ghost` | Pick by intent: muted-but-readable → subtle; barely-visible → ghost |
| `text-base-content/20` | `text-ghost` | Direct |

**Misuse-flag rule (from the spec):** when the resulting role label disagrees with the element's purpose — e.g. `<h1 class="text-base-content/50">…` collapsing to `text-subtle` for a heading — that's a probable misuse of the original emphasis level, not a migration miss. Add a brief code comment (`// TODO(design-system): review — heading rendered at subtle emphasis?`) and proceed with the closest scale stop. Do not silently demote/promote.

The compound class on `sleeve_notes_section_frag.templ:43` (`[&_a]:text-info ... [&_a:hover]:text-info/70`) is a brand-color hover idiom (per the spec's "Brand-colored text" carve-out) and stays unchanged.

- [ ] **Step 1: Walk each file and apply the mapping**

For each file in the list, replace every `text-base-content/NN` occurrence with the mapped utility. Keep `text-default` for bare `text-base-content` only where the current code explicitly writes `text-base-content` — don't add it where the class is omitted (the default text color is already `base-content` via DaisyUI).

Pay attention to heading elements (`<h1>`, `<h2>`, etc.), button labels, and other "this should be the page voice" elements — these should land at `text-default` and any current `/40` or `/50` on them is the misuse case to flag.

- [ ] **Step 2: Regenerate templ and rebuild**

```bash
task build/templ
go build ./src/...
```
Expected: clean exit on both.

- [ ] **Step 3: Visual smoke**

Run `task dev`. Browse album list, album detail, tags page, review form, sleeve notes view — confirm text hierarchy still reads correctly. Look for elements that now stand out wrongly (an element that became too dim or too bright after migration).

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor(css): collapse text-base-content/NN to text-default/muted/subtle/ghost utilities"
```

---

### Task 13: Land first use of `accent` (brand wordmark)

**Files:**
- Modify: `src/internal/library/adapters/views/library_header_bar_frag.templ:18`

The wordmark is the highest-leverage spot to introduce `accent` — it's the brand expression on every page, currently `text-primary`. Per the design-system roles, `accent` is for "decorative highlights and 'glow' moments — brand flourishes." Swapping the wordmark is a defensible first use.

- [ ] **Step 1: Change the wordmark color**

Replace:
```html
<a href="/" class="text-lg text-primary font-brand">wax</a>
```
with:
```html
<a href="/" class="text-lg text-accent font-brand">wax</a>
```

- [ ] **Step 2: Regenerate templ**

```bash
task build/templ
```

- [ ] **Step 3: Visual verify**

Run `task dev`. Open any page. Confirm the "wax" wordmark in the header now renders in the lighter amber (`#e09a4f`) rather than the warmer amber (`#c97d2e`). Confirm contrast against `bg-base-100` is still legible.

If the visual reads wrong (e.g. accent turns out to be too bright or too similar to primary in this position), revert and pick a different first-use site (e.g. the ticker animation or a small decorative flourish elsewhere). Note the choice in the commit message.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat(css): land first use of accent token on brand wordmark"
```

---

# Stream C — Documentation

### Task 14: Update `docs/design/design-system.md`

**Files:**
- Modify: `docs/design/design-system.md`

The doc shifts from "what tokens exist" to "what the visual direction is, how each token is used, and where to find the implementations." No verbatim CSS or Go in this file — it names utilities and points to where they live.

- [ ] **Step 1: Add the visual-direction prelude**

After the existing `# Design System` heading and the introductory paragraph, insert a new paragraph:

```markdown
The visual direction is analog warmth: dark, warm-toned surfaces (browns rather than grays), amber-orange accents (the wax of a sealed record sleeve, the glow of warm light), light text on dark surfaces by default, and intentional motion where motion serves meaning. The tokens and utilities below are how that direction is enacted.
```

- [ ] **Step 2: Replace the existing Foundation/Typography/Animations sections with the new structure**

Restructure the body to:

```markdown
## Foundation

(keep the existing Foundation paragraph about Tailwind + DaisyUI + the wax theme)

## Icons

Wax uses **Bootstrap Icons** (MIT, ~2000 icons) as its single icon source. The vendored CSS and font live under `static/`; the layout primitive loads them. All icons in the app are emitted by the single `Icon` primitive in `core/templates/icons.templ` — call sites pass the BI catalog name (without the `bi-` prefix) and an optional `IconStyle` (Outline or Fill). Sizing comes from the parent's `text-{size}`; color comes from the parent's text color (BI inherits `currentColor`).

**Outline / fill convention.** Outline is the default presentation; Fill marks the current page or selected state. Used wherever a UI surface has a paired notion of "this one vs the others" (today: the nav header). Most icons are decorative or single-meaning — for those, leave `Style` at its default.

**Single-variant icons.** Some BI icons exist in only one style (`vinyl`, `cassette`, `arrow-repeat`, etc.). Check BI's catalog before passing `Fill` for an icon name; if no `-fill` variant exists, omit the prop.

See `core/templates/icons.templ` for the primitive's signature.

## Colors

The wax theme defines three groups of tokens, each with a distinct role.

### Surfaces (backgrounds, borders, dividers)

Stratified by elevation:
- `bg-base-100` — page background. The default canvas.
- `bg-base-200` — raised surface (cards, panels, dropdowns, modal bodies).
- `bg-base-300` — highest elevation (hovered rows, pressed states, the topmost layer); also the default border color (`border-base-300`).

### Brand tones (emphasis, identity, decorative chrome)

Each token has a defined role; reach for the role, not the color that "looks right" in isolation.

- **`primary`** — interactive emphasis. Links, brand wordmark, selected/active text states. Reserved; scarcity is the point.
- **`accent`** — decorative highlights and "glow" moments. Brand flourishes, animated chrome.
- **`secondary`** — supporting actions and tags-domain affordances. Weighted-but-not-primary.
- **`neutral`** — chrome that isn't a surface and isn't a brand expression: tooltips, kbd hints, neutral badges.

Each tone has a paired `-content` token for legible text **on** that color. Always use the pair; never put `text-base-content` on a brand background.

### Semantic (status only)

- `info` — neutral informational status.
- `success` — completed actions, positive validation.
- `warning` — recoverable problems.
- `error` — failed actions, validation errors, **and destructive actions** (delete buttons, remove-tag, irreversible CTAs).

### Text emphasis scale

Four named utilities express the text hierarchy. Use the named utility, not raw `text-base-content/NN`.

| Utility | Role |
|---|---|
| `text-default` | Body copy, headings, primary values — the voice of the page. |
| `text-muted` | Section labels, captions, supporting meta-context. |
| `text-subtle` | Timestamps, helper text, low-priority metadata. |
| `text-ghost` | Placeholders, empty-state hints, dimmed icons. |

Brand-colored text (`text-primary`, `text-error`, etc.) is a separate mechanism — emphasis by color, not by hierarchy. Brand colors don't get the four-stop scale.

### Element opacity

Two narrow roles, each wrapped as a utility class. Raw `opacity-NN` on a whole element should not appear in templ markup outside these.

- `.is-disabled` — disabled state (whole element non-interactive). Pairs the visual cue with `cursor: not-allowed` and `pointer-events: none` so they can't drift apart.
- `.hover-fade` — hover affordance on a whole-element block (cards, link-wrapped media). Don't layer onto buttons or controls where DaisyUI handles the hover.

See `static/CLAUDE.md` for the verbatim CSS definitions.

## Typography

(keep the existing Typography section unchanged)

## Animations

(keep the existing Animations section unchanged)
```

- [ ] **Step 3: Update the "When to add to `main.css`" section**

Find the existing bullet list in that section and add a fourth bullet:

```markdown
- **A named-role wrapper around theme-semantic atoms is needed.** E.g. text emphasis utilities, state utilities like `.is-disabled`. Bare single-class wraps (`text-amber` for `text-primary`) are still not added.
```

- [ ] **Step 4: Update the Client-side libraries section if needed**

The Client-side libraries section currently mentions HTMX, idiomorph, Alpine.js. Add a note that Bootstrap Icons CSS is also loaded by the layout (not a client-side library in the interaction sense, but a third-party stylesheet that affects every page):

Append at the end of the section:

```markdown
The layout also loads **Bootstrap Icons** as a third-party stylesheet for icon rendering — see the Icons section above.
```

- [ ] **Step 5: Verify no verbatim CSS/Go leaked into the doc**

```bash
grep -E "color-mix|@utility|opacity: 0|cursor: not-allowed|templ Icon\(" docs/design/design-system.md
```
Expected: no matches. The doc names utilities, doesn't define them.

- [ ] **Step 6: Commit**

```bash
git add docs/design/design-system.md
git commit -m "docs(design): add visual direction, icons, colors, and named utility sections"
```

---

### Task 15: Update `docs/design/principles.md` (theme-tokens addendum)

**Files:**
- Modify: `docs/design/principles.md` (the "Theme tokens, not raw colors" section)

- [ ] **Step 1: Append the addendum to the existing principle**

Find the section "Theme tokens, not raw colors" (currently a single paragraph). Append:

```markdown
The text emphasis scale is `text-default`, `text-muted`, `text-subtle`, `text-ghost`; the element-state utilities are `.is-disabled` and `.hover-fade`. Raw `text-base-content/NN` and raw `opacity-NN` should not appear in templ markup outside these utilities. See `design-system.md` for the role of each utility and `static/CLAUDE.md` for the definitions.
```

- [ ] **Step 2: Commit**

```bash
git add docs/design/principles.md
git commit -m "docs(principles): name the design-system utilities in the theme-tokens principle"
```

---

### Task 16: Create `static/CLAUDE.md` (verbatim catalog)

**Files:**
- Create: `static/CLAUDE.md`

This is the new home for verbatim CSS and the BI vendoring note. It's auto-loaded when working under `static/`, which is the right scope.

- [ ] **Step 1: Create the file**

```markdown
# static/ — frontend assets (singleton)

This directory holds the application's frontend asset pipeline.

- `src/` — sources Tailwind compiles. Today: `main.css` (the wax theme + the wax-specific utilities catalogued below).
- `public/` — files served at `/static/*` by the server. Tailwind's compiled `main.css` lands here, alongside vendored third-party assets (HTMX, Bootstrap Icons + font, etc.) and brand assets (favicon, manifest, ticker JS).

`static/src/main.css` is the source of truth for the wax theme and for all wax-specific utilities and keyframes. Do not add per-page or per-component styling here — templ files use Tailwind utilities directly.

## Wax theme tokens

The `[data-theme="wax"]` block in `main.css` defines the semantic color tokens (base / primary / secondary / accent / neutral / info / success / warning / error, each paired with a `-content` variant), corner radii, and the dark color scheme. See `main.css` for current values.

## Element-state utilities

Two utility classes wrap the legitimate uses of raw element opacity. Use the class name in markup; never duplicate the underlying atoms.

```css
.is-disabled {
    opacity: 0.5;
    cursor: not-allowed;
    pointer-events: none;
}

.hover-fade {
    transition: opacity 150ms;
}
.hover-fade:hover {
    opacity: 0.8;
}
```

`.is-disabled` is for non-form elements that should appear disabled (form controls use the HTML `disabled` attribute and DaisyUI's existing styles instead). `.hover-fade` is for whole-element click targets (cards, link-wrapped media) where the entire surface should react to hover.

## Text emphasis utilities

Four named utilities express the text-on-base-content hierarchy. Defined as Tailwind v4 `@utility` blocks so variants like `hover:text-muted` work.

```css
@utility text-default {
    color: var(--color-base-content);
}

@utility text-muted {
    color: color-mix(in oklab, var(--color-base-content) 70%, transparent);
}

@utility text-subtle {
    color: color-mix(in oklab, var(--color-base-content) 40%, transparent);
}

@utility text-ghost {
    color: color-mix(in oklab, var(--color-base-content) 20%, transparent);
}
```

Roles: see `docs/design/design-system.md` "Text emphasis scale."

## Bootstrap Icons (vendored)

`bootstrap-icons.css` (vendored to `static/public/`) and `fonts/bootstrap-icons.woff2` (vendored to `static/public/fonts/`) are loaded by the root layout primitive (`core/templates/root.templ`). The application's icon primitive (`core/templates/Icon`) emits `<i class="bi bi-{name}{-fill}?">` and relies on this stylesheet.

When updating BI's pinned version, replace both files at the same time and verify the CSS still references the font via `./fonts/bootstrap-icons.woff2`.

## Custom keyframes

`main.css` defines a `ticker-scroll` keyframe used by the ticker primitive. Add a brief note here when a new keyframe lands; keyframes are global and should not multiply.

## After editing

- After editing `static/src/main.css`: run `task build/tailwind` to regenerate `static/public/main.css`.
```

- [ ] **Step 2: Commit**

```bash
git add static/CLAUDE.md
git commit -m "docs(static): add CLAUDE.md cataloguing wax-specific CSS additions"
```

---

### Task 17: Update `src/internal/core/templates/CLAUDE.md` (Icon primitive note)

**Files:**
- Modify: `src/internal/core/templates/CLAUDE.md`

- [ ] **Step 1: Add a short note about the Icon primitive**

Append to the existing CLAUDE.md, before the "After editing" section:

```markdown
## The Icon primitive

`icons.templ` defines the single `Icon` primitive, which wraps Bootstrap Icons. Pass a BI catalog name (without the `bi-` prefix) and an optional `IconStyle` (Outline | Fill). Sizing and color come from the parent (`text-{size}` for size, parent text color for color). The CSS that powers it is vendored under `static/public/`; the BI catalog lives at https://icons.getbootstrap.com/.
```

- [ ] **Step 2: Commit**

```bash
git add src/internal/core/templates/CLAUDE.md
git commit -m "docs(core/templates): note the Icon primitive's design"
```

---

# Wrap-up

### Task 18: Final test pass

- [ ] **Step 1: Run all unit tests**

```bash
task test/unit
```
Expected: all PASS.

- [ ] **Step 2: Run e2e tests**

Start the server in one terminal: `task dev`. In another:
```bash
task test/e2e
```
Expected: all PASS. If anything fails, fix before squashing.

- [ ] **Step 3: Manual visual tour**

Browse the major surfaces with `task dev`:
- `/app/library/dashboard` — header bar (active library nav), albums list (text emphasis, hover-fade rows, format icons)
- `/app/library/discover` — header bar (active discover nav), discover search results (text emphasis)
- An album detail page — tag chips (secondary color), tags section, sleeve notes view (text-info hover idiom)
- The formats modal — disabled item (`is-disabled`), owned vs not-owned format icons (`text-muted` / `text-ghost`)
- The feeds dropdown — sync spinner (animate-spin wrapper), feed status icons
- The review form — confirm modal with question icon

Look for:
- Icons rendering correctly in BI style (not blank, not the wrong shape)
- Active nav still visually distinct from inactive nav
- Brand wordmark in lighter amber
- Text hierarchy reads coherently (no element jumps out at the wrong emphasis)
- Hover affordances feel right (cards fade, buttons don't double-react)

Note any regressions in a follow-up note; fix before squashing if they're caused by this branch.

---

### Task 19: Squash to three commits per the spec

The spec calls for three commits at PR level: Stream A (icons), Stream B (colors), Stream C (docs).

- [ ] **Step 1: Confirm current commit count**

```bash
git log --oneline main..HEAD
```
Expected: one commit per task plus the initial spec commit (if it's on this branch — it's actually on `spec/design-system-icons-colors`, this work is on the feature branch). Adjust the base ref accordingly. The exact commit count depends on which task commits were made; expect roughly 17–18.

- [ ] **Step 2: Interactive rebase to squash by stream**

Note: this environment doesn't support `git rebase -i`. Use a non-interactive approach: a soft reset to the merge base, then three commits.

```bash
# Find the base
BASE=$(git merge-base main HEAD)

# Stream A — squash all icon commits
git reset --soft $BASE
git add static/public/bootstrap-icons.css static/public/fonts/ \
        src/internal/core/templates/root.templ \
        src/internal/core/templates/root_templ.go \
        src/internal/core/templates/icons.go \
        src/internal/core/templates/icons.templ \
        src/internal/core/templates/icons_templ.go \
        src/internal/core/templates/icons_test.go \
        src/internal/library/adapters/views/album_rating_history_frag.templ \
        src/internal/library/adapters/views/album_detail_page.templ \
        src/internal/library/adapters/views/library_header_bar_frag.templ \
        src/internal/library/adapters/views/feeds_dropdown_frag.templ \
        src/internal/library/adapters/views/format_icon_frag.templ \
        src/internal/review/adapters/views/rating_confirm_form_frag.templ \
        src/internal/library/adapters/views/album_row_tags_section_frag.templ
# Plus all the *_templ.go siblings regenerated by task build/templ
git add 'src/**/*_templ.go'
git commit -m "feat: migrate icons to a single Icon primitive backed by Bootstrap Icons"

# Stream B — colors (everything in static/src/, the call-site templ edits not yet committed)
git add static/src/main.css static/public/main.css \
        'src/**/*.templ' 'src/**/*_templ.go'
git commit -m "feat: text emphasis scale, .is-disabled, .hover-fade utilities + per-site migration"

# Stream C — docs
git add docs/design/design-system.md docs/design/principles.md \
        static/CLAUDE.md src/internal/core/templates/CLAUDE.md
git commit -m "docs: design-system updates for icons + colors; verbatim catalog in static/CLAUDE.md"
```

If any files end up staged in the wrong commit (because templ-generated files cross stream boundaries), use `git reset HEAD~3` to undo and re-stage more carefully.

- [ ] **Step 3: Verify the final commit shape**

```bash
git log --oneline main..HEAD
```
Expected: exactly three commits, in the order A → B → C, each with the message above.

```bash
git log --stat main..HEAD
```
Expected: each commit's file list aligns with its stream's responsibility.

---

### Task 20: Open the PR

- [ ] **Step 1: Push the branch**

```bash
git push -u origin <feature-branch-name>
```

- [ ] **Step 2: Open a PR**

Create a PR from this branch to `main` with a body that summarises the three streams and links to the spec:

```
Implements the design-system spec at docs/superpowers/specs/2026-05-15-design-system-icons-colors-design.md.

Three commits, in order:
- Stream A: migrate icons to a single Icon primitive backed by Bootstrap Icons (replaces 19 hand-inlined SVG templs).
- Stream B: introduce text-default/muted/subtle/ghost utilities and .is-disabled/.hover-fade element-state utilities; per-site migration replaces ad-hoc opacity usage.
- Stream C: documentation — conceptual updates in docs/design/, verbatim CSS catalog in new static/CLAUDE.md.

Sites flagged for design review (look for `// TODO(design-system): review` comments) are non-blocking; address in follow-up.
```

---

## Self-review notes

After writing this plan, I checked it against the spec:

**Spec coverage:** Every section of the spec maps to a task — visual direction (Task 14), Icons API/source/convention/single-variant rule (Tasks 1–8 + 14), Colors palette structure (Task 14), text emphasis scale (Tasks 10 + 12 + 14), element opacity (Tasks 9 + 11 + 14), `static/CLAUDE.md` (Task 16), `principles.md` addendum (Task 15), `core/templates/CLAUDE.md` update (Task 17), three-commit shape (Task 19).

**Type consistency:** `Icon` primitive's signature is consistent across Tasks 4, 5, 6, 7, 8 — `templates.Icon(templates.IconProps{Name: "...", Style: ...})`, with `Style` omitted for the zero value. `iconStyleSuffix` helper is named consistently across Tasks 3, 4, 8.

**Placeholder scan:** No "TBD" / "implement later" / "similar to Task N" — each step shows the exact code or command. The misuse-flag rule (Tasks 11, 12) is genuinely a per-site judgment by design (the spec calls for human-flagged review, not mechanical mapping); the plan instructs the agent to add a code comment rather than guess silently.

**Open spec items addressed in the plan:**
- The spec said BI lives in `static/`; the plan refines this to `static/public/` (where the server actually serves from) — confirmed by inspecting `start.go`.
- The spec said `static/src/CLAUDE.md`; the plan places it at `static/CLAUDE.md` so it can document both `src/` and `public/` (BI vendoring lives in `public/`). This is a small refinement to the spec, noted in Stream C of the spec body indirectly.
