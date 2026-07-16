# Project Compendium

Last updated: 2026-07-16 | Total entries: 5

## Hooks

| Question | Answer | Last Updated | Sessions |
|----------|--------|--------------|----------|
| Do VS Code Copilot hooks fire for `runSubagent` tool calls? | No. `SubagentStart` and `SubagentStop` only fire for VS Code native agent delegation (e.g., `@agent` in chat input). They do NOT fire when a subagent is invoked via the `runSubagent` tool inside an agent. `PostToolUse` does fire for all tool calls including `run_in_terminal`, `read_file`, etc. | 2026-05-17 | 1 |
| What hook format does VS Code Copilot require? | VS Code uses PascalCase event names (`SubagentStart`, `PostToolUse`), the `command` key (not `bash`), no `version` field, and a `hookSpecificOutput` wrapper object: `{"hookSpecificOutput": {"hookEventName": "SubagentStart", "additionalContext": "..."}}`. Inline commands with `echo` work reliably; external scripts also work once the hook pipeline is active. | 2026-05-17 | 1 |

## Agent Orchestration

| Question | Answer | Last Updated | Sessions |
|----------|--------|--------------|----------|
| What Microsoft AI agent orchestration pattern does smaqit's phase orchestration match? | Sequential Workflow (phase agent as orchestrator invoking spec agents in fixed order). The assisted mode (user reviews each spec) maps to Maker-Checker. Spec agent invocations are Nested Composition. Microsoft guidance: deterministic routing must be hardcoded — never delegated to agents at runtime. | 2026-05-17 | 1 |

## Claude Code Support

| Question | Answer | Last Updated | Sessions |
|----------|--------|--------------|----------|
| What's the Claude Code equivalent of `copilot-setup-steps.yml`? | There isn't a direct one — no magic-filename auto-detection exists in Claude Code. Copilot's coding agent auto-detects that exact filename as a GitHub-native convention. Closest local equivalent: a `SessionStart` hook in `.claude/settings.json` (fires when a `claude` session starts/resumes, can run setup commands). Closest cloud/CI equivalent: `anthropics/claude-code-action`, but it requires manually adding setup steps to your own workflow YAML — no auto-detection. `CLAUDE.md` is NOT the equivalent — it's just project-context markdown loaded into the system prompt, not a bootstrap mechanism. smaqit ruled this out of scope since GitHub Actions deployments continue to use Copilot. | 2026-07-16 | 1 |
| What is `installer/commands-claude/`? | The gitignored compiled-output directory holding Claude Code slash command files, generated from `commands/*.md` at repo root by `scripts/generate-agents.py`, installed to `.claude/commands/` in target projects. It exists because Claude Code needs a separate command file per slash command, unlike Copilot where an agent's own `name:` frontmatter field doubles as its `/name` invocation — so there is no `commands-copilot` equivalent. Only agents meant to be directly user-invocable get a command file (development/deployment/validation/qa); the five spec agents do not, reproducing Copilot's `user-invocable: false` on the Claude Code side. | 2026-07-16 | 1 |
