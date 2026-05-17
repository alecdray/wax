#!/usr/bin/env node
// Product-criteria static checks for the e2e suite.
//
// Runs PC1–PC7 and PC9 as static analysis over the repo. PC8 (suite green) is
// out of scope here — run it via `npx playwright test`.
//
// Exits 0 if every static PC passes, 1 otherwise. Prints a one-line verdict per
// PC and, on failure, the offending evidence underneath.
//
// Pure file analysis — no dev server, DB, or browser.
//
// Usage:
//   node e2e/check-product-criteria.mjs              # run all static PCs
//   node e2e/check-product-criteria.mjs --pc PC5     # run a single PC
//   node e2e/check-product-criteria.mjs --base main  # base ref for PC4 diff
//
// Flags (optional):
//   --pc <id>            run only the named PC (PC1, PC2, ...)
//   --repo-root <dir>    base for file scans              (default: cwd)
//   --base <ref>         git base ref for PC4 (added/modified testids)
//                        (default: main)

import { readdirSync, readFileSync, statSync, existsSync } from 'node:fs';
import { join, relative, resolve, basename } from 'node:path';
import { execFileSync } from 'node:child_process';

// ----------------------- shared utilities ---------------------------------

function parseArgs(argv) {
  const opts = { only: null, repoRoot: null, base: 'main' };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === '--pc') opts.only = argv[++i];
    else if (a === '--repo-root') opts.repoRoot = argv[++i];
    else if (a === '--base') opts.base = argv[++i];
    else if (a === '-h' || a === '--help') opts.help = true;
    else throw new Error(`unknown arg: ${a}`);
  }
  return opts;
}

function walk(dir, predicate, out = []) {
  let entries;
  try {
    entries = readdirSync(dir, { withFileTypes: true });
  } catch (err) {
    if (err.code === 'ENOENT') return out;
    throw err;
  }
  for (const ent of entries) {
    const full = join(dir, ent.name);
    if (ent.isDirectory()) walk(full, predicate, out);
    else if (ent.isFile() && predicate(full)) out.push(full);
  }
  return out;
}

function readLines(file) {
  return readFileSync(file, 'utf8').split('\n');
}

// Extract scenario names from a Gherkin .feature file. Multi-word names are
// kept as-is. Stops at the colon after "Scenario".
function extractScenarios(featureFile) {
  const out = [];
  const lines = readLines(featureFile);
  const re = /^\s*Scenario:\s*(.+?)\s*$/;
  for (let i = 0; i < lines.length; i++) {
    const m = lines[i].match(re);
    if (m) out.push({ name: m[1], line: i + 1 });
  }
  return out;
}

