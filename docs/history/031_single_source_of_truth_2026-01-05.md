# Task 055: Formalize Single Source of Truth Principle

**Date:** 2026-01-05  
**Session Type:** Framework Enhancement  
**Task Completed:** 055  
**Related:** Task 048 (E2E Testing), Issue 3

## Overview

Formalized "Single Source of Truth" as an explicit Level 0 principle in the smaqit framework and cascaded it to all specification agents and templates. This addresses Issue 3 from E2E testing where Stack specs duplicated information (python-console-stack.md vs cli-stack.md).

## What Was Done

### Critical Assessment Phase

Before implementation:
- Loaded full project context via session.recap
- Read task 55 detailed requirements
- Examined Issue 3 from E2E testing report (Task 048)
- Reviewed framework files for existing "single source of truth" mentions (none found)
- Confirmed this is a Medium-priority improvement, not a release blocker
- Identified smaqit levels hierarchy approach: Level 0 → Level 1 → Level 2

### Level 0: Framework Foundation

**1. Added "Single Source of Truth" principle to framework/SMAQIT.md**
- Positioned after "Layer Independence" principle (related concepts)
- Bold statement: "Each piece of information should exist in exactly one place"
- Rationale: Prevents conflicting sources of truth, reduces maintenance burden, ensures consistency
- Agent directives:
  - MUST NOT duplicate information from existing specs
  - SHOULD update existing specs when extending concepts
  - SHOULD reference existing specs for shared information

**2. Enhanced framework/AGENTS.md with incremental spec guidance**
- Added to specification agents MUST: "Check for existing specs in same layer before creating new specs"
- Added to MUST NOT: "Duplicate information present in existing specs—use cross-references instead"
- Added to SHOULD: Three new directives for checking, updating, creating, and referencing
- Created new "Incremental Spec Updates vs New Specs" section with decision table:
  - Feature extends existing concept → Update existing spec
  - Feature is distinct new concept → Create new spec with cross-references
  - Shared infrastructure/base requirements → Reference existing spec
  - Uncertainty → Favor updating existing spec
- Included concrete examples (argparse to Python console, auth service creation)

**3. Enhanced framework/ARTIFACTS.md cross-reference pattern**
- Added "Same-Layer Reference" type to reference table
- Documented cross-layer reference format (already existed)
- Added same-layer reference format for avoiding duplication:
  ```markdown
  ### Base Requirements
  - [STK-CONSOLE](./python-console-stack.md) — See for base Python requirements
  ```
- Added same-layer reference rules with preference guidance

### Level 1: Template Structure

**4. Updated templates/specs/stack.template.md**
- Added "Base Requirements (if applicable)" section to References
- Included HTML comments explaining when to use same-layer references
- Example: "See [STK-CONSOLE](./python-console-stack.md) for base Python requirements"
- Guidance to omit section if no same-layer dependencies exist

**5. Updated templates/specs/functional.template.md**
- Added same "Base Requirements (if applicable)" section
- Parallel structure to Stack template for consistency
- Example: "See [FUN-AUTH](./auth-flow.md) for base authentication requirements"

**6. Updated templates/specs/infrastructure.template.md**
- Added same "Base Requirements (if applicable)" section
- Example: "See [INF-NETWORKING](./base-network.md) for base network configuration"

### Level 2: Agent Instances

**7. Updated agents/smaqit.business.agent.md**
- Added to MUST NOT: "Duplicate information from existing specs—use cross-references instead"
- Added to SHOULD:
  - Check for existing Business specs before creating new specs
  - Update existing specs when adding to existing concept
  - Create new specs only for distinct new use cases
  - Reference existing specs for shared information using cross-references

**8. Updated agents/smaqit.functional.agent.md**
- Same pattern as Business agent
- Example in SHOULD: "adding endpoint to existing API" for update case

**9. Updated agents/smaqit.stack.agent.md**
- Same pattern as other agents
- Specific example in SHOULD: "See [STK-CONSOLE](./python-console-stack.md) for base Python requirements"
- Example: "adding library to existing platform" for update case

**10. Updated agents/smaqit.infrastructure.agent.md**
- Same pattern as other agents
- Example: "adding monitoring to existing deployment" for update case

**11. Updated agents/smaqit.coverage.agent.md**
- Verified existing layer-specific MUST NOT already had "Add requirements not present in upstream specs"
- Added same directive to general MUST NOT for consistency across agents
- Added SHOULD guidance for checking, updating, and creating specs

## Design Decisions

### Level Hierarchy Approach

**Decision:** Follow smaqit's own methodology - work Level 0 first, then cascade to subsequent levels.

