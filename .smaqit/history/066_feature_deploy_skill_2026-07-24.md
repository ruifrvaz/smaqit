# Feature Deploy Skill

**Date:** 2026-07-24
**Session Focus:** Clean external project references from the smaqit repository, design and implement `smaqit.feature-deploy` — a standalone post-MVP deployment skill — and cut release v1.9.0.
**Tasks Referenced:** 091
**Tasks Completed:** 091

---

## Actions Taken

### Project Cleanup
- Removed all external project names (`magnificah`, `fashion-app-poc`, `iodis-crm-poc`, `areaoffice-poc`, `assistente-escolas-poc`, `hello-mario` as a Vault project slug) from 14 files across shipped artifacts, task files, and history files. `mario-hello` as smaqit's own test fixture was preserved.
- Sanitized CHANGELOG.md, `vault-loader/SKILL.md`, `rotate-credential.sh`, and completed task files (084, 085, 086, 089, 090).

### Deployment Entry-Point Assessment
- Identified a gap: no clear entry point for deploying post-MVP features. `smaqit.new-greenfield-project` is a full lifecycle, `smaqit.feature-new` runs all 5 phases, and `/smaqit.deployment` lacks the operational knowledge encoded in greenfield's Phase 4/5 provisioning_mode branching.
- Designed a planner/executor split: `smaqit.feature-deploy` as the domain-aware planner, `/smaqit.deployment` as the agnostic executor (zero agent modifications).
- Defined the architecture: Pre-phase (feature identification) → Phase 1 (infrastructure readiness) → Phase 2 (CI/CD deploy) → Phase 3 (close-out).

### Task 091 — `smaqit.feature-deploy` Implementation
- Used `smaqit.create-skill` + `smaqit.L2` to author and compile the skill from a definition file.
- Extracted greenfield Phase 4/5 provisioning_mode branching (provision, existing-owned, existing-shared) with all per-branch callouts.
- CI/CD-only deploy: pushes to main, monitors existing GitHub Actions workflow — no deploy-rsync skills, no dev-VM sweep.
- Invokes `/smaqit.deployment` for spec validation and reports only — agent is not modified.
- Initially placed in `.github/skills/` (consumer-project path); corrected to `skills/` (product skill path).
- Added installation wiring: bumped hardcoded skill counts in `installer/main_test.go` and `scripts/smoke-test-installer.sh` from 25→26.
- Self-containment fix: copied `check-amendments.sh` into the skill's own `scripts/` directory rather than cross-referencing `smaqit.new-greenfield-project`'s copy.
- Added README mention alongside `smaqit.feature-new`.

### Release v1.9.0
- Assessed changes as MINOR (new feature, backward-compatible).
- Updated CHANGELOG.md and `installer/main.go` Version const.
- Verified build and tests pass.
- Committed (`e22f851`), tagged (`v1.9.0` annotated), pushed main and tag via GNOME Keyring SSH agent.

## Problems Solved

- **Deployment entry-point gap** — users had no clear way to deploy a post-MVP feature. `smaqit.feature-deploy` provides a single `/smaqit.feature-deploy` command that handles the full deployment lifecycle.
- **Cross-skill file references** — the compiled skill initially referenced `check-amendments.sh` from `smaqit.new-greenfield-project/scripts/`. Corrected to ship its own copy, making the skill fully self-contained.
- **Wrong skill location** — initially compiled to `.github/skills/` (consumer-project install path). Corrected to `skills/` (canonical product skill path).
- **Stale skill count assertions** — `main_test.go` and `smoke-test-installer.sh` had hardcoded 25; bumped to 26 to match the new 26th skill.

## Decisions Made

- **Planner/executor split**: `smaqit.feature-deploy` plans and orchestrates; `/smaqit.deployment` executes agnostically. The deployment agent receives an explicit plan rather than deriving one from abstract phase descriptions. Zero modifications to `agents/deployment.md`.
- **CI/CD-only deploy**: Post-MVP deploys go through existing GitHub Actions workflows. No deploy-rsync skills, no dev-VM sweep. If CI/CD workflows don't exist, the skill stops and flags it.
- **Self-contained skill convention**: Shipped skills must own their own scripts, references, and assets — cross-skill file references are prohibited.
- **Provisioning mode default-override lives in feature-deploy**, not in `smaqit.input-deployment`.
- **Release tagging is conditional**: skipped when invoked from within `smaqit.feature-new`.

## Files Modified

### New files
- `skills/smaqit.feature-deploy/SKILL.md` — compiled product skill (153 lines)
- `skills/smaqit.feature-deploy/scripts/check-amendments.sh` — self-contained amendment scanner
- `.smaqit/definitions/skills/smaqit.feature-deploy.md` — skill definition (source of truth)
- `.smaqit/tasks/091_smaqit_feature_deploy_skill.md` — task file

### Modified files
- `CHANGELOG.md` — promoted `[1.9.0]` section
- `installer/main.go` — Version bumped to `1.9.0`
- `installer/main_test.go` — skill count 25→26
- `scripts/smoke-test-installer.sh` — skill count 25→26
- `README.md` — added `smaqit.feature-deploy` mention
- `.smaqit/tasks/PLANNING.md` — task 091 added and completed

### Cleaned files (external project references removed)
- CHANGELOG.md, `vault-loader/SKILL.md`, `rotate-credential.sh`
- Task files: 084, 085, 086, 089, 090
- History files: 062, 064, 065

## Next Steps

- Remaining active tasks: 077 (Retroactive Specifications), 074 (Extensible Through Templates), 071 (Q&A Agent), 070 (E2E Boundary Enforcement)
- `smaqit.feature-new` may eventually delegate its Phase 3 (Deployment) to `smaqit.feature-deploy`

## Session Metrics

- **Tasks completed:** 1 (091)
- **New files created:** 4 (SKILL.md, check-amendments.sh, definition, task)
- **Files modified:** 6
- **Files cleaned:** 14 (external project references)
- **Release:** v1.9.0 (published, `post-merge-release` triggered)
- **Agent modifications:** 0
