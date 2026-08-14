# Vault Loader: Wrong Project-Slug Derivation, and Non-Interactive Runs Silently Write Placeholder Secrets

**Status:** Not Started
**Created:** 2026-08-15
**Mode:** Assisted

## Description

`smaqit.infrastructure-vault-loader`'s `load-credentials.sh` has two distinct, real bugs found live in `iodis-crm-poc` (2026-08-13), both in the "new scheme" (`apps/`+`machines/`) code path.

### Bug 1 — project-slug derivation truncates multi-word `AGENTS.md` project names

`load-credentials.sh:112-139` derives `PROJECT_SLUG` from `AGENTS.md`'s "Project Name" field, with two extraction strategies depending on format:

- **Inline format** (`Project Name: value`, line 129): `grep ... | tr ' ' '-'` — replaces every space with a hyphen, so a multi-word name survives intact.
- **Heading + next-line format** (`## Project Name\n\nvalue`, lines 130-133, the fallback when the inline grep finds nothing): `awk '...' | sed 's/ .*$//'` — this `sed` keeps only the text **before the first space**, silently dropping everything after the first word.

`iodis-crm-poc`'s `AGENTS.md` uses the heading+next-line format, with the value `IODIS CRM (PoC)`. The inline grep (needs `": "` on the same line as "project name") doesn't match a bare `## Project Name` heading line, so it falls through to the awk+sed fallback — which truncates `IODIS CRM (PoC)` to `iodis`, not the project's actual established slug `iodis-crm-poc` (which matches the repository/directory name, used everywhere else in that project's Vault layout: `secret/apps/iodis-crm-poc/*`, `secret/machines/iodis-test/*`).

Two separate problems compound here:
1. The two extraction strategies are inconsistent with each other (one preserves multi-word names, the other truncates to the first word) — likely unintentional, since there's no reason the heading+next-line format should behave differently.
2. Even a fixed hyphenation (`iodis-crm-(poc)` → sanitized → `iodis-crm-poc`, coincidentally) is fragile in general: it derives a *technical* slug from a *human-readable display title* with no guaranteed relationship to the real slug. The actual source of truth for a project's identity throughout the rest of this framework's Vault conventions is the repository/directory name (`agents.md`'s heading is prose for humans, not a machine identifier). Deriving from `git remote get-url origin` (basename, `.git` stripped) or the current working directory's basename would be a more reliable source than parsing a title field — worth considering as the actual fix rather than just hardening the string manipulation.

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

In a non-interactive shell, this `read` doesn't fail — it returns immediately (empty value, or whatever happens to be on stdin), and the script proceeds unconditionally to `vault kv put` with that value. Observed live: an agent running this script non-interactively got `GH_TOKEN="n/a"` written as a real Vault secret value at `secret/apps/iodis/github` (compounding with Bug 1's wrong slug), reported as `[1/1] GitHub fine-grained PAT ... DONE` — no error, no warning, a secret path that now looks populated with a real credential but isn't. Found and manually deleted before it could mislead a later session into trusting that path.

## Design Decisions

TBD — open questions:

- For Bug 1: fix the `sed` truncation to match the inline path's hyphenation, or replace the whole prose-parsing approach with `basename "$(git remote get-url origin 2>/dev/null || pwd)" .git`-style derivation from the repo identity instead? The latter is more robust but is a bigger behavioral change — worth deciding whether `AGENTS.md`'s Project Name field is meant to be authoritative for the slug at all, or was only ever intended as a convenient first guess.
- For Bug 2: should the fix be narrowly "use the existing `read_secret` helper here too" (minimal, consistent with the rest of the script), or should there be a broader non-interactive-safety pass across every raw `read` call in this script family (`bootstrap-app-to-machine.sh`, `rotate-credential.sh` likely have the same ad hoc pattern)?
- Should `vault kv put` calls in this script family generally refuse to write an empty-string secret value, as a defense-in-depth backstop independent of how the value was read?

## Implementation Steps

TBD — sketch, not committed:

1. Audit every raw `read`/`read -s`/`read -p` call across `load-credentials.sh`, `bootstrap-app-to-machine.sh`, and `rotate-credential.sh` for the same ad hoc (non-`/dev/tty`, non-`read_secret`) pattern; replace with `read_secret` (or a non-secret equivalent with the same `/dev/tty` safety) everywhere.
2. Add an empty-value guard immediately before every `vault kv put` in this script family — refuse to write and exit non-zero rather than silently persisting an empty/placeholder value.
3. Fix the heading+next-line slug-extraction `sed` to hyphenate multi-word values consistently with the inline path (minimum fix), or replace both paths with repo-identity-based derivation (preferred fix, pending the Design Decision above).
4. Add regression coverage: a multi-word `## Project Name` heading value produces the expected hyphenated slug; a non-interactive invocation (no `/dev/tty` available) fails loudly instead of writing an empty/placeholder secret.

## Known Issues Triage

**Triaged:** 2026-08-15
**Tools searched:** HashiCorp Vault
**Result:** Clear — this is an internal bug in this framework's own script, not a Vault defect.

## Acceptance Criteria

- [ ] A multi-word `AGENTS.md` "Project Name" value (either format) produces a correctly hyphenated slug, with both extraction paths behaving consistently
- [ ] Running any secret-writing prompt in this script family with no `/dev/tty` available fails loudly (non-zero exit, clear error) instead of silently writing an empty or placeholder secret
- [ ] Regression tests cover both fixes

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
| `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` | Modify — slug derivation, GH token prompt |
| `skills/smaqit.infrastructure-vault-loader/scripts/bootstrap-app-to-machine.sh` | Audit/modify — same ad hoc `read` pattern likely present |
| `skills/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh` | Audit/modify — same ad hoc `read` pattern likely present |

## Notes

Found live in `iodis-crm-poc` during task 058 (2026-08-13) — that project's established slug (`iodis-crm-poc`) is confirmed correct and already populated at `secret/apps/iodis-crm-poc/*`/`secret/machines/iodis-test/*`; the bogus `secret/apps/iodis/github` (`token: n/a`) placeholder written by this bug was found and deleted the same session. Not yet independently verified whether `bootstrap-app-to-machine.sh`/`rotate-credential.sh` share the same ad hoc `read` pattern — flagged as an audit step above rather than assumed.
