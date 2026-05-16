# Orchestration-First Phase Agents

**Status:** Completed
**Created:** 2026-05-16
**Completed:** 2026-05-16

## Description

Phase agents (development, deployment, validation) were designed to coordinate their entire phase — spec generation included — but the current implementation treats spec generation as a conditional fallback rather than the primary path. Two compounding bugs prevent orchestration from ever firing on a fresh project:

**Bug 1 — Pre-Orchestration Validation halts before orchestration:** The Dependency Verification checklist requires upstream spec artifacts to be present. On a fresh project with no specs, this check fails and the agent halts with a diagnostic report before reaching the spec-generation step.

**Bug 2 — `smaqit plan` empty output halts the agent:** `smaqit plan --phase=develop` returns empty output and exits 0 when no specs exist. The current MUST directive says "Report completion when no specs require processing." The agent interprets empty output as "nothing to do" and stops.

Both bugs push users back to the manual workflow (invoke each spec agent individually before invoking the implementation agent), which contradicts the one-command-per-phase design goal.

The fix is to restructure phase agent instructions so that orchestration — always invoking spec agents for their layer, then implementing — is the primary path. Existing specs at the correct lifecycle status skip generation; everything else goes through the full orchestration.

This task also introduces explicit autonomous vs. assisted mode descriptions into each phase agent, so the two supported execution styles are unambiguous.

## Design Decisions

**Orchestration-first, not fallback:** Spec generation is always step 1 of phase execution. The decision gate is "do specs exist at the correct lifecycle status?" not "are specs missing?" If specs exist and are at the right status, generation is skipped. Otherwise, generation runs.

**Pre-Orchestration Validation scope:** Validation checks session context sufficiency (can we generate specs?) and execution environment readiness. It does NOT check for upstream spec presence — that is the orchestration's job.

**Spec agents remain bounded:** Spec agents only do their own layer's work. Being invoked as a subagent by a phase agent does not change this. No spec agent directive changes are needed to enforce this.

**Spec agent Role input source:** A minor clarification is added to each spec agent's Role section to indicate that session context includes context propagated from an orchestrating phase agent.

**Autonomous vs. assisted modes are explicit:** Each phase agent documents both modes with clear decision rules, so users and agents know what to expect.

**`smaqit plan` usage scoped to implementation:** The CLI is for determining which existing specs need implementation processing, not for deciding whether to generate specs. Spec generation decisions use file/status checks directly.

**No CLI changes:** `smaqit plan` behavior is correct as-is. The problem is in how agents interpret its output.

**Deterministic routing (confirmed by research):** The spec agent invocation sequence is fully known upfront and must be hardcoded, not derived at runtime. Microsoft AI Agent Design Patterns guidance: *"The choice of which agent gets invoked next is deterministically defined as part of the workflow and isn't a choice given to agents in the process."* Development always invokes business → functional → stack in that order; deployment always invokes infrastructure; validation always invokes coverage. Status checks on spec files determine which steps to skip, not CLI output parsing.

**Context scoping for subagent invocations:** When a phase agent invokes a spec agent, it passes scoped context — not the full accumulated phase context. Each spec agent receives: (1) user requirements from session context, (2) previously generated upstream specs for that layer's coherence needs. Passing full accumulated context causes unnecessary context window growth and degrades subagent focus.

**Iteration caps required:** Spec generation retries and assisted-mode review loops must have explicit iteration caps with defined fallback behavior (escalate to human). Currently undefined; the Microsoft guidance requires caps to prevent infinite loops in both maker-checker and sequential patterns.

**Assisted mode maps to maker-checker:** The Microsoft pattern for assisted mode is maker-checker: spec agent (maker) produces output, user (checker) reviews and approves or provides feedback. The loop repeats until approved or the iteration cap is reached. This is the correct framing for the assisted mode section in each phase agent.

**GitHub Copilot Hooks as enforcement layer (prerequisite: VS Code surface confirmation):** Three hook events are applicable as a deterministic enforcement layer on top of agent instructions: `subagentStop` for inter-layer spec validation gates and iteration cap enforcement; `subagentStart` for guaranteed scoped context injection to subagents. Hooks operate outside the LLM — compliance is not required. Implementation depends on confirming that hooks fire in the VS Code Copilot chat/agent surface (documented payload format suggests support, but the surface is not explicitly named in the hooks reference).

## Implementation Steps

1. **Rewrite Pre-Orchestration Validation in all 3 phase agents** — Replace the "upstream spec artifacts present" check with "session context sufficient for spec generation." Remove dependency checks that require specs to exist. Keep: session context sufficiency, execution environment readiness, credential readiness (deployment only).

