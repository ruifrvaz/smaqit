# Task 031: Review Implementation Artifacts

**Status:** New  
**Priority:** Medium  
**Created:** 2025-12-26

## Problem

Implementation agents (Development, Deployment, Validation) produce artifacts but the framework does not prescribe standard folder structures or artifact organization. This creates inconsistency and makes phase completion detection challenging.

Current state per [ARTIFACTS.md](../../framework/ARTIFACTS.md):
- Agents expected to follow "stack-specific standards" per Anchoring Principle
- No prescribed folder structure for implementation outputs
- No standardization across phases

## Goal

Review and standardize implementation artifact generation with focus on:
1. README content structure and placement
2. Artifact folder organization across phases
3. Phase completion detection mechanisms
4. Stack-specific flexibility vs framework consistency

## Approach

### README Review

**Current documentation:**
- Development agent produces README with "build, test, and run instructions"
- No prescribed location (project root vs src/)
- No template or structure requirements

**Questions:**
- What sections should README contain?
- Where should README be placed?
- Should there be a template?
- How does README relate to phase completion?

### Artifact Organization

**Current state:**
- No prescribed folders (src/, deployment/, validation/)
- Agents choose organization based on stack conventions

**Trade-offs:**
- Prescribing folders enables consistent `smaqit status` detection
- But conflicts with Anchoring Principle (follow stack conventions)
- Alternative: Use state.json for phase tracking (implemented in task 015)

**Investigation needed:**
- Survey common stack conventions (Node.js, Go, Python, Java)
- Identify if standardization is beneficial or restrictive
- Consider hybrid approach (standard top-level + stack-specific internals)

### Documentation & Templates

**Deliverables:**
- Artifact structure guidelines in [ARTIFACTS.md](../../framework/ARTIFACTS.md)
- Optional: README template for Development agent
- Optional: Report templates for Validation agent
- Updated agent directives if folder standards adopted

## Acceptance Criteria

- [ ] README structure defined or documented as flexible
- [ ] Artifact folder organization decision made (standardized, flexible, or hybrid)
- [ ] Framework documentation updated with artifact guidelines
- [ ] Implementation agents updated if standardization adopted
- [ ] No conflicts with Anchoring Principle (stack-specific freedom)

## Related

- **Task 015:** Implemented state.json for phase tracking (alternative to folder-based detection)
- **ARTIFACTS.md:** Current artifact documentation
- **Anchoring Principle:** Framework constraint on implementation flexibility

## Notes

- State.json tracking (task 015) reduces need for folder-based phase detection
- README is user-facing documentation, not just artifact marker
- Artifact review should balance consistency with stack-specific best practices
