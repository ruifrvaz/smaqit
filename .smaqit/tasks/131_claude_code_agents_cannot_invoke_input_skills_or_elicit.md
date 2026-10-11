---
status: In Progress
mode: Assisted
started: "2026-10-11"
created: "2026-10-11"
---

# Claude Code Agents Can't Invoke Their Input Skills or Ask the User Questions

## Description

Every smaqit agent body tells the agent to invoke an input skill before doing any work. Spec agents must run `smaqit.input-[layer]` to check that requirements are sufficient (`framework/AGENTS.md:14`, a MUST). Phase agents run `smaqit.input-[phase]` to confirm execution parameters. On Claude Code, none of the agents can do this:

1. **No Skill tool.** Each Claude agent has a fixed `tools:` list in `.smaqit/definitions/agents/*.frontmatter.yaml`, for example `Read, Write, Edit, Grep, Glob, Bash, WebFetch, TodoWrite, Task` plus the PlantUML MCP tools for spec agents. `Skill` is not on any of them, and no agent uses the `skills:` frontmatter field. The Claude Code docs say that leaving `Skill` out of `tools:` stops a subagent from invoking skills at all, and that `skills:` is how a skill's full content gets injected into a subagent at startup. So "Invoke `smaqit.input-business`" has no mechanism behind it: the agent skips the gate or improvises it from memory. The docs don't describe this case, so that outcome is inferred.
2. **No way to ask the user.** The core of every input skill is "ask the targeted question for each gap". Claude Code removes `AskUserQuestion` from all subagents, even when it is listed in `tools:`. Spec agents run two levels deep (command → phase subagent → spec subagent), so even with access to the skill, a spec agent could only return a question, never ask one.
3. **What this breaks:**
   - The requirements gates (`input-business`, `-functional`, `-stack`, `-infrastructure`, `-coverage`) don't run on Claude Code.
   - Assisted mode can't be selected through `input-development`/`input-validation`/`input-deployment`, so phases always run autonomously.
   - Assisted mode's "pause after each spec agent for user review" can't happen inside a subagent anyway.
   - Under a direct `/smaqit.deployment` call, `provisioning_mode` resolution depends on improvisation. Only `smaqit.feature-new` and `smaqit.new-greenfield-project` invoke `input-deployment` reliably, because those skills run in the main conversation, where the Skill tool exists.

A related hypothesis, not yet verified: if a phase subagent returns with a question and is re-invoked, `agents/development.md` Step 2 treats any spec with `status: draft` as missing. A spec generated earlier in the same run is still `draft`, so the spec agents would be invoked again and might regenerate it.

Found in session 089 (2026-10-11) during the smaqit product Q&A. The Claude Code facts above were confirmed against code.claude.com/docs/en/sub-agents.md and skills.md on 2026-10-11.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Claude Code, GitHub Copilot, Codex CLI
**Platforms/Environments:** Claude Code CLI, Claude Code VS Code extension, VS Code with GitHub Copilot custom agents
**Features/Integrations:** subagents (Task/Agent tool), subagent `skills:` preload frontmatter, SendMessage subagent resume, AskUserQuestion, Copilot runSubagent, Copilot vscode/askQuestions, Agent Skills, custom slash commands
**Versions/Constraints:** smaqit 3.7.0, subagents never receive AskUserQuestion, a `tools:` allowlist without `Skill` blocks skill invocation, subagent spawn depth limit 3, a subagent resuming its own nested subagent is undocumented

## Design Decisions

All confirmed by the user on 2026-10-11, during `task.plan`.

- **Platform scope: a platform-neutral contract plus Claude and Copilot frontmatter.**
  - The ask-or-return contract lives in the shared agent bodies, so every platform gets it.
  - Claude gets `skills:` preloads. Copilot gets the missing `vscode/askQuestions`.
  - Codex gets the shared contract only, with no frontmatter change and no live verification.
