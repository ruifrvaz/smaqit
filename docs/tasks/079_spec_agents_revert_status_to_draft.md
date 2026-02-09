# Spec Agents Revert Status to Draft on Modification

**Status:** Complete  
**Created:** 2026-02-02  
**Completed:** 2026-02-09

## Description

When specification agents (Business, Functional, Stack, Infrastructure, Coverage) modify their respective specs during iterative development, the spec status should be automatically reverted to `draft`.

This is part of the iterative development workflow which currently implements acceptance criteria checkbox resetting (Task 060). When a spec is modified by its agent, it indicates the spec is no longer in its previous state and needs to go through validation again.

This change requires updates across all three levels of the Level Up Architecture:
- **L0 (Framework):** Document the principle that modifications revert status
- **L1 (Templates):** Add directive to agent templates for status reversion
- **L2 (Product Agents):** Compile the directive into all five specification agents

## Acceptance Criteria

- [x] L0: Framework documents principle that spec modifications trigger status reversion to draft
- [x] L1: Agent template includes directive for status reversion on spec modification
- [x] L2: All five spec agents (business, functional, stack, infrastructure, coverage) include status reversion logic
- [x] Status reversion behavior is consistent with existing acceptance criteria reset behavior
- [x] Documentation explains when and why status is reverted

## Notes

- This aligns with the existing pattern from Task 060 (Reset Checkboxes on Requirement Refinement)
- Currently, spec agents already reset acceptance criteria checkboxes when requirements change
- Status reversion is the natural complement: checkboxes track granular progress, status tracks overall validation state
- This enforces the validation cycle: draft → validated → [modifications] → draft → validated again
