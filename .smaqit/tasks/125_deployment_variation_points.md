---
status: Not Started
created: "2026-10-11"
---

# Deployment Variation Points: Slot-Driven Deployment Orchestration With Project-Bound Fillers

## Description

smaqit's deployment flow is not orchestrated by the deployment agent. `agents/deployment.md` is a
generic phase shell ("Transform consolidated specifications into IaC…", "Deploy… via trusted execution
layer"); the actual skill wiring — credentials → CI generation → provision → host bootstrap → deploy →
verify → TLS — is hardcoded per `provisioning_mode` inside two app-centric orchestrator skills,
`smaqit.new-greenfield-project` (Phases 4–6) and `smaqit.feature-new` (Phase 3). The L0 principle
behind it is explicit: "Deterministic routing — The invocation sequence is hardcoded per phase, not
derived at runtime" (`framework/AGENTS.md`, Phase Orchestration).

That leaves a platform repository — one that owns machines rather than an app — with no way to run its
provisioning through smaqit. A downstream platform repository (Cyso VMs running single-node k3s,
mesh-only access, gated `workflow_dispatch` applies, one GitHub Environment per machine, Vault only
through project scripts) found every canonical infrastructure skill misfitting
(`provision-cyso`, `vm-bootstrap`, `repo-config`, `vault-loader`, `deploy-verify`, `domain-tls` all
assume an app-owned VM over public SSH) and built a parallel provisioning path in its own project
skills, which smaqit's agents never reach. Tasks 114 (Finding 5: "a mode of `smaqit-deployment` or a
distinct machine-deployment agent", TBD), 115, 122 and 123 each patch one axis of the same problem by
adding branches to the closed `provisioning_mode` enum.

Make `smaqit-deployment` the authoritative deployment orchestrator for app and platform repositories
alike: each phase declares a fixed, ordered sequence of **slots** (variation points). Each slot has a
contract (purpose, inputs, handoff, completion signal) and a canonical default filler skill. A project
binds any slot to its own skill (and operation) in `.smaqit/slots.yml`, referenced from its
`AGENTS.md`; an unbound slot uses the default; `none` skips the slot. Fillers declare the slots they
fill in their own metadata, and every filler keeps the phase's invariants (guarded plan/apply, no
out-of-band fixes, credential references only). Task 116's thin-dispatcher precedent ("each
infra-owning repo owns its own contract") is the philosophical basis.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Claude Code, GitHub Copilot, Codex CLI, smaqit-adk, Go
**Platforms/Environments:** Claude Code, GitHub Copilot, Codex CLI
**Features/Integrations:** subagents (Task tool), Agent Skills, AGENTS.md project instructions, YAML configuration
**Versions/Constraints:** smaqit 3.6.0, installer skill-count test, AGENTS.md managed block is never rewritten by the installer

## Design Decisions

- **Orchestrator = the deployment agent** (user, 2026-10-10): `smaqit-deployment` owns the slot
  sequence and invokes the bound filler per slot; `new-greenfield-project` and `feature-new` stop
  hardcoding infrastructure skills and route through it. Resolves task 114 Finding 5 as "a mode of
  `smaqit-deployment`, through slots" — no separate machine-deployment agent.
- **Configuration = a separate `.smaqit/slots.yml`, referenced from `AGENTS.md`** (user, 2026-10-10):
  machine-parseable, project-owned; `AGENTS.md` carries the pointer (outside the managed
  `smaqit:instructions` block, which `installInstructionsFile` never rewrites once present).
- **Unbound slot = canonical default; `none` = skip** (user, 2026-10-10): existing app projects keep
  working unchanged with no configuration.
- **Lands first and re-scopes 114/115/122/123** (user, 2026-10-10): 114 Finding 5 resolved here; 115
  becomes optional default fillers and contracts (the platform repo owns its fillers); 122's
  reachability class becomes slot parameters of `verify`/`tls`/`ci-generate`; 123's "Vault through a
  project-provided script interface" becomes the `credentials` slot contract.
- **Fillers declare their slots** in `metadata.smaqit-slots` (slot → operation); `input-deployment`
  refuses a binding whose skill doesn't exist or doesn't declare the slot.
- **Invariants are slot-independent:** the deployment agent's guard-script MUST, the no-out-of-band
  MUST NOT and credential-references-only apply to every filler; a filler that cannot honour them is
  rejected at plan time, not tolerated.
- **Day-2 runs are single-slot runs** (e.g. `replace`), through the same agent and invariants.
- **No downstream project is named** in this task's shipped artifacts (CONTRIBUTING rule).

## Implementation Steps

1. **L0** — `framework/AGENTS.md` Phase Orchestration: replace "Deterministic routing… hardcoded per
   phase" with the slot-sequence principle (fixed order, binding resolution, default fallback,
   resolution recorded, invariants bind every filler). `framework/SKILLS.md`: filler skills and
   `metadata.smaqit-slots`. `framework/SMAQIT.md` "Extensible Through Templates": extend to variation
   points (note on task 074). Decide and record whether these go through `smaqit.new-principle` / L0
   or the direct-edit practice (ADK scope detection would treat smaqit as extension scope).
