# Status Command Intelligent Next Step Logic

**Status:** Not Started  
**Created:** 2025-12-28  
**Source:** User Testing Report Issue #1 (2025-12-27)

## Description

Make `smaqit status` command phase-aware with intelligent next step suggestions. Currently, after Business layer spec is created, status still suggests `/smaqit.development` which is premature when only 1 of 3 Phase 1 specs exists.

## Acceptance Criteria

- [ ] Status command detects incomplete specification layers within current phase
- [ ] Suggests next layer prompt when phase specs are incomplete (e.g., Business exists → suggest `/smaqit.functional`)
- [ ] Suggests implementation agent only when all required phase specs are complete
- [ ] Phase 1 (Develop) requires Business + Functional + Stack before suggesting `/smaqit.development`
- [ ] Phase 2 (Deploy) requires Infrastructure before suggesting `/smaqit.deployment`
- [ ] Phase 3 (Validate) requires Coverage before suggesting `/smaqit.validation`

## Impact

**Severity:** Medium  
**User Impact:** Misleads users who might try to start implementation phase prematurely with incomplete specs

## Notes

Related to Issue #6 (phase-first workflow clarity). This fix makes the CLI better reflect the phase-based workflow.
