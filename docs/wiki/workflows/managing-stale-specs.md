# Managing Stale Specs

## Overview

Specs become stale when their source prompt evolves. smaqit tracks prompt versions but leaves regeneration decisions to users.

## Stale Detection

**Automated Detection (Future):**

Specs contain `prompt_version` field capturing the git commit hash when generated:

```yaml
---
id: BUS-LOGIN-001
status: implemented
prompt_version: abc123def
---
```

**Manual Detection (Current):**

Compare:
1. Prompt file's latest commit: `git log -1 --format=%H .github/prompts/smaqit.business.prompt.md`
2. Spec's `prompt_version` field

If different, prompt has evolved since spec was generated.

## Stale Spec Workflow

### 1. Identify Staleness

Check prompt version mismatch:
```bash
# Get current prompt commit
git log -1 --format=%H .github/prompts/smaqit.business.prompt.md

# Compare to spec frontmatter
grep "prompt_version:" specs/business/login.md
```

### 2. Assess Impact

Questions to ask:
- What changed in the prompt? (Check git diff)
- Does this spec capture the new requirements?
- Is regeneration necessary or can I manually update?

### 3. Decide Action

| Scenario | Action |
|----------|--------|
| Minor prompt clarification | Keep spec as-is (judgment call) |
| New requirements added | Regenerate spec to incorporate |
| Requirements modified | Regenerate spec with updated criteria |
| Requirements deprecated | Regenerate or manually mark deprecated |

### 4. Regenerate If Needed

Execute: `/smaqit.[layer]` (e.g., `/smaqit.business`)

Agent will:
- Read current prompt
- Generate updated spec
- Preserve existing spec state (implemented, deployed, validated)
- Update `prompt_version` to current commit

**State preservation:**
- If spec was `status: implemented`, regeneration keeps status
- If spec structure changes significantly, state resets to `draft`
- Implementation agents decide if existing code still satisfies regenerated spec

### 5. Re-implement If Needed

If regeneration introduced new acceptance criteria:
- Run `/smaqit.development` to update implementation
- Spec state transitions: `implemented` → `draft` → `implemented` (after code update)

## Batch Regeneration

**Regenerate all specs for a layer:**
```
User: /smaqit.business
```

Agent regenerates all business specs with current prompt version.

**Regenerate specific specs:**
Currently not supported. Regenerate entire layer, then cherry-pick changes.

## Staleness Tolerance

Not all stale specs need regeneration:

- **Prompt typo fixes**: Usually ignorable
- **Prompt formatting changes**: Usually ignorable
- **New features added**: Only regenerate if affected specs overlap
- **Requirements changed**: Always regenerate

Users decide based on judgment and context.

## Git Workflow Integration

**Before regenerating:**
```bash
# See what changed in prompt
git diff abc123def HEAD .github/prompts/smaqit.business.prompt.md
```

**After regenerating:**
```bash
# Commit with reference to prompt change
git commit -m "Regenerate business specs after adding social login requirement"
```

## Future Enhancements

Potential tooling (not yet implemented):

```bash
# List stale specs
smaqit stale

# Regenerate only stale specs
smaqit regenerate --stale

# Show prompt diff for stale spec
smaqit diff BUS-LOGIN-001
```

## Related

- [Stateful Specifications](../concepts/stateful-specifications.md) — How state tracking works
- [Amending Requirements](amending-requirements.md) — Prompt amendment workflow
- [Prompt Evolution](../patterns/prompt-evolution.md) — Why prompts evolve
