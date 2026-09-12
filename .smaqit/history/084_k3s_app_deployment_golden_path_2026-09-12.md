# K3s App Deployment Golden Path

## Metadata

- **Date:** 2026-09-12
- **Session focus:** Implemented and shipped the app-side half of smaqit's k3s deployment golden path: task 117 (`smaqit.infrastructure-deploy-k3s-app`) and its child task 118 (`smaqit.infrastructure-request-k3s-onboarding`), released together as v3.5.0
- **Tasks completed:** 117 (owner, released v3.5.0), 118 (child of 117, bookkeeping-only)
- **Tasks referenced:** 116 (concurrent session, corrected mid-review to a thin dispatcher — directly informed this session's design and required a reconciliation pass)

## Actions Taken

### Task 117 — k3s App-Deployment Skill With Routing and CI/CD
- Assessed the pre-existing task file, started it in Assisted mode, ran research-map refresh and issue triage (Advisory result — a Traefik multi-`tls[]` issue and a Kubernetes RBAC name-length DoS, both confirmed non-applicable to this task's actual design; two closed issues flagged as live-verification sanity checks).
- Delegated the bulk mechanical implementation to a background agent operating in the task worktree: the new `smaqit.infrastructure-deploy-k3s-app` skill (`namespace-guard.sh`, `manifest-lint.sh`, Ingress/Deployment templates), `provisioning_mode: existing-k3s` wiring into `smaqit.input-deployment`/`smaqit.new-greenfield-project`/`smaqit.feature-new`, and a `k3s` mode in `smaqit.infrastructure-cicd-generate`. Reviewed every file directly and independently reran `go vet`/`go test`/the installer build rather than trusting the agent's self-report — caught and had it fix a real bug in `manifest-lint.sh` (container-only Pod Security fields incorrectly required at the Pod level).
- Vault kubeconfig storage was revised twice, both per explicit user direction: first added entirely (the plan had referenced a `kubeconfig` credential `smaqit.infrastructure-vault-loader` had no concept of), then restructured mid-implementation from a flat `secret/apps/<app-slug>/kubeconfig` (test/prod fields) to a machine-keyed `secret/apps/<app-slug>/<machine-slug>/kubeconfig` (one value per machine-slug, no `secret/machines/*` counterpart) after identifying the flat shape couldn't represent an app onboarded onto genuinely different machines per environment.

### Discovery — the k3s onboarding 3-way decomposition
- User asked directly whether both onboarding and deploy were wired into the greenfield/feature-new flows. Assessment revealed a conflation: there are three distinct concerns — (1) the infra repo's own onboarding automation (task 116's playbook), (2) an app-side request-and-gate step (missing), (3) the app-side deploy mechanism (task 117). Corrected this reasoning directly when the user pointed it out.
- User proposed simplifying the missing piece to a PR-based request model (open a PR with a registry entry, gate on merge) rather than direct commit/`workflow_dispatch` access — assessed as sound and much lighter-credentialed than the alternative.

### Task 118 — Request k3s App Onboarding via Platform-Repo PR (child of 117)
- Planned via `smaqit.task-plan` (Mode A): two Explore agents investigated `agents/deployment.md` (confirmed deliberately mode-agnostic — no `provisioning_mode` branching lives there, all of it lives in the calling skills) and the Infrastructure spec template / existing cross-repo `gh` patterns (none existed in this codebase; a new `Platform Repo` Constraints row was the best fit, no new section). Resolved three design forks via `AskUserQuestion`: child of 117 (not standalone), two PRs per app (one per environment/machine-slug), and format-agnostic registry-entry handling.
- Created and started as a child of task 117 (shared branch/worktree/Assisted mode). Authored the new skill's definition + compiled `SKILL.md`, a new `Platform Repo` spec-template row, and a new `secret/apps/<app-slug>/platform-repo` Vault credential distinct from `github`.
- Completed with one AC deliberately left unmet (a live PR-open/merge-detection dry-run against a real scratch repo — no such repo or credential available in this environment), confirmed with the user to complete anyway and record as follow-up.

### Reconciliation with task 116's mid-review correction
- A concurrent session corrected task 116's original design (a specific registry-file/converge-workflow mechanism) to a thin dispatcher, after determining it wrongly hardcoded one downstream project's mechanism as a universal smaqit contract. This invalidated task 118's own design premise, discovered via the auto-memory system after task 118's completion.
- User clarified the resolving principle directly: task 116 must stay generic because it lives in an infra repo smaqit doesn't control; tasks 117/118 can stay declaratively opinionated because they live in the smaqit-managed app repo. Corrected task 118's two skill files (definition + compiled) to stop citing task 116 as documenting a specific mechanism, reframing the PR-request model as smaqit's own declared app-side convention with an explicit compatibility precondition.

### Task 117 Completion — PR #88, v3.5.0
- Ran `smaqit.release-analysis` in Task mode; two rebases were needed before the version/severity computation was trustworthy: (1) the task branch predated task 116's PR merge, so its skill-count assertion (29) was stale relative to current `origin/main` (28 already merged) — corrected to 30 after rebasing; (2) a concurrent session pushed another commit between an initial fetch and the actual rebase, caught via `merge-base --is-ancestor` failing and resolved with a second fetch+rebase.
- Opened PR #88 ("Prepare release v3.5.0") covering both tasks' combined work; verified title, pushed the pending `CHANGELOG.md` entry to local `main`, promoted it on the branch after a second rebase to pick up the pending entry, force-pushed with `--force-with-lease`.
- User merged the PR; confirmed via `gh pr view --json state,mergedAt`, merged `origin/main` into local `main` (clean, no conflicts), marked both tasks `Completed`, removed the worktree, force-deleted the local branch (remote preserved as audit trail).
- User ran `smaqit update` locally and asked for confirmation the release landed correctly on the global install (`~/.claude/skills/`, `~/.agents/skills/`) — verified version (`v3.5.0`), all three k3s skills present with correct, non-stale content on both paths, and confirmed `agents/deployment.md` was correctly left untouched (all `provisioning_mode` branching lives in the calling skills, not the agent).

## Problems Solved

- The app side of smaqit's k3s deployment story had no deploy mechanism at all before this session — an app targeting a platform-owned k3s cluster had no deploy skill to route to, and greenfield's no-match synthesis path would have produced exactly the wrong thing (a VM/SSH/rsync skill). Closed by `smaqit.infrastructure-deploy-k3s-app` plus family-aware routing.
- `smaqit.infrastructure-vault-loader`'s namespace convention had no concept of a kubeconfig credential at all, then (after a first fix) had one that couldn't represent an app onboarded onto multiple k3s machines. Closed by the final machine-keyed `secret/apps/<app-slug>/<machine-slug>/kubeconfig` convention.
- No app-side mechanism existed for requesting k3s onboarding without also requiring direct commit/`workflow_dispatch` access to a repo the app doesn't own. Closed by `smaqit.infrastructure-request-k3s-onboarding`'s PR-and-gate model.
- A design-premise mismatch between task 116's corrected scope and task 118's dependent assumptions was caught before it could ship silently wrong, reconciled with a small, precisely-scoped documentation fix rather than a structural rework.
- A stale release-boundary/skill-count assertion (caused by the task branch predating a concurrent task's merge) was caught before it could ship an incorrect test assertion, via two careful rebases rather than a blind PR merge.

## Decisions Made

- Task 118 ships as a child of task 117 (shared branch/worktree/PR) rather than standalone, per explicit user choice.
- Onboarding requests use a PR-to-a-registry-file model with a single merge-state check per invocation (mirroring `smaqit.feature-new`'s existing deploy-PR gate) — never a busy-poll loop, never `workflow_dispatch` access on the app side.
- `machine_slug`, not `environment`, is the primary targeting dimension across both the Vault convention and the onboarding-request skill's inputs, kept consistent after the machine-keyed Vault revision.
- Tasks 117/118 stay declaratively opinionated about the app-side onboarding/deploy contract; task 116 stays a thin dispatcher — resolved as a structural, not incidental, asymmetry: one lives in a repo smaqit owns the conventions for, the other doesn't.
- Both tasks' unmet live-verification acceptance criteria (no k3s cluster or scratch-repo access in this environment) were completed anyway with explicit user confirmation, recorded as Follow-up identified rather than blocking indefinitely.

## Files Modified

| File | Action |
|------|--------|
| `.smaqit/definitions/skills/smaqit.infrastructure-deploy-k3s-app.md`, `skills/smaqit.infrastructure-deploy-k3s-app/{SKILL.md,scripts/*.sh,assets/*.template}` | Task 117: new skill |
| `.smaqit/definitions/skills/smaqit.infrastructure-request-k3s-onboarding.md`, `skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md` | Task 118: new skill |
| `skills/smaqit.infrastructure-vault-loader/{SKILL.md,scripts/load-credentials.sh,scripts/rotate-credential.sh}` | Both tasks: machine-keyed kubeconfig + `platform-repo` credential |
| `skills/smaqit.input-deployment/SKILL.md`, `skills/smaqit.new-greenfield-project/SKILL.md`, `skills/smaqit.feature-new/SKILL.md` | `existing-k3s` provisioning mode, family-aware routing, onboarding-request gating |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md`, `assets/deploy.yml.k3s.template` | `k3s` generation mode |
| `skills/smaqit.infrastructure-repo-config/SKILL.md` | `KUBECONFIG`/`APP_HOST` Environment secrets |
| `templates/specs/infrastructure.template.md` | New `Platform Repo` Constraints row |
| `installer/main_test.go` | Skill count 27→30 across the session |
| `CHANGELOG.md` | v3.5.0 entry |
| `.smaqit/tasks/{117,118}_*.md`, `PLANNING.md` | Full lifecycle for both tasks |

## Next Steps

- Live verification against the real downstream platform (positive/negative paths, private-registry pull, a real GitHub Actions run of the generated `deploy.yml`, task 118's PR-open/merge-detection dry-run against a real scratch repo) is recorded as Follow-up in both task files — not yet run, needed before relying on either skill in production.
- Task 116's dispatcher skill has one stale line in its Scope section ("a separate, app-side deploy mechanism (not yet a smaqit skill)") that now understates reality — flagged to the user as a small, non-urgent doc fix, not yet acted on.
- Task 115 ("Tenant Reconcile") still names a pre-k3s tenancy model the downstream platform has retired — flagged in task 117's own Notes as needing re-scoping before it starts, not acted on this session.

## Session Metrics

- Tasks completed: 2 (117 → v3.5.0, 118 → bookkeeping-only child completion)
- PRs opened/merged: 1 (#88, covering both tasks)
- New skills shipped: 2 (`smaqit.infrastructure-deploy-k3s-app`, `smaqit.infrastructure-request-k3s-onboarding`); total installed skill count 27→30 across this session's work plus the concurrent task 116
- Vault convention revisions: 2 (add kubeconfig; then flat→machine-keyed)
- Rebases required for a clean, correct release: 2 (stale skill-count baseline; a concurrent session's mid-flow commit)
- Cross-task design reconciliation: 1 (task 118 vs. task 116's mid-review correction), resolved with a 2-file, 4-spot documentation fix
