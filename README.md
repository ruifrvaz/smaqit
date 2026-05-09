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

1. Open GitHub Copilot chat and run `/smaqit.development`
2. Describe your requirements in the conversation when the agent asks
3. Watch specs generate, then code build

See the full [Mario Hello quickstart](docs/wiki/workflows/quickstart.md) for a complete walkthrough.

## Copilot Automation

smaqit includes a GitHub Action workflow that automatically installs smaqit before Copilot coding agent sessions. When Copilot coding agent runs in a GitHub Actions context, it automatically executes `.github/workflows/copilot-setup-steps.yml` by convention. No additional configuration needed.

## Commands

**CLI:**

| Command | Description |
|---------|-------------|
| `smaqit init` | Scaffold `.smaqit/` and `.github/` directories |
| `smaqit status` | Show project state and spec coverage |
| `smaqit plan` | Show specs to process (for agents) |
| `smaqit validate` | Verify project structure integrity |
| `smaqit help` | Show detailed command help |
| `smaqit uninstall` | Remove smaqit from project |
| `smaqit version` | Show smaqit version |

**Agents** (invoke in GitHub Copilot chat with `/`):

| Agent | Purpose |
|-------|---------|
| `/smaqit.development` | Build working app from specs |
| `/smaqit.deployment` | Deploy to target environment |
| `/smaqit.validation` | Run tests against deployed system |

Run `smaqit help` for all specification agents (`/smaqit.business`, `/smaqit.functional`, etc.).

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