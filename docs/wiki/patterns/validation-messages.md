# Validation Messages

## Principle

When prompts are empty or insufficient, agents guide users naturally, not with template references or error codes.

## Good Examples

✅ "Please specify measurable success criteria for this use case"  
✅ "Could you describe the actors who will interact with this system?"  
✅ "What technologies are you planning to use for the backend?"

These are natural questions that guide users toward what's needed.

## Bad Examples

❌ "Missing: Success Metrics section"  
❌ "ERROR: Incomplete prompt structure"  
❌ "Section 2.3.1 is required but not found"

These reference template structure rather than guiding requirement capture.

## Pattern

**If prompt empty or insufficient:**
1. Agent halts execution
2. Agent suggests what's missing using natural language guidance
3. Agent waits for user to fill prompt and re-invoke

**If prompt filled sufficiently:**
1. Agent proceeds with spec generation
2. Agent uses prompt content as authoritative input

## Why Natural Language?

- Users think in questions, not form fields
- Reduces friction in requirement capture
- Makes agents feel collaborative, not bureaucratic
- Aligns with free-style prompt philosophy

## Implementation

Agents should:
- Ask clarifying questions ("What problem are you solving?")
- Suggest categories ("Consider describing actors, goals, and constraints")
- Avoid referencing template structure ("Section 2.1 is missing")
- Use conversational tone ("Could you tell me more about...?")

## Related

- [Free-Style Prompts](../design-decisions/free-style-prompts.md) — Why prompts are natural language
- [HTML Comment Convention](html-comment-convention.md) — How templates provide guidance
