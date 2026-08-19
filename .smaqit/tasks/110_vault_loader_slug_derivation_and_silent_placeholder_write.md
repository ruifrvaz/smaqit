# Vault Loader: Wrong Project-Slug Derivation, and Non-Interactive Runs Silently Write Placeholder Secrets

**Status:** Not Started
**Created:** 2026-08-15
**Mode:** Assisted

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
- **Bug 2 — broad fix, all confirmed sites.** Fix all 5 confirmed ad hoc-read sites, not just the 1 originally flagged: `load-credentials.sh:167` (new-scheme github token), `load-credentials.sh:334` (legacy-scheme github token, not in the original report), `rotate-credential.sh:139` (github token), `rotate-credential.sh:169` (Cyso secret), `rotate-credential.sh:176` (Terraform secret key) — all confirmed live by direct inspection. `bootstrap-app-to-machine.sh` needs no change; confirmed it already uses the safe `read_secret` helper throughout with no ad hoc secret-feeding reads.
- **Empty-value guard: yes**, added before every affected `vault kv put` call, independent of the `read_secret` fix, as defense-in-depth (e.g. a user hitting Enter on an empty prompt with a real tty attached).
- **Regression coverage — two layers.** (1) A fast, deterministic static-analysis check (`check-no-ad-hoc-secret-reads.sh`) grepping this skill's `scripts/*.sh` for any bare `read -s`/`-p` feeding a `vault kv put` without going through `read_secret` — guards against this exact regression class recurring via a future ad hoc read. (2) A new agent-driven `.smaqit/bench/skills/smaqit.infrastructure-vault-loader/bench.yaml` Bench case (first Bench manifest in this repo — existing manifests are all in the sibling `smaqit-adk` repo) proving both fixes end-to-end through a real Codex agent following the skill's documented flow, against a real ephemeral `vault server -dev` (not a mock) — a single with-artifact-only variant, since there's no meaningful without-artifact baseline for a pure bug fix (an agent without the skill staged wouldn't be running these scripts at all). This deliberately deviates from Bench's headline with/without-artifact comparison pattern.
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
6. Replace all 5 confirmed ad hoc reads with `read_secret`: `load-credentials.sh:167`, `load-credentials.sh:334`, `rotate-credential.sh:139`, `rotate-credential.sh:169`, `rotate-credential.sh:176`.
7. Add an empty-value guard immediately before each of those 5 `vault kv put` calls — exit non-zero with a clear error rather than writing.

**Phase C — Regression coverage**
8. New static-analysis check `check-no-ad-hoc-secret-reads.sh`: greps this skill's `scripts/*.sh` for a bare `read -s`/`-p` feeding a `vault kv put` without going through `read_secret`.
9. New `.smaqit/bench/skills/smaqit.infrastructure-vault-loader/bench.yaml` (and `.smaqit/bench/README.md` if this repo doesn't already have one): single with-artifact-only case; `prepare` launches an ephemeral `vault server -dev`; fixture is a project with a multi-word, heading-format `AGENTS.md` Project Name; prompt directs Codex to run `load-credentials.sh --print-slug` then attempt the github-token load non-interactively; expectations assert the printed slug matches the expected git-remote/dirname value and that nothing was written to Vault.
10. `smaqit-adk bench validate` the new manifest, then a live trial via `smaqit-adk bench suite run .smaqit/bench`; report the result.

## Known Issues Triage

**Triaged:** 2026-08-15
**Tools searched:** HashiCorp Vault
**Result:** Clear — this is an internal bug in this framework's own script, not a Vault defect.

## Acceptance Criteria

- [ ] Project slug derives from `git remote get-url origin` (basename, `.git` stripped) or the current directory name — no longer parsed from `AGENTS.md` — via one shared function used by both `load-credentials.sh` and `rotate-credential.sh`
- [ ] Running any secret-writing prompt in this script family with no `/dev/tty` available fails loudly (non-zero exit, clear error) instead of silently writing an empty or placeholder secret, confirmed across all 5 identified sites
- [ ] Every affected `vault kv put` refuses to write when its secret value is empty, independent of how it was read
- [ ] A static-analysis safety check and a new Bench case both pass, covering both fixes together against a real ephemeral Vault

## Findings

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` | Modify — slug derivation, both GH token sites (:167, :334), new `--print-slug` flag |
| `skills/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh` | Modify — slug derivation, all 3 secret-read sites (:139, :169, :176), add `read_secret` |
| `skills/smaqit.infrastructure-vault-loader/scripts/bootstrap-app-to-machine.sh` | No change — confirmed already uses `read_secret` throughout, no ad hoc secret-feeding reads found |
| `skills/smaqit.infrastructure-vault-loader/scripts/lib-project-slug.sh` | Create — shared `derive_project_slug()` |
| `skills/smaqit.infrastructure-vault-loader/scripts/check-no-ad-hoc-secret-reads.sh` | Create — static regression check |
| `.smaqit/bench/skills/smaqit.infrastructure-vault-loader/bench.yaml` | Create — first Bench manifest in this repo |
| `.smaqit/bench/README.md` | Create if absent — adapted from `smaqit-adk`'s conventions |

## Notes

Found live in a downstream project during task 058 (2026-08-13) — that project's established slug (`acme-case-manager-poc`) is confirmed correct and already populated at `secret/apps/acme-case-manager-poc/*`/`secret/machines/acme-test/*`; the bogus `secret/apps/acme/github` (`token: n/a`) placeholder written by this bug was found and deleted the same session.

**Planning follow-up (2026-08-19):** Re-verified both bugs are still live by direct code inspection (not assumed) before planning: `load-credentials.sh:131-132`'s fallback path still truncates via `sed 's/ .*$//'`; the ad hoc-read pattern is confirmed present at 5 sites, not just the 1 originally flagged (`load-credentials.sh:167,334`; `rotate-credential.sh:139,169,176`) — `bootstrap-app-to-machine.sh` is confirmed clean (already uses `read_secret` throughout). Also found `rotate-credential.sh` has its own duplicated, fallback-less copy of the slug-derivation logic — it never actually reproduced Bug 1's truncation (no heading-format fallback path at all; on inline-grep failure it goes straight to manual entry), which is why only `load-credentials.sh` hit the bug live.

Before approving the git-remote/dirname derivation approach, cross-referenced this machine's real, populated local Vault (9+ app slugs) against every checkable real project's actual `AGENTS.md`/git-remote/dirname — every one already matches its real Vault path today. `areaoffice-poc` and `iodis-crm-poc` are notable: their `AGENTS.md` titles are multi-word display text (`Commerce CRM — Proof of Concept`, `IODIS CRM (PoC)`) that the *current* buggy fallback would already mis-derive (to `commerce`/`iodis`), yet their real Vault paths use the correct repo-identity slug — meaning those two projects are likely already silently exposed to Bug 1 today, with the correct slug having reached Vault some other way (manual entry). The planned fix closes that gap rather than opening a new one, on this machine at least; the general risk (a project whose current slug genuinely differs from its repo identity) remains an accepted, undefended trade-off for machines/projects not checked here.
