---
status: PR Open
pr: 105
created: "2026-10-17"
mode: Assisted
started: "2026-10-18"
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
   **Moved to task 137** (planned 2026-10-18), with its design decisions; out of scope here.

Also observed (smaller):
- the k3s template has no `concurrency` group, so two dispatches can roll out at once;
- one-off in-Namespace Jobs (admin bootstrap) have no template, so every project hand-writes the
  restricted securityContext, `imagePullSecrets`, Secret-sourced arguments and cleanup. A follow-up,
  out of scope here.

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub Actions, GitHub CLI, Kubernetes (k3s), kubectl, Docker, jq
**Platforms/Environments:** None
**Features/Integrations:** k3s CI/CD template, post-merge dispatcher, VM deploy templates, namespace guard, exec-refused check, in-cluster probe Job (python3), Container Registry pull Secret, workflow concurrency
**Versions/Constraints:** breaking change allowed (no legacy rendering path); stays app- and platform-agnostic (no platform names in canonical text); secrets reach steps only through `env:`

## Design Decisions

Resolved by `task.plan` on 2026-10-18.

- **Breaking, no opt-in** (matching task 122): regenerated workflows change shape, and the golden
  files are updated deliberately rather than kept byte-identical. Released as MINOR with an
  action-required CHANGELOG note.
- **Dispatcher, every mode:** `post-merge-deploy.yml.template` is copied verbatim into `full`,
  `deploy-only` and `k3s`. Its step becomes `gh workflow run deploy.yml --repo "$GH_REPO" --ref "$BASE_REF"`,
  with both values passed through `env:`.
- **No secret in rendered script text or on argv, in every template:**
  - every secret reaches a step through `env:`;
  - the k3s kubeconfig is written with `umask 077` to `$RUNNER_TEMP/kubeconfig` and shredded by a
    final `if: always()` step;
  - the VM templates' `VM_SSH_KEY` gets the same treatment (`full` and `deploy-only`);
  - the pull Secret is built as `.dockerconfigjson` by `jq` reading `env.*`, then passed to
    `kubectl create secret generic --type=kubernetes.io/dockerconfigjson --from-file=.dockerconfigjson=<(…)`.
- **Private registry from the spec:** the existing `Container Registry` Constraints row
  (`<host>, public|private`) is parsed into `smaqit spec environments --json`. It is optional;
  absent means a public image. A malformed row fails with `SLOTS-SPEC-FIELD-INVALID`.
  `render-k3s-deploy.sh` keeps a `# smaqit:pull-secret:begin/end` block only for a private
  registry, mirroring the mesh-join block.
- **Exec checks ship as vendored scripts**, so local deploys get them too and the stub harness tests them:
  - before any apply, `namespace-guard.sh --deployer-exec denied|allowed` checks
    `kubectl auth can-i create pods --subresource=exec` and `can-i get pods --subresource=exec`;
    each must answer exactly `no`;
  - after the rollout, a new `exec-refused.sh` attempts a real `kubectl exec deploy/<slug> -- true`
    with a timeout. It passes only on a non-zero exit whose output contains `Forbidden`;
  - both bracket their captured commands with `set +e` and decide from the output.
- **Spec-controlled exec opt-out:** a `Deployer Exec` Constraints row (`denied` by default, or
  `allowed`). Under `allowed`, a `yes` or a successful exec is a warning instead of a failure. Errors,
  timeouts, NotFound and unclear answers fail closed under both settings.
- **Document the two traps** (`can-i` TYPE/NAME, `bash -e`) in the cicd-generate and deploy-k3s-app
  skills, where projects add their own steps.
- **Probe client `python3`:** a stdlib `urllib.request.urlopen(sys.argv[1], timeout=5)` one-liner
  (non-zero exit on HTTP ≥400 or a connection error). `infra_spec.go` accepts `python3` next to
  `curl`/`wget`.
- **Concurrency:** on the `deploy` job only, grouped per target machine with
  `cancel-in-progress: false`, so builds still run in parallel.
- **Moved to task 137 (independent task):** credential slots under a scripts-only Vault policy (gap 6).
  Its decisions were settled in this plan:
  - a spec `Vault Access: operator-scripts` row switches fillers to an exact operator handoff
    followed by a Vault-free check of the outcome;
  - a deploy-verify `ci-evidence` operation reads the CI run's verify steps;
  - app runtime secrets get a spec table, an environment-config sync or handoff, and manifest wiring.
