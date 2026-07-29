# Add `existing-unmanaged` Provisioning Mode — Dedicated VM, Never Terraform-Managed

**Status:** Completed
**Created:** 2026-07-29
**Completed:** 2026-07-29

## Description

Task 084 introduced `provisioning_mode` (`provision` / `existing-owned` / `existing-shared`) so `smaqit.new-greenfield-project` and `smaqit.feature-new` could branch correctly depending on who owns and manages the target VM's Terraform state. All three values assume Terraform governs the VM somewhere — either this project's own state (`provision`/`existing-owned`) or a different project's (`existing-shared`, co-hosting).

A fourth, real case doesn't fit any of them: a VM **dedicated solely to one project** (not co-hosted with anything else) that was provisioned out-of-band (manually, or via a different provider's own console/tooling) and will **never** be Terraform-managed, by design. Discovered on a downstream project's own deploy-to-a-second-provider task: a second live target (a plain VPS, no Terraform involvement now or ever) needs to sit alongside that project's existing Cyso/Terraform deployment.

Run naively today, this would either misuse `existing-shared` (wrong semantics — implies *another project* owns the VM, and pulls in co-hosting-specific behavior that doesn't apply to a dedicated VM) or force a `provision`-mode Terraform state to be written for a VM Terraform will never actually apply against. Neither is correct. This was confirmed by directly reading `smaqit.input-deployment`, `smaqit.new-greenfield-project`, `smaqit.feature-new`, `smaqit.infrastructure-provision-cyso` (including `ownership-guard.sh`), `smaqit.infrastructure-repo-config` (including `sync-secrets.sh`), `smaqit.infrastructure-cicd-generate` (including its `deploy-only` template assets), and `smaqit.infrastructure-vault-loader` (including `bootstrap-app-to-machine.sh`) — an exhaustive line-by-line inventory of every `existing-shared` branch point exists in this task's originating session and is summarized in Implementation Steps below.

## Design Decisions

- **New value: `existing-unmanaged`**, alongside the existing three, resolved in `smaqit.input-deployment` exactly as the other three are — no new skill. This mirrors task 084's own precedent ("resolution lives in `smaqit.input-deployment`, not a new skill... per `CONTRIBUTING.md`'s 'avoid introducing new concepts unless needed'"), and was chosen over building a separate, redundant skill (the alternative considered and rejected on the downstream project) specifically because it slots into an axis the framework already has.
- **Mechanically ~90% identical to `existing-shared`**: skip `smaqit.infrastructure-provision-cyso` entirely, `deploy-only`-shaped CI/CD generation (no `provision.yml`, no Terraform step), the same 3-secret GitHub Actions set (`VM_SSH_KEY`/`VM_SSH_PUBLIC_KEY`/`GH_TERRAFORM_TOKEN`) plus a manually-set `VM_HOST`-equivalent variable, no Terraform state to `terraform destroy`.
- **Two genuine divergences from `existing-shared` — do not copy these callouts verbatim:**
  1. **Machine registration state.** `existing-shared` assumes the target machine is *already registered* in Vault (`secret/machines/<slug>/base-ssh` exists, because some other project provisioned it). `existing-unmanaged` typically targets a machine registering for the *first* time — route through `bootstrap-app-to-machine.sh`'s existing fresh-registration branch (prompts host/provider/owner_project, generates a keypair, defers public-key installation to the operator) rather than assuming pre-existing `base-ssh`. `owner_project` for this metadata record should be the *requesting* project's own slug (there is no "other project" in this mode) — this is itself a small open question the script's current two-branch design (discoverer vs. shared co-host) doesn't explicitly name; resolve it during implementation rather than leaving it implicit.
  2. **nginx vhost `default_server` logic.** `existing-shared`'s "must be name-based, never `default_server`" callout is a *co-hosting* concern, not a Terraform-management one. A dedicated `existing-unmanaged` VM is (by this mode's own definition) the only site on the box, so `write-vhost.sh`'s existing live-inspection logic already yields `default_server` correctly, same as `provision`/`existing-owned` — **no new callout should be added at this point**; adding one by rote (copying `existing-shared`'s) would be actively wrong.
- **`ownership-guard.sh`'s messaging is currently hardcoded to a co-hosting narrative** ("this project owns and manages via its own Terraform state (co-hosting)... use `provisioning_mode: existing-shared` instead"). Since orchestration skips this skill entirely for both `existing-shared` and `existing-unmanaged`, this only matters for the guard's own stated defense-in-depth purpose (a direct/manual invocation bypassing the orchestrator) — but the message text should stop asserting "another project owns it" when the real cause could equally be "nobody's Terraform owns it, by design."
- **Vault multi-machine-per-app question is explicitly out of scope here, flagged for the downstream project instead.** `secret/apps/<slug>/ssh` + `.../machine` (task 090's namespace) currently point at exactly one machine per app. A project adopting `existing-unmanaged` for a *second* target while keeping an existing `provision`/`existing-owned` target for the first (the originating downstream project's own situation) needs one app's identity to relate to two machines — the current schema has no documented answer (machine-suffixed app-key convention? one key authorized on multiple machines?). Do not guess a resolution as a side effect of this task; if it needs a framework answer, that is a follow-up task, not folded in here silently.

