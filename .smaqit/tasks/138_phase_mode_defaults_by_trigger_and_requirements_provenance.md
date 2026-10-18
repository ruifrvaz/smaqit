---
status: Not Started
created: "2026-10-18"
---

# Phase Mode Defaults by Trigger (Commands Assisted, Skills Autonomous) and Requirements Provenance

## Description

Direct phase runs silently default to `autonomous`: `smaqit.input-development` (and its deployment and validation siblings) defaults `mode` to `autonomous` and asks only when the request already mentions review, so a bare `/smaqit.development` runs Business → Functional → Stack → code with no pause. The end-to-end orchestrators (`smaqit.new-greenfield-project`, `smaqit.feature-new`) have the opposite problem: their autonomous mode is defined as "without gate interruptions", yet their Phase 0, spec-set approval, and validation sign-off gates are written unconditionally.

Separately, the layer input gates accept inferred requirements. Step 1 "Extract from session context" plus lenient "Sufficient when" thresholds let the model count a project or directory name, its own world knowledge, or the orchestrating conversation's paraphrase of the request as user requirements. Spec and phase agents also carry pre-task-131 directives ("Flag assumptions explicitly when clarification is unavailable", "Document assumptions when specs are underspecified") that license filling gaps with assumptions. A blank project titled "mario-app" can therefore trigger a highly opinionated development cycle.

The approved model: smaqit has human-triggered entry points (the three phase commands, or the phase agents a user picks on Copilot/Codex), which default to **assisted**, and automated runs (the greenfield and feature-new skills, and phases the model starts on its own), which default to **autonomous**. A mode stated in the request always wins. Assisted means each layer works only from what the user stated, pauses for review, and each phase ends by asking the user to review it and naming the next phase. Autonomous means best effort end to end, stopping only for `NEEDS-INPUT` or human operations. In both modes a gate never treats a title, a name, model knowledge, or a paraphrase as requirements.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Claude Code, GitHub Copilot, OpenAI Codex
**Platforms/Environments:** None
**Features/Integrations:** Claude Code slash commands as model-invocable skills, Claude Code subagents (proactive delegation), Copilot custom agents (runSubagent, askQuestions), Codex custom agents, headless claude -p stream-json
**Versions/Constraints:** behavior-changing defaults allowed (MINOR with action-required note), `## When to use smaqit` heading verbatim (smoke-test and e2e grep it), skill/command descriptions ≤400 chars

## Design Decisions

- **Mode default by trigger:** a user-started phase entry point → `assisted`; orchestrator skills (greenfield, feature-new) and model-started phases → `autonomous`; a mode stated in the request always wins. The mode question is dropped.
- **Assisted semantics:** user-stated input only, `REVIEW-REQUESTED` after each layer, and a review prompt naming the next phase's entry point at phase end.
- **Autonomous orchestrators** stop only for `NEEDS-INPUT` and human operations (PR merges, the synthesized-skill checkpoint, credential prompts, the release PR). Phase 0 confirmation, greenfield's inventory check, spec-set approval, and validation sign-off become assisted-only.
- **No start confirmation:** a `CONFIRM-START` signal was considered and dropped. Self-invoked phases run autonomous; the provenance baseline is their guard.
- **No-block fallback:** a phase agent invoked without a `PHASE-PARAMETERS` block assumes the user picked it and defaults to `assisted`, so a Claude self-delegation that forgets its block fails safe (one extra review pause).
- **Verbatim request:** `PHASE-PARAMETERS` carries `user_request` with the user's words unchanged; phase agents pass `user_request`, `answers`, and `mode` unchanged to specification agents, replacing "user requirements from session context".
- **Empty request:** blocks only when a layer must be generated; the layer gate decides (implementing already-current specs with no new request still works).
- **Provenance baseline (both modes):** required sections are satisfied only by `user_request`, `answers`, and files the user pointed at (including `.smaqit/requirements-extract.md`). Project/directory names, model world knowledge, and orchestrator paraphrase never count; a title alone is never sufficient.
- **Mode-dependent upstream rule:** in autonomous mode upstream specs also count as requirement sources; in assisted mode each layer's required sections must trace to the user's own statements, and upstream specs serve cross-layer coherence only. Coherence rules (`framework/AGENTS.md:159`) are unchanged.
- **Assumptions:** only required sections are protected; optional sections may use a stated default that is listed in the spec. `qa.md` is out of scope (it answers questions and generates no specs).
- **Self-routing kept:** task 130's "Use proactively" phase-subagent descriptions and model-invocable commands stay.

## Implementation Steps

