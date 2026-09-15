# K3s Image Build-Push Hardening

## Metadata

- **Date:** 2026-09-15
- **Session focus:** Task 121 — harden the k3s app-deployment skill family (image build-push, lint, and Vault conventions)
- **Tasks completed:** 121 (released v3.6.0, PR #91)

## Actions Taken

- Ran `session.start`, loading project state, memory, `PLANNING.md`, and the compendium. Identified task 121 (created the previous session, never formally planned) as the most urgent ready candidate — seven live-verified findings from a downstream project's first real `existing-k3s` deployment — over task 113's older, lower-urgency diagram-authoring gap.
- Planned task 121 via `smaqit.task-plan`: read the downstream project's (magnificah-website) task 005 file and its synthesized skill in full, plus every smaqit file the task touches, before drafting a phased plan. Discovery surfaced two refinements not in the original task text: (1) the build-and-push CI wiring belongs in `deploy.yml.k3s.template` itself, not the reconciled skill, since the template already carried an unfilled `__IMAGE__` token; (2) the Infrastructure spec template needed a new `Container Registry` Constraints row so `repo-config`'s "if the spec names a private registry" condition had something to key off. User reviewed and adjusted the plan (Phase A should go through `smaqit.create-skill`, not hand-authored files) before approval.
- Updated the task file in place with the phased Design Decisions/Implementation Steps/ACs, then started task 121 (`smaqit.task-start`): created its branch/worktree, refreshed the project-research map (added Docker/GHCR doc entries), and ran issue triage (Result: Historical — a closed Docker issue confirming the lowercase-tag requirement is documented behavior, not a regression).
- Implemented across six phases:
  - **Phase A:** Reconciled `smaqit.infrastructure-image-build-push-static` via `smaqit.create-skill` → `smaqit.L2`, after discovering `smaqit.create-skill`'s documented output paths (`.agents/skills/`, `.claude/skills/`) are the wrong convention for this framework's own source repo — worked around by invoking `smaqit.L2` directly against a definition file at this repo's real convention (`skills/<name>/SKILL.md`).
  - **Phase B:** Added a `build` job to `deploy.yml.k3s.template` (GHCR login, lowercase-tag fix, push by SHA) and wired `__IMAGE__` substitution into the existing stamp step; validated with `actionlint`/`shellcheck` and a fresh throwaway-slug generation.
  - **Phase C:** Extended `manifest-lint.sh` to require an effective numeric `runAsUser` wherever `runAsNonRoot` is true, mirroring the script's existing pod/container inheritance pattern; verified against 5 fixture manifests.
  - **Phase D:** Added the `Container Registry` Constraints row, documented the new org-scoped `organizations/<org-slug>/github-package-read` Vault path, and sharpened `repo-config`'s sync step.
  - **Phase E:** Added two new Gotchas to `cicd-generate` (Ingress-must-render-first, `workflow_dispatch` default-branch requirement), cross-referenced them from `new-greenfield-project`, and reordered `deploy-k3s-app`'s Examples to CI-first.
  - **Phase F:** Bumped two hardcoded skill-count assertions in `installer/main_test.go` (29→30) after `make prepare`; `go test`/`go vet` passed; rebuilt and ran `--install-global`, confirmed all changes live.
- Completed task 121 (`smaqit.task-complete`): Phase 1 committed implementation, computed the release version via `smaqit.release-analysis` (Task mode, MINOR severity — new skill, new CI capability, new spec field), opened PR #91 ("Prepare release v3.6.0"), wrote and promoted the CHANGELOG entry. Assisted mode stopped for review.
- Diagnosed and resolved a CI failure the user flagged mid-review: 3 of 4 CodeQL matrix jobs failed at the SARIF-upload step with no error text or real findings — traced to a transient GitHub default-setup upload hiccup (confirmed via a prior PR's clean run and the `go` job succeeding in the same run), not anything in the PR's own content. Attempted `gh run rerun`/the check-runs `rerequest` API (both refused — no committed CodeQL workflow file, and no App-level token), then pushed an empty commit to retrigger the dynamic scan; the retry passed. User merged the PR. Phase 2 verified the merge via `gh pr view`, merged `origin/main` into local `main`, marked the task Completed, and cleaned up the worktree/branch.

## Problems Solved

- **`smaqit.create-skill` output-path mismatch:** the generic downstream-consumer skill-authoring tool writes to `.agents/skills/`/`.claude/skills/`, which don't exist in this framework's own source repo (its canonical skill source is `skills/<name>/SKILL.md`). Caught before any file was misplaced; resolved by invoking `smaqit.L2` directly against a correctly-located definition file.
- **`smaqit.create-skill`'s validator flagged a false positive:** its `validate-skill.go` rejects a "Use when..." description opening as a conversational anti-pattern, but two sibling canonical skills in this same repo already violate that exact rule — kept the new skill's description consistent with its siblings rather than making it a style outlier.
- **CodeQL default-setup upload flake:** diagnosed as transient (not a real security finding, not caused by this PR) by comparing against the previous PR's clean run and confirming the `go` job's upload succeeded in the same run. No committed CodeQL workflow exists in this repo (GitHub's dynamic default setup), so neither `gh run rerun` nor the check-runs `rerequest` API could force a retry — an empty commit to the branch was the working fix.
- **Anticipated but avoided merge conflict:** the task branch's frozen git history (forked right after task 121's own "create task" commit, before that commit reached `origin/main`) meant PR #91's file diff included the task file/`PLANNING.md` with stale pre-implementation content. Flagged as a likely add/add conflict risk for Phase 2 ahead of time; git's `ort` merge strategy resolved it cleanly with no manual intervention needed.

## Decisions Made

- Kept the reconciled skill name `-static`-suffixed (no second real non-static use case to generalize from yet), per the task's own Design Decision.
- Build-and-push CI wiring (the lowercase-tag fix included) lives in the generic `deploy.yml.k3s.template`, not the reconciled skill — a discovery-driven scope refinement that makes every future k3s project inherit the fix for free rather than depending on which image-build skill gets matched/synthesized.
- Added the `Container Registry` Infrastructure-spec Constraints row as new, discovery-driven scope (not in the original task text) — mirrors task 120's own precedent for other existing-k3s fields.
- No new automated test harness invented for `manifest-lint.sh` — verified with throwaway fixture manifests instead, consistent with no skill-bundled script in this repo having automated tests today.
- `organizations/<org-slug>/github-package-read` deliberately excluded from `rotate-credential.sh`'s supported paths (org-scoped, not app/machine-scoped) — rotation stays the same manual `vault kv put`/`delete` used to populate it.

## Files Modified

- `skills/smaqit.infrastructure-image-build-push-static/SKILL.md`, `.smaqit/definitions/skills/smaqit.infrastructure-image-build-push-static.md` — new canonical skill + provenance file
- `skills/smaqit.infrastructure-cicd-generate/SKILL.md`, `assets/deploy.yml.k3s.template` — `build` job, `__IMAGE__` substitution, two new Gotchas
- `skills/smaqit.infrastructure-deploy-k3s-app/scripts/manifest-lint.sh`, `assets/deployment.yaml.template`, `SKILL.md` — `runAsUser` lint check, Examples reorder
- `skills/smaqit.infrastructure-vault-loader/SKILL.md` — new `organizations/` Vault path documentation
- `skills/smaqit.infrastructure-repo-config/SKILL.md` — sharpened private-registry credential sync
- `skills/smaqit.new-greenfield-project/SKILL.md` — k3s sequence cross-references
- `templates/specs/infrastructure.template.md`, `agents/infrastructure.md` — new `Container Registry` Constraints row + MUST rule
- `installer/main_test.go` — bumped hardcoded skill-count assertions (29→30)
- `CHANGELOG.md` — v3.6.0 entry
- `.smaqit/tasks/121_harden_k3s_app_deployment_family_image_build_push_lint_vault.md`, `.smaqit/tasks/PLANNING.md` — task lifecycle bookkeeping
- `.smaqit/references/project-research.md` — task 121 research block (Docker, GHCR doc entries)
- `smaqit.code-workspace` — worktree lifecycle bookkeeping

## Next Steps

- Whether to broaden `smaqit.infrastructure-image-build-push-static` beyond "static site, no build step of its own" remains open — resolve once a second real (non-static) use case exists.
- `rotate-credential.sh` doesn't support `organizations/<org-slug>/github-package-read` — acceptable today since rotation is rare and manual; revisit if that changes.
- Carried over from prior sessions, still unfixed and un-filed as a task: the installer skill-retirement pruning gap (`cmdInstallGlobal`/`copyEmbeddedDir` never deletes a directory dropped from the current embed).
- `PLANNING.md`'s remaining Active candidates: task 113 (self-contained, no blockers) is the most ready pick; task 114 needs a scoping conversation first (115 likely depends on it); tasks 071/074 may already be satisfied by current code and worth a reconciliation pass.

## Session Metrics

- Tasks completed: 1 (task 121)
- Release shipped: v3.6.0 (PR #91)
- Files modified: 14 (12 source/skill files + CHANGELOG + task-bookkeeping) + 1 workspace-rebuild commit
- PRs opened/merged: 1
- CI issues diagnosed and resolved: 1 (transient GitHub default-setup CodeQL upload flake)