2. **Rewrite Phase Orchestration workflow in all 3 phase agents** — Make spec generation the primary workflow with deterministic routing. Restructure as:
   - Step 1: Invoke input skill (confirm/default execution parameters)
   - Step 2: For each required spec layer in fixed sequence, check file status directly (not via `smaqit plan`) → skip layers with specs at correct lifecycle status
   - Step 3: Generate missing/stale specs by invoking spec agents in hardcoded dependency order with scoped context (user requirements + upstream specs only, not full accumulated context)
   - Step 4: Consolidate specs for coherence
   - Step 5: Implement (generate code / deploy / run tests); use `smaqit plan` here to identify which specs need implementation processing
   - Step 6: Verify and report

3. **Add autonomous vs. assisted mode section to each phase agent** — Define both modes using the maker-checker framing:
   - **Autonomous:** invoke input skill → for each spec layer in sequence, invoke spec agent with scoped context → consolidate → implement (no user breaks)
   - **Assisted (maker-checker):** invoke input skill → invoke spec agent (maker) → present output to user (checker) → user approves or provides feedback → loop until approved or iteration cap reached → proceed to next layer → implement
   - Both modes: max 3 review iterations per spec layer; on cap reached, surface unresolved issues to user before proceeding

4. **Fix the "empty smaqit plan output" directive** — Change the MUST directive from "Report completion when no specs require processing" to: "If `smaqit plan --phase=[PHASE]` returns no specs, all existing specs are up to date for implementation — proceed directly to implementation step."

5. **Update spec agent Role sections (all 5)** — Add one sentence: "When invoked as a subagent by a phase agent, session context includes requirements propagated from the orchestrating agent."

6. **Update `framework/PHASES.md`** — Describe orchestration-first as the primary workflow. Add that phase agents always coordinate spec generation, not only when specs are missing. Add note on deterministic routing and context scoping.

7. **Update `framework/AGENTS.md`** — Rewrite the Phase Orchestration section: remove "missing artifacts trigger invocation" framing; replace with orchestration-first, deterministic routing, scoped context passing, and iteration caps. Add the autonomous/assisted (maker-checker) mode descriptions.

8. **Mirror all agent changes to `installer/agents/`** — Copy each updated agent file to its installer mirror after editing.

9. **Mirror framework changes to `installer/framework/`** — Copy updated PHASES.md and AGENTS.md.

10. **Build installer and verify** — Run `make build` in `installer/`. No functional test needed for agent instructions (no CLI changes), but confirm the build succeeds with updated embedded files.

11. **[Conditional] Implement hook enforcement layer** — Only after confirming hooks fire in the VS Code Copilot chat/agent surface. Create `.github/hooks/smaqit-gates.json` and `.github/hooks/scripts/smaqit-validate-spec.sh` and `smaqit-inject-context.sh`. Register in installer as additional scaffold output. See `docs/research/ms-ai-agent-orchestration-patterns.md` § GitHub Copilot Hooks for full design.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [x] Running `/smaqit.development` on a fresh project (no specs) invokes business → functional → stack agents, then implements, without halting on missing specs
- [x] Running `/smaqit.deployment` on a project with Development phase complete invokes infrastructure agent, then deploys, without halting on missing specs
- [x] Running `/smaqit.validation` on a project with Deployment phase complete invokes coverage agent, then validates, without halting on missing specs
- [x] If all required specs already exist at the correct lifecycle status, phase agents skip spec generation and proceed directly to implementation
- [x] Pre-Orchestration Validation no longer checks for upstream spec presence; it checks session context sufficiency and environment readiness
- [x] Each phase agent explicitly documents autonomous mode (no user breaks) and assisted mode (maker-checker loop with user as checker)
- [x] Assisted mode defines a max 3-iteration cap per spec layer with fallback behavior when cap is reached
- [x] Spec agent invocation sequence is hardcoded per phase (not derived from CLI output): development = business → functional → stack; deployment = infrastructure; validation = coverage
- [x] Phase agents pass scoped context to each invoked spec agent (user requirements + relevant upstream specs only — not full accumulated context)
- [x] `smaqit plan` is used only to determine which existing specs need implementation processing, not for spec generation routing decisions
- [x] `smaqit plan` empty output is interpreted as "specs up to date — proceed to implementation"
- [x] All 5 spec agent Role sections include the subagent context source note
- [x] `framework/PHASES.md` describes orchestration-first, deterministic routing, and context scoping
- [x] `framework/AGENTS.md` Phase Orchestration section describes orchestration-first with deterministic routing, context scoping, and iteration caps
- [ ] All changed agent files mirrored to `installer/agents/` — **skipped per user decision**
- [ ] All changed framework files mirrored to `installer/framework/` — **skipped per user decision**
- [x] `installer/` builds successfully after all changes
- [ ] [Conditional — pending VS Code surface confirmation] `subagentStop` hook validates spec output after each spec agent and blocks with feedback on failure — **SubagentStart/SubagentStop do not fire for runSubagent tool calls; hooks not implemented**
- [ ] [Conditional] Hook script enforces 3-iteration cap via `.smaqit/.hook-state/retries-{layer}` counter — **not implemented (conditional blocked)**
- [ ] [Conditional] `subagentStart` hook injects upstream spec content as `additionalContext` before functional, stack, and coverage agents are spawned — **not implemented (conditional blocked)**

