---
status: Completed
mode: Assisted
started: "2026-10-12"
created: "2026-08-24"
completed: "2026-10-12"
---

# Platform Repositories: Guarded Replacement, Cyso Facts, and a Pattern Guide

## Description

**Rescoped 2026-10-12 (`task.plan`).** The four canonical skills this task originally asked for
are dropped. Since the task was filed, the downstream platform repository moved every machine to
single-server k3s reached over a mesh network: its provisioning and host-baseline mechanisms are
live but materially changed, machine verification was never packaged as one unit, and the
`tenants.yml` VM tenancy was retired on 2026-09-05 in favour of k3s Namespace onboarding. Those
mechanisms are now that repository's own project skills, which it plans to bind into smaqit's
deployment slots itself. Shipping canonical copies would duplicate live skills and hardcode one
contributor's mechanics as a universal contract (the task 116 thin-dispatcher rule).

What remains is the generic, smaqit-owned work the downstream exposed:

1. **Guarded replacement.** `plan-guard.sh` rejects every planned delete, which also blocks a
   deliberate `terraform plan -replace=<address>` — the only way to replace an instance whose
   `user_data` is under `ignore_changes`. The guard also discards the plan it inspected, so the
   caller's `terraform apply` re-plans instead of applying what was checked.
2. **Cyso facts.** The reference's S3 backend block uses superseded syntax and lacks the flags
   proven against Cyso Object Storage; Nova's 65,535-byte `user_data` limit and the gzip
   embedding that avoids it are undocumented.
3. **A platform-repository guide.** The pattern (layout, per-machine state and Environments,
   gated dispatch, day-two through workflows, verification checklist) and its live-found gotchas
   have no home in smaqit's documentation.
4. **A cross-reference** from the app-side skills to how an app is hosted on a platform
   repository's machine.

### Original scope (superseded)

Companion to task 114. A downstream project's infrastructure repo task 001 (2026-08-22 → 2026-08-24, PRs #1–#7)
established and live-verified the **machine-monorepo pattern** — an app-agnostic repository
owning cloud top-level resources, per-machine Terraform state, host baselines, and a per-machine
tenancy registry, with application repositories reduced to deploy-only tenants that register as
opaque `tenants.yml` rows. Every major piece of that pattern had to be invented from scratch
because no skill covers it; this task turns the four proven, working implementations into
reusable smaqit skills so the next machine (or the next org adopting the pattern) doesn't
re-derive them.

A cross-cutting lesson worth baking into all four: **live exercising found every bug** — five
real defects (S3-backend credential env scoping, wrong Cinder tier default, wrong volume-attach
bus assumption, ssh silently re-splitting a multi-word `deploy_key` across argument boundaries,
an unconditional `touch` breaking strict idempotency) — and *none* were caught by
`terraform validate`, YAML parsing, design gates, or code review. The verify and reconcile skills
are where that leverage lives.

### Skill 1 — `smaqit.infrastructure-provision-machine`

The machine-monorepo provisioning runbook, distinct from `smaqit.infrastructure-provision-cyso`
(app-owned VM, local apply, flat Vault scheme):

- Layout: shared machine-agnostic modules (`modules/{compute-instance,data-volume,
  security-group}`) composed by a per-machine root (`machines/<name>/terraform`); volume creation
  separated from attachment (the attachment is composed at the root because cloud-init user-data
  needs the volume ID before the instance exists).
- State: per-machine key in S3-compatible Object Storage — isolation by **key**, not bucket
  (bucket reuse is fine and avoids manual bucket creation); `skip_s3_checksum = true` and
  companion flags for OpenStack-based stores.
