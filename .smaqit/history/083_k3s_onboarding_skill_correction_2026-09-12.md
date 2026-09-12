# K3s Onboarding Skill Correction

## Metadata

- **Date:** 2026-09-12
- **Session focus:** Planned, implemented, mid-review corrected, and released task 116 (`smaqit.infrastructure-onboard-k3s-app`); reconciled a concurrent-session conflict discovered along the way; redacted four leaked real downstream-project names from the repo.
- **Tasks completed:** 116 (released v3.4.0)
- **Tasks referenced:** 114, 115 (unstarted, blocked on separate scoping conversations), 117/118 (concurrent session's work, explicitly left for the user to reconcile)

## Actions Taken

### Task 116 — Contribute a k3s App-Onboarding Skill Definition
- Planned via `task.plan`: initial plan was a definitions-only contribution mirroring the `deploy-rsync-python-tornado` precedent. During clarification, the user's actual goal (a skill usable standalone *and* wired into `smaqit.new-greenfield-project`/`smaqit.feature-new` as a deployment-strategy alternative) triggered a `session.assess` — Discovery (3 parallel Explore agents) found that "app onboarding" (cluster-repo-side tenancy grant) and "app deployment" (app-repo-side workload push) are structurally distinct concerns, and that `provisioning_mode`/Phase 4 Step 6 are pervasively VM-shaped with no cluster analogue. User confirmed task 116 should stay onboarding-only ("full stop"), then separately pushed back on the plan's "never compiled" design decision — scope widened to compile and ship the skill as a real product capability, not just a definitions file, given the source mechanism's proven track record. Renamed the skill from `smaqit.infrastructure-deploy-k3s` to `smaqit.infrastructure-onboard-k3s-app` before implementation to correctly signal onboarding vs. deployment.
- Started via `task.start`; refreshed `.smaqit/references/project-research.md` (19 days stale) as part of the triage gate, adding a missed project-layer tool (`github-copilot-sdk`) and task 116's technology block.
- Implemented: authored the definitions file and compiled skill with the full detailed mechanism (registry file + converge workflow, Namespace/RBAC/PSA/NetworkPolicy specifics, six real bugs with fixes, two known gaps), bumped installer skill-count assertions 27→28, added a CHANGELOG entry. `make -C installer test`/`smoke-test` both passed.
- Completed via `task.complete` Phase 1: `release-analysis` suggested MINOR/v3.4.0; PR #87 opened, changelog pending entry promoted.
- **Mid-review correction:** while reviewing PR #87, the user asked which operator would invoke the skill, then identified the whole design as wrong — a generic smaqit product skill cannot prescribe one specific downstream org's onboarding mechanics (registry schema, RBAC scheme, its own six gotchas) as a universal contract; each infra-owning repo owns its own onboarding contract, per the machine-monorepo pattern's own principle. Corrected in place on the same branch: deleted the definitions file entirely, rewrote the compiled skill as a thin dispatcher (recognize the target cluster is infra-repo-owned, defer entirely to that repo's own onboarding skill/workflow/instructions, store the resulting credential via `smaqit.infrastructure-vault-loader`), dropped `validated`/`validated-stack` metadata (doesn't apply to a generic dispatcher). Re-ran the full test/smoke-test cycle — both pass. Updated the PR description via `gh api` directly (`gh pr edit` hit an unrelated GraphQL Projects-classic deprecation error on this repo).
- **Concurrent-session conflict found and flagged, not resolved here:** task 118 (child of task 117, built by a different session) was designed assuming task 116's *original* detailed mechanism was a legitimate "direct-access playbook" it complements with a no-direct-access PR-based alternative. That assumption broke once the detailed mechanism was deleted. Reported to the user, who took ownership of reconciling task 117/118 separately (confirmed later in-session as resolved on their end: task 118 completed, task 117 kept declarative).
- Completed via `task.complete` Phase 2 after the user merged PR #87: confirmed `MERGED` via `gh pr view`, pulled `main`, marked `Completed`, removed the worktree, force-deleted the local branch, rebuilt the workspace (carefully staging only the workspace file, since the concurrent session had its own uncommitted task 118/`PLANNING.md` edits in the same working tree at that moment). Confirmed `post-merge-release.yml` fired and completed: **v3.4.0** tagged and published.

### Downstream Project Name Redaction
- User requested cleanup of four real downstream-project names that had leaked into task files, a history entry, and a *shipped* skill, in four separate incremental requests: `Magnificah`/`Magnificah/infrastructure` (tasks 114, 115, history 070, `PLANNING.md` — 9 occurrences across 4 files), `iodis-crm-poc` (tasks 110, 113, history 070 — 7 occurrences across 3 files, including a quoted `AGENTS.md` title that directly spelled the name), `fashion-app-poc` (history 070 and the **shipped** `smaqit.infrastructure-deploy-rsync-python-nextjs` skill's frontmatter `description:` and intro prose — the first two cleanups only touched internal task/history files, this one touched a real, installed product artifact), and `areaoffice-poc` (history 070 and task 110 — including a second quoted `AGENTS.md` title and its derived-slug example, both dropped rather than replaced with another fabricated example).
- Each pass: grepped the full repo (content and filenames) before and after, replaced with generic "a/two downstream project(s)" phrasing or restructured sentences where a real name was grammatically load-bearing, and committed+pushed separately per name as requested.

## Problems Solved

- Task 116 shipped a real, working, correctly-scoped `smaqit.infrastructure-onboard-k3s-app` skill — caught and corrected mid-review before a fundamentally wrong design (one org's implementation hardcoded as a universal contract) could ship in a release.
- A cross-session design conflict (task 118's dependency on task 116's soon-to-be-deleted mechanism) was surfaced before it caused silent breakage, rather than discovered later as a confusing inconsistency.
- Four real downstream-project-name leaks were removed from committed history, active task files, and — in `fashion-app-poc`'s case — a shipped, installable product skill.

## Decisions Made

- Onboarding (cluster-repo-side tenancy grant) and deployment (app-repo-side workload push) are permanently distinct concerns in smaqit's infrastructure-skill architecture; conflating them wastes planning effort (happened once, during task 116's own initial scoping).
- A smaqit product skill must never hardcode one specific downstream repo's implementation details as if they were a universal contract — the corrected task 116 design (thin dispatcher, defer entirely to the infra repo's own process) is the reusable pattern for this class of skill going forward.
- Detailed, org-specific mechanism content that doesn't generalize gets deleted from smaqit outright, not preserved as reference material "just in case."
- The existing branch/PR is corrected in place on a wrong-design finding, not closed and reopened — preserves review history and avoids redundant PR churn.

## Files Modified

| File | Action |
|------|--------|
| `.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md` | Created, then deleted on correction |
| `skills/smaqit.infrastructure-onboard-k3s-app/SKILL.md` | Created (detailed mechanism), then rewritten (thin dispatcher) |
| `installer/main_test.go`, `docs/wiki/workflows/testing-smaqit.md` | Skill-count assertions bumped 27→28 |
| `CHANGELOG.md` | v3.4.0 entry added, then rewritten to match the corrected design |
| `.smaqit/tasks/116_contribute_k3s_app_onboarding_skill_definition.md`, `.smaqit/tasks/PLANNING.md` | Full task lifecycle (start → PR Open → Completed) |
| `.smaqit/references/project-research.md` | Refreshed (was 19 days stale); added `github-copilot-sdk` and task 116's technology block |
| `.smaqit/tasks/114_*.md`, `.smaqit/tasks/115_*.md` | Magnificah/infrastructure references redacted |
| `.smaqit/tasks/110_*.md`, `.smaqit/tasks/113_*.md` | iodis-crm-poc and areaoffice-poc references redacted |
| `skills/smaqit.infrastructure-deploy-rsync-python-nextjs/SKILL.md` | fashion-app-poc references redacted (shipped skill) |
| `.smaqit/history/070_release_and_project_sanitization_2026-07-29.md` | All four names redacted from its own audit-tally lines |
| `smaqit.code-workspace` | Rebuilt after task 116 worktree removal |

## Next Steps

- Task 117/118 reconciliation — confirmed by the user as resolved on their end during this session (task 118 completed, task 117 kept declarative); no further action expected from this repo's task backlog on that front.
- Task 114 remains blocked on a scoping conversation (one of its acceptance criteria belongs in `smaqit-extensions`, not this repo) — flagged at session start, not addressed this session.
- Task 113 remains the most immediately actionable, self-contained backlog item (no blockers, same shape as already-shipped tasks 111/112).
- Tasks 071 and 074 both appear already satisfied by current code/framework state (found during session-start's codebase read) — worth a reconciliation pass to close them out rather than leaving them listed as active.

## Session Metrics

- Tasks completed: 1 (task 116 → v3.4.0)
- PRs opened/merged: 1 (#87)
- Releases published: 1 (v3.4.0)
- Design corrections caught during review (before shipping to a wider audience): 1 (task 116's full mechanism rescoped to a thin dispatcher)
- Cross-session conflicts found and handed off: 1 (task 118's dependency on task 116's deleted mechanism)
- Real downstream-project names redacted: 4 (`Magnificah`, `iodis-crm-poc`, `fashion-app-poc`, `areaoffice-poc`) across 8 files, including 1 shipped product skill
