# Task 055: Formalize Single Source of Truth Principle

**Status:** completed  
**Priority:** Medium  
**Created:** 2026-01-05  
**Completed:** 2026-01-05  
**Related:** Task 048 (E2E Testing), Issue 3

## Problem

New Stack spec (`cli-stack.md`) duplicates information from existing Stack spec (`python-console-stack.md`), including Python version, build tools, and development environment requirements. This creates conflicting sources of truth and maintenance burden.

**Impact:**
- Creates potential for conflicting sources of truth if specs diverge
- Maintenance burden—must update multiple specs for same information
- Confusion about which spec is authoritative for shared requirements
- Increases likelihood of spec inconsistency over time

**Framework Gap:** While "single source of truth" is mentioned in framework context, it's not formalized as an explicit principle with agent directives.

**Root Cause:** Stack agent treated incremental feature as requiring completely new spec rather than updating existing spec. No framework guidance on incremental spec updates vs new spec creation.

## Objective

Elevate "single source of truth" to explicit Level 0 principle in `framework/SMAQIT.md`, then cascade to agent directives to prevent information duplication across specs.

## Acceptance Criteria

- [x] Added "Single Source of Truth" as explicit principle in `framework/SMAQIT.md`
- [x] Principle includes: Definition, rationale, agent directives
- [x] Cascaded to specification agents: Added directive "MUST NOT duplicate information from existing specs"
- [x] Added guidance on when to update existing specs vs create new specs
- [x] Documented cross-reference pattern for shared information
- [x] Updated Stack agent specifically (covered by general principle + specific guidance)
- [x] Updated Stack template with "Base Requirements" section for same-layer references
- [x] Updated Functional and Infrastructure templates with same pattern

## Implementation Plan

### Phase 1: Framework Level 0

1. **Add principle to `framework/SMAQIT.md`:**
   - Section: "Core Principles"
   - Title: "Single Source of Truth"
   - Definition: "Each piece of information should exist in exactly one place. When information is needed in multiple contexts, reference the source rather than duplicate."
   - Rationale: Prevents conflicting sources of truth, reduces maintenance burden, ensures consistency
   - Directive: "Agents MUST NOT duplicate information from existing specs. Use cross-references for shared information."

### Phase 2: Agent Directives (Level 2)

2. **Update all specification agents** (Business, Functional, Stack, Infrastructure, Coverage):
   - Add to MUST NOT section: "Duplicate information present in existing specs—use cross-references instead"
   - Add to SHOULD section: "Reference existing specs for shared information (e.g., 'See STK-CONSOLE for Python version requirements')"

3. **Add guidance on incremental spec updates:**
   - Document in `framework/AGENTS.md` or agent files:
     - **Update existing spec when:** Adding to existing concept (e.g., adding argparse to Python console app)
     - **Create new spec when:** Introducing distinct new concept (e.g., separate service/component)
   - Provide decision criteria for agents

### Phase 3: Cross-Reference Pattern

4. **Document cross-reference pattern:**
   - Add to `framework/ARTIFACTS.md` or create wiki article
   - Example: "Spec B references Spec A for shared requirements: 'See [STK-CONSOLE](./console-stack.md) for base Python requirements. This spec adds CLI argument parsing with argparse.'"

5. **Optional: Update Stack template:**
   - Add "References to Existing Specs" section to `templates/specs/stack.template.md`
   - Encourage agents to check for existing Stack specs and reference them

## Files to Modify

- `framework/SMAQIT.md` (add principle)
- `framework/AGENTS.md` (add guidance on incremental updates)
- `agents/smaqit.business.agent.md` (add directive)
- `agents/smaqit.functional.agent.md` (add directive)
- `agents/smaqit.stack.agent.md` (add directive)
- `agents/smaqit.infrastructure.agent.md` (add directive)
- `agents/smaqit.coverage.agent.md` (already has anti-duplication directive, verify)
- `framework/ARTIFACTS.md` (optional, document cross-reference pattern)
- `templates/specs/stack.template.md` (optional, add References section)

## Testing

**Manual verification:**
1. Read updated principle in SMAQIT.md
2. Verify principle is clear and unambiguous
3. Check agent directives for anti-duplication rules
4. Verify guidance on update vs create is actionable

**Optional agent test:**
1. Create test project with existing spec
2. Add incremental feature requirement
3. Run spec agent
4. Verify agent either: (a) updates existing spec, OR (b) creates minimal new spec with cross-reference

## Estimated Effort

2 hours (includes principle definition, agent cascading, and guidance documentation)

## Dependencies

None (can be implemented independently)

## Blocks

None directly, but prevents future duplication issues

## Related Tasks

- Task 048: E2E Agent Workflow Testing (discovered this issue)
- Task 054: Strengthen Stack Agent Code Directive (related Stack agent improvement)

## Notes

**Current mentions:** "Single source of truth" appears in context throughout framework but is not formalized as principle with explicit directives.

**Why formalize:** Elevating to Level 0 principle ensures all agents understand and enforce this pattern. Without explicit directive, agents will continue creating duplicate information.

**Tradeoff:** Cross-referencing increases spec coupling (Spec B depends on Spec A). However, duplication increases maintenance burden and consistency risk. Cross-referencing is lesser evil.

**Coverage layer example:** Coverage agent already has strong anti-duplication directive ("MUST NOT add requirements not present in upstream specs"). This establishes precedent that can be applied to other layers.

**Decision criteria for agents:**

| Scenario | Action | Example |
|----------|--------|---------|
| Feature extends existing concept | Update existing spec | Adding argparse to Python console app → Update python-console-stack.md |
| Feature is distinct new concept | Create new spec with cross-reference | Adding separate authentication service → Create auth-service-stack.md, reference shared base requirements |
| Uncertainty | Favor update over duplication | When in doubt, extend existing spec |

**User guidance:** This principle should also be documented in user-facing materials (README or wiki) to help users understand why specs are structured certain ways.
