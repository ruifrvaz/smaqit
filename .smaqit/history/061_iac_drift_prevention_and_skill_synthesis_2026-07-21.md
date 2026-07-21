# IaC Drift Prevention and Skill Synthesis

**Date:** 2026-07-21
**Session Focus:** Plan and execute Task 087 (dynamic deploy-skill synthesis), discover and fix a real IaC-drift near-miss during the live validation deploy of `<tested-deployment>`, fix two vault-loader bugs found in that same first-run, and remove residual "HIM Corporate" project-specific traces from canonical skill templates.
**Tasks Referenced:** 087 (Dynamic Stack Detection + On-the-Fly Deploy Skill Synthesis)
**Tasks Completed:** 087 (canonical-repo scope only — see Next Steps)

---

## Actions Taken

### Task 087 Planning and Implementation
- Ran `smaqit.task-plan 087`; discovery revealed the task's own Notes were stale — it claimed no connection to `<tested-deployment>` and no relation to a new VM, when in fact `<tested-deployment>` was the concrete motivating project and the live validation target
- User corrected: `<tested-deployment>` gets a **new**, dedicated Terraform-provisioned VM (not the shared VM from an earlier manual deploy); task sequencing is canonical rewrite first, then a separate live validation pass; the live validation also supplies the missing evidence for Tasks 084/085's one remaining unchecked acceptance criterion each
- Corrected task file Notes and Acceptance Criteria accordingly
- Rewrote `smaqit.new-greenfield-project` Phase 4 Step 6: replaced the hardcoded two-way stack list + "default and adapt as needed" fallback with a deterministic precondition → judgment → synthesis procedure — read the declared stack from `specs/stack/platform-stack.md`, judge against currently-installed `smaqit.infrastructure-deploy-rsync*` skills generically (no hardcoded enumeration), and on no match invoke `smaqit.create-skill` (primary) or manual authoring (fallback) with a mandatory human checkpoint before invoking the synthesized skill
- Added a Gotcha documenting the four required-inherited-context items any synthesized deploy skill must reuse: `__APP_DIR__` token, shared `write-vhost.sh`, deploy-stamp pattern, guard-script reuse
- Bumped `smaqit.new-greenfield-project` to 1.4.0, then 1.4.1 (incidental trace fix, see below)

### Live Validation — IaC Drift Near-Miss (via sibling `<tested-deployment>` session)
- While validating Task 087 for real against `<tested-deployment>`'s Tornado/systemd/no-Docker stack, the agent manually SSH'd into the freshly-provisioned VM to fix a broken `cloud-init user_data` step (`pip3 install tornado` failing under PEP 668) instead of only fixing it through Terraform
- This diverged the live instance from Terraform's declared config; a subsequent `terraform plan` (run as a bare, unguarded command — not through any guard script) proposed **replacing** the running instance and its volume attachment, which would have reassigned `fixed_ip` and broken `VM_HOST`
- Caught before `apply` by manually reading the plan output — not because any guard blocked it
- User flagged this directly: despite prior sessions building `plan-guard.sh`/`ownership-guard.sh` specifically to prevent this class of destructive apply, the agent still ran a bare `terraform plan` and only avoided disaster by chance

