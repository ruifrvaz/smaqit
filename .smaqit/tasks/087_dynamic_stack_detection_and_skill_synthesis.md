# Dynamic Stack Detection + On-the-Fly Deploy Skill Synthesis

**Status:** Completed
**Mode:** Assisted
**Created:** 2026-07-20
**Started:** 2026-07-21
**Completed:** 2026-07-23

## Description

While preparing to test-deploy `<tested-deployment>` (a plain Tornado app — pure Python, no framework,
no `backend/`/`frontend` split, deployed via systemd + nginx directly on the VM, no Docker at all)
to a fresh Cyso VM, its stack turned out to match **neither** existing deploy skill:
`smaqit.infrastructure-deploy-rsync` (Node.js + Vite/React, Docker Compose) nor
`smaqit.infrastructure-deploy-rsync-python-nextjs` (FastAPI + Next.js, Docker Compose, Task 086).

`smaqit.new-greenfield-project`'s current Phase 4 Step 6 fallback for exactly this case reads: "If
neither matches, default to `smaqit.infrastructure-deploy-rsync` and adapt as needed." That's
precisely the low-determinism escape hatch Tasks 084/085 were built to eliminate elsewhere in this
same flow.

Rather than statically pre-building a third skill for this one stack shape (there is no limit to
how many stack shapes could show up next), this task upgrades the fallback itself into a
deterministic procedure: read the already-declared stack from the stack spec (authoritative — no
independent re-detection), judge whether an existing deploy skill matches it, and if none does,
synthesize a new, properly-conventioned one on the fly using `smaqit.create-skill` (confirmed
present and working in `smaqit-adk` — gathers a name, scans project context, infers a full skill
spec, compiles it via `smaqit.L2`, no back-and-forth questions), present it for a quick human
checkpoint, then invoke it to perform the actual deploy. This is a capability change to the
orchestration logic in `smaqit.new-greenfield-project` only — it does not pre-author any specific
new stack skill itself; that happens live, driven by whatever project actually triggers the
fallback next (immediately: `<tested-deployment>`).

## Design Decisions

- **The stack spec is the authoritative source of truth for what stack is being deployed — Step 6
  reads it, it does not re-derive the stack from the filesystem.** `specs/stack/platform-stack.md`
  (produced by the `/smaqit.stack` agent in Phase 2) is where the project's stack is declared once;
  the infrastructure spec (Phase 2, `/smaqit.infrastructure`) is itself downstream of and depends on
  that declaration. Re-inferring the stack independently at deploy time (e.g. by grepping for
  `package.json`+Vite vs `pyproject.toml`+FastAPI) would create a second, potentially-diverging
  source of truth and is explicitly rejected.
- **Matching against existing deploy skills is a judgment call over whatever skills currently
  exist, not a hardcoded enumeration baked into `new-greenfield-project`'s prose.** Task 086 wrote
  Step 6 as a fixed two-way list ("Node+Vite → X, Python/FastAPI+Next.js → Y"); this task
  generalizes that to "read the declared stack, then judge which of the currently-installed
  `smaqit.infrastructure-deploy-rsync*` skills (by description/metadata) matches it." This scales
  without needing Step 6 edited every time a Task-086-style reconciliation adds another skill to
  the family. If no existing skill matches → synthesize (below). Report what was checked either way
  (declared stack, which skills were compared against it, match or no match).
- **`smaqit.create-skill` is the primary synthesis mechanism**, when available. Availability check:
  the skill file itself under the project's compiled skills directory (`.github/skills/smaqit.create-skill/`
  for Copilot-style installs, `.claude/skills/` for Claude-Code-style), plus its `smaqit.L2`
  compiler dependency (`.github/agents/smaqit.L2.agent.md` or equivalent) and `.smaqit/templates/skills/`
  — all three are required for the pipeline to actually run, not just the skill file's presence.
  **Confirmed present in `<tested-deployment>`** (the project that triggered this task): `smaqit.create-skill`
  v2.0.0 at `.github/skills/smaqit.create-skill/`, `smaqit.L2` at `.github/agents/smaqit.L2.agent.md`,
  and `.smaqit/templates/skills/` all verified — so the primary path applies for that project's
  upcoming real test, not the fallback. It already does exactly what's needed: gather a name, scan
  context, infer a complete spec, compile via `smaqit.L2`, report ambiguous fields. Do not reinvent
  it.