2. **Slot catalogue** for the deploy phase (purpose, inputs, handoff, completion signal, default
   filler): `credentials` (vault-loader), `environment-config` (repo-config), `platform-prereqs`
   (none), `iac-generate` (agent's generic generation), `provision` (provision-cyso), `host-bootstrap`
   (vm-bootstrap), `platform-onboarding` (request-k3s-onboarding), `ci-generate` (cicd-generate),
   `app-deploy` (deploy family selection), `tls` (domain-tls), `verify` (deploy-verify), `replace`
   (none), `teardown` (inline destroy).
3. **L1** — `.smaqit/slots.yml` schema and template (`schemaVersion`, `phase: deploy`, per-slot
   `skill`/`operation`/`params`/`environments`, values `default`/`none`); `smaqit init` creates a
   commented skeleton when absent; `templates/AGENTS.md.template` gains a pointer line; a wiki page.
4. **L2** — `skills/smaqit.input-deployment`: read and validate `.smaqit/slots.yml`, output a slot
   plan alongside `provisioning_mode`. `agents/deployment.md` (+ frontmatter yaml): Execution
   Parameters take the slot plan; Steps 5–6 walk it; single-slot day-2 entry; slot resolution in the
   deployment report; invariants restated per filler. `new-greenfield-project` Phases 4–6 and
   `feature-new` Phase 3: route infrastructure steps through the slot plan; default fillers preserve
   current behaviour. Bump each changed skill's `metadata.version`.
5. **Tests and bench** — installer Go tests (init skeleton; skill-count test unchanged unless a skill
   is added); smoke test; a bench on a fixture project: bound slot → project filler invoked, unbound →
   default, `none` → skipped, undeclared binding → refused.
6. **Re-scope notes** on tasks 114, 115, 122, 123 and 074.
7. `smaqit.qa` consistency review (L0 ↔ L1 ↔ L2), CHANGELOG, release 3.7.0.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `framework/AGENTS.md` Phase Orchestration states the slot-sequence principle (binding resolution, default fallback, recorded resolution, invariants for every filler); `framework/SKILLS.md` defines `metadata.smaqit-slots`; `framework/SMAQIT.md` mentions variation points
- [ ] A deploy-phase slot catalogue documents every slot's purpose, inputs, handoff, completion signal and default filler
- [ ] `.smaqit/slots.yml` schema/template exists; `smaqit init` creates a commented skeleton when absent; the AGENTS.md template points to it
- [ ] `smaqit.input-deployment` resolves a validated slot plan (bound skill exists and declares the slot; `default`/`none` honoured)
- [ ] `agents/deployment.md` executes the slot plan in Steps 5–6, supports a single-slot day-2 run, records the resolution in the deployment report, and keeps the guard/out-of-band/credential invariants for every filler
- [ ] `new-greenfield-project` and `feature-new` route infrastructure steps through the slot plan; with no `.smaqit/slots.yml`, behaviour is unchanged
- [ ] `make -C installer test` and the smoke test pass; bench cases (bound, unbound, none, undeclared) pass
- [ ] Tasks 114, 115, 122, 123 and 074 carry re-scope notes; no downstream project is named in shipped artifacts
- [ ] `smaqit.qa` review passes; released as 3.7.0

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
| `framework/AGENTS.md`, `framework/SKILLS.md`, `framework/SMAQIT.md` | Modify (L0) |
| `templates/slots.yml.template` (or equivalent), `templates/AGENTS.md.template` | Create / Modify (L1) |
| `installer/main.go`, `installer/main_test.go` | Modify (init skeleton) |
| `skills/smaqit.input-deployment/SKILL.md` | Modify |
| `agents/deployment.md`, `.smaqit/definitions/agents/deployment.frontmatter.yaml` | Modify |
| `skills/smaqit.new-greenfield-project/SKILL.md`, `skills/smaqit.feature-new/SKILL.md` | Modify |
| `docs/wiki/` (slot catalogue and configuration page) | Create |
| `.smaqit/bench/skills/<deployment-slots>/` | Create |
| `CHANGELOG.md` | Modify |

## Notes

- Insertion points mapped 2026-10-10: Phase Orchestration principle (`framework/AGENTS.md`, the
  "Deterministic routing" line); `input-deployment` mode resolution; deployment agent Execution
  Parameters and Steps 5–6; greenfield Step 6's family lookup is the closest existing slot (deploy
  only, prefix-matched, stack-driven) and is subsumed by the `app-deploy` slot.
- Several canonical skills read the project's AGENTS.md only for identity facts (project slug,
  owner/repo); none reads it for bindings.
- A downstream platform repository is waiting on this release to bind its own fillers.