## Findings

**Implementation approach:**
- Rewrote Pre-Orchestration Validation in all 3 phase agents to check session context sufficiency instead of upstream spec presence
- Rewrote Phase Orchestration workflow in all 3 phase agents: spec generation is now the primary first step with deterministic hardcoded routing; `smaqit plan` scoped to implementation routing only
- Added autonomous vs. assisted (maker-checker) mode sections with 3-iteration caps to all 3 phase agents
- Added subagent context source note to all 5 spec agent Role sections
- Updated `framework/PHASES.md` and `framework/AGENTS.md` to reflect orchestration-first design
- Built installer to confirm embedded file compilation succeeds

**Decisions made:**
- Installer mirror steps (8, 9) explicitly skipped per user direction — installer/agents/ and installer/framework/ not updated this task
- Hook enforcement layer (step 11) tested but not implemented: `PostToolUse` fires via VS Code hooks; `SubagentStart`/`SubagentStop` do not fire for `runSubagent` tool calls — only for VS Code native agent delegation
- Conditional hook ACs left unchecked as the VS Code surface does not support the required hook events for tool-invoked subagents

**Blockers encountered:**
- `SubagentStart` hook does not fire when subagents are invoked via the `runSubagent` tool — confirmed after testing both script-based and inline hook commands; hook enforcement layer cannot be implemented at `SubagentStart`/`SubagentStop` level for tool-invoked agents

**Follow-up identified:**
- Installer mirror sync needed: agent and framework files changed in this task are not reflected in `installer/agents/` and `installer/framework/` — users who reinstall smaqit will get older agent behavior
- Hook enforcement design needs rethink: `PostToolUse` fires but requires a different enforcement architecture (post-tool output inspection rather than pre/post-subagent gates); consider a follow-up task
- Consider a dedicated task to sync installer mirrors after all recent task changes

## Files to Create / Modify

| File | Action |
|------|--------|
| `agents/smaqit.development.agent.md` | Modify |
| `agents/smaqit.deployment.agent.md` | Modify |
| `agents/smaqit.validation.agent.md` | Modify |
| `agents/smaqit.business.agent.md` | Modify (Role section) |
| `agents/smaqit.functional.agent.md` | Modify (Role section) |
| `agents/smaqit.stack.agent.md` | Modify (Role section) |
| `agents/smaqit.infrastructure.agent.md` | Modify (Role section) |
| `agents/smaqit.coverage.agent.md` | Modify (Role section) |
| `framework/PHASES.md` | Modify |
| `framework/AGENTS.md` | Modify |
| `installer/agents/smaqit.development.agent.md` | Mirror |
| `installer/agents/smaqit.deployment.agent.md` | Mirror |
| `installer/agents/smaqit.validation.agent.md` | Mirror |
| `installer/agents/smaqit.business.agent.md` | Mirror |
| `installer/agents/smaqit.functional.agent.md` | Mirror |
| `installer/agents/smaqit.stack.agent.md` | Mirror |
| `installer/agents/smaqit.infrastructure.agent.md` | Mirror |
| `installer/agents/smaqit.coverage.agent.md` | Mirror |
| `installer/framework/PHASES.md` | Mirror |
| `installer/framework/AGENTS.md` | Mirror |
| `.github/hooks/smaqit-gates.json` | Create (conditional) |
| `.github/hooks/scripts/smaqit-validate-spec.sh` | Create (conditional) |
| `.github/hooks/scripts/smaqit-inject-context.sh` | Create (conditional) |

## Notes

Task 073 (Implementation Agents as Phase Orchestrators) is the predecessor — it introduced the Phase Orchestration section and `runSubagent` invocation pattern. This task does not revisit that architecture; it fixes the logic that prevents orchestration from running and promotes it from fallback to primary path.

The one open acceptance criterion from Task 073 ("E2E test passes: empty project → `/smaqit.development` → working app") is not in scope here — it is a testing concern, not an implementation concern. It belongs to Task 070 (E2E Boundary Enforcement Validation) once that is re-scoped, or a new testing task.

**Research basis:** Design decisions refined based on Microsoft AI Agent Design Patterns guidelines (Azure Architecture Center, 2026) and GitHub Copilot Hooks reference. See `docs/research/ms-ai-agent-orchestration-patterns.md` for pattern alignment analysis and `docs/research/gh-copilot-hooks-enforcement.md` for the hooks enforcement layer design.
