# Framework Evolution (Meta)

**Date:** 2025-12-21  
**Type:** Meta-framework documentation

## Overview

This document captures how smaqit framework evolves and the lessons learned from experimentation.

## Iteration Through Experimentation

smaqit framework emerged from prototyping agent workflows. Key lessons learned:

### What Didn't Work

**Agents deriving requirements from upstream layers**
- Created false dependencies between layers
- Made it unclear where requirements originated
- Led to cascading changes when one layer changed

**Rigid prompt structures**
- Fought against natural user expression
- Created friction in requirement capture
- Didn't accommodate diverse project types

**Template suggestions instead of mandates**
- LLM variance broke downstream consumption
- Inconsistent output structure required manual normalization
- Downstream agents couldn't reliably parse inputs

### What Worked

**Layer independence through prompt files**
- Users control each layer explicitly
- No false derivation chains
- Clear source of truth for requirements

**Free-style prompts with suggested structure**
- Captured natural requirements
- Reduced friction
- Agents interpret natural language well

**Template constraints**
- Reduced variance
- Enabled predictable downstream consumption
- Consistent structure across runs

## Amendment Protocol

When framework concepts change, smaqit follows a structured refactoring process:

### Steps

1. **Update principles** in framework files (SMAQIT.md, LAYERS.md, etc.)
2. **Update templates** to reflect new principles (templates/, agents/)
3. **Regenerate agents** from updated templates
4. **Test workflows** with user testing agent
5. **Document changes** in history files

### Self-Application

The framework is self-applying—it uses its own principles to evolve:
- Specs before code (framework files are specs for agents)
- Traceability (changes trace through framework → templates → agents)
- Layer independence (framework concepts are independent but coherent)
- Template-constrained output (agents follow templates strictly)

## Historical Pivots

### December 2025: Layer Independence

**Problem:** "Deterministic from Input" principle conflicted with layer independence concept.

**Solution:** 
- Renamed to "Reproducible from Input Set"
- Updated terminology: "user input" → "prompt file"
- Each layer reads from its own prompt file

**Impact:** 26 files updated (framework/, agents/, templates/)

### December 2025: Documentation Separation

**Problem:** Framework files mixed execution instructions with rationale.

**Solution:**
- Created docs/wiki/ for human context
- Stripped framework files to pure instructions
- Moved rationale to wiki (7 documents)

**Impact:** SMAQIT.md and PROMPTS.md reduced by ~150 lines combined

## Lessons for Future Evolution

1. **Test with user testing agent** before considering changes complete
2. **Update all three levels** (framework → templates → agents) when principles change
3. **Document rationale in wiki**, not framework files
4. **Keep framework files pure** - instructions only, no examples or history
5. **Commit framework changes as atomic units** - don't mix principle changes with unrelated edits

## Related

- [session.wrap command](../tasks/PLANNING.md) — How to document significant sessions
- [Documentation Structure](../../README.md#documentation-structure) — Where different content types belong
