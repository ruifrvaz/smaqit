# Deploy Target Resolution — Branch the Flow for Existing / Shared VMs

**Status:** In Progress
**Mode:** Assisted
**Created:** 2026-07-20
**Started:** 2026-07-20

## Description

`smaqit.new-greenfield-project`'s Phase 4/5 flow, and every skill it invokes
(`smaqit.infrastructure-vault-loader`, `smaqit.infrastructure-provision-cyso`,
`smaqit.infrastructure-vm-bootstrap`, `smaqit.infrastructure-cicd-generate`,
`smaqit.infrastructure-repo-config`, `smaqit.infrastructure-deploy-rsync`), assumes exactly one
scenario: this project provisions its own brand-new VM via Terraform. That assumption was never
wrong until now — every project built with smaqit so far has owned its own infrastructure.

It breaks for a real, already-encountered case: deploying a *second* project onto a VM a *different*
project already owns and manages via its own Terraform state (co-hosted apps on one VM, e.g. two
sites sharing an IP behind nginx). Run naively, the flow:

1. Tries to provision a second VM nobody asked for, via `smaqit.infrastructure-provision-cyso`.
2. Generates a brand-new SSH keypair (`smaqit.infrastructure-vault-loader`'s default when one is
   missing) that the existing, shared VM has never seen and will refuse to authenticate.
3. Hard-fails in `smaqit.infrastructure-repo-config`, which reads all four Vault paths
   (`ssh`, `tfstate`, `cyso`, `github`) unconditionally — `tfstate`/`cyso` were never populated
   because this project was never meant to provision anything.

This was confirmed by directly reading the deployment agent, `smaqit.new-greenfield-project`, and
the five infrastructure skills above, in the context of a real deploy (`fashion-app-poc`, plus a
second project — `<tested-deployment>` — that needs to land on the same VM). None of this is a bug in any
one skill; it's a decision the flow currently never makes explicitly, because it's never had to.

A companion flowchart mapping every branch point, which skill each touches, and what changes on
each path is at [`084-flowchart.md`](084-flowchart.md) in this same directory — read it before
starting implementation. (An interactive rendered copy also exists at
https://claude.ai/code/artifact/a035b9bd-b1e4-4750-a2c7-4a9326b4f235, but the in-repo file is the
source of truth.)

## Design Decisions

- **New concept: `provisioning_mode`**, one of `provision` (default, current behavior unchanged) /
  `existing-owned` (redeploying to a VM this project's own Terraform state already manages —
  already works today via `plan-guard.sh`'s idempotent no-op, just needs to be a named, documented
  path rather than an accident of how Terraform happens to behave) / `existing-shared` (targeting a
  VM a *different* project owns — the actually-broken case).
