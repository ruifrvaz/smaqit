# Prompts as Input Records

## Philosophy

Prompts in smaqit are **versioned input records**, not temporary commands.

## What They Capture

Prompts capture user requirements at each layer and serve as:

- **Source of truth** for what was requested (specs are derived outputs)
- **Audit trail** showing requirement evolution over time  
- **Reproducibility anchor** enabling workflow re-execution

## Version Control Integration

Filled prompts should be committed to version control alongside specs. When requirements change, users edit prompt files and regenerate specs—the prompt amendment creates explicit requirement history.

## Benefits

- **Traceability**: Every spec traces to a prompt file
- **Reproducibility**: Identical prompt set produces equivalent validated behavior
- **Auditability**: Requirement changes are explicit in version history
- **Documentation**: Prompts document what was requested at each layer

## Related

- [Free-Style Prompts](../design-decisions/free-style-prompts.md) — Why prompts are natural language
- [Amending Requirements](../workflows/amending-requirements.md) — How to evolve prompts over time
