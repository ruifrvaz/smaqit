---
status: Not Started
created: "2026-10-11"
---

# Make smaqit Self-Invoking: Directive Instructions, Block Refresh, Trigger-Form Descriptions

## Description

Claude sessions in projects that use smaqit for development, deployment and validation do not invoke smaqit on their own; they edit code, deploy and test directly. Four causes were confirmed in session 089:

1. The managed `AGENTS.md` block (`templates/AGENTS.md.template`) describes smaqit but never tells the coding agent *when* to use it. Its only invocation guidance is "Invoke with `/`…", which reads as user-initiated.
2. Existing projects never receive template changes: `installInstructionsFile` (`installer/main.go`) returns `"up to date"` as soon as the begin marker exists, so `smaqit update` (which re-scaffolds through `cmdInit`) never refreshes the block.
3. Entry-point descriptions describe smaqit's process ("Orchestrate the Development phase — …") rather than the user's request. The Claude phase subagent descriptions end with "Invoked via the /smaqit.X command", which discourages direct delegation.
4. Claude Code puts skill names and descriptions in a listing that is limited to about 1% of the context window. When the listing is too long, it drops descriptions from the least-invoked skills first. Many smaqit skill descriptions are 400–1,000 characters long. Three of them (`smaqit.infrastructure-{deploy-k3s-app,repo-config,request-k3s-onboarding}`) are not valid YAML, because their unquoted descriptions contain `provisioning_mode: existing-k3s`.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Go, Python, PyYAML, Claude Code
**Platforms/Environments:** Linux, WSL2, macOS, Windows
**Features/Integrations:** Claude Code skills, custom commands, subagents, CLAUDE.md @-imports, AGENTS.md project instructions
**Versions/Constraints:** smaqit 3.6.1, Claude Code skill-listing budget (1% of context window, least-invoked descriptions dropped first), 1,536-character per-entry listing cap

## Design Decisions

