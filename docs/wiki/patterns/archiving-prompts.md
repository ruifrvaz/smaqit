# Archiving Prompts

## Philosophy

smaqit **suggests** but does not enforce prompt archiving.

## Suggested Pattern

```
project/
├── .github/prompts/        # Current prompts (active)
└── docs/archives/          # Historical prompts (reference)
    └── 2025-01-15_v1.0/
        └── prompts/
            └── smaqit.business.prompt.md
```

## Archiving Strategies

Users decide their own approach:

### 1. Git History (Implicit)
- **Approach**: No explicit archiving, rely on version control commits
- **Pros**: Zero overhead, complete history via git
- **Cons**: Requires git knowledge to access historical versions

### 2. Periodic Snapshots (Explicit)
- **Approach**: Copy prompts to `docs/archives/YYYY-MM-DD_vX.Y/` at milestones
- **Pros**: Easy to access historical versions without git
- **Cons**: Manual overhead, duplication

### 3. Version Tagging (Inline)
- **Approach**: Embed version/date in prompt filenames
- **Example**: `smaqit.business.prompt.v1.0.md`
- **Pros**: Self-documenting, simple to implement
- **Cons**: Multiple files per layer

## Critical Rule

**Agents always read from `.github/prompts/`**

Archived prompts are documentation only. Agents never read from archives.

## When to Archive

Consider archiving at:
- Major version releases
- Significant requirement pivots  
- Before large refactorings
- Project milestones or phases

## What to Archive

- All 8 prompt files (5 layer + 3 phase)
- Associated metadata (version, date, release notes)
- Optional: Generated specs for that version

## Benefits

- **Historical reference** for requirement evolution
- **Rollback capability** if needed
- **Audit trail** for compliance or retrospectives
- **Learning** from past decisions

## Related

- [Prompts as Input Records](../concepts/prompts-as-input-records.md) — Why prompts are versioned
- [Prompt Evolution](prompt-evolution.md) — How prompts grow with projects
- [Amending Requirements](../workflows/amending-requirements.md) — How to update prompts