- **Out of scope / follow-ups:** a one-off Job template; VM modes deploy twice on a marked merge
  (push trigger plus dispatch).

## Implementation Steps

*Phase A — spec contract (`installer/infra_spec.go`)*
1. Add `python3` to `probeClients` and update the allowed-values text.
2. Parse the optional `Container Registry` (`<host>, public|private`) and `Deployer Exec`
   (`denied|allowed`, default `denied`) Constraints rows into `infraEnvironments` and the `--json`
   output. Validate them in `validateSharedConstraints`. Update `templates/specs/infrastructure.template.md`
   and `agents/infrastructure.md`.

*Phase B — guard scripts (depends on A)*

3. `namespace-guard.sh`: add `--deployer-exec denied|allowed` and the two `can-i --subresource=exec`
   checks, using the existing `set +e` / `RC` / `ANSWER` pattern. Fail closed on anything unclear.
4. New `deploy-k3s-app/scripts/exec-refused.sh`: the post-rollout real-exec check, with the same policy flag.

*Phase C — templates and renderer (depends on A and B)*

5. `post-merge-deploy.yml.template`: `--repo` and `--ref` from `env:`.
6. `deploy.yml.k3s.template`:
   - the kubeconfig through `env:`, `umask 077`, `$RUNNER_TEMP`, cleaned up by an `always()` step;
   - the guard gets `--deployer-exec "$DEPLOYER_EXEC"`, and a new "Exec refused" step runs after
     "Wait for rollout";
   - `deploy` job concurrency per machine;
   - the marker-wrapped dockerconfigjson pull-Secret step.
7. `render-k3s-deploy.sh`:
   - keep or drop the pull-Secret block with a `__REGISTRY_HOST__` token, plus a post-check;
   - render `DEPLOYER_EXEC` into the job `env:`;
   - refuse a private registry with no host.
8. `deploy.yml.full.template` and `deploy.yml.deploy-only.template`: the "Set up SSH key" step takes
   `VM_SSH_KEY` through `env:`, writes with `umask 077`, and is cleaned up by an `always()` step.

*Phase D — probe and skill text (parallel with C)*

9. `probe-job.yaml.template` and the `__PROBE_FETCH__` mapping in deploy-k3s-app SKILL.md Step 5:
   add `python3`, and fix the hardcoded wget example.
10. Skill text:
    - cicd-generate SKILL.md:
      - gotchas for `can-i` TYPE/NAME and `bash -e`;
      - rules for secrets and concurrency;
      - update line 122's list of what the renderer fills, and line 59's description of the secret names.
    - deploy-k3s-app SKILL.md:
      - rewrite the Step 4 pull Secret in dockerconfigjson form;
      - describe the exec checks;
      - add the same two gotchas;
      - point its "never `pods/exec`" lines at the enforced check.
    - deploy-verify SKILL.md: the `python3` client.

*Phase E — tests (depends on C)*

11. `k3s_render_test.go`:
    - `runStep` runs under `bash --noprofile --norc -eo pipefail`;
    - `k3sStep` gets `Env` and `k3sWorkflow` gets `Concurrency`;
    - `runsUnder` handles `always()`.
12. Run the guard, "Exec refused" and kubeconfig steps with the stub `kubectl`. This needs new stub
    modes in `skill_scripts_test.go`: exec `yes`, Forbidden, exec succeeds, NotFound, timeout. Cover
    pass, fail and the `allowed` warning, plus the kubeconfig's mode 0600 and its removal.
13. Dispatcher test: run the rendered step outside a git repository, with a stub `gh` that refuses to
    run without `--repo`.
14. A contract test across every `skills/*/assets/*.template` file: no `${{ secrets.` or
    `${{ github.event.` inside a `run:`, and no `--docker-password` / `--password` arguments.
15. Probe test: the `python3` fetch command run under `sh` against an `httptest` server, with only
    `python3` on `PATH`; cover both pass and fail. Skip if `python3` is missing.
16. Golden files and contract tests:
    - add a `private-registry` golden case;
    - regenerate every golden with `-update` and review the diff;
    - update the rendered-contract step order;
    - add renderer-refusal cases and infra_spec parse/refusal cases.
17. Update `docs/wiki/workflows/deployment-slots.md` and `.smaqit/compendium.md`, plus a CHANGELOG
    entry with the action-required note.

