---
status: Not Started
created: "2026-10-17"
---

# existing-k3s: Fixes From the First Live Private-Mesh Deploy

## Description

The first app deployed live through task 122's private-mesh / platform-owned path (a Python/Tornado +
SQLite app onto a mesh-only, SSO-gated k3s machine, on smaqit v3.14.0) took three fix PRs after the
generated pipeline was merged, and needed project-side workarounds for three more gaps. Every
problem below was hit on a real run, not found by review:

1. **The post-merge dispatcher never dispatches.**
   `skills/smaqit.infrastructure-cicd-generate/assets/post-merge-deploy.yml.template:20` runs
   `gh workflow run deploy.yml --ref …` in a job with no checkout and no `--repo`, so `gh` fails with
   `failed to run git: fatal: not a git repository`. A merged PR carrying `smaqit:deploy` therefore
   never deploys. The base ref is also interpolated straight into the script.
2. **Secrets interpolated into scripts or passed on argv.**
   - `deploy.yml.k3s.template:126` writes `printf '%s' "${{ secrets.KUBECONFIG }}" > "$HOME/.kube/config"`.
     The secret is pasted into the rendered script text, the file is world-readable by default, and it
     is never removed.
   - The commented imagePullSecret step (`:151`) passes `--docker-password="${{ secrets.REGISTRY_TOKEN }}"`
     on kubectl's argv, where it is visible in the runner's process list.
3. **The `kubectl auth can-i <verb> pods/exec` trap.** kubectl parses its argument as
   `[TYPE | TYPE/NAME | NONRESOURCEURL]`, so `can-i create pods/exec` asks "can I create a pod named
   exec?", and any deployer allowed to create pods gets `yes`. A downstream platform spec had
   recorded this as a misleading `can-i` quirk. Nothing in smaqit asserts that the deployer cannot
   exec into pods. The skills only say "never `pods/exec`" (`deploy-k3s-app/SKILL.md:213,252`,
   `deploy-verify/SKILL.md:110`), so projects write their own check and hit the trap.
4. **`bash -e` ends checks that expect a non-zero exit.** GitHub runs every `run:` step with
   `bash -e`. A project-added check that captures a command expected to exit non-zero ends the step
   on the correct answer before evaluating it. Examples: `ANSWER="$(kubectl auth can-i …)"` (exit 1
   means "no"), or a `kubectl exec` that must be refused. The template and skills say nothing about
   this, and the first deploy failed on it twice.
5. **No usable probe client for Python-slim images.** `installer/infra_spec.go:24` accepts only
   `curl` or `wget` as the in-cluster Probe Client, and the probe Job runs the app's own image.
   `python:*-slim` images contain neither, so the project had to add an OS package to its production
   image just for verification.
6. **The credential fillers are unusable under a scripts-only Vault policy.** `vault-loader`,
   `repo-config` (`sync-secrets.sh`, `sync-k3s-env.sh`, `sync-machine-env.sh`), the PAT loading in
   `request-k3s-onboarding`, and `deploy-verify`'s in-cluster mode (which needs a local kubeconfig
   materialized from Vault) all call Vault directly. A project whose operator allows Vault access only
   through its own reviewed scripts has to bind `access-credentials`, `target-credentials` and
   `environment-config` to `none` and perform those slots by hand. `deploy-verify` has to be replaced
   by the deploy run's own verification steps, and the onboarding PR by plain `gh` auth. App runtime
   secrets (a session-signing key, a first-admin password) have no filler at all.

Also observed (smaller):
- the k3s template has no `concurrency` group, so two dispatches can roll out at once;
- one-off in-Namespace Jobs (admin bootstrap) have no template, so every project hand-writes the
  restricted securityContext, `imagePullSecrets`, Secret-sourced arguments and cleanup.

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub Actions, GitHub CLI, Kubernetes (k3s), kubectl, HashiCorp Vault, Docker
**Platforms/Environments:** None
**Features/Integrations:** k3s CI/CD template, post-merge dispatcher, namespace guard, in-cluster probe Job, existing-k3s credential slots, repo-config, vault-loader, request-k3s-onboarding, deploy-verify
**Versions/Constraints:** breaking change allowed (no legacy rendering path); stays app- and platform-agnostic (no platform names in canonical text)

## Design Decisions

- **Breaking, no opt-in** (matching task 122): regenerated workflows change shape, and the golden
  files are updated deliberately rather than kept byte-identical.
- **Dispatcher:** `gh workflow run deploy.yml --repo "${{ github.repository }}" --ref "$BASE_REF"`,
  with the base ref passed through `env:`.
- **No secret in rendered script text or on argv:**
  - every secret reaches a step through `env:`;
  - the kubeconfig is written with `umask 077` under `$RUNNER_TEMP` and shredded in an `always()` step;
  - the pull Secret is built as `.dockerconfigjson` by `jq` reading `env.*`, passed to
    `kubectl create secret generic --type=kubernetes.io/dockerconfigjson --from-file=…=<(…)`.
