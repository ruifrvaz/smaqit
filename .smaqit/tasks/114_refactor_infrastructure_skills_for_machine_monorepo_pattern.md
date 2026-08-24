---
status: Not Started
created: "2026-08-24"
---

# Refactor Infrastructure Skills for the Machine-Monorepo Pattern (Learnings from Magnificah/infrastructure Task 001)

## Description

Found live while provisioning `magnificah-test-01` through `Magnificah/infrastructure` task 001
(2026-08-22 → 2026-08-24, PRs #1–#7 in that repo) — the first real use of the **machine-monorepo
pattern**: an app-agnostic repository that owns Cyso top-level resources (instances, volumes,
security groups, keypairs), per-machine Terraform state, host baselines, and a per-machine
tenancy registry, with application repositories reduced to deploy-only tenants. Nearly the entire
`smaqit.infrastructure-*` family assumes the app-owned-VM pattern (`deployment/terraform/`,
`secret/<project-slug>/*`, `backend/`/`frontend/`, health-endpoint verification, rsync deploys),
so most of it was bypassed and re-derived by hand. Along the way, several skills were found to
carry **actively wrong facts** or misfire in the machine context. Full non-secret evidence:
`Magnificah/infrastructure` → `.smaqit/reports/provisioning-evidence-2026-08-24.md`.

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
  `fd91e198-f162-4b6b-a23e-123304fb408a` (also what the legacy `magnificah-test` VM's applied
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
- **Latent bug left behind**: `secret/machines/magnificah-test-01/metadata` (host/provider/
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
install it) before opening a release-titled PR. (`Magnificah/infrastructure`'s `v0.1.0` remains
untagged as of this filing.)

## Design Decisions

- Scope is refactors to existing artifacts only; new skills are sibling task 115. The two tasks
  together constitute the "smaqit framework update for the machine-repo pattern" that
  `Magnificah/infrastructure` task 001 deferred to this repository.
- Finding 1's corrections update both `skills/smaqit.infrastructure-provider-cyso/references/`
  and the wiki source they mirror, per that skill's own stated convention.
- Finding 2 builds on task 110 / PR #85's `lib-project-slug.sh` work rather than reopening it —
  the remaining gap is the missing registration flow and required metadata, not derivation.
- [TBD at task start: whether Finding 3 extends `infrastructure-repo-config` with a machine
  scheme or delegates to a machine-specific config skill under task 115; whether Finding 5 is a
  mode of `smaqit-deployment` or a distinct machine-deployment agent.]

## Implementation Steps

1. Correct `smaqit.infrastructure-provider-cyso` references (+ wiki source): virtio-scsi attach
   convention with both by-id shapes, explicit `volume_type` requirement + tier default warning +
   exact type names, detach-retype-reattach procedure, single confirmed image UUID, fine-grained
   PAT org-level Secrets/Variables requirement with 403/404 diagnostic mapping.
2. Add a machine-registration flow to `smaqit.infrastructure-vault-loader`
   (`register-machine.sh <machine-slug>`): generate + store `base-ssh`, prompt `cyso` and
   `tfstate` (using the hardened `read_secret` + empty-guards from task 110), write `metadata`
   (provider/owner_project at registration; `host` backfilled post-apply). Document that
   `rotate-credential.sh` is rotation-only. Backfill `secret/machines/magnificah-test-01/metadata`
   as the live proof.
3. Extend `smaqit.infrastructure-repo-config` (or split) for machine-scoped GitHub Environments:
   environment creation/verification, Vault→environment secret/variable mapping for the machine
   scheme, PAT-permission preflight.
4. Encode Finding 4's two workflow gotchas in the owning generation skill's gotchas + templates;
   add a plan-guard/`prevent_destroy` recommendation for gated machine applies.
5. Update `agents/deployment.md` (smaqit-deployment): machine-mode guidance or explicit
   deferral to a machine-deployment path, mandatory provider-knowledge routing for platform
   facts, and the agent-worktree/uncommitted-state caveat.
6. Add the `post-merge-release.yml` existence preflight to `smaqit.task-complete` (Phase 1,
   before the release-titled PR is opened).

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `cyso-reference.md` (skill references + wiki source) reflects every Finding 1 correction;
      no `/dev/vdb` claim, no duplicate image UUID, tier default and PAT org-level requirements
      documented.
- [ ] `register-machine.sh` exists, populates all four `secret/machines/<slug>/` paths
      (base-ssh, cyso, tfstate, metadata) idempotently, takes the machine slug as a required
      argument (no cwd derivation), and passes the skill's no-ad-hoc-secret-reads static check.
- [ ] `secret/machines/magnificah-test-01/metadata` is backfilled and
      `rotate-credential.sh machines/magnificah-test-01/base-ssh` no longer fails its
      metadata precondition (dry-check acceptable; no live rotation required).
- [ ] Machine-scoped GitHub Environment configuration is covered by a skill (extended or new),
      including the PAT-permission preflight with the three diagnostic signatures.
- [ ] The workflow-generation skill's docs/templates encode job-level backend credentials and a
      destroy-guard for gated machine applies.
- [ ] `agents/deployment.md` addresses machine-context misfit and mandates provider-knowledge
      routing for platform facts.
- [ ] `smaqit.task-complete` warns when `post-merge-release.yml` is absent before opening a
      `Prepare release` PR.

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
| `skills/smaqit.infrastructure-provider-cyso/references/cyso-reference.md` | Modify (+ wiki source) |
| `skills/smaqit.infrastructure-vault-loader/scripts/register-machine.sh` | Create |
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Modify |
| `skills/smaqit.infrastructure-repo-config/` | Modify (or split) |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md` | Modify (gotchas) |
| `agents/deployment.md` | Modify |
| `skills/smaqit.task-complete/SKILL.md` | Modify (release-workflow preflight) |

## Notes

Source material: `Magnificah/infrastructure` task file
`.smaqit/tasks/001_provision_magnificah_test_01.md` (Findings section) and
`.smaqit/reports/provisioning-evidence-2026-08-24.md` in that repo. Sibling: task 115 (new
machine-monorepo skills). Related: task 110 / PR #85 (vault-loader slug derivation — in flight,
complementary, not duplicated here).
