![smaqit banner](assets/banner.png)

Welcome to smaQit, an agent orchestration toolkit for developers building software using Spec Driven Development. You describe requirements, specification agents generate stateful specs, then implementation agents turn those specs into working, tested and deployed applications. Built for teams that value auditability, clear boundaries, and reproducible workflows.

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

## Automated Setup for GitHub Copilot

If you're using GitHub Copilot coding agent, smaqit can be installed automatically before each session:

1. **Copy the setup workflow** to your project:
   ```bash
   mkdir -p .github/workflows
   curl -o .github/workflows/copilot-setup.yml \
     https://raw.githubusercontent.com/ruifrvaz/smaqit/main/.github/workflows/copilot-setup.yml
   ```

2. **Configure Copilot** to use it by adding to `.github/copilot-instructions.md`:
   ```markdown
   ## Environment Setup
   
   Before starting work, ensure smaqit is installed:
   - Setup steps: `.github/workflows/copilot-setup.yml`
   ```

3. **That's it!** Copilot will automatically install smaqit if not present.

The workflow:
- ✅ Detects existing installations (idempotent)
- ✅ Downloads latest smaqit release
- ✅ Runs `smaqit init` if needed
- ✅ Validates installation success
- ✅ Provides clear status feedback

This eliminates manual installation for Copilot workflows.

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