# Copilot Codex Skills Unification

**Date:** 2026-08-14
**Session focus:** Assess, plan, implement, and complete task 107 — merge the duplicate Copilot/Codex skill renderings into one shared tree and make `AGENTS.md` the canonical instructions file.

**Tasks completed:** 107 (Merge Copilot/Codex Skills; Canonicalize AGENTS.md)

## Actions taken

- Assessed a downstream project's claim that `cmdInstallGlobal` ships only the Copilot-rendered skills to `~/.agents/skills/` while the Codex variant is never installed globally. Verified the claim structurally, then proved via regeneration that both renders are byte-identical under the current generator — the difference seen by the other project came from stale, gitignored build artifacts, not real per-platform variance.
- Planned task 107 (Mode A) with full discovery and three clarified design decisions: merged tree name, repo dogfood scope, and CLAUDE.md style.
- Created and started task 107 (Assisted), implemented all changes, verified, and completed it: merged branch into `main`, cleaned worktree, deleted branch, rebuilt workspace.

## Problems solved

- **Redundant skill trees:** `skills-copilot` and `skills-codex` were byte-identical renders of the same source; the Codex one was embedded and staged but never installed. Merged into a single `installer/skills-shared/` tree consumed by both Copilot and Codex.
- **Dead legacy migration path:** `removeLegacyProjectMirrors()` existed solely for pre-v3 project-mirror cleanup. Deleted outright per user decision that pre-8/11 upgrades are resolved via clean reinstall — no backwards compatibility retained.
- **Copilot-only instructions file:** `.github/copilot-instructions.md` was stale (described the pre-global-install layout, omitted `skills/` and `.smaqit/definitions/`, undercounted agents). Its content was rewritten into a root `AGENTS.md` (read natively by both Copilot and Codex), a thin `CLAUDE.md` (`@AGENTS.md`) added, and the old file deleted.
- **Stale `.gitignore`:** still listed `installer/skills-copilot/` and `skills-codex/` after the rename; fixed to cover `skills-shared/`.
- **Smoke-test mismatch:** a legacy-mirror-migration block exercised the now-deleted function; removed it.
- **Unenforced invariant:** added `TestSharedSkillsServeCopilotAndCodex` asserting 26 skill dirs, no unresolved `[SMAQIT_SKILLS_DIR]` placeholders, and no Claude/legacy paths in the shared tree.

## Decisions made

- Copilot is not deprioritized as a platform; only provably identical outputs (skills, shared instructions) merge. Agent formats stay fully platform-specific.
- Merged tree named `skills-shared` (embed `skillFilesShared`), mirroring the existing `shared-skills` kind.
- `CLAUDE.md` stays a thin `@AGENTS.md` hook; `copilot-setup-steps.yml` keeps shipping.
- Kept `smaqit.infrastructure-vault-loader`'s `AGENTS.md → CLAUDE.md → copilot-instructions.md` fallback chain: it is intentional compatibility logic for arbitrary downstream projects, not a stale self-reference.
- Left `.github/scripts/SUMMARY.md` and historical logs untouched (historical record).

## Files modified

- `scripts/generate-agents.py` — merged skill outputs into single `skills-shared` tree; docstring updated
- `installer/main.go` — single `skillFilesShared` embed; install/uninstall mappings updated; `removeLegacyProjectMirrors()` and call site deleted; comments corrected to "GitHub Copilot and Codex"
- `installer/main_test.go` — skill test renamed to `...SharedContent`; new invariant test added
- `installer/Makefile` — prepare echo and clean lists reference `skills-shared`
- `.gitignore` — `installer/skills-shared/` replaces the two old entries
- `scripts/smoke-test-installer.sh` — removed legacy-mirror-migration block
- `AGENTS.md` — created; canonical instructions for Copilot + Codex, rewritten for current architecture
- `CLAUDE.md` — created; thin `@AGENTS.md` hook
- `.github/copilot-instructions.md` — deleted
- `.github/workflows/copilot-setup-steps.yml`, `templates/workflows/copilot-setup-steps.yml` — trigger comment now references `AGENTS.md`
- `.smaqit/definitions/skills/smaqit.infrastructure-repo-config.md`, `smaqit.infrastructure-cicd-generate.md` — context resolution references `AGENTS.md`
- `.smaqit/tasks/107_*.md`, `.smaqit/tasks/PLANNING.md` — task state lifecycle

## Next steps

- Push `main` to `origin` (local branch is ahead; includes the task 107 merge).
- Consider a release once the installer merge settles (`smaqit.release.local` or `/smaqit.release.pr`).
- Optionally regenerate the repo's own committed dogfood platform skill trees to stay in sync with any future `skills/` source changes.

## Session Metrics

- Tasks completed: 1 (107)
- Files changed in implementation: 13 (249 insertions, 238 deletions)
- Verification: `go vet` clean, `go test ./...` pass (incl. new invariant test), `make build` success, sandboxed smoke test pass
