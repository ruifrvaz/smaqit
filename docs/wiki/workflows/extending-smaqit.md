# Extending smaqit

This guide covers how to extend the smaqit framework by creating new agents, modifying principles, and contributing framework improvements.

## Overview

smaqit's source is structured as a three-level compilation chain:

- **Level 0 (Framework)** — `framework/*.md` — Philosophy and principles (WHY/WHAT)
- **Level 1 (Templates + Compilation Rules)** — `templates/` — Structure and directives
- **Level 2 (Agents)** — `agents/*.agent.md` — Concrete compiled agents shipped to users

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

- **Modify agent behavior** — Edit `agents/` source files or their `templates/agents/compiled/*.rules.md` rules
- **Change framework principles** — Edit `framework/*.md` and propagate changes downward
- **Add a new skill** — Create `skills/<name>/SKILL.md` and add to `installer/skills/`
- **Add a new agent** — Create compilation rules in `templates/agents/compiled/`, write `agents/` source, add to installer

## Compilation Convention

The L0→L1→L2 chain is a convention, not automated tooling. Follow it when making changes:

1. **Principle changes (L0)** — Edit `framework/*.md`. No directives, no file paths. Philosophical only.
2. **Directive changes (L1)** — Update the relevant `templates/agents/compiled/*.rules.md` file.
3. **Agent update (L2)** — Edit `agents/smaqit.<name>.agent.md` to reflect the new directives.
4. **Validate** — Use `/smaqit.qa` to check consistency across levels.

Commit sequentially: L0 change → L1 change → L2 change, each in its own commit with prefix `L0:`, `L1:`, `L2:`.

## Using the QA Agent

`/smaqit.qa` is the framework validation agent. Run it after changes to:

- Detect level contamination (directives in L0, philosophy in L2)
- Verify agent structure matches templates
- Check compilation chain integrity
- Run full framework consistency checks before releasing

## Installer Sync

After editing source files, sync the installer mirror so builds pick up the changes:

```bash
cd installer && make build
```

`make build` copies `agents/`, `skills/`, `templates/specs/`, and `framework/` into `installer/` subdirectories and embeds them into the binary.

## Adding a New Layer Agent

1. Create `templates/agents/compiled/<layer>.rules.md` with L0→L1 compilation rules
2. Write `agents/smaqit.<layer>.agent.md` following the spec agent structure
3. Create `skills/smaqit.input-<layer>/SKILL.md` as the input validation skill
4. Add agent and skill to installer embed paths in `installer/main.go`
5. Run `make build` and validate with `/smaqit.qa`

## Best Practices

### Level Boundaries

- **L0 files** — No directives (MUST/MUST NOT/SHOULD), no file paths, no workflows. Philosophical only.
- **L1 files** — Transform L0 concepts into directives and structure.
- **L2 files** — Self-contained and executable. No references to L0 or L1.

### Validation

Run `/smaqit.qa` after every significant change and before committing.

## Contributing

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for contribution guidelines.

## Further Reading

- [Level Up Compilation Architecture](../designs/level-up-compilation.md) — How the L0→L1→L2 chain works
- [Hierarchical Levels](../designs/hierarchical-levels.md) — Why four levels exist
- [Quickstart](quickstart.md) — Using smaqit in a project