- **Directive strength — default with exemption:** the block directs the agent to use the matching smaqit workflow for requests that add or change behavior, deploy, or test the deployed system. An explicit exemption list (typos, formatting, comments, dependency bumps with no behavior change) avoids forcing trivial edits through a full phase. Chosen over "suggest and ask" and "mandatory, no exemption".
- **Triggers on both commands and subagents:** the four commands (`development`, `deployment`, `validation`, `qa`) get "Use when…" descriptions plus `when_to_use` trigger phrases. Commands are merged into skills, so Claude can invoke them itself. The four Claude phase/qa subagents get "Use proactively when…" descriptions and lose "Invoked via the /smaqit.X command". Both paths end at the same subagent, because the command only delegates to it, so a double match is harmless. The subagent descriptions must tell Claude to pass the user's full request and relevant session context, because a direct delegation skips the command body's "pass full session context" instruction.
- **Skill descriptions capped at 400 characters, key use case first, enforced at build time:** `scripts/generate-agents.py` fails `make prepare` on invalid frontmatter YAML or on a skill/command description over 400 characters. Any detail that is cut and still needed to run the skill moves into the skill body.
- **No `disable-model-invocation` on `smaqit.input-*`:** it would block Claude from invoking these skills, which task 131's fix depends on. On Claude Code they currently cannot be invoked from inside agents at all: no agent's `claude:` `tools:` list includes `Skill`, and none uses the `skills:` preload. Task 131 addresses this (see Notes); this task does not change agent tool lists.
- **Block refresh replaces only content between the markers:** content outside `<!-- smaqit:instructions:begin -->`/`<!-- smaqit:instructions:end -->` stays byte-identical. This is consistent with the history 059 decision (never overwrite the user's own instructions; "always overwrite" was the rejected alternative). It is also compatible with task 125, whose `.smaqit/slots.yml` pointer lives outside the managed block, and it lets 125's own template change reach existing projects.
- **Malformed block fails safe:** a begin marker without an end marker, reversed markers, or a duplicate begin marker leaves the file untouched and prints a warning. `init` does not fail, and the installer never guesses where a block ends.
- **Out of scope:**
  - Copilot and Codex agent descriptions. Revisit after the live check.
  - smaqit-extensions skills (session/task/release); they live in the sibling repo.
  - `smaqit uninstall` removing the block. It doesn't do that today either; a separate gap.
  - Advising users on Claude Code listing-budget settings. These are user settings that smaqit shouldn't touch.

## Implementation Steps

*Phase A — instruction block*
1. `templates/AGENTS.md.template`: add a "When to use smaqit" section near the top of the block. Keep it neutral across Claude Code, Copilot and Codex. Route requests:
   - new feature on a deployed project → `smaqit.feature-new`
   - new project end to end → `smaqit.new-greenfield-project`
   - build or change the app → development
   - deploy → deployment
   - test the deployed system → validation
   - framework questions → qa

   Add a spec-first rule: run `smaqit status` and read the relevant `specs/` before changing behavior. Add the explicit trivial-change exemption list.
2. `installer/main.go`: add an `instructionsMarkerEnd` constant next to `instructionsMarkerBegin` (line 661). Change `installInstructionsFile` (670-701):
   - Replace everything from the begin marker through the end marker with the current template.
   - Return `"updated"` when the content changed, and `"up to date"` with no write when it is identical.
   - Malformed block → warn and leave the file untouched.
   - Keep the create and append paths.

   Update the comment at 623-625. The status print at 653 needs no logic change; nothing branches on the status.
3. `installer/main_test.go`: add `TestInstallInstructionsFile…` tests, using `TestEnsureManagedToolsGitignorePreservesExistingRules` as the model. Cover created, appended, refreshed (content before and after the block is byte-identical), identical (file not rewritten) and malformed (file untouched).
4. `scripts/smoke-test-installer.sh`: after `init` (around line 68), assert both markers are present in `AGENTS.md` and `CLAUDE.md`. Then plant an outdated block plus a user line outside it, re-run `init`, and assert the block was refreshed and the user line kept.

*Phase B — entry-point triggers (parallel with A)*

5. `commands/smaqit.{development,deployment,validation,qa}.md`: rewrite `description` as "Use when…" with the key use case first. Add `when_to_use` with phrases users actually say ("build/add/change a feature", "deploy", "test the deployed app", "how does smaqit…"). `copy_commands()` copies these into `installer/commands-claude/` unchanged.
6. `.smaqit/definitions/agents/{development,deployment,validation,qa}.frontmatter.yaml`, `claude:` entries only: drop "Invoked via the /smaqit.X command". Rewrite as "Use proactively when…" with the same triggers, plus an instruction to pass the user's full request and relevant session context.

*Phase C — skill descriptions (parallel with A/B)*

7. Rewrite the descriptions of `skills/smaqit.infrastructure-{deploy-k3s-app,repo-config,request-k3s-onboarding}/SKILL.md` as `>-` block scalars so they parse as strict YAML.
8. Rewrite every skill description over 400 characters (about 17, including `infrastructure-image-build-push-static`, `infrastructure-cicd-generate`, `new-greenfield-project`, `feature-new` and `test-e2e-playwright`). Put the key use case first and keep each at 400 characters or fewer. Move any detail that is needed to run the skill into its body, if it isn't there already (for example, deploy-k3s-app's "never touches anything cluster-scoped").
9. `scripts/generate-agents.py`: parse every skill, command and agent frontmatter with strict YAML, and fail `make prepare`, naming the file, on a parse error or on a skill/command description over 400 characters. Depends on 7 and 8.

*Phase D — docs and release*

10. Update the docs:
    - `README.md` "Reinstallation and Updates": the smaqit block is refreshed, and content outside it is preserved.
    - `.smaqit/compendium.md`: the entry "Does `smaqit init` install … `AGENTS.md`?".
    - `docs/wiki/agent-tools-reference.md`: commands are skills Claude can invoke itself; phase subagents are proactive; the description listing budget.
11. Add a MINOR `CHANGELOG.md` entry, and sync the `installer/main.go` `Version` constant with `framework/SMAQIT.md`. The PR title must be `Prepare release vX.Y.Z`.
12. Live check in Claude Code in a smaqit-initialized project (for example mario-hello):
    - Without slash commands, ask "add a farewell page and deploy it" and "what's the status of this project". Record whether smaqit is invoked.
    - Run `/doctor` for the cost of the skill listing.
    - Record the result as behavior evidence, not pass/fail.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `templates/AGENTS.md.template` contains a "When to use smaqit" section that routes feature, build, deploy, test and framework-question requests to the matching smaqit entry point, with a spec-first rule and an explicit trivial-change exemption list
