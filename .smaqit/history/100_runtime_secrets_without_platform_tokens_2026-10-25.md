# Runtime Secrets Without Platform Tokens

## Metadata

- **Date:** 2026-10-25
- **Session focus:**
  - replan task 137 against the post-145 codebase, then implement and release it;
  - file two framework gaps reported by downstream platform sessions.
- **Tasks completed:** 137 (v3.22.0, PR #114).
- **Tasks created:** 147 (ci-bootstrap re-run state checks), 148 (optional image policy action).
- **Releases:** v3.22.0 (#114).

## Actions Taken

**Session start**
- Loaded project context. Local `main` was already in sync with origin; the previous session's
  "reword a commit before pushing" item had been done.

**Task 137: plan (`task.plan 137`)**
- Found the task file stale against task 145:
  - `sync-k3s-env.sh` and repo-config `sync` for existing-k3s no longer existed;
  - the only remaining Vault contact was the `platform-repo` PAT, loaded on every existing-k3s path,
    including `existing-cicd-pr`, where nothing used it.
- Three discovery agents mapped the spec parser, the fillers that still touched Vault, and the
  runtime-secret render path.
- You decided:
  - retire the PAT for every existing-k3s project (onboarding uses the operator's own `gh` login);
  - runtime secrets go Vault → machine-slug GitHub Environment;
  - feature deploys should touch no Vault.
- Rewrote the task file (description, triage context, decisions, 13 steps, 6 criteria) with your
  confirmation.

**Task 137: start**
- Created the branch and worktree, added the task's research-map block (18 verified URLs), and ran
  issue triage: Advisory. kubectl#1610 (`--from-env-file` can't take multi-line values) changed the
  plan to build the Secret with `jq`.

**Task 137: implementation**
- **Parser:** a `Vault Access` row and a `## Runtime Secrets` table, with each environment's own
  names in `smaqit spec environments --json`.
- **PAT retired:**
  - request-k3s-onboarding 3.0.0 preflights `permissions.push` with the operator's `gh` login;
  - the PAT is gone from `load-credentials.sh`, `credential-file.sh` and `rotate-credential.sh`.
- **Runtime secrets:**
  - vault-loader 5.0.0 reads the spec before touching Vault;
  - repo-config 4.0.0 adds the `runtime` operation (`sync-runtime-env.sh`), with an exact `HANDOFF`
    block and exit 4 under `operator-scripts`;
  - `deploy.yml` gains "Create runtime Secret" (server-side apply, values only through `env:`);
  - manifest-lint gains a `secretKeyRef` rule.
- **Tests:**
  - a stub `smaqit` serving real-parser JSON;
  - tests for zero Vault calls, the handoff text, the sync, the render, step execution and lint;
  - a new test that keeps skill definitions equal to their shipped skills;
  - six deliberate breaks, all caught;
  - `make test` and `make smoke-test` green.

**Task 137: completion**
- Wrote the Findings and computed v3.22.0 (MINOR with action-required notes, per the v3.20/v3.21
  convention).
- The first push got a 403. With your approval, switched to the `apps/smaqit` token, opened PR #114,
  pushed the changelog section, and switched back.
- After you merged it: merged `origin/main` locally, completed the task, removed the worktree and
  the local branch. The v3.22.0 GitHub Release was published.

**Peer requests**
- A downstream platform session reported that `ci-bootstrap` re-runs onboarding and artifact-build on
  an already-onboarded app. Filed as task 147 with your approval (state checks, not a new path).
- A second downstream session asked for an optional image-policy action around the image push. Filed
  as task 148, after task 137.
- Replied to both sessions with their task numbers.

**Sync**
- Committed the research map and pushed local `main` (6 bookkeeping commits plus the map) with the
  same token switch, leaving local and origin equal.

## Problems Solved

- **Zero Vault calls when nothing applies.** `load-credentials.sh` started and unsealed Vault before
  branching on mode. The existing-k3s branch now reads the spec first and exits before Step 0.
- **One `deployment.yaml` for every environment.** A secret declared for only some environments would
  break pods elsewhere. Such a reference now sets `optional: true`, and manifest-lint enforces it per
  machine.
- **Secrets in annotations.** Client-side `kubectl apply` would copy values into
  `last-applied-configuration`. The step builds `data` with `jq` from the environment and applies it
  server-side.
- **Golden-file churn.** `runtimeSecrets` is added to `K3S_TARGETS` only when non-empty, so every
  existing golden stayed byte-identical. That also matters to task 148.
- **Smaller fixes:**
  - a repo-config description over the 400-character limit;
  - a duplicate test helper;
  - a no-op leak check in a refusal loop.

## Decisions Made

- The PAT is retired for every existing-k3s project, not only under `operator-scripts`. The PAT's
  least-privilege scoping is given up: the operator needs collaborator push on the platform
  repository.
- `access-credentials` stays vault-loader `access` on every path, but makes no Vault call without
  runtime secrets. That keeps your "no Vault on feature deploys" outcome and still lets a feature add a
  secret.
- Under `operator-scripts`, the single handoff sits at `environment-config`.
- One Name per secret serves as the Environment secret, the Secret key and the pod's env var.
- Task 147 uses state checks rather than a re-bootstrap path. Task 148 waits for 137.

## Files Modified

- `installer/infra_spec.go`, `installer/slots.go`, and installer tests: `infra_spec_test.go`,
  `skill_scripts_test.go`, `k3s_render_test.go`, `k3s_scripts_test.go`, `skill_text_test.go`,
  `slots_test.go`; plus a new golden file, `installer/testdata/k3s-deploy/runtime-secrets.deploy.yml`.
- Skills:
  - request-k3s-onboarding;
  - vault-loader: SKILL.md, `load-credentials.sh`, `credential-file.sh`, `rotate-credential.sh`;
  - repo-config: SKILL.md and the new `sync-runtime-env.sh`;
  - cicd-generate: SKILL.md, template and renderer;
  - deploy-k3s-app: SKILL.md, deployment template and `manifest-lint.sh`;
  - image-build-push-static, project-new, feature-new and input-deployment: wording.
- `.smaqit/definitions/skills/` (five definitions), `templates/specs/infrastructure.template.md`,
  `agents/infrastructure.md`.
- `docs/wiki/workflows/deployment-slots.md`, `.smaqit/compendium.md`, `README.md`, `CHANGELOG.md`,
  `scripts/smoke-test-installer.sh`.
- Task files 137, 147 and 148, `PLANNING.md`, and `.smaqit/references/project-research.md`.

## Next Steps

- Task 147 (Medium): ci-bootstrap state checks. Its peer platform is waiting on it.
- Task 148 (Medium): optional image policy action, now unblocked by v3.22.0.
- Tasks 146 (VM modes CI-only) and 143 (Bench).
- Follow-ups from 137:
  - `rotate-credential.sh` cannot rotate runtime secrets;
  - the platform's deployer Role must allow `secrets` in the Namespace;
  - the project-new definition is still stale.
- Earlier follow-ups:
  - a smaqit-adk task so `smaqit.L2` honours product conventions;
  - a live OIDC deploy run.

## Session Metrics

- **Duration:** one day, 2026-10-25.
- **Tasks:** completed 1 (137); created 2 (147, 148); rescoped 1 (137).
- **Releases:** 1 (v3.22.0).
- **Task 137:**
  - 38 files changed, about 1,900 lines added;
  - 6 of 6 deliberate breaks caught;
  - `make test` and `make smoke-test` green;
  - every existing golden byte-identical.
