# Subagent Elicitation Return Signals

## Metadata

- **Date:** 2026-10-11
- **Session focus:** Product Q&A on how smaqit runs on Claude Code. It found that agents running as subagents can't run their input skills or ask the user questions. The fix was planned, implemented and released.
- **Tasks completed:** 131 (released v3.8.0, PR #95)
- **Tasks created:** 131; smaqit-adk 037 (Bench harness selector for Claude Code)
- **Tasks referenced:** 125, 127, 128, 129 (deployment slots, completed by a parallel session during this one), 130 (parallel session), 114 (started by another session)

## Actions Taken

- **Session start:** loaded project docs, memory, history 088, `PLANNING.md` and the compendium. While mapping task 127's code, found that a parallel session ("125: variation points") was already implementing it in the task-125 worktree. Stayed read-only there and answered later questions from `main`.
- **Product Q&A, three questions:**
  - **How smaqit is configured to run:** three phases (develop/deploy/validate), each run by a phase agent that orchestrates its spec agents.
  - **How to set up a new project:** `install.sh` installs agents and skills globally, then `smaqit init` adds `.smaqit/`, `specs/`, `.mcp.json`, and `AGENTS.md`/`CLAUDE.md`.
  - **The develop-only entry point:** `/smaqit.development` → Task → `smaqit-development`, which runs the business, functional and stack spec agents.
- **Claude Code facts,** checked against the official docs through a claude-code-guide subagent:
  - Subagents nest up to 3 levels deep.
  - `AskUserQuestion` is removed from every subagent.
  - A `tools:` list without `Skill` blocks skill invocation entirely; `skills:` preloads a skill's full content at startup.
  - The main conversation can resume a subagent that has returned by sending its agent ID a message (`SendMessage`).
- **Input-skill finding:** traced every `smaqit.input-*` reference. No Claude agent can reach its input skill, and no subagent can ask the user, so on Claude Code the requirements gates never ran. Filed task 131.
- **Coordination with task 130:** a parallel session had filed task 130, whose rationale assumed agents can invoke skills. Sent it a cross-session note. It corrected 130 and agreed an ownership split: 130 owns `description`/`when_to_use`; 131 owns command bodies and `tools:`/`skills:`. It also flagged the risk of direct-delegation bypass, which was added to 131.
- **Planning task 131 (`task.plan`):** three Explore subagents covered the generator, the input-skill touchpoints, and Copilot/Codex. Discovery confirmed three more problems:
  - Assisted mode was unreachable on every platform, because no input skill defined a mode.
  - Greenfield regenerated approved draft specs.
  - Copilot phase agents lacked `vscode/askQuestions`.

  Two rounds of Q&A then settled the design.
- **Starting the task:** after task 125 shipped v3.7.0, started 131 from `origin/main`. The research-map block and issue triage came back Advisory.
- **Implementation in five reviewed phases:**
  - **A:** the L0 contract
  - **B:** the input skills
  - **C:** the agents plus Claude/Copilot metadata
  - **D:** the command loop and orchestrator skills
  - **E:** the generator guard, docs, CHANGELOG and tests
- **Bench question:** found that Bench can't validly exercise Claude Code yet. Personal skills shadow treatment skills, an isolated config directory needs credentials, and there's no multi-turn support. Filed smaqit-adk task 037 for a harness selector.
- **Completion:** with the user's approval, the live Claude Code check was deferred to the next real use. Completed 131 as PR #95, `Prepare release v3.8.0`. After the user merged it, ran Phase 2 cleanup and confirmed the release workflow published v3.8.0 with 5 binaries.

## Problems Solved

- **Input gates never ran on Claude Code.** Every agent now preloads its own input skill through `skills:`, and `generate-agents.py` fails `make prepare` when a named input skill is unreachable.
- **Subagents can't ask the user.** A platform-neutral contract replaces asking: agents return `NEEDS-INPUT` (writing nothing) or `REVIEW-REQUESTED`, orchestrators pass these up unchanged, and the main conversation asks and resumes.
- **Assisted mode was unreachable.** `mode` is now a phase input-skill parameter carried in an explicit `PHASE-PARAMETERS` block.
- **Greenfield regenerated specs.** Phases 3–7 hand the Phase 2 spec set on as `prevalidated`. Phases 4–7 were added to the scope after the user reviewed the change.
- **`/smaqit.<layer>` calls in orchestrators:** those commands don't exist on Claude Code, so both orchestrators now invoke spec agents by name on every platform.

## Decisions Made

- **Platform-neutral contract:** it lives in the shared agent bodies. Claude gets `skills:`, Copilot gets `vscode/askQuestions`, and Codex gets the contract only.
- **Caller passes parameters, agent falls back:** a phase agent without a parameter block applies its input skill's defaults. This keeps task 130's proactive direct delegation safe.
- **Bubble, don't nest-resume:** spec gates write nothing on `NEEDS-INPUT`. The main conversation resumes the *phase* agent, a documented mechanism, and the phase agent invokes the spec agent again.
- **`approved_spec_paths`:** makes a fresh re-invocation safe when resume isn't available.
- **Preload only each agent's own input skill,** because resume re-injects preloaded skills (closed claude-code issue #80489).
- **L0 edits were direct,** following task 125's precedent.
- **Version:** v3.8.0, a MINOR release (new contract, backward compatible). The `Version` constant in `installer/main.go` stays at 3.1.0, matching recent releases.
- **Live check deferred** by the user to actual use of smaqit, rather than a hand-rolled Bench run.

## Files Modified

- **Task 131, shipped in PR #95:**
  - `framework/AGENTS.md`: User Interaction, Return Signals, Phase Parameters, input-skill rule, Execution Modes
  - `framework/SKILLS.md`: rules for reaching the user and reaching the skill
  - `skills/smaqit.input-*/SKILL.md` (8): ask-or-return gates; `mode` and the Output block on the phase skills
  - `agents/*.md` (9): `NEEDS-INPUT` and `answers` in the spec agents; parameter block, `approved_spec_paths` and `REVIEW-REQUESTED` in the phase agents; a QA line
  - `.smaqit/definitions/agents/*.frontmatter.yaml` (8): Claude `skills:`; Copilot `vscode/askQuestions` on 5 agents
  - `commands/smaqit.{development,deployment,validation}.md`: ask-and-resume loop
  - `skills/smaqit.new-greenfield-project/SKILL.md` (2.1.0) and `skills/smaqit.feature-new/SKILL.md` (2.3.0)
  - `scripts/generate-agents.py`: `check_input_skill_reachability`
  - `docs/wiki/agent-tools-reference.md`, `.smaqit/compendium.md`, `README.md`, `CHANGELOG.md` (3.8.0)
- **Bookkeeping:**
  - `.smaqit/tasks/131_claude_code_agents_cannot_invoke_input_skills_or_elicit.md` and `.smaqit/tasks/PLANNING.md`: lifecycle from create through Completed
  - `.smaqit/references/project-research.md`: task 131 block
  - `smaqit.code-workspace`
- **smaqit-adk:** `.smaqit/tasks/037_bench_harness_selector_claude_code.md` and `.smaqit/tasks/PLANNING.md`, committed locally only
- **Memory:** `task_state_131.md` and `MEMORY.md`

## Next Steps

- **Live Claude Code check on the next real use:** run `smaqit update`, then send a thin `/smaqit.development` request. Expect `NEEDS-INPUT`, then an answer, then a resume. Assisted mode should produce `REVIEW-REQUESTED` after each layer. Confirm `skills:` preload injects the skill content (claude-code#67251).
- **Task 130** (parallel session, uncommitted on the primary checkout) has to reconcile with 131's command bodies and `skills:`/`tools:`.
- **smaqit-adk task 037** (Bench harness selector) needs `task.plan`. Its commit `a1a8fc6` is local and not pushed.
- Task 114 is In Progress in another session.
- **Carried over:** tasks 113 and 126; the unfiled gap where the installer never prunes retired skills.

## Session Metrics

- **Tasks:** 1 created and completed (131); 1 smaqit-adk task created (037)
- **Release:** v3.8.0 (PR #95), with 5 platform binaries published
- **Implementation:** 37 files, +651/−146. `make test` and `make smoke-test` pass. A negative test of the reachability guard confirmed it fails as intended
- **Subagents:** 4 discovery threads (3 Explore + 1 docs guide), 1 cross-session coordination exchange
