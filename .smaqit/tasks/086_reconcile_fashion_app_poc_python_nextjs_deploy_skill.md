# Reconcile fashion-app-poc's Python/FastAPI + Next.js Deploy Skill Into Canonical smaqit

**Status:** In Progress
**Mode:** Assisted
**Created:** 2026-07-20
**Started:** 2026-07-20

## Description

While assessing whether Task 085's deterministic CI/CD templates were too Node.js-specific, a diff
against a real downstream project — `fashion-app-poc` (a FastAPI 0.115 + Next.js 15 + pnpm +
PostgreSQL app, confirmed from its own `backend/pyproject.toml` / `frontend/package.json`) —
surfaced that its installed `.github/skills/` already contains a validated skill,
`smaqit.infrastructure-deploy-rsync-python-nextjs` (`validated: "2026-07-17"`, tested against a real
Cyso Cloud deploy), plus a Phase 4 branching instruction in its copy of
`smaqit.new-greenfield-project` ("Invoke the appropriate deploy skill based on project type") that
selects between it and `smaqit.infrastructure-deploy-rsync` by stack.

Neither exists in this canonical `smaqit` repo. This settles, with evidence rather than
speculation, whether multi-stack deploy support is a real need: it is — someone already built and
validated it, downstream, and it was never contributed back upstream. A full diff of every shared
`smaqit.infrastructure-*` skill between this repo and `fashion-app-poc`'s install also surfaced a
handful of smaller, independent divergences in both directions (this repo is ahead on
`vm-bootstrap`'s `reload-or-restart` fix and `provider-cyso`'s refined ForceNew guidance;
`fashion-app-poc` is ahead by one small convenience in `vault-loader`'s credential script). This
task reconciles the Python/Next.js deploy skill and the smaller upstream-bound fix into this repo.
It does not attempt to push this repo's own fixes back down into `fashion-app-poc`'s installed
copy — see Design Decisions.

## Design Decisions

- **Direction of reconciliation: fashion-app-poc → smaqit canonical, one way, this task.**
  Pushing this repo's own fixes (the `vm-bootstrap` / `provider-cyso` refinements
  `fashion-app-poc` is missing) back down into that project's installed copy is explicitly out of
  scope here — that happens naturally the next time `fashion-app-poc` runs its own
  update/reinstall against this improved canonical source. A `smaqit` task should not reach into
  and edit a different project's repository directly.
- **Harmonize placeholder conventions on import, don't preserve them verbatim.** The imported
  skill uses its own ad-hoc tokens (`{deploy_path}`, `{project}`) and literal
  `.github/skills/...`-compiled paths (an artifact of being read from a Copilot-target install,
  not the abstract smaqit source). Rewrite to match this repo's existing conventions:
  `{deploy_path}` → `__APP_DIR__` (same token `smaqit.infrastructure-deploy-rsync` already uses
  post-Task-085, for consistency across sibling deploy skills), `{project}` → `<project-slug>`,
  and `.github/skills/...` → `[SMAQIT_SKILLS_DIR]/...`.
- **Reuse `write-vhost.sh`, don't import a second nginx-vhost implementation.** The imported
  skill's Step 5 currently `scp`s the nginx conf directly and symlinks it manually — it predates
  Task 085 and has no `default_server`-vs-name-based awareness at all. Since vhost-writing logic
  is not stack-specific (it operates purely on the nginx conf and `sites-enabled/`, regardless of
  what's running behind it), point this skill at the same
  `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh` rather than importing (and now
  permanently maintaining) a second, less-safe implementation.
- **Merge Phase 4 Step 6 branching without regressing Task 084's `provisioning_mode` branching.**
  `fashion-app-poc`'s copy of `new-greenfield-project` is a pre-Task-084 snapshot (v1.0.0) and
  doesn't have the `→ existing-owned:` / `→ existing-shared:` callouts at all. The stack-based
  deploy-skill selection and the provisioning-mode branching are two independent axes that both
  apply to the same step; merge them legibly rather than picking one and dropping the other.
- **Fix the stray third-party project reference while importing.** `run-migrations.sh`'s own usage
  example references a project called `prior-shared-vm` — neither `fashion-app-poc` nor this repo's
  generic placeholder convention. Cosmetic only (an illustrative comment, not functional code),
  but real evidence of cross-project copy history worth cleaning up rather than propagating.
  Replace with the generic `<project-slug>` example convention used elsewhere.
- **Explicitly out of scope: a Python/Next.js variant of Task 085's `cicd-generate` templates.**
  The imported skill is a manual/agent-followed deploy runbook, not a GitHub Actions workflow
  generator. Extending `smaqit.infrastructure-cicd-generate` to also emit a Python/Next.js CI/CD
  workflow is a natural, real follow-up given this same evidence, but is a separate, sizeable
  piece of work (different setup action, no Vite timing rule, `alembic`/migration step, `pnpm`
  caching) — flagged as follow-up, not bundled here.

## Implementation Steps

1. **Import the skill**: copy
   `fashion-app-poc/.github/skills/smaqit.infrastructure-deploy-rsync-python-nextjs/` (`SKILL.md`
   + `scripts/run-migrations.sh`) into
   `smaqit/skills/smaqit.infrastructure-deploy-rsync-python-nextjs/`.

2. **Harmonize tokens and paths** in the imported `SKILL.md` and `run-migrations.sh`:
   - `{deploy_path}` → `__APP_DIR__` throughout.
   - `{project}` → `<project-slug>` throughout.
   - `.github/skills/smaqit.infrastructure-vm-bootstrap/scripts/remove-default-nginx-site.sh` →
     `[SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vm-bootstrap/scripts/remove-default-nginx-site.sh`.
   - `run-migrations.sh`'s usage example: `/opt/prior-shared-vm` → `__APP_DIR__` (or a generic
     `<project-slug>`-based example), matching the fix already made in the SKILL.md prose.

