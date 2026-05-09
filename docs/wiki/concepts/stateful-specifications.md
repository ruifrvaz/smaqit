# Stateful Specifications

## Overview

Specifications in smaqit track their lifecycle through YAML frontmatter, enabling incremental development and progress visibility.

## State Lifecycle

```
draft → implemented → deployed → validated
  ↓          ↓            ↓
failed    failed       failed
  ↓          ↓            ↓
deprecated (optional)
```

### State Definitions

| State | Meaning | Set By |
|-------|---------|--------|
| `draft` | Spec generated but not yet implemented | Specification agents |
| `implemented` | Code written and tests passing | Development agent |
| `deployed` | Infrastructure provisioned and running | Deployment agent |
| `validated` | Coverage verified and tests executed | Validation agent |
| `failed` | Implementation/deployment/validation failed | Implementation agents |
| `deprecated` | Spec no longer active (manual user action) | User |

## Frontmatter Schema

**Required fields (set at creation):**
```yaml
---
id: BUS-LOGIN-001
status: draft
created: 2025-01-03T14:23:00Z
---
```

**Optional fields (set during implementation):**
```yaml
implemented: 2025-01-04T09:15:00Z
deployed: 2025-01-04T10:30:00Z
validated: 2025-01-04T11:45:00Z
```

## Incremental Development

State tracking enables:

1. **Partial progress** — Implement 3 out of 10 specs without regenerating all
2. **Resume workflows** — Pick up where you left off across sessions
3. **Targeted regeneration** — Regenerate only failed specs
4. **Progress visibility** — See which specs are done vs. pending

## Design Rationale

### Why Frontmatter?

- **Colocation**: State lives with spec content (like tasks)
- **Simplicity**: No external database or state files per spec
- **Git-friendly**: Text-based, mergeable, diffable
- **Self-contained**: Specs carry their own state

### Why Not External State File Per Spec?

- **Overhead**: One JSON file per spec = filesystem clutter
- **Fragmentation**: State separated from content (harder to review)
- **Merge complexity**: Two-file updates for every state change

### How Does CLI Track Phase Status?

- **Scanning**: CLI scans all spec files and parses frontmatter on demand
- **Aggregation**: Counts specs by layer and status to determine phase completion
- **No cache**: Always reads current state from spec files (source of truth)

## Related

- [Progressive Refinement](../designs/progressive-refinement.md) — How specs evolve through phases
