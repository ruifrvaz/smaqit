---
status: PR Open
pr: 114
created: "2026-10-18"
mode: Assisted
started: "2026-10-25"
---

# existing-k3s: Credential Slots Under a Scripts-Only Vault Policy

## Description

Some operators allow Vault access only through their own reviewed scripts, and the existing-k3s
credential fillers can't run under that policy: they call the `vault` binary themselves, with no
way to point them elsewhere. App runtime secrets have no home either: a session-signing key or a
first-admin password has no spec field and no filler, and no k3s manifest template has a
`secretKeyRef`, so the generated manifests can't inject one into the pod.

**Rescoped 2026-10-25 (`task.plan 137`), after task 145.** Task 145 removed every deploy credential
(kubeconfig, per-app mesh key, pull credential, `sync-k3s-env.sh`; `environment-config` is `none`).
The only Vault contact left was the `platform-repo` PAT, loaded by vault-loader `access` (on every
existing-k3s path, including `existing-cicd-pr`, where nothing uses it) and read by
request-k3s-onboarding through `credential-file.sh`. This task therefore:

- **retires the platform-repo PAT** for every existing-k3s project: onboarding uses the operator's
  own `gh` login, so an app with no runtime secrets never contacts Vault;
- **adds runtime secrets**: a `## Runtime Secrets` spec table, values held in Vault and synced into
  the machine-slug GitHub Environment, a Kubernetes Secret built by the deploy job from step `env:`,
  consumed through `secretKeyRef`;
- **adds a `Vault Access: direct|operator-scripts` row**: under `operator-scripts` the fillers make
  zero `vault` calls, hand the operator an exact command, and verify the result with `gh` alone.