**Rationale:**
- Level 0 (Framework) defines principles that agents must follow
- Level 1 (Templates) provides structure for implementing principles
- Level 2 (Agents) are concrete instances that consume templates and follow framework
- Changes cascade naturally: Framework → Templates → Agents

**Benefits:**
- Ensures consistency across all levels
- Framework files embedded in agents will contain new principle
- Future agents generated from templates will include guidance automatically

### Principle Positioning

**Decision:** Placed "Single Source of Truth" after "Layer Independence" in SMAQIT.md.

**Rationale:**
- Both principles relate to avoiding unintended dependencies and duplication
- Layer Independence prevents vertical duplication (deriving requirements from upstream layers)
- Single Source of Truth prevents horizontal duplication (duplicating within same layer)
- Logical grouping improves framework comprehension

### Update vs Create Guidance

**Decision:** Favor updating existing specs over creating new specs when uncertain.

**Rationale:**
- Prevents premature spec proliferation
- Easier to split a spec later than to consolidate duplicate specs
- Reduces maintenance burden
- Aligns with "single source of truth" principle

**Example scenarios:**
- Adding argparse to Python console → Update python-console-stack.md
- Adding separate auth service → Create auth-service-stack.md, reference base stack

### Template Enhancement Scope

**Decision:** Added "Base Requirements" section to Stack, Functional, and Infrastructure templates (not Business or Coverage).

**Rationale:**
- Business layer is entry point, has no upstream specs to reference (no need)
- Coverage layer already has strong anti-duplication directive (derives from all upstream)
- Stack, Functional, Infrastructure are where same-layer duplication risk is highest
- These three layers most likely to have incremental spec additions

### Cross-Reference Pattern

**Decision:** Documented both cross-layer and same-layer reference patterns in ARTIFACTS.md.

**Rationale:**
- Cross-layer already existed (Implements/Enables)
- Same-layer needed for avoiding duplication within a layer
- Separate sections make distinction clear
- Examples show concrete usage patterns

## Problems Solved

### Problem 1: Stack Spec Duplication (Issue 3)

**Impact:** Medium - Maintenance burden, conflicting sources of truth

**Root Cause:** Stack agent created new spec (cli-stack.md) that duplicated base requirements from python-console-stack.md because no framework guidance existed on incremental updates vs new specs.

**Solution:**
- Level 0: Added principle with explicit directives
- Level 1: Templates now include "Base Requirements" section with guidance
- Level 2: All specification agents now have MUST NOT duplicate directive and SHOULD guidance

**Validation:** Stack agent now has explicit guidance: "Update existing specs when adding to an existing technology stack (e.g., adding library to existing platform)"

### Problem 2: Lack of Decision Criteria

**Impact:** Medium - Agents had no guidance on update vs create decisions

**Root Cause:** Framework mentioned single source of truth conceptually but didn't formalize decision criteria for agents.

**Solution:** Created "Incremental Spec Updates vs New Specs" decision table in AGENTS.md with:
- 4 scenario types with clear actions
- Concrete examples from real-world cases
- "Favor update" guidance for uncertainty

**Validation:** Decision table provides clear guidance agents can follow without human intervention.

### Problem 3: No Cross-Reference Pattern for Same-Layer

**Impact:** Low - Agents could avoid duplication but no standard format

**Root Cause:** ARTIFACTS.md documented cross-layer references (Implements/Enables) but not same-layer references.

**Solution:** Added "Same-Layer Reference" type with format and rules to ARTIFACTS.md and templates.

**Validation:** All three affected templates (Stack, Functional, Infrastructure) now include Base Requirements section with example format.

## Files Modified

**Framework (3):**
- framework/SMAQIT.md — Added "Single Source of Truth" principle
- framework/AGENTS.md — Added incremental spec update guidance and directives
- framework/ARTIFACTS.md — Documented same-layer reference pattern

**Templates (3):**
- templates/specs/stack.template.md — Added Base Requirements section
- templates/specs/functional.template.md — Added Base Requirements section
- templates/specs/infrastructure.template.md — Added Base Requirements section

**Agents (5):**
- agents/smaqit.business.agent.md — Added anti-duplication directives
- agents/smaqit.functional.agent.md — Added anti-duplication directives
- agents/smaqit.stack.agent.md — Added anti-duplication directives
- agents/smaqit.infrastructure.agent.md — Added anti-duplication directives
- agents/smaqit.coverage.agent.md — Added anti-duplication directives (already had layer-specific)

**Task Management (2):**
- docs/tasks/055_formalize_single_source_of_truth.md — Updated status to completed
- docs/tasks/PLANNING.md — Moved task 055 from Active to Completed

**Total:** 13 files modified

## Validation

**Installer build:**
```bash
cd installer && make build
# Result: SUCCESS - version d129d62
```

