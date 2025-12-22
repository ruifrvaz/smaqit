# Explicit Over Implicit

## Overview

smaqit deliberately favors explicit documentation over implicit understanding. This design decision reduces ambiguity in LLM agent execution and improves human comprehension.

## The Principle

**When in doubt, make it explicit:**
- State assumptions rather than assume shared context
- Define scope boundaries rather than imply them
- Reference sources rather than expect inference

## Why Explicitness Matters

### For LLM Agents

LLMs excel at pattern matching but struggle with implicit context:
- They don't retain state between invocations
- They can't infer organizational knowledge
- They benefit from explicit constraints

**Example:** Instead of "implement the authentication feature," specify:
```
Implement JWT-based authentication:
- Token expires after 24 hours
- Refresh tokens valid for 30 days
- Store tokens in HTTP-only cookies
```

### For Human Developers

Explicit documentation reduces onboarding time and prevents misinterpretation:
- New team members don't need tribal knowledge
- Decisions are traceable to their rationale
- Changes are obvious (explicit text changed, not implicit meaning)

## Manifestations in smaqit

### 1. Requirement Identifiers

Every acceptance criterion has a unique ID:
```
BUS-LOGIN-001: User can authenticate with valid credentials
```

Not just "user authentication works"—explicit, traceable, unambiguous.

### 2. Reference Sections

Specs explicitly state their sources:
```markdown
## References

### Implements
- [BUS-LOGIN](../business/login.md) — Implements login use case
```

Not implicit "this relates to business somehow"—explicit relationship documented.

### 3. Scope Boundaries

Every spec defines what's included AND what's excluded:
```markdown
## Scope

### Included
- Username/password authentication
- JWT token generation

### Excluded
- OAuth integration (deferred to Phase 2)
- Multi-factor authentication
```

Not just "authentication"—explicit boundaries prevent scope creep.

### 4. Template Constraints

Templates are mandatory structure, not suggestions:
- Agents MUST follow templates exactly
- No optional sections—everything is explicit
- Placeholder format is standardized: `[PLACEHOLDER]`

## Trade-offs

**Benefits:**
- Reduced ambiguity in agent execution
- Lower cognitive load for humans
- Easier to detect inconsistencies
- Simpler impact analysis (explicit references)

**Costs:**
- More verbose documentation
- Takes longer to write specs
- Can feel bureaucratic for small projects

The cost is intentional—explicit documentation is an investment that pays off in reduced miscommunication and faster debugging.

## When to Bend the Rule

Explicitness has diminishing returns. Avoid:
- Explaining obvious implementation details
- Documenting every micro-decision
- Over-specifying flexible areas

**Good:** "API returns 401 for invalid credentials"  
**Bad:** "API returns HTTP status code 401 Unauthorized, which is defined in RFC 7235 Section 3.1, with..."

Use judgment. Be explicit where ambiguity would cause problems, concise elsewhere.

## Related

- [Traceability](../concepts/traceability.md) — How explicit references enable impact analysis
- [Template Constraints](template-constraints.md) — Why templates enforce structure
