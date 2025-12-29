# smaqit

## What is it?

A Spec-driven agent orchestration kit where AI agents write specifications, then implement from those specs. **Work in phases**—each phase generates specs and implements them together for fast feedback.

## How it Works

**Phase-First Workflow (Recommended):**
```
Phase 1 (Develop):  Business → Functional → Stack specs → Build → Working app
Phase 2 (Deploy):   Infrastructure spec → Deploy → Running system
Phase 3 (Validate): Coverage spec → Test → Validation report
```

**Spec-First Workflow (Alternative):**
```
Generate all specs: Business → Functional → Stack → Infrastructure → Coverage
Execute phases:     Develop → Deploy → Validate
```

Phase-first gives you faster feedback. Spec-first works for upfront design requirements.

## Getting Started

### Installation

**Download the latest release for your platform:**

Visit [Releases](https://github.com/ruifrvaz/smaqit/releases) and download the appropriate binary:

- **Linux**: `smaqit_linux_amd64`
- **macOS Intel**: `smaqit_darwin_amd64`
- **macOS Apple Silicon**: `smaqit_darwin_arm64`
- **Windows**: `smaqit_windows_amd64.exe`

**Make it executable (Linux/macOS):**
```bash
chmod +x smaqit_*
mv smaqit_* /usr/local/bin/smaqit
```

**Or add to PATH (Windows):**
```powershell
# Move to a directory in your PATH
move smaqit_windows_amd64.exe C:\Windows\smaqit.exe
```

### Usage

**Phase-First Workflow (Recommended):**
```bash
# Initialize in your project
smaqit init

# Phase 1: Develop
# Generate specs, then build
smaqit develop

# Phase 2: Deploy
# Generate infrastructure spec, then deploy
smaqit deploy

# Phase 3: Validate
# Generate coverage spec, then test
smaqit validate
```

**Spec-First Workflow (Alternative):**
```bash
# Initialize
smaqit init

# Generate all specifications first
# (invoke agents via GitHub Copilot chat)

# Then execute phases
smaqit develop
smaqit deploy
smaqit validate
```

## Commands

| Command | Description |
|---------|-------------|
| `smaqit init` | Scaffold `.smaqit/` and `.github/agents/` |
| `smaqit develop` | Run develop phase |
| `smaqit deploy` | Run deploy phase |
| `smaqit validate` | Run validate phase |
| `smaqit status` | Show current state |

---

## Layers

1. **Business** — Use cases, actors, goals
2. **Functional** — Behaviors, contracts, data models
3. **Stack** — Languages, frameworks, tools
4. **Infrastructure** — Compute, networking, observability
5. **Coverage** — Tests against deployed app

## Phases

**Phases are the primary workflow unit.** Each phase includes specifications and implementation together.

1. **Develop** — Generate specs (business → functional → stack), then build → working application
2. **Deploy** — Generate infrastructure spec, then deploy → running system
3. **Validate** — Generate coverage spec, then test → validation report

**Phase-first is recommended** for faster feedback. Complete each phase before moving to the next. Alternatively, you can generate all specs first (spec-first), but implementation still happens in phases.

## Team Alignment

smaqit layers align with Agile/Scrum team roles, enabling specialists to work in their domain.

| Role | Layer | Focus |
|------|-------|-------|
| Stakeholders | Input | Requirements and business needs |
| Product Owner | Business | Why, for whom, success criteria |
| Engineers | Functional | What behaviors, contracts, data models |
| Software Developer | Stack | With what languages, frameworks, tools |
| DevOps | Infrastructure | Where and how it runs |
| Testers | Coverage | How we verify it works |

Layer boundaries respect role boundaries:
- Product Owners define *what* success looks like, not *how* to build it
- Engineers translate business goals into system behaviors
- Developers choose technologies that satisfy functional requirements
- Each role focuses on their expertise without immediate cross-concerns

## Architecture

smaqit kit is organized in hierarchical levels:

```
smaqit
├── Level 0: Framework (Foundation)
│   ├── framework/SMAQIT.md      # Core principles
│   ├── framework/LAYERS.md      # Layer definitions
│   ├── framework/PHASES.md      # Phase workflows
│   ├── framework/TEMPLATES.md   # Template rules
│   ├── framework/AGENTS.md      # Agent behaviors
│   ├── framework/ARTIFACTS.md   # Artifact rules
│   └── framework/PROMPTS.md     # Prompt architecture
│
├── Level 1: Templates (Structure)
│   ├── templates/specs/         # Specification templates (5)
│   ├── templates/prompts/       # Prompt templates (2)
│   └── templates/agents/        # Agent templates (2)
│
├── Level 2: Agents & Artifacts (Instances)
│   ├── agents/*.agent.md        # Agent definitions (8)
│   ├── prompts/*.prompt.md      # Prompt files (8)
│   └── specs/**/*.md            # Specification documents
│
└── Level 3: Application (Output)
    └── The built system
```

**Level dependencies:**
- Level 1 consumes Level 0 (templates follow framework rules)
- Level 2 consumes Level 1 (agents/specs/prompts follow templates)
- Level 3 consumes Level 2 (application follows specs)

## Documentation Structure

### Framework Files (`framework/`)
**Audience:** LLM agents  
**Purpose:** Pure execution instructions

These files contain ONLY what agents need to execute workflows:
- Core principles (SMAQIT.md)
- Layer definitions (LAYERS.md)
- Phase workflows (PHASES.md)
- Template rules (TEMPLATES.md)
- Agent behaviors (AGENTS.md)
- Artifact rules (ARTIFACTS.md)

### Wiki (`docs/wiki/`)
**Audience:** Human developers  
**Purpose:** Context and rationale

These files explain WHY the framework is designed this way:
- `concepts/` — Core concepts explained
- `designs/` — Why we chose these patterns
- `patterns/` — Common usage patterns
- `workflows/` — Step-by-step processes

**See [User vs Agent Documentation](docs/wiki/concepts/user-vs-agent-documentation.md) for detailed guidance on this distinction.**

---

## Contributors

### Building from Source

**Prerequisites:**
- Go 1.25 or later
- make (optional, can use build scripts)

**Build:**
```bash
# Clone the repository
git clone https://github.com/ruifrvaz/smaqit.git
cd smaqit/installer

# Build for your platform
make build

# Or build for all platforms
make build-all

# Or use shell scripts
./build.sh build          # Unix-like systems
build.bat build           # Windows
```

**Development:**
```bash
# Show version that would be built
make version

# Install to ~/.local/bin (adds to PATH if needed)
make install

# Remove binary (prompts for PATH and artifact cleanup)
make uninstall

# See all available targets
make help
```

## Testing

**Test location:** `installer/test/` (standardized test directory)

**Automated end-to-end testing:** See `.github/agents/smaqit.user-testing.agent.md`

**Testing philosophy and manual workflows:** See `docs/wiki/workflows/testing-smaqit.md`

## License

MIT
