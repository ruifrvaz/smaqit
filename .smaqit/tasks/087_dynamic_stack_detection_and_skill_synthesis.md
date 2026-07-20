# Dynamic Stack Detection + On-the-Fly Deploy Skill Synthesis

**Status:** Not Started
**Mode:** Assisted
**Created:** 2026-07-20

## Description

While preparing to test-deploy `tested-deployment` (a plain Tornado app — pure Python, no framework,
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
fallback next (immediately: `tested-deployment`).

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
  **Confirmed present in `tested-deployment`** (the project that triggered this task): `smaqit.create-skill`
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

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] Phase 4 Step 6 reads the declared stack from `specs/stack/platform-stack.md` (not
      filesystem re-detection) and judges the match against currently-installed deploy skills
      generically — no hardcoded per-stack list — reporting what was checked either way, rather
      than silently defaulting
- [ ] `smaqit.create-skill` is documented as the primary synthesis path, with an always-available
      manual-authoring fallback when it isn't present
- [ ] The four required-inherited-context items (`__APP_DIR__`, `write-vhost.sh` reuse, deploy-stamp
      pattern reuse, guard-script reuse) are documented as explicit synthesis inputs, not left
      implicit
- [ ] The human checkpoint between synthesis and execution is documented and not silently skippable
      in Assisted mode
- [ ] Synthesized-skill provenance metadata convention is documented
- [ ] Task 084's `provisioning_mode` branching is still intact and legible after this change; Task
      086's hardcoded two-way stack list is deliberately superseded by the generic lookup (not kept
      alongside it) — confirm no leftover hardcoded enumeration remains
- [ ] **Not part of this task's own completion, but the reason it exists:** this capability gets
      exercised for real against `tested-deployment`'s actual Tornado/systemd stack as the next,
      separate step (the live test-deploy playbook) once this task lands

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
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify — Phase 4 Step 6 dynamic detection + synthesis procedure, Gotchas updates |

## Notes

- Originates directly from attempting to plan a real test-deploy of `tested-deployment` to a new Cyso
  VM and discovering its stack (Tornado + systemd + nginx, no Docker) matches neither existing
  deploy skill. User explicitly asked for a *dynamic, real-time* adaptation (detect → synthesize →
  run) rather than either pre-building a third static skill or leaving the fallback as unstructured
  prose.
- `smaqit.create-skill`'s own preconditions — the skill itself, its `smaqit.L2` compiler
  dependency, and `.smaqit/templates/skills/` — were checked directly in `tested-deployment` (not
  assumed) and are confirmed present: `smaqit.create-skill` v2.0.0 at
  `.github/skills/smaqit.create-skill/`, `smaqit.L2` at `.github/agents/smaqit.L2.agent.md`,
  `.smaqit/templates/skills/` present. The primary synthesis path applies for the upcoming real
  test there.
- This task does not touch `tested-deployment` at all — it only changes `smaqit.new-greenfield-project`
  in this canonical repo. The actual live test (synthesizing a Tornado/systemd deploy skill and
  using it to deploy `tested-deployment` to a new VM) is deliberately a separate, subsequent piece of
  work — the original ask this whole detour started from.
