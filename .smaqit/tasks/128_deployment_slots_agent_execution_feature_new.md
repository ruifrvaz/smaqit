---
status: Completed
completed: "2026-10-11"
mode: Assisted
started: "2026-10-11"
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
**Triaged:** 2026-10-11
**Tools searched:** Claude Code (`anthropics/claude-code`), GitHub Copilot (`github/copilot-cli`), Codex CLI (`openai/codex`)
**Result:** Advisory

### Advisory Issues
- [#88987 Task/general-purpose subagent fabricates "verbatim" file quote instead of returning actual file content](https://github.com/anthropics/claude-code/issues/88987) — `anthropics/claude-code` — labels: bug, area:model, area:agents, stale. The agent must act on the filler SKILL.md it reads, not on a paraphrase; record the resolved SKILL.md path in the report.
- [#40278 followup_task resets a full-access subagent to read-only/on-request](https://github.com/openai/codex/issues/40278) — `openai/codex` — labels: bug, sandbox, app, subagent. A resumed Codex deployment subagent may lose write/exec access mid slot plan; affects filler execution, not plan resolution.

### Historical (Closed)
- [#3688 Repository-level custom agents resolved relative to git root, but skills and .mcp.json relative to cwd](https://github.com/github/copilot-cli/issues/3688) — `github/copilot-cli` — closed 2026-09-30. `smaqit slots plan` resolves project-local skills from the project root, independent of client cwd behaviour.
- [#81440 Subagents in plan mode can read files outside the working dir](https://github.com/anthropics/claude-code/issues/81440) — `anthropics/claude-code` — closed 2026-10-01. Global filler SKILL.md paths live outside the project; reading them is expected.

### Omitted Tools
- Agent Skills — specification only (covered by task 127's triage)

## Acceptance Criteria

- [x] `smaqit.input-deployment` outputs the slot plan via the CLI, stops on `SLOTS-*`, and is invoked once per deployment
- [x] `agents/deployment.md` executes the slot plan in Steps 5–6, supports a single-slot day-2 run, records the resolution in the deployment report, and keeps the guard/out-of-band/credential invariants for every filler
- [x] The compiled Copilot deployment agent has a read-file tool
- [x] `feature-new` Phase 3 routes infrastructure steps through the slot plan; under `existing-cicd-pr`, bindings to `iac-generate`/`ci-generate`/`provision`/`host-bootstrap` are refused

## Findings

**Implementation approach:**
- `agents/deployment.md` takes a `slot_plan` (or a single `slot`); Step 5 resolves or accepts the plan and stops on `SLOTS-*`; Step 6 walks it: local fillers are read from `skillPath` and followed, families pick a stack-matching candidate, builtins cover IaC generation and guarded destroy, `ci` executors are monitored
- The `existing-cicd-pr` PR/merge/monitor/verify sequence is unchanged in substance and now starts at the first `ci`-executor slot; `verify` runs the plan's filler with `--expected-sha`
- `smaqit.input-deployment` gained Environment and Deployment Path parameters and a Slot Plan step wrapping `smaqit slots plan --json`
- `smaqit.feature-new` Phase 3 resolves the plan in preflight and passes it to the agent; its direct vault-loader and repo-config steps became the agent's credential and `environment-config` slots
- Copilot deployment agent frontmatter gained `read/readFile`; regenerated agents for all three platforms carry the slot text

**Decisions made:**
- The agent pauses and hands the operator the exact command when a filler needs secret input, since credential slots now run inside the subagent; secrets never pass through chat
- Family slots with no fitting candidate return to the caller, which keeps match-or-synthesize ownership in the orchestrating skill
- `status: deployed` is not set by a day-two run that changes no deployed revision
- The feature-new gotcha about `smaqit.input-deployment` now forbids changing its `provision` default rather than modifying the skill at all

**Blockers encountered:**
- None

**Follow-up identified:**
- Child 129 adds `deployment_path: ci-push` to the agent and routes greenfield Phases 4–6 through it
- The deployment-slots bench (parent close-out) should cover that the agent reads the bound filler's SKILL.md rather than a paraphrase (claude-code#88987)

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.input-deployment/SKILL.md` | Modify |
| `agents/deployment.md`, `.smaqit/definitions/agents/deployment.frontmatter.yaml` | Modify |
| `skills/smaqit.feature-new/SKILL.md` | Modify |

## Notes

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
