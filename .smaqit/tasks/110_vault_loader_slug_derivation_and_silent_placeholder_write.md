---
status: PR Open
created: "2026-08-15"
mode: Assisted
started: "2026-08-19"
pr: 85
---

# Vault Loader: Wrong Project-Slug Derivation, and Non-Interactive Runs Silently Write Placeholder Secrets

## Description

`smaqit.infrastructure-vault-loader`'s `load-credentials.sh` has two distinct, real bugs found live in a downstream project (2026-08-13), both in the "new scheme" (`apps/`+`machines/`) code path.

### Bug 1 — project-slug derivation truncates multi-word `AGENTS.md` project names

`load-credentials.sh:112-139` derives `PROJECT_SLUG` from `AGENTS.md`'s "Project Name" field, with two extraction strategies depending on format:

- **Inline format** (`Project Name: value`, line 129): `grep ... | tr ' ' '-'` — replaces every space with a hyphen, so a multi-word name survives intact.
- **Heading + next-line format** (`## Project Name\n\nvalue`, lines 130-133, the fallback when the inline grep finds nothing): `awk '...' | sed 's/ .*$//'` — this `sed` keeps only the text **before the first space**, silently dropping everything after the first word.

The downstream project's `AGENTS.md` uses the heading+next-line format, with a multi-word value like `Acme Case Manager (PoC)`. The inline grep (needs `": "` on the same line as "project name") doesn't match a bare `## Project Name` heading line, so it falls through to the awk+sed fallback — which truncates `Acme Case Manager (PoC)` to `acme`, not the project's actual established slug `acme-case-manager-poc` (which matches the repository/directory name, used everywhere else in that project's Vault layout: `secret/apps/acme-case-manager-poc/*`, `secret/machines/acme-test/*`).

