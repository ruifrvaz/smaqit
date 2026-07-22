---
version: "1.0.0"
---

# Spec Lifecycle Reference

Distilled mechanics this skill depends on for spec revalidation and status tracking. Self-contained — no dependency on `framework/` (not shipped to installed projects).

## Spec Frontmatter and State

```yaml
---
id: [LAYER_PREFIX]-[CONCEPT]
status: draft | implemented | deployed | validated | failed | deprecated
created: [ISO8601_TIMESTAMP]
implemented: [ISO8601_TIMESTAMP]
deployed: [ISO8601_TIMESTAMP]
validated: [ISO8601_TIMESTAMP]
---
```

**State Transitions:**

| From State | To State | Triggered By |
|------------|----------|---------------|
| (none) | `draft` | Spec generation |
| `draft` | `implemented` | Code generated, tests pass |
| `draft` | `failed` | Code generation failed |
| `implemented` | `deployed` | Deployment succeeded |
| `implemented` | `failed` | Deployment failed |
| `deployed` | `validated` | All tests passed |
| `deployed` | `failed` | Tests failed |
| Any | `deprecated` | Feature removed |

**Acceptance criteria checkbox states:** `[ ]` not yet implemented/validated · `[x]` satisfied · `[!]` failed, untestable, or not satisfied.

**On modification of an existing spec:** status reverts to `draft` regardless of prior state; any checkbox on a criterion whose text changed resets `[x]`/`[!]` → `[ ]`. Checkboxes on unchanged criteria are left as-is.

| Previous Status | After Modification |
|------------------|---------------------|
| `implemented` | `draft` |
| `deployed` | `draft` |
| `validated` | `draft` |
| `failed` | `draft` |

## Incremental Plan Resolution

| Mode | Command | Processes |
|------|---------|-----------|
| Incremental | `smaqit plan --phase=develop` | Only specs with `status: draft` or `status: failed` |
| Regeneration | `smaqit plan --phase=develop --regen` | All specs regardless of status |

Adding a feature: (1) spec agent invoked with updated requirements in session context → generates/updates specs (`status: draft`); (2) `smaqit plan --phase=develop` (no `--regen`) returns only the new/changed draft or failed specs; (3) existing `implemented`/`deployed`/`validated` specs not touched by the feature are skipped.

## Incremental Spec Updates vs New Specs

| Scenario | Action |
|----------|--------|
| Feature extends existing concept | Update existing spec |
| Feature is distinct new concept | Create new spec with Foundation Reference |
| Shared infrastructure/base requirements | Create foundation spec, reference from feature specs |
| Uncertainty | Favor updating existing spec |

**Examples:**

| Requirement | Existing Spec | Decision |
|-------------|---------------|----------|
| Add argparse CLI to Python console app | `python-console-stack.md` exists | Update existing spec |
| Add authentication service to app | `app-stack.md` exists | Create `auth-service-stack.md`, reference `[STK-APP](./app-stack.md)` |
| Add logging to existing feature | `feature-functional.md` exists | Update existing spec |

**Foundation Reference format** (feature spec extends a foundation spec in the same layer):

```markdown
## References

### Foundation Reference
- [STK-[FOUNDATION-CONCEPT]](./base-stack.md) — Shared requirements referenced here

### Implements
- [FUN-[CONCEPT]-NNN](../functional/feature.md) — Implements feature functionality
```

## Specification Agent Mappings

| Agent | Layer | Context (for coherence) | Output |
|-------|-------|--------------------------|--------|
| `smaqit.business` | Business | None | `specs/business/*.md` |
| `smaqit.functional` | Functional | Business specs | `specs/functional/*.md` |
| `smaqit.stack` | Stack | Business and Functional specs | `specs/stack/*.md` |
| `smaqit.infrastructure` | Infrastructure | Phase 1 specs | `specs/infrastructure/*.md` |
| `smaqit.coverage` | Coverage | All layer specs | `specs/coverage/*.md` |

## Frontmatter Fields Updated by Implementation Agents

| Agent | Fields Updated |
|-------|-----------------|
| Development | `status: implemented` or `failed`; `implemented: [ISO8601_TIMESTAMP]` |
| Deployment | `status: deployed` or `failed`; `deployed: [ISO8601_TIMESTAMP]` |
| Validation | `status: validated` or `failed`; `validated: [ISO8601_TIMESTAMP]`; checkboxes `[ ]` → `[x]`/`[!]` |
