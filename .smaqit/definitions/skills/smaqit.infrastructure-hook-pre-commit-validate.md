# Skill Definition: smaqit.infrastructure-hook-pre-commit-validate

## Identity
- **Name:** smaqit.infrastructure-hook-pre-commit-validate
- **Version:** 1.0.0
- **Description:** Validate staged files before a commit is accepted. Checks for: `.env` files, plaintext API keys/secrets, spec files stuck in `draft` status that should be `implemented`, and large binary files. Ships a `scripts/install.sh` to symlink itself as a git pre-commit hook. Can also be invoked manually as a smaqit skill before committing.

## Steps

### When invoked as a smaqit skill (manual pre-commit check)
1. Run `git diff --cached --name-only` to get staged file list.
2. Run validation checks against staged files (see checks below).
3. Report: PASS or FAIL per check. If any fail, list the failing files and the issue.

### When running as a git hook (automatic on `git commit`)
The `scripts/pre-commit.sh` script (installed via `scripts/install.sh`) runs the same checks using only shell tools (no agent invocation).

### Validation checks
1. **No `.env` files staged:** `git diff --cached --name-only | grep -E '\.env$|\.env\.'` — fail if any match.
2. **No plaintext secret patterns:** Scan staged file contents for patterns:
   - `ANTHROPIC_API_KEY=sk-` (literal key value)
   - `sk-ant-` (Anthropic key prefix)
   - `-----BEGIN (RSA|EC|OPENSSH) PRIVATE KEY-----`
   - Any line matching `(password|secret|token)\s*=\s*[^\$\{]{8,}` (non-templated value with ≥8 chars)
   Run: `git diff --cached | grep -E '(sk-ant-|ANTHROPIC_API_KEY=sk-|BEGIN.*PRIVATE KEY|password\s*=\s*[^\$\{]{8,})'`
3. **No draft specs staged for merge to main:** If the current branch is `main` or the commit is a merge, check staged `.md` files in `specs/` for `status: draft`. Warn (not fail) if found.
4. **No large files (>1 MB) staged:** `git diff --cached --name-only | xargs -I{} sh -c 'test -f "{}" && du -k "{}" | awk "{if (\$1 > 1024) print \"{}\"}"'`

### scripts/install.sh
The skill ships an install script at `.github/skills/smaqit.infrastructure-hook-pre-commit-validate/scripts/install.sh` that:
1. Copies `scripts/pre-commit.sh` to `.git/hooks/pre-commit`
2. Makes it executable
3. Reports: "pre-commit hook installed. Run `git commit` to test."

### scripts/pre-commit.sh
Minimal bash script that runs checks 1–3 using git and grep. Runs without agent context — shell only.

## Output
- Validation report (PASS/FAIL per check)
- `.git/hooks/pre-commit` installed (when `install.sh` is run)
- `.github/skills/smaqit.infrastructure-hook-pre-commit-validate/scripts/install.sh` — install script
- `.github/skills/smaqit.infrastructure-hook-pre-commit-validate/scripts/pre-commit.sh` — hook script

## Scope
- Does NOT prevent all possible security mistakes — it catches common patterns only.
- Does NOT scan un-staged files. Only staged changes are inspected.
- Does NOT block commits to non-main branches for draft spec status (warning only, not failure).

## Gotchas
- Git hooks live in `.git/hooks/` which is NOT tracked by version control. The `install.sh` script must be run once per clone. Mention this in project README under "Development Setup".
- `git diff --cached` shows only staged changes. Developers who use `git add -p` may have secrets in their working tree but not staged — this hook does not catch those.
- Anthropic key detection pattern — `sk-ant-` prefix. If this ever changes, update the pattern in `pre-commit.sh`.
- The draft spec check is a WARNING, not a BLOCK. It is acceptable to commit draft specs during active development.

## Completion
- [ ] Staged file list obtained
- [ ] All 4 checks run
- [ ] Report delivered (PASS/FAIL per check)
- [ ] Install script and hook script created in `scripts/`

## Failure Handling
| Situation | Action |
|-----------|--------|
| `.env` file staged | FAIL — block commit; report the filename |
| Plaintext secret pattern found | FAIL — block commit; report the file and matching line |
| Draft spec on main/merge | WARN — do not block; surface as advisory note |
| Large file staged | WARN — do not block; advise using Git LFS |

## Examples
**Input:** Developer runs `git add backend/.env && git commit -m "fix"`.
**Output:** FAIL: `.env` file staged. Commit blocked. "Remove `backend/.env` from staged files before committing."
