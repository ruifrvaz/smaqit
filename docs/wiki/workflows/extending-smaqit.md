# Extending smaqit via smaqit-sdk

This guide covers how to extend the smaqit framework by creating new agents, modifying principles, and contributing framework improvements using **smaqit-sdk**.

## Overview

**smaqit-sdk** is the framework development toolkit that contains:

- **Level agents** (L0, L1, L2) for principle documentation, template compilation, and agent compilation
- **QA agent** for framework validation and testing
- **Framework files** (7 files in `framework/`) defining smaqit architecture
- **Templates** (14 total) for specs, prompts, and agents
- **new-agent prompt** for creating custom agents

## When to Use smaqit-sdk

Use **smaqit-sdk** instead of **smaqit** when:

- Creating custom layer or phase agents for specialized domains
- Modifying framework principles or compilation rules
- Contributing to smaqit framework development
- Building organization-specific agent extensions
- Debugging or improving agent templates

Use **smaqit** (product) when building applications with existing agents.

## Installation

```bash
# Install SDK
curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit/main/install-sdk.sh | bash

# Initialize SDK project
smaqit-sdk init
```

## SDK Project Structure

After `smaqit-sdk init`, you'll have:

```
.smaqit/
├── framework/                    # 7 framework principle files
│   ├── SMAQIT.md                # Core principles index
│   ├── LAYERS.md                # Layer definitions
│   ├── PHASES.md                # Phase workflows
│   ├── TEMPLATES.md             # Template structure rules
│   ├── AGENTS.md                # Agent behaviors
│   ├── ARTIFACTS.md             # Artifact rules
│   └── PROMPTS.md               # Prompt architecture
├── templates/
│   ├── specs/                   # 5 spec templates
│   │   ├── business.template.md
│   │   ├── functional.template.md
│   │   ├── stack.template.md
│   │   ├── infrastructure.template.md
│   │   └── coverage.template.md
│   ├── prompts/                 # 2 prompt templates
│   │   ├── layer.template.md
│   │   └── phase.template.md
│   └── agents/                  # 7 agent templates
│       ├── layer.template.md
│       ├── phase.template.md
│       └── compiled/            # L0→L1 compilation rules
│           ├── validate.rules.md
│           ├── develop.rules.md
│           ├── deploy.rules.md
│           ├── business.rules.md
│           ├── functional.rules.md
│           ├── stack.rules.md
│           └── infrastructure.rules.md
└── logs/                        # Session and task logs
.github/
├── agents/                      # 4 meta agents
│   ├── smaqit.L0.agent.md      # Principle documentation
│   ├── smaqit.L1.agent.md      # Template compilation
│   ├── smaqit.L2.agent.md      # Agent compilation
│   └── smaqit.qa.agent.md      # Framework testing
└── prompts/
    └── smaqit.new-agent.prompt.md  # Agent creation workflow
```

## Level Agent Architecture

smaqit uses a **three-level compilation chain**:

### Level 0 (L0): Principles

**Purpose:** Define framework philosophy and concepts without implementation details.

**Agent:** `/smaqit.L0`

**Content types:**
- WHY: Philosophical foundations
- WHAT: Definitions and categorizations
- HOW arranged: Structural organization

**Example:** `framework/SMAQIT.md` defines "Agents validate their own output" as a concept.

**When to use Agent-L0:**
- Documenting new framework principles
- Clarifying architectural rationale
- Defining new concepts or mappings
- Updating framework files (`framework/*.md`)

### Level 1 (L1): Templates

**Purpose:** Compile L0 principles into directive-based templates with structure.

**Agent:** `/smaqit.L1`

**Compilation mechanism:**
- Generic templates (`templates/agents/*.template.md`) with placeholders
- Compilation files (`templates/agents/compiled/*.rules.md`) with L0→L1 transformation rules
- Output: Templates with directives (MUST/MUST NOT/SHOULD)

**Example:** Compilation file transforms L0 "Agents validate output" into L1 "Agents MUST validate output before declaring completion".

**When to use Agent-L1:**
- Creating or updating spec templates (`templates/specs/`)
- Creating or updating prompt templates (`templates/prompts/`)
- Creating or updating agent templates (`templates/agents/*.template.md`)
- Creating or updating compilation files (`templates/agents/compiled/*.rules.md`)
- Compiling L0 principles into structured directives

### Level 2 (L2): Product Agents

**Purpose:** Compile L1 templates into concrete product agents.

**Agent:** `/smaqit.L2`

**Compilation mechanism:**
- Reads agent template + compilation file
- Merges according to compilation guidance section
- Produces concrete agent (e.g., `agents/smaqit.business.agent.md`)

**Example:** Merges `layer.template.md` + `business.rules.md` → `smaqit.business.agent.md`.

**When to use Agent-L2:**
- Compiling product agents from templates (`agents/*.agent.md`)
- Regenerating agents after template changes
- Creating custom agents via compilation

## Creating a New Agent

### Using the new-agent prompt

1. **Define requirements:**

Fill `.github/prompts/smaqit.new-agent.prompt.md` with:
- Agent name and purpose
- Layer or phase it operates on
- Input requirements and output artifacts
- Validation criteria

2. **Compile L1 template:**

```bash
# Open GitHub Copilot chat
/smaqit.L1
```

