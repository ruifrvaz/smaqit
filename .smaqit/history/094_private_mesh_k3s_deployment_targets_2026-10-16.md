# Private-Mesh k3s Deployment Targets

## Metadata

- **Date:** 2026-10-16
- **Session focus:** Plan and ship support for `existing-k3s` targets that aren't publicly reachable (private mesh only) or whose Ingress the platform owns (for example behind its SSO gate)
- **Tasks completed:** 122 (v3.14.0, PR #103)
- **Tasks referenced:** 135 (uncommitted, created by another session), 126, 113, 094, 077, 074, 071

## Actions Taken

- Ran `session.start`: loaded README, history 093, PLANNING.md, compendium, and `plan-guard.sh` for task 135 (the next unblocked task by priority). The user chose to plan task 122 instead.
- **`task.plan 122`:** three discovery agents covered the spec contract, CI/CD rendering, and the k3s skills and slot catalogue. They surfaced two gaps behind the task's recorded decisions:
  - the Infrastructure spec had no per-environment structure, and no Go code read spec bodies;
  - one `deploy.yml` serves several environments, which can differ in class.
  - Four Q&A decisions resolved them: a parsed `## Environments` table; the gate in `smaqit slots plan`; run-time branching on a rendered per-machine map; and a probe-Job verdict from Job status, running the app's own image. Wrote the plan into the task file (Design Decisions, Implementation Steps, three new ACs, Files table).
- **`task.start 122`:**
  - Created the branch and worktree from `origin/main`.
  - Built the task 122 research-map block (25 verified URLs).
  - Issue triage came back Advisory: cert-manager ingress-shim bugs, including #8733, a stale Certificate after a mode change; a Traefik TLS-store leak; and the `actions/runner` searches were rate-limited.
- **Implementation (assisted), in four phases:**
  - **Spec contract and gate:** `installer/infra_spec.go`, the `SLOTS-SPEC-*` gate in `resolveSlotPlan`, `smaqit spec environments`, checks in `smaqit validate`, the template table, and the agent and input-deployment text.
  - **CI rendering:** a dispatch-only k3s `deploy.yml` template, `render-k3s-deploy.sh`, seven golden files, and contract, refusal, and step-execution tests.
  - **k3s skills:** default-TLS Ingress and probe Job templates; `mesh-check` (the existing-k3s `platform-prereqs` default); deploy-verify's in-cluster mode; a fail-closed `namespace-guard.sh` with a stub `kubectl`.
  - **Credentials and orchestration:** mesh setup-key load, rotate, and sync; greenfield, feature-new, and deployment-agent `ci-push` dispatch; wiki, README, compendium, and CHANGELOG.
- Verified with `make test`, `make smoke-test` (the gate and the new command through the compiled binary), actionlint on every render, and five deliberate breaks, each caught by a test.
- **`task.complete 122`:**
  - Phase 1: wrote Findings and checked all 11 ACs; release-analysis suggested v3.14.0, MINOR; opened PR #103 and committed the changelog on the branch.
  - Phase 2, after the user merged: merged `origin/main`, marked the task Completed, and removed the worktree and local branch. v3.14.0 published at 19:43 UTC.

## Problems Solved

- **No machine-readable environments:** a new `## Environments` table, parsed with the same section-slicing as design references, and one spec allowed to declare it (`SLOTS-SPEC-AMBIGUOUS` otherwise).
- **One workflow, many classes:** `uses:` can't be an expression, so the mesh-join step is rendered in (and deleted for all-public apps). Everything else branches at run time on `K3S_TARGETS`.
- **No conditional rendering existed:** one sanctioned renderer (jq + awk, literal token replacement), with golden files regenerated via `-update`.
- **Proving run-time behaviour:** a test runs the rendered "Resolve target" and "Select and stamp" steps per combination and evaluates each step's `if:`.
- **`namespace-guard.sh` failed open:** it now checks reachability first and accepts only an explicit "no" as denied.
- **Dispatch-only knock-on:** the deployment agent's `ci-push` path waited for a push run; it now dispatches when there's no push trigger. Feature-new's decision table already handled it.
- **Caught during implementation:** `getent` exit 2 under `pipefail` aborted `mesh-check`; an unescaped apostrophe broke a single-quoted skill description in `make prepare`; the smoke test planned existing-k3s without a spec; actionlint flagged an unused loop variable.

## Decisions Made

- Breaking with no default, released as MINOR with an action-required CHANGELOG note (the v3.13.0 precedent).
- The default fillers read the spec, not binding `params`; `provisioning_mode` stays closed and no slot was added.
- The revision is proven from the Deployment's image tag for every class. The external check requests only the declared health path and `/`.
- The probe Job runs the app's own image with the spec's `Probe Client`; its verdict is the Job's status (no `pods/log`). It needs Jobs RBAC and same-Namespace NetworkPolicy from the platform.
- `own-tls` keeps `deployment/k3s/ingress.yaml`. Host, Ingress Class, and ClusterIssuer became apply-time tokens.
- The mesh setup key lives at `secret/apps/<app>/<machine>/mesh-setup-key`, is synced under the spec-named secret, and is kept out of `credential-file.sh`.
- The certificate-wait step stays in the shared workflow, skipped by `if:` for non-own-tls machines; AC1 is met at run time.
- Task 135's uncommitted file and PLANNING row were never swept into task 122 commits; PLANNING was staged with only task 122's lines via `git update-index --cacheinfo`.

## Files Modified

- `installer/infra_spec.go` (new), `installer/slots.go`, `installer/main.go` — parser, gate, `spec` command, validate checks, mesh-check default
- `installer/infra_spec_test.go`, `installer/k3s_render_test.go`, `installer/k3s_scripts_test.go` (new); `installer/slots_test.go`, `installer/skill_scripts_test.go` (stub `kubectl`), `installer/cicd_templates_test.go` (old contract test removed); `installer/testdata/k3s-deploy/` (7 golden files)
- `templates/specs/infrastructure.template.md`, `agents/infrastructure.md`, `agents/deployment.md`
- `skills/smaqit.infrastructure-cicd-generate/` — `SKILL.md`, `assets/deploy.yml.k3s.template`, new `scripts/render-k3s-deploy.sh`
- `skills/smaqit.infrastructure-deploy-k3s-app/` — `SKILL.md` (2.0.0), `assets/ingress.yaml.template`, new `assets/ingress-default-tls.yaml.template`, `assets/probe-job.yaml.template`, new `scripts/mesh-check.sh`, `scripts/namespace-guard.sh`
- `skills/smaqit.infrastructure-deploy-verify/SKILL.md`, `skills/smaqit.infrastructure-repo-config/` (`SKILL.md`, `scripts/sync-k3s-env.sh`), `skills/smaqit.infrastructure-vault-loader/` (`SKILL.md`, `scripts/load-credentials.sh`, `scripts/rotate-credential.sh`)
- `skills/smaqit.input-deployment`, `skills/smaqit.new-greenfield-project`, `skills/smaqit.feature-new`, `skills/smaqit.infrastructure-image-build-push-static` — `SKILL.md`
- `scripts/smoke-test-installer.sh`, `docs/wiki/workflows/deployment-slots.md`, `docs/wiki/workflows/platform-repositories.md`, `README.md`, `.smaqit/compendium.md`, `CHANGELOG.md`
- `.smaqit/tasks/122_existing_k3s_private_mesh_only_targets.md`, `.smaqit/tasks/PLANNING.md`, `.smaqit/references/project-research.md`

## Next Steps

- Task 135 (`plan-guard.sh --allow-delete`): its file and PLANNING row are still uncommitted from another session.
- Existing `existing-k3s` projects must add the Environments table and regenerate `deploy.yml` (v3.14.0 action-required note). The first downstream mesh-only, SSO-gated app can now proceed.
- Follow-ups from task 122:
  - sync `.smaqit/definitions/skills/` for deploy-k3s-app, cicd-generate, and deploy-verify;
  - make a first live run against a private-mesh / platform-owned machine;
  - clean up the stale Certificate when an environment leaves `own-tls` (cert-manager#8733).
- Carried over: unseal-key and legacy writes via argv; the installer skill-retirement pruning gap; tasks 126 and 113; smaqit-extensions task 042 is still uncommitted.

## Session Metrics

- Duration: one day (2026-10-16)
- Tasks completed: 1 (122); release: v3.14.0 (PR #103)
- Discovery agents: 3 Explore agents; plan Q&A: 4 decisions
- New files: 13 (4 Go, 7 golden, 2 scripts and 2 templates under skills); 29 modified
- New tests: 10 test functions (parser/gate tables, goldens, rendered contract, renderer refusals, step execution across 12 combinations, namespace guard, mesh-check, template lint, mesh-key load/sync)
- Deliberate-break checks: 5, each caught by a test