- **Phase parameters: the caller passes them, and the agent falls back.**
  - Command bodies and orchestrator skills run the phase input skill in the main conversation and pass an explicit parameter block to the phase agent.
  - A phase agent invoked without a block applies its preloaded phase-skill defaults, and returns `NEEDS-INPUT` only for a blocking ambiguity. This keeps task 130's direct, proactive delegation safe.
- **Spec gates: preloaded, passing `NEEDS-INPUT` up the chain.**
  - Each spec agent preloads its `smaqit.input-[layer]` skill with `skills:`.
  - When the gate finds gaps, the agent writes nothing and returns `NEEDS-INPUT`. The phase agent passes it up unchanged.
  - The main conversation asks the user, then resumes the *phase* agent by sending its agent ID the answers (`SendMessage`), which the Claude Code docs describe.
  - The phase agent invokes the spec agent again with the answers. This avoids the undocumented case of a subagent resuming its own child, and nothing gets regenerated because nothing was written.
- **`approved_spec_paths` parameter:** a phase agent treats the listed paths as current even while their status is `draft`. This makes a fresh re-invocation safe whenever resume isn't available.
- **Assisted mode: review by returning and resuming.**
  - The phase input skills gain a `mode` parameter, default Autonomous.
  - In Assisted mode, a phase agent that can't ask directly returns `REVIEW-REQUESTED` with the spec and design paths after each layer.
  - Approval adds the paths to `approved_spec_paths`. Feedback re-invokes the spec agent, with a maximum of 3 iterations as today.
- **The greenfield regeneration bug is in scope.** `smaqit.new-greenfield-project` Phase 3 passes `specification_mode: prevalidated` with the Phase 2 spec paths, as `smaqit.feature-new` already does.
- **Spec-agent calls in orchestrators are in scope.** Both orchestrator skills stop calling `/smaqit.business` … `/smaqit.coverage`, which aren't commands on Claude Code. They invoke the `smaqit.<layer>` specification agent in platform-neutral wording and handle the same signals.
- **Ownership split with task 130:** 130 owns `description`/`when_to_use` wording. 131 owns the command bodies and the agents' `tools:`/`skills:` fields.
- **Compatibility with task 125 (released in v3.7.0):** `input-deployment`'s existing `slot_plan` output, plus `environment` and `deployment_path`, become part of the parameter block. Slot resolution and the Phases 4–6 slot routing are not changed.
- **Delivery:** a standalone task on its own branch and worktree from `origin/main`, with its own `Prepare release vX.Y.Z` PR. The version is chosen at merge.
- **L0 edits are direct,** following task 125's precedent. Record this in Findings.
- **Out of scope:**
  - Codex live verification and Codex frontmatter.
  - Changing the slot catalogue or slot resolution.
  - A bench case for the `NEEDS-INPUT` contract, which is a follow-up.

## Implementation Steps

*Phase A: the contract, in the framework (L0, direct edit)*

