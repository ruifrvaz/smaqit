---
name: smaqit.test-e2e-playwright
description: Use when running Playwright E2E smoke tests against a live application instance to validate primary user flows before or after deployment. Accepts a BASE_URL, executes the project's `e2e/` smoke suite, and produces a PASS/FAIL report per scenario. Also use when generating the initial `e2e/` test suite: reads machine-readable scenario declarations from the coverage spec, instantiates generic flow templates, and produces project-specific test files via the `generate-e2e.js` script. The skill has no hardcoded project knowledge — all scenario parameters come from the project's own specs.
metadata:
  version: "2.0.0"
compatibility: "Node.js 18+, npm, Chromium (auto-installed via --with-deps). Runs on Linux, macOS, WSL2."
allowed-tools: Bash(node:*), Bash(npm:*), Bash(npx:*), Bash(curl:*), Read, Write, Grep
---

# Playwright E2E Smoke Tests

## Template System

The skill ships three reusable flow templates under `templates/flows/`:

| Template file | Type token | Covers |
|---|---|---|
| `login-flow.template.js` | `login-flow` | Multi-user login; asserts login form disappears |
| `navigation-flow.template.js` | `navigation-flow` | Login → navigate → optional pre-click → assert element visible |
| `api-trigger-flow.template.js` | `api-trigger-flow` | Login → navigate → optional pre-click → click trigger → wait for loading state to resolve |

Templates use `{{token_name}}` placeholders (lowercase with underscores) that map directly to fields in the coverage spec declaration YAML.

**Token catalog:**

| Token | Used in | Description |
|---|---|---|
| `{{id}}` | all | Scenario ID (e.g. `e2e-001`) |
| `{{title}}` | all | Human-readable scenario title |
| `{{timeout_ms}}` | all | Test timeout in milliseconds |
| `{{username_input_selector}}` | all | CSS selector for username input |
| `{{password_input_selector}}` | all | CSS selector for password input |
| `{{submit_selector}}` | all | CSS/text selector for login submit |
| `{{usernames}}` | login-flow | Comma-separated list of usernames |
| `{{password}}` | login-flow | Password for all users |
| `{{username}}` | navigation, api-trigger | Single username |
| `{{nav_text}}` | navigation, api-trigger | Sidebar nav item text to click |
| `{{pre_click_selector}}` | navigation, api-trigger | Optional extra click before assertion; use `__none__` to skip — the line is removed from output |
| `{{target_selector}}` | navigation-flow | Playwright selector for expected visible element |
| `{{trigger_text}}` | api-trigger-flow | Text of button/link that starts the async call |
| `{{loading_text}}` | api-trigger-flow | Text that appears while call is in progress |

**Coverage spec declaration format:** Declare each E2E scenario as a fenced YAML block tagged `e2e-declaration` in `specs/coverage/*.md`:

~~~
```yaml e2e-declaration
id: e2e-001
title: All seed users can log in
type: login-flow
usernames: user1,user2,user3
password: secret
username_input_selector: input[placeholder="username"]
password_input_selector: input[type="password"]
submit_selector: button:has-text("Sign in")
timeout_ms: 5000
output_file: 01-login.spec.js
```
~~~

## Steps

### Mode A — Generate (when `e2e/` does not exist)

1. Check for `e2e/` in the project root. If it exists, skip to Mode B.
2. Confirm `specs/coverage/*.md` contains at least one `e2e-declaration` block. If none are found: read the functional spec and stack spec, then add the declarations to the coverage spec before proceeding.
3. Run the generation script from the project root:
   ```bash
   node [SMAQIT_SKILLS_DIR]/smaqit.test-e2e-playwright/scripts/generate-e2e.js
   ```
   The script reads all `e2e-declaration` blocks from `specs/coverage/`, extracts the local dev port from `specs/stack/STK-001.md`, generates `e2e/playwright.config.js` and all `e2e/smoke/<output_file>` test files, and reports each generated path.
4. Install Playwright and the Chromium browser:
   ```bash
   npm install --save-dev @playwright/test   # in project root, or backend/ if no root package.json
   npx playwright install chromium --with-deps
   ```
5. Commit: `git add e2e/ && git commit -m "test: add Playwright E2E smoke suite"`

### Mode B — Execute (against a running app)

1. Confirm `e2e/` exists. If not, run Mode A first.
2. Resolve `BASE_URL` in priority order:
   - `BASE_URL` env var provided by the caller
   - Fixed IP or domain from `specs/infrastructure/INF-001.md`
   - Ask the user
