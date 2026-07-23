# Feature New Skill and Release

**Date:** 2026-07-23
**Session Focus:** Plan and implement Task 089 (`smaqit.feature-new` post-MVP workflow skill), incorporate review feedback on shipped-content hygiene, cut a local release, and diagnose a CI/release-trigger investigation that surfaced a real regression.
**Tasks Referenced:** 089
**Tasks Completed:** None (089 remains In Progress, Assisted mode — implementation complete and handed back to operator, not yet closed via `/task.complete`)

---

## Actions Taken

### Session Setup and Planning
- Loaded project context, planning state, and prior history via `session.start`.
- Ran `task.plan 089` against the task's pre-existing, fully-drafted design (Description, Design Decisions, Implementation Steps, 8 ACs from a prior session). Scored Complex (8 ACs, multi-area touch) and ran four parallel Explore subagents against the greenfield skill, `smaqit.input-deployment`/`smaqit.spec-status-update`, framework docs, and cross-platform skill install paths.
- Discovery surfaced two stale assumptions in the task file: Implementation Step 3 specified a literal `../` relative path for cross-skill script reference (repo convention is actually the `[SMAQIT_SKILLS_DIR]` placeholder, resolved at compile time by `scripts/generate-agents.py`); Implementation Step 6 assumed an existing framework-doc skill index that does not exist anywhere in the repo, for any skill.
- Asked the operator how to handle the nonexistent-index gap; operator chose redirecting to a one-line `README.md` mention. Updated the task file with both corrections before starting.

### Implementation (Task 089)
- Ran `task.start 089` (Assisted mode) — research map confirmed current, issue triage returned Clear (no third-party tools; the task is pure orchestration prose over existing smaqit skills).
- Authored `skills/smaqit.feature-new/SKILL.md`: six phases (Task Creation, Spec Revalidation, Development, Deployment, Validation, Close-out), modeled on `smaqit.new-greenfield-project`'s structure but dropping Requirements Extraction, from-scratch spec generation, and the dev-VM sweep. Deployment phase resolves `provisioning_mode` with an existing-target-first override, supports an explicit deploy-now/defer parameter, and runs the amendment gate (`check-amendments.sh`, referenced via the compile-time placeholder) on every cycle regardless of defer state.
- Authored `skills/smaqit.feature-new/references/phase-differences-from-greenfield.md` and added a one-line mention to `README.md`.
- Verified rather than assumed: ran `scripts/generate-agents.py` and confirmed `[SMAQIT_SKILLS_DIR]` resolves correctly across all three compiled platform trees; reproduced the originating false-negative live against a real downstream project's specs (`check-amendments.sh` returned `PASS` before, `AMENDMENTS FOUND` after rewriting its amendment into canonical tag form, with explicit operator approval for that cross-project edit); checked off all 8 ACs; left Findings blank per the task template's own instruction (task-complete's responsibility, not the implementer's).

### Review Feedback and Rework
- Operator flagged two defects: the references file leaked a consumer project's name and internal task numbers into shipped product content, and `SKILL.md` referenced `framework/AGENTS.md` — confirmed via `installer/main.go`'s `go:embed` manifest that `framework/` is never shipped to installed projects, and no other skill in the tree references it.
- Fixed by inlining the Incremental Spec Updates mechanics directly, then — per further operator direction — distilled the actually-relevant mechanics from `framework/AGENTS.md`, `ARTIFACTS.md`, and `PHASES.md` into a new self-contained `skills/smaqit.feature-new/references/spec-lifecycle-reference.md`, and rewrote `phase-differences-from-greenfield.md` as a bare declarative table with all "why" narrative and consumer-project grounding removed.

### Release v1.7.0
- Committed the feature (`5b8a62d`) and invoked the `smaqit-release-local` agent. It classified the change as MINOR (v1.7.0), prepared `CHANGELOG.md`/`installer/main.go`, and reported push was not possible from this sandbox (no SSH agent/askpass, `gh` token lacks write scope).
- Operator asked to try a WSLg GUI passphrase unlock; investigation found no askpass tool installed (`zenity`, `kdialog`, `ssh-askpass*`, `python3-tkinter` all absent) and that installing one via `sudo apt` is blocked outright by the Claude Code auto-mode permission classifier. Operator chose to skip the GUI path and push from their own shell instead.
- Release agent committed `29fe606` ("Release v1.7.0"), tagged `v1.7.0`, and verified the build locally; push was left to the operator.

