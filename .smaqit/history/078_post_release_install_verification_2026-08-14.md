# Post-Release Global Install Verification

**Date:** 2026-08-14
**Session focus:** Verify the v3.1.0 global install against its own release notes and confirm task 107's changes shipped correctly.
**Tasks completed:** None (verification-only session; task 107 was completed in the prior session)

## Actions taken

- Confirmed `main`, `origin/main`, and the `v3.1.0` tag are aligned after the release PR (#82) merged, and that task 107's commits (skill tree merge, `AGENTS.md` canonicalization) landed in the release alongside task 106 (Python/Tornado rsync deployment skill).
- Fetched the GitHub release notes for v3.1.0 and cross-checked them against the actual post-`curl install.sh` state of every global install directory: `~/.copilot/agents/`, `~/.claude/{agents,commands,skills}/`, `~/.codex/agents/`, `~/.agents/skills/`.
- Verified the shared skill tree contains exactly 27 smaqit-owned skills (26 + the new Tornado deploy skill from task 106), with no leftover `skills-copilot`/`skills-codex` split artifacts, and that `smaqit.feature-new`'s installed amendment-gate path correctly resolves to `${HOME}/.agents/skills` (task 107's fix).
- Confirmed the 29 additional entries under `~/.agents/skills/` and the extra Codex/Claude agents beyond the 9 canonical ones belong to the separate `smaqit-extensions` tool sharing the same global directories by design, and were left untouched (older mtimes, predating this install run).
- Confirmed no stray project-local `AGENTS.md`/`CLAUDE.md` were written into `$HOME` by the global installer (correct — those are project-scoped, written only by `smaqit init`).
- Updated `.smaqit/compendium.md`: fixed a stale "26 shipped product skills" reference (now 27 after task 106) and added a new entry on distinguishing smaqit-owned global-install content from `smaqit-extensions` content when verifying a release.

## Problems solved

None — this was a clean verification pass; no discrepancies were found between the release notes and the actual installed state.

## Decisions made

- None specific to this session beyond the compendium documentation choice (see Files modified).

## Files modified

- `.smaqit/compendium.md` — fixed stale skill count in the task-lifecycle entry; added a new Installation-category entry on post-release install verification methodology

## Next steps

- Push `main` to `origin` if any local-only commits remain beyond what was already merged via PR #82 (verify with `git status`/`git log origin/main..main`).
- Continue with task 106 follow-up work if any remains (release notes show it shipped as part of v3.1.0).
- Active tasks unchanged from prior session: 048 (Currency Labels), 009 (deferred GitHub/CI/CD phases) in other projects.

## Session Metrics

- Tasks completed: 0 (verification session)
- Files modified: 1 (`.smaqit/compendium.md`)
- Verification scope: 4 global directories, 27 skills, 9 agents per platform, cross-checked against GitHub release notes for v3.1.0
