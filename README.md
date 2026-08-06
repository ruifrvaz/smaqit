![smaqit banner](assets/banner.png)

Welcome to smaQit, an orchestration toolkit for agentic software development. You describe requirements, specification agents generate stateful specs, then implementation agents turn those specs into working, tested and deployed applications. 

Built for teams that value auditability, clear boundaries, and reproducible workflows.

## Features

- **One installer** — The released binary embeds smaqit's pinned PlantUML MCP, JavaScript/WASM renderer, font, templates, and agent integrations. Node.js 22+ is the only host prerequisite; the materialized local runtime under `.smaqit/tools/` is automatically Git-ignored.
- **Visual design gates** — Minimal UML-style PlantUML/PNG design pairs are linked to every active specification, visually reviewed by specification agents, and automatically gated before implementation agents consume their PlantUML source.
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
| OpenAI Codex | ✅ Supported |
| Other AI assistants | Planned |

## Getting Started

**Prerequisite:** Node.js 22 or newer. Consumer installation and execution never run npm/npx or resolve packages over the network.

**Install:**

```bash
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit/main/install.sh | bash
```

**Initialize:**

```bash
smaqit init
```

**Build something:**

1. Open GitHub Copilot chat or Claude Code and run `/smaqit.development`; in Codex, ask it to spawn the `smaqit.development` agent
2. The agent will validate your requirements are sufficient before proceeding — describe what you want to build when prompted
3. Watch specs generate, then code build

See the full [Mario Hello quickstart](docs/wiki/workflows/quickstart.md) for a complete walkthrough.

Once your MVP is deployed, use `smaqit.feature-new` for iterative post-MVP feature cycles — it applies the same task-per-phase discipline without requirements extraction, from-scratch specs, or a dev-VM sweep. Deployment goes through the existing CI/CD pipeline via a pull request as the human approval gate.

## Copilot Automation

smaqit includes a GitHub Action workflow that automatically installs smaqit before Copilot coding agent sessions. When Copilot coding agent runs in a GitHub Actions context, it automatically executes `.github/workflows/copilot-setup-steps.yml` by convention. No additional configuration needed.

## Commands

**CLI:**

| Command | Description |
|---------|-------------|
| `smaqit init` | Scaffold `.smaqit/`, `.github/`, `.claude/`, `.codex/`, and `.agents/`, plus `AGENTS.md`/`CLAUDE.md` |
| `smaqit status` | Show project state and spec coverage |
| `smaqit plan` | Show specs to process (for agents) |
| `smaqit validate` | Verify project structure integrity |
| `smaqit design render <file>` | Syntax-check PlantUML and render its canonical PNG |
| `smaqit design attest <file>` | Record the active agent's visual review against current hashes |
| `smaqit design validate [file]` | Run structural, PlantUML, and visual-attestation gates |
| `smaqit mcp verify` | Verify VS Code, Claude Code, and Codex PlantUML MCP registration plus local stdio transport |
| `smaqit help` | Show detailed command help |
| `smaqit uninstall` | Remove smaqit from project |
| `smaqit update` | Update smaqit to the latest release |
| `smaqit version` | Show smaqit version |

**Agents** (invoke with `/` in GitHub Copilot chat or Claude Code; ask Codex to spawn the named agent):

| Agent | Purpose | Claude Code | Codex |
|-------|---------|-------------|-------|
| `smaqit.business` | Generate Business specifications | via development only | named subagent |
| `smaqit.functional` | Generate Functional specifications | via development only | named subagent |
| `smaqit.stack` | Generate Stack specifications | via development only | named subagent |
| `smaqit.infrastructure` | Generate Infrastructure specifications | via deployment only | named subagent |
| `smaqit.coverage` | Generate Coverage specifications | via validation only | named subagent |
| `smaqit.development` | Build working app from specs | direct command | named subagent |
| `smaqit.deployment` | Deploy to target environment | direct command | named subagent |
| `smaqit.validation` | Run tests against deployed system | direct command | named subagent |
| `smaqit.qa` | Answer questions about the smaqit framework | direct command | named subagent |

On Claude Code, the five specification agents are Task-delegated subagents rather than standalone slash commands — the same `user-invocable: false` boundary they already have in GitHub Copilot. Codex discovers all nine project agents from `.codex/agents/*.toml` and the 26 repository skills from `.agents/skills/`; skills can be selected with `/skills` or mentioned with `$`.

### Reinstallation and Updates

Running `smaqit init` on an existing installation will:

- **Detect conflicts** — The installer checks which files would be overwritten
- **Preserve user data** — Your specs and custom extensions in `.smaqit/` are never touched
- **Preserve shared client configuration** — Unrelated `.vscode/mcp.json`, `.mcp.json`, and `.codex/config.toml` content is retained; only smaqit's exact PlantUML registration is managed
- **Prompt for confirmation** — If smaqit files would be overwritten, you'll be asked to confirm
- **Skip if no conflicts** — If only custom files exist, installation proceeds automatically
- **Ignore managed runtime** — The installer adds the narrow `.smaqit/tools/` rule to your root `.gitignore`, preserving existing rules; canonical `docs/designs/` artifacts remain tracked

This makes it safe to:
- Upgrade to a new version of smaqit
- Reinstall after manual changes to agent or template files
- Add smaqit to projects with existing `.smaqit` extensions

## Documentation

- **[Quickstart](docs/wiki/workflows/quickstart.md)** — Build "Hello, Mario!" from scratch
- **[Team Alignment](docs/wiki/concepts/team-alignment.md)** — How layers map to Agile roles
- **[Wiki](docs/wiki/)** — Concepts, designs, patterns, workflows
