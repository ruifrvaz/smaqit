# Project Compendium

Last updated: 2026-05-17 | Total entries: 3

## Hooks

| Question | Answer | Last Updated | Sessions |
|----------|--------|--------------|----------|
| Do VS Code Copilot hooks fire for `runSubagent` tool calls? | No. `SubagentStart` and `SubagentStop` only fire for VS Code native agent delegation (e.g., `@agent` in chat input). They do NOT fire when a subagent is invoked via the `runSubagent` tool inside an agent. `PostToolUse` does fire for all tool calls including `run_in_terminal`, `read_file`, etc. | 2026-05-17 | 1 |
| What hook format does VS Code Copilot require? | VS Code uses PascalCase event names (`SubagentStart`, `PostToolUse`), the `command` key (not `bash`), no `version` field, and a `hookSpecificOutput` wrapper object: `{"hookSpecificOutput": {"hookEventName": "SubagentStart", "additionalContext": "..."}}`. Inline commands with `echo` work reliably; external scripts also work once the hook pipeline is active. | 2026-05-17 | 1 |

## Agent Orchestration

| Question | Answer | Last Updated | Sessions |
|----------|--------|--------------|----------|
| What Microsoft AI agent orchestration pattern does smaqit's phase orchestration match? | Sequential Workflow (phase agent as orchestrator invoking spec agents in fixed order). The assisted mode (user reviews each spec) maps to Maker-Checker. Spec agent invocations are Nested Composition. Microsoft guidance: deterministic routing must be hardcoded — never delegated to agents at runtime. | 2026-05-17 | 1 |