1. **Mode contract.** In `skills/smaqit.input-{development,deployment,validation}/SKILL.md`: default `mode` by trigger (assisted when the user started the entry point, autonomous when started automatically or passed by an orchestrator, a stated mode always wins), drop the mode question, and add a verbatim `user_request` to the `PHASE-PARAMETERS` output and example block (`mode: assisted`). In `commands/smaqit.{development,deployment,validation}.md`: tell the input skill whether the user typed the command or the model invoked it, pass `user_request` verbatim, and echo the resolved block before delegating.
2. **Phase agents** (`agents/{development,deployment,validation}.md`):
   - Execution Parameters: add `user_request`; no-block default is assisted (the user picked the agent).
   - Execution Mode: Assisted becomes the default text (dev `:121-130`, deploy `:149-158`, validation `:131-140`).
   - Scoped context to spec agents (dev `:168`, deploy `:198`, validation `:178`): pass `user_request`, `answers`, and `mode` unchanged.
   - Replace "Document assumptions when specs are underspecified" (dev `:87`, deploy `:114`, validation `:98`).
   - Assisted phase end: the final output asks the user to review the phase's output and names the next entry point.
   - Fix Workflow Handover targets: development → `/smaqit.deployment` (not the `smaqit.infrastructure` agent), deployment → `/smaqit.validation` (not `smaqit.coverage`, `:435`).
3. **Self-started phases.** `templates/AGENTS.md.template` "When to use smaqit": a phase you start on your own runs `autonomous` and passes the request verbatim; fix line 49's claim that Copilot spec agents are "directly invocable" (they are `user-invocable: false`). In `.smaqit/definitions/agents/{development,deployment,validation}.frontmatter.yaml` Claude and Codex descriptions: a direct delegation passes a block with `mode: autonomous` and `user_request`.
4. **Orchestrators.** `skills/smaqit.new-greenfield-project/SKILL.md` and `skills/smaqit.feature-new/SKILL.md`:
   - Phase 0 defaults to Autonomous unless the user asks for assisted.
   - Make assisted-only: Phase 0 "Operator confirms mode and approves the task set" (greenfield `:62`, feature-new `:59`), greenfield Phase 1 inventory check (`:69`), spec-set approval (greenfield `:83`, feature-new `:76`), validation sign-off (greenfield `:171`, feature-new `:174`).
   - feature-new's task-start mode (`:42`, "in Assisted mode by default") follows Phase 0.
   - Greenfield Phase 2 passes `.smaqit/requirements-extract.md` as the requirements source.
   - Align `.smaqit/definitions/skills/smaqit.new-greenfield-project.md:45` (hard-coded `mode: autonomous`).
5. **Provenance.**
   - Layer gates `skills/smaqit.input-{business,functional,stack,infrastructure,coverage}/SKILL.md`: step 1 sources, the baseline exclusions (including "a title alone is never sufficient"), the mode-dependent upstream rule, Stack's "agent's recommendation" only when user-stated (`:24-25`), remove Coverage's inferred risk-profile clause (`:25`), bump each version.
   - Spec agents `agents/{business,functional,stack,infrastructure,coverage}.md`: Role line 7 and Session Context match the gates; assisted when no mode is passed; replace "Flag assumptions explicitly when clarification is unavailable" (business `:73`, functional `:85`, stack `:89`, infrastructure `:100`, coverage `:75`) with a MUST NOT: never fill a required section with an assumption, return `NEEDS-INPUT`; optional sections may use a stated default listed in the spec.
   - `framework/AGENTS.md`: Fail-Fast bullet (`:29-33`), Session Context (`:9-17`, `:153-159`), Phase Parameters table (`:267`, mode default by trigger, `user_request`), Execution Mode (`:280`), implementation SHOULD (`:429`). `framework/PHASES.md`: context-sufficiency lines (`:57-58`, `:100-101`, `:180-181`).
6. **Tests.** `scripts/e2e-claude.sh`:
   - Add `empty-request` and `title-only` in all five hard-coded scenario-list sites (usage comment `:14`, default `:30`, validation case `:81-87`, dispatch loop `:305-316`).
   - `new_project` (`:232-239`) takes an optional directory name and README seed.
   - `empty-request`: bare `/smaqit.development` → `NEEDS-INPUT`, `no_specs`.
   - `title-only`: project dir `mario-app`, `README.md` `# mario-app`, plain prompt "Let's build mario-app" with no command → `NEEDS-INPUT`, `no_specs`.
   - `assisted`: drop the "Use assisted mode…" sentence and expect `REVIEW-REQUESTED` by default.
   - `happy-path` and `needs-input`: prompts state "Run autonomously, with no review pauses."; `self-invocation` unchanged (autonomous, one turn).
   - `scripts/e2e/log_check.py`: docstring note only.
