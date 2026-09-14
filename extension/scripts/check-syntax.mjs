// check-syntax.mjs — syntax-check every extension source file.
//
// Two files are ES modules and two are classic scripts, so they need different
// checks:
//   - service-worker.js is declared "type": "module" in the manifest, so it must
//     parse as an ES module.
//   - popup.js uses no imports and is loaded via a plain <script> tag.
//   - content scripts are classic scripts.
//
// Running `node --check` on a module without module context reports a spurious
// failure at the first `export`, which is why this distinguishes them rather
// than treating every file the same.

import { execFileSync } from 'node:child_process';
import { copyFileSync, rmSync, existsSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const root = join(here, '..');

const MODULE_FILES = ['service-worker.js'];
const SCRIPT_FILES = ['content-shopee.js', 'popup.js', 'main-network-observer.js'];

let failed = 0;

function checkScript(file) {
  const path = join(root, file);
  try {
    execFileSync(process.execPath, ['--check', path], { stdio: 'pipe' });
    console.log(`  ok    ${file} (classic script)`);
    return true;
  } catch (err) {
    console.error(`  FAIL  ${file}: ${err.stderr?.toString().trim() || err.message}`);
    return false;
  }
}

function checkModule(file) {
  const path = join(root, file);
  // node --check needs a module extension or type:module context; a temporary
  // .mjs copy is the least invasive way to ask the question.
  const tmp = join(root, `_syntax_check_${Date.now()}.mjs`);
  try {
    copyFileSync(path, tmp);
    execFileSync(process.execPath, ['--check', tmp], { stdio: 'pipe' });
    console.log(`  ok    ${file} (ES module)`);
    return true;
  } catch (err) {
    console.error(`  FAIL  ${file}: ${err.stderr?.toString().trim() || err.message}`);
    return false;
  } finally {
    if (existsSync(tmp)) rmSync(tmp, { force: true });
  }
}

console.log('Syntax check:');
for (const f of MODULE_FILES) if (!checkModule(f)) failed++;
for (const f of SCRIPT_FILES) if (!checkScript(f)) failed++;

// The manifest is hand-edited, so a malformed trailing comma is worth catching
// here rather than at chrome://extensions load time.
try {
  const manifest = JSON.parse(
    execFileSync(process.execPath, ['-e', 'process.stdout.write(JSON.stringify(require("./manifest.json")))'], {
      cwd: root,
      stdio: 'pipe',
    }).toString(),
  );
  if (manifest.manifest_version !== 3) {
    console.error(`  FAIL  manifest.json: expected manifest_version 3, got ${manifest.manifest_version}`);
    failed++;
  } else {
    console.log(`  ok    manifest.json (mv${manifest.manifest_version}, v${manifest.version})`);
  }
} catch (err) {
  console.error(`  FAIL  manifest.json: ${err.message}`);
  failed++;
}

if (failed > 0) {
  console.error(`\n${failed} file(s) failed the syntax check.`);
  process.exit(1);
}
console.log('\nAll extension files parse cleanly.');
