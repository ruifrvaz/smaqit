---
status: Completed
mode: Assisted
started: "2026-10-12"
created: "2026-10-12"
completed: "2026-10-12"
---

# Copilot and Codex Entry-Agent Descriptions in Trigger Form

## Description

Task 130 (v3.10.0) rewrote the Claude Code descriptions of the four entry agents (`development`, `deployment`, `validation`, `qa`) in "Use when…"/"Use proactively when…" form. The Copilot and Codex descriptions of the same agents still use the older wording:
- Copilot: "Implementation agent for the Development phase.", "Fetch and answer questions about smaqit framework documentation"
- Codex: "Implementation agent for the Development phase that spawns Business, Functional, and Stack agents and builds a tested application."

These descriptions say what the agent *is*, not which user request it serves. That makes it harder for the user to pick the right agent in the Copilot agent picker, and harder for Codex to decide when to spawn it.

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub Copilot, Codex CLI
**Platforms/Environments:** VS Code, Linux, WSL2, macOS, Windows
**Features/Integrations:** Copilot custom agents, Codex custom agents, agent descriptions
**Versions/Constraints:** smaqit 3.10.0, Copilot agents can ask the user directly (vscode/askQuestions), Codex spawned agents return signals to the spawning session

## Design Decisions

- **Scope: the four entry agents only.** The five specification agents keep their delegation-only wording: Copilot `user-invocable: false` with a short description, and Codex "spawned by the … phase agent". It is accurate and keeps them from being chosen directly.
- **Same triggers as Claude.** Each description leads with the user request it serves, using the trigger set task 130 established for the Claude commands and agents (build/add/change/fix; deploy/redeploy/provision; test/validate/verify the deployed system; how smaqit works).
- **Copilot: trigger only.** Copilot phase agents have `vscode/askQuestions` and ask the user directly, so their descriptions carry no return-signal instruction.
- **Codex: trigger plus relay instruction.** A Codex agent's questions travel back to the spawning session, so the three phase descriptions also say: pass the user's full request and any resolved `PHASE-PARAMETERS` block, and if the result starts with `NEEDS-INPUT` or `REVIEW-REQUESTED`, ask the user and invoke the agent again with their answers, never answering it yourself.
- **Description only:** tools, `user-invocable`, and the agent bodies stay as they are.

## Implementation Steps

1. `.smaqit/definitions/agents/{development,deployment,validation,qa}.frontmatter.yaml`: rewrite the `copilot:` and `codex:` `description` values; verify every other key is unchanged.
2. Run `make -C installer prepare`, and check the generated `installer/agents-copilot/*.agent.md` frontmatter (YAML) and `installer/agents-codex/*.toml` (TOML) for the new descriptions.
3. Update the "Triggers: How a Session Picks smaqit Up" section of `docs/wiki/agent-tools-reference.md` to cover Copilot and Codex.
4. Add a `CHANGELOG.md` `[Unreleased]` entry.
5. Run `make -C installer test` and `make -C installer smoke-test`.

## Known Issues Triage
**Triaged:** 2026-10-12
**Tools searched:** GitHub Copilot, Codex CLI
**Result:** Advisory

### Advisory Issues
- [#41703 Remove unintended whitespace from spawn agent tool descriptions](https://github.com/openai/codex/issues/41703) — `openai/codex` — bug, subagent. Concerns the spawn tool's own description text, not custom-agent descriptions; noted only because it touches how Codex presents agents
- [#336761 Agent Host ignores the selected custom agent: picker resets to "Agent" and the agent prompt is never applied](https://github.com/microsoft/vscode/issues/336761) — `microsoft/vscode` — open. A host-side bug that affects whether a selected Copilot custom agent runs at all, independent of its description

### Historical (Closed)
- [#32418 Add Claude Code-style reusable subagent types to MultiAgentV2 spawn_agent](https://github.com/openai/codex/issues/32418) — `openai/codex` — closed
- [#340610 Agent picker and runtime desynchronize during Agent Host hydration and harness switching](https://github.com/microsoft/vscode/issues/340610) — `microsoft/vscode` — closed

### Search Warnings
- GitHub Copilot has no public issue repository; `microsoft/vscode` (the custom-agent host) was searched instead

## Acceptance Criteria

- [x] The Copilot and Codex descriptions of `development`, `deployment`, `validation`, and `qa` lead with "Use when…" and the user requests each serves
- [x] The three Codex phase-agent descriptions tell the spawning session to pass the full request and any `PHASE-PARAMETERS` block, and to relay `NEEDS-INPUT`/`REVIEW-REQUESTED` to the user without answering them itself
- [x] The five specification agents' descriptions, every agent's tools, `user-invocable`, and the agent bodies are unchanged
- [x] Generated Copilot frontmatter parses as YAML and generated Codex agents parse as TOML; `make test` and `make smoke-test` pass
- [x] The wiki trigger section and the CHANGELOG describe the change

## Findings

**Implementation approach:**
- Rewrote only the `copilot:` and `codex:` `description` lines of the four entry-agent definitions (two lines per file), using a script that located each platform block by its line-anchored key and asserted that every other key in every platform was unchanged
- Copilot descriptions lead with "Use when…" plus the phase summary (174–198 characters); Codex phase descriptions add the pass-request and relay-signal instruction (429–466 characters); QA is trigger-only on both
- Wiki triggers table gained a Copilot/Codex row; CHANGELOG `[Unreleased]` entry under Changed
- Verified: `make test`, `make smoke-test`, all generated Copilot frontmatter parses as YAML and all Codex agents parse as TOML, and the five specification agents and every agent body are identical to `origin/main`

**Decisions made:**
- The relay instruction goes only into Codex descriptions: Copilot phase agents have `vscode/askQuestions` and ask the user directly, while a Codex agent's signals return to the spawning session, as on Claude Code
- Specification agents keep their delegation-only wording: it is accurate, and Copilot's `user-invocable: false` already keeps them out of the picker

**Blockers encountered:**
- None; a first script pass failed safely before writing (it assumed a newline before the first platform key) and was corrected

**Follow-up identified:**
- Open VS Code issue #336761 (the Agent Host ignores a selected custom agent) can stop a Copilot custom agent from applying at all, independent of its description
- Behavioral confirmation that Codex spawns these agents from their descriptions is not covered here; task 132's headless end-to-end work is Claude Code-only

## Files to Create / Modify

| File | Action |
|------|--------|
| `.smaqit/definitions/agents/{development,deployment,validation,qa}.frontmatter.yaml` | Modify — `copilot:`/`codex:` descriptions |
| `docs/wiki/agent-tools-reference.md` | Modify — triggers section |
| `CHANGELOG.md` | Modify — `[Unreleased]` entry |

## Notes

- Follow-up from task 130 (v3.10.0), requested by the user on 2026-10-12 after the post-release live check.
