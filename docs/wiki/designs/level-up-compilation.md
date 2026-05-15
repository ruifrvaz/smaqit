# Level Up Compilation Architecture

**Status:** Implemented (PR #36)  
**Date:** 2026-01-14  
**Category:** Meta-Framework Design

## Overview

The **Level Up** architecture is smaqit's internal compilation process for building the framework itself. It transforms abstract principles (L0) into executable templates (L1) and finally into concrete agents (L2). This is meta-framework work—not used by shipped agents, but essential for maintaining framework consistency and enabling future extensibility.

**Key Insight:** Level Up is analogous to compiler architecture (C# → IL → Machine Code), where each level serves as an intermediate representation that enables transformation and optimization.

## Three Levels

### Level 0: Framework Principles (Philosophy)

**Location:** `framework/*.md`  
**Audience:** Humans understanding smaqit, agents consuming foundational concepts  
**Content:** WHY and WHAT (conceptual)  
**Purity:** NO directives, NO implementation details, NO procedural workflows

**Purpose:** Human-readable philosophy that explains the reasoning behind smaqit's design.

**Current files:**
- `framework/SMAQIT.md` — Core principles
- `framework/LAYERS.md` — Layer philosophy
- `framework/PHASES.md` — Phase philosophy
- `framework/TEMPLATES.md` — Template philosophy
- `framework/AGENTS.md` — Agent philosophy
- `framework/ARTIFACTS.md` — Artifact philosophy

**Example principle:**
```markdown
The Test Independence Principle

Test artifacts exist independently of agent execution. Tests can run in any 
environment with the appropriate runtime, enabling continuous integration, 
local developer workflows, and automated verification outside the validation phase.
```

### Level 1: Template Directives (Transformation Rules)

**Location:** `templates/**/*.md`  
**Audience:** Template consumers (Agent-L2 compiler, installer)  
**Content:** HOW (operational)  
**Purity:** Directives with placeholders, NO concrete values

**Purpose:** Intermediate representation that compiles L0 principles into actionable directives.

**Structure:**
- **Templates** (`templates/agents/*.template.md`) — Generic structure with placeholders
- **Compilation Files** (`templates/agents/compiled/*.rules.md`) — L0→L1 transformation rules

**Key Innovation: Compilation Files Architecture**

Like intermediate languages in traditional compilers (IL in .NET, LLVM IR), L1 serves as an optimizable transformation layer rather than simply being a templated version of the final output. smaqit achieves this through **compilation files** that document how L0 principles transform into L1 directives while keeping templates generic.

**Compilation file structure:**
1. **Frontmatter** — Metadata (layer/phase, target, sources, created)
2. **Source L0 Principles** — Tabulated references (Source File | Section)
3. **L1 Directive Compilation** — Philosophy → directives transformation
4. **Compilation Guidance for Agent-L2** — Step-by-step merge instructions

**Example (validate.rules.md):**
```markdown
---
phase: validate
target: agents/smaqit.validation.agent.md
sources:
  - framework/PHASES.md
  - framework/ARTIFACTS.md
created: 2026-01-14
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| ARTIFACTS.md | The Test Independence Principle |
| PHASES.md | Validate Phase Activities |
| ARTIFACTS.md | Implementation Artifacts by Phase |

## L1 Directive Compilation

**MUST directives:**
- Generate executable test artifacts from Coverage specifications
- Create test files in `tests/` directory implementing Coverage spec test cases
- Use test framework specified in Stack spec
- Organize tests by feature with clear mapping to Coverage spec scenarios
- Preserve Given/When/Then structure from Gherkin scenarios in test code
- Generate test framework configuration file
- Ensure test artifacts are executable independently (outside agent context)
```

### Level 2: Concrete Agents (Shipped Artifacts)

**Location:** `agents/*.agent.md`  
**Audience:** LLMs executing workflows in user projects  
**Content:** Concrete directives for specific layers/phases  
**Purity:** NO placeholders, self-contained and executable

**Purpose:** Final compiled agents shipped to user projects via installer.

**Compilation process:**
```
L1 Template + L1 Compilation File → Agent-L2 compiler → L2 Agent

Result: Self-contained agent with:
- All placeholders replaced (e.g., [PHASE] → validation)
- Phase-specific directives merged from compilation file
- No references to L0 or L1 sources
```

**Example transformation:**
```
L1 Template: "MUST read requirements from session context"
L1 Compilation File: validate.rules.md with Test Independence directives
    ↓ [Agent-L2 compiles]
L2 Agent (smaqit.validation.agent.md):
    "MUST read requirements from session context"
    "MUST generate executable test artifacts in tests/ directory"
    "MUST create test framework configuration file"
```

## Information Flow

```
┌─────────────────────────────────────────────────────┐
│ L0: framework/ARTIFACTS.md                          │
│ "Test Independence Principle: Tests exist           │
│  independently of agent execution..."                │
└─────────────────┬───────────────────────────────────┘
                  │ Agent-L1 compiles
                  ↓
┌─────────────────────────────────────────────────────┐
│ L1: templates/agents/compiled/validate.rules.md     │
│ "MUST generate executable test artifacts"           │
│ "MUST create test framework configuration"          │
│ "MUST organize tests with Coverage spec mapping"    │
└─────────────────┬───────────────────────────────────┘
                  │ Agent-L2 merges with template
                  ↓
┌─────────────────────────────────────────────────────┐
│ L2: agents/smaqit.validation.agent.md                │
│ Complete Validation Agent with:                      │
│ - Test Independence directives (from L1)             │
│ - Phase-specific rules (from L1)                     │
│ - No placeholders (ready for execution)             │
└─────────────────────────────────────────────────────┘
```

## Meta-Agents (Internal Development Tools)

smaqit uses three internal agents for framework development:

### Agent-L0: Principle Curator
**Responsibility:** Maintain L0 purity  
**Scope:** `framework/*.md` files  
**Role:** Remove directive contamination, ensure principles stay philosophical

**Example cleanup:**
```
BEFORE (directive contamination):
"Agents MUST NOT duplicate information from existing specs"

AFTER (pure principle):
"Single Source of Truth: Each piece of information exists in exactly one place"
```

### Agent-L1: Template Compiler
**Responsibility:** Compile L0 principles into L1 directives and compilation files  
**Scope:** `templates/**/*.template.md` and `templates/agents/compiled/*.rules.md`  
**Role:** Transform philosophy into actionable directives

**Compilation rules:**
- **Principles → Directives:** Abstract concepts become MUST/SHOULD/MUST NOT rules
- **Philosophy → Workflows:** Conceptual flows become step-by-step sequences
- **Rationale removal:** "Why" explanations stay at L0, only "what/how" compiles to L1
- **Examples → Placeholders:** Specific examples become generic `[PLACEHOLDER]` patterns
- **Template purity:** Templates maintain generic structure, specifics in compilation files

### Agent-L2: Agent Compiler
**Responsibility:** Compile L1 templates + compilation files into L2 agents  
**Scope:** `agents/*.agent.md` files  
**Role:** Generate concrete agents by merging templates with compilation rules

**Compilation process:**
1. Read L1 template structure
2. Read L1 compilation file for specific phase/layer
3. Merge directives from compilation file into template
4. Replace all placeholders with concrete values
5. Validate result is self-contained and executable

## Why Compilation Files?

### The Problem: "L2 with Placeholders" Anti-Pattern

Initial approach attempted to put phase-specific directives directly in L1 templates:

```markdown
<!-- L1 Template (WRONG) -->
## Phase-Specific Rules

### Validation Phase
- MUST generate executable test artifacts
- MUST create test framework configuration
... (32 lines of validation-specific directives)

### Development Phase
- MUST generate application code
... (different phase-specific content)
```

**Issue:** If L1 templates contain final directives (just with placeholders), then L1→L2 is merely string substitution, not transformation. This defeats L1's purpose as an intermediate representation.

### The Solution: Compilation Files Architecture

Separate transformation rules from template structure:

```markdown
<!-- L1 Template (CORRECT) -->
## Phase-Specific Rules

[PHASE_SPECIFIC_RULES]

<!-- L1 Transformation Instructions:
     Agent-L2 compiles [PHASE_SPECIFIC_RULES] by reading:
     templates/agents/compiled/[phase].rules.md § Phase-Specific Rules
-->
```

**Benefit:** L1 template remains generic, compilation files document L0→L1 transformations, L2 compilation is true transformation (not just substitution).

## Compiler Analogy

The Level Up architecture parallels traditional compiler design:

| Level | Traditional Compiler | smaqit Level Up |
|-------|---------------------|----------------|
| Source | C# (human-readable) | L0 Principles (philosophy) |
| Intermediate | IL/MSIL (optimizable) | L1 Templates + Compilation Files |
| Target | Machine Code (executable) | L2 Agents (concrete) |

**Key insight:** Just as IL enables optimization and platform-independence, L1 compilation files enable principle-to-directive transformation while preserving traceability.

## Version Control Conventions

Level Up changes follow strict commit ordering:

```bash
# Correct sequence (PR #36 example)
git commit -m "L0: Add Test Independence Principle to ARTIFACTS.md"
git commit -m "L1: Create validate.rules.md with Test Independence directives"
git commit -m "L2: Apply validate.rules.md to smaqit.validation.agent.md"
git commit -m "docs: Document compilation files architecture"
```

**Rules:**
- **Sequential commits:** L0 → L1 → L2 → docs
- **Level isolation:** Never mix levels in a single commit
- **Commit prefixes:** Use `L0:`, `L1:`, `L2:`, `docs:` for traceability
- **Rationale:** Preserves compilation chain, enables bisection, documents transformation

## Current State (2026-01-19)

### Completed
- ✅ L0 purity achieved for ARTIFACTS.md and PHASES.md (PR #36)
- ✅ Compilation files architecture established
- ✅ Eight compilation files created:
  - 3 phase files: validate.rules.md, develop.rules.md, deploy.rules.md
  - 5 layer files: business.rules.md, functional.rules.md, stack.rules.md, infrastructure.rules.md, coverage.rules.md
- ✅ implementation-agent.template.md updated to reference compilation files
- ✅ specification-agent.template.md updated to reference compilation files
- ✅ Frontmatter structure standardized (layer/phase, target, sources, created)
- ✅ Source L0 Principles converted to tabulated format

### Pending
- ⏳ Agent-L2 compilation execution (apply rules to generate L2 agents)
- ⏳ Automated compilation tooling

## Not Shipped to Users

**Critical:** The Level Up compilation process is **internal development work**. User projects receive:

- ✅ **Compiled L2 agents** (`agents/*.agent.md`) — Final products
- ✅ **Spec templates** (`templates/specs/*.template.md`) — For generating specs
- ❌ **Framework files** (`framework/*.md`) — NOT copied by installer
- ❌ **Compilation files** (`templates/agents/compiled/*.rules.md`) — NOT copied by installer
- ❌ **Meta-agents** (Agent-L0, Agent-L1, Agent-L2) — Internal tools only

**Rationale:** Shipped agents don't need to know how they were compiled. They're self-contained executables.

## Future: Extensibility

Once the Level Up pipeline is complete and proven, it enables extensibility:

1. **Custom Layers** — Define new layer principles at L0, compile to L1 templates, compile to L2 agents
2. **Custom Phases** — Define new phase workflows at L0, compile to L1 templates, compile to L2 agents
3. **Domain-Specific Frameworks** — Healthcare, finance, e-commerce variants all compile through same pipeline

The compilation architecture ensures consistency: all domains follow the same L0→L1→L2 transformation pattern.

## References

- **Task B001:** Extensible Meta-Framework (long-term extensibility vision)
- **Task B002:** Iterating Extensible Framework (immediate compilation execution)
- **Session 037:** Compilation Files Architecture (2026-01-14 design session)
- **Copilot Instructions:** `.github/copilot-instructions.md` § Level Agents
