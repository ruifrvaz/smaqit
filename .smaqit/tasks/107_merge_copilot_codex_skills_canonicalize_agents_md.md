# Merge Copilot/Codex Skills; Canonicalize AGENTS.md

**Status:** In Progress
**Created:** 2026-08-13
**Started:** 2026-08-14
**Mode:** Assisted

## Description

The installer generates and ships two byte-identical skill trees — `skills-copilot` and `skills-codex` — both installed to the same shared `~/.agents/skills/` path with the same `${HOME}/.agents/skills` self-reference substitution. The redundant `skills-codex` tree is embedded and staged but never actually installed globally (only used by legacy project-mirror cleanup and tests), making the split pure duplication with no functional difference.

Separately, the repo's own `.github/copilot-instructions.md` is stale: it still describes the pre-global-install (pre-task-105) project-local artifact layout, omits `skills/`, `.smaqit/definitions/`, and `templates/designs/`, and undercounts agents (states 8, actual is 9).

This task merges the redundant skill trees into a single `skills-shared` tree, deletes the now-fully-redundant `removeLegacyProjectMirrors()` legacy migration path (pre-8/11 project-mirror installs are resolved via a clean reinstall, not automated migration), and makes root `AGENTS.md` the single canonical instructions file for GitHub Copilot and Codex — replacing `.github/copilot-instructions.md` — while Claude Code keeps its existing thin `CLAUDE.md` → `@AGENTS.md` hook. Copilot is not deprioritized as a platform; only outputs that are provably identical across Copilot and Codex are merged. Agent formats remain fully separate (`agents-copilot/*.agent.md`, `agents-codex/*.toml`, `agents-claude/*.md`).

## Design Decisions

- **Merged skill tree name:** `installer/skills-shared/`, Go embed variable `skillFilesShared` — mirrors the existing `shared-skills` kind label in `resolveGlobalDir`.
- **Scope of the merge:** only skill trees and shared instructions merge. Agent rendering stays fully platform-specific (Copilot `.agent.md`, Codex `.toml`, Claude `.md`) — this is not a reduction of Copilot support, only removal of provably duplicate output.
- **`removeLegacyProjectMirrors()` is deleted outright**, along with its `cmdInit` call. Pre-8/11 users upgrading from the old project-local-mirror layout perform a clean reinstall; no automated migration path is retained.
- **`AGENTS.md` becomes canonical** for both GitHub Copilot and Codex (both read it natively). `.github/copilot-instructions.md` is deleted; its content is rewritten (not copy-pasted) into root `AGENTS.md` to reflect current v3 architecture (global install model, `skills/`, `.smaqit/definitions/`, `templates/designs/`, correct 9-agent count, `skills-shared` naming).
- **`CLAUDE.md` stays a thin `@AGENTS.md` hook** — no content duplication for Claude Code.
- **`copilot-setup-steps.yml` keeps shipping** — Copilot bootstrap workflow is platform-specific tooling, not first-class-instruction duplication, and is out of scope for removal.
- **Committed dogfood platform trees out of scope** — the 213 git-tracked files under `.github/skills/`, `.agents/skills/`, `.claude/skills/` are this repo's own live installs, not installer logic, and are not touched by this task.
- **Historical docs/logs untouched** — references to `copilot-instructions.md` in `docs/logs/` and `docs/wiki/designs/level-up-compilation.md` are historical record and not updated.

## Implementation Steps