## Implementation Steps

1. **`skills/smaqit.input-deployment/SKILL.md`** — add `existing-unmanaged` as a fourth `Provisioning Mode` value (next to `provision`/`existing-owned`/`existing-shared`): "a VM dedicated to this project but never managed by any Terraform state — provisioned out-of-band and staying that way. Skips Terraform entirely, like `existing-shared`, but the VM is not co-hosted with another project." Update the elicitation question to mention this case.

2. **`skills/smaqit.new-greenfield-project/SKILL.md`**:
   - Pre-conditions/Applicability block (~lines 29-41): widen the `[existing-shared]` tag on the "Target VM's fixed IP known..." line to cover both modes.
   - Provisioning Mode section (~lines 88-96): add the fourth bullet; "three modes" → "four modes" everywhere that phrase appears in this file.
   - Add `→ existing-unmanaged:` callouts wherever a `→ existing-shared:` callout exists today (Phase 4 steps ~104/106/109/111/147/151; Phase 5 steps ~159/161/163), reusing `existing-shared`'s text **except** at the machine-registration step (~104, use the fresh-registration wording from Design Decisions) and the nginx-vhost step (~143, add no callout there at all) and the teardown step (~151, drop "owned by another project," keep "no Terraform state to destroy").
   - Scope section (~line 209): "three" → "four" `provisioning_mode` values.

3. **`skills/smaqit.feature-new/SKILL.md`** — Phase 3 Step 5's resolution logic (~line 76): add an `existing-unmanaged` example. Phase 3 Steps 9-10 (~lines 86, 88): note this mode shares `existing-shared`'s restricted Vault/repo-config behavior, with the same machine-registration caveat as item 2 above.

4. **`skills/smaqit.infrastructure-provision-cyso/SKILL.md`** — Scope (~lines 10-14), Completion checklist (~line 229), Failure Handling (~line 244): every "use `existing-shared` instead" becomes "use `existing-shared` or `existing-unmanaged` instead." Fix `scripts/ownership-guard.sh`'s hardcoded co-hosting message text per Design Decisions (message only — guard logic itself is unchanged).

5. **`skills/smaqit.infrastructure-repo-config/SKILL.md`** — every `existing-shared`-restricted-mode mention (~lines 22-25, 36, 63-64, 93, 109) widens to include `existing-unmanaged` (identical 3-secret set and manual-`VM_HOST` reasoning — arguably an even more direct case, since there's no Terraform output anywhere in the picture, not even another project's). `scripts/sync-secrets.sh` skip-message strings (~lines 83, 96, 123): cosmetic text update only — the skip logic is already mode-agnostic (path-existence check).

6. **`skills/smaqit.infrastructure-cicd-generate/SKILL.md`** — document that `deploy-only` generation mode now serves two `provisioning_mode` values (`existing-shared`, `existing-unmanaged`) — same generated shape, different reason ("another project's Terraform manages it" vs. "nobody's Terraform manages it"). Update `assets/deploy.yml.deploy-only.template`'s header comment (~lines 8-9) and nginx-reload comment (~lines 79-81), which currently narrate "co-hosted" specifically, so they don't assert something that may not hold under the new mode.

7. **`skills/smaqit.infrastructure-vault-loader/SKILL.md`** — no script logic change (the fresh-machine-registration branch in `bootstrap-app-to-machine.sh` already does what this mode needs). Add prose explicitly naming `existing-unmanaged` as the mode that always exercises this branch, and resolve the `owner_project` question from Design Decisions item 1.

8. Bump `metadata.version` on every skill touched (patch for doc-only changes, minor where new steps/parameters are added), matching task 084's own convention.

## Known Issues Triage
**Triaged:** 2026-07-29
**Tools searched:** Terraform, HashiCorp Vault
**Result:** Clear

### Blocking Issues
- None

### Advisory Issues
- None — the task is documentation-only over already-existing script behavior (`ownership-guard.sh`, `sync-secrets.sh`, `bootstrap-app-to-machine.sh`); it introduces no new Terraform or Vault API usage, so upstream issue search against those repos returned nothing specific to this task's actual scope.

### Historical (Closed)
- None

### Unresolvable Tools
- None

## Acceptance Criteria