- [ ] `installInstructionsFile` replaces the content between the `smaqit:instructions` begin/end markers with the current template; content outside the markers is byte-identical afterwards; an identical block is not rewritten
- [ ] A malformed block (begin without end, reversed, or duplicate begin) leaves the file untouched and prints a warning without failing `init`
- [ ] Go tests cover the created, appended, refreshed, identical and malformed cases; the smoke test asserts the markers after `init` and a refresh that keeps the user's content on re-`init`
- [ ] The four phase/qa commands have "Use when…" descriptions plus `when_to_use` trigger phrases
- [ ] The four Claude phase/qa subagent descriptions contain no "Invoked via the /smaqit.X command", use "Use proactively when…" wording, and instruct passing the user's full request
- [ ] Every `skills/*/SKILL.md`, `commands/*.md` and generated agent frontmatter parses as strict YAML, including the three k3s-family skills
- [ ] Every skill and command description is 400 characters or fewer with the key use case first; detail that was cut and is needed to run the skill lives in its body
- [ ] `make prepare` fails, naming the file, on invalid frontmatter YAML or a skill/command description over 400 characters
- [ ] README, compendium and the agent-tools-reference wiki describe the block refresh and the trigger model; CHANGELOG and Version sync updated

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
| `templates/AGENTS.md.template` | Modify — "When to use smaqit" section |
| `installer/main.go` | Modify — `instructionsMarkerEnd`, `installInstructionsFile` refresh between markers |
| `installer/main_test.go` | Modify — `installInstructionsFile` test group |
| `scripts/smoke-test-installer.sh` | Modify — marker and refresh assertions |
| `commands/smaqit.{development,deployment,validation,qa}.md` | Modify — "Use when…" + `when_to_use` |
| `.smaqit/definitions/agents/{development,deployment,validation,qa}.frontmatter.yaml` | Modify — Claude descriptions |
| `skills/*/SKILL.md` (3 broken + ~17 over 400 characters) | Modify — descriptions; bodies where detail moves |
| `scripts/generate-agents.py` | Modify — strict YAML + description-length checks |
| `README.md`, `.smaqit/compendium.md`, `docs/wiki/agent-tools-reference.md` | Modify — docs |
| `CHANGELOG.md`, `framework/SMAQIT.md`, `installer/main.go` (`Version`) | Modify — release |

## Notes

- Found in session 089 (2026-10-11) during product analysis: the user noticed that Claude sessions in projects doing development, deployment and validation never invoked smaqit.
- Claude Code mechanics, confirmed in the official docs on 2026-10-11 (code.claude.com/docs/en/skills.md and sub-agents.md):
  - Commands in `.claude/commands/` are merged into skills, so Claude can invoke them itself unless `disable-model-invocation: true` is set.
  - The skill listing keeps every name but drops descriptions from the least-invoked skills first once it exceeds its budget (1% of the context window).
  - The per-entry cap is 1,536 characters across `description` + `when_to_use`.
  - Subagent descriptions with "use proactively" encourage automatic delegation.
- **Overlap with task 125 and its children 127–129 (in progress in a parallel session):** 125's plan also modifies `templates/AGENTS.md.template` (adding a `.smaqit/slots.yml` pointer line), and may touch phase-agent definitions and `feature-new`/`new-greenfield-project`. Rebase onto whichever lands first, and re-read both files before editing.
- **Overlap with task 131** (`.smaqit/tasks/131_claude_code_agents_cannot_invoke_input_skills_or_elicit.md`): Claude Code agents can't invoke their input skills (no `Skill` in `tools:`, no `skills:` preload) or ask the user (no AskUserQuestion in subagents). 131's proposed direction, not yet approved:
  - phase input skills run in the main conversation before delegation
  - spec-layer checks are preloaded with `skills:` and return their questions
  - the command owns the ask-and-resume loop

  Both tasks edit `commands/smaqit.{development,deployment,validation}.md` and the `claude:` frontmatter of the phase agents. Whichever lands second reconciles with the other: 130 owns the `description`/`when_to_use` wording, 131 owns the command bodies and the `tools:`/`skills:` fields. If 131 moves elicitation into the command, check that 130's direct-delegation path (proactive subagents) doesn't bypass it.
- Historical decision: `.smaqit/history/059_release_v1.3.1_and_instructions_files_2026-07-16.md` (create-if-absent, append-if-exists, never overwrite user instructions). This task keeps that guarantee for content outside the markers.
