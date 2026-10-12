# Self-Invoking Entry Point Triggers

## Metadata

- **Date:** 2026-10-11 → 2026-10-12 (one session, across two days)
- **Session focus:** Find out why Claude sessions in smaqit projects don't invoke smaqit on their own, then fix it across project instructions, the installer, and agent, command, and skill descriptions
- **Tasks created:** 130, 133
- **Tasks completed:** 130 (released v3.10.0, PR #97), 133 (released v3.10.1, PR #98)
- **Tasks referenced:** 113, 126 (assessed at session start); 125/127–129 and 131 (parallel sessions, overlapping files); 114 (parallel, claimed v3.9.0); 132 (planned headless e2e, owned by another session)

## Actions Taken

- **Session start:** loaded project context. Assessed task 113 (still no warning-only output path in `design.go`; the `Warnings` field holds PlantUML's own syntax warnings) and task 126 (recommended design: lock only the rename, then a short stale threshold; drop the duplicate validation).
- **Product analysis:** the user reported that Claude sessions never invoke smaqit. Read the installed templates, command, agent, and skill frontmatter, and `installInstructionsFile`. Found four causes:
  - The managed `AGENTS.md` section described smaqit but never said when to use it.
  - Existing projects never received template changes.
  - Entry-point descriptions described the process ("Orchestrate the Development phase") rather than the user's request, and the Claude phase agents said "Invoked via the /smaqit.X command".
  - Claude Code's skill listing drops descriptions under its budget; this session's own listing showed about 15 smaqit skills without descriptions.
- **Planning:** three discovery agents covered the installer write path, where descriptions come from, and Claude Code invocation mechanics (confirmed in the official docs). Created task 130. The user chose: a default-with-exemption directive, triggers on both commands and proactive subagents, and a 400-character cap on skill descriptions.
- **Cross-session coordination:** a parallel session filed task 131 (agents can't reach their input skills or ask the user). Corrected task 130's rationale and agreed an ownership split: 130 owns `description`/`when_to_use`, 131 owns command bodies and `tools:`/`skills:`. Re-planned task 130 twice: after PR #94 (v3.7.0, which had already quoted the three invalid-YAML descriptions) and after v3.8.0 (task 131), which exposed that its ask-and-resume loop lives only in the command bodies.
- **Task 130 implementation:**
  - A "When to use smaqit" section in `templates/AGENTS.md.template`, including a return-signal rule.
  - `installInstructionsFile` now refreshes the section between its markers, with an end-marker constant, CRLF preservation, and a malformed-markers warning.
  - Six Go tests and new smoke-test assertions.
  - "Use when…" descriptions plus `when_to_use` on the four Claude commands, and "Use proactively when…" on the four Claude entry agents.
  - 17 skill descriptions trimmed to 400 characters or fewer.
  - A strict-YAML and 400-character frontmatter check in `scripts/generate-agents.py`.
  - README, compendium, and wiki updates.
- **Task 130 release:** opened PR #97 as v3.10.0, since task 114's PR #96 had claimed v3.9.0. Resolved the CHANGELOG-only conflict after #96 merged by merging `origin/main` into the branch, with no rebase. Phase 2 confirmed the merge and the release (5 binaries), then cleaned up the worktree and local branch.
- **Live check:** `smaqit update` to v3.10.0 from a neutral directory, then validated the installed global payload. In a scratch project, `init` wrote the new section, and `validate`, `status`, `mcp verify`, and `slots` passed. Re-init reported `up to date`, an outdated section was refreshed with user content kept, and a malformed section was left untouched with a warning and exit 0.
- **Task 133:** Copilot and Codex descriptions for the four entry agents in trigger form; Codex phase agents also relay return signals. Released as v3.10.1 via PR #98, then wrapped up after the merge.
- **Q&A:** whether "use proactively" wording or `CLAUDE.md` instructions are enough. Both are needed, and the specification agents stay non-proactive.
- **Final housekeeping:**
  - Updated the global install to v3.10.1.
  - Re-checked claude-code #85027: still open, but it affects the VS Code agent host, not the Claude Code extension used here.
  - Pushed 24 local `main` bookkeeping commits (fast-forward).

## Problems Solved

- **Existing projects never received template changes** (including v3.7.0's slots guidance): fixed by refreshing between the markers, with content outside them preserved byte for byte.
- **Proactive delegation skipped the command's question relay:** fixed by putting the return-signal rule in the `AGENTS.md` section, which loads in every session.
- **Over-budget, process-worded descriptions:** fixed by trigger-form rewrites plus a build-time 400-character check.
- **Version and file overlap with parallel tasks:**
  - Task 114's five overlapping skills changed only on the description line, with no version bump, so the branches merged cleanly.
  - Task 114 had claimed v3.9.0, so this release moved up to v3.10.0.
  - The PR #97 CHANGELOG conflict was resolved by keeping both sections, newest first.
- **A first script pass failed safely** for task 133 (it assumed a newline before the first YAML key); it was fixed before anything was written.

## Decisions Made

- **The directive defaults to using smaqit,** with an explicit list of trivial changes that may skip it; preferred over "suggest and ask" or "mandatory".
- **Commands and Claude entry agents both carry triggers.** The commands stay the preferred path; the proactive agents are the fallback, made safe by the `AGENTS.md` return-signal rule. The specification agents are never proactive.
- **Skill descriptions are capped at 400 characters** and enforced at build time. No `disable-model-invocation` on input skills, since the commands invoke them.
- **A malformed marker section warns rather than fails,** so one damaged file never blocks `smaqit update`.
- **Copilot descriptions carry only the trigger,** because Copilot agents ask the user directly. Codex phase descriptions add the relay instruction, because Codex returns signals to the spawning session.
- **User decision:** `smaqit uninstall` intentionally keeps the managed `AGENTS.md`/`CLAUDE.md` section.
- **Version sync is a no-op:** `framework/SMAQIT.md` carries no version, and `Version` in `main.go` is an ldflags fallback.

## Files Modified

- `templates/AGENTS.md.template`: "When to use smaqit" section (v3.10.0)
- `installer/main.go`, `installer/main_test.go`: marker-section refresh and its tests (v3.10.0)
- `scripts/smoke-test-installer.sh`: marker and refresh assertions (v3.10.0)
- `scripts/generate-agents.py`: `check_listing_frontmatter` / `check_all_listing_frontmatter` (v3.10.0)
- `commands/smaqit.{development,deployment,validation,qa}.md`: frontmatter only (v3.10.0)
- `.smaqit/definitions/agents/{development,deployment,validation,qa}.frontmatter.yaml`: Claude descriptions (v3.10.0); Copilot and Codex descriptions (v3.10.1)
- `skills/*/SKILL.md` (17): descriptions, plus patch version bumps on 12 of them; one body sentence in `smaqit.test-e2e-playwright` (v3.10.0)
- `README.md`, `.smaqit/compendium.md`, `docs/wiki/agent-tools-reference.md`, `CHANGELOG.md`: documentation for both releases
- `.smaqit/tasks/130_*.md`, `.smaqit/tasks/133_*.md`, `.smaqit/tasks/PLANNING.md`: task lifecycle
- `.smaqit/references/project-research.md`: task 130 and 133 research blocks
- Memory: `task_state_130`, `task_state_133`, `MEMORY.md`

## Next Steps

- **Headless Claude Code check (task 132, owned by another session, In Progress):** confirm that, without slash commands, a session reaches for smaqit and a directly delegated `NEEDS-INPUT` reaches the user. If Claude routinely bypasses the commands in favor of the proactive agents, tone down the agent wording.
- **Codex spawning from descriptions** has no behavioral check yet; task 132 is Claude-only.
- **Carried over:** task 126 (install-path lock; recommended design noted above), task 113 (decide where the advisory output goes), the unfiled installer skill-retirement pruning gap.
- **Downstream projects:** run `smaqit update` to receive the refreshed `AGENTS.md` section.

## Session Metrics

- Releases shipped: 2 (v3.10.0, PR #97; v3.10.1, PR #98), each with 5 platform binaries
- Tasks created: 2 (130, 133); tasks completed: 2
- Implementation diff: task 130 had 34 files (+344/−84); task 133 had 6 files (+11/−10)
- New tests: 6 Go test functions (9 cases including the malformed subtests) and 7 smoke-test assertions
- Skill descriptions over 400 characters: 17 before, 0 after
- Merge conflicts resolved: 1 (CHANGELOG only)
- Bookkeeping commits pushed at session end: 24
