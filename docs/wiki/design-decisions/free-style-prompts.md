# Free-Style Prompts

## Decision

Prompts are **natural language inputs**, not rigidly structured forms.

## Rationale

### Why Free-Style?

1. **Reduces friction in requirement capture**  
   Users express requirements naturally without fighting form fields

2. **Accommodates diverse project types and domains**  
   No rigid structure works for all domains—let users adapt

3. **Allows users to express requirements naturally**  
   People think in paragraphs, not bulleted lists

4. **Agents interpret natural language**  
   LLMs excel at this—leverage their strength

## Implementation

Templates provide **suggested structure** (sections, sub-sections) but users write requirements in their own words.

Example structure:
```markdown
## Requirements

### Actors
[Describe who interacts with the system]

### Use Cases
[Describe what users want to accomplish]

### Success Metrics
[Describe measurable outcomes]
```

Users fill in natural language. Agents interpret and request clarification if needed.

## Trade-offs

**Benefits:**
- Lower barrier to entry
- More expressive requirement capture
- Better domain fit for varied projects

**Costs:**
- Agents must validate completeness (Fail-Fast on Ambiguity)
- More variability in input format
- Potential for incomplete requirements

Until proven wrong, we accept these costs because the flexibility benefits outweigh them.

## Related

- [Prompts as Input Records](../concepts/prompts-as-input-records.md) — What prompts capture
- [Validation Messages](../patterns/validation-messages.md) — How agents guide users
