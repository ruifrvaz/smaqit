---
status: PR Open
pr: 97
mode: Assisted
started: "2026-10-11"
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
**Triaged:** 2026-10-11
**Tools searched:** Claude Code, PyYAML, Go
**Result:** Advisory

### Advisory Issues
- [#85027 VS Code agent host drops user skill descriptions, so skills never auto-activate](https://github.com/anthropics/claude-code/issues/85027) — `anthropics/claude-code` — opened 2026-08-08 — area:ide, area:skills. Directly relevant: the user runs the VS Code extension. If this still reproduces, trimming descriptions won't make skills auto-activate in VS Code. The AGENTS.md directive (loaded through CLAUDE.md, not the skill listing) is the lever that doesn't depend on it. Re-check during the live check.
- [#81081 [BUG] Session-start skill listing silently truncates most skill descriptions under a size budget](https://github.com/anthropics/claude-code/issues/81081) — `anthropics/claude-code` — opened 2026-07-25 — bug, documentation, reproduced. Downgraded from Blocking: the skills docs (research map) describe the listing budget and dropped descriptions as intended behavior. This is the mechanism the 400-character cap works around.
- [#95582 Skill catalog descriptions intermittently missing from system prompt (frontmatter on disk is correct)](https://github.com/anthropics/claude-code/issues/95582) — `anthropics/claude-code` — opened 2026-09-19 — bug, platform:windows, area:skills
- [#93044 Project skills beyond a session-level cap return "Unknown skill" from the Skill tool despite valid SKILL.md files](https://github.com/anthropics/claude-code/issues/93044) — `anthropics/claude-code` — opened 2026-09-09 — bug, has repro, platform:web, area:skills
- [#88058 Feature request: per-skill display description separate from the routing description (SKILL.md)](https://github.com/anthropics/claude-code/issues/88058) — `anthropics/claude-code` — opened 2026-08-19 — enhancement, area:skills

### Historical (Closed)
- [#68135 [Bug] Skill listing truncated with descriptions dropped due to context budget constraints](https://github.com/anthropics/claude-code/issues/68135) — `anthropics/claude-code` — closed 2026-08-19
- [#86703 [Bug] Skill descriptions not loaded into context window](https://github.com/anthropics/claude-code/issues/86703) — `anthropics/claude-code` — closed 2026-08-14
- [#98674 [Bug] Agent ignores custom skills and performs unrelated tasks instead](https://github.com/anthropics/claude-code/issues/98674) — `anthropics/claude-code` — closed 2026-10-08
- [#95367 [BUG] No disk-sourced skills load in 2.1.271 — user skills and enabled-plugin skills both absent](https://github.com/anthropics/claude-code/issues/95367) — `anthropics/claude-code` — closed 2026-10-04
- [#318 Valid .yaml generates error "yaml.scanner.ScannerError: mapping values are not allowed here"](https://github.com/yaml/pyyaml/issues/318) — `yaml/pyyaml` — closed 2019-07-03. The same error class the three k3s-family descriptions hit before v3.7.0; it's expected YAML behavior for an unquoted `key: value` in a plain scalar.

### Unresolvable Tools
- Python — the helper resolved an unrelated repository (`TheAlgorithms/Python`); not searched

### Search Warnings
- `golang/go` open/closed searches — no results matched the task's feature dimension (the search terms were too generic for a language repository); not evidence of absence

## Acceptance Criteria

- [x] `templates/AGENTS.md.template` contains a "When to use smaqit" section that routes feature, build, deploy, test and framework-question requests to the matching smaqit entry point, prefers the phase commands, states the return-signal rule (`NEEDS-INPUT`/`REVIEW-REQUESTED`: ask the user, never answer yourself, resume with `answers`), and includes a spec-first rule and an explicit trivial-change exemption list
- [x] `installInstructionsFile` replaces the content between the `smaqit:instructions` begin/end markers with the current template; content outside the markers is byte-identical afterwards; an identical block is not rewritten
- [x] A malformed block (begin without end, reversed, or duplicate begin) leaves the file untouched and prints a warning without failing `init`
- [x] Go tests cover the created, appended, refreshed, identical and malformed cases; the smoke test asserts the markers after `init` and a refresh that keeps the user's content on re-`init`
- [x] The four phase/qa commands have "Use when…" descriptions plus `when_to_use` trigger phrases; only their frontmatter changes, and their bodies are identical to v3.8.0
- [x] The four Claude phase/qa subagent descriptions contain no "Invoked via the /smaqit.X command", use "Use proactively when…" wording, and instruct passing the user's full request and any resolved `PHASE-PARAMETERS` block; their `tools:`/`skills:` fields are unchanged from v3.8.0
- [x] Every `skills/*/SKILL.md`, `commands/*.md` and generated agent frontmatter parses as strict YAML, including the three k3s-family skills
- [x] Every skill and command description is 400 characters or fewer with the key use case first; detail that was cut and is needed to run the skill lives in its body
- [x] `make prepare` fails, naming the file, on invalid frontmatter YAML or a skill/command description over 400 characters
- [x] README, compendium and the agent-tools-reference wiki describe the block refresh and the trigger model; CHANGELOG updated (Version sync: nothing to sync — `framework/SMAQIT.md` carries no version and `installer/main.go`'s `Version` is an ldflags-overridden fallback that releases don't touch)

## Findings

**Implementation approach:**
- `templates/AGENTS.md.template` gains "When to use smaqit": a request-to-entry-point table, specifications first, prefer the phase commands, a return-signal rule (`NEEDS-INPUT`/`REVIEW-REQUESTED` go to the user, never answered by the agent), and an explicit trivial-change exemption list
- `installInstructionsFile` replaces the begin-through-end marker span with the current template: `updated` / `up to date` (no write) / malformed → `errMalformedInstructionsBlock`, which `installInstructionsFileOrExit` turns into a warning and `left unchanged` without failing init. CRLF files keep CRLF
- Six `TestInstallInstructionsFile…` tests (created, appended, refreshed, current-not-rewritten via mtime, CRLF, four malformed layouts); the smoke test asserts markers after init and a refresh that keeps user content, reusing the existing slots re-init
- Commands: frontmatter-only rewrite to "Use when…" `description` + `when_to_use`; bodies byte-identical to v3.8.0. Claude phase/qa agents: "Use proactively when…" descriptions carrying the parameter-block and return-signal instructions; `tools:`/`skills:`, Copilot and Codex entries verified unchanged
- 17 skill descriptions rewritten by a script that touched only `description` (and `metadata.version`), refused anything over 400 characters, and re-parsed each result; `check_all_listing_frontmatter()` in `scripts/generate-agents.py` runs first in `main()` and fails `make prepare` naming the file
- Verified: `make test`, `make smoke-test`, planted-defect negative checks, a 34-file source scan and 91-file generated-output scan, and a real v3.8.0-initialized project refreshed by the dev build (content outside the markers byte-identical, `slots.yml` untouched, second run `up to date`)

**Decisions made:**
- Re-planned against v3.8.0 after tasks 125 and 131 shipped: the YAML-quoting step was already done in v3.7.0, and the return-signal rule was added because 131's ask-and-resume loop lives only in command bodies, which a proactive direct delegation skips
- Task 114's five overlapping skills (`provider-cyso`, `vault-loader`, `provision-cyso`, `repo-config`, `cicd-generate`) changed only their description line, with no version bump: 114 already bumps their versions on a separate line, so the branches merge without conflict
- `cicd-generate` kept its "greenfield generates workflows itself" and "never generates copilot-setup-steps.yml" facts in the description rather than moving them into a body 114 is editing; `test-e2e-playwright` gained one body sentence for the only other description-only detail
- The five specification agents' "Invoked via Task by … not a direct user entry point" wording stays: it is accurate and keeps them delegation-only
- Phase agent descriptions (~450 characters) are outside the 400-character limit, which applies only to skills and commands, the entries that share Claude Code's listing budget
- Malformed markers warn rather than fail, so one hand-damaged file never blocks `smaqit update`

**Blockers encountered:**
- None

**Follow-up identified:**
- Live Claude Code check not run (installing the dev build globally would overwrite `~/.claude`): after release, in a smaqit-initialized project, ask "add a farewell page and deploy it" and "what's the status of this project" without slash commands, confirm smaqit is used, and confirm a directly delegated `NEEDS-INPUT` reaches the user
- Open claude-code issue #85027 (VS Code agent host drops user skill descriptions, so skills never auto-activate) may still limit description-based triggering in the VS Code extension; the `AGENTS.md` directive doesn't depend on it. Re-check during the live check
- Copilot and Codex agent descriptions remain in the older form; revisit after the live check
- `smaqit uninstall` still leaves the managed `AGENTS.md`/`CLAUDE.md` section in place

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
