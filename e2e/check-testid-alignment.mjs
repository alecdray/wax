#!/usr/bin/env node
// Spec/templ alignment check.
//
// Detects `data-testid` values referenced by Playwright specs (via
// `page.getByTestId('...')` or equivalent) that are not declared by any templ
// file under `src/internal/**/*.templ`.
//
// Exits 0 with a terse "OK" if every spec reference is declared somewhere.
// Exits 1 and prints `<spec-path>:<line>: <testid>` for every orphan otherwise.
//
// Pure static analysis — reads files only. No dev server, DB, or browser.
//
// Usage:
//   node e2e/check-testid-alignment.mjs                # scan the repo
//   node e2e/check-testid-alignment.mjs --spec-dir D --templ-dir D ...
//
// Flags (optional, primarily for self-tests):
//   --spec-dir <dir>     directory to scan for *.spec.ts   (default: e2e/spec)
//   --templ-dir <dir>    directory to scan for *.templ     (default: src/internal)
//   --repo-root <dir>    base for default dirs             (default: cwd)

import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative, resolve } from 'node:path';

function parseArgs(argv) {
  const opts = {};
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === '--spec-dir') opts.specDir = argv[++i];
    else if (a === '--templ-dir') opts.templDir = argv[++i];
    else if (a === '--repo-root') opts.repoRoot = argv[++i];
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

// Collects every `data-testid="X"` literal under the given templ dir.
function collectDeclaredTestIds(templDir) {
  const declared = new Set();
  const files = walk(templDir, (p) => p.endsWith('.templ'));
  const re = /data-testid="([^"]+)"/g;
  for (const file of files) {
    const src = readFileSync(file, 'utf8');
    let m;
    while ((m = re.exec(src)) !== null) declared.add(m[1]);
  }
  return { declared, fileCount: files.length };
}

// Collects every spec reference to a testid string.
// Matches `getByTestId('X')` and `getByTestId("X")`.
function collectSpecReferences(specDir) {
  const refs = []; // { file, line, testid }
  const files = walk(specDir, (p) => p.endsWith('.spec.ts'));
  const re = /getByTestId\(\s*['"]([^'"]+)['"]\s*\)/g;
  for (const file of files) {
    const src = readFileSync(file, 'utf8');
    const lines = src.split('\n');
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      let m;
      re.lastIndex = 0;
      while ((m = re.exec(line)) !== null) {
        refs.push({ file, line: i + 1, testid: m[1] });
      }
    }
  }
  return { refs, fileCount: files.length };
}

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
      'usage: check-testid-alignment [--spec-dir DIR] [--templ-dir DIR] [--repo-root DIR]\n',
    );
    process.exit(0);
  }

  const repoRoot = resolve(opts.repoRoot ?? process.cwd());
  const specDir = resolve(opts.specDir ?? join(repoRoot, 'e2e', 'spec'));
  const templDir = resolve(opts.templDir ?? join(repoRoot, 'src', 'internal'));

  const { declared, fileCount: templCount } = collectDeclaredTestIds(templDir);
  const { refs, fileCount: specCount } = collectSpecReferences(specDir);

  const orphans = refs.filter((r) => !declared.has(r.testid));

  if (orphans.length === 0) {
    process.stdout.write(
      `OK — ${refs.length} spec testid reference(s) across ${specCount} spec file(s) all declared (scanned ${templCount} templ file(s), ${declared.size} declared testid(s)).\n`,
    );
    process.exit(0);
  }

  // Stable order: by spec path, then line, then testid.
  orphans.sort((a, b) => {
    if (a.file !== b.file) return a.file < b.file ? -1 : 1;
    if (a.line !== b.line) return a.line - b.line;
    return a.testid < b.testid ? -1 : 1;
  });

  for (const o of orphans) {
    const rel = relative(repoRoot, o.file) || o.file;
    process.stdout.write(`${rel}:${o.line}: ${o.testid}\n`);
  }

  const distinct = new Set(orphans.map((o) => o.testid)).size;
  process.stderr.write(
    `\n${orphans.length} orphan reference(s) across ${distinct} distinct missing testid(s); spec testid is not declared in any templ under ${relative(repoRoot, templDir) || templDir}.\n`,
  );
  process.exit(1);
}

main();
