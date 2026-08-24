---
status: Not Started
created: "2026-08-24"
---

# New Machine-Monorepo Infrastructure Skills: Provision, Verify, Day-2 Baseline, Tenant Reconcile

## Description

Companion to task 114. `Magnificah/infrastructure` task 001 (2026-08-22 → 2026-08-24, PRs #1–#7)
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

## Design Decisions

- Four separate skills rather than one mega-skill, matching the existing family's granularity
  (provision / verify / day-2 / reconcile map to distinct invocation moments).
- Reference implementation is `Magnificah/infrastructure` (modules/, machines/magnificah-test-01/,
  scripts/reconcile/, .github/workflows/{provision,reconcile-tenants,host-baseline}.yml) — the
  skills generalize it (machine-agnostic, provider constants parameterized where sensible) rather
  than embedding Magnificah specifics; the pattern's rule that no application name may appear in
  machinery applies to the skills too.
- [TBD at task start: whether provision-machine stays Cyso/OpenStack-specific (matching
  provision-cyso's precedent) or abstracts the provider; whether `smaqit.new-greenfield-project`
  gains an `existing-shared` cross-reference pointing app projects at machine-monorepo-managed
  hosts.]
- Depends on task 114 for: corrected provider facts (skill 1 consumes them), machine-scoped
  GitHub Environment configuration (skill 1's credential step), and machine registration in
  Vault (precondition for everything).

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `smaqit.infrastructure-provision-machine` exists covering layout, per-machine state-key
      isolation, machine-scoped environment credentials, plan-on-PR + gated confirm-input apply,
      fixed-IP recording, and ForceNew/user_data safety.
- [ ] `smaqit.infrastructure-machine-verify` exists producing a PASS/FAIL report over SSH,
      cloud-init, mount (UUID-pinned + write test), services, API-verified security group,
      reboot survival, and zero-drift plan; documents the stale-known_hosts gotcha.
- [ ] `smaqit.infrastructure-host-baseline` exists encoding the day-2 workflow pattern
      (Terraform-output-sourced targeting, idempotent script streaming, machine-scoped
      environment) with an explicit no-ad-hoc-SSH stance.
- [ ] `smaqit.infrastructure-tenant-reconcile` exists covering the tenants.yml schema,
      fail-closed validation, idempotent grants, DNS ledger semantics, filesystem-level
      idempotency, and both the shlex-quoting and unconditional-touch gotchas.
- [ ] All four skills contain no application-specific names or logic (tenants as opaque data
      only), matching the pattern's own app-agnostic rule.
- [ ] Cross-references are in place: provision-machine ↔ provider-cyso (facts), ↔ task 114's
      environment-config work, tenant-reconcile ↔ host-baseline (grants assume a baselined host).

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
| `skills/smaqit.infrastructure-provision-machine/SKILL.md` | Create |
| `skills/smaqit.infrastructure-machine-verify/SKILL.md` | Create |
| `skills/smaqit.infrastructure-host-baseline/SKILL.md` | Create |
| `skills/smaqit.infrastructure-tenant-reconcile/SKILL.md` | Create |
| `skills/smaqit.new-greenfield-project/SKILL.md` | Possibly modify (cross-reference) |

## Notes

Source material: `Magnificah/infrastructure` (repo layout, `scripts/reconcile/`, the three
workflows) and its `.smaqit/reports/provisioning-evidence-2026-08-24.md`. Together with task 114,
this is the "smaqit framework update for the machine-repo pattern" that infrastructure task 001
deferred to this repository. Task 114 should land first (this task consumes its corrected facts
and registration/config flows).