3. **Replace Step 5's inline nginx handling** with an invocation of
   `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh`, matching how
   `smaqit.infrastructure-deploy-rsync`'s own Step 6 now calls it (Task 085). Remove the
   now-redundant manual `scp`/`ln -sf`/`remove-default-nginx-site.sh` sequence from this skill's
   prose — one shared, deterministic vhost writer for both stacks, not two implementations of the
   same decision.

4. **Add Phase 4 Step 6 branching to `smaqit.new-greenfield-project`**, merged with the existing
   `provisioning_mode` callouts from Task 084:
   ```
   6. Invoke the deploy skill matching the stack declared in `specs/stack/platform-stack.md`:
      - Node.js + Vite/React → `smaqit.infrastructure-deploy-rsync`
      - Python/FastAPI + Next.js → `smaqit.infrastructure-deploy-rsync-python-nextjs`
      If neither matches, default to `smaqit.infrastructure-deploy-rsync` and adapt as needed.
      → **`existing-shared`:** if this is not the first site on the VM, the nginx vhost must be
        name-based only — never `default_server` (both deploy skills call the same
        `write-vhost.sh`, which enforces this).
   ```

5. **Update the "Source path contract" Gotcha** in `smaqit.new-greenfield-project` to mention
   stack-based deploy skill selection (matching `fashion-app-poc`'s own wording), reconciled with
   the `__APP_DIR__` terminology from Task 085.

6. **Backport the `~/.vault-token` convenience** into canonical
   `smaqit.infrastructure-vault-loader/scripts/load-credentials.sh`: read
   `VAULT_TOKEN="${VAULT_TOKEN:-$(cat ~/.vault-token 2>/dev/null || true)}"` before the existing
   "authenticate if no valid token" step, so a machine that already ran `vault login` doesn't
   re-prompt.

7. **Validate**: `bash -n` and `shellcheck` on the imported/modified `run-migrations.sh`; read
   through the merged Phase 4 Step 6 in `new-greenfield-project` to confirm both branching axes
   (`provisioning_mode` × stack type) remain legible together, not just individually correct.

8. Bump `metadata.version` on every skill touched; set an initial version + carry forward the
   `validated`/`validated-stack` metadata fields on the newly-imported skill.

## Known Issues Triage
**Triaged:** 2026-07-20
**Tools searched:** docker/compose, vercel/next.js
**Result:** Clear

### Blocking Issues
- None

### Advisory Issues
- None — `docker/compose` keyword search (`health exec`) returned one loosely-matched networking
  bug in an unrelated version/scenario; `vercel/next.js` search (`NEXT_PUBLIC build-time`) returned
  one loosely-matched issue about standalone-build server-side env vars, not the client-side
  `NEXT_PUBLIC_*` build-time baking this task's imported skill already relies on (standard,
  well-documented Next.js behavior, not a bug). Neither matches both a platform and feature
  keyword; neither is relevant to the specific patterns being imported.

### Historical (Closed)
- None

### Unresolvable Tools
- FastAPI, Alembic, PostgreSQL, pnpm — no task-specific bug search performed; this task imports
  already-validated content (tested against a real Cyso Cloud deploy per the source skill's own
  metadata) rather than newly integrating with these tools, so the risk this triage step guards
  against is low.

## Acceptance Criteria

- [x] `smaqit.infrastructure-deploy-rsync-python-nextjs` exists under canonical `skills/`, with
      `__APP_DIR__`/`<project-slug>`/`[SMAQIT_SKILLS_DIR]` conventions matching the rest of the
      skill family (not the ad-hoc tokens or compiled-path literals from the source install)
