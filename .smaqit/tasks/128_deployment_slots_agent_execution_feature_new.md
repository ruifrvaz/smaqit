---
status: Not Started
created: "2026-10-11"
parent: "125"
---

# Deployment Slots B: Deployment Agent Executes the Slot Plan; feature-new Routed Through It

## Description

Child B of task 125, depends on child A (127). Makes `smaqit-deployment` the executor of the resolved
slot plan: `smaqit.input-deployment` obtains the plan from `smaqit slots plan`, the deployment agent
walks it in Steps 5–6 by reading each filler's SKILL.md from its resolved path, and `smaqit.feature-new`
Phase 3 stops calling infrastructure skills directly and routes them through the agent's slots. This is
the part a waiting downstream platform repository needs to bind its own fillers.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Claude Code, GitHub Copilot, Codex CLI, Agent Skills
**Platforms/Environments:** Claude Code, GitHub Copilot, Codex CLI
**Features/Integrations:** subagents (Task tool), custom agent tool lists, AGENTS.md project instructions
**Versions/Constraints:** smaqit 3.6.1, Copilot deployment agent currently has no read-file tool, Claude deployment subagent has no Skill tool (by design: SKILL.md is read, not invoked)

## Design Decisions

- Inherits task 125's Design Decisions (invocation = read SKILL.md from the resolved path; `existing-cicd-pr` refuses `iac-generate`/`ci-generate`/`provision`/`host-bootstrap` bindings; invariants bind every filler).
- **Single resolution per deployment:** `input-deployment` does not re-invoke the CLI when the caller already passed a slot plan (removes feature-new's current double invocation).

## Implementation Steps

1. `skills/smaqit.input-deployment/SKILL.md` — after resolving `provisioning_mode`, call `smaqit slots plan` and output the slot plan; stop on `SLOTS-*`; skip when a plan was passed in. Bump `metadata.version`.
2. `agents/deployment.md` — Execution Parameters take `slot_plan`; Steps 5–6 walk it in order (read the filler's SKILL.md at its resolved path and follow it; skip `none`; `ci`-executor slots are monitored, not run locally); single-slot day-2 entry; slot resolution recorded in the deployment report; guard-script / no-out-of-band / credential-references invariants restated as binding every filler.
3. `.smaqit/definitions/agents/deployment.frontmatter.yaml` — add a read-file tool to the Copilot tools list.
4. `skills/smaqit.feature-new/SKILL.md` Phase 3 — vault-loader/repo-config preflight becomes the agent's `access-credentials` / `target-credentials` / `environment-config` slots; `existing-cicd-pr` profile and refusals apply. Bump `metadata.version`.
5. Regenerate (`make -C installer prepare`) and confirm the compiled Claude/Copilot/Codex deployment agents carry the changes.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `smaqit.input-deployment` outputs the slot plan via the CLI, stops on `SLOTS-*`, and is invoked once per deployment
- [ ] `agents/deployment.md` executes the slot plan in Steps 5–6, supports a single-slot day-2 run, records the resolution in the deployment report, and keeps the guard/out-of-band/credential invariants for every filler
- [ ] The compiled Copilot deployment agent has a read-file tool
- [ ] `feature-new` Phase 3 routes infrastructure steps through the slot plan; under `existing-cicd-pr`, bindings to `iac-generate`/`ci-generate`/`provision`/`host-bootstrap` are refused

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
| `skills/smaqit.input-deployment/SKILL.md` | Modify |
| `agents/deployment.md`, `.smaqit/definitions/agents/deployment.frontmatter.yaml` | Modify |
| `skills/smaqit.feature-new/SKILL.md` | Modify |

## Notes

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
