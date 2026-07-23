# Machine Credential Namespace and App Bootstrap

**Status:** In Progress
**Mode:** Assisted
**Created:** 2026-07-23
**Started:** 2026-07-23

## Description

Task `084` gave the deployment flow a way to *decide* whether a target VM is new, owned, or shared —
but it didn't change what gets stored in Vault to make the `existing-shared` path actually work
smoothly. Vault's credential model today has exactly one namespacing axis:
`secret/<project-slug>/*`. A machine (a provisioned VM) has no representation of its own — only
whichever project happened to provision it has a `secret/<project-slug>/ssh` entry for the keypair
Terraform installed on it.

This was confirmed against real state, not designed in the abstract: on a shared local Vault instance
(one Vault, many projects' credentials, as documented in `smaqit.infrastructure-vault-loader`), two
real projects' handling of the same VM was inspected directly:

- One project's SSH key was never in this Vault at all — it had been sourced from a different device
  entirely and placed directly into GitHub Secrets, bypassing Vault. Vault has no record of it, and
  GitHub Secrets are structurally write-only (no API or CLI can read a secret's value back once set)
  — there is no way to recover that key through Vault or GitHub, only through wherever it was
  originally issued.
- A second project targeting the *same* VM had generated its own, disconnected SSH keypair, stored
  under its own project slug. It was never authorized on the VM — attempting to connect with it
  failed with `Permission denied (publickey)`.

Both symptoms trace to the same root cause: nothing in the credential model represents "this machine"
as an entity with its own identity and its own base credential that other things can bootstrap against.
Every project independently invents or sources its own disconnected keypair, with no shared anchor.

## Design Decisions

- **New Vault namespace: `secret/machines/<machine-slug>/*`**, parallel to the existing flat
  `secret/<project-slug>/*` scheme still used by projects that haven't migrated (`hello-mario`,
  `iodis-crm-poc`, `assistente-escolas-poc` — migration explicitly out of scope, see Notes). Reserved:
  no app-slug may be `machines`. Holds everything that belongs to *whoever provisions the machine*,
  not to any one app running on it:
  - `secret/machines/<machine-slug>/base-ssh` — `private_key`, `public_key`: the credential Terraform
    actually installs on the VM at provision time (via `key_pair`/`user_data`). Used **only** to
    bootstrap an app's own access or to rotate itself — never used for routine, day-to-day deploys by
    anyone, including the project that originally provisioned the machine.
  - `secret/machines/<machine-slug>/cyso` — `app_credential_id`, `app_credential_secret`: the cloud
    provider API credential Terraform uses to provision/manage this machine. Machine-scoped, not
    app-scoped — provisioning is a property of the machine, not of any individual app on it.
  - `secret/machines/<machine-slug>/tfstate` — `access_key`, `secret_key`: credential for the
    Terraform remote-state backend tracking this machine. Same reasoning as `cyso`.
  - `secret/machines/<machine-slug>/metadata` — non-secret: `host` (IP), `provider`, `owner_project`
    (which project's Terraform state actually provisions this machine).
  - **Confirmed against real state, not designed in the abstract**: this exact shape already exists
    at `secret/machines/magnificah-test/*` in the shared local Vault instance as of 2026-07-23 — the
    design here documents and completes that prototype rather than inventing a new one.
- **Every app always gets its own distinct keypair — no exceptions, including the owning project.**
  The machine's base credential is a bootstrap-only mechanism, never a shared day-to-day credential.
  This costs one extra keypair + one extra bootstrap step even in the simple single-app-per-machine
  case, in exchange for a uniform model with no special-casing: rotating the machine's base credential
  never has to touch any app's ongoing access, and one app's key compromise never implicates another
  app sharing the machine.
  - **New Vault namespace: `secret/apps/<app-slug>/*`**, replacing the flat `secret/<project-slug>/*`
    root for anything that belongs to a specific app rather than the machine it runs on:
    `secret/apps/<app-slug>/ssh` (the app's own keypair — always populated via the bootstrap flow
    below, never by copying another app's key material or by Terraform installing it directly),
    `secret/apps/<app-slug>/github` (unchanged in meaning from today's `secret/<project-slug>/github`,
    just relocated), and `secret/apps/<app-slug>/machine` (new, small — records which `machine-slug`
    this app is bootstrapped against, so future sessions don't have to re-derive it).
  - **Confirmed against real state**: `secret/apps/areaoffice-poc/ssh` already exists in the same
    shared Vault instance, matching this shape.
  - Terminology: earlier drafts of this task used `project-slug` throughout. The real namespace uses
    `app-slug` for anything under `secret/apps/*` and `machine-slug` for anything under
    `secret/machines/*` — Implementation Steps and Acceptance Criteria below use `app-slug`
    accordingly.
- **Bootstrap logic lives in `smaqit.infrastructure-vault-loader`**, not a new skill. It already owns
  keypair generation/loading; this adds a machine-aware mode rather than introducing a new concept,
  consistent with task 084's "extend, don't invent" bias.
- **Bootstrap procedure**, run once per app-to-machine pairing (idempotent — a no-op if
  `secret/apps/<app-slug>/ssh` is already populated *and* reachable, verified via a lightweight
  `ssh -o BatchMode=yes` connectivity check, not just presence in Vault):
  1. Resolve the target `machine-slug`. If `secret/machines/<machine-slug>/base-ssh` doesn't exist yet,
     this is a fresh machine being registered for the first time (the `provisioning_mode: provision`
     case from task 084) — generate the base keypair here, store it along with `cyso`/`tfstate`/
     `metadata`, and continue to step 2 as normal; the provisioning project is not exempt from getting
     its own app-specific key.
  2. Generate a new, distinct keypair for `<app-slug>`.
  3. Fetch the machine's `base-ssh` private key into a temp file (never displayed, never logged),
     use it to SSH in, append the new app keypair's public half to `~/.ssh/authorized_keys`, then
     discard the temp file.
  4. Store the app's new keypair at `secret/apps/<app-slug>/ssh`; record `secret/apps/<app-slug>/machine`.
  5. Verify the new keypair actually authenticates before reporting success.
- **Rotating a machine's base credential is independent of app access.** Since apps only ever touch
  the base credential once (at bootstrap), rotating it (generate new base keypair, use the *old* one
  one last time to install the new public key, retire the old one) never requires touching any app's
  already-bootstrapped `secret/apps/<app-slug>/ssh`. `cyso`/`tfstate` rotation for a machine reuses the
  existing per-path rotation `rotate-credential.sh` already implements, just re-targeted at
  `secret/machines/<machine-slug>/*` instead of `secret/<project-slug>/*` — only `base-ssh` needs the
  bespoke generate/install-with-old-key/retire sequence, since it's the one credential type that must
  install itself onto the remote machine rather than simply being replaced in Vault. This means
  `rotate-credential.sh` needs a real restructuring, not just a new case-statement branch: it must
  resolve its target slug and path root by credential type — `ssh`/`github` under
  `secret/apps/<app-slug>/*` (app-slug), `base-ssh`/`cyso`/`tfstate` under
  `secret/machines/<machine-slug>/*` (machine-slug) — replacing today's single-root assumption
  (`FULL_PATH=secret/${PROJECT_SLUG}/${CREDENTIAL_PATH}` for all four legacy types).
- **Out of scope:** de-provisioning an app's access when it's decommissioned (removing its public key
  from `authorized_keys`), multi-machine apps, and migrating existing `secret/<project-slug>/*`-scoped
  projects (`hello-mario`, `iodis-crm-poc`, `assistente-escolas-poc`) onto the new `apps/`/`machines/`
  scheme. All three are real future needs, not solved here — flag as follow-up, don't guess at a
  design now.

## Implementation Steps

1. **`smaqit.infrastructure-vault-loader/SKILL.md`** — update the documented Vault path convention
   (currently `secret/<project-slug>/<type>` for `cyso`/`ssh`/`tfstate`/`github`, SKILL.md:31–40) to
   describe the two-namespace scheme: `secret/machines/<machine-slug>/{base-ssh,cyso,tfstate,metadata}`
   and `secret/apps/<app-slug>/{ssh,github,machine}`. Note the old flat scheme remains in place,
   unmigrated, for `hello-mario`/`iodis-crm-poc`/`assistente-escolas-poc`. Add the bootstrap procedure
   as a new documented flow, distinct from the existing "every session" (`load-credentials.sh`) and
   "one-time setup" (`install-vault.sh`/`setup-vault.sh`) flows.
2. **`smaqit.infrastructure-vault-loader/scripts/bootstrap-app-to-machine.sh <app-slug> <machine-slug>`**
   (new) — implement the five-step procedure above. Match `load-credentials.sh`'s existing idioms
   exactly: hidden/piped credential handling (`read_secret()`-style helper), `ssh-keygen -t ed25519 -f
   ... -N "" -q` inside a `mktemp -d` always `rm -rf`'d, private keys written via `@file` syntax
   (never `$(cat ...)` — see the documented libcrypto Gotcha), public keys via `$(cat ... | tr -d
   '\n')`, `vault kv put ... > /dev/null`, `set -euo pipefail`.
3. **`smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh`** — restructure per the Design
   Decisions above: today it's `rotate-credential.sh <path>` with a hardcoded `cyso|ssh|tfstate|github`
   case statement and a single `FULL_PATH=secret/${PROJECT_SLUG}/${CREDENTIAL_PATH}`. Split into two
   roots by credential type (`ssh`/`github` → `secret/apps/<app-slug>/*`; `base-ssh`/`cyso`/`tfstate` →
   `secret/machines/<machine-slug>/*`), and give `base-ssh` its own generate/install-with-old-key/retire
   sequence instead of delegating to `load-credentials.sh` (which has no machine-scoped keygen path).
4. **`smaqit.infrastructure-vault-loader/scripts/load-credentials.sh`** — newly identified scope (not
   in the original draft): this is the "every session" script that populates credentials today at
   `secret/<project-slug>/*`. For projects on the new scheme, it must populate `secret/apps/<app-slug>/
   {ssh,github}` instead, and stop prompting for `cyso`/`tfstate` per-app entirely — those are now
   populated once at machine registration (Step 5 below), not per app per session.
5. **`smaqit.infrastructure-provision-cyso/SKILL.md`** — on a fresh provision (`provisioning_mode:
   provision`), insert a new step between the existing step 7 (`terraform apply`, lines 105–119) and
   step 8 (verify SSH access, which today only reads the already-loaded project-slug `ssh` secret):
   register the new machine (write `secret/machines/<machine-slug>/{base-ssh,cyso,tfstate,metadata}`),
   then invoke `bootstrap-app-to-machine.sh` for the provisioning project itself — it does not get a
   shortcut around getting its own app-specific key.
6. **`smaqit.new-greenfield-project/SKILL.md`** — replace the two exact spots carrying the "reuse
   owner's key across Vault namespaces" guidance: line 36 (Phase 4 precondition checklist) and line 104
   (Phase 4 step 2 `existing-shared` callout). Both become a single instruction to invoke
   `bootstrap-app-to-machine.sh` against the target machine's registered base credential.
7. **`smaqit.infrastructure-repo-config`** — this needs an actual change, not just confirmation as
   originally scoped: its `ssh`/`github` reads must move from `secret/<project-slug>/*` to
   `secret/apps/<app-slug>/*` for projects on the new scheme, while its existing "skip absent Vault
   paths cleanly" behavior (from task 084) continues to correctly leave `secret/machines/*` alone —
   that namespace is never something a project's own `repo-config` run should read or sync to GitHub
   Secrets directly.
8. Update `084-flowchart.md` (or add a small companion diagram here) showing the bootstrap sequence,
   if the existing decision-tree diagram doesn't already make it legible on its own.

## Known Issues Triage
**Triaged:** 2026-07-23
**Tools searched:** hashicorp/vault, openssh/openssh-portable
**Result:** Advisory

### Blocking Issues
None

### Advisory Issues
- [#31669 Creating a secret in kv v2 with a forwards slash leads to the slash escaped as %2f](https://github.com/hashicorp/vault/issues/31669) — `hashicorp/vault` — opened 2025-12-04 — `ui`, `bug` — UI-only edge case about a literal slash *within* a single path segment; does not apply to this task's use of `/` as a standard multi-segment KV v2 path separator (`secret/machines/<slug>/base-ssh`) via the `vault` CLI, which all existing `smaqit.infrastructure-vault-loader` scripts already rely on without issue.

### Historical (Closed)
None

### Unresolvable Tools
None

## Acceptance Criteria

- [x] `secret/machines/<machine-slug>/{base-ssh,cyso,tfstate,metadata}` exist as a documented, reserved
      Vault namespace, distinct from `secret/apps/<app-slug>/*` — matching the real prototype already
      present in the shared Vault instance (`secret/machines/magnificah-test/*`)
- [x] `bootstrap-app-to-machine.sh` exists, is idempotent, and never exposes either the machine's base
      private key or the app's newly-generated private key outside of piped/file-based handling
- [ ] A fresh VM provision registers itself as a machine and bootstraps the provisioning project's own
      app-specific key through the same path a later, different app would use — no special-casing
- [ ] A second app targeting an already-registered machine can run the bootstrap script and end up
      with a working, independently-revocable SSH key at `secret/apps/<app-slug>/ssh`, verified by an
      actual successful connection (not just "no error"), without ever touching the first app's key
      material
- [ ] `rotate-credential.sh` correctly targets `secret/apps/<app-slug>/*` for `ssh`/`github` and
      `secret/machines/<machine-slug>/*` for `base-ssh`/`cyso`/`tfstate`, and rotating a machine's
      `base-ssh` never requires any already-bootstrapped app's key to change
- [ ] `load-credentials.sh` populates `secret/apps/<app-slug>/{ssh,github}` for new-scheme projects
      each session, without prompting for `cyso`/`tfstate` per app
- [ ] Walked through end-to-end against the real scenario that motivated this: a second project
      (already has a disconnected, unauthorized keypair as of this writing) successfully bootstraps
      onto the machine the first project provisioned, and can connect

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
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Modify — document `apps/`/`machines/` namespace + bootstrap flow, supersedes documented `<project-slug>/*` convention for new projects |
| `skills/smaqit.infrastructure-vault-loader/scripts/bootstrap-app-to-machine.sh` | Create |
| `skills/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh` | Modify — split into `apps/<app-slug>/*` (ssh/github) and `machines/<machine-slug>/*` (base-ssh/cyso/tfstate) roots |
| `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` | Modify — populate `apps/<app-slug>/{ssh,github}` each session; drop per-app cyso/tfstate prompts |
| `skills/smaqit.infrastructure-provision-cyso/SKILL.md` | Modify — register machine + self-bootstrap on fresh provision |
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify — `existing-shared` branch (lines 36, 104) uses bootstrap script, not cross-namespace key copying |
| `skills/smaqit.infrastructure-repo-config/SKILL.md` | Modify — read `apps/<app-slug>/{ssh,github}` instead of `<project-slug>/{ssh,github}` |
| `.smaqit/tasks/084-flowchart.md` | Modify (optional) — add bootstrap sequence if needed for legibility |

## Notes

- Builds directly on task `084` (Deploy Target Resolution — Branch the Flow for Existing / Shared VMs),
  which is complete: it gave the flow `provisioning_mode` (`provision`/`existing-owned`/
  `existing-shared`) and the `ownership-guard.sh` defense-in-depth check in `provision-cyso`. This task
  replaces 084's original "reuse owner's key" / "generate own key, manually append" options for
  `existing-shared` SSH access with a single, automated, principled mechanism — read 084 first for why
  that decision point exists at all, but everything actionable for *this* task is self-contained above.
- Surfaced by direct inspection of real Vault state (a shared, multi-project local Vault instance) and
  a real failed SSH connection attempt — not designed speculatively.