- [x] The imported skill's nginx vhost step invokes the shared `write-vhost.sh` — no second,
      independent implementation of the `default_server`-vs-name-based decision
- [x] The stray `prior-shared-vm` example reference is gone
- [x] `smaqit.new-greenfield-project` Phase 4 Step 6 branches on stack type (Node vs Python/Next.js)
      *and* still carries the full Task 084 `provisioning_mode` branching — neither axis regressed
      by the other
- [x] `smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` reads `~/.vault-token` when
      `VAULT_TOKEN` isn't already exported
- [x] `run-migrations.sh` passes `bash -n` and `shellcheck`
- [x] This task does NOT modify anything under `/home/ruifrvaz/projects/fashion-app-poc/` — the
      reconciliation is one-directional, into this repo only (verified via `git -C
      .../fashion-app-poc status --short`: only that project's own pre-existing, unrelated local
      changes are present, nothing from this session)

## Findings

**Implementation approach (interim — task not yet complete, Assisted mode):**
- Followed the 8 Implementation Steps in order. Copied `SKILL.md` + `scripts/run-migrations.sh`
  verbatim first, then harmonized in place (sed for the mechanical token/path substitutions,
  manual edit for the nginx-handling replacement, which needed real restructuring, not just
  substitution).
- Reused `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh` as-is for the Python/Next.js
  skill rather than writing a second copy — it was already stack-agnostic by construction (Task
  085), so no changes were needed to that script itself, only to how this skill invokes it.
- Verified the one-directional constraint explicitly at the end via `git -C
  .../fashion-app-poc status --short` in a separate invocation (an earlier combined-invocation
  check was accidentally checking the same directory twice due to a persisted `cd`; redone
  correctly with `-C`).

**Decisions made:**
- Kept the imported skill's `validated`/`validated-stack` metadata fields as-is (2026-07-17,
  Python 3.12/FastAPI 0.115/Next.js 15/pnpm 9/PostgreSQL 16) — real provenance worth preserving,
  not something to regenerate or generalize away.
- Bumped the imported skill to `1.2.0` (from `1.1.0` at the source) to reflect the write-vhost.sh
  integration change made during import, not just a mechanical copy.

**Blockers encountered:**
- None.

**Post-handback fix (user review, before `/task.complete`):** the imported skill's Step 2 and
Step 1 each carried an inline troubleshooting remedy (a root-container `rm -rf .next` for a
Docker-owned-permissions EACCES, and a base64-encode/decode fix for an "error in libcrypto" SSH
key issue) that duplicated content already correctly documented in Gotchas and Failure Handling.
User flagged the `.next` one specifically as concerning given its destructive-looking shape
(`docker run ... rm -rf .next`) sitting in the main happy-path step rather than a troubleshooting
section. Trimmed both to one-line pointers at the exact spot they'd trigger, keeping the full
remedy only in Gotchas/Failure Handling — verified nothing was lost by checking both target
sections still contain the complete fix. Bumped to `1.3.0`.

**Follow-up identified:**
- Two follow-ups already flagged in this task's own Notes (not part of this task): a Python/Next.js
  variant of Task 085's `cicd-generate` templates, and backporting this repo's own `vm-bootstrap`/
  `provider-cyso` refinements down into `fashion-app-poc`'s installed copy.

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-deploy-rsync-python-nextjs/SKILL.md` | New (imported + harmonized) |
| `skills/smaqit.infrastructure-deploy-rsync-python-nextjs/scripts/run-migrations.sh` | New (imported + harmonized) |
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify — Phase 4 Step 6 stack-based branching, merged with existing `provisioning_mode` branching |
| `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` | Modify — `~/.vault-token` read |

## Notes

- Originates from a live diff against `/home/ruifrvaz/projects/fashion-app-poc/.github/skills/`,
  triggered by questioning whether Task 085's Node.js-only templates were too rigid. Confirmed via
  `fashion-app-poc/backend/pyproject.toml` and `frontend/package.json` that its real stack is
  FastAPI + Next.js — the branching logic and new skill are legitimate, validated downstream work,
  not accidental cross-project contamination (the initial hypothesis).
- **Follow-up identified, not part of this task:** a Python/Next.js variant of Task 085's
  `smaqit.infrastructure-cicd-generate` templates (deploy.yml with `pnpm`/Next.js build +
  `alembic` migration step, no Vite timing rule). Real candidate given the same evidence, but
  sizeable enough to deserve its own task.
- **Follow-up identified, not part of this task:** backporting this repo's own `vm-bootstrap`
  (`reload-or-restart`, co-hosted `default_server` awareness) and `provider-cyso` (refined ForceNew
  guidance) improvements down into `fashion-app-poc`'s installed copy — deliberately out of scope
  here (see Design Decisions), but worth doing whenever that project's own maintenance next runs.