## Known Issues Triage
**Triaged:** 2026-10-18
**Tools searched:** GitHub CLI (`cli/cli`), GitHub Actions (`actions/runner`), kubectl (`kubernetes/kubectl`), Kubernetes (k3s) (`k3s-io/k3s`), jq (`jqlang/jq`)
**Result:** Advisory

### Advisory Issues
- [#1217 auth/can-i: check subresource if it exists and belongs to resource](https://github.com/kubernetes/kubectl/issues/1217) — `kubernetes/kubectl` — opened 2022-05-23 — kind/bug, lifecycle/rotten, needs-triage. `can-i --subresource` does not validate that the subresource exists. The exec check names the real `exec` subresource, so it is unaffected, but a mistyped subresource would answer silently. Tests should pin the exact `--subresource=exec` arguments.
- [#643 add-mask doesn't work with workflow_dispatch inputs](https://github.com/actions/runner/issues/643) — `actions/runner` — opened 2020-08-05 — enhancement, Runner Feature. Dispatch inputs are not masked. The k3s `environment` input is a machine slug, not a secret, so this is no concern provided no secret is ever passed as a dispatch input.

### Historical (Closed)
- [#1797 `kubectl auth can-i patch deployments/scale` says yes despite lacking permission to patch deployments/scale](https://github.com/kubernetes/kubectl/issues/1797) — `kubernetes/kubectl` — closed 2025-11-11. This is the same TYPE/NAME parsing trap as gap 3: `deployments/scale` is read as a deployment named `scale`. It confirms the planned `--subresource` form and the gotcha text.
- [#3028 Cannot write in $RUNNER_TEMP](https://github.com/actions/runner/issues/3028) — `actions/runner` — closed 2024-02-16. Fixed; `$RUNNER_TEMP` is writable on hosted runners.
- [#2498 Inconsistent value for `runner.temp` when running in a container](https://github.com/actions/runner/issues/2498) — `actions/runner` — closed 2023-03-22. Relevant only to container jobs, which the k3s deploy job does not use.
- [#14584 Subprocess environment inheritance leaks GH_TOKEN/GITHUB_TOKEN to local Git hooks and filters](https://github.com/cli/cli/issues/14584) — `cli/cli` — closed 2026-10-05. The dispatcher step runs with no checkout, so no Git hooks exist.

### Omitted Tools
- Docker — excluded after relevance prioritization and the five-repository limit (only the `.dockerconfigjson` file format is used)

## Acceptance Criteria

- [x] A merged PR carrying `smaqit:deploy` dispatches `deploy.yml` from a job with no checkout, in every CI mode
- [x] No rendered workflow (k3s, `full`, `deploy-only`, post-merge) contains a `${{ secrets.* }}` expression inside a `run:` script or passes a secret on a command line, enforced by a test across every template; the kubeconfig and the VM SSH key are owner-only and removed whatever the outcome
- [x] The rendered deploy workflow fails when the deployer can exec into pods, and passes on this platform shape (`can-i --subresource=exec` answering `no` and a real exec refused as Forbidden), proven by a stub-kubectl test under `bash -e`
- [x] `Deployer Exec: allowed` turns an exec `yes` or a successful exec into a warning; errors, timeouts and unclear answers fail closed under both settings
- [x] A private `Container Registry` renders a pull-Secret step that passes no secret on argv; a public or absent one renders none
- [x] The `deploy` job has a per-machine concurrency group with `cancel-in-progress: false`
- [x] An app image with only `python3` (no curl or wget) verifies in-cluster with Probe Client `python3`
- [x] The cicd-generate and deploy-k3s-app skills document the `can-i` TYPE/NAME and `bash -e` traps
- [x] Golden files regenerated deliberately; no platform-specific names in canonical skill text

## Findings

**Implementation approach:**
- Spec contract first: `infra_spec.go` gained the `python3` Probe Client and two optional Constraints rows, `Container Registry` (`<host>, public|private`) and `Deployer Exec` (`denied` default, or `allowed`). Both are exposed in `smaqit spec environments --json`.
- Exec checks went into vendored scripts rather than inline workflow steps, so local deploys get them too and the stub harness tests them: `namespace-guard.sh --deployer-exec` asks `can-i create|get pods --subresource=exec`, and the new `exec-refused.sh` attempts a real exec after the rollout.
- Templates:
  - every secret reaches its step through `env:`;
  - the kubeconfig and VM SSH key are written under `umask 077` and removed by an `always()` step;
  - the post-merge dispatcher passes `--repo`;
  - the k3s `deploy` job gained per-machine concurrency;
  - a marker-wrapped dockerconfigjson pull-Secret step is kept by the renderer only for a private registry.
- Tests run rendered steps the way GitHub does (`bash --noprofile --norc -eo pipefail`, step `env:` applied) against stub `kubectl`/`gh`. A contract test covers every workflow template. The `python3` probe runs with only `python3` on `PATH`. Seven deliberate breaks were each caught.

**Decisions made:**
- Gap 6 (credential slots under a scripts-only Vault policy) moved to the independent task 137, together with its settled design: a spec handoff-and-verify row, deploy-verify `ci-evidence`, and runtime secrets.
- The exec check fails by default; `Deployer Exec: allowed` downgrades a permitted exec to a warning, while errors, timeouts and NotFound fail closed under both settings. An exec that reaches the container but whose command fails (for example on a distroless image) counts as permitted.
- The secrets and dispatcher fixes apply to every CI mode, since the post-merge template is shared and both VM templates interpolated `VM_SSH_KEY`. The PR number in `provision.yml` also moved to `env:` so the rule has no exceptions.
- `namespace-guard.sh` usage errors now exit 2, matching the other guard scripts; no caller depended on the old exit 1.
- `deploy-verify` needed no change: it runs the committed probe Job, which carries the fetch command.
- The renderer now drops the blank line after a removed block, which also removes a double blank line that all-public renders had before this task.

**Blockers encountered:**
- None.

**Follow-up identified:**
- Task 137: credential slots under a scripts-only Vault policy, app runtime secrets, and verify from the CI run.
- A one-off in-Namespace Job template (admin bootstrap) is still missing.
- VM `full`/`deploy-only` workflows still deploy twice on a marked merge (a push trigger plus the dispatch).
- The first live merge carrying `smaqit:deploy` after this release is the real proof of the dispatcher fix; the test uses a stub `gh`.
- Task 135 (PR #104) also edits the cicd-generate skill and the CHANGELOG, so whichever task merges second rebases.

## Files to Create / Modify

| File | Action |
|------|--------|
| installer/infra_spec.go, installer/infra_spec_test.go | Modify |
| templates/specs/infrastructure.template.md, agents/infrastructure.md | Modify |
| skills/smaqit.infrastructure-deploy-k3s-app/scripts/namespace-guard.sh | Modify |
| skills/smaqit.infrastructure-deploy-k3s-app/scripts/exec-refused.sh | Create |
| skills/smaqit.infrastructure-deploy-k3s-app/assets/probe-job.yaml.template, SKILL.md | Modify |
| skills/smaqit.infrastructure-cicd-generate/assets/post-merge-deploy.yml.template, deploy.yml.k3s.template, deploy.yml.full.template, deploy.yml.deploy-only.template | Modify |
| skills/smaqit.infrastructure-cicd-generate/scripts/render-k3s-deploy.sh, SKILL.md | Modify |
| skills/smaqit.infrastructure-deploy-verify/SKILL.md | Modify |
| installer/k3s_render_test.go, installer/k3s_scripts_test.go, installer/skill_scripts_test.go, installer/cicd_templates_test.go | Modify |
| installer/testdata/k3s-deploy/* (regenerated, plus a private-registry case) | Modify / Create |
| docs/wiki/workflows/deployment-slots.md, .smaqit/compendium.md, CHANGELOG.md | Modify |

## Notes

- Source: the first live private-mesh / platform-owned deploy (a downstream app's onboarding task, 2026-10-16/17), smaqit v3.14.0. Failed runs and their fixes: the dispatcher (`not a git repository`), then `can-i create pods/exec` → `yes`, then `bash -e` ending the corrected check. Each fixed downstream by hand before the deploy succeeded.
- Builds on task 122 (shipped in v3.14.0); same breaking-change policy.
- The downstream workaround for gap 6 was a scripts-only Vault helper with Environment syncs (`--sync-only`, generated values). Its interface is a reference point for option (a), not a dependency. Gap 6 now belongs to task 137.
- Task 135 (running in another session) also edits cicd-generate SKILL.md and CHANGELOG.md; whichever
  merges second rebases.
