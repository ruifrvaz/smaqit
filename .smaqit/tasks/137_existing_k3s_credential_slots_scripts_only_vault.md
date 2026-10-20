---
status: Not Started
created: "2026-10-18"
---

# existing-k3s: Credential Slots Under a Scripts-Only Vault Policy

## Description

Some operators allow Vault access only through their own reviewed scripts, and the existing-k3s
credential fillers can't run under that policy. Every one of them calls the `vault` binary itself,
with no way to point it elsewhere:

- `smaqit.infrastructure-vault-loader` `access` and `target` run `load-credentials.sh`;
- `smaqit.infrastructure-repo-config` `sync` runs `sync-k3s-env.sh`, which pipes
  `vault kv get` into `gh secret set`;
- `smaqit.infrastructure-request-k3s-onboarding` loads the platform-repo PAT through
  `credential-file.sh`;
- `smaqit.infrastructure-deploy-verify`'s in-cluster mode, on a local (`standard`-path) deploy, needs
  a local kubeconfig, which it also reads from Vault through `credential-file.sh`. CI-path deploys
  no longer need it: task 142 made deploy-verify's `ci-evidence` the CI-path default.

Under that policy, a project has to bind `access-credentials`, `target-credentials` and
`environment-config` to `none` and perform those slots by hand. It also has to replace the onboarding
PR's token with plain `gh` authentication.

App runtime secrets have no home either: a session-signing key or a first-admin password has no
spec field and no filler. No k3s manifest template has a `secretKeyRef` or `envFrom`, so the
generated manifests can't inject one into the pod.

This is gap 6 of task 136, found on the first live private-mesh deploy. It was split out during
`task.plan 136` on 2026-10-18, which also settled the design below.

## Issue Triage Context

**Mode:** Auto
**Technologies:** HashiCorp Vault, GitHub CLI, GitHub Actions, Kubernetes (k3s), kubectl, jq, Bash, Go
**Platforms/Environments:** None
**Features/Integrations:** existing-k3s credential slots, vault-loader, repo-config sync-k3s-env, request-k3s-onboarding, deploy-verify in-cluster mode, runtime secrets, secretKeyRef
**Versions/Constraints:** breaking change allowed (MINOR with action-required note), no raw vault commands in skill text, secrets only via env: or stdin, app- and platform-agnostic

## Design Decisions

- **Policy declared in the spec:** a Constraints row `Vault Access: direct|operator-scripts`
  (default `direct`), parsed by `installer/infra_spec.go` into `smaqit spec environments --json`
  and validated with `SLOTS-SPEC-*` errors. It lives in the spec, like the rest of the existing-k3s
  target data, not in binding `params`.
- **Handoff, then verify:** under `operator-scripts`, each credential filler never calls Vault:
  1. it emits an exact operator handoff: Vault path, field, GitHub Environment, secret name, and the
     shape of the command to run;
  2. it pauses the slot, following the deployment agent's rule to hand the operator the exact
     command (`agents/deployment.md:78`);
  3. it verifies the outcome without Vault, for example by checking that `gh secret list --env <machine>`
     shows the expected names.

  The slot completes; it is not bound to `none`.
- **Onboarding:** under `operator-scripts`, `request-k3s-onboarding` may use the operator's own
  `gh` authentication instead of the Vault-held platform-repo PAT.
- **CI-path verification belongs to task 142:** deploy-verify's `ci-evidence` (the default `verify`
  for existing-k3s on `existing-cicd-pr` and `ci-push`) needs no local kubeconfig, Vault or mesh
  membership. This task covers only the local (`standard`-path) in-cluster verify under
  `operator-scripts`, whose kubeconfig follows the handoff-then-verify rule above.
- **App runtime secrets:**
  - a new `## Runtime Secrets` spec table (name, Kubernetes Secret key, generated or
    operator-provided, environments), parsed by `infra_spec.go`;
  - `environment-config` syncs each secret into the machine's GitHub Environment, or under
    `operator-scripts` hands it off;
  - the rendered `deploy.yml` creates a Kubernetes Secret from `env:`, never on argv;
  - the deployment template references it through `secretKeyRef`.
- **Keep** `installer/skill_text_test.go`'s rule against raw `vault` commands in skill text.

## Implementation Steps

1. `installer/infra_spec.go`:
   - parse and validate the `Vault Access` row and the `## Runtime Secrets` table;
   - expose both in the `--json` output;
   - update `templates/specs/infrastructure.template.md` and `agents/infrastructure.md`.
2. vault-loader `access` and `target`: an `operator-scripts` branch that emits the handoff and then
   verifies without Vault.
3. repo-config `sync` and `sync-k3s-env.sh`:
   - an `operator-scripts` handoff with a `gh`-only verification;
   - syncing or handing off the runtime secrets.
4. request-k3s-onboarding: under `operator-scripts`, use the operator's `gh` authentication.
5. deploy-verify `verify`, `standard` path only: under `operator-scripts`, the in-cluster mode's
   kubeconfig follows the handoff-then-verify rule instead of `credential-file.sh`. (CI-path
   verification is task 142's `ci-evidence`.)
6. Templates:
   - in `deploy.yml.k3s.template` and `render-k3s-deploy.sh`, create the runtime-secret
     Kubernetes Secret from `env:`;
   - in `deployment.yaml.template`, reference it through `secretKeyRef`;
   - regenerate the golden files.
7. Tests:
   - parser refusals;
   - a stub `vault` that proves each filler makes zero `vault` calls under `operator-scripts`;
   - exact handoff text;
   - the runtime-secret render, with no secret in `run:` or on argv.
8. Update `docs/wiki/workflows/deployment-slots.md`, `.smaqit/compendium.md`, and the CHANGELOG
   (with an action-required note), then release.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] The `Vault Access` row is parsed and validated, and a bad value fails with an `SLOTS-SPEC-*` error
- [ ] Under `operator-scripts`, every existing-k3s credential filler makes zero `vault` calls (proven
      with a stub `vault`), emits an exact, tested handoff, and verifies the outcome without Vault;
      a project completes these slots without binding them to `none`
- [ ] A `## Runtime Secrets` table is parsed, synced or handed off into the machine's Environment, and
      wired into the rendered `deploy.yml` and the deployment template, with no secret in `run:` or on argv
- [ ] Golden files regenerated deliberately; no raw `vault` commands in skill text; no platform-specific
      names in canonical skill text
- [ ] Wiki, compendium and CHANGELOG updated; released

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
| installer/infra_spec.go, installer/infra_spec_test.go | Modify |
| templates/specs/infrastructure.template.md, agents/infrastructure.md | Modify |
| skills/smaqit.infrastructure-vault-loader/** | Modify |
| skills/smaqit.infrastructure-repo-config/** | Modify |
| skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md | Modify |
| skills/smaqit.infrastructure-deploy-verify/SKILL.md (`standard`-path in-cluster mode only) | Modify |
| skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template, scripts/render-k3s-deploy.sh | Modify |
| skills/smaqit.infrastructure-deploy-k3s-app/assets/deployment.yaml.template | Modify |
| installer tests and installer/testdata/k3s-deploy/* | Modify |
| docs/wiki/workflows/deployment-slots.md, .smaqit/compendium.md, CHANGELOG.md | Modify |

## Notes

- Depends on task 136, and builds on its `env:`-only secret handling and template changes. Start it
  after 136 merges.
- Task 142 owns CI-path verification (deploy-verify `ci-evidence`, moved out of this task on
  2026-10-20). Start this task after 142 merges.
- A downstream project worked around this with its own scripts-only Vault helper (`--sync-only`,
  generated values). That helper is a reference point, not a dependency.