Two separate problems compound here:
1. The two extraction strategies are inconsistent with each other (one preserves multi-word names, the other truncates to the first word) — likely unintentional, since there's no reason the heading+next-line format should behave differently.
2. Even a fixed hyphenation (`acme-case-manager-(poc)` → sanitized → `acme-case-manager-poc`, coincidentally) is fragile in general: it derives a *technical* slug from a *human-readable display title* with no guaranteed relationship to the real slug. The actual source of truth for a project's identity throughout the rest of this framework's Vault conventions is the repository/directory name (`agents.md`'s heading is prose for humans, not a machine identifier). Deriving from `git remote get-url origin` (basename, `.git` stripped) or the current working directory's basename would be a more reliable source than parsing a title field — worth considering as the actual fix rather than just hardening the string manipulation.

### Bug 2 — a non-interactive run silently writes a placeholder secret instead of failing

`load-credentials.sh:39-49` defines a `read_secret` helper that correctly reads from `/dev/tty` explicitly:

```sh
read_secret() {
  local _var="$1" _prompt="$2" _value
  IFS= read -rs -p "  ${_prompt}: " _value </dev/tty && echo >&2
  printf -v "$_var" '%s' "$_value"
  unset _value
}
```

Reading from `/dev/tty` explicitly is the right call — in a context with no real terminal attached (e.g., an agent's shell tool), this **fails loudly** (`read: /dev/tty: No such device or address` or similar) rather than silently succeeding with garbage.

But the new-scheme GitHub-token prompt at `load-credentials.sh:167` does **not** use this helper — it's a bare, ad hoc read with no `/dev/tty` redirect and no post-read validation:

```sh
read -s -p "  github_token: " GH_TOKEN && echo
vault kv put "secret/apps/${APP_SLUG}/github" token="$GH_TOKEN" > /dev/null
```

In a non-interactive shell, this `read` doesn't fail — it returns immediately (empty value, or whatever happens to be on stdin), and the script proceeds unconditionally to `vault kv put` with that value. Observed live: an agent running this script non-interactively got `GH_TOKEN="n/a"` written as a real Vault secret value at `secret/apps/acme/github` (compounding with Bug 1's wrong slug), reported as `[1/1] GitHub fine-grained PAT ... DONE` — no error, no warning, a secret path that now looks populated with a real credential but isn't. Found and manually deleted before it could mislead a later session into trusting that path.

## Design Decisions

- **Bug 1 — full replacement, not a minimal patch.** Replace `AGENTS.md` Project Name parsing entirely with `git remote get-url origin` (basename, `.git` stripped) → current-directory-basename derivation, extracted into a new shared `lib-project-slug.sh` sourced by both `load-credentials.sh` and `rotate-credential.sh`. Accepted risk, deliberately not mitigated: a project whose current slug differs from its repo identity would compute a different slug post-fix, with no mismatch-detection safety net added. Verified live against this machine's real Vault (9+ populated app slugs, cross-referenced against each checked-out project's actual `AGENTS.md`/git remote/dirname) — every project's derived slug already matches its real Vault path; `areaoffice-poc` and `iodis-crm-poc` in particular have multi-word `AGENTS.md` titles that the *current* buggy fallback would already mis-derive, yet their real Vault paths use the correct repo-identity slug, meaning the fix closes a gap those two are already silently sitting in rather than opening a new one.
- **Bug 2 — broad fix, all confirmed sites.** Fix all 7 confirmed ad hoc-read sites, not just the 1 originally flagged: `load-credentials.sh:167` (new-scheme github token), `load-credentials.sh:216` (Cyso `app_credential_secret`), `load-credentials.sh:318` (Terraform `s3_secret_key`), `load-credentials.sh:334` (legacy-scheme github token), `rotate-credential.sh:139` (github token), `rotate-credential.sh:169` (Cyso secret), `rotate-credential.sh:176` (Terraform secret key) — all confirmed live by direct inspection. (Corrected 2026-08-19 during implementation: the planning-phase grep matched only `github_token`/`GH_TOKEN`-named patterns and missed `load-credentials.sh:216,318`, which use the same unguarded `read -s` shape under different variable names — found on a full-file read before editing.) `bootstrap-app-to-machine.sh` needs no change; confirmed it already uses the safe `read_secret` helper throughout with no ad hoc secret-feeding reads.
- **Empty-value guard scope**: checks every field about to be written in a given `vault kv put` call, not only its secret field — e.g. `load-credentials.sh:216`'s guard covers both `CYSO_ID` (plain read, not itself a secret) and `CYSO_SECRET`, since an incomplete credential (real secret, empty ID) is equally bad and both are available at the same guard point. Plain, non-secret companion reads (`CYSO_ID`, `S3_KEY`/access key, menu choices, confirmations) are not otherwise touched — no `/dev/tty` redirect added to those, since the empty-guard already catches the actual harm (a placeholder silently persisted) without redesigning this script family's non-secret prompts.
- **Empty-value guard: yes**, added before every affected `vault kv put` call, independent of the `read_secret` fix, as defense-in-depth (e.g. a user hitting Enter on an empty prompt with a real tty attached).
- **Regression coverage — two layers.** (1) A fast, deterministic static-analysis check (`check-no-ad-hoc-secret-reads.sh`) grepping this skill's `scripts/*.sh` for any bare `read -s`/`-p` feeding a `vault kv put` without going through `read_secret` — guards against this exact regression class recurring via a future ad hoc read; this is the primary, actually-executable proof for Bug 2 (also manually verified live against this machine's real dev Vault: a non-interactive run failed loudly with nothing written). (2) A new agent-driven `.smaqit/bench/skills/smaqit.infrastructure-vault-loader/bench.yaml` Bench case (first Bench manifest in this repo) — **scoped to Bug 1 (slug derivation) only**, not both fixes as originally planned. Reason for the narrowed scope, decided mid-implementation: reading `smaqit-adk`'s actual Bench engine source (`src/bench/run.go`/`process.go`) confirmed a real, unmanaged process-leak gap — a backgrounded process launched in `Case.prepare` (which a live `vault server -dev` would require, to stay alive through the variant's execution and the `expect` checks) survives past a successful run with no engine-provided teardown; `terminateProcessTree`'s process-group kill only fires on timeout/cancellation. Filed as `smaqit-adk` task 033 rather than worked around here. Additionally, Codex credits are exhausted on this machine, so no live Bench trial (`bench suite run`) is possible this session regardless — the manifest is authored and structurally validated (`bench validate`) only; a live trial is deferred.
- **Slug derivation is extracted to one shared library**, not duplicated per-script — this also fixes a pre-existing inconsistency where `rotate-credential.sh`'s own copy of the derivation logic has no fallback path at all (falls straight to manual entry). `read_secret` itself stays duplicated per-script, matching its existing convention (already duplicated between `load-credentials.sh` and `bootstrap-app-to-machine.sh`) — not forcing an unrelated refactor of an already-working pattern.
- **New `--print-slug` dry-run flag** added to `load-credentials.sh` — independently useful for manual debugging, and needed by the Bench case to check derived slug without a full interactive/Vault-writing run.

## Implementation Steps

