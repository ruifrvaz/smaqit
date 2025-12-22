# Template Constraints

## Overview

smaqit templates are mandatory structures that agents MUST follow when producing output. This design decision reduces LLM variance and ensures predictable consumption by downstream agents and humans.

## The Principle

**Templates are cognitive scaffolds, not suggestions.**

Templates enforce:
- **Consistent output** across runs
- **Predictable input** for downstream consumers
- **Reduced LLM variance** through structural constraints

## Why Templates Must Be Mandatory

### The Problem with Optional Structure

When templates are suggestions:

1. LLM generates output in its preferred style
2. Style varies between invocations (non-deterministic)
3. Downstream consumers can't rely on structure
4. Manual normalization required
5. Human review time increases

**Cost:** Unpredictable output + manual cleanup + downstream brittleness

### The Mandatory Alternative

When templates are mandatory:

1. LLM generates output following exact template
2. Structure is consistent across invocations
3. Downstream consumers parse reliably
4. Automated processing possible
5. Human review focuses on content, not structure

**Benefit:** Predictable output + automated consumption + reduced variance

## Template Types in smaqit

### 1. Specification Templates

Define the structure for spec documents:

```
templates/specs/
├── business.template.md
├── functional.template.md
├── stack.template.md
├── infrastructure.template.md
└── coverage.template.md
```

**Every spec MUST:**
- Use its layer template from `templates/specs/[LAYER].template.md`
- Include all required sections
- Replace all placeholders with actual content
- Not add sections not defined in template
- Not omit required sections

### 2. Agent Templates

Define the structure for agent definitions:

```
templates/agents/
├── specification-agent.template.md
└── implementation-agent.template.md
```

**Every agent MUST:**
- Follow GitHub Custom Agent format (YAML frontmatter + markdown)
- Include all required sections (Role, Input, Output, Directives, Completion Criteria)
- Use standardized placeholder format: `[PLACEHOLDER]`

### 3. Prompt Templates

Define the structure for prompt files:

```
templates/prompts/
├── specification-prompt.template.md
└── phase-prompt.template.md
```

**Prompts are free-style with suggested structure**—templates provide guidance without rigidity. This is the exception to mandatory templates because user input should be natural.

## Placeholder Conventions

All placeholders use `[PLACEHOLDER]` format:
- Brackets: `[` and `]`
- SCREAMING_CASE: `LAYER`, `CONCEPT`, `NNN`
- No spaces inside brackets: `[LAYER]` not `[ LAYER ]`

### Common Placeholders

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `[LAYER]` | Lowercase layer name | `business` |
| `[LAYER_NAME]` | Title case layer name | `Business` |
| `[LAYER_PREFIX]` | 3-letter layer code | `BUS` |
| `[CONCEPT]` | Concept name | `LOGIN` |
| `[NNN]` | Sequential 3-digit number | `001` |

Standardized placeholders enable:
- Easy search/replace during instantiation
- Clear identification of incomplete content
- Automated validation of template compliance

## Enforcement Mechanisms

### Agent Self-Validation

Agents validate their output against templates before completion:

```markdown
**Completion Criteria:**
- [ ] All template sections are filled (no placeholders remain)
- [ ] Output follows designated template structure
- [ ] No sections added beyond template definition
- [ ] No required sections omitted
```

Agents check themselves—quality is built in, not inspected later.

### Compliance Rules

From TEMPLATES.md:

> **Compliance Rules**
> - Agents MUST use the template
> - Agents MUST NOT add sections not defined in the template
> - Agents MUST NOT omit required sections from the template
> - Agents MUST NOT leave placeholder text in completed specs

These are constraints, not guidelines.

## Trade-offs

**Benefits:**
- Predictable output structure
- Reduced LLM variance
- Automated downstream processing
- Faster human review
- Clear quality criteria

**Costs:**
- Less flexibility in output format
- Templates require maintenance
- Can feel rigid for simple cases
- Creativity constrained

The cost is intentional—predictability is more valuable than flexibility for specification documents. We accept reduced creativity in structure to gain reliability in consumption.

## When Templates Are Too Rigid

Templates work well for:
- Specification documents (structured contracts)
- Agent definitions (standard interfaces)
- Framework files (consistent reference)

Templates work poorly for:
- User input (prompts are free-style)
- Exploratory artifacts (research, spikes)
- One-off scripts (not reused)

Use judgment. Apply template constraints where consumption requires predictability, relax elsewhere.

## Evolution Strategy

Templates evolve through:

1. **Identify gap**: Template doesn't capture needed information
2. **Propose amendment**: Update template with new section
3. **Update all instances**: Regenerate affected specs/agents
4. **Document rationale**: Why this section was added

Template changes are expensive (require regeneration), so changes should be deliberate and justified.

## Related

- [Accept Mutability, Validate Behavior](../concepts/accept-mutability.md) — Why artifacts can vary despite templates
- [Self-Validating Agents](../concepts/self-validating-agents.md) — How agents enforce compliance
