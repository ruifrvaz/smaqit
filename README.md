![smaqit banner](assets/banner.png)

Welcome to smaQit, an agent orchestration toolkit for developers who wish to build software using Spec Driven Development. You describe requirements in prompt files, specification agents generate traceable specs across business, functional, and technical layers, then implementation agents turn those specs into working, tested applications. Built for teams that value auditability, clear boundaries, and reproducible workflows.

## Features

- **Lightweight** — Single binary, no dependencies. `smaqit init` scaffolds everything.
- **Auditable prompts** — Requirements captured in versioned prompt files with full traceability.
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

1. Fill `.github/prompts/smaqit.business.prompt.md` with your requirements
2. Open GitHub Copilot chat and run `/smaqit.development`
3. Watch specs generate, then code build

See the full [Mario Hello quickstart](docs/wiki/workflows/quickstart.md) for a complete walkthrough.

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

## Documentation

- **[Quickstart](docs/wiki/workflows/quickstart.md)** — Build "Hello, Mario!" from scratch
- **[Team Alignment](docs/wiki/concepts/team-alignment.md)** — How layers map to Agile roles
- **[Wiki](docs/wiki/)** — Concepts, designs, patterns, workflows