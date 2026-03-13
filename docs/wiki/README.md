# Wiki

The wiki is the source of truth for the product and architecture of Wax. It captures what the product is, where it's going, how it's structured, and why decisions were made.

## Scope

The wiki covers **product** and **architecture** — not code implementation. It should be possible to understand Wax completely from the wiki without reading any source code. Implementation details (specific functions, SQL queries, file paths) do not belong here; the code is the source of truth for those.

## Structure

```
wiki/
├── README.md       ← this file
├── wiki.md         ← entry point and graph index
└── pages/          ← one file per node
```

`wiki.md` is the only file at the top level. All content lives in `pages/`.

## Graph Format

The wiki is a graph, not a hierarchy. Pages link to each other based on relevance — there is no tree of parent/child topics. `wiki.md` is the entry point and lists all nodes, but it is not a "root" that owns the pages beneath it.

When a concept connects to another page, link to it inline rather than restating the content.

## Page Conventions

### Frontmatter

Every page must have YAML frontmatter with the following fields:

**`description`** *(required)* — defines the scope of the page. Should answer the question "does this note belong here?" and explicitly state what does *not* belong and where those things go instead.

**`links`** *(optional)* — related pages, as markdown links to the file.

```yaml
---
description: >
  What belongs here and what does not. Should answer the question:
  "does this note belong in this page?"
links:
  - "[page-a](page-a.md)"
  - "[page-b](page-b.md)"
---
```

### Parent link

The first line after frontmatter is a link to the page's logical parent — the page a reader would naturally come from, or the one that provides the most context for this page. This is usually `wiki.md` for top-level nodes, but can be another page. For example, if a feature grows large enough to warrant its own dedicated page, its parent would be `features.md`.

```markdown
[parent-page](../wiki.md)
```

## When to Break Out a New Page

Keep content on an existing page unless one of these is true:

- **The section has grown too large** — if a topic within a page has expanded to the point where it dominates the page or makes it hard to scan, it deserves its own page.
- **The topic is referenced from multiple pages** — if two or more pages need to link to the same concept, that concept should be its own page rather than living inside one of them.
- **The scope is clearly distinct** — if you find yourself writing content that keeps bumping against the boundaries of the current page's `description`, that's a signal the content belongs elsewhere.

When breaking out a page, set the parent link to the page it came from. Only add it to the Pages table in `wiki.md` if it is a top-level page (i.e. its parent is `wiki.md`).

## No Duplication

Information lives in exactly one place. If a concept is relevant to multiple pages, the secondary pages link to the primary rather than restating it.

When deciding where something belongs, read the `description` frontmatter of candidate pages. If a page's description says "does not belong here → see X", put it in X.

## Adding a New Page

1. Create the file in `pages/` with frontmatter, a parent link, and content.
2. If the page's parent is `wiki.md`, add a row to the Pages table in `wiki.md`.
3. Add cross-links in any existing pages that relate to the new one.