- Credentials: machine-scoped GitHub Environment named after the machine, backed by
  `secret/machines/<slug>/*` (see task 114's repo-config work); provider `auth_url`/`region`
  hardcoded as public platform constants so a misconfigured environment can't silently target the
  wrong region.
- Governance: `terraform plan` on PR (paths-scoped to the machine dir + modules), apply only via
  human-gated `workflow_dispatch` with an exact-match `confirm_machine` input; record the fixed
  IP from outputs and set it as the environment's `FIXED_IP` variable post-apply.
- ForceNew safety: `ignore_changes = [user_data]` on the instance (first-boot-only content must
  never force replacement via an incidental diff), plus the destroy-guard from task 114.

### Skill 2 — `smaqit.infrastructure-machine-verify`

The machine analog of `smaqit.infrastructure-deploy-verify` (which is app-health-shaped: health
endpoint, SHA, SPA root — none applicable to a machine). PASS/FAIL report over:

- SSH reachability with the machine's base-ssh key; `cloud-init status` done
- Data volume: `findmnt` on the mount path, UUID-pinned fstab entry, write test
- Host services: expected packages installed and `systemctl is-active` (e.g. docker, nginx),
  reverse proxy answering on localhost
- Security group verified via the **cloud API directly** (`openstack security group show`),
  independent of Terraform state — exactly the declared ingress rules and nothing else
- Reboot survival: reboot, wait for return, re-verify mount + services
- `terraform plan` reports zero drift
- Gotcha inherited from live use: a reused floating IP leaves a stale `known_hosts` entry — the
  skill should anticipate the host-key-changed failure and document the `ssh-keygen -R` fix
  rather than letting it read as a MITM scare.

### Skill 3 — `smaqit.infrastructure-host-baseline` (day-2 ops)

Encodes the principle the downstream operator set explicitly mid-task: *no ad-hoc remediations —
fix provisioning, and run day-2 ops through workflows*. Cloud-init runs first-boot only, so a
fixed host script can never retroactively heal an already-provisioned machine; the proven
alternative is a `host-baseline.yml` workflow:

- `workflow_dispatch` with a `machine` input; uses that machine's own GitHub Environment
- Reads the target's fixed IP and data-volume ID **from the machine's own Terraform outputs**
  (`terraform init` + `output -raw`) — no manually-maintained host variables to drift
- Streams each host-baseline script over SSH (`sudo bash -s -- <args> < script`) in order; every
  script must be idempotent (guard-then-act, no unconditional mutations)
- Proven live: it remediated the failed first-boot volume mount (wrong attach-bus assumption)
  with zero manual SSH, correctly skipping the already-installed packages.

### Skill 4 — `smaqit.infrastructure-tenant-reconcile`

The tenancy half of the pattern — the contract between a machine and its (opaque) application
tenants:

- `tenants.yml` schema: `slug`, `uid`/`gid`, `hostnames` (map of environment → list), `tier`
  (`static` | `full-stack`), `deploy_key` (public key material only), `dns_state`
  (`pending` | `active`), `port` required-iff-full-stack
- Fail-closed registry validation before any grant: unique slugs/uids/ports; PR-triggered dry-run
  validation job (safe, host-untouched) + gated `workflow_dispatch` apply
- Idempotent per-tenant grants over SSH: system user/group with declared-vs-existing UID/GID
  divergence treated as an error (never silently adopted), data-dir ownership under the shared
  volume, single-key `authorized_keys` contract, sudoers scoped to the tenant's own nginx site
  (visudo-validated), loopback port registry for full-stack tier only (removed when a tenant
  drops to static)
- DNS ledger verification: `active` hostnames must resolve to the machine's fixed IP (fail
  closed); `pending` hostnames reported, never failing
- Idempotency defined at the **filesystem level** (repeat run leaves mtimes untouched), not just
  "logs said no changes"
- Two hard-won gotchas that must be in the skill: (1) ssh does not preserve argv boundaries for
  remote commands — every remote argument must be `shlex.quote()`d (or equivalent) or a
  multi-word `deploy_key` silently corrupts the positional args behind it; (2) no unconditional
  `touch`/`chmod`-style mutations outside a guard, or idempotency silently degrades to
  content-only.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Terraform, Bash, jq, OpenStack, GitHub Actions, cloud-init
**Platforms/Environments:** Linux, WSL2, Cyso Cloud (OpenStack)
**Features/Integrations:** terraform plan -replace, saved plan files, Terraform S3 backend, Nova user_data, cloud-init gzip encoding, smaqit deployment slots
**Versions/Constraints:** smaqit 3.10.0, Terraform >= 1.6 S3 backend syntax, Nova user_data limit 65535 bytes base64-encoded

## Design Decisions

Confirmed by the user on 2026-10-12, during `task.plan`:

- **No canonical machine skills.** Provision, verify, host-baseline, and the k3s onboarding
  successor stay platform-repository-owned fillers, bound through deployment slots. This follows
  the task 116 rule that smaqit defines hand-off points, not one contributor's mechanisms.
- **`tenant-reconcile` is out:** superseded downstream by k3s Namespace onboarding, whose
  receiving side is the platform's own contract (smaqit ships only the app side).
