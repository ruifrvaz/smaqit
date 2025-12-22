# smaqit

## What is it?

A Spec-driven agent orchestration kit where AI agents write specifications, then implement from those specs.

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

```bash
# Initialize in your project
smaqit init

# Start developing
smaqit develop

# Deploy
smaqit deploy

# Validate
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

1. **Develop** — Write specs (business → functional → stack), then build
2. **Deploy** — Write infrastructure spec, then deploy
3. **Validate** — Write coverage spec, then test

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
│   ├── framework/LAYERS.md      # Layer definitions (Business → Coverage)
│   └── framework/PHASES.md      # Phase workflows (Develop → Validate)
│
├── Level 1: Templates (Structure)
│   ├── framework/TEMPLATES.md   # Template rules
│   ├── templates/specs/         # Specification templates (5)
│   └── templates/agents/        # Agent templates (2)
│
├── Level 2: Agents & Artifacts (Instances)
│   ├── framework/AGENTS.md      # Agent behaviors
│   ├── framework/ARTIFACTS.md   # Artifact rules
│   ├── agents/*.agent.md        # Agent definitions (8)
│   └── specs/**/*.md            # Specification documents
│
└── Level 3: Application (Output)
    └── The built system
```

**Level dependencies:**
- Level 1 consumes Level 0 (templates follow framework rules)
- Level 2 consumes Level 1 (agents/specs follow templates)
- Level 3 consumes Level 2 (application follows specs)

## Documentation Structure

smaqit documentation serves three audiences with distinct separation of concerns:

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

**No rationale, no examples, no history**—just instructions.

### Wiki (`docs/wiki/`)
**Audience:** Human developers  
**Purpose:** Context and rationale

These files explain WHY the framework is designed this way:
- `concepts/` — Core concepts explained
- `design-decisions/` — Why we chose these patterns
- `patterns/` — Common usage patterns
- `workflows/` — Step-by-step processes

**For learning and understanding**, not for LLM consumption.

### README (this file)
**Audience:** Users and contributors  
**Purpose:** Project overview

High-level introduction:
- What smaqit is
- How to install and use it
- Architecture overview
- Contributor guidelines

**Entry point for humans**, links to framework and wiki as needed.

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

## License

MIT