1. `framework/AGENTS.md`: define **Phase Parameters** and **Return Signals**.
   - **Parameter block:** `mode`, `specification_mode`, `spec_paths`, `approved_spec_paths`, `deployment_path`, `provisioning_mode`, `environment`, `slot_plan`, and `answers` on resume.
   - **`NEEDS-INPUT` signal:** the source agent or skill, its layer, a list of questions (each with an id, the question and why it's needed), and a statement that nothing was written.
   - **`REVIEW-REQUESTED` signal:** the layer, the spec and design paths, and the iteration count out of 3.
   - Each signal ends with a one-line instruction to its receiver ("ask the user these, then resume me with the answers"), so it also works under direct delegation.
   - Rewrite the "Invoke input skill" rule (line 14) and the input-skill bullet (~138): input skills are *loaded* by the agent, and their questions are *asked* by whoever can reach the user.
   - Rewrite **Execution Modes** (after the slot section added by task 125): mode is a parameter, and Assisted means returning `REVIEW-REQUESTED` after each layer when the agent can't ask directly.
2. `framework/SKILLS.md` Skill Invocation Context (~105–126): add the rule "ask if you have a user-question tool; otherwise return `NEEDS-INPUT` and write nothing".

*Phase B: input skills (depends on A)*

3. `skills/smaqit.input-{business,functional,stack,infrastructure,coverage}/SKILL.md`:
   - Step 5 "Elicit gaps" becomes "ask directly if you can; otherwise stop before generating and return `NEEDS-INPUT` listing every gap at once".
   - Step 2's "confirm add vs revise" becomes a `NEEDS-INPUT` question as well.
   - Bump `metadata.version`.
4. `skills/smaqit.input-{development,deployment,validation}/SKILL.md`:
   - Add a `mode` parameter: default Autonomous, elicited only when the context mentions review or approval.
   - Add an explicit **Output** section: the parameter block the caller passes in.
   - In `input-deployment`, replace "surface `provisioning_mode` … in session context" with the block, which already carries the `slot_plan`, environment and deployment path from task 128.
   - Bump `metadata.version`.

*Phase C: agents (depends on A; can run in parallel with B)*

5. `agents/{business,functional,stack,infrastructure,coverage}.md`:
   - Input step: run the preloaded gate. When it finds gaps, return `NEEDS-INPUT` without writing files.
   - When the input contains `answers`, treat them as requirements.
   - Failure Handling: "Ambiguous input → request clarification" maps to `NEEDS-INPUT` when the agent can't ask.
6. `agents/{development,deployment,validation}.md`, **Execution Parameters**: read the parameter block. If none was passed, apply the preloaded phase-skill defaults and return `NEEDS-INPUT` only for a blocking ambiguity.
   - `deployment.md` keeps its existing "unless the caller already passed a `slot_plan`" rule.
7. `agents/{development,deployment,validation}.md`, **Step 2:**
   - Pass a spec agent's `NEEDS-INPUT` up unchanged.
   - On resume with `answers`, invoke that spec agent again.
   - Treat paths in `approved_spec_paths` as current even while `draft`.
8. `agents/{development,deployment,validation}.md`, **Assisted mode:**
   - Return `REVIEW-REQUESTED` after each layer.
   - On resume, approval adds the paths to `approved_spec_paths`, and feedback re-invokes the spec agent (max 3).
   - Fix "Mode is set by the `smaqit.input-*` skill" (`development.md:122`, `deployment.md:148`, `validation.md:132`): mode comes from the parameter block.
9. `.smaqit/definitions/agents/*.frontmatter.yaml`:
   - **Claude:** add `skills: [smaqit.input-<own layer or phase>]` to all 8 `claude:` entries. `generate-agents.py` passes extra keys through unchanged. Leave `description` to task 130.
   - **Copilot:** add `vscode/askQuestions` to development, deployment and validation. Confirm functional, stack, infrastructure and coverage have it, as business does, and add it where missing.

*Phase D: callers, which own the loop (depends on A–C)*

10. `commands/smaqit.{development,deployment,validation}.md` body:
    - Run the phase input skill in the main conversation.
    - Invoke the subagent with the parameter block plus the user's request.
    - On `NEEDS-INPUT` or `REVIEW-REQUESTED`, ask with the user-question tool, then resume the same subagent by sending its agent ID the `answers`. If resume isn't available, invoke it again with the same parameters plus `answers` and `approved_spec_paths`.
    - Repeat until a final result arrives, then relay it.
    - Leave the `description` frontmatter to task 130.
11. `skills/smaqit.new-greenfield-project/SKILL.md`:
    - Phase 2 (lines 75–79): replace `/smaqit.business` … `/smaqit.coverage` with "invoke the `smaqit.<layer>` specification agent", handling `NEEDS-INPUT`.
    - Phase 3 (line 87): pass `specification_mode: prevalidated` with the Phase 2 spec paths.
    - Pass the Phase 0 `mode` into every phase-agent call, including Phases 4–6, without changing their slot logic.
    - Bump `metadata.version`.
12. `skills/smaqit.feature-new/SKILL.md`:
    - Phase 1 (line 72): the same platform-neutral spec-agent wording.
    - Pass `mode` into the Phase 2, 3 and 4 agent calls.
    - Phase 3 already passes the `input-deployment` output and `slot_plan`; keep it as is.
    - Bump `metadata.version`.

*Phase E: guard, docs, verification, release*

13. `scripts/generate-agents.py`: add a reachability check. Every `smaqit.input-*` an agent body tells it to run must appear in that agent's Claude `skills:` list; otherwise fail `make prepare` and name the file.
14. Docs:
    - `docs/wiki/agent-tools-reference.md`: the signal contract and the loop per platform. Also fix the stale project-level Codex paths (lines ~151/153).
    - `.smaqit/compendium.md`: a new entry, "how smaqit agents ask the user questions on each platform".
    - `README.md`: one line under Build something.
15. Run `make -C installer prepare test smoke-test`, then the `grep` checks from Verification.
16. Live check in Claude Code on a scratch `smaqit init` project, recording the results in Findings:
    - Thin requirements, which should produce `NEEDS-INPUT` → answer → resume → spec written.
    - Assisted mode, which should produce `REVIEW-REQUESTED` after each layer.
    - No spec agent re-runs for approved specs.
17. CHANGELOG entry and a `Prepare release vX.Y.Z` PR, with the version chosen at merge.

## Known Issues Triage
**Triaged:** 2026-10-11
**Tools searched:** Claude Code (`anthropics/claude-code`), GitHub Copilot (`github/copilot-docs`), Codex CLI (`openai/codex`)
**Result:** Advisory

### Advisory Issues
- [#100813 Claude can't see your skills: `/skills` says they're loaded, but the listing reaches the model late or never (2.1.295)](https://github.com/anthropics/claude-code/issues/100813) — `anthropics/claude-code` — opened 2026-10-09 — bug, platform:macos, area:skills. Affects the main-thread skill listing, not `skills:` preload. Still relevant: the command bodies run the phase input skills in the main conversation.
- [#93171 Custom subagent silently missing from Agent tool's available list despite valid, sibling-identical frontmatter](https://github.com/anthropics/claude-code/issues/93171) — `anthropics/claude-code` — opened 2026-09-09 — bug, platform:windows, platform:vscode, area:agents. A spec or phase subagent that isn't listed can't be invoked; frontmatter edits must stay valid.
- [#97849 AskUserQuestion invoked with placeholder content while waiting on background agent](https://github.com/anthropics/claude-code/issues/97849) — `anthropics/claude-code` — opened 2026-09-28 — bug, platform:windows, area:agents. Relevant to the main-conversation ask-and-resume loop. The loop should ask only after the subagent's result has arrived.
- [#96561 Cloud session: AskUserQuestion asked right after SendUserMessage is dropped on worker resume, so a late answer is silently lost](https://github.com/anthropics/claude-code/issues/96561) — `anthropics/claude-code` — opened 2026-09-24 — bug, has repro, area:core, area:cowork. Cloud sessions only.
- [#89835 Plan mode's subagent reminder tells background subagents to write a plan and call AskUserQuestion, which they cannot do](https://github.com/anthropics/claude-code/issues/89835) — `anthropics/claude-code` — opened 2026-08-26 — no labels. Corroborates that subagents can't ask.
- [#93477 Feature request: suppress skill_listing/agent_listing_delta for a subagent spawn whose declared tools don't include Skill/Agent](https://github.com/anthropics/claude-code/issues/93477) — `anthropics/claude-code` — opened 2026-09-10 — enhancement, area:cost, area:agents
- [#51685 Add a per-skill main-agent-only scope that excludes ordinary subagents](https://github.com/openai/codex/issues/51685) — `openai/codex` — opened 2026-10-07 — enhancement, app, skills, subagent
- [#16475 Subagents do not inherit parent session's registered MCP tools](https://github.com/openai/codex/issues/16475) — `openai/codex` — opened 2026-04-01 — bug, mcp, subagent. Concerns Codex, which is out of scope here.

### Historical (Closed)
- [#80489 Messaging a running subagent re-injects its full frontmatter skills: bodies and re-fires SubagentStart hooks — undocumented and token-costly](https://github.com/anthropics/claude-code/issues/80489) — `anthropics/claude-code` — closed 2026-10-10 (stale). Directly relevant: resuming a phase agent with `SendMessage` may re-inject its preloaded input skill. Preload only the agent's own input skill to keep that cost small.
- [#67251 Agent skills: frontmatter field does not inject skill content at startup](https://github.com/anthropics/claude-code/issues/67251) — `anthropics/claude-code` — closed 2026-07-20 (stale). Verify during the live check that `skills:` preload actually injects content.
- [#87536 disable-model-invocation: true still hides skills from model entirely (regression from #43660)](https://github.com/anthropics/claude-code/issues/87536) — `anthropics/claude-code` — closed 2026-08-19. Bears on task 130's decision to keep model invocation enabled.
- [#101169 ${CLAUDE_EFFORT} in a skill preloaded into a subagent expands to the session effort, not the subagent's effort](https://github.com/anthropics/claude-code/issues/101169) — `anthropics/claude-code` — closed 2026-10-11
- [#82240 context: fork skills dispatch to background agent with no task/instructions](https://github.com/anthropics/claude-code/issues/82240) — `anthropics/claude-code` — closed 2026-10-08

### Search Warnings
- `github/copilot-docs` open/closed searches returned no items. The helper resolved GitHub Copilot to a documentation repository rather than to the VS Code custom-agent tracker, so Copilot coverage is limited. This is not a Clear result.

## Acceptance Criteria

- [ ] `framework/AGENTS.md` defines the Phase Parameters block and the `NEEDS-INPUT`/`REVIEW-REQUESTED` return signals. Its input-skill rule and Execution Modes are rewritten to match. `framework/SKILLS.md` states the ask-or-return rule. The direct L0 edit is recorded in Findings.
- [ ] The five layer gates return `NEEDS-INPUT` and write no files when they can't ask the user. The three phase input skills define `mode` and an Output parameter block. No agent still says "Mode is set by the … skill".
- [ ] All 8 Claude agent entries preload their own input skill with `skills:`. Copilot development, deployment and validation, plus any spec agent that lacked it, have `vscode/askQuestions`.
- [ ] Phase agents:
  - read the parameter block
  - fall back to defaults plus `NEEDS-INPUT` when no block was passed
  - pass a spec agent's `NEEDS-INPUT` up unchanged
  - honour `approved_spec_paths`
  - implement Assisted mode as `REVIEW-REQUESTED`, keeping the 3-iteration cap
- [ ] The three phase commands run the phase input skill in the main conversation, pass the parameter block, and loop ask-and-resume. They resume through `SendMessage`, falling back to a fresh invocation with `answers` and `approved_spec_paths`.
- [ ] `new-greenfield-project` Phase 3 passes `specification_mode: prevalidated` with the Phase 2 spec paths. Neither orchestrator skill calls spec agents as `/smaqit.<layer>` commands, and both pass `mode` into every phase-agent call.
- [ ] `make -C installer prepare` fails, naming the file, when an input skill an agent body tells it to run is missing from that agent's Claude `skills:` list.
- [ ] `make -C installer test` and `make -C installer smoke-test` pass. A Claude Code live check records three results: the `NEEDS-INPUT` round trip, the Assisted `REVIEW-REQUESTED` round trip, and no re-run of approved specs.
- [ ] `docs/wiki/agent-tools-reference.md` (including its stale Codex paths), the compendium and `README.md` describe the contract. `CHANGELOG.md` is updated.

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
| `framework/AGENTS.md`, `framework/SKILLS.md` | Modify — Phase Parameters, Return Signals, input-skill rule, Execution Modes |
| `skills/smaqit.input-*/SKILL.md` (8) | Modify — ask-or-return gates; `mode` + Output block on the phase skills |
| `agents/{business,functional,stack,infrastructure,coverage}.md` | Modify — run the gate, `NEEDS-INPUT`, `answers` |
| `agents/{development,deployment,validation}.md` | Modify — parameter block, signal handling, `approved_spec_paths`, Assisted as `REVIEW-REQUESTED` |
| `.smaqit/definitions/agents/*.frontmatter.yaml` | Modify — Claude `skills:`; Copilot `vscode/askQuestions` |
| `commands/smaqit.{development,deployment,validation}.md` | Modify — body: run the phase input skill, then the ask-and-resume loop |
| `skills/smaqit.new-greenfield-project/SKILL.md`, `skills/smaqit.feature-new/SKILL.md` | Modify — spec-agent wording, `prevalidated` in greenfield Phase 3, `mode` passed through |
| `scripts/generate-agents.py` | Modify — reachability check |
| `docs/wiki/agent-tools-reference.md`, `.smaqit/compendium.md`, `README.md` | Modify — docs |
| `CHANGELOG.md` | Modify |

## Notes

- **Discovery (`task.plan`, 2026-10-11)**, confirmed by reading the code:
  - **No input skill defines a mode,** so phase-agent Assisted mode is unreachable on every platform. Neither orchestrator passes a mode to any agent.
  - **Draft-spec re-run is real.** Each phase agent's Step 2 re-invokes a spec agent for any `draft` spec. `new-greenfield-project` Phase 3 calls the development agent without `prevalidated` right after Phase 2 drafted every spec.
  - **There is no parameter channel.** The command bodies pass no parameters. `input-deployment` was the only skill that said where its output went ("session context"), which a fresh subagent can't see.
  - **Copilot:** only `business` has `vscode/askQuestions`. Task 128 has since added `read/readFile` to Copilot `deployment`.
  - **Generator:** any key under `claude:` passes through `generate_agent`/`dump_frontmatter` unchanged. Nothing checks compiled agent or command frontmatter today.
  - **Claude Code docs:** a returned subagent can be resumed with its full history by sending its agent ID a message (`SendMessage`). A subagent resuming its own nested subagent is undocumented.
- **Overlaps with task 130** (directive instructions and trigger descriptions, from a parallel session). One of 130's design decisions says "No `disable-model-invocation` on `smaqit.input-*`: … phase/spec agents invoke these skills that way". On Claude Code the agents currently can't invoke skills at all, so that rationale rests on this task's fix. 130 also rewrites the four phase command descriptions and the Claude phase-agent frontmatter, which this task may also touch. Reconcile whichever lands second.
  - **Ownership split, agreed with task 130's session on 2026-10-11:** 130 owns the `description`/`when_to_use` wording of the commands and the Claude phase agents. 131 owns the command bodies and the agents' `tools:`/`skills:` fields. 130's rationale now states that the agents can't currently invoke input skills.
  - **Bypass risk:** 130 adds "use proactively" to the Claude phase subagent descriptions, so Claude may delegate to a phase subagent directly and never run the command. If 131 puts elicitation or the ask-and-resume loop in the command body, a direct delegation skips it. The fix must also work, or degrade explicitly, when the phase subagent is invoked directly. Options include the subagent returning its unresolved parameters and questions as its result, or moving the loop out of the command body.
- **Overlap with task 128** (deployment slots B): it changes `smaqit.input-deployment` to call `smaqit slots plan` and to be invoked once per deployment. This task's decision about where `input-deployment` runs has to fit that.
- Claude Code docs, checked 2026-10-11:
  - "To preload Skills into context, use the `skills` field rather than listing `Skill` here."
  - "To prevent a subagent from invoking skills entirely, omit `Skill` from the `tools` list."
  - With `skills:`, "the full content of each listed skill is injected into the subagent's context at startup".
  - A subagent "can spawn subagents of its own, up to three layers below the main conversation".
  - `AskUserQuestion` is removed from subagents even when listed.