- **Generic manual-authoring fallback when `smaqit.create-skill` isn't available.** The agent
  writes the new `SKILL.md` directly, following the exact structural shape of the two existing
  deploy-rsync skills (Pre-conditions, numbered Steps, Output, Scope, Gotchas, Completion, Failure
  Handling). This is always possible on any platform (Claude Code, Copilot, Codex) — it's following
  a documented template, not a special tool.
- **Point synthesis explicitly at the existing, validated `smaqit.infrastructure-deploy-rsync*`
  family as reference exemplars — not a generic "write a deploy skill" prompt.**
  `smaqit.create-skill`'s own Step 2 already scans a project's existing skills for conventions, but
  a generic scan doesn't know these specific siblings are the relevant family to weight heavily
  among a potentially large skill set. Name them explicitly (`smaqit.infrastructure-deploy-rsync`,
  `smaqit.infrastructure-deploy-rsync-python-nextjs`) as required reading before inferring the new
  skill's spec.
- **Constrain what synthesis must inherit** from those exemplars, so it only fills in the
  stack-specific parts rather than reinventing everything already solved:
  - `__APP_DIR__` token convention (Task 085).
  - Nginx vhost writing delegates to the shared
    `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh` — a synthesized skill must never
    invent its own `default_server`-vs-name-based logic.
  - Deploy-stamp writing reuses `smaqit.infrastructure-hook-post-deploy-stamp`'s pattern.
  - Any Terraform-touching step reuses `plan-guard.sh`/`ownership-guard.sh` (vendored per Task 085),
    never a bare `terraform apply`.
  These four are passed as explicit required-context to `smaqit.create-skill` (or the manual
  fallback), not left for it to rediscover independently.
- **Human checkpoint between synthesis and execution, not full autonomy end-to-end.** Balances
  "dynamic, real-time adaptation" (no back-and-forth during detection/scanning itself) against
  blindly running freshly-generated SSH/rsync/systemctl commands against a real cloud VM. Present
  the synthesized `SKILL.md` before invoking it. Skippable only in Autonomous mode with the user's
  prior, explicit sign-off for that mode.
- **Provenance metadata on synthesized skills** — same convention as Task 086's imported skill
  (`validated`/`validated-stack`-style fields, or `synthesized`/`synthesized-for-project`/
  `synthesized-date`). Marks it as a candidate for a future Task-086-style reconciliation back into
  canonical `smaqit` once proven by real use — closing the loop this session has traced twice now
  (downstream need → validated in the field → reconciled upstream).

## Implementation Steps

1. **Rewrite `smaqit.new-greenfield-project` Phase 4 Step 6** replacing both the hardcoded
   two-way stack list (Task 086) and the "adapt as needed" fallback with one precondition +
   judgment + synthesis procedure:
   - **Precondition:** read the declared stack from `specs/stack/platform-stack.md` — the
     authoritative source. Do not re-derive it from the filesystem.
   - **Judgment:** compare the declared stack against the currently-installed
     `smaqit.infrastructure-deploy-rsync*` skills (by description/metadata) and pick the one that
     matches, if any.
   - **Matched** → invoke it. (This replaces Task 086's hardcoded "Node+Vite → X, Python/FastAPI+Next.js
     → Y" list with a lookup that scales to however many skills currently exist in the family,
     without needing this step edited every time one is added.)
   - **No match** →
     a. Report what was checked: the declared stack, which skills were compared against it, and
        why none matched.
     b. Check whether `smaqit.create-skill` is available: the skill file under the project's
        compiled skills directory, its `smaqit.L2` compiler dependency, and `.smaqit/templates/skills/`
        must all be present.
     c. If available: invoke it with a name derived from the declared stack (e.g.
        `smaqit.infrastructure-deploy-rsync-<stack-slug>`), explicitly pointed at the existing
        `smaqit.infrastructure-deploy-rsync*` skills as reference exemplars, and the four
        required-inherited-context items from Design Decisions as explicit input, not left implicit.
     d. If unavailable: author the `SKILL.md` by hand, following the existing deploy-rsync skills'
        structural shape, with the same four required-inherited-context items applied manually.
     e. Present the synthesized skill to the user for the checkpoint.
     f. Invoke the synthesized skill to perform the actual deploy.