**Phase A — Slug derivation (Bug 1)**
1. New shared lib `skills/smaqit.infrastructure-vault-loader/scripts/lib-project-slug.sh`: `derive_project_slug()` tries `git remote get-url origin` (basename, `.git` stripped) → current-directory basename → empty (caller prompts manually), sanitized to the existing lowercase-hyphenated form.
2. `load-credentials.sh`: replace the AGENTS.md-parsing block (current lines ~128-133) with a sourced call to `derive_project_slug`; keep the existing manual-entry fallback (~135-137) unchanged.
3. `rotate-credential.sh`: replace its duplicated, fallback-less inline derivation (current lines ~80-82) with the same sourced call — also fixes the pre-existing inconsistency between the two scripts' derivation logic.
4. Add a `--print-slug` flag to `load-credentials.sh`: prints the derived slug and exits 0 before any Vault interaction.

**Phase B — Non-interactive secret-write safety (Bug 2)**
5. Add `read_secret` (verbatim, matching its existing definition) to `rotate-credential.sh`, which currently lacks it.
6. Replace all 7 confirmed ad hoc reads with `read_secret`: `load-credentials.sh:167,216,318,334`, `rotate-credential.sh:139,169,176`.
7. Add an empty-value guard immediately before each of those 6 `vault kv put` calls (2 sites at :216/:318 share one `vault kv put` call each with a companion non-secret field) — check every field about to be written, exit non-zero with a clear error rather than writing.

**Phase C — Regression coverage**
8. New static-analysis check `check-no-ad-hoc-secret-reads.sh`: greps this skill's `scripts/*.sh` for a bare `read -s`/`-p` feeding a `vault kv put` without going through `read_secret`.
9. New `.smaqit/bench/skills/smaqit.infrastructure-vault-loader/bench.yaml` (and `.smaqit/bench/README.md`, since this repo has no `.smaqit/bench/` yet): single with-artifact-only case, **scoped to Bug 1 only** (no live Vault — see Design Decisions for why); fixture is a project with a multi-word, heading-format `AGENTS.md` Project Name and a known git remote; prompt directs Codex to run `load-credentials.sh --print-slug`; expectation asserts the printed output matches the expected git-remote-derived slug (not the old buggy AGENTS.md-truncated one).
10. `smaqit-adk bench validate` the new manifest (structural only). No live trial this session — Codex credits are exhausted; defer `smaqit-adk bench suite run` to whenever credits are available.

## Known Issues Triage

**Triaged:** 2026-08-19
**Tools searched:** HashiCorp Vault (resolved to `hashicorp/vault-guides` — a fuzzy-search mismatch against the tutorial/examples repo, not the main `hashicorp/vault` product repo; recorded as a categorization limitation below, not treated as evidence of Vault-side clearance), OpenAI Codex CLI (`openai/codex`, newly relevant since this plan adds a Codex-driven Bench case)
**Result:** Advisory — one known, already-mitigated Codex issue; no blockers.

