# Trigger-Based Phase Mode Defaults

## Metadata

- **Date:** 2026-10-18
- **Session focus:** Assess why a bare or title-only smaqit project could start a highly opinionated, autonomous development cycle, and ship a fix: phase mode defaults by what started the phase, and layer gates that count only user-stated requirements
- **Tasks completed:** 138 (v3.17.0, PR #106)
- **Tasks referenced:** 135 and 136 (completed by parallel sessions as v3.15.0 and v3.16.0), 137, 139 and 140 (created or completed by parallel sessions), 126, 113, 094, 077, 074, 071

## Actions Taken

- **`session.start`:** loaded README, CONTRIBUTING, history 094, PLANNING, and the compendium. Found task 135 already implemented in its worktree, and confirmed every line task 136 cites against the code.
- **`session.assess` on the phase triggers:** traced `/smaqit.development` → `smaqit.input-development` → `smaqit-development` → the layer gates, and found four holes:
  - phase input skills defaulted `mode` to `autonomous` and asked only when the request already mentioned review;
  - layer gates' "Extract from session context" let a project name, the model's world knowledge, or the orchestrator's paraphrase satisfy a required section;
  - the main conversation retold the request to the subagent in its own words;
  - pre-task-131 "flag/document assumptions" directives licensed filling gaps.
  Also found the orchestrators' autonomous mode stopping at unconditional review gates.
- **`task.plan` (Mode A):** three Explore agents mapped per-platform phase entry paths, the e2e harness and text tests, and every mode/assumption directive. Two Q&A rounds settled the design. A start-confirmation signal (`CONFIRM-START`) was drafted, then dropped when the user restated the model: commands are assisted, skills (and model-started phases) are autonomous.
- **`task.create 138`** (137 had been taken by a parallel session), then synced local `main` with `origin/main` (merging task 136's PR) and pushed.
- **`task.start 138`:** branch and worktree from `origin/main`, a task-only research-map block (10 verified URLs), and issue triage (Advisory: claude-code #82456 on unrequested Agent-tool use, codex #46048 on subagent completion relay).
- **Implementation (assisted), seven groups:**
  1. Phase input skills and commands: mode by trigger, the mode question dropped, a verbatim `user_request` field, the resolved mode echoed before delegating.
  2. Phase agents: assisted as the no-block default, `user_request`/`answers`/`mode` passed to spec agents, the assumption SHOULD replaced, an assisted phase-end review prompt, corrected handovers.
  3. `AGENTS.md` template mode rule and the Copilot spec-agent invocability fix; Claude/Codex descriptions.
  4. Orchestrators: Phase 0 defaults to Autonomous; four review/sign-off gates marked assisted-only.
  5. Requirement Sources in all five layer gates, matching spec-agent Session Context, `framework/AGENTS.md` and `framework/PHASES.md`.
  6. e2e-claude `empty-request` and `title-only` scenarios.
  7. Docs, compendium and CHANGELOG.
- Verified with `make prepare`, `make test`, `make smoke-test`, and greps of the generated Copilot, Claude, and Codex artifacts.
- **`task.complete 138` Phase 1:** Findings, release analysis (MINOR, v3.17.0, no claimed collision), PR #106, and the changelog section on the branch. The live-e2e criterion was marked `[!]` because the user asked to open the PR first and plan the run after.
- **Live e2e-claude**, staged at the user's direction:
  - Stage 1: `empty-request` and `title-only` passed (about $0.48).
  - Stage 2: `assisted`, `self-invocation`, `happy-path`, and `needs-input` passed (about $4.71).
  The criterion was ticked and the PR description updated with a results table.
- **`task.complete 138` Phase 2** after the user merged: verified `MERGED`, merged `origin/main`, marked Completed, removed the worktree and local branch, and rebuilt the workspace. Confirmed the v3.17.0 GitHub Release (latest, five binaries, notes from the changelog section).

## Problems Solved

- **A bare or title-only request produced an opinionated cycle.** Layer gates now accept only the user's request, `answers`, the user's own messages, and files the user pointed at. A name, the model's own knowledge, or a paraphrase never counts. Live, "Let's build mario-app." in a `mario-app` project stopped at the Business gate: "names a project but doesn't say what it should do".
- **Mode was silently autonomous.** It now follows the trigger, without a question. The live `assisted` scenario paused per layer with no mode stated.
- **Orchestrators' autonomous mode still stopped at human review gates.** Spec-set approval, validation sign-off, the Phase 0 confirmation, and the inventory check are now assisted-only.
- **The orchestrator retold the request.** `user_request` now carries the user's words unchanged to every spec agent.
- **Stale facts:** the development and deployment handovers pointed at spec agents, and the `AGENTS.md` template claimed Copilot spec agents are directly invocable. Both fixed.
- **Caught during the work:**
  - Edit-tool edits failed until each worktree file was opened with Read.
  - A plain YAML description can't contain `mode: autonomous`; reworded to "set its mode to autonomous".
  - The new section in `agent-tools-reference.md` first absorbed two unrelated paragraphs; moved it to its own section.
  - `gh pr edit` failed on the classic-Projects GraphQL deprecation; the PR body was updated through the REST API.

## Decisions Made

- **Mode default by trigger:**
  - a typed command, a picked agent, or a no-block phase agent → `assisted`;
  - a model-started phase, or greenfield/feature-new → `autonomous`;
  - a stated mode always wins.
- **Assisted:** user-stated input only, `REVIEW-REQUESTED` per layer, and a review prompt at phase end.
- **Autonomous:** stops only for `NEEDS-INPUT` and human operations (PR merges, credential prompts, the synthesized-skill checkpoint, the release).
- **No start confirmation:** `CONFIRM-START` was considered and dropped. For self-invoked runs, the requirement-source rules are the guard, and the live `title-only` run proved them.
- **Provenance depends on mode:** autonomous also counts upstream specs; assisted traces every layer's required sections to the user. Business traces to the user in both modes.
- **An empty request blocks only when a layer must be generated.**
- **Only required sections are protected from assumptions.** `qa.md` is out of scope.
- **Deviation from the approved plan:** feature-new's parent task stays in Assisted task mode, because the task lifecycle's Autonomous mode self-merges the release PR, and a PR merge is a human operation.
- **Kept in autonomous runs:** greenfield's "structural divergence needs operator approval" stop. It fires when code can't follow an approved spec, so it is closer to `NEEDS-INPUT` than to a review gate.
- **Live e2e policy:** keep the strict `NEEDS-INPUT` check (a conversation that only asks questions itself would fail), and run the cheap guard scenarios before the expensive ones.
- **Released as MINOR with an action-required note,** following the v3.13.0/v3.14.0 precedent for behavior-changing defaults.

## Files Modified

- `skills/smaqit.input-{development,deployment,validation}/SKILL.md`: Mode by trigger, the User Request parameter, output blocks
- `skills/smaqit.input-{business,functional,stack,infrastructure,coverage}/SKILL.md`: the Requirement Sources section, tightened Stack and Coverage sufficiency, version bumps
- `commands/smaqit.{development,deployment,validation}.md`: trigger reporting, verbatim request, mode echo
- `agents/{development,deployment,validation}.md`: Execution Parameters, Execution Mode, scoped context, assumption SHOULD, handovers
- `agents/{business,functional,stack,infrastructure,coverage}.md`: Role context, Session Context, the assumption MUST NOT
- `.smaqit/definitions/agents/{development,deployment,validation}.frontmatter.yaml`: Claude/Codex descriptions
- `.smaqit/definitions/skills/smaqit.new-greenfield-project.md`: spec-set gate assisted-only
- `skills/smaqit.new-greenfield-project/SKILL.md`, `skills/smaqit.feature-new/SKILL.md`: Phase 0 default, assisted-only gates, `user_request`
- `templates/AGENTS.md.template`: the mode rule and the spec-agent invocability fix
- `framework/AGENTS.md`, `framework/PHASES.md`: provenance, the Phase Parameters table, Execution Modes, fail-fast
- `scripts/e2e-claude.sh`, new `scripts/e2e/hello-site/title-only-request.md`: the two new scenarios, `new_project` seeding, updated prompts
- `README.md`, `docs/wiki/workflows/{quickstart,testing-smaqit}.md`, `docs/wiki/agent-tools-reference.md`, `docs/wiki/designs/fail-fast-on-ambiguity.md`
- `.smaqit/compendium.md` (a new mode/provenance entry and three updated entries), `CHANGELOG.md` (v3.17.0)
- `.smaqit/tasks/138_phase_mode_defaults_by_trigger_and_requirements_provenance.md`, `.smaqit/tasks/PLANNING.md`, `.smaqit/references/project-research.md`, `smaqit.code-workspace`

## Next Steps

- Task 137 (existing-k3s credential slots under a scripts-only Vault policy) is the next High-priority task.
- Follow-up from task 138: a `scripts/generate-agents.py` check that every phase command and agent states the trigger-based default.
- Watch claude-code #82456: model-started phases now rely on the command-as-skill path, which the live `title-only` run used.
- Existing smaqit projects must run `smaqit update` to refresh their `AGENTS.md` section (v3.17.0 action-required note).
- Carried over: the unseal key and legacy writes still go via argv; the installer skill-retirement pruning gap; tasks 126 and 113; smaqit-extensions task 042 is uncommitted.

## Session Metrics

- Duration: one day (2026-10-18)
- Tasks completed: 1 (138); release v3.17.0 (PR #106)
- Discovery: 3 Explore agents; 4 question rounds with the user (assessment, two plan rounds, model confirmation), plus a live-e2e policy question
- Files: 37 changed in the implementation commit (1 new fixture); 412 insertions, 182 deletions
- Verification: `make prepare`, `make test`, `make smoke-test`, and generated-artifact greps all pass; live e2e-claude 6/6 scenarios on Claude Code 2.1.296, about $5.19 total, 0 permission denials
