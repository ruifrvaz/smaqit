# Extending smaqit

This guide covers how to extend the smaqit framework by creating new agents, modifying principles, and contributing framework improvements.

## Overview

smaqit's source is structured as a three-level compilation chain:

- **Level 0 (Framework)** — `framework/*.md` — Philosophy and principles (WHY/WHAT)
- **Level 1 (Templates + Compilation Rules)** — `templates/` — Structure and directives
- **Level 2 (Agents)** — `agents/*.md` plus `.smaqit/definitions/agents/*.frontmatter.yaml` — Canonical bodies and platform metadata compiled for users

Extension means working in this repo by editing source files and following the compilation convention. The `smaqit.qa` agent is the primary tool for validating framework consistency.

## Repository Structure

```
smaqit/
├── framework/              # L0: principles (SMAQIT.md, LAYERS.md, PHASES.md, etc.)
├── templates/
│   ├── specs/              # Specification templates (installed to user projects)
│   └── agents/
│       └── compiled/       # L1 compilation rules (*.rules.md per layer/phase)
├── agents/                 # L2: compiled agents (shipped to user projects)
├── skills/                 # Input and utility skills (shipped to user projects)
├── installer/              # Go CLI tool
└── docs/wiki/              # Human-readable rationale and guides
```

## When to Extend

- **Modify agent behavior** — Edit `agents/` source files, their `.smaqit/definitions/agents/` metadata, or their `templates/agents/compiled/*.rules.md` rules
- **Change framework principles** — Edit `framework/*.md` and propagate changes downward
- **Add a new skill** — Create `skills/<name>/SKILL.md`; generation installs it for Copilot, Claude Code, and Codex
- **Add a new agent** — Create compilation rules in `templates/agents/compiled/`, write the canonical `agents/` body, and add platform metadata under `.smaqit/definitions/agents/`

## Compilation Convention

The L0→L1→L2 chain is a convention, not automated tooling. Follow it when making changes:

1. **Principle changes (L0)** — Edit `framework/*.md`. No directives, no file paths. Philosophical only.
2. **Directive changes (L1)** — Update the relevant `templates/agents/compiled/*.rules.md` file.
3. **Agent update (L2)** — Edit `agents/<name>.md` to reflect the new directives and its platform metadata when discovery fields or placeholders change.
4. **Generate** — Run `make -C installer prepare` to compile all three platform formats.
5. **Validate** — Use the `smaqit.qa` agent to check consistency across levels.

Commit sequentially: L0 change → L1 change → L2 change, each in its own commit with prefix `L0:`, `L1:`, `L2:`.

## Using the QA Agent

`smaqit.qa` is the framework validation agent. Run it after changes to:

- Detect level contamination (directives in L0, philosophy in L2)
- Verify agent structure matches templates
- Check compilation chain integrity
- Run full framework consistency checks before releasing

## Installer Sync

After editing source files, regenerate the installer staging trees so builds pick up the changes:

```bash
make -C installer prepare
```

The generator compiles agents into `installer/agents-copilot/`, `installer/agents-claude/`, and `installer/agents-codex/`; skills are copied into the corresponding three staging trees. These outputs are ephemeral, gitignored build inputs embedded in the Go binary.

## Adding a New Layer Agent

1. Create `templates/agents/compiled/<layer>.rules.md` with L0→L1 compilation rules
2. Write `agents/<layer>.md` following the spec agent structure
3. Create `skills/smaqit.input-<layer>/SKILL.md` as the input validation skill
4. Create `.smaqit/definitions/agents/<layer>.frontmatter.yaml` with `copilot`, `claude`, and `codex` metadata and all platform placeholder values
5. Run `make -C installer test` and validate with the `smaqit.qa` agent

## Best Practices

### Level Boundaries

- **L0 files** — No directives (MUST/MUST NOT/SHOULD), no file paths, no workflows. Philosophical only.
- **L1 files** — Transform L0 concepts into directives and structure.
- **L2 files** — Self-contained and executable. No references to L0 or L1.

### Validation

Run the `smaqit.qa` agent after every significant change and before committing.

## Contributing

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for contribution guidelines.

## Further Reading

- [Level Up Compilation Architecture](../designs/level-up-compilation.md) — How the L0→L1→L2 chain works
- [Hierarchical Levels](../designs/hierarchical-levels.md) — Why four levels exist
- [Quickstart](quickstart.md) — Using smaqit in a project
