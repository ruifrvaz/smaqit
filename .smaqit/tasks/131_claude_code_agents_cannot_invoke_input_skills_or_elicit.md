---
status: Not Started
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
**Technologies:** Claude Code
**Platforms/Environments:** Claude Code CLI, Claude Code VS Code extension
**Features/Integrations:** subagents (Task/Agent tool), Agent Skills, subagent `skills:` preload frontmatter, custom slash commands, AskUserQuestion
**Versions/Constraints:** smaqit 3.6.1, subagents never receive AskUserQuestion, a `tools:` allowlist without `Skill` blocks skill invocation, subagent spawn depth limit 3

## Design Decisions

TBD — to be confirmed during assessment. A proposed direction, not yet approved, is in Notes.

## Implementation Steps

1. **Verify the current behavior** in a scratch smaqit-initialized project on Claude Code:
   - Run `/smaqit.development` with deliberately thin requirements, and record whether `smaqit-business` attempts `input-business` and what it does instead.
   - Ask for assisted mode, and record whether the run actually pauses.
   - Make a subagent return a question and re-invoke it, to confirm or refute the draft-spec regeneration hypothesis.
   - Check how Copilot and Codex agents reach their input skills, to decide whether the fix is Claude-only.
2. **Decide the split** between (a) input skills that run in the main conversation, inside the command or calling skill, before delegation, and (b) gates preloaded into the agent with `skills:` that return structured questions instead of asking. Record the decision.
3. **Phase input skills** (`input-development`, `input-deployment`, `input-validation`): move the call to wherever step 2 decides, and pass the resolved parameters (mode, `provisioning_mode`) to the phase subagent explicitly. Coordinate `input-deployment` with task 128, which changes how it is called and what it outputs.
4. **Spec-layer gates** (`input-business`, `-functional`, `-stack`, `-infrastructure`, `-coverage`): make them reachable on Claude Code (e.g. `skills:` in each `claude:` frontmatter entry). Define the "gaps found → return questions, don't guess, don't generate" contract in the agent bodies.
5. **Return-and-resume loop:** define how a phase run continues after the user answers questions returned by a gate. Fix draft-spec regeneration if step 1 confirms it, for example by recording which specs this run generated, or by resuming the same subagent instead of starting a new one.
6. **Assisted mode on Claude Code:** either implement review pauses as return-and-resume, or document that assisted mode isn't available inside Claude subagents.
7. **Framework text:** update `framework/AGENTS.md` (line 14 and the input-skill bullet near line 138) and `framework/SKILLS.md` so the invocation rule describes what the platform can actually do.
8. Regenerate with `make -C installer prepare`, then run `make -C installer test` and `make -C installer smoke-test`. Add a check that every skill an agent body tells it to invoke is reachable from that agent's Claude frontmatter.
9. CHANGELOG entry and release.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] Step 1's findings are recorded: whether the gates run today, whether assisted mode pauses, whether re-invocation regenerates draft specs, and how Copilot and Codex reach input skills
- [ ] On Claude Code, every input skill an agent body tells it to invoke is reachable from that agent, or runs in the main conversation before delegation, and a test or generator check enforces this
- [ ] A spec gate that finds insufficient requirements returns its questions to the main conversation instead of guessing or generating a spec, and the main conversation asks the user and resumes
- [ ] Resuming after answered questions does not regenerate specs the same run already produced
- [ ] The phase parameters (execution mode, `provisioning_mode`) resolved by a phase input skill reach the phase subagent explicitly on Claude Code
- [ ] Assisted mode on Claude Code either works through return-and-resume or is documented as unavailable
- [ ] `framework/AGENTS.md` and `framework/SKILLS.md` describe input-skill invocation consistently with the platform mechanics
- [ ] `make -C installer test` and `make -C installer smoke-test` pass

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
| `.smaqit/definitions/agents/*.frontmatter.yaml` (`claude:` entries) | Modify — skill reachability (`skills:` and/or `Skill`) |
| `agents/{business,functional,stack,infrastructure,coverage,development,deployment,validation}.md` | Modify — input-skill invocation, return-questions contract, resume behavior |
| `commands/smaqit.{development,deployment,validation}.md` | Modify — run phase input skills or the return-and-resume loop in the main conversation, if chosen |
| `skills/smaqit.input-*/SKILL.md` | Modify — "elicit" step behavior when running inside a subagent |
| `scripts/generate-agents.py` | Modify — emit `skills:`, if not already passed through; reachability check |
| `framework/AGENTS.md`, `framework/SKILLS.md` | Modify — invocation rule |
| `installer/main_test.go` or a generator check | Modify — reachability guard |
| `CHANGELOG.md` | Modify |

## Notes

- **Proposed direction (not yet approved):**
  - Phase input skills run in the main conversation before delegation, because their questions (mode, `provisioning_mode`) must reach the user.
  - Spec-layer gates are preloaded with `skills:` and return their questions instead of asking.
  - The command file owns the ask-and-resume loop.
- **Overlaps with task 130** (directive instructions and trigger descriptions, from a parallel session). One of 130's design decisions says "No `disable-model-invocation` on `smaqit.input-*`: … phase/spec agents invoke these skills that way". On Claude Code the agents currently can't invoke skills at all, so that rationale rests on this task's fix. 130 also rewrites the four phase command descriptions and the Claude phase-agent frontmatter, which this task may also touch. Reconcile whichever lands second.
- **Overlap with task 128** (deployment slots B): it changes `smaqit.input-deployment` to call `smaqit slots plan` and to be invoked once per deployment. This task's decision about where `input-deployment` runs has to fit that.
- Claude Code docs, checked 2026-10-11:
  - "To preload Skills into context, use the `skills` field rather than listing `Skill` here."
  - "To prevent a subagent from invoking skills entirely, omit `Skill` from the `tools` list."
  - With `skills:`, "the full content of each listed skill is injected into the subagent's context at startup".
  - A subagent "can spawn subagents of its own, up to three layers below the main conversation".
  - `AskUserQuestion` is removed from subagents even when listed.
