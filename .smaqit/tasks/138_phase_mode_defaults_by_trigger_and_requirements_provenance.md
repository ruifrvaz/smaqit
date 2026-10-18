---
status: Completed
created: "2026-10-18"
mode: Assisted
started: "2026-10-18"
completed: "2026-10-18"
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
**Triaged:** 2026-10-18
**Tools searched:** Claude Code (`anthropics/claude-code`), GitHub Copilot (`github/copilot-docs`), OpenAI Codex (`openai/codex`)
**Result:** Advisory

### Advisory Issues
- [#82456 Undocumented system-prompt directive "Do not call the AgentTool unless the user requested it" contradicts documented description-based subagent delegation](https://github.com/anthropics/claude-code/issues/82456) — `anthropics/claude-code` — opened 2026-07-30 — area:agents, stale. Bears directly on model-started phases: Claude may decline to delegate to a "Use proactively" phase subagent without being asked. The task routes self-starts through the model-invocable command as well as the subagent, and the e2e `self-invocation` and `title-only` scenarios observe which path actually fires; don't rely on proactive delegation alone.
- [#77795 Skills with disable-model-invocation: true are silently ignored when invoked via agent prompts](https://github.com/anthropics/claude-code/issues/77795) — `anthropics/claude-code` — opened 2026-07-15 — bug, platform:macos, area:agents, area:skills. Not hit: the task keeps the phase commands model-invocable and adds no `disable-model-invocation`. Relevant only if that option is revisited.
- [#95730 Background subagents give the user zero visibility; model waits silently for 15+ minutes](https://github.com/anthropics/claude-code/issues/95730) — `anthropics/claude-code` — opened 2026-09-20 — enhancement, platform:macos, area:agents, area:agent-view. Same family as the known detached-subagent behavior (#75043). Assisted-by-default means more `REVIEW-REQUESTED` round-trips through background phase agents; the e2e `assisted` scenario's resume loop is the check.
- [#101519 Every subagent is told "Do NOT Write report/summary/findings/analysis .md files" even when its task prompt asks for exactly that](https://github.com/anthropics/claude-code/issues/101519) — `anthropics/claude-code` — opened 2026-10-12 — bug, area:agents. Pre-existing and not changed by this task, but the phase agents write `.smaqit/reports/*-phase-report-*.md` as subagents; if the e2e `report_exists` check starts failing, check this first.
- [#46048 Codex CLI subagent completes, but main thread reports "No agents completed yet"](https://github.com/openai/codex/issues/46048) — `openai/codex` — opened 2026-09-16 — bug, CLI, subagent. Affects relaying `NEEDS-INPUT`/`REVIEW-REQUESTED` from Codex phase subagents, which assisted-by-default makes more frequent. Not reproducible in this repo's tests (no Codex e2e).
- [#54896 Injected subagent instructions incorrectly claim full-history forks always inherit the parent model and effort](https://github.com/openai/codex/issues/54896) — `openai/codex` — opened 2026-10-17 — documentation, CLI, subagent, config. Unrelated to mode or provenance; recorded for completeness of the subagent surface.

### Historical (Closed)
- [#31893 GPT-5.6 forces a restricted spawn_agent schema that makes documented custom agents unusable](https://github.com/openai/codex/issues/31893) — `openai/codex` — closed 2026-07-09. Fixed; Codex named custom agents spawn as documented.

### Search Warnings
- `github/copilot-docs` — the helper resolved GitHub Copilot to its documentation repository; its open and closed searches returned no issues, which is weak evidence about Copilot's custom-agent runtime (the VS Code extension's tracker was not searched).

## Acceptance Criteria

- [x] With no stated mode, a user-started phase entry point (typed `/smaqit.<phase>`, or a Copilot/Codex agent the user picked) resolves `mode: assisted`; a model-started phase and greenfield/feature-new resolve `autonomous`; a mode stated in the request always wins
- [x] In assisted mode each layer pauses with `REVIEW-REQUESTED` and each phase's final output asks the user to review it and names the next phase's entry point; the stale Workflow Handover targets are corrected
- [x] In autonomous mode greenfield and feature-new stop only for `NEEDS-INPUT` and human operations (PR merges, synthesized-skill checkpoint, credential prompts); Phase 0 confirmation, inventory check, spec-set approval, and validation sign-off are assisted-only
- [x] `PHASE-PARAMETERS` carries `user_request` verbatim; phase agents pass `user_request`, `answers`, and `mode` unchanged to specification agents
- [x] Layer gates count only user-stated requirements (`user_request`, `answers`, user-pointed files including `.smaqit/requirements-extract.md`); project/directory names, model world knowledge, and orchestrator paraphrase never satisfy a required section in either mode; autonomous also counts upstream specs; assisted requires each layer's required sections to trace to the user's own statements, with upstream specs for cross-layer coherence only
- [x] No specification or phase agent directive permits filling a required section with an assumption; Coverage's inferred risk-profile clause is removed, and Stack's "agent's recommendation" counts only when the user states it
- [x] e2e-claude has `empty-request` (bare `/smaqit.development`) and `title-only` (plain self-invoked "Let's build mario-app" in a `mario-app` project) scenarios, each asserting `NEEDS-INPUT` and no files under `specs/`; `assisted` passes without asking for assisted mode; all six scenarios pass live
- [x] `framework/AGENTS.md`, `framework/PHASES.md`, `templates/AGENTS.md.template` (heading `## When to use smaqit` unchanged), README, wiki, and compendium describe the new defaults; `make prepare`, `make test`, `make smoke-test` pass; CHANGELOG has an action-required note

## Findings

**Implementation approach:**
- Mode now defaults by trigger in the three phase input skills: assisted when the user started the phase (typed command, picked agent, or a phase agent invoked with no block), autonomous when the model or an orchestrator did; a stated mode always wins and the mode question was dropped
- Added a verbatim `user_request` field to `PHASE-PARAMETERS`; phase commands, phase agents, and both orchestrators pass it, plus `mode`, unchanged to every specification agent
- Added a shared Requirement Sources section to all five layer gates (user request, answers, the user's own messages, user-pointed files including `requirements-extract.md`; names, model knowledge, and paraphrase never count) with a mode-dependent upstream-spec rule; spec agents' Session Context matches it
- Replaced every assumption-filling directive in spec and phase agents and in `framework/AGENTS.md` with a `NEEDS-INPUT` rule; removed Coverage's inferred risk-profile clause and tightened Stack's "agent's recommendation"
- Greenfield and feature-new Phase 0 default to Autonomous, with task-set confirmation, inventory check, spec-set approval, and validation sign-off marked assisted-only
- Assisted phases end with a review prompt naming the next phase command; fixed the stale development and deployment handover targets
- e2e-claude gained `empty-request` and `title-only` scenarios (with `new_project` directory and README seeding); `assisted` now tests the default, and `happy-path`/`needs-input` state "run autonomously"
- Verified with `make prepare`, `make test`, `make smoke-test`, and greps of the generated Copilot, Claude, and Codex artifacts

**Decisions made:**
- feature-new still starts its parent task in Assisted task mode in both execution modes: the task lifecycle's Autonomous mode self-merges the task PR, which is the release, and a PR merge is a human operation (deviates from the plan's "task-start mode follows Phase 0")
- A phase agent invoked without a parameter block defaults to assisted, so a Claude self-delegation that forgets its block fails toward an extra review pause
- Greenfield's "structural divergences need operator approval" stop stays in autonomous runs: it fires when code cannot follow an approved spec, which is closer to `NEEDS-INPUT` than to a review gate
- Business, the entry layer, traces required sections to the user in either mode; only downstream layers count upstream specs in autonomous mode
- The older `.smaqit/definitions/skills/smaqit.new-greenfield-project.md` keeps `mode: autonomous` on its MVP task, which now matches the orchestrator default; only its spec-set gate was marked assisted-only
- `scripts/e2e/log_check.py` was left unchanged: its existing `signal` and `delegated` checks cover both new scenarios
- `qa.md` keeps its assumption directives; it answers questions and generates no specifications
- The `AGENTS.md` template no longer claims the spec agents are directly invocable on Copilot (they are `user-invocable: false`)

**Blockers encountered:**
- None. The live `make -C installer e2e-claude` run ran after the PR opened, at the user's direction, and all six scenarios passed on Claude Code 2.1.296 (about $5.19 in total): `empty-request` and `title-only` stopped at the Business gate with `NEEDS-INPUT` and wrote nothing; `assisted` paused per layer on the default; `self-invocation`, `happy-path`, and `needs-input` resolved autonomous and completed

**Follow-up identified:**
- A `scripts/generate-agents.py` check that every phase command and agent states the trigger-based default, as input-skill reachability is checked today
- claude-code #82456 (an undocumented directive against unrequested Agent-tool use) did not bite: `title-only` routed by invoking the `/smaqit.development` command as a skill, and `self-invocation` delegated; keep watching it, since the command path is now the one model-started phases rely on

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