### Advisory Issues
- [#36570 exec: approvals_reviewer = "auto_review" silently defeats an explicit --sandbox level](https://github.com/openai/codex/issues/36570) — `openai/codex` — opened 2026-08-02 — bug, sandbox, exec, CLI, config. Directly relevant to this task's new Bench case (`codex exec` driven non-interactively). Downgraded from Blocking: already a documented, mitigated known limitation in this framework's own Bench conventions (`smaqit-adk/.smaqit/bench/README.md`'s reusable Codex process-variant block already pins `--sandbox danger-full-access` explicitly rather than relying on defaults) — the planned manifest copies that same block verbatim per Implementation Step 9.

### Historical (Closed)
- None directly relevant — closed-search results for `openai/codex` were mostly unrelated feature requests/enhancements (headless fork, replay mode, structured metrics) or fixed bugs in unrelated areas (hooks, apply_patch rollout); none concerning `--sandbox`/non-interactive secret-write behavior specifically.

### Unresolvable Tools
- None

### Omitted Tools
- None (2 of 5 max repositories used)

### Search Warnings
- None

## Acceptance Criteria

- [x] Project slug derives from `git remote get-url origin` (basename, `.git` stripped) or the current directory name — no longer parsed from `AGENTS.md` — via one shared function used by both `load-credentials.sh` and `rotate-credential.sh`
- [x] Running any secret-writing prompt in this script family with no `/dev/tty` available fails loudly (non-zero exit, clear error) instead of silently writing an empty or placeholder secret, confirmed across all 7 identified sites
- [x] Every affected `vault kv put` refuses to write when its secret value is empty, independent of how it was read
- [x] A static-analysis safety check passes for Bug 2 (covering all 7 sites, verified live against a real dev Vault); a new Bench case structurally validates for Bug 1 (live trial deferred — Codex credits exhausted; see `smaqit-adk` task 033 for the process-leak gap that scoped Bug 2 out of the Bench case)

## Findings

**Implementation approach:**
- New `skills/smaqit.infrastructure-vault-loader/scripts/lib-project-slug.sh` defines `derive_project_slug()` (git-remote basename with `.git` stripped → dirname fallback → empty), sourced by both `load-credentials.sh` and `rotate-credential.sh`, replacing each script's own AGENTS.md-parsing logic entirely.
- `load-credentials.sh` gained a `--print-slug` flag checked immediately after sourcing the lib, before any Vault status/unseal/auth interaction — exits 0 with just the derived slug.
- All 7 confirmed ad hoc-read sites (`load-credentials.sh:158,207,309,325`; `rotate-credential.sh:140,170,177` — line numbers shifted slightly from the original grep-time estimates after the derivation-block edits) now use `read_secret` (added to `rotate-credential.sh`, which previously lacked it) plus an empty-value guard immediately before their `vault kv put`, checking every field about to be written (not just the secret one — e.g. `CYSO_ID`/`CYSO_SECRET` together).
- New `scripts/check-no-ad-hoc-secret-reads.sh`: greps for any `read ... -s`/`-rs` line lacking `/dev/tty`, i.e. anything outside a `read_secret`-shaped definition. Verified both directions live: passes clean against the fixed scripts, and correctly fails (exit 1, correct file/line) against a deliberately reintroduced ad hoc `read -s` in a throwaway copy.
- New `.smaqit/bench/skills/smaqit.infrastructure-vault-loader/bench.yaml` (first Bench manifest in this repo, `.smaqit/bench/README.md` also authored) — single with-artifact-only Case, scoped to Bug 1 only (see Decisions). Structurally validated via `smaqit-adk bench validate`/`bench suite validate`; no live trial (Codex credits exhausted this session).

**Decisions made:**
- Both Bugs verified still live by direct code inspection before implementing, not assumed from the task file (see Notes' 2026-08-19 planning entries).
- Bug 2's scope corrected mid-implementation from 5 to 7 sites: a full-file read of `load-credentials.sh` (not just the narrower `github_token`-targeted grep used during planning) surfaced `CYSO_SECRET` (was line 216, now 207) and `S3_SECRET` (was line 318, now 309) using the identical unguarded `read -s` shape under different variable names.
- Manually verified Bug 2's fix live against this machine's real local dev Vault: a non-interactive `MACHINE_SLUG=... load-credentials.sh </dev/null` on a throwaway scratch app slug failed loudly (`/dev/tty: No such device or address`, exit 1) and confirmed via `vault kv get` that nothing was written — then cleaned up the scratch test directory (no real Vault path was ever touched; the scratch slug was never real).
- Manually verified Bug 1's fix live: `derive_project_slug` tested against this worktree (git remote → `smaqit`), a non-repo directory (dirname fallback), an SSH host-alias remote matching the real `areaoffice-poc` shape from planning, and an https remote with mixed-case/punctuation — all correct. Also reproduced the original bug's exact multi-word heading-format `AGENTS.md` scenario end-to-end via `--print-slug` and confirmed no truncation.
- **Bench case scope reduced from both bugs to Bug 1 only, mid-implementation.** Reading `smaqit-adk`'s actual Bench engine source (`src/bench/run.go`, `process.go`, `process_unix.go`) before authoring the manifest — since no existing manifest anywhere backgrounds a long-lived process from `Case.prepare` — confirmed a real gap: `terminateProcessTree`'s process-group kill only fires on timeout/cancellation, never on a successful run, so a backgrounded `vault server -dev` (needed to stay alive through the Case's variant execution and `expect` checks) would leak as an orphaned host process with no engine-provided cleanup. Filed as `smaqit-adk` task 033 rather than worked around here (per `smaqit.bench-scaffold`'s own guidance: report a genuine engine limitation as a follow-up task, don't patch the engine inline from a downstream task). Bug 2's regression coverage instead rests on the static check plus the manual live-Vault verification above — both already real and already executed, not deferred.
- Also confirmed live (from the user, after I'd only done source-level analysis, not a live run): Codex credits are exhausted on this machine, so no live Bench trial (`bench suite run`) is possible this session for either bug's coverage regardless. The manifest is authored and structurally validated only; a live trial is deferred to whenever credits are available.
- Kept `read_secret` duplicated per-script (matching its existing convention, already duplicated between `load-credentials.sh` and `bootstrap-app-to-machine.sh`) rather than also extracting it to the new shared lib — only slug derivation, which had a real behavioral inconsistency between the two scripts to fix, was worth centralizing.

**Blockers encountered:**
- Task 110's file used the pre-frontmatter `**Status:**` header format (same as every task file in this repo predating the recent lifecycle-resolver change) — converted just this task's header to unblock `task-start`, matching the fix already applied to task 111 earlier this session.
- Bench manifest authoring was blocked on a genuine, previously-undocumented engine limitation (see Decisions above) — resolved by scoping the Case down rather than waiting on an upstream fix, with the gap filed as its own tracked task in the owning repo.

**Follow-up identified:**
- `smaqit-adk` task 033 (Bench `Case.prepare` background-process cleanup gap) — filed and pushed to the sibling repo, not part of this task's own completion criteria.
- Once `smaqit-adk` task 033 ships and Codex credits are available, `.smaqit/bench/skills/smaqit.infrastructure-vault-loader/bench.yaml` could be extended with a second Case covering Bug 2 against a real ephemeral Vault — not filed as a separate smaqit task since it's a natural, low-effort extension of an existing manifest rather than new tracked work; noted here for whoever next touches this manifest.

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` | Modify — slug derivation, all 4 secret-read sites (:167, :216, :318, :334), new `--print-slug` flag |
| `skills/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh` | Modify — slug derivation, all 3 secret-read sites (:139, :169, :176), add `read_secret` |
| `skills/smaqit.infrastructure-vault-loader/scripts/bootstrap-app-to-machine.sh` | No change — confirmed already uses `read_secret` throughout, no ad hoc secret-feeding reads found |
| `skills/smaqit.infrastructure-vault-loader/scripts/lib-project-slug.sh` | Create — shared `derive_project_slug()` |
| `skills/smaqit.infrastructure-vault-loader/scripts/check-no-ad-hoc-secret-reads.sh` | Create — static regression check |
| `.smaqit/bench/skills/smaqit.infrastructure-vault-loader/bench.yaml` | Create — first Bench manifest in this repo; scoped to Bug 1 only (see Decisions) |
| `.smaqit/bench/skills/smaqit.infrastructure-vault-loader/fixtures/multi-word-project/AGENTS.md` | Create — Case fixture |
| `.smaqit/bench/README.md` | Create — adapted from `smaqit-adk`'s conventions (this repo had no `.smaqit/bench/` at all) |
| `.smaqit/bench/runs/.gitignore` | Create — matches `smaqit-adk`'s own convention |

## Notes

Found live in a downstream project during task 058 (2026-08-13) — that project's established slug (`acme-case-manager-poc`) is confirmed correct and already populated at `secret/apps/acme-case-manager-poc/*`/`secret/machines/acme-test/*`; the bogus `secret/apps/acme/github` (`token: n/a`) placeholder written by this bug was found and deleted the same session.

**Planning follow-up (2026-08-19):** Re-verified both bugs are still live by direct code inspection (not assumed) before planning: `load-credentials.sh:131-132`'s fallback path still truncates via `sed 's/ .*$//'`; the ad hoc-read pattern was initially confirmed at 5 sites during planning (`load-credentials.sh:167,334`; `rotate-credential.sh:139,169,176`), then corrected to **7** during implementation once a full-file read caught 2 more (`load-credentials.sh:216,318` — the planning-phase grep matched only `github_token`/`GH_TOKEN`-named patterns and missed the Cyso/Terraform secret reads under different variable names) — `bootstrap-app-to-machine.sh` is confirmed clean (already uses `read_secret` throughout). Also found `rotate-credential.sh` has its own duplicated, fallback-less copy of the slug-derivation logic — it never actually reproduced Bug 1's truncation (no heading-format fallback path at all; on inline-grep failure it goes straight to manual entry), which is why only `load-credentials.sh` hit the bug live.

Before approving the git-remote/dirname derivation approach, cross-referenced this machine's real, populated local Vault (9+ app slugs) against every checkable real project's actual `AGENTS.md`/git-remote/dirname — every one already matches its real Vault path today. `areaoffice-poc` and `iodis-crm-poc` are notable: their `AGENTS.md` titles are multi-word display text (`Commerce CRM — Proof of Concept`, `IODIS CRM (PoC)`) that the *current* buggy fallback would already mis-derive (to `commerce`/`iodis`), yet their real Vault paths use the correct repo-identity slug — meaning those two projects are likely already silently exposed to Bug 1 today, with the correct slug having reached Vault some other way (manual entry). The planned fix closes that gap rather than opening a new one, on this machine at least; the general risk (a project whose current slug genuinely differs from its repo identity) remains an accepted, undefended trade-off for machines/projects not checked here.
