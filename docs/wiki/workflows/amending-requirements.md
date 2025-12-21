# Amending Requirements Workflow

## Overview

When requirements change, users edit prompts and regenerate specs. Prompts are the source, specs are derived.

## Step-by-Step Workflow

### 1. Edit Prompt File

Location: `.github/prompts/smaqit.[layer].prompt.md`

Actions:
- Add new requirements
- Modify existing requirements  
- Deprecate outdated requirements (mark as such, don't delete for history)

### 2. Commit Prompt Changes

Version control captures requirement evolution:
- Prompt diffs show explicit requirement changes
- Commit message explains what requirements changed and why

### 3. Re-invoke Agent

Execute: `/smaqit.[layer]` (e.g., `/smaqit.business`)

Agent actions:
- Reads amended prompt
- Regenerates specs incorporating changes
- Updates existing specs or creates new ones

### 4. Review Generated Specs

Validation checks:
- Verify specs reflect prompt amendments
- Check for conflicts with upstream layers
- Validate acceptance criteria updated appropriately

### 5. Commit Spec Changes

Link spec changes to prompt changes:
- Commit message references prompt amendment
- Maintains traceability: requirement change → spec change

## Amendment Types

### Modify Existing Feature
- **Action**: Modify prompt → regenerate existing specs
- **Example**: Change performance requirement from 100ms to 50ms

### Add New Feature
- **Action**: Add to prompt → generate new specs  
- **Example**: Add new use case to business prompt

### Deprecate Feature
- **Action**: Mark requirement as deprecated in prompt → mark specs as deprecated
- **Example**: Mark "Social Login" as deprecated in prompt

## Prompt Evolution Patterns

Prompts should accommodate ongoing evolution. Users may add markers like:

```markdown
## Feature: User Login
**Status:** Active
[Requirements here]

## Feature: Social Login
**Status:** Deprecated (2025-12-20)
[Requirements kept for history]
```

This is optional and user-defined. Agents should tolerate varied styles.

## Traceability Chain

```
Prompt Amendment → Agent Re-invocation → Spec Regeneration → Implementation Update
```

Each step is explicit and traceable through version control.

## Related

- [Prompts as Input Records](../concepts/prompts-as-input-records.md) — Why prompts are versioned
- [Prompt Evolution](../patterns/prompt-evolution.md) — How prompts grow with projects
- [Archiving Prompts](../patterns/archiving-prompts.md) — Managing historical versions