This is gap 6 of task 136, found on the first live private-mesh deploy, split out during
`task.plan 136` on 2026-10-18 and rescoped on 2026-10-23 (after 144/145's plans) and 2026-10-25.

## Issue Triage Context

**Mode:** Auto
**Technologies:** HashiCorp Vault, GitHub CLI, GitHub Actions, Kubernetes (k3s), kubectl, jq, Bash, Go
**Platforms/Environments:** None
**Features/Integrations:** existing-k3s spec parser (Vault Access, Runtime Secrets), request-k3s-onboarding operator gh auth, vault-loader access, repo-config runtime sync, deploy.yml runtime Secret, secretKeyRef, manifest-lint
**Versions/Constraints:** smaqit v3.21.0, breaking change allowed (MINOR v3.22.0 with action-required note), platform-repo PAT retired with no opt-out, spec JSON schemaVersion stays 1, no raw vault commands in skill text, secrets only via env: or stdin, app- and platform-agnostic

## Design Decisions

- **Platform-repo PAT retired everywhere (user, 2026-10-25):** request-k3s-onboarding authenticates
  with the operator's own `gh` login for every existing-k3s project, not only under
  `operator-scripts`. The PAT's least-privilege scoping is given up; the operator needs collaborator
  push on the platform repository, and the fork-based flow (`-H user:branch`) stays forbidden. An
  existing-k3s app with no runtime secrets makes no Vault contact under either policy.
- **Policy declared in the spec:** a Constraints row `Vault Access: direct|operator-scripts` (default
  `direct`), parsed by `installer/infra_spec.go` into `smaqit spec environments --json` and validated
  with `SLOTS-SPEC-FIELD-INVALID`, following the `Deployer Exec` pattern. Fillers read it from the spec
  JSON; `schemaVersion` stays 1 (additive fields).
- **Runtime secrets: Vault → Environment (user, 2026-10-25):** a `## Runtime Secrets` table (Name,
  Source `generated|operator-provided`, Environments). Under `direct`, vault-loader `access`
  populates `secret/apps/<app-slug>/<machine-slug>/runtime` (one field per Name; `generated` values
  created locally and written on stdin, never printed; `operator-provided` via hidden prompt), and a
  new repo-config `runtime` operation (bound to `environment-config`) pipes each value into the
  machine-slug Environment. Generated values live in Vault so they can be rotated and recovered.
- **One Name per secret:** the same Name is the GitHub Environment secret, the key in the Kubernetes
  Secret `<app-slug>-runtime`, and the pod's env var — no mapping column to drift.
- **operator-scripts: one handoff, at `environment-config`:** `access-credentials` has nothing to do
  and completes without a Vault call. `environment-config` makes no Vault call: for each missing name
  it prints the suggested Vault path and field, the Environment, the secret name and the command shape
  for the operator's own script, pauses the slot (`agents/deployment.md`, operator hand-off rule), and
  completes once `gh secret list --env <machine>` shows every name. The slot is never bound to `none`.
- **Feature deploys touch no Vault unless needed (user, 2026-10-25):** the user chose
  `access-credentials` none on `existing-cicd-pr`. With the PAT retired, `access` loads only runtime
  secrets, which a feature may add, so the default stays `vault-loader access` on every path but is a
  true no-op (no Vault call at all) when the spec declares none; feature-new asks for Vault only when
  the resolved plan's filler will use it. Revisit if a hard `none` is preferred.
- **Deploy workflow:** an optional "Create runtime Secret" step between Lint and Apply (kept only when
  the table is non-empty, like the mesh-join block); step `env:` carries static
  `NAME: ${{ secrets.NAME }}` entries; `run:` refuses an empty value for this machine and pipes an env
  file on stdin into `kubectl create secret generic … --dry-run=client -o yaml | kubectl apply -f -`.
  Pods pick up changed values on the next deploy's rollout (every deploy changes `DEPLOY_SHA`).
- **Manifest enforcement:** `artifact-build` authors the `secretKeyRef` references; `manifest-lint.sh
  --runtime-secrets <names>` refuses a deployment that doesn't reference every name for the machine.
- **Keep** the generic `## Secrets Management` template table; Runtime Secrets is the machine-readable
  existing-k3s table.
- **Keep** `installer/skill_text_test.go`'s rule against raw `vault` commands in skill text.
- **No verify or kubeconfig work:** existing-k3s verifies only with deploy-verify `ci-evidence`
  (tasks 142, 144); task 145 removed the kubeconfig.
- **Out of scope:** VM modes (task 146); task 147's ci-bootstrap state checks. Task 147 also edits
  request-k3s-onboarding; whichever merges second rebases.

## Implementation Steps

**Phase A — spec parser**
1. `installer/infra_spec.go`: `vaultAccessModes` and `VaultAccess` (default `direct`, refusal on any
   other value); `RuntimeSecrets []runtimeSecret` parsed from `## Runtime Secrets` with
   `markdownSection`/`markdownTable`; validate Name against `secretNamePattern` (refuse `GITHUB_`
   prefix and duplicates), Source values, and Environments against declared rows (on the filtered
   `rows`, like mesh/probe); `cmdSpec` text prints `vault access:` and the secret names.
2. `installer/infra_spec_test.go`: add `Vault Access` to `k3sSpec`'s constraint list, an optional
   runtime-secrets section, refusal cases in `TestInfraEnvironmentsRejectIncompleteRows`, defaults in
   `TestInfraEnvironmentsParseTheTable`.
3. `templates/specs/infrastructure.template.md` (row beside `Deployer Exec`, new table) and
   `agents/infrastructure.md` (authoring bullet: omit the row unless the operator states the policy;
   list only secrets the user named).

**Phase B — retire the platform-repo PAT** (parallel with A)

4. request-k3s-onboarding (definition first, then skill): drop the `credential-file.sh` load and
   `GH_TOKEN`; preflight push rights with the operator's login (`gh api repos/<repo> --jq
   .permissions.push`) plus a failure row; reword pre-condition, scope, completion and failure rows.
5. vault-loader: the existing-k3s branch of `load-credentials.sh` stops loading the PAT (Phase C adds
   runtime secrets); drop `apps/<slug>/platform-repo` from `credential-file.sh`'s allowlist and from
   `rotate-credential.sh`; SKILL.md and definition.
6. `installer/slots.go` existing-k3s notes for `access-credentials` and `platform-onboarding`;
   project-new (PAT pre-condition, unconditional Vault pre-condition), feature-new step 9 (Vault only
   when the plan's filler uses it), input-deployment wording.

**Phase C — runtime secrets into Vault and the Environment** (depends on 1)

7. vault-loader `access` (existing-k3s, `direct`): populate
   `secret/apps/<app-slug>/<machine-slug>/runtime` per Runtime Secret and environment, skipping
   populated fields; read the table from `smaqit spec environments --json`. No table, or
   `operator-scripts`: complete with "nothing to load" and no Vault call.
8. repo-config `runtime` operation (`environment-config=runtime`), `scripts/sync-runtime-env.sh`
   modelled on `sync-machine-env.sh`: preflight, every Vault field present before any write,
   `vault kv get -field … | gh secret set --env <machine>`, verify with `gh secret list`. Under
   `operator-scripts`: no Vault call, the exact handoff per missing name, pause, `gh`-only verify.
9. `installer/slots.go`: `environment-config` defaults to `repo-config runtime` for existing-k3s
   (completes immediately with no runtime secrets); rewrite its existing-k3s note; update pinned
   defaults in `installer/slots_test.go`.

**Phase D — deploy workflow and manifests** (depends on 1)

10. `render-k3s-deploy.sh` + `deploy.yml.k3s.template`: project per-machine runtime-secret names into
    `K3S_TARGETS`; optional "Create runtime Secret" block between Lint and Apply with a survival check;
    static step `env:` entries; stdin-only `run:`; extend the leftover-token regex.
11. `deployment.yaml.template`: commented `secretKeyRef` example; `artifact-build` contract note and
    image-build-push-static Step 2; `manifest-lint.sh --runtime-secrets <names>`, passed by the Lint
    step.

**Phase E — tests, docs, release**

12. Tests: stub-`vault` zero-call tests (`h.argv("vault") == ""`) for vault-loader `access`, repo-config
    `runtime` under `operator-scripts`, and onboarding; exact handoff text; `sync-runtime-env.sh` with
    stub `gh` (values land, none on argv or in output); a `runtime-secrets` golden case (bump the
    `2*3+3` count) and step order; extend `stubKubectl` for `create secret` and run the step under
    `execStep`; `TestWorkflowTemplatesKeepSecretsOutOfScripts` stays green; a manifest-lint rule test.
13. Docs: `docs/wiki/workflows/deployment-slots.md` (regenerate the contract block; hand-edit the
    defaults table and existing-k3s Targets prose); `.smaqit/compendium.md` (existing-k3s Vault
    credentials, platform fillers, private-mesh entries); README; CHANGELOG v3.22.0 with an
    action-required note (grant operator push on the platform repo, retire
    `secret/apps/<slug>/platform-repo` with your own tooling, declare runtime secrets); release.

**Verification:** `make -C installer test` and `make -C installer smoke-test` (smoke spec gains a
`Vault Access` row and a Runtime Secrets row); regenerate goldens and the wiki block once with
`go test -run 'TestK3sDeployGoldenFiles|TestSlotContractsMatchWiki' -update` and review the diff;
deliberate breaks (bad Vault Access value, a vault call in the operator-scripts path, a secret moved
into `run:`, a missing `secretKeyRef`, a leftover token) are caught.

## Known Issues Triage
**Triaged:** 2026-10-25
**Tools searched:** GitHub CLI (`cli/cli`), kubectl (`kubernetes/kubectl`), HashiCorp Vault (`hashicorp/vault`), k3s (`k3s-io/k3s`), jq (`jqlang/jq`)
**Result:** Advisory

### Blocking Issues
- None

### Advisory Issues
- [#1610 Kubectl does not support secret creation from `.env` files with multiline values](https://github.com/kubernetes/kubectl/issues/1610) — `kubernetes/kubectl` — opened 2024-06-08 — kind/feature, sig/cli, triage/accepted. Directly relevant: the planned `--from-env-file` stdin step would break a multi-line runtime secret (e.g. a PEM key); build the Secret manifest with `jq` (`stringData`) and pipe it to `kubectl apply -f -` instead.
- [#8817 gh variable/secret commands should url-encode environment name](https://github.com/cli/cli/issues/8817) — `cli/cli` — opened 2024-03-15 — bug, priority-3. Not hit by machine slugs (lowercase letters, digits, inner hyphens), which `secretNamePattern`/slug validation already enforce.
- [#13293 Unable to get just the secret values (data) if the secret itself contains a "data" key](https://github.com/hashicorp/vault/issues/13293) — `hashicorp/vault` — opened 2021-11-29 — secret/kv, devex. Avoided: runtime-secret fields are uppercase names, never `data`.

### Historical (Closed)
- [#14708 gh secret set --body - stores literal "-" instead of reading stdin](https://github.com/cli/cli/issues/14708) — `cli/cli` — closed 2026-10-15. Pipe values to `gh secret set` on stdin with no `--body`, as `sync-machine-env.sh` already does.
- [#13588 Trailing whitespace in secret value when using gh secret set](https://github.com/cli/cli/issues/13588) — `cli/cli` — closed 2026-06-18. Generated values must be written without a trailing newline.

### Unresolvable Tools
- None

### Omitted Tools
- GitHub Actions — no single issue repository; covered by the mapped official docs (secrets context, environment secrets)
- Bash, Go — general-purpose language/runtime; excluded after relevance prioritization and the five-repository limit

### Search Warnings
- None

## Acceptance Criteria

- [x] The `Vault Access` row and `## Runtime Secrets` table are parsed and validated, and bad values
      fail with `SLOTS-SPEC-*` errors
- [x] The platform-repo PAT is gone from request-k3s-onboarding, vault-loader, `credential-file.sh` and
      `rotate-credential.sh`; onboarding preflights push rights with the operator's `gh` login
- [x] Under `operator-scripts`, a stub `vault` records zero calls; the slots complete through an exact,
      tested handoff verified with `gh secret list`, without binding them to `none`
- [x] With no runtime secrets declared, no existing-k3s slot calls Vault on any path
- [x] Runtime secrets reach the rendered `deploy.yml` via step `env:` and stdin (never `run:` or argv),
      and manifest-lint enforces the `secretKeyRef` references
- [x] Golden files regenerated deliberately; no raw `vault` commands or platform-specific names in
      skill text; wiki, compendium, README and CHANGELOG updated; released

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- Parser: `Vault Access` row (default `direct`) and `## Runtime Secrets` table in `installer/infra_spec.go`, with per-row resolved names and a shared `only()` narrowing used by `smaqit spec environments --env`.
- PAT retired: request-k3s-onboarding 3.0.0 uses the operator's `gh` login after a `permissions.push` preflight; `platform-repo` removed from `load-credentials.sh`, `credential-file.sh` and `rotate-credential.sh`.
- Runtime secrets: vault-loader `access` reads the spec before Step 0 and exits without a Vault call when nothing applies; new repo-config `runtime` operation (`sync-runtime-env.sh`, the existing-k3s `environment-config` default) syncs or hands off; optional "Create runtime Secret" deploy step; manifest-lint `secretKeyRef` rule.
- Tests: stub `smaqit` in the script harness serving real-parser JSON; zero-vault-call, handoff-text, sync, render, step-execution and lint tests; a definitions-match-skills test; six deliberate breaks all caught.

**Decisions made:**
- PAT retired for every existing-k3s project (user), not only under `operator-scripts`.
- Runtime secrets Vault → machine-slug Environment, path `secret/apps/<app>/<machine>/runtime`, one field per name (user).
- `access-credentials` kept as vault-loader `access` on every path but a true no-op without runtime secrets, preserving the user's "no Vault on feature deploys" choice while letting a feature add a secret.
- One Name is the Environment secret, the Secret key and the pod env var; a secret not declared for every environment uses `optional: true` since one `deployment.yaml` serves all.
- Secret built with `jq` from step `env:` and applied server-side (no argv, no last-applied annotation; multi-line values work, avoiding kubectl#1610).
- `runtimeSecrets` added to `K3S_TARGETS` only when non-empty, so every existing golden stays byte-identical.

**Blockers encountered:**
- None. One description exceeded the 400-character limit and was shortened; a duplicate test helper was removed.

**Follow-up identified:**
- `rotate-credential.sh` does not rotate runtime secrets yet.
- The platform's deployer Role must allow `secrets` in the Namespace; documented as a failure row only.
- The project-new definition is still stale against its skill (no skill-body markers, so the new sync test doesn't cover it).
- Task 147 and task 148 touch request-k3s-onboarding and the k3s template; rebase onto this release.

## Files to Create / Modify

| File | Action |
|------|--------|
| installer/infra_spec.go, installer/infra_spec_test.go | Modify |
| installer/slots.go, installer/slots_test.go | Modify |
| templates/specs/infrastructure.template.md, agents/infrastructure.md | Modify |
| skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md | Modify |
| skills/smaqit.infrastructure-vault-loader/** (SKILL.md, load-credentials.sh, credential-file.sh, rotate-credential.sh) | Modify |
| skills/smaqit.infrastructure-repo-config/SKILL.md, scripts/sync-runtime-env.sh | Modify / Create |
| skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template, scripts/render-k3s-deploy.sh | Modify |
| skills/smaqit.infrastructure-deploy-k3s-app/assets/deployment.yaml.template, scripts/manifest-lint.sh | Modify |
| skills/smaqit.infrastructure-image-build-push-static/SKILL.md | Modify |
| skills/smaqit.project-new/SKILL.md, skills/smaqit.feature-new/SKILL.md, skills/smaqit.input-deployment/SKILL.md | Modify |
| .smaqit/definitions/skills/ (vault-loader, repo-config, request-k3s-onboarding, cicd-generate, deploy-k3s-app) | Modify |
| installer/skill_scripts_test.go, installer/k3s_render_test.go, installer/k3s_scripts_test.go, installer/testdata/k3s-deploy/* | Modify / Create |
| scripts/smoke-test-installer.sh | Modify |
| docs/wiki/workflows/deployment-slots.md, .smaqit/compendium.md, README.md, CHANGELOG.md | Modify |

## Notes

- Task 142 owns CI-path verification (deploy-verify `ci-evidence`, moved out of this task on
  2026-10-20).
- Tasks 136, 144 and 145 have merged; nothing blocks this task.
- A downstream project worked around the missing runtime-secret support with its own scripts-only
  Vault helper (`--sync-only`, generated values). That helper is a reference point, not a dependency.
- Planning discovery (2026-10-25): spec Constraints other than the parsed ones (`Platform Repo`,
  `Registry File Path`) are read from the spec Markdown by fillers; only the matching Environments row
  reaches `smaqit slots plan` as `Target`, so set-level fields reach fillers through
  `smaqit spec environments --json`.
