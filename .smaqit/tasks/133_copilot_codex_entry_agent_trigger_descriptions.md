---
status: Not Started
created: "2026-10-12"
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

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] The Copilot and Codex descriptions of `development`, `deployment`, `validation`, and `qa` lead with "Use when…" and the user requests each serves
- [ ] The three Codex phase-agent descriptions tell the spawning session to pass the full request and any `PHASE-PARAMETERS` block, and to relay `NEEDS-INPUT`/`REVIEW-REQUESTED` to the user without answering them itself
- [ ] The five specification agents' descriptions, every agent's tools, `user-invocable`, and the agent bodies are unchanged
- [ ] Generated Copilot frontmatter parses as YAML and generated Codex agents parse as TOML; `make test` and `make smoke-test` pass
- [ ] The wiki trigger section and the CHANGELOG describe the change

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
| `.smaqit/definitions/agents/{development,deployment,validation,qa}.frontmatter.yaml` | Modify — `copilot:`/`codex:` descriptions |
| `docs/wiki/agent-tools-reference.md` | Modify — triggers section |
| `CHANGELOG.md` | Modify — `[Unreleased]` entry |

## Notes

- Follow-up from task 130 (v3.10.0), requested by the user on 2026-10-12 after the post-release live check.
