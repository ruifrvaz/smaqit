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
**Versions/Constraints:** smaqit 3.8.0 (v3.9.0 target), Claude Code skill-listing budget (1% of context window, least-invoked descriptions dropped first), 1,536-character per-entry listing cap

## Design Decisions

- **Base: `origin/main` at v3.8.0, one branch, all phases together.** Tasks 125 (v3.7.0) and 131 (v3.8.0) have shipped. The earlier hold on 131 no longer applies. Re-planned on 2026-10-11 against v3.8.0.
- **Directive strength — default with exemption:** the block directs the agent to use the matching smaqit workflow for requests that add or change behavior, deploy, or test the deployed system. An explicit exemption list (typos, formatting, comments, dependency bumps with no behavior change) avoids forcing trivial edits through a full phase. Chosen over "suggest and ask" and "mandatory, no exemption".
- **Triggers on both commands and subagents:**
  - The four commands (`development`, `deployment`, `validation`, `qa`) get "Use when…" descriptions plus `when_to_use` trigger phrases. Commands are merged into skills, so Claude can invoke them itself.
  - The four Claude phase/qa subagents get "Use proactively when…" descriptions and lose "Invoked via the /smaqit.X command". Their descriptions tell Claude to pass the user's full request plus any `PHASE-PARAMETERS` block it has already resolved.
  - **Frontmatter only:** the command bodies (parameter resolution plus the ask-and-resume loop) and the agents' `tools:`/`skills:` fields belong to task 131 and stay exactly as v3.8.0 shipped them.
- **Return-signal rule in the `AGENTS.md` block (makes direct delegation safe):**
  - A phase agent invoked directly, without a parameter block, applies its preloaded defaults. It can still return `NEEDS-INPUT` or `REVIEW-REQUESTED`.
  - In v3.8.0, the rules for handling those signals live only in the command bodies, so a proactive direct delegation would hand a signal to a main conversation with no instruction for it.
  - The "When to use smaqit" section therefore states: prefer the phase commands, because they resolve parameters and own the ask-and-resume loop. When any smaqit agent returns `NEEDS-INPUT` or `REVIEW-REQUESTED`, ask the user, never answer it yourself, and resume or re-invoke with `answers`.
  - The `AGENTS.md` block is loaded in every session on every platform, so this closes the gap without touching 131's files.
- **Skill descriptions capped at 400 characters, key use case first, enforced at build time:**
  - `scripts/generate-agents.py` fails `make prepare` on skill or command frontmatter that is not valid YAML, or on a description over 400 characters. Skills and commands are copied without being parsed today, which is why the check is needed. Agent metadata is already loaded as YAML.
  - Detail that is cut and still needed to run the skill moves into the skill body.
  - Each edited skill gets a patch bump of `metadata.version`, following v3.7.0's convention. Its `smaqit-slots` metadata stays untouched.
