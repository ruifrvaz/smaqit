---
status: PR Open
pr: 96
mode: Assisted
started: "2026-10-11"
created: "2026-08-24"
---

# Refactor Infrastructure Skills for the Machine-Monorepo Pattern (Learnings from a Downstream Project's Infrastructure Task 001)

## Description

Found live while provisioning `<machine-slug>` through a downstream project's infrastructure repo task 001
(2026-08-22 → 2026-08-24, PRs #1–#7 in that repo) — the first real use of the **machine-monorepo
pattern**: an app-agnostic repository that owns Cyso top-level resources (instances, volumes,
security groups, keypairs), per-machine Terraform state, host baselines, and a per-machine
tenancy registry, with application repositories reduced to deploy-only tenants. Nearly the entire
`smaqit.infrastructure-*` family assumes the app-owned-VM pattern (`deployment/terraform/`,
`secret/<project-slug>/*`, `backend/`/`frontend/`, health-endpoint verification, rsync deploys),
so most of it was bypassed and re-derived by hand. Along the way, several skills were found to
carry **actively wrong facts** or misfire in the machine context. Full non-secret evidence:
a downstream project's infrastructure repo → `.smaqit/reports/provisioning-evidence-2026-08-24.md`.

This task covers refactoring **existing** skills/agents. Sibling task 115 covers the **new**
machine-monorepo skills.

### Finding 1 — `smaqit.infrastructure-provider-cyso` reference facts are wrong or internally inconsistent

The knowledge router earned its place (it caught a deployment-agent draft that had guessed
`s3.cyso.cloud`/`nl-ams-1` for the Object Storage backend — real values `core.fuga.cloud:8080`/
`ams2`), but its `references/cyso-reference.md` (verified 2026-04-05) carries facts contradicted
by the live machine:

- **Volume attach bus**: the doc says data volumes attach at `/dev/vdb`. The real host attaches
  Cinder volumes over **virtio-scsi**: device `/dev/sdb`, stable udev symlink
  `/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_<full-volume-id>` (plus a `scsi-S...` variant). The
  first-boot mount script written against the documented virtio convention
  (`/dev/disk/by-id/virtio-<id[:20]>`) failed on the real machine (its fail-loud design caught it
  cleanly in `cloud-init-output.log`).
- **Cinder volume-type default**: creating a volume with no explicit `volume_type` resolves to
  **`tier-2`** (~2.6× tier-1's cost), not tier-1 — the doc presents Tier-1 as "the proven
  baseline" without warning that it must be requested explicitly. Exact selectable type names
  (via `openstack volume type list`): lowercase `tier-1`, `tier-1m`, `tier-2`, `tier-2m`.
- **Retype of an attached volume is unsupported** (HTTP 400 despite the error text implying
  in-use is fine): the working remediation is unmount → detach → `openstack volume set --type` →
  reattach → remount.
- **Two inconsistent Ubuntu 24.04 image UUIDs** appear in the same reference (the "Confirmed
  images" table vs. an inline Terraform example). The live-confirmed one is
  `fd91e198-f162-4b6b-a23e-123304fb408a` (also what the legacy `<legacy-machine-slug>` VM's applied
  Terraform uses).
- **Fine-grained PAT permissions for environment-scoped secrets/variables**: repo-level
  Environments + Secrets + Variables grants were **not sufficient** — writes to a GitHub
  Environment's secrets/variables 403'd until **organization-level Secrets and Variables
  read/write** was also enabled on the token. Cost three troubleshooting rounds; documented
  nowhere in the skill family. (Diagnostic signatures: env create → 403 "Resource not accessible
  by personal access token"; var set against a nonexistent env → 404; var set with repo-only
  grants → 403.)

The skill's own gotchas note that `references/` mirrors `docs/wiki/` content — both copies need
the correction.

### Finding 2 — `smaqit.infrastructure-vault-loader` has no machine-registration flow

First-time population of `secret/machines/<slug>/{base-ssh,cyso,tfstate}` has no entrypoint:

- `load-credentials.sh MACHINE_SLUG=<slug>` is **app-bootstrap** mode — it only prompts for a
  GitHub PAT under `secret/apps/<derived-slug>/github` and explicitly never writes machine paths
  ("populated once at machine registration" — a flow that doesn't exist). Run from an unrelated
  cwd, it silently wrote a real PAT to `secret/apps/scripts/github` (slug derived from
  `~/projects/scripts`); the stray credential had to be traced by `created_time` and deleted.
  Task 110 / PR #85 fixed slug *derivation*; this incident happened **with** the fixed
  derivation — a correct slug for the wrong directory. A dedicated machine-registration
  entrypoint taking the machine slug as a required argument sidesteps the entire class.
- The workaround was `rotate-credential.sh machines/<slug>/{cyso,tfstate}` (a rotation tool doing
  first-time population) plus hand-running `ssh-keygen` for `base-ssh`.
- **Latent bug left behind**: `secret/machines/<machine-slug>/metadata` (host/provider/
  owner_project) was never written, and `rotate-credential.sh machines/<slug>/base-ssh` hard-fails
  without `metadata.host` — so base-ssh rotation for the machine provisioned by task 001 is
  currently broken until metadata is backfilled. Registration must make metadata a first-class
  step (host filled in post-apply).

### Finding 3 — `smaqit.infrastructure-repo-config` cannot configure machine-scoped GitHub Environments

The skill is repo-level-secrets, app-scheme only (`VM_SSH_KEY`, `VM_HOST`, `GH_TERRAFORM_TOKEN`,
`secret/<slug>/*`). The machine model needs: one GitHub **Environment** per machine (named after
the machine slug), populated from `secret/machines/<slug>/*` with machine-shaped names
(`BASE_SSH_PUBLIC_KEY`, `BASE_SSH_PRIVATE_KEY`, `CYSO_APPLICATION_CREDENTIAL_ID`,
`CYSO_APPLICATION_CREDENTIAL_SECRET`, `CYSO_S3_ACCESS_KEY`, `CYSO_S3_SECRET_KEY`) plus variables
`RECONCILE_SSH_USER` and post-apply `FIXED_IP`. It was all hand-rolled with
`vault kv get | gh secret set --env`. The skill also needs a PAT-permission preflight mapping the
diagnostic signatures from Finding 1's last bullet to actionable fixes.

### Finding 4 — CI/CD generation gotchas proven live

Two workflow-authoring mistakes cost a failed run each; whichever skill owns workflow generation
(`smaqit.infrastructure-cicd-generate` today, or task 115's machine skills) must encode them:

- **S3-backend credentials must be job-level `env:`**, not step-level on `terraform init` only —
  GitHub Actions steps don't inherit sibling-step env, and *every* `terraform` invocation
  (validate excepted, plan/apply included) reads the backend. First live `plan` failed with "No
  valid credential sources found".
- **The gated apply job has no `plan-guard.sh`/`prevent_destroy` equivalent** in the machine
  workflow set — a `user_data`/`image_id` diff would destroy-and-recreate the machine with only
  human plan-reading as the gate. `infrastructure-provision-cyso` has this protection for the app
  pattern; the machine pattern needs its analog.

### Finding 5 — `smaqit-deployment` agent misfits the machine context

The agent was usable only under heavy inline fencing (author files only; never commit/push/PR;
never apply/dispatch; credentials by reference only). Structural issues:

- Its workflow assumes app deployment (deploy-verify SHA/health coupling, `deployment_path`
  branching, Stack-spec `backend/`/`frontend/` contracts). Verification for a machine is
  SSH/cloud-init/mounts/services/security-group/reboot — none of which it knows.
- It **guessed platform facts instead of consulting `infrastructure-provider-cyso`** — every
  guessed fact (endpoint, region, network id vs name, volume type, bucket) was wrong or
  suboptimal. The agent definition should require routing platform facts through provider
  knowledge skills.
- Operational caveat worth recording: its isolated agent worktree was checked out from the
  branch's last *commit* and couldn't see the task worktree's uncommitted state (smaqit's
  Assisted-mode flow deliberately keeps implementation uncommitted until task-complete) — the
  agent had to escape into the real task worktree to function.

### Finding 6 — `smaqit.task-complete` / release chain assumes `post-merge-release.yml` exists

The full `Prepare release vX.Y.Z` convention was followed in a repo that doesn't have
`post-merge-release.yml` installed — so merging PR #1 never created the `v0.1.0` tag or GitHub
Release, silently. The chain should preflight the workflow's existence and warn (or offer to
install it) before opening a release-titled PR. (a downstream project's infrastructure repo's `v0.1.0` remains
untagged as of this filing.)

## Issue Triage Context

**Mode:** Auto
**Technologies:** Bash, Go, HashiCorp Vault, GitHub CLI, Terraform, OpenStack, GitHub Actions
**Platforms/Environments:** Linux, WSL2, Cyso Cloud (OpenStack)
**Features/Integrations:** Vault KV v2, GitHub Environment secrets/variables, fine-grained PATs, Cinder volumes (virtio-scsi), Terraform S3 backend, smaqit deployment slots
**Versions/Constraints:** smaqit 3.7.0, Agent Skills metadata values must be strings, fine-grained PAT needs org-level Secrets/Variables for Environment writes

## Design Decisions

- Scope is refactors to existing artifacts only; new skills are sibling task 115. The two tasks
  together constitute the "smaqit framework update for the machine-repo pattern" that
  a downstream project's infrastructure repo task 001 deferred to this repository.
- Finding 1's corrections update `skills/smaqit.infrastructure-provider-cyso/references/` only:
  no `docs/wiki/` copy of the Cyso reference exists, so the skill's own claim that its
  references mirror the wiki (SKILL.md:59) is corrected instead.
- Finding 2 builds on task 110 / PR #85's `lib-project-slug.sh` work rather than reopening it —
  the remaining gap is the missing registration flow and required metadata, not derivation.
- Confirmed during `task.plan` (2026-10-11):
  - **Finding 6 moves to smaqit-extensions task 042,** together with the smaqit-extensions half
    of Finding 5's worktree caveat. The task-lifecycle and release skills live only in that
    repository; nothing in this repository changes for Finding 6.
  - **Finding 2:** a new `register-machine.sh`, declared in vault-loader's `metadata.smaqit-slots`
    as `target-credentials=register`. It replaces only provision-cyso Step 7.5's raw
    `vault kv put …/metadata`; the rest of the raw-Vault cleanup stays in task 123.
  - **Finding 3:** a new `sync-machine` operation on `smaqit.infrastructure-repo-config`
    (`environment-config=sync-machine`), implemented in scripts rather than inline `vault`
    commands. The PAT-permission preflight is a shared script used by both the machine
    operation and the existing k3s per-environment step.
  - **Finding 4:** gotchas only. Machine workflow templates are task 115's.
  - **Finding 5:** machine contexts are a mode of `smaqit-deployment` through deployment slots
    (resolved by task 125). This task adds mandatory provider-knowledge routing, the
    isolated-worktree caveat, and verify taken from the bound `verify` filler.
  - **Live Vault step:** the metadata backfill is run by the user through their own Vault
    tooling; the implementing agent runs no `vault` command and records the commands in
    Findings.
  - **Delivery:** its own branch from `origin/main` and its own `Prepare release vX.Y.Z` PR.

## Implementation Steps

*Phase A — Cyso facts (Finding 1)*

1. `skills/smaqit.infrastructure-provider-cyso/references/cyso-reference.md`:
   - Replace the `/dev/vdb` data-volume convention (~219–225) with virtio-scsi `/dev/sdb` and
     both stable symlink shapes (`/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_<full-volume-id>`,
     `scsi-S…`).
   - Volume types: lowercase `tier-1`, `tier-1m`, `tier-2`, `tier-2m`; warn that omitting
     `volume_type` resolves to `tier-2`; add an explicit `volume_type` to the Terraform example
     (~244–247).
   - Add the detach → `openstack volume set --type` → reattach → remount retype procedure.
   - Keep the single confirmed Ubuntu 24.04 UUID `fd91e198-f162-4b6b-a23e-123304fb408a`;
     replace `bcdcd204-…` (~411).
   - Add the fine-grained PAT org-level Secrets/Variables requirement for Environment writes,
     with the three diagnostic signatures (env create 403, var set against a missing env 404,
     var set with repo-only grants 403).
   - Bump the verified date.
2. Apply the same corrections to the sibling references: `cyso-deployment-setup.md` (~160–163,
   `/dev/vdb`) and `cyso-pricing.md` (type names, default warning).
3. `provider-cyso/SKILL.md`: correct the stale "copies of `docs/wiki/` content" claim (~59);
   bump `metadata.version`.

*Phase B — machine registration (Finding 2), parallel with A*

4. Create `skills/smaqit.infrastructure-vault-loader/scripts/register-machine.sh <machine-slug>
   [--host <ip>]`:
   - The slug is a required argument; never derived from the working directory.
   - Hidden reads use the `/dev/tty` `read_secret` pattern so `check-no-ad-hoc-secret-reads.sh`
     passes.
   - Idempotent: keeps existing fields; generates `base-ssh` only if missing; prompts `cyso` and
     `tfstate`; writes `metadata` (`provider`, `owner_project` defaulting to the repo slug via
     `lib-project-slug.sh`); `--host` backfills `metadata.host` after apply.
5. `vault-loader/SKILL.md`: declare `target-credentials=register` alongside `access` and
   `target`; document registration and that `rotate-credential.sh` is rotation-only and requires
   `metadata.host`; fix the claim (~250) that `bootstrap-app-to-machine.sh` prompts for
   cyso/tfstate; bump `metadata.version`.
6. `provision-cyso/SKILL.md` Step 7.5 (~141–144): replace the raw metadata write with
   `register-machine.sh <slug> --host <fixed_ip>`; bump `metadata.version`.

*Phase C — machine Environments (Finding 3), depends on B for the path schema*

7. Create `skills/smaqit.infrastructure-repo-config/scripts/sync-machine-env.sh <owner>/<repo>
   <machine-slug>`: create or verify the Environment named after the machine slug
   (`gh api repos/<owner>/<repo>/environments/<slug> -X PUT`); map `secret/machines/<slug>/*` to
   `BASE_SSH_PUBLIC_KEY`, `BASE_SSH_PRIVATE_KEY`, `CYSO_APPLICATION_CREDENTIAL_ID`,
   `CYSO_APPLICATION_CREDENTIAL_SECRET`, `CYSO_S3_ACCESS_KEY`, `CYSO_S3_SECRET_KEY`; set the
   variables `RECONCILE_SSH_USER` and `FIXED_IP` (from `metadata.host`). Secrets travel by stdin,
   never by argument or stdout.
8. Create `skills/smaqit.infrastructure-repo-config/scripts/check-pat-permissions.sh`: probe the
   token and map the three signatures to the fix (org-level Secrets/Variables read/write). Used
   by `sync-machine` and by the existing k3s per-environment step.
9. `repo-config/SKILL.md`: declare `environment-config=sync, environment-config=sync-machine`;
   add the `sync-machine` operation section and the preflight; add Failure Handling rows; bump
   `metadata.version`.
10. `docs/wiki/workflows/deployment-slots.md`: extend the platform-repository example to bind
    `target-credentials` → vault-loader `register` and `environment-config` → repo-config
    `sync-machine`.

*Phase D — gotchas and agent (Findings 4–5)*

11. `cicd-generate/SKILL.md` Gotchas: S3-backend credentials belong in job-level `env:` (steps
    don't inherit sibling-step env); a gated apply against an existing machine needs plan-guard
    and `lifecycle { prevent_destroy = true }` on the instance and volumes. Bump
    `metadata.version`.
12. `agents/deployment.md`: platform facts (endpoints, regions, image/volume/network IDs) must
    come from the provider-knowledge skill, never be guessed; a delegated subagent in an
    isolated worktree sees only committed state, so run it in the task worktree; for non-app
    targets verification is whatever the `verify` slot binds. Rebase onto task 131 first if it
    has merged (its edits are in other sections).

*Phase E — verification, docs, release*

13. `installer/slots_test.go`: `smaqit slots plan` resolves bindings to `register` and
    `sync-machine`; a binding without `operation` against a skill declaring several operations
    fails with `SLOTS-OPERATION-INVALID`.
14. Stub harness: run `register-machine.sh` and `sync-machine-env.sh` against PATH-shadowed
    `vault`/`gh` stubs that record calls; assert idempotency, the required slug, and that no
    secret value reaches stdout or argv. Run `check-no-ad-hoc-secret-reads.sh`.
15. `make -C installer prepare test smoke-test`; `grep -rn '/dev/vdb\|bcdcd204' skills/` returns
    nothing.
16. `.smaqit/compendium.md`: entry on how a platform repository registers a machine and
    configures its Environment through slots. `CHANGELOG.md` (MINOR) and a
    `Prepare release vX.Y.Z` PR.
17. Record in Findings the user-run commands for the live backfill and its check (the user's own
    `vault-add-credential machines/<slug>/metadata` and `--check`).

## Known Issues Triage
**Triaged:** 2026-10-11
**Tools searched:** GitHub CLI (`cli/cli`), Terraform (`hashicorp/terraform`), OpenStack Terraform provider (`terraform-provider-openstack/terraform-provider-openstack`), OpenStack CLI (`openstack/python-openstackclient`)
**Result:** Advisory

### Blocking Issues
- None

### Advisory Issues
- [#8817 gh variable/secret commands should url-encode environment name](https://github.com/cli/cli/issues/8817) — `cli/cli` — opened 2024-03-15 — bug, priority-3. Machine-slug Environment names are lowercase-hyphenated, so no encoding is needed; `sync-machine-env.sh` should validate the slug charset.
- [#5149 create/edit/list/delete repo environment](https://github.com/cli/cli/issues/5149) — `cli/cli` — opened 2022-02-01 — enhancement. `gh` has no environment command; create Environments through `gh api …/environments/<name> -X PUT`, as planned.
- [#37130 Regression for certain third-party S3 backends over S3 v2 API as of 1.11.2 (ceph, hetzner)](https://github.com/hashicorp/terraform/issues/37130) — `hashicorp/terraform` — opened 2025-05-20 — bug, backend/s3. Cyso Object Storage is a non-AWS S3 backend; relevant to the cicd-generate gotchas, not confirmed for Cyso.
- [#37366 Backend S3 try to load aws credentials on non aws backends](https://github.com/hashicorp/terraform/issues/37366) — `hashicorp/terraform` — opened 2025-07-23 — enhancement, backend/s3.

### Historical (Closed)
- [#37985 Terraform S3 Backend fails with "No valid credential sources found" using aws login](https://github.com/hashicorp/terraform/issues/37985) — `hashicorp/terraform` — closed 2025-12-10. Same error signature as Finding 4, different cause.
- [#37203 (maybe) 1.12 S3 Backend (nonAWS) Regression: skip_s3_checksum ignored?](https://github.com/hashicorp/terraform/issues/37203) — `hashicorp/terraform` — closed 2025-09-23
- [#13588 Trailing whitespace in secret value when using gh secret set](https://github.com/cli/cli/issues/13588) — `cli/cli` — closed 2026-06-18. Relevant to piping secrets through stdin.

### Unresolvable Tools
- HashiCorp Vault — helper resolved to `hashicorp/vault-guides`, not the product repository; not searched

### Omitted Tools
- Bash, Go, GitHub Actions — excluded after relevance prioritization and the five-repository limit

### Search Warnings
- None. Research-map task block had no GitHub URLs; all repositories came from the helper fallback.

## Acceptance Criteria

- [x] `cyso-reference.md` reflects every Finding 1 correction: no `/dev/vdb` claim, tier default
      and lowercase type names, retype procedure, and PAT org-level requirement documented. No
      `bcdcd204` UUID anywhere in `skills/`; `cyso-deployment-setup.md` and `cyso-pricing.md`
      are corrected; `provider-cyso/SKILL.md` no longer claims a `docs/wiki/` mirror.
- [x] `register-machine.sh` exists, populates all four `secret/machines/<slug>/` paths
      (base-ssh, cyso, tfstate, metadata) idempotently, takes the machine slug as a required
      argument (no cwd derivation), and passes the skill's no-ad-hoc-secret-reads static check.
      vault-loader declares it as `target-credentials=register`, and provision-cyso Step 7.5
      calls it instead of a raw `vault kv put`.
- [x] Findings record the user-run commands that backfill `secret/machines/<machine-slug>/metadata`
      and check it, through the user's own Vault tooling. The implementing agent runs no `vault`
      command.
- [x] repo-config's `sync-machine` operation creates or verifies the machine-slug Environment and
      maps the six machine secrets and two variables. `check-pat-permissions.sh` maps the three
      diagnostic signatures and is used by both the machine and the k3s per-environment step.
- [x] `cicd-generate`'s Gotchas encode job-level backend credentials, and plan-guard plus
      `prevent_destroy` for gated applies against an existing machine.
- [x] `agents/deployment.md` mandates provider-knowledge routing for platform facts, states the
      isolated-worktree caveat, and states that non-app verification comes from the bound
      `verify` filler.
- [x] `make -C installer prepare test smoke-test` passes; Go tests cover both new operations; a
      stub vault/gh harness proves idempotency and that no secret reaches stdout or argv.
- [x] The `deployment-slots.md` platform example, the compendium, and `CHANGELOG.md` are updated.

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- Corrected the Cyso facts in `cyso-reference.md` and its two sibling references: virtio-scsi `/dev/sdb` with both by-id shapes, lowercase volume types with the `tier-2` default warning and an explicit `volume_type` in the Terraform example, the detached retype procedure, one image UUID, and a GitHub-token section with the three PAT signatures. Header records the 2026-08-24 correction date rather than claiming a fresh full verification.
- Added `register-machine.sh` (vault-loader, `target-credentials=register`): required slug with a URL-safe charset, keeps existing paths, terminal-only prompts, secrets handed to Vault on stdin (`field=-`), `--host`/`--provider`/`--owner-project` patch metadata. provision-cyso Step 7.5 now calls it.
- Added repo-config `sync-machine` (`sync-machine-env.sh`) and the shared `check-pat-permissions.sh` (read-only probes, exit 2 for a missing Environment, 3 for a missing grant); the latter also gates the k3s per-environment step.
- Added cicd-generate gotchas for workflows outside the templates, and deployment-agent directives for provider-knowledge routing, the isolated-worktree caveat, and filler-defined verification.
- Tests: `TestSlotPlanPlatformRepositoryBindings` plus `installer/skill_scripts_test.go`, which runs the real scripts against file-backed stub `vault`/`gh` binaries, using `script(1)` for a terminal and `setsid` for the no-terminal case. A deliberate argv leak was injected to confirm the leak assertion fails.
- `make -C installer test` and `make -C installer smoke-test` pass.

**Decisions made:**
- `sync-machine` treats `metadata.host` as optional: `environment-config` precedes `provision` in the slot order and CI needs the credentials before the gated apply, so the first run skips `FIXED_IP` with a notice and a single-slot re-run after the `--host` backfill sets it. This changes the planned behaviour, which required the host.
- Secrets flow by stdin rather than argv, stricter than the existing vault-loader scripts, so they never appear in the process list.
- No `docs/wiki/` copy of the Cyso reference exists; the skill's claim that its references mirror the wiki was replaced rather than creating a wiki page.
- The isolated-worktree caveat is a deployment-agent MUST NOT here; the task-lifecycle half moved to smaqit-extensions task 042, together with Finding 6.
- The provision-cyso image check was moved from `Ubuntu 22.04` to `Ubuntu 24.04` as part of the image-fact correction.
- CHANGELOG is written by task completion, not by the implementer.

**Blockers encountered:**
- Issue triage resolved HashiCorp Vault to `hashicorp/vault-guides`, so Vault was not searched; recorded in the triage block.
- The planned verification grep (`/dev/vdb` returns nothing) is too strict: the corrected references deliberately say "not `/dev/vdb`". Verified instead that every remaining match is such a warning.

**Follow-up identified:**
- User-run live backfill of the downstream machine's metadata, through the user's own Vault tooling: `~/projects/scripts/vault-add-credential machines/<machine-slug>/metadata`, then `~/projects/scripts/vault-add-credential --check machines/<machine-slug>/metadata`. The implementing agent ran no `vault` command.
- Task 123: provision-cyso Step 2 still generates `base-ssh` with an inline raw `vault kv put`; replace it with `register-machine.sh <machine-slug>`. The remaining vault-loader scripts still pass secrets as arguments rather than stdin.
- Task 115: machine workflow templates and default machine fillers for `provision`, `verify`, `host-bootstrap`, `replace` can now bind next to `register` and `sync-machine`.
- smaqit-extensions task 042: release-workflow preflight and the task-lifecycle worktree caveat.
- The triage helper's repository resolution for HashiCorp Vault returns the wrong repository.

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-provider-cyso/references/{cyso-reference,cyso-deployment-setup,cyso-pricing}.md`, `SKILL.md` | Modify |
| `skills/smaqit.infrastructure-vault-loader/scripts/register-machine.sh` | Create |
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Modify |
| `skills/smaqit.infrastructure-provision-cyso/SKILL.md` | Modify (Step 7.5 only) |
| `skills/smaqit.infrastructure-repo-config/scripts/{sync-machine-env,check-pat-permissions}.sh` | Create |
| `skills/smaqit.infrastructure-repo-config/SKILL.md` | Modify |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md` | Modify (gotchas) |
| `agents/deployment.md` | Modify |
| `installer/slots_test.go` | Modify |
| `docs/wiki/workflows/deployment-slots.md`, `.smaqit/compendium.md`, `CHANGELOG.md` | Modify |

## Notes

Source material: a downstream project's infrastructure repo task file
`.smaqit/tasks/001_provision_<machine-slug>.md` (Findings section) and
`.smaqit/reports/provisioning-evidence-2026-08-24.md` in that repo. Sibling: task 115 (new
machine-monorepo skills). Related: task 110 / PR #85 (vault-loader slug derivation — in flight,
complementary, not duplicated here).
- **Re-scoped by task 125 (2026-10-11):** Finding 5's open decision is resolved there — machine contexts run as a mode of `smaqit-deployment` through deployment slots, not a distinct agent. Findings 2–3 (machine registration, machine-scoped Environments) become requirements on the `credentials` and `environment-config` slot contracts; a platform repository may fill those slots with its own skills.
- **Planned 2026-10-11 (`task.plan`):** Finding 6 moved to smaqit-extensions task 042; machine
  workflow templates (Finding 4) deferred to task 115. Overlaps task 123 (`repo-config`,
  `vault-loader`) and task 131 (`agents/deployment.md`); whichever lands second rebases.
- **Refreshed after task 125 implementation (2026-10-11):** the shipped slot names are `access-credentials` / `target-credentials` (not a single `credentials` slot) and `environment-config`; machine registration maps to `target-credentials`, machine-scoped Environments to `environment-config`. Bind them in `.smaqit/slots.yml`; fillers declare `metadata.smaqit-slots` (a string of `slot=operation` pairs). Inspect with `smaqit slots plan`.