Agent-L1 will:
- Create or update agent template in `templates/agents/`
- Create or update compilation file in `templates/agents/compiled/`
- Document L0 principles that informed the template

3. **Compile L2 product agent:**

```bash
/smaqit.L2
```

Agent-L2 will:
- Read template and compilation file
- Merge into concrete product agent in `agents/`
- Validate output structure

4. **Test with QA agent:**

```bash
/smaqit.qa
```

QA agent will:
- Validate agent structure
- Check for level contamination
- Verify compilation chain integrity

## Modifying Existing Agents

### Updating principles (L0)

To change framework philosophy or concepts:

```bash
/smaqit.L0
```

Provide context about which principle to modify and why. Agent-L0 will:
- Update relevant framework file (`framework/*.md`)
- Preserve principle purity (no directives or implementation details)
- Document rationale

### Updating templates (L1)

To change agent structure or directives:

```bash
/smaqit.L1
```

Provide context about which template or compilation file to modify. Agent-L1 will:
- Update agent template or compilation file
- Compile new directives from L0 principles
- Maintain compilation chain integrity

### Recompiling agents (L2)

After L0 or L1 changes, recompile product agents:

```bash
/smaqit.L2
```

Agent-L2 will:
- Detect which agents need recompilation
- Merge updated templates and compilation files
- Validate output

## Release Choreography

smaqit uses **dual versioning** with SDK releases driving product releases:

### SDK Release (sdk-v1.x.x)

1. **Make framework changes:**
   - Update principles (L0), templates (L1), or compilation files (L1)
   - Use Level agents to maintain compilation chain integrity

2. **Validate with QA:**
   ```bash
   /smaqit.qa
   ```

3. **Tag SDK release:**
   ```bash
   git tag sdk-v1.1.0
   git push origin sdk-v1.1.0
   ```

4. **GitHub Actions:**
   - Builds `smaqit-sdk` binaries for all platforms
   - Extracts SDK section from CHANGELOG
   - Creates GitHub release with binaries

### Product Release (v0.x.x)

After SDK release, compile new product agents:

1. **Recompile agents:**
   ```bash
   /smaqit.L2
   ```

2. **Validate with QA:**
   ```bash
   /smaqit.qa
   ```

3. **Tag product release:**
   ```bash
   git tag v0.8.0
   git push origin v0.8.0
   ```

4. **GitHub Actions:**
   - Builds `smaqit` binaries for all platforms
   - Extracts Product section from CHANGELOG
   - Creates GitHub release with binaries

### Version Stability

- **SDK (sdk-v1.x.x):** Can reach v1.0.0+ (stable framework principles)
- **Product (v0.x.x):** May remain <1.0.0 longer (evolving product features)

Independent versioning allows SDK stability while product continues rapid iteration.

## Best Practices

### Level Boundaries

- **L0 files** should contain NO directives (MUST/MUST NOT/SHOULD), NO file paths, NO implementation details
- **L1 files** should transform L0 concepts into directives, structure, and mappings
- **L2 files** should be pure compilation outputs, not manually edited

### Opportunistic Cleanup

When working in contaminated areas (mixed levels), actively extract and relocate content:

1. **Don't introduce new contamination** — Respect level boundaries for new content
2. **Clean contamination within session scope** — If you spot it, fix it
3. **Document cleanup** — Note what was cleaned and where it moved
4. **Prioritize session goals** — Don't derail work, but seize opportunities

### Validation

Run `/smaqit.qa` frequently:
- After L0 principle changes
- After L1 template compilation
- After L2 agent compilation
- Before committing changes
- Before releases

### Documentation

Document decisions in:
- **Wiki** (`docs/wiki/`) — Human-readable context and rationale
- **Session history** (`docs/history/`) — Detailed implementation logs
- **CHANGELOG** — User-facing changes (separate SDK and Product sections)

## Troubleshooting

### Level contamination detected

**Symptom:** QA agent reports L0 files contain directives or L1 files contain philosophy.

**Solution:**
1. Invoke appropriate Level agent (L0, L1, or L2)
2. Extract contaminated content
3. Relocate to proper level
4. Update compilation chain if needed

### Agent compilation fails

**Symptom:** Agent-L2 can't merge template and compilation file.

**Solution:**
1. Check template structure matches compilation file expectations
2. Verify compilation file has proper sections (L0 Principles, L1 Directives, Compilation Guidance)
3. Run `/smaqit.qa` to identify specific issues
4. Use `/smaqit.L1` to fix template or compilation file

### Framework changes don't propagate

**Symptom:** L0 principle changes don't appear in product agents.

**Solution:**
1. Compile L1 templates first: `/smaqit.L1`
2. Update compilation files if needed
3. Recompile L2 agents: `/smaqit.L2`
4. Validate with `/smaqit.qa`

## Contributing

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for contribution guidelines.

When contributing framework improvements:

1. **Start with principles** — Use `/smaqit.L0` to document WHY
2. **Compile to templates** — Use `/smaqit.L1` to create structure
3. **Generate product agents** — Use `/smaqit.L2` to produce concrete agents
4. **Validate everything** — Use `/smaqit.qa` before submitting PR
5. **Document decisions** — Update wiki with rationale and examples

## Further Reading

- [Team Alignment](../concepts/team-alignment.md) — How layers map to Agile roles
- [Quickstart](quickstart.md) — Building apps with smaqit product
- [Wiki Index](../README.md) — All documentation
