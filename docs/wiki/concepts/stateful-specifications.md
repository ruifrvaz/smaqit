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
prompt_version: abc123def
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

## Phase State Aggregation

Implementation agents update `.smaqit/state.json` with aggregate counts:

```json
{
  "version": "1.0",
  "phases": {
    "develop": {
      "completed": true,
      "timestamp": "2025-01-04T12:00:00Z",
      "specs_processed": 15,
      "specs_succeeded": 14,
      "specs_failed": 1
    },
    "deploy": {
      "completed": false,
      "specs_processed": 0,
      "specs_succeeded": 0,
      "specs_failed": 0
    },
    "validate": {
      "completed": false,
      "specs_processed": 0,
      "specs_succeeded": 0,
      "specs_failed": 0
    }
  }
}
```

## Prompt Version Tracking

`prompt_version` field captures the git commit hash of the prompt used to generate the spec:

- **Purpose**: Detect when prompt has evolved beyond spec
- **Detection**: Compare spec's `prompt_version` against current prompt file commit
- **Action**: User decides whether to regenerate (see [Managing Stale Specs](../workflows/managing-stale-specs.md))

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

### Why Aggregate to state.json?

- **CLI efficiency**: Single file read for phase status
- **Progress visibility**: Quick overview without scanning all specs
- **Phase tracking**: Unified view of completion across layers

## Related

- [Managing Stale Specs](../workflows/managing-stale-specs.md) — Detecting and handling outdated specs
- [Amending Requirements](../workflows/amending-requirements.md) — How state changes on regeneration
- [Prompt Evolution](../patterns/prompt-evolution.md) — Why prompts evolve and specs follow