- **The guard stays backward compatible:** with no flags, `plan-guard.sh` behaves exactly as
  today, so vendored copies and the cicd-generate templates are unaffected.
- **Fail closed on replacement:** an allowed address passes only when the plan replaces it
  (delete plus create). A pure destroy of it, or an allowed address the plan doesn't replace, is
  blocked, as is every other delete.
- **`--out` is opt-in** and closes the gap between the guarded plan and the applied one.
- **Machine verification is a checklist in the guide, not a skill:** the checks depend on the
  machine's class.
- **No downstream project names** in shipped content or the guide (CONTRIBUTING).
- **Delivery:** its own branch from `origin/main` and its own `Prepare release vX.Y.Z` PR.

## Known Issues Triage
**Triaged:** 2026-10-12
**Tools searched:** Terraform (`hashicorp/terraform`), OpenStack Terraform provider (`terraform-provider-openstack/terraform-provider-openstack`), cloud-init (`canonical/cloud-init`)
**Result:** Advisory

### Blocking Issues
- None

### Advisory Issues
- [#37130 Regression for certain third-party S3 backends over S3 v2 API as of 1.11.2 (ceph, hetzner)](https://github.com/hashicorp/terraform/issues/37130) — `hashicorp/terraform` — opened 2025-05-20 — bug, backend/s3, new. Relevant to the documented non-AWS backend flags; not confirmed for Cyso.
- [#38337 S3 backend: skip_s3_checksum does not disable aws-chunked content-encoding on PutObject](https://github.com/hashicorp/terraform/issues/38337) — `hashicorp/terraform` — opened 2026-04-01 — enhancement, backend/s3, new
- [#36704 Apply `skip_s3_checksum` option to `GetObject` calls](https://github.com/hashicorp/terraform/issues/36704) — `hashicorp/terraform` — enhancement, backend/s3, new
- [#35305 change suggested `apply` command after saving plan file to account for `-chdir`](https://github.com/hashicorp/terraform/issues/35305) — `hashicorp/terraform` — enhancement, cli. The guard changes directory before planning, so `--out` must resolve to an absolute path.

### Historical (Closed)
- [#39076 Replacement planned as ["delete","create"] is applied create-before-destroy, and the destroy never runs](https://github.com/hashicorp/terraform/issues/39076) — `hashicorp/terraform` — closed. A replacement can appear in either action order; the guard must accept both.
- [#35302 Terraform does not respect -replace option](https://github.com/hashicorp/terraform/issues/35302) — `hashicorp/terraform` — closed, working as designed. Supports failing closed when a listed address is not replaced in the plan.
- [#37648 The state can not be persisted in custom S3 endpoint. The Content-SHA256 you specified did not match](https://github.com/hashicorp/terraform/issues/37648) — `hashicorp/terraform` — closed. The reason for `skip_s3_checksum` on non-AWS stores.
- [#37203 (maybe) 1.12 S3 Backend (nonAWS) Regression: skip_s3_checksum ignored?](https://github.com/hashicorp/terraform/issues/37203) — `hashicorp/terraform` — closed
- [#4042 text/plain encoding for write_files unsupported?](https://github.com/canonical/cloud-init/issues/4042) — `canonical/cloud-init` — closed

### Unresolvable Tools
- None

### Omitted Tools
- Bash, jq, GitHub Actions — excluded after relevance prioritization; not material to the task's failure modes

### Search Warnings
- None. Research-map task block had no GitHub URLs; all repositories came from the helper fallback.

## Implementation Steps

*Phase A — guarded replacement*

1. `skills/smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh`: add repeatable
   `--allow-replace <address>` (passed to `terraform plan` as `-replace=<address>`) and
   `--out <planfile>` (keep the inspected binary plan), keeping the positional directory argument.
   An allowed address passes only as a replacement; a pure destroy of it, an allowed address the
   plan doesn't replace, and every other delete are blocked. No flags: behaviour unchanged.
2. `installer/skill_scripts_test.go`: stub-`terraform` cases — no changes, update-only,
   unexpected delete blocked, allowed replacement passes with `-replace=` passed through, pure
   destroy of an allowed address blocked, allowed address not replaced blocked, allowed
   replacement plus another delete blocked, `--out` keeps the plan, unknown flag fails. Skip when
   `jq` is missing.
3. `provision-cyso/SKILL.md`: a deliberate-replacement path in Steps 6–7 (apply the `--out` plan),
   a Failure Handling row, version bump. `cicd-generate/SKILL.md`: a gotcha for gated workflows
   (pass the dispatch input through `--allow-replace`, apply the `--out` plan), version bump;
   templates unchanged.

*Phase B — Cyso facts (parallel with A)*

4. `cyso-reference.md` S3 backend block: `endpoints = { s3 = … }`, `use_path_style`,
   `skip_requesting_account_id`, `skip_s3_checksum`, `skip_credentials_validation`,
   `skip_region_validation`; Terraform >= 1.6; state isolation by key in a shared bucket, or a
   dedicated bucket for full separation.
5. `openstack-forcenew-attributes.md`: Nova's 65,535-byte base64 `user_data` limit and its
   empty-error-box symptom, a plan-time `precondition`, `base64gzip()` with cloud-init
   `encoding: gz+b64`, and `ignore_changes = [user_data]` requiring an explicit guarded
   `-replace=`. `provider-cyso/SKILL.md`: gotcha line, version bump.

*Phase C — platform-repository guide*

6. New `docs/wiki/workflows/platform-repositories.md` (app-agnostic, no downstream names): what a
   platform repository is and which slots it binds; layout (shared modules, per-machine roots,
   state, and Environments; validate-only PR checks under main-only branch policies); governance
   (gated dispatch with an exact confirm input, guarded `--allow-replace`/`--out`); day-two
   through workflows, sharing first-boot scripts, targeting from Terraform outputs; a generic
   machine-verification checklist; gotchas (ssh argv joining, `</dev/null` in loops, secrets on
   stdin, guard-then-act idempotency at the filesystem level, SSH keepalive, `user_data` limit,
   host keys changing on replacement).
7. Link it from `deployment-slots.md` and any wiki index.

*Phase D — cross-reference*

8. `smaqit.input-deployment` and `smaqit.new-greenfield-project` mode definitions: an app on a
   platform repository's machine uses `existing-k3s` (onboarded via `request-k3s-onboarding`) or
   `existing-shared`; the platform repository binds its own fillers; see the guide. Version
   bumps; descriptions within the 400-character cap.

*Phase E — verification and release*

9. `.smaqit/compendium.md`: extend the platform-repository entry with the guarded-replacement
   flow and the guide.
10. `make -C installer prepare test smoke-test`; `grep -rn force_path_style skills/` is empty;
    break the guard deliberately once to prove the tests fail. CHANGELOG via `task.complete`.

## Acceptance Criteria

- [x] `plan-guard.sh` accepts repeatable `--allow-replace <address>` and `--out <planfile>`,
      passes each address to `terraform plan -replace=`, allows only a true replacement of a
      listed address, blocks every other delete, a pure destroy of a listed address, and a listed
      address the plan doesn't replace; with no flags its behaviour is unchanged.
- [x] Stub-terraform tests cover every case in Implementation Step 2 and fail when the guard is
      deliberately broken.
- [x] `provision-cyso` documents the guarded replacement path and applying the `--out` plan;
      `cicd-generate` documents it for gated workflows; templates are unchanged.
- [x] `cyso-reference.md`'s backend block uses the current S3 backend syntax and flags, with no
      `force_path_style` anywhere in `skills/`; the `user_data` limit, gzip embedding, and
      explicit replacement are documented in `openstack-forcenew-attributes.md`.
- [x] `docs/wiki/workflows/platform-repositories.md` covers the pattern, slots, governance,
      day-two, the verification checklist, and every gotcha in Step 6, naming no downstream
      project; `deployment-slots.md` links to it.
- [x] `smaqit.input-deployment` and `smaqit.new-greenfield-project` cross-reference hosting an
      app on a platform repository's machine.
- [x] `make -C installer test` and `make -C installer smoke-test` pass; the compendium and
      `CHANGELOG.md` are updated.

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- Rescoped during `task.plan` after reading the downstream platform repository as it stands: provisioning and host-baseline live but rebuilt around k3s and a mesh network, verification never packaged, and VM tenancy (`tenants.yml`) retired on 2026-09-05 for k3s Namespace onboarding. The four canonical skills were dropped in favour of generic hand-off pieces.
- `plan-guard.sh`: repeatable `--allow-replace <address>` (passed to `terraform plan -replace=`; passes only a delete-plus-create of a listed address, in either action order) and `--out <planfile>` (resolved to an absolute path before changing directory; kept only when the guard passes). No flags keeps the old behaviour; usage errors exit 2.
- `installer/skill_scripts_test.go`: a stub `terraform` (records argv, writes a marker plan at `-out=`, answers `show -json` only for that plan) and four tests covering every planned case. A deliberate break (dropping the replacement check) made the pure-destroy case fail; restored.
- provision-cyso now guards with `--out tfplan` and applies `tfplan`, with a deliberate-replacement path and two failure rows; cicd-generate gained a gated-replacement gotcha.
- Cyso facts: Terraform 1.6+ S3 backend block proven against Cyso Object Storage, state-isolation guidance, the Nova `user_data` limit with a plan-time precondition and gzip embedding, and the `ignore_changes`/deliberate-replacement interplay.
- New `docs/wiki/workflows/platform-repositories.md`, linked from `deployment-slots.md` and the README; app-side cross-references in `input-deployment` and `new-greenfield-project`; compendium entry extended.
- `make -C installer prepare test smoke-test` pass; no downstream names in the diff; no `force_path_style` left in `skills/`.

**Decisions made:**
- No canonical machine skills and no tenancy skill: the platform repository's mechanisms changed wholesale within weeks and remain its own slot fillers (task 116 rule).
- Machine verification is a checklist in the guide rather than a skill, because the checks depend on the machine's class.
- provision-cyso's default flow now applies the saved, guarded plan instead of re-planning on apply.
- The cross-reference sentences in shipped skills are self-contained: `docs/wiki/` is not installed with the skills, so a skill cannot link to the guide.
- The per-bucket isolation note was limited to what is verified (a bucket as the unit of listing, lifecycle, and deletion); per-bucket credential scoping on Cyso was not verified and is not claimed.

**Blockers encountered:**
- A fresh worktree has no generated embed directories, so `go test` fails at setup until `make prepare` runs.

**Follow-up identified:**
- cicd-generate's `deploy.yml.full.template` still runs the guard and then a bare `terraform apply -auto-approve`, which re-plans; moving it to `--out`/`apply tfplan` changes generated workflows and deserves its own task.
- Downstream platform repository task 051 can use `--allow-replace` instead of writing its own guard.
- Task 114's `sync-machine` sets an optional `FIXED_IP` that the downstream no longer reads; documented as optional in the guide.

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh`, `SKILL.md` | Modify |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md` | Modify (gotcha) |
| `skills/smaqit.infrastructure-provider-cyso/{SKILL.md,references/cyso-reference.md,references/openstack-forcenew-attributes.md}` | Modify |
| `skills/smaqit.input-deployment/SKILL.md`, `skills/smaqit.new-greenfield-project/SKILL.md` | Modify (cross-reference) |
| `installer/skill_scripts_test.go` | Modify |
| `docs/wiki/workflows/platform-repositories.md` | Create |
| `docs/wiki/workflows/deployment-slots.md`, `.smaqit/compendium.md`, `CHANGELOG.md` | Modify |

## Notes

Source material: a downstream project's infrastructure repo (repo layout, `scripts/reconcile/`, the three
workflows) and its `.smaqit/reports/provisioning-evidence-2026-08-24.md`. Together with task 114,
this is the "smaqit framework update for the machine-repo pattern" that infrastructure task 001
deferred to this repository. Task 114 should land first (this task consumes its corrected facts
and registration/config flows).
- **Re-scoped by task 125 (2026-10-11):** with deployment slots, a platform repository owns its machine fillers (provision, verify, host baseline, replace). These four skills become optional canonical default fillers and the slot contracts they must meet, not mandatory canonical machine skills.
- **Rescoped 2026-10-12 (`task.plan`):** discovery against the downstream repository found
  provisioning and host-baseline live but changed (k3s, mesh access, machine classes), no packaged
  verification, and tenancy replaced by k3s Namespace onboarding on 2026-09-05. Its own task 051
  binds its project skills into smaqit slots; it can use `--allow-replace` once this ships.
  Task 114's `sync-machine` still sets an optional `FIXED_IP`; the downstream now reads the IP
  from Terraform outputs instead — the guide notes the variable as optional.
- **Refreshed after task 125 implementation (2026-10-11):** the shipped slots for these four skills are `provision`, `verify`, `host-bootstrap` ("host baseline"), and the day-two `replace`; any default filler added here must declare them in `metadata.smaqit-slots` and be added to the catalogue in `installer/slots.go` (a Go test enforces the declaration).
