# Fail-Fast on Ambiguity

## Overview

smaqit agents are designed to stop and request clarification when input is unclear, rather than proceeding with assumptions. This design decision prevents costly rework from incorrect interpretations.

## The Principle

**When input is unclear:**
- Stop and request clarification
- Do not invent requirements
- Flag assumptions explicitly

The cost of clarification is lower than the cost of rework from incorrect assumptions.

## Why Fail-Fast

### The Problem with "Best Effort"

LLMs can generate plausible-sounding content for almost any prompt. This creates a dangerous pattern:

1. User provides vague requirement: "build a user system"
2. LLM invents details: assumes username/password, assumes email verification, assumes...
3. Implementation proceeds with invented details
4. User sees result: "That's not what I meant!"
5. Rework required

**Cost:** Wasted agent time + wasted compute + user frustration

### The Fail-Fast Alternative

1. User provides vague requirement: "build a user system"
2. Agent stops: "Please clarify: What authentication methods? What user attributes? What registration flow?"
3. User provides details
4. Implementation proceeds with correct requirements
5. User sees result: "Exactly what I needed!"

**Benefit:** Faster time-to-correct-solution + lower cost + higher user confidence

## Manifestations in smaqit

### 1. Pre-Run Validation

Phase prompts check for content before starting:

```markdown
**Pre-Run Validation:**
If `.github/prompts/smaqit.business.prompt.md` is empty or insufficient:
- Halt with natural language guidance
- Example: "Please specify your use cases, actors, and success criteria"
```

Agents don't proceed with blank input. They fail fast with guidance.

### 2. Sufficiency Checks

Agents evaluate whether prompt content is sufficient:

```markdown
**Validate sufficiency:** Agents MUST request clarification if prompt content 
is insufficient, using natural language guidance.
```

Not "make do with what you have"—stop and ask for more.

### 3. Conflict Detection

Implementation agents validate cross-layer coherence:

```markdown
**Cross-Layer Consolidation:**
1. Coherence check: Verify specs across layers are compatible
2. Conflict detection: Identify contradictions between layers
3. Gap analysis: Ensure all upstream requirements have downstream specs
4. Amendment request: If conflicts or gaps exist, request amendments
```

Don't silently choose one layer over another. Stop and report the conflict.

### 4. Explicit Assumption Flagging

When agents must proceed with uncertainty:

```markdown
**Assumption:** User data is stored in relational database (not specified)
**Flag:** Confirm database type before implementation
```

Make assumptions visible so users can correct them.

## Trade-offs

**Benefits:**
- Prevents wasted work from misinterpretation
- Forces users to think through requirements
- Makes uncertainty visible early
- Builds user confidence in agent behavior

**Costs:**
- More back-and-forth with users
- Slower initial progress (but faster overall)
- Requires users to be more specific upfront

The cost is intentional—it's cheaper to clarify early than to rebuild later.

## Common Clarification Patterns

**Vague scope:**
```
User: "Build an authentication system"
Agent: "Please clarify: username/password, OAuth, both? Email verification required? Password reset flow needed?"
```

**Contradictory requirements:**
```
User: "Make it fast and thorough" (for tests)
Agent: "Trade-off detected: Fast tests typically skip edge cases. Thorough tests take longer. Which priority: speed or coverage?"
```

**Missing constraints:**
```
User: "Deploy to cloud"
Agent: "Please specify: Which cloud provider? Budget constraints? Geographic requirements? Scaling needs?"
```

## When to Bend the Rule

Don't fail-fast on:
- Minor stylistic details (agent can choose)
- Standard industry practices (agent can assume best practices)
- Easily reversible decisions (agent can proceed with default)

**Good:** Stop for "what authentication methods?"  
**Bad:** Stop for "what hex color for error messages?"

Use judgment. Fail-fast on decisions that affect requirements or architecture, proceed with reasonable defaults elsewhere.

## Related

- [Validation Messages](../patterns/validation-messages.md) — How agents communicate clarification needs
- [Free-Style Prompts](free-style-prompts.md) — Why prompts are natural language (reduces friction)