- **Task 114 ordering:** task 114 (in progress) edits 5 of the 17 long-description skills: `provider-cyso`, `vault-loader`, `provision-cyso`, `repo-config` and `cicd-generate`. Trim those last. If 114 merges first, rebase onto it before editing them. They stay in this task's scope, because the 400-character check would fail without them.
- **No `disable-model-invocation` on `smaqit.input-*`:** since v3.8.0, the phase commands apply the phase input skills in the main conversation, and agents preload them through `skills:`. Disabling model invocation would block the command path.
- **Block refresh replaces only content between the markers:**
  - Content outside `<!-- smaqit:instructions:begin -->`/`<!-- smaqit:instructions:end -->` stays byte-identical. This is consistent with the history 059 decision (never overwrite the user's own instructions; "always overwrite" was the rejected alternative).
  - v3.7.0 added its deployment-slots paragraph and its `.smaqit/slots.yml` exception *inside* the managed block, so existing projects receive them only through this refresh.
- **Malformed block fails safe:** a begin marker without an end marker, reversed markers, or a duplicate begin marker leaves the file untouched and prints a warning. `init` does not fail, and the installer never guesses where a block ends.
- **Out of scope:**
  - Copilot and Codex agent descriptions. Revisit after the live check.
  - smaqit-extensions skills (session/task/release); they live in the sibling repo.
  - `smaqit uninstall` removing the block. It doesn't do that today either; a separate gap.
  - Advising users on Claude Code listing-budget settings. These are user settings that smaqit shouldn't touch.
  - Command bodies, and agents' `tools:`/`skills:` (task 131).

## Implementation Steps

*Phase A — instruction block*
1. `templates/AGENTS.md.template`: add a "When to use smaqit" section near the top of the block. Keep it neutral across Claude Code, Copilot and Codex, and leave v3.7.0's slots paragraph in place. Route requests:
   - new feature on a deployed project → `smaqit.feature-new`
   - new project end to end → `smaqit.new-greenfield-project`
   - build or change the app → development
   - deploy → deployment
   - test the deployed system → validation
   - framework questions → qa

   The section also contains:
   - **Prefer the phase commands**, because they resolve parameters and own the ask-and-resume loop.
   - **The return-signal rule:** for `NEEDS-INPUT`/`REVIEW-REQUESTED`, ask the user, never answer yourself, and resume or re-invoke with `answers`.
   - **The spec-first rule:** run `smaqit status` and read the relevant `specs/` before changing behavior.
   - **The explicit trivial-change exemption list.**
2. `installer/main.go`: add an `instructionsMarkerEnd` constant next to `instructionsMarkerBegin`. Change `installInstructionsFile`:
   - Replace the span from the begin marker through the end marker with the current template.
   - Return `"updated"` when the content changed, and `"up to date"` with no write when it is identical.
   - Malformed block → warn and leave the file untouched.
   - Keep the create and append paths.

   Update the comment on the two `cmdInit` callers, which sit just before `installSlotsSkeleton`. Find the code by name; line numbers moved in v3.7.0.
3. `installer/main_test.go`: add `TestInstallInstructionsFile…` tests, using `TestEnsureManagedToolsGitignorePreservesExistingRules` as the model. Cover created, appended, refreshed (content before and after the block is byte-identical), identical (file not rewritten) and malformed (file untouched).
4. `scripts/smoke-test-installer.sh`: after the first `init`, assert both markers are present in `AGENTS.md` and `CLAUDE.md`. Then plant an outdated block plus a user line outside it, re-run `init`, and assert the block was refreshed and the user line kept. Place these next to v3.7.0's `slots.yml` assertions.

*Phase B — entry-point triggers*

5. `commands/smaqit.{development,deployment,validation,qa}.md`, **frontmatter only**: rewrite `description` as "Use when…" with the key use case first. Add `when_to_use` with phrases users actually say ("build/add/change a feature", "deploy", "test the deployed app", "how does smaqit…"). Bodies stay as v3.8.0 shipped them.
6. `.smaqit/definitions/agents/{development,deployment,validation,qa}.frontmatter.yaml`, **`claude:` `description` only**: drop "Invoked via the /smaqit.X command". Write "Use proactively when…" with the same triggers, plus "pass the user's full request and any resolved `PHASE-PARAMETERS` block". Also drop "Invoked via" from any spec-agent description that has it, keeping their "not a direct user entry point" wording. `tools:`/`skills:` stay unchanged.

*Phase C — skill descriptions*

7. Rewrite the 17 skill descriptions over 400 characters (re-measured on v3.8.0):
   - feature-new
   - new-greenfield-project
   - requirements-extract
   - spec-status-update
   - test-e2e-playwright
   - infrastructure-:
     - deploy-k3s-app
     - deploy-rsync-python-tornado
     - deploy-verify
     - domain-tls
     - hook-post-deploy-stamp
     - image-build-push-static
     - request-k3s-onboarding
     - provider-cyso, provision-cyso, repo-config, vault-loader, cicd-generate. These five overlap task 114: do them last.

   Put the key use case first and keep each at 400 characters or fewer. Move needed detail into the body, and patch-bump `metadata.version`.
8. `scripts/generate-agents.py`: next to `check_input_skill_reachability`, add a check that parses every `skills/*/SKILL.md` and `commands/*.md` frontmatter as strict YAML. Fail `make prepare`, naming the file, on a parse error or on a description over 400 characters. Depends on 7.

*Phase D — docs and release*

9. Update the docs:
   - `README.md` "Reinstallation and Updates": the smaqit block is refreshed, and content outside it is preserved.
   - `.smaqit/compendium.md`: the entry "Does `smaqit init` install … `AGENTS.md`?", plus how a directly delegated agent's return signals are handled.
   - `docs/wiki/agent-tools-reference.md`: the trigger model (commands Claude can invoke itself, proactive phase subagents, the return-signal rule) and the description listing budget, next to v3.8.0's "Asking the User" section.
10. Add a `CHANGELOG.md` entry (v3.9.0 unless something else lands first), and sync the `installer/main.go` `Version` constant with `framework/SMAQIT.md`. The PR title must be `Prepare release vX.Y.Z`.
11. Live check in Claude Code in a smaqit-initialized project (for example mario-hello):
    - Without slash commands, ask "add a farewell page and deploy it" and "what's the status of this project". Record whether smaqit is invoked.
    - If a direct delegation returns `NEEDS-INPUT`, confirm the main conversation asks the user rather than answering itself.
    - Run `/doctor` for the cost of the skill listing.
    - Record the result as behavior evidence, not pass/fail.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `templates/AGENTS.md.template` contains a "When to use smaqit" section that routes feature, build, deploy, test and framework-question requests to the matching smaqit entry point, prefers the phase commands, states the return-signal rule (`NEEDS-INPUT`/`REVIEW-REQUESTED`: ask the user, never answer yourself, resume with `answers`), and includes a spec-first rule and an explicit trivial-change exemption list
- [ ] `installInstructionsFile` replaces the content between the `smaqit:instructions` begin/end markers with the current template; content outside the markers is byte-identical afterwards; an identical block is not rewritten
- [ ] A malformed block (begin without end, reversed, or duplicate begin) leaves the file untouched and prints a warning without failing `init`
- [ ] Go tests cover the created, appended, refreshed, identical and malformed cases; the smoke test asserts the markers after `init` and a refresh that keeps the user's content on re-`init`
- [ ] The four phase/qa commands have "Use when…" descriptions plus `when_to_use` trigger phrases; only their frontmatter changes, and their bodies are identical to v3.8.0
- [ ] The four Claude phase/qa subagent descriptions contain no "Invoked via the /smaqit.X command", use "Use proactively when…" wording, and instruct passing the user's full request and any resolved `PHASE-PARAMETERS` block; their `tools:`/`skills:` fields are unchanged from v3.8.0
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
| `skills/*/SKILL.md` (17 over 400 characters) | Modify — descriptions; bodies where detail moves |
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
- **Overlap with task 125 — resolved:** shipped in v3.7.0 (PR #94). It added a deployment-slots paragraph and a `.smaqit/slots.yml` exception *inside* the managed `AGENTS.md` block, added `installSlotsSkeleton` right after the `installInstructionsFile` calls, added `slots.yml` smoke assertions, and quoted the three invalid-YAML skill descriptions.
- **Overlap with task 131 — resolved:** shipped in v3.8.0 (PR #95).
  - What 131 rewrote: the command bodies (parameter resolution and the ask-and-resume loop), the agents' Claude `skills:` preloads and Copilot `vscode/askQuestions`, and the input skills. It also added `check_input_skill_reachability` to `scripts/generate-agents.py`.
  - It left all `description`/`when_to_use` wording to this task, per the agreed split. Its follow-up list says this task must reconcile with its command bodies and `skills:` fields, which this task does by editing frontmatter descriptions only.
  - 131's return-signal handling exists only in the command bodies. The return-signal rule in the `AGENTS.md` block (Design Decisions) covers direct delegation.
- **Overlap with task 114 (in progress, parallel session):** it edits `provider-cyso`, `vault-loader`, `provision-cyso`, `repo-config` and `cicd-generate` (`SKILL.md`), plus `agents/deployment.md`, the compendium and CHANGELOG. Trim those five descriptions last; rebase onto 114 first if it merges before this task.
- Historical decision: `.smaqit/history/059_release_v1.3.1_and_instructions_files_2026-07-16.md` (create-if-absent, append-if-exists, never overwrite user instructions). This task keeps that guarantee for content outside the markers.
