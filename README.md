![smaqit banner](assets/banner.png)

Welcome to smaQit, an orchestration toolkit for agentic software development. You describe requirements, specification agents generate stateful specs, then implementation agents turn those specs into working, tested and deployed applications. 

Built for teams that value auditability, clear boundaries, and reproducible workflows.

## Features

- **Lightweight** — Single binary, no dependencies. `smaqit init` scaffolds everything.
- **Traceable requirements** — Requirements captured in session context with full traceability from input to spec to implementation.
- **Stateful specs** — Specifications track lifecycle: draft → implemented → deployed → validated.
- **Bounded agents** — Each agent owns one layer or phase. No scope creep.
- **Self-validating** — Agents verify their own output before completion.
- **Spec-first** — Code follows specs, not the other way around.

## Compatibility

Currently supported:

| Platform | Status |
|----------|--------|
| GitHub Copilot (VS Code) | ✅ Supported |
| Claude Code | ✅ Supported |
| Other AI assistants | Planned |

## Getting Started

**Install:**

```bash
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit/main/install.sh | bash
```

**Initialize:**

```bash
smaqit init
```

**Build something:**

1. Open GitHub Copilot chat (or start Claude Code) and run `/smaqit.development`
2. The agent will validate your requirements are sufficient before proceeding — describe what you want to build when prompted
3. Watch specs generate, then code build

See the full [Mario Hello quickstart](docs/wiki/workflows/quickstart.md) for a complete walkthrough.

## Copilot Automation

smaqit includes a GitHub Action workflow that automatically installs smaqit before Copilot coding agent sessions. When Copilot coding agent runs in a GitHub Actions context, it automatically executes `.github/workflows/copilot-setup-steps.yml` by convention. No additional configuration needed.

## Commands

**CLI:**

| Command | Description |
|---------|-------------|
| `smaqit init` | Scaffold `.smaqit/`, `.github/`, and `.claude/` directories, plus `AGENTS.md`/`CLAUDE.md` |
| `smaqit status` | Show project state and spec coverage |
| `smaqit plan` | Show specs to process (for agents) |
| `smaqit validate` | Verify project structure integrity |
| `smaqit help` | Show detailed command help |
| `smaqit uninstall` | Remove smaqit from project |
| `smaqit version` | Show smaqit version |

**Agents** (invoke with `/` in GitHub Copilot chat or Claude Code):

| Agent | Purpose | Claude Code |
|-------|---------|-------------|
| `/smaqit.business` | Generate Business specifications | via `/smaqit.development` only |
| `/smaqit.functional` | Generate Functional specifications | via `/smaqit.development` only |
| `/smaqit.stack` | Generate Stack specifications | via `/smaqit.development` only |
| `/smaqit.infrastructure` | Generate Infrastructure specifications | via `/smaqit.deployment` only |
| `/smaqit.coverage` | Generate Coverage specifications | via `/smaqit.validation` only |
| `/smaqit.development` | Build working app from specs | ✅ direct command |
| `/smaqit.deployment` | Deploy to target environment | ✅ direct command |
| `/smaqit.validation` | Run tests against deployed system | ✅ direct command |
| `/smaqit.qa` | Answer questions about the smaqit framework | ✅ direct command |

On Claude Code, the five specification agents are Task-delegated subagents rather than standalone slash commands — the same `user-invocable: false` boundary they already have in GitHub Copilot.

### Reinstallation and Updates

Running `smaqit init` on an existing installation will:

- **Detect conflicts** — The installer checks which files would be overwritten
- **Preserve user data** — Your specs and custom extensions in `.smaqit/` are never touched
- **Prompt for confirmation** — If smaqit files would be overwritten, you'll be asked to confirm
- **Skip if no conflicts** — If only custom files exist, installation proceeds automatically

This makes it safe to:
- Upgrade to a new version of smaqit
- Reinstall after manual changes to agent or template files
- Add smaqit to projects with existing `.smaqit` extensions

## Documentation

- **[Quickstart](docs/wiki/workflows/quickstart.md)** — Build "Hello, Mario!" from scratch
- **[Team Alignment](docs/wiki/concepts/team-alignment.md)** — How layers map to Agile roles
- **[Wiki](docs/wiki/)** — Concepts, designs, patterns, workflows