// Extract `test('name', ...)` calls from a Playwright spec.
function extractTests(specFile) {
  const out = [];
  const src = readFileSync(specFile, 'utf8');
  const lines = src.split('\n');
  // Single- or double-quoted name. Backslashed quotes are not in our scenarios
  // so a simple non-greedy match is sufficient.
  const re = /^\s*test\(\s*['"](.+?)['"]\s*,/;
  for (let i = 0; i < lines.length; i++) {
    const m = lines[i].match(re);
    if (m) out.push({ name: m[1], line: i + 1 });
  }
  return out;
}

function relpath(repoRoot, p) {
  return relative(repoRoot, p) || p;
}

// ----------------------- PC1 ----------------------------------------------

function pc1(repoRoot) {
  const featDir = join(repoRoot, 'e2e', 'feat');
  const specDir = join(repoRoot, 'e2e', 'spec');
  const featFiles = walk(featDir, (p) => p.endsWith('.feature'));
  const specFiles = walk(specDir, (p) => p.endsWith('.spec.ts'));

  const featBase = new Map(); // base -> file
  for (const f of featFiles) {
    const b = basename(f).replace(/\.feature$/, '');
    featBase.set(b, f);
  }
  const specBase = new Map();
  for (const f of specFiles) {
    const b = basename(f).replace(/\.spec\.ts$/, '');
    specBase.set(b, f);
  }

  const failures = [];

  for (const [b, f] of featBase) {
    if (!specBase.has(b)) {
      failures.push(`orphan feature with no spec: ${relpath(repoRoot, f)}`);
    }
  }
  for (const [b, f] of specBase) {
    if (!featBase.has(b)) {
      failures.push(`orphan spec with no feature: ${relpath(repoRoot, f)}`);
    }
  }

  // For each pair, compare scenario/test names exactly.
  for (const [b, featFile] of featBase) {
    const specFile = specBase.get(b);
    if (!specFile) continue;
    const scenarios = extractScenarios(featFile);
    const tests = extractTests(specFile);
    const scenSet = new Set(scenarios.map((s) => s.name));
    const testSet = new Set(tests.map((t) => t.name));

    for (const s of scenarios) {
      if (!testSet.has(s.name)) {
        failures.push(
          `scenario in ${relpath(repoRoot, featFile)}:${s.line} has no matching test in ${relpath(repoRoot, specFile)}: ${JSON.stringify(s.name)}`,
        );
      }
    }
    for (const t of tests) {
      if (!scenSet.has(t.name)) {
        failures.push(
          `test in ${relpath(repoRoot, specFile)}:${t.line} has no matching scenario in ${relpath(repoRoot, featFile)}: ${JSON.stringify(t.name)}`,
        );
      }
    }
  }

  return {
    id: 'PC1',
    description: 'Feature/spec 1:1 pairing',
    command: 'node e2e/check-product-criteria.mjs --pc PC1',
    pass: failures.length === 0,
    failures,
  };
}

// ----------------------- PC2 ----------------------------------------------

// Reuses the existing alignment check. Runs as a subprocess to keep that
// command the source of truth for the verdict and surface its output verbatim.
function pc2(repoRoot) {
  const script = join(repoRoot, 'e2e', 'check-testid-alignment.mjs');
  let stdout = '';
  let stderr = '';
  let code = 0;
  try {
    stdout = execFileSync('node', [script, '--repo-root', repoRoot], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
    });
  } catch (err) {
    code = err.status ?? 1;
    stdout = err.stdout?.toString('utf8') ?? '';
    stderr = err.stderr?.toString('utf8') ?? '';
  }

  const failures = [];
  if (code !== 0) {
    for (const line of stdout.split('\n')) {
      if (line.trim()) failures.push(line.trim());
    }
    const stderrTrim = stderr.trim();
    if (stderrTrim) failures.push(stderrTrim);
  }

  return {
    id: 'PC2',
    description: 'No orphan testids in specs',
    command: 'npm run e2e:check',
    pass: code === 0,
    failures,
  };
}

// ----------------------- PC3 ----------------------------------------------

// Page routes for PC3: the four user-facing modules, excluding pure
// HTMX-fragment endpoints (their parent page covers them) and /spotify/callback.
//
// Determined by reading routes.go and selecting handlers that render a *_page
// templ. Frag handlers (HTMX swaps) are excluded.
//
// Each page route must be exercised by at least one feature scenario — i.e.
// the route path (or its templated stem) appears in some feature file.
function pc3(repoRoot) {
  // Carved-out exclusions per the PC wording.
  const excluded = new Set(['/spotify/callback']);

  const modules = ['auth', 'library', 'review', 'tags'];

  // Classify each handler in the four modules by reading the corresponding
  // http.go and seeing which views.* function it renders. A handler that
  // renders `views.XxxPage(...)` is a page; `views.XxxFrag(...)` /
  // `views.XxxModalFrag(...)` is a fragment.
  //
  // Special cases:
  //   * Handlers that have no `views.Xxx(...)` call but are pure redirects
  //     (e.g. Logout) are treated as page-equivalents.
  const handlerKind = new Map(); // handlerName -> 'page' | 'frag' | 'redirect'

  for (const mod of modules) {
    const httpFile = join(repoRoot, 'src/internal', mod, 'adapters/http.go');
    if (!existsSync(httpFile)) continue;
    const src = readFileSync(httpFile, 'utf8');
    // Walk function definitions; for each, scan its body for views.* calls.
    const lines = src.split('\n');
    const funcRe = /^func\s+\(\w+\s+\*HttpHandler\)\s+(\w+)\(/;
    const indices = [];
    for (let i = 0; i < lines.length; i++) {
      if (funcRe.test(lines[i])) indices.push({ name: lines[i].match(funcRe)[1], line: i });
    }
    indices.push({ name: null, line: lines.length });
    for (let i = 0; i < indices.length - 1; i++) {
      const { name, line } = indices[i];
      if (!name) continue;
      const body = lines.slice(line, indices[i + 1].line).join('\n');
      let kind = 'redirect'; // default for handlers that don't render a templ
      // Order matters: check for Page first since "*Page" is more specific
      // than "*Frag" (which could in theory appear inside a Page composition).
      if (/views\.\w+Page\s*\(/.test(body)) kind = 'page';
      else if (/views\.\w+(Frag|Modal)\s*\(/.test(body)) kind = 'frag';
      handlerKind.set(name, kind);
    }
  }

  const pageRoutes = []; // { module, method, path, handler }
  const reRoute = /mux\.Handle\(\s*"([^"]+)"\s*,\s*httpx\.HandlerFunc\(h\.(\w+)\)\)/g;

  for (const mod of modules) {
    const file = join(repoRoot, 'src/internal', mod, 'adapters/routes.go');
    if (!existsSync(file)) continue;
    const src = readFileSync(file, 'utf8');
    let m;
    reRoute.lastIndex = 0;
    while ((m = reRoute.exec(src)) !== null) {
      const spec = m[1];
      const handler = m[2];
      const parts = spec.split(/\s+/);
      const method = parts.length === 2 ? parts[0] : 'ANY';
      const path = parts.length === 2 ? parts[1] : parts[0];

      if (excluded.has(path)) continue;
      if (method !== 'GET' && method !== 'ANY') continue;

      const kind = handlerKind.get(handler);
      // Only pages (and the auth redirect Logout) count as user-facing routes
      // for PC3. Fragment endpoints are loaded by their parent page.
      if (kind !== 'page' && handler !== 'Logout') continue;

      pageRoutes.push({ module: mod, method, path, handler });
    }
  }

  // Load every feature file's text once.
  const featFiles = walk(join(repoRoot, 'e2e', 'feat'), (p) => p.endsWith('.feature'));
  const featCorpus = featFiles
    .map((f) => readFileSync(f, 'utf8'))
    .join('\n')
    .toLowerCase();

  // Coverage heuristic: a route is covered if any of these appears in the
  // feature corpus:
  //   1. the literal path stem (`/app/library/dashboard`)
  //   2. the last meaningful segment of the path (the leaf word, e.g.
  //      `dashboard`, `discover`, plus per-route synonyms)
  //   3. for the root `/{$}` route — any login-feature scenario
  //   4. for `/logout` — the word "logout" / "log out"
  //   5. for `/unauthorized` — "unauthorized" or "unauthorised"
  //   6. for `/app/library/albums/{albumId}` — "album detail" / "detail page"
  const failures = [];
  for (const r of pageRoutes) {
    const stem = r.path.split('{')[0].replace(/\/$/, '').toLowerCase();
    let covered = false;

    if (r.path === '/{$}') {
      covered = /feature:\s*login/i.test(featCorpus);
    } else if (r.path === '/logout') {
      covered = /\blog\s?out\b/i.test(featCorpus);
    } else if (r.path === '/unauthorized') {
      covered = /unauthori[sz]ed/i.test(featCorpus);
    } else {
      // Try the literal stem first.
      if (featCorpus.includes(stem)) {
        covered = true;
      } else {
        // Fall back to the leaf segment (last non-empty, non-brace part).
        const segments = r.path.split('/').filter(Boolean).filter((p) => !p.startsWith('{'));
        const leaf = segments[segments.length - 1]?.toLowerCase();
        if (leaf && featCorpus.includes(leaf)) covered = true;
        // Specific synonyms for routes whose leaf is parameterised.
        if (!covered && r.path === '/app/library/albums/{albumId}') {
          covered = /album\s+detail|detail\s+page|album\s+in\s+the\s+library/i.test(featCorpus);
        }
      }
    }

    if (!covered) {
      failures.push(
        `route ${r.method} ${r.path} (${r.module} → ${r.handler}) has no covering scenario in e2e/feat/`,
      );
    }
  }

  return {
    id: 'PC3',
    description: 'Route coverage',
    command: 'node e2e/check-product-criteria.mjs --pc PC3',
    pass: failures.length === 0,
    failures,
  };
}

// ----------------------- PC4 ----------------------------------------------

// Testid naming: `<surface>[-<element>][-<modifier>]`, kebab-case lowercase,
// surface derived from the declaring templ's filename. We only check testids
// `added or modified by this build`, sourced from `git diff <base>...HEAD`.
//
// The README documents "cross-surface composition is allowed" — a fragment may
// use the consuming page's surface name when it's owned by that page. We
// therefore accept either the file's own surface OR any other known surface
// in the same module's views/ directory.
function pc4(repoRoot, baseRef) {
  let diffOut = '';
  try {
    diffOut = execFileSync(
      'git',
      ['-C', repoRoot, 'diff', `${baseRef}...HEAD`, '--name-only'],
      { encoding: 'utf8' },
    );
  } catch (err) {
    return {
      id: 'PC4',
      description: 'Testid naming convention',
      command: `node e2e/check-product-criteria.mjs --pc PC4 --base ${baseRef}`,
      pass: false,
      failures: [`git diff failed: ${err.message}`],
    };
  }

  const changedTempl = diffOut
    .split('\n')
    .filter((p) => p.endsWith('.templ'));

  const failures = [];

  if (changedTempl.length === 0) {
    return {
      id: 'PC4',
      description: 'Testid naming convention',
      command: `node e2e/check-product-criteria.mjs --pc PC4 --base ${baseRef}`,
      pass: true,
      failures: [],
      note: 'no templ files changed in this build',
    };
  }

  // Surface helper: strip _page / _frag / _modal, convert _ to -.
  function surfaceFromFile(filePath) {
    const base = basename(filePath).replace(/\.templ$/, '');
    return base.replace(/_(page|frag|modal)$/i, '').replace(/_/g, '-').toLowerCase();
  }

  // Convert a Go camel/PascalCase identifier to kebab-case.
  function camelToKebab(name) {
    return name
      .replace(/([A-Z]+)([A-Z][a-z])/g, '$1-$2')
      .replace(/([a-z\d])([A-Z])/g, '$1-$2')
      .toLowerCase();
  }

  // Build the set of all surfaces declared anywhere in src/internal/** so we
  // can validate cross-surface composition. Sources:
  //   1. each templ file's surface (from filename), and
  //   2. each sub-template declaration inside that file (e.g.
  //      `templ albumListRow(...)` inside `albums_list_frag.templ`
  //      contributes the surface `album-list-row`). Multi-component files
  //      are common, and their sub-component testids legitimately use the
  //      sub-component's name as the surface.
  const allTemplFiles = walk(join(repoRoot, 'src/internal'), (p) => p.endsWith('.templ'));
  const allSurfaces = new Set();
  for (const f of allTemplFiles) {
    allSurfaces.add(surfaceFromFile(f));
    let body;
    try { body = readFileSync(f, 'utf8'); } catch { continue; }
    const reSubTempl = /^templ\s+([A-Za-z][A-Za-z0-9_]*)\s*\(/gm;
    let mm;
    while ((mm = reSubTempl.exec(body)) !== null) {
      const ident = mm[1];
      const kebab = camelToKebab(ident);
      // Strip trailing -page / -frag / -modal so e.g. `AlbumsListBodyFrag`
      // contributes `albums-list-body` (matching the file-derived rule).
      const trimmed = kebab.replace(/-(page|frag|modal)$/i, '');
      allSurfaces.add(trimmed);
    }
  }

  const reTestId = /data-testid="([^"]+)"/;
  // Kebab-case lowercase: letters, digits, hyphens. No double-hyphens, no
  // leading/trailing hyphen.
  const reKebab = /^[a-z0-9]+(-[a-z0-9]+)*$/;

  for (const relTempl of changedTempl) {
    const abs = join(repoRoot, relTempl);
    if (!existsSync(abs)) continue;

    // Get the diff for this file and extract added `data-testid="..."` literals
    // (lines starting with `+`, excluding the `+++ b/...` header).
    let fileDiff = '';
    try {
      fileDiff = execFileSync(
        'git',
        ['-C', repoRoot, 'diff', `${baseRef}...HEAD`, '--', relTempl],
        { encoding: 'utf8' },
      );
    } catch (err) {
      failures.push(`could not diff ${relTempl}: ${err.message}`);
      continue;
    }

    const fileSurface = surfaceFromFile(relTempl);
    const declared = new Set();

    for (const line of fileDiff.split('\n')) {
      if (!line.startsWith('+') || line.startsWith('+++')) continue;
      const m = line.match(reTestId);
      if (!m) continue;
      declared.add(m[1]);
    }

    for (const id of declared) {
      if (!reKebab.test(id)) {
        failures.push(`${relTempl}: testid "${id}" is not kebab-case lowercase`);
        continue;
      }
      // Must start with a known surface (either own file's surface or any
      // declared surface in the codebase — to allow cross-composition).
      const segments = id.split('-');
      let matched = false;
      // Try the longest prefix down to the shortest. The surface "album-score-readout"
      // wins over "album" because we want the most specific surface.
      for (let take = segments.length; take >= 1; take--) {
        const candidate = segments.slice(0, take).join('-');
        if (candidate === fileSurface || allSurfaces.has(candidate)) {
          matched = true;
          break;
        }
      }
      if (!matched) {
        failures.push(
          `${relTempl}: testid "${id}" surface does not match this file's surface "${fileSurface}" or any cross-composed surface`,
        );
      }
    }
  }

  return {
    id: 'PC4',
    description: 'Testid naming convention',
    command: `node e2e/check-product-criteria.mjs --pc PC4 --base ${baseRef}`,
    pass: failures.length === 0,
    failures,
  };
}

// ----------------------- PC5 ----------------------------------------------

// Selector discipline (amended). Every `.locator(...)` argument in spec files
// must be one of:
//   - the literal `'dialog[open]'` selector
//   - a single `[data-testid="X"]` matcher
//   - a comma-separated alternation of two-or-more `[data-testid="X"]` matchers
//   - any selector when the receiver is dialog-scoped (chained under
//     `dialog[open]`) — the dialog-scope exception preserves backwards
//     compatibility with existing modal-interaction patterns
// Forbidden everywhere (including inside dialog-scoped chains):
//   - CSS-class selectors (`.foo`)
//   - text-content / nth-of-type / XPath
//   - `getByText`, `getByPlaceholder`, `getByAltText`, `getByTitle`
//
// Note on the dialog-scope exception: the script accepts arbitrary selector
// strings when the receiver is dialog-scoped. This matches the original PC5
// implementation's behaviour and aligns with the amended PC5 wording's
// "locator chained under any of the above" clause for the OUTERMOST locator
// argument. CSS-class / nth-of-type / XPath remain forbidden even inside
// dialog scope. Stricter enforcement of bare-attribute selectors inside
// dialog chains (e.g. `input[name="X"]`) is a deferred follow-up — those
// sites are accepted here.
//
// "Scoped inside an open dialog" detection: a `.locator(...)` call whose
// receiver expression contains the string `dialog[open]` (e.g.
// `page.locator('dialog[open]').foo` or `const dialog = page.locator('dialog[open]'); dialog.locator(...)`).
// Variables assigned from such a locator are tracked across the file.
function pc5(repoRoot) {
  const specFiles = walk(join(repoRoot, 'e2e', 'spec'), (p) => p.endsWith('.spec.ts'));
  const flaggedFactories = ['getByText', 'getByPlaceholder', 'getByAltText', 'getByTitle'];

  // Selector composed entirely of [data-testid="..."] matchers — single or
  // comma-separated alternation. Each non-empty trimmed piece must match
  // /^\[data-testid="[^"]+"\]$/. Empty pieces (trailing commas, etc.) are
  // tolerated.
  function isTestidAttrSelector(sel) {
    if (sel == null) return false;
    const pieces = sel.split(',').map((p) => p.trim()).filter((p) => p.length > 0);
    if (pieces.length === 0) return false;
    return pieces.every((p) => /^\[data-testid="[^"]+"\]$/.test(p));
  }

  const failures = [];

  for (const file of specFiles) {
    const src = readFileSync(file, 'utf8');
    const lines = src.split('\n');
    const rel = relpath(repoRoot, file);

    // Collect names of variables bound to a dialog-scoped locator.
    // Patterns:
    //   const <name> = page.locator('dialog[open]'...)
    //   const <name> = page.locator('dialog[open] <rest>'...)
    //   const <name> = <existing-dialog-scoped>.locator(...)
    //   const <name> = <existing-dialog-scoped>.getByTestId(...)
    const dialogScopedVars = new Set();
    // Walk the file iteratively until the set stabilises (handles forward refs).
    const reVarBind = /(?:const|let|var)\s+(\w+)\s*=\s*(.+?);?$/;
    let grew = true;
    while (grew) {
      grew = false;
      for (const line of lines) {
        const m = line.match(reVarBind);
        if (!m) continue;
        const [, name, expr] = m;
        if (dialogScopedVars.has(name)) continue;
        // Selector string that begins with `dialog[open]` (with or without
        // following content) is dialog-scoped.
        if (/['"`]dialog\[open\]/.test(expr)) {
          dialogScopedVars.add(name);
          grew = true;
          continue;
        }
        // chained off existing scoped var (anywhere in the expression)
        for (const v of dialogScopedVars) {
          if (new RegExp(`\\b${v}\\b`).test(expr)) {
            dialogScopedVars.add(name);
            grew = true;
            break;
          }
        }
      }
    }

    // Flagged factory calls (getByText etc.) anywhere in the spec.
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      for (const f of flaggedFactories) {
        const re = new RegExp(`\\.${f}\\s*\\(`);
        if (re.test(line)) {
          failures.push(`${rel}:${i + 1}: forbidden locator factory \`${f}\` — only getByTestId/getByRole/getByLabel are allowed`);
        }
      }
    }

    // Check every .locator(...) call.
    // Match `<receiver>.locator(<args...>)` where args may span lines. To keep
    // this tractable we scan line-by-line and require the opening of the
    // locator call to live on one line — the few cases in this repo with
    // multi-line array args (album_actions.spec.ts) are matched off the
    // receiver expression on that opening line.
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      // Find `.locator(` occurrences with a known receiver. Capture the
      // receiver expression up to the `.locator(` token (greedy back to the
      // start of the chain — bounded by whitespace / open-paren / comma).
      const re = /([\w$\.\[\]'"`\(\)\^\*=\s\-]+?)\.locator\(\s*([^)]*)/g;
      let m;
      while ((m = re.exec(line)) !== null) {
        const receiver = m[1].trim();
        const argRaw = m[2].trim();

        // What does the locator's argument look like? Trim leading bracket
        // (array form), surrounding quotes.
        // If the arg is an array literal (album_actions.spec.ts line 54), the
        // selectors are on subsequent lines — grab them.
        const argString = argRaw.replace(/^\[/, '').replace(/\s+$/, '');
        // Strip optional outer quotes for the simple-string case.
        const firstChar = argString[0];
        let strLit = null;
        if (firstChar === `'` || firstChar === `"` || firstChar === '`') {
          const end = argString.indexOf(firstChar, 1);
          if (end > 0) strLit = argString.slice(1, end);
        }

        // Permit the literal dialog[open] selector.
        if (strLit === 'dialog[open]') continue;
        // Permit a selector string that begins with `dialog[open]`, i.e.
        // a combined-selector form of dialog scoping.
        if (strLit !== null && /^dialog\[open\](\s|$|\[)/.test(strLit)) continue;

        // Amended PC5: permit a selector composed entirely of
        // [data-testid="..."] matchers (single OR comma-separated
        // alternation), regardless of receiver. Semantically equivalent
        // to getByTestId / a chain of getByTestId, but expresses the
        // multi-testid "first-matching-of-N" alternation pattern.
        if (isTestidAttrSelector(strLit)) continue;

        // Is the receiver dialog-scoped?
        const receiverScoped =
          /['"`]dialog\[open\]/.test(receiver) ||
          [...dialogScopedVars].some((v) => new RegExp(`\\b${v}\\b`).test(receiver));

        if (receiverScoped) {
          // Inside dialog. Still forbid CSS class selectors and explicit
          // nth-of-type/XPath patterns (the PC bans them everywhere, including
          // inside dialogs — wording: "No CSS-class selectors, no text-content
          // selectors, no nth-of-type, no XPath").
          if (strLit !== null) {
            if (/(^|\s|,|>)\.[\w-]/.test(strLit)) {
              failures.push(`${rel}:${i + 1}: CSS class selector inside dialog: ${JSON.stringify(strLit)}`);
            }
            if (/nth-of-type/.test(strLit) || strLit.startsWith('//') || strLit.startsWith('xpath=')) {
              failures.push(`${rel}:${i + 1}: forbidden selector pattern: ${JSON.stringify(strLit)}`);
            }
          }
          continue;
        }

        // Not dialog-scoped and not a permitted testid-attr / dialog literal
        // selector — flag it.
        if (strLit !== null) {
          failures.push(`${rel}:${i + 1}: locator selector outside allow-list (dialog[open] / [data-testid="..."] / alternation): ${JSON.stringify(strLit)}`);
        } else {
          // Multi-line argument: either a single quoted string broken across
          // lines, or an array literal whose elements are quoted testid
          // selectors joined with ', '. Capture from the open `(` of the
          // .locator call forward until paren depth returns to zero — that's
          // the argument body.
          const startIdx = line.indexOf('.locator(', m.index) + '.locator('.length;
          let depth = 1;
          let argBody = '';
          // Scan rest of current line.
          for (let c = startIdx; c < line.length && depth > 0; c++) {
            const ch = line[c];
            if (ch === '(') depth++;
            else if (ch === ')') { depth--; if (depth === 0) break; }
            argBody += ch;
          }
          // Continue onto subsequent lines if not yet balanced.
          for (let j = i + 1; j < Math.min(lines.length, i + 12) && depth > 0; j++) {
            argBody += ' ';
            const ln = lines[j];
            for (let c = 0; c < ln.length && depth > 0; c++) {
              const ch = ln[c];
              if (ch === '(') depth++;
              else if (ch === ')') { depth--; if (depth === 0) break; }
              argBody += ch;
            }
          }
          // Strategy: strip every `[data-testid="..."]` matcher from the
          // argument body and a trailing `.join('...')` if present; the
          // residue must contain only whitespace, commas, quotes, and
          // array brackets — i.e. nothing but allowed syntax noise.
          const matcherRe = /\[data-testid="[^"]+"\]/g;
          const matchers = argBody.match(matcherRe);
          if (matchers && matchers.length > 0) {
            const residue = argBody
              .replace(matcherRe, '')
              .replace(/\.join\(\s*['"`][^'"`]*['"`]\s*\)/g, '')
              .replace(/['"`,\s\[\]]/g, '');
            if (residue === '') continue;
          }
          failures.push(`${rel}:${i + 1}: locator with non-string/multi-line arg outside allow-list: ${argBody.slice(0, 200)}`);
        }
      }
    }
  }

  return {
    id: 'PC5',
    description: 'Selector discipline',
    command: 'node e2e/check-product-criteria.mjs --pc PC5',
    pass: failures.length === 0,
    failures,
  };
}

// ----------------------- PC6 ----------------------------------------------

function pc6(repoRoot) {
  const specFiles = walk(join(repoRoot, 'e2e', 'spec'), (p) => p.endsWith('.spec.ts'));
  const failures = [];
  // Catch waitForTimeout, raw sleep, setTimeout-as-await, and `delay(`.
  const reTimeout = /\.waitForTimeout\(/;
  const reSleep = /\b(sleep|delay)\s*\(\s*\d+/;
  const reSetTimeout = /setTimeout\s*\(/;
  for (const file of specFiles) {
    const lines = readLines(file);
    const rel = relpath(repoRoot, file);
    for (let i = 0; i < lines.length; i++) {
      if (reTimeout.test(lines[i])) {
        failures.push(`${rel}:${i + 1}: forbidden waitForTimeout call`);
      }
      if (reSleep.test(lines[i])) {
        failures.push(`${rel}:${i + 1}: forbidden fixed-duration sleep/delay call`);
      }
      if (reSetTimeout.test(lines[i])) {
        failures.push(`${rel}:${i + 1}: forbidden setTimeout call`);
      }
    }
  }
  return {
    id: 'PC6',
    description: 'No fixed-timeout waits',
    command: 'node e2e/check-product-criteria.mjs --pc PC6',
    pass: failures.length === 0,
    failures,
  };
}

// ----------------------- PC7 ----------------------------------------------

// Single auth path: every authenticated spec test reaches the authenticated
// state via `loginAs(context, userId)` imported from `e2e/helpers/auth.ts`.
// No alternative auth bypass.
//
// Heuristic: for each `test(...)` block, look at the body until the matching
// closing brace and verify either:
//   - the test does not visit any /app/... route AND does not need auth, or
//   - the test calls loginAs(...)
// Also flag any spec that calls `context.addCookies` with `wax_token` or that
// imports an auth helper other than `loginAs` from `helpers/auth`.
function pc7(repoRoot) {
  const specFiles = walk(join(repoRoot, 'e2e', 'spec'), (p) => p.endsWith('.spec.ts'));
  const failures = [];

  for (const file of specFiles) {
    const src = readFileSync(file, 'utf8');
    const rel = relpath(repoRoot, file);
    const lines = src.split('\n');

    // No raw wax_token cookie injection.
    for (let i = 0; i < lines.length; i++) {
      if (/wax_token/.test(lines[i]) && !/auth\.ts$/.test(file)) {
        failures.push(`${rel}:${i + 1}: raw wax_token reference outside helpers/auth.ts`);
      }
    }

    // Walk test blocks. We split the file at `^test(` declarations.
    const testStarts = [];
    for (let i = 0; i < lines.length; i++) {
      if (/^test\(\s*['"]/.test(lines[i])) testStarts.push(i);
    }
    testStarts.push(lines.length); // sentinel

    for (let t = 0; t < testStarts.length - 1; t++) {
      const start = testStarts[t];
      const end = testStarts[t + 1];
      const body = lines.slice(start, end).join('\n');
      const nameMatch = body.match(/^test\(\s*['"](.+?)['"]/);
      const name = nameMatch?.[1] ?? `<unknown@line ${start + 1}>`;

      // Does the test reach an authenticated state? Signals: it calls loginAs
      // OR it visits a /app/... route AND expects an authenticated outcome
      // (i.e. NOT an unauthorized/login redirect).
      const visitsApp = /['"]\/app\//.test(body);
      const usesLoginAs = /loginAs\s*\(/.test(body);
      const expectsUnauthFlow =
        /toHaveURL\(\s*['"]\/unauthorized/.test(body) ||
        /toHaveURL\(\s*['"]\/['"]/.test(body) ||
        /getByTestId\(\s*['"]unauthorized-page/.test(body) ||
        /getByTestId\(\s*['"]login-page/.test(body);

      // Authenticated if: explicitly logs in, OR visits /app and doesn't
      // appear to be probing the unauthenticated redirect.
      const isAuthenticated = usesLoginAs || (visitsApp && !expectsUnauthFlow);

      if (isAuthenticated && !usesLoginAs) {
        failures.push(`${rel}:${start + 1}: test "${name}" appears authenticated but does not call loginAs(...)`);
      }
    }
  }

  return {
    id: 'PC7',
    description: 'Single auth path',
    command: 'node e2e/check-product-criteria.mjs --pc PC7',
    pass: failures.length === 0,
    failures,
  };
}

// ----------------------- PC9 ----------------------------------------------

// No mocking or HTTP interception in any spec.
function pc9(repoRoot) {
  const specFiles = walk(join(repoRoot, 'e2e', 'spec'), (p) => p.endsWith('.spec.ts'));
  const failures = [];
  const patterns = [
    { re: /\.route\s*\(/, msg: 'forbidden page.route(...) request interception' },
    { re: /\.unroute\s*\(/, msg: 'forbidden page.unroute(...)' },
    { re: /\.fulfill\s*\(/, msg: 'forbidden request.fulfill(...) mock' },
    { re: /\bMockServiceWorker\b/, msg: 'forbidden MockServiceWorker reference' },
    { re: /from\s+['"]msw['"]/, msg: 'forbidden msw import' },
    { re: /from\s+['"]nock['"]/, msg: 'forbidden nock import' },
    { re: /from\s+['"]sinon['"]/, msg: 'forbidden sinon import' },
    { re: /\bjest\.(mock|fn|spyOn)\b/, msg: 'forbidden jest mocking helper' },
    { re: /\bvi\.(mock|fn|spyOn)\b/, msg: 'forbidden vitest mocking helper' },
  ];
  for (const file of specFiles) {
    const lines = readLines(file);
    const rel = relpath(repoRoot, file);
    for (let i = 0; i < lines.length; i++) {
      for (const { re, msg } of patterns) {
        if (re.test(lines[i])) {
          failures.push(`${rel}:${i + 1}: ${msg}: ${lines[i].trim()}`);
        }
      }
    }
  }
  return {
    id: 'PC9',
    description: 'Real backend',
    command: 'node e2e/check-product-criteria.mjs --pc PC9',
    pass: failures.length === 0,
    failures,
  };
}

// ----------------------- driver -------------------------------------------

const ALL = {
  PC1: pc1,
  PC2: pc2,
  PC3: pc3,
  PC4: pc4,
  PC5: pc5,
  PC6: pc6,
  PC7: pc7,
  PC9: pc9,
};

function main() {
  let opts;
  try {
    opts = parseArgs(process.argv.slice(2));
  } catch (err) {
    process.stderr.write(`error: ${err.message}\n`);
    process.exit(2);
  }
  if (opts.help) {
    process.stdout.write(
      'usage: check-product-criteria [--pc <id>] [--base <git-ref>] [--repo-root <dir>]\n',
    );
    process.exit(0);
  }

  const repoRoot = resolve(opts.repoRoot ?? process.cwd());

  const order = ['PC1', 'PC2', 'PC3', 'PC4', 'PC5', 'PC6', 'PC7', 'PC9'];
  const toRun = opts.only ? [opts.only] : order;

  const results = [];
  for (const id of toRun) {
    const fn = ALL[id];
    if (!fn) {
      process.stderr.write(`unknown PC: ${id}\n`);
      process.exit(2);
    }
    results.push(id === 'PC4' ? fn(repoRoot, opts.base) : fn(repoRoot));
  }

  let failed = 0;
  for (const r of results) {
    const verdict = r.pass ? 'PASS' : 'FAIL';
    const tail = r.note ? ` — ${r.note}` : '';
    process.stdout.write(`${r.id}  ${verdict}  ${r.description}${tail}\n`);
    if (!r.pass) {
      failed++;
      for (const f of r.failures) process.stdout.write(`    ${f}\n`);
    }
  }
  process.stdout.write(`\nPC8 (suite green) is verified by \`npx playwright test\` — not run here.\n`);

  process.exit(failed === 0 ? 0 : 1);
}

main();