1. **Generator merge** (`scripts/generate-agents.py`): replace the `skills-copilot` + `skills-codex` outputs with a single `installer/skills-shared/`; keep `skills-claude` unchanged. Collapse `copilot`/`codex` in `SKILLS_DIR_BY_PLATFORM` into one shared `${HOME}/.agents/skills` substitution; update `generate_skills()` to render the shared tree once and the Claude tree once. Update the module docstring (currently describes three skill output trees).
2. **Installer Go** (`installer/main.go`): replace the `skillFilesCopilot` + `skillFilesCodex` embeds with a single `//go:embed skills-shared` → `skillFilesShared`; update the `cmdInstallGlobal` and `cmdUninstall` mappings to use it for the `shared-skills` kind. Delete `removeLegacyProjectMirrors()` entirely and its call site in `cmdInit`. Update the `installInstructionsFile` comment ("AGENTS.md (read natively by GitHub Copilot)") to reflect both Copilot and Codex reading it natively.
3. **Build plumbing** (`installer/Makefile`): update the `prepare` target's echo message and the `uninstall`/clean target's directory lists (`skills-copilot`, `skills-codex` → `skills-shared`).
4. Regenerate gitignored build inputs: `make -C installer prepare`.
5. **Tests** (`installer/main_test.go`): update `TestRemoveEmbeddedSkillDirsPreservesUnownedCodexContent` (and its Codex-agent sibling if affected) to reference `skillFilesShared`/`skills-shared`. Remove/adjust any test relying on `removeLegacyProjectMirrors()`. Add a new invariant test (e.g. `TestSharedSkillsServeCopilotAndCodex`) that walks the embedded `skills-shared` tree and asserts: 26 top-level skill directories, no unresolved `[SMAQIT_SKILLS_DIR]` placeholders remain, and every resolved path is exactly `${HOME}/.agents/skills` (never `.github/skills`, never a Claude-specific path).
6. **Canonical instructions**: create root `AGENTS.md`, rewriting `.github/copilot-instructions.md` content to reflect current v3 architecture (global install paths, `skills/`, `.smaqit/definitions/`, `templates/designs/`, corrected 9-agent count, `skills-shared` naming). Delete `.github/copilot-instructions.md`. Create root `CLAUDE.md` containing the existing thin `@AGENTS.md` hook pattern (matching what the installer already ships to user projects).
7. **Reference sweep**: update remaining canonical references from `copilot-instructions.md` to `AGENTS.md` in: `skills/smaqit.infrastructure-vault-loader/SKILL.md`, `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh`, `skills/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh`, `.smaqit/definitions/skills/smaqit.infrastructure-repo-config.md`, `.smaqit/definitions/skills/smaqit.infrastructure-cicd-generate.md`, `templates/workflows/copilot-setup-steps.yml` (trigger comment only), `.github/scripts/SUMMARY.md`.
8. **Verify**: `cd installer && make prepare && go vet ./... && go test ./...`; then `make build && bash ../scripts/smoke-test-installer.sh`; then `git grep -l "copilot-instructions"` to confirm only historical docs remain and `ls installer` shows no `skills-copilot`/`skills-codex`.

## Known Issues Triage
**Triaged:** 2026-08-14
**Tools searched:** none
**Result:** Clear

No third-party tools identified — this task is an internal refactor of smaqit's own installer (Go embeds, `generate-agents.py`), affecting no new external library, service, or platform integration. Triage not applicable.

## Acceptance Criteria

- [ ] `scripts/generate-agents.py` produces exactly two rendered skill trees: `installer/skills-shared/` (Copilot + Codex, `${HOME}/.agents/skills` paths) and `installer/skills-claude/`; no `skills-copilot` or `skills-codex` output remains.
- [ ] `installer/main.go` embeds and installs only `skillFilesShared` for the `shared-skills` kind in both `cmdInstallGlobal` and `cmdUninstall`; `removeLegacyProjectMirrors()` and its call site are removed.
- [ ] `installer/Makefile` prepare/clean targets reference `skills-shared` only.
- [ ] `go test ./...` in `installer/` passes, including a new invariant test proving the shared tree serves both Copilot and Codex (26 skill dirs, no unresolved placeholders, all paths `${HOME}/.agents/skills`).
- [ ] Repo root has `AGENTS.md` with all coding-agent-specific instructions, reflecting current v3 architecture; `.github/copilot-instructions.md` is deleted; root `CLAUDE.md` contains the `@AGENTS.md` hook.
- [ ] No canonical source, skill, template, or definition file references `copilot-instructions.md` (historical docs/logs excluded); `copilot-setup-steps.yml` still ships unchanged in function.
- [ ] `bash scripts/smoke-test-installer.sh` passes: 26 shared skills present at `~/.agents/skills/`, Claude skills at the Claude-specific path, Copilot/Codex agents unaffected.

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

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
| `scripts/generate-agents.py` | Modify — merge skill output trees |
| `installer/main.go` | Modify — embeds, install/uninstall mappings, delete legacy mirror removal |
| `installer/main_test.go` | Modify — update skill tests, add invariant test |
| `installer/Makefile` | Modify — prepare echo + clean lists |
| `templates/workflows/copilot-setup-steps.yml` | Modify — trigger comment reference |
| `AGENTS.md` | Create — canonical instructions |
| `CLAUDE.md` | Create — thin `@AGENTS.md` hook |
| `.github/copilot-instructions.md` | Delete |
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Modify — reference update |
| `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` | Modify — reference update |
| `skills/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh` | Modify — reference update |
| `.smaqit/definitions/skills/smaqit.infrastructure-repo-config.md` | Modify — reference update |
| `.smaqit/definitions/skills/smaqit.infrastructure-cicd-generate.md` | Modify — reference update |
| `.github/scripts/SUMMARY.md` | Modify — reference update |

## Notes

Discovery confirmed (2026-08-13) that under the current generator, `skills-copilot` and `skills-codex` render byte-identical (`diff -rq` = 0) — the split is pure duplication, not intentional platform divergence. Committed dogfood platform trees (`.github/skills/`, `.agents/skills/`, `.claude/skills/` — 213 tracked files) are this repo's own live installs and are explicitly out of scope.

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