- [x] `smaqit.input-deployment` resolves and surfaces `existing-unmanaged` as a fourth `provisioning_mode` value, defined distinctly from `existing-shared` (dedicated vs. co-hosted; never-Terraform-managed vs. another-project's-Terraform)
- [x] `smaqit.new-greenfield-project` Phase 4/5 document `→ existing-unmanaged:` callouts everywhere `existing-shared` has one, with the machine-registration and nginx-vhost divergences correctly reflected (not copied verbatim from `existing-shared`)
- [x] `smaqit.infrastructure-provision-cyso` (including `ownership-guard.sh`'s message text) and `smaqit.infrastructure-repo-config` (including `sync-secrets.sh`'s message text) reference `existing-unmanaged` alongside `existing-shared` wherever the latter is mentioned today
- [x] `smaqit.infrastructure-cicd-generate`'s `deploy-only` mode is documented as serving both `existing-shared` and `existing-unmanaged`, with the deploy-only template's co-hosting-specific comments corrected to not assume co-hosting
- [x] `smaqit.infrastructure-vault-loader` names `existing-unmanaged` as the mode driving `bootstrap-app-to-machine.sh`'s fresh-registration branch, with `owner_project`'s value for this mode resolved and documented
- [x] Full-tree grep for `existing-shared` across `skills/` shows no orphaned mention lacking a parallel `existing-unmanaged` reference, except the two documented, deliberate divergences — **this was false as first reported; see Review Findings below. True as of the review pass's fix.**
- [x] No remaining "three `provisioning_mode` values" language anywhere in the touched files
- [ ] Exercised end-to-end via the originating downstream project's real second-VPS deployment — not merely documented in the abstract, mirroring task 084's own acknowledged gap (documented-but-not-live-walked) and explicitly avoiding repeating it if a live walkthrough is available this time — **deliberately still open**, see Findings

## Findings

**Implementation approach:**
- Followed the 8 Implementation Steps in order across all 7 originally-scoped files, plus 2 collateral files found during the full-tree `existing-shared` grep that the original scope missed: `smaqit.infrastructure-deploy-rsync/SKILL.md` (clarified that `write-vhost.sh`'s `default_server`-vs-name-based decision is driven by live VM state, not `provisioning_mode`, so `existing-unmanaged` needs no special callout there) and `smaqit.infrastructure-cicd-generate/assets/deploy.yml.full.template` (a maintainer comment naming "the existing-shared case," widened for consistency even though this template path is never used by `existing-unmanaged` itself).
- Verification step (full-tree grep for `existing-shared`, check for leftover "three modes" language) ran clean after all edits — every hit either has a parallel `existing-unmanaged` mention or falls into one of the two deliberate divergences documented in Design Decisions.

**Decisions made:**
- All branching documented inline via `→ existing-unmanaged:` callouts (or `[existing-shared/existing-unmanaged]` tag widening), matching task 084's own established convention, rather than duplicating step sequences.
- `bootstrap-app-to-machine.sh`'s fresh-machine-registration prompt for `owner_project` was reworded (not just documented in prose) to stop assuming a Terraform-provisioning project always exists — it now explicitly names the `existing-unmanaged` case as "the requesting project's own slug." This was flagged as an open question in Design Decisions and resolved during implementation rather than left as a dangling TBD.
- Version bumps followed the file's own existing semver convention: minor bump where a new value/parameter was added (`smaqit.input-deployment` 1.1.0→1.2.0, `smaqit.new-greenfield-project` 1.4.2→1.5.0, `smaqit.feature-new` 1.0.0→1.1.0, `smaqit.infrastructure-cicd-generate` 2.0.0→2.1.0, `smaqit.infrastructure-vault-loader` 3.3.0→3.4.0), patch bump for message-text-only changes (`smaqit.infrastructure-provision-cyso` 1.5.0→1.5.1, `smaqit.infrastructure-repo-config` 1.4.0→1.4.1). `smaqit.infrastructure-deploy-rsync` was not version-bumped — its edit was a clarifying addition to already-correct behavior, not a behavior change.

**Blockers encountered:**
- None.

**Follow-up identified:**
- The one deliberately unchecked acceptance criterion — a live walkthrough of `existing-unmanaged` via the originating downstream project's real second-VPS deployment — is a separate, later stage of the effort that produced this task, tracked in that project's own task planning, not yet started as of this task's completion. Closing this task now regardless mirrors task 084's own precedent (documented-but-not-live-walked, closed anyway with the gap explicit) rather than leaving a purely-documentation task open indefinitely while waiting on a separate downstream project's own scheduling.
- The Vault multi-machine-per-app question flagged in Design Decisions (one app's `secret/apps/<slug>/ssh`+`.../machine` currently pointing at exactly one machine, with no documented answer for an app that legitimately targets two) was deliberately left unresolved here, as scoped. If that downstream project's live walkthrough hits it in practice, that's the trigger for a follow-up task, not a reason to guess a resolution now.

**Review Findings (2026-07-29, retroactive `task.start` + review — implementation had been written but never committed):**
- The original implementation's own AC6 claim (clean full-tree `existing-shared` grep) did not hold: `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` was never touched, and isn't mentioned anywhere in the original Implementation Steps, Design Decisions, or Findings, despite being the credential loader the skill's own docs say to run every session and the one place other than `bootstrap-app-to-machine.sh` where `PROVISIONING_MODE`-aware branching lives. This is a **legacy flat-scheme** script (pre-dates the `apps/`+`machines/` namespace from task 090) — reachable whenever a legacy-scheme project's deployment target is itself `existing-unmanaged`, not only in the multi-machine scenario Design Decisions already deferred.
- Concretely, before the fix: `PROVISIONING_MODE=existing-unmanaged` didn't match the script's hardcoded `"existing-shared"` checks, so it silently fell through to the `provision`/`existing-owned` branch. `ssh`/`github` still got created (not missing), but two things were wrong: (1) the operator was wrongly prompted for `cyso`/`tfstate` credentials a Terraform-unmanaged VM will never use; (2) more seriously, the SSH-keypair step took the branch built for `provision`/`existing-owned`, which generates a keypair silently because Terraform is expected to push it onto the VM via cloud-init — for `existing-unmanaged` there is no Terraform run to do that, so the operator would get a keypair with no instruction that they must manually install it on the VM, a silent precondition for the next deploy's SSH step to work at all.
- Fixed: `load-credentials.sh` now treats `existing-shared` and `existing-unmanaged` alike for the cyso/tfstate skip and the 2-path/`REQUIRED_PATHS` accounting (`RESTRICTED_MODE` variable), and gives `existing-unmanaged` its own SSH branch — always generate-and-print manual-install instructions, without `existing-shared`'s "copy the owning project's keypair" choice, since `existing-unmanaged` has no owning project to copy from. `smaqit.infrastructure-vault-loader/SKILL.md`'s scheme-detection section and usage examples updated to match; a second minor gap (`smaqit.feature-new/SKILL.md`'s Scope section citing only `existing-shared` as its co-hosted-example parenthetical) was also corrected.
- No version bump applied to `smaqit.infrastructure-vault-loader/SKILL.md` beyond the original run's 3.3.0→3.4.0 (still a minor-scope addition of the same kind) or to `smaqit.feature-new/SKILL.md` beyond 1.0.0→1.1.0 (prose-only correction within the same already-bumped scope).

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.input-deployment/SKILL.md` | Modify — add `existing-unmanaged` provisioning mode value |
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify — applicability tags, Provisioning Mode section, `→ existing-unmanaged:` callouts |
| `skills/smaqit.feature-new/SKILL.md` | Modify — Phase 3 resolution example + Vault/repo-config notes + Scope parenthetical (review pass) |
| `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` | Modify (review pass) — `existing-unmanaged`-aware restricted mode + dedicated SSH branch |
| `skills/smaqit.infrastructure-provision-cyso/SKILL.md` | Modify — Scope/Completion/Failure Handling wording |
| `skills/smaqit.infrastructure-provision-cyso/scripts/ownership-guard.sh` | Modify — correct hardcoded co-hosting message text |
| `skills/smaqit.infrastructure-repo-config/SKILL.md` | Modify — restricted-mode wording |
| `skills/smaqit.infrastructure-repo-config/scripts/sync-secrets.sh` | Modify — skip-message text |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md` | Modify — `deploy-only` mode now serves two provisioning_mode values |
| `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.deploy-only.template` | Modify — de-couple comments from co-hosting assumption |
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Modify — name `existing-unmanaged`, resolve `owner_project` question |

## Notes

Originates from a downstream project's own deploy-to-a-second-VPS task, during that project's session assessment of a user proposal to build a bespoke new skill (`smaqit.deploy-new`) for this case. Investigation found the existing `provisioning_mode` axis (task 084) already generalizes almost exactly what was needed — the user's instinct that a new PR-marker-gated workflow was needed turned out to be correct in principle (a second, Terraform-free deploy workflow does need its own marker/dispatcher), but the framework-level gap was narrower than a whole new skill: one missing enum value. Mirrors the task 089/092/094/095 precedent of tracking real framework gaps discovered on downstream projects, rather than fixing them ad hoc in place and losing the feedback loop back to the framework.

Cross-reference: that downstream project's own second-VPS deployment work is blocked on this task landing (or at minimum its `existing-unmanaged` mode being usable) before it can proceed.