- **Exec check shipped by the template:** before any apply, `kubectl auth can-i create pods --subresource=exec`
  and `can-i get pods --subresource=exec` must each answer exactly `no`. After the rollout, a real
  `kubectl exec` into the Deployment must be refused as Forbidden. Both steps start with `set +e` and
  decide from the captured output.
- **Document the two traps** (`can-i` TYPE/NAME, `bash -e`) in the cicd-generate and deploy-k3s-app
  skills, where projects add their own steps.
- **Probe client `python3`:** the probe Job runs a stdlib `urllib` one-liner against the Service's
  health path. `infra_spec.go` accepts `python3` next to `curl`/`wget`.
- **Credential slots:** design open, to settle in `task.plan`. Options:
  - (a) fillers delegate every Vault read/write to a project-declared operator command, through a
    binding param;
  - (b) fillers emit an exact operator handoff (paths, secret names, Environment), never calling
    Vault, when the project declares a scripts-only policy;
  - (c) keep direct Vault and document `none` bindings.

  Whichever is chosen, `environment-config` should cover app runtime secrets declared in the spec,
  not only platform credentials. `deploy-verify`'s in-cluster mode should be satisfiable from the
  deploy run itself, without a local kubeconfig.
- **Template `concurrency`** group per target machine (`cancel-in-progress: false`).

## Implementation Steps

1. `post-merge-deploy.yml.template`: `--repo`, base ref via `env:`.
2. `deploy.yml.k3s.template`:
   - kubeconfig via `env:`, `umask 077` and shredded;
   - the dockerconfigjson pull Secret through a file descriptor (turned on when the spec declares a
     private registry);
   - the two exec steps with `set +e`;
   - a concurrency group;
   - golden files regenerated.
3. `installer/infra_spec.go` plus the probe Job template and deploy-verify: a `python3` probe client.
4. Skill text (cicd-generate, deploy-k3s-app): the `can-i --subresource` and `bash -e` gotchas.
5. Credential-slot design (open decision above), then implement it across vault-loader, repo-config,
   request-k3s-onboarding and deploy-verify.
6. Optional: a one-off Job template (restricted, `imagePullSecrets`, Secret-sourced args, cleanup and
   evidence), e.g. for admin bootstrap.
7. Tests: golden renders, a step-execution test under `bash -e` with a stub kubectl (`no`/`yes`,
   Forbidden/success, timeouts), and a dispatcher test without a checkout.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] A merged PR carrying `smaqit:deploy` dispatches `deploy.yml` from a job with no checkout
- [ ] No rendered workflow contains a `${{ secrets.* }}` expression inside a `run:` script or passes a secret on a command line; the kubeconfig is owner-only and removed whatever the outcome
- [ ] The rendered deploy workflow fails when the deployer can exec into pods, and passes on this platform shape (`can-i --subresource=exec` answering `no` and a real exec refused as Forbidden), proven by a stub-kubectl test under `bash -e`
- [ ] An app image with only `python3` (no curl or wget) verifies in-cluster with Probe Client `python3`
- [ ] The credential-slot design is decided, and a project with a scripts-only Vault policy completes the existing-k3s slots without binding them to `none` or hand-performing them, or the chosen design's handoff is exact and tested
- [ ] The cicd-generate and deploy-k3s-app skills document the `can-i` TYPE/NAME and `bash -e` traps
- [ ] Golden files regenerated deliberately; no platform-specific names in canonical skill text

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
| skills/smaqit.infrastructure-cicd-generate/assets/post-merge-deploy.yml.template | Modify |
| skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template, scripts/render-k3s-deploy.sh, SKILL.md | Modify |
| skills/smaqit.infrastructure-deploy-k3s-app/** (probe Job template, SKILL.md) | Modify |
| skills/smaqit.infrastructure-deploy-verify/** | Modify |
| skills/smaqit.infrastructure-repo-config/**, skills/smaqit.infrastructure-vault-loader/**, skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md | Modify |
| installer/infra_spec.go, installer tests and golden files | Modify |
| CHANGELOG.md | Modify |

## Notes

- Source: the first live private-mesh / platform-owned deploy (a downstream app's onboarding task, 2026-10-16/17), smaqit v3.14.0. Failed runs and their fixes: the dispatcher (`not a git repository`), then `can-i create pods/exec` → `yes`, then `bash -e` ending the corrected check. Each fixed downstream by hand before the deploy succeeded.
- Builds on task 122 (shipped in v3.14.0); same breaking-change policy.
- The downstream workaround for gap 6 was a scripts-only Vault helper with Environment syncs (`--sync-only`, generated values). Its interface is a reference point for option (a), not a dependency.
