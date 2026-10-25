# Credential-Free k3s Deploys

## Metadata

- **Date:** 2026-10-23 – 2026-10-25
- **Session focus:**
  - land "every existing-k3s deploy goes through CI/CD";
  - make existing-k3s apps deploy with the CI job's own identity, carrying no deploy credential;
  - turn the existing-k3s slots into contracts that a hosting platform can fill with its own skills.
- **Tasks completed:** 144 (v3.20.0, PR #112), 145 (v3.21.0, PR #113).
- **Tasks created:** 146 (VM modes: CI-only deploys, split from 144).
- **Tasks rescoped:** 145 (OIDC replaces the kubeconfig), 137 (follows 145; scripts-only Vault
  handoff and runtime secrets only).
- **Releases:** v3.20.0 (#112), v3.21.0 (#113).

## Actions Taken

**Session start and fixes**
- Refreshed the Bench compendium entry. Removed downstream project names from task files, as
  CONTRIBUTING requires.
- Rescoped task 145 through three of your decisions:
  - GitHub Actions OIDC replaces the issued kubeconfig outright;
  - one shared mesh key per environment (machine slug) replaces the per-app key, never one key for
    the whole platform;
  - the platform pulls the images, so the app's registry pull credential goes.

**Task 144: all deploys go through CI/CD**
- Implemented the `ci-bootstrap` path for existing-k3s. Its manifests are rendered by cicd-generate
  (your choice), and local deploys are refused.
- Released v3.20.0 (PR #112). Split the VM-mode half out as task 146.

**Task 145: CI-identity deploys without carried credentials**
- **Plan** (`task.plan 145`). Three discovery agents mapped the platform contract, the credential
  plumbing and the workflow generator. You decided:
  - the connect step is the platform's own pinned action, named by new `Cluster Connect` rows;
  - the mesh join runs whenever the row names a mesh secret, so a `public` row may name one;
  - the skill definitions are resynced first.
- **Start:** waited until the platform's identity-based deploy work was finished.
- **Definitions:**
  - Resynced five skill definitions to the compiled skills, then made every change in the
    definition first.
  - They were applied by hand, not compiled by `smaqit.L2` (your choice). L2's skill rules forbid
    the "Use when" descriptions, drop `smaqit-slots`, and impose fixed sections.
- **OIDC build:**
  - `deploy.yml` grants `id-token: write` and `contents: read`. It connects through the spec's
    action, joins the mesh when the row names a secret, and shreds `$KUBECONFIG`.
  - Removed the kubeconfig, per-app mesh-key and registry-credential paths, `sync-k3s-env.sh`, the
    pull-secret step and the `Container Registry` row.
  - existing-k3s `target-credentials` and `environment-config` default to none.
  - The namespace guard retries an unreachable API for a bounded window, and gives Unauthorized and
    Forbidden their own immediate failures.
  - request-k3s-onboarding prints and checks the repository's name and numeric ID.
- **Reframe ("abstract base, concrete platform fillers").** It followed the platform session's
  proposal. You chose:
  - slot contracts live in the resolver, printed by `smaqit slots contract`, with the wiki catalogue
    generated from them and test-guarded;
  - under a CI-only mode, `app-deploy` means start the run, select it, watch it and hand its run ID
    to `verify`;
  - an `Allowed Ingress Modes` row, plus a ClusterIP-only Service lint;
  - the namespace guard stays the reconcile check.
- Sent the existing-k3s contract notes to the platform session.
- **Completion:**
  - Wrote the Findings and fixed one stale definition sentence about the pull-secret step.
  - Opened PR #113 as v3.21.0 and you merged it; the post-merge release published v3.21.0.
  - Removed the worktree and the local branch, and rebuilt the workspace.

## Problems Solved

- **L2 conflict.** The compiler's rules contradict this repo's skill conventions. Applying the
  definitions by hand kept them authoritative without losing those conventions.
- **CI-executed `app-deploy`.** Under a CI-only mode the slot looked like it had nothing to fill. It
  was redefined as start-and-watch, which is the part the agent actually performs.
- **Platform's port wait not finished yet.** Until it is, the guard's bounded API retry covers a mesh
  that resolves the API's name before its port answers.
- **Telling 401 from 403.** An identity the platform rejected (401) and a binding not yet in place
  (403) now get separate messages, so an operator knows which side to fix.
- **Smaller fixes:**
  - a stale SendMessage address was recovered with ListAgents;
  - a misnamed smoke-test variable was fixed;
  - a load test that matched the worktree path was scoped to the script's own output.

## Decisions Made

- OIDC replaces the kubeconfig model with no auth-mode switch, following your no-backwards-compatibility
  principle.
- One mesh key per environment, held by the platform; the app neither stores nor syncs it.
- The platform's connect action owns the API address, the cluster CA, the audience and token refresh.
  smaqit adds no fields for them and never handles a token.
- Slot contracts are code (`installer/slots.go`), not prose only, so a platform's skills can fill them
  with nothing else to change.
- `tls`, `replace` and `teardown` stay refused for existing-k3s bindings.

## Files Modified

- `installer/slots.go`, `installer/main.go`: slot contracts, the `slots contract` command, and the
  k3s defaults.
- `installer/infra_spec.go`: the `Cluster Connect` and `Allowed Ingress Modes` rows, mesh
  decoupling, and the `Container Registry` row removed.
- Installer tests, plus the `installer/testdata/k3s-deploy/*` golden files (two new,
  `private-registry` removed).
- Skills:
  - cicd-generate: template and renderer;
  - deploy-k3s-app: the trigger-and-monitor operation, namespace guard and manifest lint;
  - vault-loader 4.0.0, repo-config 3.0.0 (`sync-k3s-env.sh` deleted), request-k3s-onboarding 2.0.0;
  - image-build-push-static, project-new, input-deployment and feature-new: wording.
- `.smaqit/definitions/skills/`: six definitions (vault-loader new).
- `agents/deployment.md`, `agents/infrastructure.md`, `templates/specs/infrastructure.template.md`.
- `docs/wiki/workflows/deployment-slots.md`, `.smaqit/compendium.md`, `CHANGELOG.md`.
- `scripts/smoke-test-installer.sh`.
- Task files 137, 144, 145 and 146, and `PLANNING.md`.

## Next Steps

- Reword local commit `ee4629d`: its subject still names the downstream platform. Hold the push of
  local `main` until then.
- Create a smaqit-adk task so `smaqit.L2` honours a product's description, metadata and section
  conventions.
- Run an OIDC deploy live against the platform's probe repositories.
- Resync the project-new definition, which is stale against its compiled skill.
- Next tasks in the queue: 137 (runtime secrets under a scripts-only Vault policy), 146 (VM modes
  CI-only), 143 (Bench).

## Session Metrics

- **Duration:** about three days, 2026-10-23 to 2026-10-25.
- **Tasks:**
  - completed: 2 (144, 145);
  - created: 1 (146);
  - rescoped: 2 (145, 137).
- **Releases:** 2 (v3.20.0, v3.21.0).
- **Task 145:**
  - about 50 files changed;
  - 8 of 8 deliberate breaks caught;
  - `make test` and `smoke-test` green.