3. Confirm target is reachable: `curl -sf <BASE_URL>/` — must return HTTP 200. Stop and report if it fails.
4. Run: `BASE_URL=<url> npx playwright test e2e/smoke/ --reporter=list`
5. Capture exit code. Exit 0 = all pass; non-zero = failures present.
6. Parse output and produce a PASS/FAIL table, one row per scenario.
7. Print the summary line:
   - All pass: `E2E smoke PASS — <N> scenarios, 0 failures.`
   - Any fail: `E2E smoke FAIL — <F> of <N> scenarios failed. See details below.`
8. On failure: print Playwright's error output for each failing scenario. Do not redeploy or restart the app — report only.

## Output

- **Mode A:** `e2e/` directory with `playwright.config.js` and one spec file per declared scenario under `e2e/smoke/`.
- **Mode B:** Console report — PASS/FAIL table per scenario and one-line summary.

## Scope

- Covers smoke-level validation: login, navigation, primary user flows, AI endpoint reachability.
- Does NOT replace unit or integration tests (Vitest/Supertest suite in `backend/src/`).
- Does NOT test visual regression or accessibility.
- Does NOT manage app startup or teardown — caller must ensure the target is running before invoking this skill.
- Does NOT handle authentication token refresh or long-session scenarios.
- AI endpoint tests use extended timeout from the declaration (e.g. `timeout_ms: 90000`); if the AI service is unavailable, the loading indicator never disappears — mark as skipped with a note.

## Examples

**Mode A — Generate:**
Coverage spec has 7 `e2e-declaration` blocks. No `e2e/` exists.
→ `generate-e2e.js` runs, produces `e2e/playwright.config.js` + 7 spec files. `@playwright/test` installed. Committed.

**Mode B — Local (Phase 3):**
Backend on `:3000`, Vite frontend on `:5173`. Invoke with `BASE_URL=http://localhost:5173`.
→ `E2E smoke PASS — 7 scenarios, 0 failures.`

**Mode B — Deployed (Phase 4, step 7.5):**
Dev VM at `http://81.24.10.203:8082`.
→ Same report format; any 500s or login failures surface before Phase 4 closes.

## Gotchas

- **Templates use lowercase token names** — all `{{token}}` placeholders use `lowercase_with_underscores` matching YAML field names exactly. Avoid camelCase or UPPER_CASE.
- **`__none__` sentinel** — set `pre_click_selector: __none__` to skip the pre-click step. The generate script removes any output line containing `__none__`.
- **App must be running for Mode B** — for local mode: run `npm run dev` in `backend/` and `npm run dev` in `frontend/` before invoking.
- **BASE_URL is always the frontend** — even in local mode, `BASE_URL` is the Vite/frontend server. The frontend proxy forwards `/api/` to the backend.
- **Playwright install is per-machine** — run `npx playwright install chromium --with-deps` on fresh machines and CI runners.
- **Re-generating overwrites** — Mode A's generate script overwrites existing test files without warning. Track `e2e/` in git so diffs are visible before committing.

## Completion

- [ ] `e2e/` directory exists with `playwright.config.js` and spec files matching the declarations (Mode A)
- [ ] `@playwright/test` installed and Chromium browser available (`npx playwright install chromium`)
- [ ] `BASE_URL` resolved and target confirmed reachable (HTTP 200)
- [ ] All scenarios pass (or skipped for AI scenarios when the AI service is unavailable)
- [ ] Summary line printed: `E2E smoke PASS — N scenarios, 0 failures.`

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | `e2e/` exists — skip Mode A and switch to Mode B |
| No `e2e-declaration` blocks found | Add declarations to the coverage spec before running Mode A; the generate script will report the exact error |
| Unknown `type` in declaration | No template file matches — check `templates/flows/` for available types; add a new template or change the declaration type |
| `{{token}}` not substituted | Token name in template doesn't match a field in the declaration — check spelling; the generate script warns per unresolved token |
| `e2e/` missing in Mode B | Run Mode A first |
| Target not reachable | Stop; app must be running before invoking this skill |
| Login scenario fails | Seed data likely missing — check `seedDb()` in `server.js` or run `docker exec <backend> node src/db/seed.js` |
| AI scenario times out | Mark as skipped; note that AI service credentials may be invalid or the endpoint unreachable |
| `playwright install` fails | Try `--with-deps`; may require `sudo` on Ubuntu |
| Exit code 1, no visible errors | Run with `--reporter=verbose` |