2. **Update the "Source path contract" Gotcha** to replace "adapt as needed" with a pointer to this
   new procedure.

3. **Add a Gotcha documenting the four required-inherited-context items**, so future synthesis (by
   any agent, not just this session) doesn't drift from established conventions by rediscovering
   them independently each time.

4. **Read through the merged Phase 4 Step 6** to confirm both axes — `provisioning_mode` (Task 084,
   unchanged, orthogonal) and the new stack-spec-driven judgment/synthesis procedure (superseding
   Task 086's hardcoded list) — remain legible together, not just individually correct.

5. Bump `metadata.version` on `smaqit.new-greenfield-project`.

## Known Issues Triage
**Triaged:** N/A
**Tools searched:** none
**Result:** Not run

Implementation happened in a sibling session via `smaqit.task-plan 087` rather than the full
`smaqit.task-start` ceremony, so formal triage was skipped at the time. Reconstructing this
retroactively (this task file's Status/Findings update, 2026-07-21) has no planning value now —
the code is already written, committed (`b20a1ed`), and live-validated. No third-party tools are
introduced by this task's own change (pure orchestration prose in `new-greenfield-project`); the
tools it references (`smaqit.create-skill`, `smaqit.L2`) are internal to the smaqit ecosystem, not
external dependencies triage would search for.

## Acceptance Criteria

- [x] Phase 4 Step 6 reads the declared stack from `specs/stack/platform-stack.md` (not
      filesystem re-detection) and judges the match against currently-installed deploy skills
      generically — no hardcoded per-stack list — reporting what was checked either way, rather
      than silently defaulting
- [x] `smaqit.create-skill` is documented as the primary synthesis path, with an always-available
      manual-authoring fallback when it isn't present
- [x] The four required-inherited-context items (`__APP_DIR__`, `write-vhost.sh` reuse, deploy-stamp
      pattern reuse, guard-script reuse) are documented as explicit synthesis inputs, not left
      implicit
- [x] The human checkpoint between synthesis and execution is documented and not silently skippable
      in Assisted mode
- [x] Synthesized-skill provenance metadata convention is documented
- [x] Task 084's `provisioning_mode` branching is still intact and legible after this change; Task
      086's hardcoded two-way stack list is deliberately superseded by the generic lookup (not kept
      alongside it) — confirm no leftover hardcoded enumeration remains
- [x] **Live-validated 2026-07-21** against `<tested-deployment>`'s real Tornado/systemd/no-Docker stack,
      on a newly-provisioned dedicated VM (`<tested-deployment-vm>`, `<vm-fixed-ip>`, fresh Terraform state —
      `provisioning_mode: provision`, not the old shared VM from the 2026-07-14 deploy). Step 6's
      no-match path fired exactly as designed: no existing `deploy-rsync*` skill matched, the agent
      invoked `smaqit.create-skill` → `smaqit.L2`, produced
      `smaqit.infrastructure-deploy-rsync-python-tornado`, presented it for human approval (the
      checkpoint held — not silently skipped), then invoked it to actually deploy. The synthesized
      skill correctly reused `write-vhost.sh` rather than reinventing vhost logic, confirming the
      required-inherited-context constraint was followed in practice, not just documented. See the
      validation project's own session history for the full account.

## Findings

**Note on process:** the implementation and live validation below happened in two sibling
sessions on 2026-07-21 (canonical `smaqit` work, then a switch to `<tested-deployment>` to actually
deploy) that did not go through this task's formal `task-start`/`task-complete` ceremony — Status
stayed "Not Started" and every Acceptance Criteria checkbox stayed unchecked despite the real work
being committed. This entry reconstructs and verifies that work retroactively (2026-07-21) against
primary sources (`git log`, both projects' own `.smaqit/history/` files, and direct inspection of
the committed `SKILL.md` content) before checking anything off — see the session recap in the
parent conversation for the full verification trail.

**Implementation approach:**
- Rewrote `smaqit.new-greenfield-project` Phase 4 Step 6 exactly per this task's design: read the
  stack spec (authoritative), generic lookup against installed `deploy-rsync*` skills, and on no
  match: report → check `smaqit.create-skill` availability → synthesize (primary: create-skill,
  fallback: manual) → human checkpoint → invoke. Committed as `b20a1ed`.
- The task file's own Notes were stale at the time (claimed no connection to `<tested-deployment>`) —
  corrected before implementation, per that session's own account.
- Live validation happened immediately after, in `<tested-deployment>` itself: a genuinely new dedicated
  VM (`<tested-deployment-vm>`) was provisioned, no existing deploy skill matched its Tornado/systemd/no-Docker
  stack, and the no-match path fired for real — `smaqit.create-skill` → `smaqit.L2` synthesized
  `smaqit.infrastructure-deploy-rsync-python-tornado`, it was presented for approval (checkpoint
  held), then invoked to actually deploy and verify the application.

**Decisions made:**
- No new canonical deploy skill is authored as part of Task 087 itself — `deploy-rsync-python-tornado`
  lives only in `<tested-deployment>` for now, exactly as designed (a future Task-086-style reconciliation
  is the deliberate next step, not automatic).
- `<tested-deployment>` was given a dedicated new VM rather than returning to the shared VM from its
  earlier (2026-07-14) manual deploy — this was the right target for proving `provisioning_mode:
  provision` end-to-end, not just the synthesis mechanism in isolation.

**Blockers encountered:**
- None blocking Task 087 itself. The live session did surface a real, separate IaC-drift near-miss
  (an out-of-band manual SSH fix left `main.tf` and the live VM's `user_data` diverged; a later bare,
  unguarded `terraform plan` proposed replacing the running instance) — root-caused to a stale
  downstream copy of `smaqit.infrastructure-provision-cyso` (v1.0.0, pre-dating `plan-guard.sh`
  entirely), not a flaw in Task 087's design. Tracked and being resolved separately, in
  `<tested-deployment>` — not part of this task's scope.

**Follow-up identified:**
- Consider a Task-086-style reconciliation of `smaqit.infrastructure-deploy-rsync-python-tornado`
  into canonical `smaqit` now that it's been proven by real use — not yet done.
- This task file's own process gap (real work landing without the formal task-start/complete
  ceremony) is itself worth a lighter-touch note for future sessions: `smaqit.task-plan` alone
  doesn't set `Status: In Progress` or touch `PLANNING.md` the way `smaqit.task-start` does — worth
  keeping in mind when starting substantial work directly from a plan.

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify — Phase 4 Step 6 dynamic detection + synthesis procedure, Gotchas updates |

