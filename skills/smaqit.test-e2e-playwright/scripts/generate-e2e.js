#!/usr/bin/env node
// generate-e2e.js
// Reads `yaml e2e-declaration` blocks from coverage spec files and generates Playwright test files.
//
// Usage (from project root):
//   node [SMAQIT_SKILLS_DIR]/smaqit.test-e2e-playwright/scripts/generate-e2e.js \
//     [--coverage-spec <dir>]   default: specs/coverage
//     [--stack-spec <file>]     default: specs/stack/STK-001.md
//     [--templates-dir <dir>]   default: <script-dir>/../templates
//     [--output-dir <dir>]      default: e2e/smoke

'use strict';
const fs   = require('fs');
const path = require('path');

// ─── Args ────────────────────────────────────────────────────────────────────

const args = process.argv.slice(2);
function getArg(flag, def) {
  const i = args.indexOf(`--${flag}`);
  return i !== -1 && args[i + 1] ? args[i + 1] : def;
}

const COVERAGE_DIR  = getArg('coverage-spec', 'specs/coverage');
const STACK_SPEC    = getArg('stack-spec',     'specs/stack/STK-001.md');
const TEMPLATES_DIR = getArg('templates-dir',  path.join(__dirname, '..', 'templates'));
const OUTPUT_DIR    = getArg('output-dir',      'e2e/smoke');
const CONFIG_OUT    = 'e2e/playwright.config.js';

// ─── Helpers ─────────────────────────────────────────────────────────────────

/** Parse a flat `key: value` YAML block into an object. */
function parseDeclaration(block) {
  const result = {};
  for (const line of block.split('\n')) {
    const m = line.match(/^([\w_-]+)\s*:\s*(.+)$/);
    if (m) result[m[1].trim()] = m[2].trim();
  }
  return result;
}

/** Extract all ```yaml e2e-declaration blocks from a markdown string. */
function extractDeclarations(md) {
  const re = /```yaml\s+e2e-declaration\r?\n([\s\S]*?)```/g;
  const out = [];
  let m;
  while ((m = re.exec(md)) !== null) {
    const decl = parseDeclaration(m[1]);
    if (decl.id) out.push(decl);
  }
  return out;
}

/**
 * Substitute {{token}} placeholders in a template string.
 * After substitution, removes any line containing __none__ (optional-step sentinel).
 */
function substitute(template, tokens) {
  let result = template.replace(/\{\{(\w+)\}\}/g, (_, key) => {
    if (Object.prototype.hasOwnProperty.call(tokens, key)) return tokens[key];
    console.warn(`  [warn] token {{${key}}} not found in declaration — left as-is`);
    return `{{${key}}}`;
  });
  // Remove lines containing __none__ (pre_click_selector: __none__ sentinel)
  result = result.split('\n').filter(line => !line.includes('__none__')).join('\n');
  // Collapse 3+ consecutive blank lines to a single blank line (cosmetic cleanup after __none__ removal)
  result = result.replace(/\n{3,}/g, '\n\n');
  return result;
}

/** Heuristically extract the local dev port from the stack spec. */
function extractLocalPort(stackContent) {
  // Look for explicit "port: XXXX" or ":XXXX" near dev/vite/frontend keywords
  const matchers = [
    /(?:dev|vite|frontend)[^\n]*?:(\d{4,5})/i,
    /port[:\s]+(\d{4,5})/i,
    /:(\d{4,5})/,
  ];
  for (const re of matchers) {
    const m = stackContent.match(re);
    if (m) return m[1];
  }
  return '5173'; // Vite default
}

// ─── Main ─────────────────────────────────────────────────────────────────────

console.log('smaqit.test-e2e-playwright — generate-e2e.js\n');
console.log(`Coverage dir : ${COVERAGE_DIR}`);
console.log(`Stack spec   : ${STACK_SPEC}`);
console.log(`Templates dir: ${TEMPLATES_DIR}`);
console.log(`Output dir   : ${OUTPUT_DIR}`);
console.log(`Config out   : ${CONFIG_OUT}\n`);

// Collect declarations from all .md files in coverage dir
const declarations = [];
if (!fs.existsSync(COVERAGE_DIR)) {
  console.error(`[error] Coverage spec directory not found: ${COVERAGE_DIR}`);
  process.exit(1);
}
for (const file of fs.readdirSync(COVERAGE_DIR).sort()) {
  if (!file.endsWith('.md')) continue;
  const content = fs.readFileSync(path.join(COVERAGE_DIR, file), 'utf8');
  const found   = extractDeclarations(content);
  if (found.length > 0) {
    console.log(`Found ${found.length} declaration(s) in ${file}`);
    declarations.push(...found);
  }
}

if (declarations.length === 0) {
  console.error(`\n[error] No e2e-declaration blocks found in ${COVERAGE_DIR}/`);
  console.error('  Add ```yaml e2e-declaration blocks to your coverage spec file.\n');
  process.exit(1);
}

// Sort by id for stable output ordering
declarations.sort((a, b) => (a.id || '').localeCompare(b.id || ''));
console.log(`\nTotal: ${declarations.length} scenario(s) to generate.\n`);

// Extract local dev port from stack spec
const stackContent = fs.existsSync(STACK_SPEC) ? fs.readFileSync(STACK_SPEC, 'utf8') : '';
if (!stackContent) {
  console.warn(`[warn] Stack spec not found at ${STACK_SPEC} — defaulting port to 5173`);
}
const localPort = extractLocalPort(stackContent);
console.log(`Local dev port: ${localPort}\n`);

// Ensure output directories exist
fs.mkdirSync(OUTPUT_DIR, { recursive: true });
fs.mkdirSync(path.dirname(CONFIG_OUT), { recursive: true });

// Generate playwright.config.js
const configTplPath = path.join(TEMPLATES_DIR, 'playwright.config.template.js');
if (!fs.existsSync(configTplPath)) {
  console.error(`[error] Config template not found: ${configTplPath}`);
  process.exit(1);
}
const configContent = substitute(fs.readFileSync(configTplPath, 'utf8'), { DEFAULT_PORT: localPort });
fs.writeFileSync(CONFIG_OUT, configContent);
console.log(`Generated: ${CONFIG_OUT}`);

// Generate each scenario spec file
let generated = 0;
let skipped   = 0;

for (const decl of declarations) {
  const { id, type, output_file } = decl;

  if (!type || !output_file) {
    console.warn(`  [skip] ${id || '(no id)'}: missing required field(s): ${!type ? 'type' : ''}${!type && !output_file ? ', ' : ''}${!output_file ? 'output_file' : ''}`);
    skipped++;
    continue;
  }

  const tplFile = path.join(TEMPLATES_DIR, 'flows', `${type}.template.js`);
  if (!fs.existsSync(tplFile)) {
    console.warn(`  [skip] ${id}: no template for type '${type}' (expected: ${tplFile})`);
    skipped++;
    continue;
  }

  const template = fs.readFileSync(tplFile, 'utf8');
  const content  = substitute(template, decl);
  const outPath  = path.join(OUTPUT_DIR, output_file);
  fs.writeFileSync(outPath, content);
  console.log(`Generated: ${outPath}  [${type}]`);
  generated++;
}

console.log(`\nDone: ${generated} file(s) generated, ${skipped} skipped.`);
if (skipped > 0) process.exit(1);