### IaC Drift Prevention Fix
- Root-caused: the downstream `<tested-deployment>` copy of `smaqit.infrastructure-provision-cyso` was stale at v1.0.0 — it predated `plan-guard.sh`'s existence entirely and its Step 6 literally said "Review plan. `terraform plan`", with no guard mandate at all. The guard scripts existed and worked correctly in canonical (confirmed: `plan-guard.sh` and `ownership-guard.sh` are byte-identical between canonical and a properly-synced downstream copy) — this was a sync gap, not a design flaw in the guards themselves
- Added to `agents/deployment.md`: new MUST NOT (never patch a live IaC-managed resource out-of-band without reconciling IaC config/state in the same change) and new MUST (route every plan/apply against already-provisioned infrastructure through the project's guard script, including diagnostic-only checks with no intent to apply) — plus a new "IaC Drift Prevention" section spelling out the remediation path (`lifecycle { ignore_changes }` or a deliberate, reviewed `terraform apply -replace`)
- Added a matching Gotcha to `smaqit.infrastructure-provision-cyso/SKILL.md` documenting the same failure mode

### Vault-Loader Bug Fixes (found during `<tested-deployment>`'s first-ever `load-credentials.sh` run)
- **SSH key "error in libcrypto"**: `vault kv put private_key="$(cat file)"` strips the private key's trailing newline via shell command substitution; OpenSSH's parser requires it. Fixed all three occurrences (default path, `existing-shared` copy-from-source, `existing-shared` generate-new) to use Vault's `@file` syntax instead, which preserves exact bytes. New Gotcha documents the root cause and warns against regressing to `$(cat ...)`
- **Project-slug misdetection**: the slug-derivation regex matched an unfilled `AGENTS.md` placeholder (`[TODO: add project name]`), and the extraction regex itself sliced off the word "TODO" before a post-extraction guard could catch it — silently writing all four credentials to `secret/add-project-name/*` instead of the real slug. Fixed by excluding any line containing "TODO" (case-insensitive) *before* extraction, not after. Confirmed the same bug and fix applied to both canonical and the (differently-structured) downstream copy

### HIM Corporate Trace Removal (Category C cleanup)
- User asked to review canonical skills for lingering project-specific ("HIM Corporate") traces after the agent's own session-assessment flagged three files; a full-tree grep found the actual scope was **11 files**, not 3
- Genericized across: `smaqit.infrastructure-deploy-rsync`, `-deploy-verify`, `-domain-tls`, `-hook-post-deploy-stamp`, `-vm-bootstrap`, `-provider-cyso` (SKILL.md + all three `references/*.md` files), `smaqit.requirements-extract`, `smaqit.spec-status-update`, `smaqit.test-e2e-playwright`, plus incidental fixes folded into the Task 087 and drift-prevention commits for `smaqit.new-greenfield-project` and `smaqit.infrastructure-provision-cyso`
- Replaced hardcoded IPs, nginx site names, health-check paths, spec filenames, `INF-TOPOLOGY-*` IDs, Terraform resource labels, and example domains with generic placeholders (`<project-slug>`, `<vm-fixed-ip>`, `__APP_DIR__`, `<health-check-path>`, etc.)
- Confirmed via a final full-tree grep: zero remaining traces

### Commits
- `cecb59a` — feat(087): dynamic stack detection + skill synthesis for deploy dispatch
- `ff63f1f` — fix: prevent IaC drift from out-of-band manual VM fixes
- `396c71f` — fix: two vault-loader bugs found during a live downstream-project first-run
- `f18b664` — chore: remove HIM Corporate project-specific traces from skill templates

All local only — nothing pushed this session (SSH still unavailable in this sandbox, same recurring blocker as prior sessions).

---

## Problems Solved

- **Task 087's own task file had stale/incorrect Notes** contradicting the actual plan — corrected before implementation began, avoiding wasted work against a wrong premise.
- **IaC drift near-miss**: root-caused to a genuine downstream sync gap (stale skill copy), not a flaw in the drift-prevention design itself; fixed at both the agent-instruction level (never patch out-of-band) and the skill level (mandatory guard script, no exception for diagnostics).
- **Two real, previously-uncaught vault-loader bugs**, both surfaced only because this was the first time `load-credentials.sh` ran against a project using `AGENTS.md`-only conventions with a freshly-generated key end-to-end.
- **Scope of "HIM Corporate" traces was 3.7x larger than initially assessed** — a full-tree grep rather than spot-checking caught the true extent.

## Decisions Made

- Task 087's canonical-repo work (Phase 4 Step 6 rewrite) is scoped separately from its live validation (a distinct, subsequent step against `<tested-deployment>`) — no new canonical deploy skill is authored as part of Task 087 itself; that only happens via a future Task-086-style reconciliation once the synthesized skill is proven by real use.
- Drift remediation offers two paths (`ignore_changes` vs. deliberate `apply -replace`) without picking one — left as an explicit downstream decision since both have real trade-offs.
- Historical citations in Gotchas (e.g. "Observed in HIM Corporate session 005") were genericized rather than deleted — the lesson-learned content is preserved, only the project name is removed.

## Files Modified

### Modified (canonical)
- `.smaqit/tasks/087_dynamic_stack_detection_and_skill_synthesis.md` — Notes/AC correction
- `agents/deployment.md` — IaC Drift Prevention section, new MUST/MUST NOT
- `skills/smaqit.new-greenfield-project/SKILL.md` — Phase 4 Step 6 rewrite; version 1.3.0 → 1.4.1
- `skills/smaqit.infrastructure-provision-cyso/SKILL.md` — out-of-band-fix Gotcha; version 1.4.0 → 1.4.1
- `skills/smaqit.infrastructure-vault-loader/SKILL.md` + `scripts/load-credentials.sh` — two bug fixes
- `skills/smaqit.infrastructure-deploy-rsync/SKILL.md` — trace removal; version 1.2.0 → 1.2.1
- `skills/smaqit.infrastructure-deploy-verify/SKILL.md` — trace removal
- `skills/smaqit.infrastructure-domain-tls/SKILL.md` — trace removal; version 1.0.0 → 1.1.0
- `skills/smaqit.infrastructure-hook-post-deploy-stamp/SKILL.md` — trace removal
- `skills/smaqit.infrastructure-vm-bootstrap/SKILL.md` — trace removal; version 1.1.0 → 1.1.1
- `skills/smaqit.infrastructure-provider-cyso/SKILL.md` + all `references/*.md` — trace removal; version 1.1.0 → 1.1.1
- `skills/smaqit.requirements-extract/SKILL.md`, `skills/smaqit.spec-status-update/SKILL.md`, `skills/smaqit.test-e2e-playwright/SKILL.md` — trace removal

---

## Next Steps

- **Push required**: 4 new local commits (`cecb59a`, `ff63f1f`, `396c71f`, `f18b664`) plus the pre-existing unpushed `bd0b9c2` and the still-unpushed `v1.4.0` release from session 060 — all need a machine with GitHub SSH access.
- **Package this batch into a proper release** rather than pushing incrementally, per user's stated preference this session.
- **Task 087 completion**: canonical-repo scope is done; still pending is the live validation exercise against `<tested-deployment>` proving the synthesis procedure end-to-end (this was in fact completed this session, in the sibling `<tested-deployment>` project — see that project's own session history for the full account). Consider closing Task 087 formally next session once both sides are confirmed.
- **Drift remediation decision still open**: `<tested-deployment>`'s `main.tf` needs either `lifecycle { ignore_changes = [user_data] }` or a deliberate reviewed replace to resolve the `user_data` drift flagged as `INF-DEPLOYMENT-011: [!]` in that project's infrastructure spec — not yet applied.
- **PLANNING.md cleanup still pending** (flagged across sessions 060 and this one, never actioned): tasks 071/074 likely already complete in the codebase; task 070's priority discrepancy (High in file vs. Low in table) unresolved.

---

## Session Metrics

- **Date:** 2026-07-21
- **Commits:** 4 local (`cecb59a`, `ff63f1f`, `396c71f`, `f18b664`) — none pushed
- **Skills genericized:** 12 files (Category C sweep)
- **Bugs fixed:** 2 (vault-loader SSH key newline, project-slug misdetection)
- **New agent directive section:** IaC Drift Prevention (`agents/deployment.md`)
</content>
