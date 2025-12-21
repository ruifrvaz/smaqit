# HTML Comment Convention for Examples

## Pattern

Templates and shipped prompts include examples wrapped in HTML comments to guide users without polluting requirements.

## Format

```markdown
### Actors

<!-- Example: "Mario Fan - Users who love Nintendo's Mario franchise" -->
<!-- Example: "System - Console application" -->

[User writes actual actors here]
```

## Critical Rules

**Agents MUST ignore HTML comments** (`<!-- -->`).

Examples are documentation for humans only. This prevents example requirements from contaminating generated specs.

## Implementation

When agents read prompts:
1. Strip all HTML comments before interpretation
2. Parse only user-written content
3. Do not treat example text as requirements

## Benefits

- **User guidance** without rigid enforcement
- **Flexibility** in how users express requirements
- **Clean separation** between examples and actual content
- **No pollution** of generated specs with placeholder data

## Usage

Use HTML comments for:
- Format examples (`<!-- Example: "Actor Name - Description" -->`)
- Suggested content (`<!-- Example: "Scalability: Handle 10k concurrent users" -->`)
- Clarification notes (`<!-- Tip: Be specific and measurable -->`)

## Related

- [Free-Style Prompts](../design-decisions/free-style-prompts.md) — Why prompts are natural language
- [Validation Messages](../patterns/validation-messages.md) — How agents guide users when input is insufficient