### Post-Push Diagnosis
- Operator reported the push succeeded but no `post-merge-release` run appeared. Diagnosis via `gh api` (SSH still unavailable in this sandbox) found the tag had not been pushed yet in the first pass — `post-merge-release.yml` triggers only on a `v*` tag push (or a merged PR to main), not a plain branch push.
- While diagnosing, found a real regression: `Installer Smoke Test` failed in CI because `installer/main_test.go`'s `TestRemoveEmbeddedSkillDirsPreservesUnownedCodexContent` hardcoded a stale count of 24 Codex skills (now 25 with `smaqit.feature-new` shipped). Fixed and committed (`ad58ad5`).
- With explicit operator approval (tag delete/recreate requires per-operation approval per prior feedback), moved the local `v1.7.0` tag to point at the fix commit and re-verified the build.
- Operator pushed again; further diagnosis found `post-merge-release` had in fact fired (`gh run list` lagged a few minutes behind actual trigger time) and had *already published* the `v1.7.0` GitHub Release from `ad58ad5` — but `Installer Smoke Test` failed *again* on the same commit, from a second, independent hardcoded skill count in `scripts/smoke-test-installer.sh` that the first fix missed.
- Fixed and verified locally (`make smoke-test` passes, reporting 25 skills; `go test ./...` passes); committed as `123fa7a`. Because `v1.7.0` was by then a live published release, declined to move the tag a second time — recommended a `v1.7.1` patch release instead, and left that decision with the operator (session ended before it was made).

## Problems Solved

- **Task 089's pre-drafted plan contained two stale/incorrect assumptions** (wrong cross-skill reference convention, nonexistent framework-doc index) that would have shipped a broken skill reference and an unfulfillable documentation step — caught via Discovery before implementation began, not after.
- **Shipped skill content leaked internal/consumer-project context** and depended on a repo directory (`framework/`) never distributed to installed projects — both fixed with a self-contained, declarative reference file.
- **Two independent hardcoded Codex-skill-count assertions** (`installer/main_test.go`, `scripts/smoke-test-installer.sh`) went stale in the same release, each requiring its own fix — the second was only found because the release-trigger diagnosis kept going past the first apparent fix.
- **A live, already-published GitHub Release's tag was correctly left untouched** rather than rewritten a second time, once it was confirmed public — avoiding a real risk of publishing/tag divergence.

## Decisions Made

- Cross-skill script references use the `[SMAQIT_SKILLS_DIR]` placeholder, never a literal relative path — corrected in the task file before implementation, not discovered mid-build.
- The missing framework-doc index gap was resolved by redirecting to a one-line `README.md` mention rather than fabricating a new registry section that no other skill uses.
- Shipped skill/reference content must never name a specific consumer project or reference smaqit's own internal task-tracking numbers — genericize the rationale instead.
- Shipped skills must be self-contained with respect to `framework/`, which is never installed into consumer projects; any needed mechanics get distilled into the skill's own `references/`.
- A published git tag (confirmed via a live GitHub Release) is not moved a second time — a new patch release is the correct path forward instead.

## Files Modified

### Task 089 implementation
- `skills/smaqit.feature-new/SKILL.md` — new skill, six-phase post-MVP workflow
- `skills/smaqit.feature-new/references/phase-differences-from-greenfield.md` — declarative phase-mapping table
- `skills/smaqit.feature-new/references/spec-lifecycle-reference.md` — distilled, self-contained spec-lifecycle mechanics (no `framework/` dependency)
- `README.md` — one-line `smaqit.feature-new` mention
- `.smaqit/tasks/089_feature_new_post_mvp_workflow_skill.md` — corrected Implementation Steps, triage block, all 8 ACs checked
- `.smaqit/tasks/PLANNING.md` — status → In Progress

### Release and regression fixes
- `CHANGELOG.md`, `installer/main.go` — v1.7.0 version bump
- `installer/main_test.go` — corrected hardcoded Codex skill count (24 → 25)
- `scripts/smoke-test-installer.sh` — corrected second, independent hardcoded Codex skill count (24 → 25)

### Cross-project (with explicit operator approval)
- A downstream project's `specs/stack/platform-stack.md` — rewrote existing prose amendment into the canonical `<!-- amendment: DATE — description -->` tag (2 lines added; unrelated in-flight changes in that repo untouched)

## Next Steps

- Decide whether to cut a `v1.7.1` patch release on top of `123fa7a` (recommended) rather than moving the now-published `v1.7.0` tag again.
- Push `123fa7a` to `main` (currently local-only).
- Task 089 remains In Progress (Assisted mode) — run `/task.complete 089` once satisfied with the implementation.
- Operator noted more changes are planned for this area in an upcoming session.

## Session Metrics

- **Tasks completed:** 0 (089 implemented, not yet closed)
- **New skills shipped:** 1 (`smaqit.feature-new`)
- **Feature commits:** 1
- **Fix commits:** 2
- **Release commits:** 1
- **Release:** v1.7.0 (published; one follow-up fix commit not yet included in a tag)
- **Cross-project edits (approved):** 1
- **Regressions found and fixed:** 2 (independent hardcoded skill-count assertions)
