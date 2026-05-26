---
name: smaqit.infrastructure-hook-pre-commit-validate
description: Use when validating staged files before committing to catch .env files, plaintext API keys/secrets, draft spec files on main-branch commits, and large files over 1 MB. Also use when setting up the automated git pre-commit hook via the bundled install script so checks run automatically on every commit. Produces a PASS/FAIL report per check, with filenames and matching lines listed for failures.
metadata:
  version: "1.0.0"
---

# Pre-Commit Validate

## Steps

### Manual invocation (smaqit skill)

1. Run `git diff --cached --name-only` to obtain the staged file list.
2. Run all 4 checks below against staged files.
3. Report PASS or FAIL per check. For any failure, list the affected filenames and issue details.

### Validation checks

**Check 1 — No `.env` files staged** *(FAIL)*
```bash
git diff --cached --name-only | grep -E '\.env$|\.env\.'
```
FAIL if any match. Report each filename.

**Check 2 — No plaintext secrets** *(FAIL)*
```bash
git diff --cached | grep -nE '(sk-ant-|ANTHROPIC_API_KEY=sk-|BEGIN.*PRIVATE KEY|(password|secret|token)\s*=\s*[^\$\{]{8,})'
```
FAIL if any match. Report the file path and the matching line(s).

**Check 3 — No draft specs on main/merge** *(WARN)*
Run only when `git rev-parse --abbrev-ref HEAD` returns `main` or `MERGE_HEAD` exists:
```bash
git diff --cached --name-only | grep -E '^specs/.*\.md$' | xargs grep -l 'status: draft' 2>/dev/null
```
WARN — do not block. Surface as an advisory note only.

**Check 4 — No large files staged (>1 MB)** *(WARN)*
```bash
git diff --cached --name-only | xargs -I{} sh -c 'test -f "{}" && du -k "{}" | awk "{if (\$1 > 1024) print \"{}\"}"'
```
WARN — do not block. Advise using Git LFS.

### Hook installation

Run `scripts/install.sh` from the repository root once per clone. It copies
`scripts/pre-commit.sh` to `.git/hooks/pre-commit` and makes it executable.
The installed hook runs checks 1–3 automatically on every `git commit`.

Read `scripts/install.sh` when performing or reviewing hook installation.
Read `scripts/pre-commit.sh` when reviewing or modifying the hook implementation.

## Output

- Validation report printed to stdout: PASS or FAIL per check
- `.git/hooks/pre-commit` created when `scripts/install.sh` is executed
- `scripts/install.sh` — bundled install script (symlinks hook into `.git/hooks/`)
- `scripts/pre-commit.sh` — shell-only hook implementation (runs without agent context)

## Scope

- Inspects staged changes only (`git diff --cached`). Unstaged working tree files are not checked.
- Does not block commits on non-main branches for draft spec status (warning only, never a failure).
- Catches common secret patterns — not a comprehensive secrets scanner.
- `.git/hooks/` is not tracked by version control; `scripts/install.sh` must be re-run on each new clone.

## Examples

**Input:** Developer runs `git add backend/.env && git commit -m "fix"`.

**Output:**
```
FAIL: .env file(s) staged:
  backend/.env
Remove backend/.env from staged files before committing.
```

---

**Input:** Operator asks to install the pre-commit hook on a fresh clone.

**Output:** Run `bash .github/skills/smaqit.infrastructure-hook-pre-commit-validate/scripts/install.sh` from the repo root. Confirm: "pre-commit hook installed. Run `git commit` to test."

## Gotchas

- `.git/hooks/` is not tracked by version control. The install step must be documented in the project README under "Development Setup" and re-run on each new clone.
- `git add -p` (partial staging) may leave secrets in the working tree that are not staged — this hook does not catch those.
- The Anthropic key prefix pattern is `sk-ant-`. If Anthropic changes their key format, update the pattern in `scripts/pre-commit.sh`.
- The draft spec check is advisory — committing draft specs during active development is expected and acceptable.
- Check 2 uses `\s*` which is a GNU grep extension. It works correctly on Linux (Ubuntu). For macOS compatibility, replace `\s` with `[[:space:]]` in `scripts/pre-commit.sh`.

## Completion

- [ ] `git diff --cached --name-only` staged file list obtained
- [ ] All 4 validation checks executed
- [ ] PASS/FAIL report delivered per check with filenames for failures
- [ ] `scripts/install.sh` and `scripts/pre-commit.sh` present in the skill `scripts/` directory

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| `.env` file staged | FAIL — block commit; report the filename |
| Plaintext secret pattern matched | FAIL — block commit; report the file and matching line |
| Draft spec on main/merge | WARN — do not block; surface as advisory note |
| Large file staged (>1 MB) | WARN — do not block; advise using Git LFS |