7. **Docs and release.** `README.md:53`, `docs/wiki/workflows/quickstart.md:53` (commands review each layer by default; say "autonomous" to run straight through), `docs/wiki/workflows/testing-smaqit.md` (scenario table, count, fixture claim), `docs/wiki/agent-tools-reference.md` (entry paths, defaults by trigger), `docs/wiki/designs/fail-fast-on-ambiguity.md` (assumption section), `.smaqit/compendium.md` (input skills, asking questions, self-invocation, e2e entries), `CHANGELOG.md` MINOR with an action-required note.
8. **Verify.** `make -C installer prepare`, `make -C installer test`, `make -C installer smoke-test`; grep generated `installer/agents-*`/`skills-*` for the trigger-based default wording, `user_request`, no remaining "Flag assumptions"/"Document assumptions" in spec and phase agents, no unconditional "User reviews and approves"/"User signs off" in the orchestrators; `SCENARIOS=empty-request,title-only make -C installer e2e-claude`, then all six scenarios.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] With no stated mode, a user-started phase entry point (typed `/smaqit.<phase>`, or a Copilot/Codex agent the user picked) resolves `mode: assisted`; a model-started phase and greenfield/feature-new resolve `autonomous`; a mode stated in the request always wins
- [ ] In assisted mode each layer pauses with `REVIEW-REQUESTED` and each phase's final output asks the user to review it and names the next phase's entry point; the stale Workflow Handover targets are corrected
- [ ] In autonomous mode greenfield and feature-new stop only for `NEEDS-INPUT` and human operations (PR merges, synthesized-skill checkpoint, credential prompts); Phase 0 confirmation, inventory check, spec-set approval, and validation sign-off are assisted-only
- [ ] `PHASE-PARAMETERS` carries `user_request` verbatim; phase agents pass `user_request`, `answers`, and `mode` unchanged to specification agents
- [ ] Layer gates count only user-stated requirements (`user_request`, `answers`, user-pointed files including `.smaqit/requirements-extract.md`); project/directory names, model world knowledge, and orchestrator paraphrase never satisfy a required section in either mode; autonomous also counts upstream specs; assisted requires each layer's required sections to trace to the user's own statements, with upstream specs for cross-layer coherence only
- [ ] No specification or phase agent directive permits filling a required section with an assumption; Coverage's inferred risk-profile clause is removed, and Stack's "agent's recommendation" counts only when the user states it
- [ ] e2e-claude has `empty-request` (bare `/smaqit.development`) and `title-only` (plain self-invoked "Let's build mario-app" in a `mario-app` project) scenarios, each asserting `NEEDS-INPUT` and no files under `specs/`; `assisted` passes without asking for assisted mode; all six scenarios pass live
- [ ] `framework/AGENTS.md`, `framework/PHASES.md`, `templates/AGENTS.md.template` (heading `## When to use smaqit` unchanged), README, wiki, and compendium describe the new defaults; `make prepare`, `make test`, `make smoke-test` pass; CHANGELOG has an action-required note

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
| skills/smaqit.input-{development,deployment,validation}/SKILL.md | Modify |
| skills/smaqit.input-{business,functional,stack,infrastructure,coverage}/SKILL.md | Modify |
| commands/smaqit.{development,deployment,validation}.md | Modify |
| agents/{development,deployment,validation}.md | Modify |
| agents/{business,functional,stack,infrastructure,coverage}.md | Modify |
| .smaqit/definitions/agents/{development,deployment,validation}.frontmatter.yaml | Modify |
| .smaqit/definitions/skills/smaqit.new-greenfield-project.md | Modify |
| templates/AGENTS.md.template | Modify |
| skills/smaqit.new-greenfield-project/SKILL.md, skills/smaqit.feature-new/SKILL.md | Modify |
| framework/AGENTS.md, framework/PHASES.md | Modify |
| scripts/e2e-claude.sh, scripts/e2e/log_check.py | Modify |
| README.md, docs/wiki/workflows/{quickstart,testing-smaqit}.md, docs/wiki/agent-tools-reference.md, docs/wiki/designs/fail-fast-on-ambiguity.md | Modify |
| .smaqit/compendium.md, CHANGELOG.md | Modify |

## Notes

- Origin: session 095 assessment of the phase triggers, after a blank project titled "mario-app" triggered a highly opinionated autonomous development cycle.
- Evidence at planning time: `commands/smaqit.development.md`, `skills/smaqit.input-development/SKILL.md:21-26`, `agents/business.md:73`, `agents/development.md:87,123,130`, `skills/smaqit.input-stack/SKILL.md:25`, `skills/smaqit.input-coverage/SKILL.md:25`, `templates/AGENTS.md.template`, `scripts/e2e-claude.sh`.
- The existing e2e `needs-input` scenario ("Build me a small website about Mario.") already returns `NEEDS-INPUT`; the gap is the bare request, the title-only context, and the orchestrator's paraphrase, none of which any scenario covers today.
- No Go test asserts any of the wording this task changes; `scripts/smoke-test-installer.sh` and `scripts/e2e-claude.sh` grep only the `## When to use smaqit` heading. `scripts/generate-agents.py` enforces the 400-char description limit and input-skill reachability.
- `installer/main.go` `Version` is stale (`3.1.0`) relative to CHANGELOG; it is set via ldflags at build time, so it is not part of this task.
- Follow-up candidate: a `generate-agents.py` check that every phase command and agent states the trigger-based default.

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
