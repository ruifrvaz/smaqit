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
| `vscode/memory`, `vscode/askQuestions`, `edit/createJupyterNotebook`, `edit/editNotebook` | dropped — no Claude Code equivalent |

There is no Claude Code equivalent of `user-invocable: false`. The same effect — an agent reachable only via delegation, never as a direct user command — is reproduced by simply not creating a `commands/<name>.md` for that agent, so no `.claude/commands/` entry exists for it (see the five specification agents: business, functional, stack, infrastructure, coverage).

## Codex Agent Metadata

Codex project custom agents are generated into `.codex/agents/*.toml`. Each file contains `name`, `description`, and `developer_instructions`. The five design-authoring specification agents also enable `tools.view_image` and declare the required project-local `smaqit-plantuml` MCP server. Implementation agents receive neither capability and consume validated PlantUML source after the automatic plan gate. Other capabilities continue to use the tools and sandbox permissions available in the active Codex session.

All nine smaqit agents are installed as named Codex project agents. Phase agents refer to specification agents by name and ask Codex to spawn them. The same 26 shared skills are installed under `.agents/skills/`, where Codex can discover them through `/skills` or `$` mentions.
