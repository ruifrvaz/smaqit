# Agent Tools Reference

This document describes the platform-specific tool metadata used by smaqit's GitHub Copilot, Claude Code, and Codex agents.

**Source:** Verified against `github/awesome-copilot` repository examples and VS Code Copilot agent specification.

---

## Available Tools

### Code Navigation

| Tool | Description |
|------|-------------|
| `search/codebase` | Search and read codebase files (semantic codebase exploration) |
| `search` | Full-text search across the workspace |
| `search/usages` | Find all usages/references of a symbol |
| `search/searchResults` | Access and iterate over search results |

### File Editing

| Tool | Description |
|------|-------------|
| `edit/editFiles` | Create and edit files in the workspace |

### Web

| Tool | Description |
|------|-------------|
| `web/fetch` | Fetch and read content from web URLs |

### Diagnostics & Feedback

| Tool | Description |
|------|-------------|
| `read/problems` | Access VS Code Problems panel (errors, warnings) |
| `read/terminalLastCommand` | Read the last command run in the integrated terminal |
| `read/terminalSelection` | Read selected text in the integrated terminal |

### Execution

| Tool | Description |
|------|-------------|
| `runCommands` | Execute shell commands in the workspace terminal |
| `execute/runInTerminal` | Run a command directly in the integrated terminal |
| `execute/getTerminalOutput` | Get current output from the integrated terminal |
| `execute/runTests` | Run the project's test suite |
| `execute/testFailure` | Get details about test failures |

### Source Control

| Tool | Description |
|------|-------------|
| `changes` | Access staged and unstaged git changes |
| `activePullRequest` | Inspect the current active pull request diff |

### GitHub

| Tool | Description |
|------|-------------|
| `githubRepo` | Interact with the GitHub repository (issues, PRs, metadata) |

### Agent Orchestration

| Tool | Description |
|------|-------------|
| `agent/runSubagent` | Invoke another agent as a sub-workflow |

### VS Code Integration

| Tool | Description |
|------|-------------|
| `vscode/extensions` | Access VS Code extension information |
| `vscode/openSimpleBrowser` | Open a URL in VS Code's Simple Browser |
| `vscode/vscodeAPI` | Access the VS Code API documentation |
| `findTestFiles` | Locate test files in the workspace |

---

## Tool Name Migration

Previous versions of smaqit agent files used short tool names from an older API. This table maps old names to their current replacements.

| Old Name | New Name | Notes |
|----------|----------|-------|
| `read` | `search/codebase` | Incorrect bare name — use namespaced version |
| `edit` | `edit/editFiles` | Namespaced version required |
| `usages` | `search/usages` | Moved to `search` namespace |
| `fetch` | `web/fetch` | Moved to `web` namespace |
| `problems` | `read/problems` | Moved to `read` namespace |
| `testFailure` | `execute/testFailure` | Moved to `execute` namespace |
| `runTests` | `execute/runTests` | Moved to `execute` namespace |
| `runSubagent` | `agent/runSubagent` | Correct namespaced tool name |
| `todos` | *(removed)* | No direct replacement in current API |
| `runCommands` | `runCommands` | Unchanged |
| `changes` | `changes` | Unchanged |
| `search` | `search` | Unchanged |

---

## Standard Tool Sets by Agent Type

### Specification Agents (business, functional, stack, infrastructure, coverage)

Full specification toolset — file editing, codebase search, terminal access, sub-agent delegation, and web fetch.

```yaml
tools: ['execute/getTerminalOutput', 'execute/awaitTerminal', 'execute/runInTerminal', 'read/readFile', 'agent/runSubagent', 'edit/createDirectory', 'edit/createFile', 'edit/createJupyterNotebook', 'edit/editFiles', 'edit/editNotebook', 'edit/rename', 'search/changes', 'search/codebase', 'search/fileSearch', 'search/listDirectory', 'search/textSearch', 'search/searchSubagent', 'search/usages', 'web/fetch', 'todo']
```

### Implementation Agents (development, deployment, validation)

Full execution: editing, running commands, tests, invoking sub-agents.

```yaml
tools: ['edit/editFiles', 'search', 'runCommands', 'read/problems', 'changes', 'execute/testFailure', 'execute/runTests', 'agent/runSubagent']
```

### Read-Only / Q&A Agents (qa, doc-helper)

Read and search only — no file modification.

```yaml
tools: ['search/codebase', 'search', 'web/fetch']
```

---

## Claude Code Tool Mapping

Agent tool lists live in `.smaqit/definitions/agents/<name>.frontmatter.yaml` (the shared body is in `agents/<name>.md`) and are compiled by `scripts/generate-agents.py` (see `framework/TEMPLATES.md`). When editing an agent's Copilot tool list, keep the Claude Code side in sync using this mapping:

| Copilot tool ID(s) | Claude Code tool |
|---|---|
| `edit/editFiles`, `edit/createFile`, `edit/createDirectory`, `edit/rename` | `Write`, `Edit` |
| `execute/runInTerminal`, `execute/getTerminalOutput`, `execute/sendToTerminal`, `execute/awaitTerminal`, `runCommands`, `execute/runTests`, `execute/testFailure` | `Bash` |
| `read/readFile`, `read/problems`, `read/terminalLastCommand`, `read/terminalSelection`, `read/viewImage` | `Read` |
| `smaqit-plantuml/check_syntax`, `smaqit-plantuml/render_diagram` | `mcp__smaqit-plantuml__check_syntax`, `mcp__smaqit-plantuml__render_diagram` |
| `search`, `search/codebase`, `search/usages`, `search/textSearch`, `search/fileSearch`, `search/listDirectory` | `Grep`, `Glob` |
| `web`, `web/fetch` | `WebFetch` |
| `agent`, `agent/runSubagent` | `Task` |
| `todo` | `TodoWrite` |
| `changes`, `activePullRequest`, `githubRepo`, `search/changes` | `Bash` (shell out to `git`/`gh` — no native equivalent) |
| `search/searchSubagent` | `Task` |
| `vscode/askQuestions` | none inside a subagent: Claude Code removes `AskUserQuestion` from every subagent. smaqit agents return `NEEDS-INPUT` instead (see Asking the User) |
| `vscode/memory`, `edit/createJupyterNotebook`, `edit/editNotebook` | dropped — no Claude Code equivalent |

There is no Claude Code equivalent of `user-invocable: false`. The same effect — an agent reachable only via delegation, never as a direct user command — is reproduced by simply not creating a `commands/<name>.md` for that agent, so no `.claude/commands/` entry exists for it (see the five specification agents: business, functional, stack, infrastructure, coverage).

**Skills:** a Claude Code subagent whose `tools:` list omits `Skill` cannot invoke skills at all. Each smaqit agent therefore preloads its own input skill through `skills:` in its `claude:` metadata, which injects the skill's full content at startup. `scripts/generate-agents.py` fails `make prepare` when an agent body names a `smaqit.input-*` skill that its Claude metadata neither preloads nor can invoke.

## Asking the User: Return Signals

Agents usually run as subagents, and a subagent often cannot ask the user anything. Questions and review requests therefore travel up to whichever caller can reach the user (`framework/AGENTS.md`, User Interaction):

| Signal | Returned by | Meaning |
|---|---|---|
| `NEEDS-INPUT` | An input skill or agent that found a gap it cannot ask about | Lists every open question, each with an id and why it's needed. Nothing was written |
| `REVIEW-REQUESTED` | A phase agent in Assisted mode, after a specification layer | Lists the produced spec and design paths, the iteration count, and the specs approved so far |

Callers pass phase agents an explicit `PHASE-PARAMETERS` block (`mode`, `specification_mode`, `spec_paths`, `approved_spec_paths`, deployment parameters, and `answers` on resume), produced by the phase input skill.

| Platform | Who asks the user | How the run continues |
|---|---|---|
| Claude Code | The main conversation: the `/smaqit.<phase>` command body, or an orchestrating skill such as `smaqit.feature-new` | The main conversation resumes the same subagent by sending its agent ID the `answers` (SendMessage). It falls back to invoking the agent again with `answers` and `approved_spec_paths` |
| GitHub Copilot | Any smaqit agent with `vscode/askQuestions` (all eight non-QA agents); a `runSubagent` child that cannot ask returns the signal to its parent | The parent agent asks, then invokes the child again with `answers` |
| Codex | The session that spawned the agent | The spawning session invokes the agent again with `answers` |

## Codex Agent Metadata

Codex custom agents are generated into `installer/agents-codex/*.toml` and installed once per user to `~/.codex/agents/` (or `$CODEX_HOME/agents/`). Each file contains `name`, `description`, and `developer_instructions`. The five design-authoring specification agents also enable `tools.view_image` and declare the required `smaqit-plantuml` MCP tools. Host registration is separate: `smaqit init` manages `[mcp_servers.smaqit-plantuml]` in trusted `.codex/config.toml`, while Claude Code uses root `.mcp.json` and VS Code uses `.vscode/mcp.json`. Implementation agents receive neither capability and consume validated PlantUML source after the automatic plan gate. Other capabilities continue to use the tools and sandbox permissions available in the active Codex session.

All nine smaqit agents are installed as named Codex user-level agents. Phase agents refer to specification agents by name and ask Codex to spawn them. The shared skills are installed once per user under `~/.agents/skills/`, where Codex can discover them through `/skills` or `$` mentions.