- **Resolution lives in `smaqit.input-deployment`**, not a new skill. That skill already elicits a
  "Deployment Target" parameter when ambiguous; `provisioning_mode` is a natural extension of the
  same parameter, not a new concept requiring its own skill (per `CONTRIBUTING.md`'s "avoid
  introducing new concepts unless needed"). Default `provision` if nothing in session context
  suggests otherwise; elicit when the user's request mentions an existing/shared VM, co-hosting, or
  names infrastructure another project already owns.
- **`smaqit.new-greenfield-project` Phase 4/5 branch on the resolved mode** — this is the one place
  that already orchestrates the full lifecycle; branching here (rather than scattering conditionals
  across five independently-invocable skills) keeps the decision centralized and auditable.
- **Defense in depth, not instead of the gate**: `smaqit.infrastructure-provision-cyso` and
  `smaqit.infrastructure-repo-config` also each gain a narrow, local guard (see steps below) so a
  direct/manual invocation of either skill — bypassing the orchestrator — doesn't silently do the
  wrong thing.
- **`existing-shared` mode never provisions Terraform for the deploying project.** Two Terraform
  states with opinions about the same VM is an explicit non-goal; if a future case genuinely needs
  that, it gets its own task rather than being guessed at here.
- **SSH access for `existing-shared` is a documented manual step, not automated.** Two options exist
  (reuse the owning project's key by copying it across Vault namespaces, or generate a new key and
  manually append its public half to the shared VM's `authorized_keys`) — both require the operator
  to already have access to the owning project's key once, which cannot be automated without
  assuming trust between two otherwise-independent projects' Vault instances. Document both; don't
  pick one as "the" automated path.
- **`VM_HOST` is a GitHub Actions *variable*, not a secret, in every `provisioning_mode`** — it's
  just an IP/hostname, not sensitive, and unlike a secret it can be read back via `gh variable get`.
  That read-back is what lets `smaqit.infrastructure-provision-cyso`'s ownership guard (Step 4)
  query the currently-declared target directly, instead of depending on the caller having already
  exported it into the environment. For `existing-shared` mode specifically, the variable is set
  manually (`gh variable set VM_HOST`), never derived from a Terraform output — there is no
  Terraform run on the deploying project's side to produce one.

## Implementation Steps

1. **`smaqit.input-deployment`** — add a `Provisioning Mode` parameter alongside the existing
   `Deployment Target` one. Default `provision`. Elicit with: "Is this a new VM, a redeploy of your
   own existing VM, or targeting a VM another project already manages?" when session context doesn't
   make it unambiguous. Document the three values and what each means downstream.

2. **`smaqit.new-greenfield-project`** — in Phase 4 and Phase 5, branch on `provisioning_mode`:
   - `provision`: unchanged (steps as written today).
   - `existing-owned`: same skill sequence, but document explicitly that `provision-cyso`'s
     `terraform apply` is expected to no-op (gated by `plan-guard.sh`), not skip the step outright —
     this path already works, it just needs to stop being an undocumented accident.
   - `existing-shared`: skip `smaqit.infrastructure-provision-cyso` entirely. `vault-loader` only
     needs to populate `ssh` and `github` for this project's slug (see step 3). `VM_HOST` is set via
     `gh variable set`, not read from Terraform output. `cicd-generate` runs in deploy-only mode (see
     step 6). `repo-config` runs in restricted mode (see step 5).

3. **`smaqit.infrastructure-vault-loader`** — `scripts/load-credentials.sh` needs a mode-aware branch:
   in `existing-shared` mode, skip prompting for `cyso`/`tfstate` entirely (they're never needed), and
   for `ssh`, offer a third option beyond "already populated" / "auto-generate" — "copy from another
   project's Vault namespace" (prompt for the source project-slug, read its `secret/<slug>/ssh`
   fields, write them under the new slug). Document that the alternative (generate fresh + manually
   append to `authorized_keys`) is equally valid and doesn't require this script's help.

4. **`smaqit.infrastructure-provision-cyso`** — add a pre-flight guard as step 0: if `VM_HOST` (or an
   explicitly supplied target IP) is already set but this project's Terraform state has no matching
   `openstack_compute_instance_v2` resource, stop with a clear message pointing at the
   `existing-shared` path instead of attempting to provision a second VM. This is the defense-in-depth
   guard for a direct invocation that bypasses `new-greenfield-project`'s branching.

5. **`smaqit.infrastructure-repo-config`** — steps 4 and 5 (`tfstate`, `cyso` secrets) must check
   whether the Vault path has any data before attempting to read a field from it, and skip that
   step cleanly (not hard-fail) if absent — rather than assuming, as it does today, that all four
   paths always exist.

6. **`smaqit.infrastructure-cicd-generate`** — add a `deploy-only` generation mode: `deploy.yml` gets
   a single `deploy` job (no `provision` job, no Terraform step at all), `provision.yml` is not
   generated. Document when to use it (`provisioning_mode: existing-shared`).

7. **`smaqit.infrastructure-deploy-rsync`** and the nginx vhost step it (or the generated workflow)
   writes — document the Decision 5 branch explicitly: first site on a VM claims `default_server`;
   any subsequent co-hosted site's vhost must be name-based only. This is already the correct design
   from the fashion-app-poc incident remediation; it just needs to be stated here too, since this is
   the skill that actually writes the vhost file.

8. **Update the flowchart artifact** (linked above) if the design changes materially during
   implementation review — it's the canonical reference for this task, not a one-time sketch.

9. Bump `metadata.version` on every skill touched, following existing semver-ish convention
   (patch-level bump for doc-only changes, minor for new steps/parameters).

## Known Issues Triage
**Triaged:** 2026-07-20
**Tools searched:** terraform-provider-openstack, hashicorp/vault
**Result:** Clear

### Blocking Issues
- None

### Advisory Issues
- None — Vault keyword search (`namespace copy secret`) returned 11 loosely-matched issues, none pertaining to reading a secret path in one namespace/slug and writing it under another; none labeled `bug`/`regression`; none relevant to the cross-project SSH-key-copy design in Step 3.

### Historical (Closed)
- None

### Unresolvable Tools
- GitHub CLI (`gh`), GitHub Actions, nginx — no task-specific bug search performed; these are used per their documented, stable CLI/workflow-syntax behavior only, not a version-specific feature this task depends on.

## Acceptance Criteria

- [x] `smaqit.input-deployment` resolves and surfaces a `provisioning_mode` value (`provision` /
      `existing-owned` / `existing-shared`), defaulting safely and eliciting only when genuinely
      ambiguous
- [x] `smaqit.new-greenfield-project` Phase 4/5 steps branch explicitly on `provisioning_mode`, with
      all three paths documented (not just the default)
- [x] `existing-shared` mode never invokes `smaqit.infrastructure-provision-cyso`
- [x] `smaqit.infrastructure-vault-loader` supports populating only `ssh` + `github` for a project
      slug, without prompting for `cyso`/`tfstate`, and offers a documented path to reuse an SSH
      keypair across two projects' Vault namespaces
- [x] `smaqit.infrastructure-repo-config` no longer hard-fails when `tfstate`/`cyso` Vault paths are
      absent — skips those steps cleanly and reports what was skipped and why
- [x] `smaqit.infrastructure-provision-cyso` has a pre-flight guard against provisioning a second VM
      when one is already declared for the target but not owned by this project's state
- [x] `smaqit.infrastructure-cicd-generate` supports a `deploy-only` mode producing a single-job
      `deploy.yml` with no Terraform/provision step
- [x] The nginx `default_server`-vs-name-based-vhost rule is documented in whichever skill actually
      writes the vhost file, not only in a downstream project's own task notes
- [ ] All acceptance criteria above are exercised, at minimum, by walking through the
      `<tested-deployment>`-onto-`fashion-app-poc`'s-VM scenario end-to-end in a real session and confirming
      each step does what this task says it should — **not done in this session**: requires a live
      walkthrough in those projects' own working environments, not available here

## Findings

**Implementation approach (interim — task not yet complete, Assisted mode):**
- Followed the 9 Implementation Steps in order. Added one new script not explicitly named in the
  task (`smaqit.infrastructure-provision-cyso/scripts/ownership-guard.sh`) to implement Step 4's
  pre-flight guard as a deterministic check (mirroring the existing `plan-guard.sh` pattern)
  rather than a prose instruction.
- `VM_HOST` storage went through three passes before landing: (1) initially implemented as a
  GitHub Actions *variable* only for `existing-shared`, reading the Design Decisions' literal
  `gh variable set VM_HOST` wording; (2) reverted to a **secret** in every mode after the user
  flagged that wording as a miscommunication; (3) on reassessment, moved to a **variable** in
  every mode (not just `existing-shared`) — `VM_HOST` is just an IP/hostname, not sensitive, and
  unlike a secret it can be read back via `gh variable get`. That read-back directly simplifies
  `smaqit.infrastructure-provision-cyso/scripts/ownership-guard.sh`: it can query the declared
  target itself rather than depending on the caller (operator or workflow `env:` block) having
  already exported it. Final state applies uniformly across `smaqit.infrastructure-cicd-generate`,
  `smaqit.infrastructure-repo-config`, `smaqit.new-greenfield-project`,
  `smaqit.infrastructure-provision-cyso` (SKILL.md + `ownership-guard.sh`), and
  `smaqit.infrastructure-hook-post-deploy-stamp` (collateral wording fix — not in the original
  Files to Create/Modify table, but referenced `VM_HOST` as a secret and would have been
  inconsistent otherwise).

**Decisions made:**
- All branching is documented inline in each affected SKILL.md (`→ existing-owned:` /
  `→ existing-shared:` callouts under the relevant step) rather than duplicating full step
  sequences per mode — keeps the default (`provision`) path unchanged and readable.

**Blockers encountered:**
- None.

**Follow-up identified:**
- Acceptance criterion "exercise end-to-end via the `<tested-deployment>`-onto-`fashion-app-poc` scenario"
  is unmet — it requires a live session in those projects' own working environments, not available
  in this repo's session. Flagging for the user to run separately before invoking `/task.complete 084`.

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
| `skills/smaqit.input-deployment/SKILL.md` | Modify — add `Provisioning Mode` parameter |
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify — branch Phase 4/5 on `provisioning_mode` |
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Modify — mode-aware credential population |
| `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` | Modify — skip cyso/tfstate prompts in `existing-shared` mode; add cross-project SSH key copy path |
| `skills/smaqit.infrastructure-provision-cyso/SKILL.md` | Modify — add pre-flight ownership guard |
| `skills/smaqit.infrastructure-provision-cyso/scripts/ownership-guard.sh` | New — deterministic Step 0 guard; resolves declared `VM_HOST` via arg/env/`gh variable get` and checks it against Terraform state |
| `skills/smaqit.infrastructure-repo-config/SKILL.md` | Modify — skip absent Vault paths cleanly instead of hard-failing; `VM_HOST` set via `gh variable set` |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md` | Modify — add `deploy-only` generation mode; `VM_HOST` read via `${{ vars.VM_HOST }}` |
| `skills/smaqit.infrastructure-deploy-rsync/SKILL.md` | Modify — document `default_server` vs name-based vhost rule |
| `skills/smaqit.infrastructure-hook-post-deploy-stamp/SKILL.md` | Modify (collateral) — wording fix: `VM_HOST` is a variable, not a secret |

## Notes

- Originates from hardening and deploying `fashion-app-poc`'s infrastructure, then assessing whether
  the same tooling would work for redeploying a second project (`<tested-deployment>`) onto the same VM.
  It would not, without this task. Optional background, if that repo happens to be available in the
  working environment: its own task history has the full incident and fix history that led here —
  not required to execute this task, everything actionable is captured in Description and
  Implementation Steps above.
- [`084-flowchart.md`](084-flowchart.md) is the design reference — read it before starting
  implementation; it has the full branch-by-branch reasoning this description only summarizes.
- Explicitly out of scope: automating cross-project trust for SSH key sharing, and supporting two
  independent Terraform states both managing resources on the same VM. Both are flagged as non-goals
  in Design Decisions, not deferred sub-tasks.
