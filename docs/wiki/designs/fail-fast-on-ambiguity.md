# Fail-Fast on Ambiguity

## Overview

smaqit agents are designed to stop and request clarification when input is unclear, rather than proceeding with assumptions. This design decision prevents costly rework from incorrect interpretations.

## The Principle

**When input is unclear:**
- Stop and request clarification
- Do not invent requirements
- Never fill a required section with an assumption

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

Agents check session context for content before starting:

```markdown
**Pre-Run Validation:**
If session context is empty or insufficient:
- Halt with natural language guidance
- Example: "Please specify your use cases, actors, and success criteria"
```

Agents don't proceed with blank input. They fail fast with guidance.

### 2. Sufficiency Checks

Agents evaluate whether session context is sufficient:

```markdown
**Validate sufficiency:** Agents MUST request clarification if session context 
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

### 4. Requirements Come From the User

A gate counts only what the user actually stated: their request (passed verbatim as `user_request`), their answers to an earlier question, their own messages, and files they pointed at, such as `.smaqit/requirements-extract.md`. Three things look like requirements and are not:

- **A name.** A blank project titled "mario-app" states no actor, no use case, and no success metric. An LLM will happily turn "Mario" into a fan site with a greeting and a button, and that is exactly the opinionated cycle this rule exists to prevent.
- **The agent's own knowledge.** Knowing what a typical checkout flow looks like is not knowing what this user's checkout must do.
- **A caller's retelling.** When a main conversation or orchestrator hands the request to a subagent, its summary can quietly add detail. The verbatim `user_request` is what counts.

How strictly upstream specifications count depends on the mode:

| Mode | Upstream specifications |
|---|---|
| `assisted` (the default when the user starts a phase or invokes an orchestrating skill) | Context for coherence only. Every required section traces to the user's own statements, so each layer may ask for what the user hasn't said yet |
| `autonomous` (the default when the model starts a phase or chooses an orchestrating skill) | Count as requirement sources, since they passed their own gate. The run asks only where nothing upstream covers a section |

### 5. No Assumption-Flagging for Required Sections

Earlier versions let an agent proceed with a flagged assumption when it couldn't reach the user. That escape hatch is gone: every agent can now return `NEEDS-INPUT`, so clarification is always available, and a flagged assumption was in practice an invented requirement with a label. Optional sections may still use a default, as long as the specification states it as a default.

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

Don't fail-fast on (outside required sections):
- Minor stylistic details (agent can choose)
- Standard industry practices (agent can assume best practices, and states the choice)
- Easily reversible decisions (agent can proceed with a stated default)

**Good:** Stop for "what authentication methods?"  
**Bad:** Stop for "what hex color for error messages?"

Use judgment. Fail-fast on decisions that affect requirements or architecture, proceed with reasonable defaults elsewhere.

## Related

- [Validation Messages](../patterns/validation-messages.md) — How agents communicate clarification needs
