# GitHub Copilot Agent Tools Reference

This document lists all tools available to GitHub Copilot custom agents (`.agent.md` files) running in VS Code agent mode.

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
