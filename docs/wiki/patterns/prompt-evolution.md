# Prompt Evolution Patterns

## Single Manifest per Layer

Unlike specifications (one file per concept), prompts are **single manifest files** that capture all requirements for a layer.

## Growth Pattern

As projects evolve:
- **Add** new features to existing prompts (don't create new prompt files)
- **Mark** deprecated features (keep for history)
- **Amend** existing requirements (version control tracks changes)

## Example Evolution

**Initial Business Prompt:**
```markdown
## Use Cases

### User Login
Users can authenticate with email/password.
```

**After Adding Feature:**
```markdown
## Use Cases

### User Login
Users can authenticate with email/password.

### Social Login
Users can authenticate with Google or GitHub OAuth.
```

**After Deprecation:**
```markdown
## Use Cases

### User Login
Users can authenticate with email/password.

### Social Login
**Status:** Deprecated (2025-12-20)  
**Reason:** Complexity vs. usage didn't justify maintenance

Users can authenticate with Google or GitHub OAuth.
```

## Why Single Manifest?

### Benefits
- **Consolidated record** of all project requirements at each layer
- **Single source** to review when onboarding or auditing
- **Simpler workflow** (edit one file vs. many)

### Trade-offs
- **File size** grows with project (mitigated by git history)
- **Merge conflicts** if multiple people edit simultaneously (rare for prompts)

## Status Markers

Users may adopt patterns like:

```markdown
## Feature: [Name]
**Status:** Active | Deprecated | Planned
**Added:** YYYY-MM-DD
**Deprecated:** YYYY-MM-DD (if applicable)
**Reason:** (for deprecations)

[Requirements here]
```

This is optional. Agents should tolerate varied styles.

## Related

- [Prompts as Input Records](../concepts/prompts-as-input-records.md) — Why prompts are versioned
- [Amending Requirements](../workflows/amending-requirements.md) — How to update prompts
- [Archiving Prompts](archiving-prompts.md) — Managing historical versions
