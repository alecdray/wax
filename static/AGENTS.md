# static/ — frontend assets (singleton)

This directory holds the application's frontend asset pipeline.

- `src/` — sources Tailwind compiles. Today: `main.css` (the wax theme + the wax-specific utilities catalogued below).
- `public/` — files served at `/static/*` by the server. Tailwind's compiled `main.css` lands here, alongside vendored third-party assets (HTMX, Bootstrap Icons + font, etc.) and brand assets (favicon, manifest, ticker JS).

`static/src/main.css` is the source of truth for the wax theme and for all wax-specific utilities and keyframes. Do not add per-page or per-component styling here — templ files use Tailwind utilities directly.

## Wax theme tokens

The `[data-theme="wax"]` block in `main.css` defines the semantic color tokens (base / primary / secondary / accent / neutral / info / success / warning / error, each paired with a `-content` variant), corner radii, and the dark color scheme. See `main.css` for current values.

## Element-state utilities

Three utility classes wrap the legitimate uses of raw element opacity. Use the class name in markup; never duplicate the underlying atoms.

```css
.is-disabled {
    opacity: 0.5;
    cursor: not-allowed;
    pointer-events: none;
}

.hover-fade-out {
    transition: opacity 150ms;
}
.hover-fade-out:hover {
    opacity: 0.8;
}

.hover-fade-in {
    opacity: 0.5;
    transition: opacity 150ms;
}
.hover-fade-in:hover {
    opacity: 1;
}
```

`.is-disabled` is for non-form elements that should appear disabled (form controls use the HTML `disabled` attribute and DaisyUI's existing styles instead). `.hover-fade-out` is for whole-element click targets (cards, link-wrapped media) where the entire surface should dim subtly on hover. `.hover-fade-in` is for secondary affordances (chip controls, row-scoped actions) that should be de-emphasized at rest and reveal on hover.

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

## Brand icon

`favicon.svg` is the single source mark — a large amber `--color-primary` brand "w" with a dark extruded edge (depth without lightening the brand colour) on a warm-black tile (`#0e0c0a`, theme base-100). The "w" is the brand font (Instrument Sans 600) **outlined to a `<path>`**, not live `<text>`: an SVG icon doesn't inherit the page's web font, and `rsvg-convert` uses system fonts, so a font-name reference would fall back inconsistently. To change the letter/weight/font, re-outline it (fonttools: instantiate the variable font at the target axes, pull the glyph via `SVGPathPen`, and transform it to centre the bbox). It is served directly as the favicon; the home-screen PNGs (`apple-touch-icon.png` for iOS, `icon-{192,512}.png` for the PWA manifest) are rendered from it by `task build/icons`. Edit `favicon.svg`, rerun the task, and commit the PNGs — the Docker build copies `public/` as-is and has no `rsvg-convert`.

## Bootstrap Icons (vendored)

`bootstrap-icons.css` (vendored to `static/public/`) and `fonts/bootstrap-icons.woff2` (vendored to `static/public/fonts/`) are loaded by the root layout primitive (`core/templates/root.templ`). The application's icon primitive (`core/templates/Icon`) emits `<i class="bi bi-{name}{-fill}?">` and relies on this stylesheet.

When updating BI's pinned version, replace both files at the same time and verify the CSS still references the font via `./fonts/bootstrap-icons.woff2`.

## Custom keyframes

`main.css` defines a `ticker-scroll` keyframe used by the ticker primitive, and a `wax-spin` keyframe used by the action-button busy state (`.btn-busy`). Add a brief note here when a new keyframe lands; keyframes are global and should not multiply.

## After editing

- After editing `static/src/main.css`: run `task build/tailwind` to regenerate `static/public/main.css`.