**Version check:**
```bash
./dist/smaqit --version
# Result: smaqit d129d62
```

**Consistency review:**
- Framework principle includes definition, rationale, agent directives ✓
- All 5 specification agents have consistent anti-duplication directives ✓
- Templates have Base Requirements sections with guidance ✓
- Decision table provides clear criteria ✓
- Cross-reference patterns documented ✓

## Key Learnings

### 1. Critical Assessment Before Implementation

Loading full project context via session.recap revealed:
- Issue 3 from E2E testing that motivated this task
- Framework structure (levels 0-3 hierarchy)
- Recent decisions about phase-first workflow
- Pattern from completed tasks (028, 037, 043)

This context enabled surgical implementation aligned with framework philosophy.

### 2. Level Hierarchy Matters

Working Level 0 → Level 1 → Level 2 ensured:
- Framework changes cascade to all consumers
- Templates reflect new framework rules
- Agents generated from templates inherit guidance automatically
- No conflicting directives across levels

This is smaqit's own principle applied to itself.

### 3. Favor Update Over Create

Decision criteria that "favor updating existing specs when uncertain" is pragmatic:
- Prevents premature spec proliferation
- Easier to split later than consolidate duplicates
- Reduces maintenance burden
- Aligns with principle itself

This guidance reduces agent decision paralysis.

### 4. Concrete Examples Beat Abstract Guidance

Including concrete examples throughout:
- "Adding argparse to Python console → Update existing"
- "Adding auth service → Create new with reference"
- Cross-reference format with actual paths

These examples give agents clear patterns to follow.

### 5. Consistency Across Agents

All 5 specification agents now have identical anti-duplication structure:
- Same MUST NOT directive
- Same SHOULD guidance
- Layer-specific examples in SHOULD

This consistency reduces cognitive load and ensures predictable behavior.

## Impact

**Framework Quality:**
- ✅ Single Source of Truth now explicit Level 0 principle
- ✅ Clear decision criteria for agents
- ✅ Standard cross-reference patterns documented
- ✅ Prevents Issue 3 recurrence

**Agent Behavior:**
- ✅ All specification agents have anti-duplication directives
- ✅ Clear guidance on update vs create decisions
- ✅ Concrete examples for common scenarios
- ✅ Consistent behavior across all layers

**Template Quality:**
- ✅ Base Requirements sections provide structure for avoiding duplication
- ✅ HTML comments guide agents when to use same-layer references
- ✅ Templates now support both cross-layer and same-layer references

**Maintenance:**
- ✅ Reduces future spec duplication risk
- ✅ Clear patterns for incremental development
- ✅ Easier to validate spec consistency
- ✅ No breaking changes to existing behavior

## Next Steps

**Immediate:**
- Task 055 completed and moved to Completed in PLANNING.md ✓
- Session history documented ✓

**Verification in Future E2E Testing:**
- Test incremental feature addition (Luigi scenario from Task 048)
- Verify agents check for existing specs before creating new ones
- Confirm agents use Base Requirements sections correctly
- Validate cross-reference format matches documentation

**Related Tasks:**
- Task 054: Strengthen Stack Agent Code Directive (related Stack agent improvement)
- Task 048: End-to-end agent workflow testing (discovered Issue 3)
- Tasks 049-053: High-priority blockers for v0.5.0-beta release

**Future Enhancements:**
- Consider adding `smaqit validate` check for duplicate information across specs
- Wiki article explaining single source of truth pattern to users
- Examples section in wiki showing good vs bad spec organization

## Session Metrics

**Duration:** ~1.5 hours  
**Tasks completed:** 1 (055)  
**Files modified:** 13  
**Lines added:** 105  
**Lines removed:** 8  
**Net change:** +97 lines  
**Commits:** 2  
**Builds:** 1 successful  

**Coverage by Level:**
- Level 0 (Framework): 3 files
- Level 1 (Templates): 3 files
- Level 2 (Agents): 5 files
- Task Management: 2 files

**Agent updates:**
- Business ✓
- Functional ✓
- Stack ✓
- Infrastructure ✓
- Coverage ✓

**Validation:**
- Build successful ✓
- Version matches commit ✓
- Consistency review passed ✓
- All acceptance criteria met ✓

## Conclusion

Task 055 successfully formalized "Single Source of Truth" as an explicit framework principle and cascaded it through all levels of smaqit architecture. The implementation follows smaqit's own methodology (Level 0 → Level 1 → Level 2), ensures consistency across agents, and provides clear decision criteria to prevent future spec duplication like Issue 3.

The changes are non-breaking, surgical, and will be embedded in agents through the installer's framework embedding mechanism. Future E2E testing will validate that agents correctly follow the new guidance when adding incremental features.