## Notes

- Originates directly from attempting to plan a real test-deploy of `<tested-deployment>` to a new Cyso
  VM and discovering its stack (Tornado + systemd + nginx, no Docker) matches neither existing
  deploy skill. User explicitly asked for a *dynamic, real-time* adaptation (detect → synthesize →
  run) rather than either pre-building a third static skill or leaving the fallback as unstructured
  prose.
- `smaqit.create-skill`'s own preconditions — the skill itself, its `smaqit.L2` compiler
  dependency, and `.smaqit/templates/skills/` — were checked directly in `<tested-deployment>` (not
  assumed) and are confirmed present: `smaqit.create-skill` v2.0.0 at
  `.github/skills/smaqit.create-skill/`, `smaqit.L2` at `.github/agents/smaqit.L2.agent.md`,
  `.smaqit/templates/skills/` present. The primary synthesis path applies for the upcoming real
  test there.
- This task's own scope only changes `smaqit.new-greenfield-project` in this canonical repo — it
  does not author or reconcile any new stack-specific deploy skill itself. `<tested-deployment>` is the
  concrete downstream project that motivates it and is the immediate next step once this task
  lands: a separate, subsequent live test-deploy that synthesizes a Tornado/systemd deploy skill
  (via `smaqit.create-skill`, confirmed present and working there) and uses it to deploy
  `<tested-deployment>` to a **newly provisioned** VM — fresh Terraform VM + Object Storage state bucket
  + Cinder volume, not a reuse of the shared VM from its earlier manual deploy. That live run also
  supplies the outstanding "real target project" live-walkthrough evidence still open on Tasks 084
  and 085 (each has exactly one unchecked acceptance criterion for lack of a real project to test
  against in-sandbox) — the same session that validates this task's Step 6 rewrite closes out that
  gap for both.
