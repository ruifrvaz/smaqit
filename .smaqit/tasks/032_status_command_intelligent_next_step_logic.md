# Status Command Intelligent Next Step Logic

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2026-01-02  
**Source:** User Testing Report Issue #1 (2025-12-27)

## Description

Make `smaqit status` command phase-aware with intelligent next step suggestions. Currently, after Business layer spec is created, status still suggests `/smaqit.development` which is premature when only 1 of 3 Phase 1 specs exists.

## Acceptance Criteria

- [x] Status command detects incomplete specification layers within current phase
- [x] Suggests next layer prompt when phase specs are incomplete (e.g., Business exists → suggest `/smaqit.functional`)
- [x] Suggests implementation agent only when all required phase specs are complete
- [x] Phase 1 (Develop) requires Business + Functional + Stack before suggesting `/smaqit.development`
- [x] Phase 2 (Deploy) requires Infrastructure before suggesting `/smaqit.deployment`
- [x] Phase 3 (Validate) requires Coverage before suggesting `/smaqit.validation`

## Solution

Replaced simple phase completion check with intelligent spec-aware logic:

1. **Empty project** → suggests `/smaqit.business`
2. **Missing Phase 1 layers** → suggests missing layer in order (business → functional → stack)
3. **All Phase 1 specs present** → suggests `/smaqit.development`
4. **Phase 1 complete, no infrastructure** → suggests `/smaqit.infrastructure`
5. **Phase 1 complete, has infrastructure** → suggests `/smaqit.deployment`
6. **Similar logic for Phase 3** (coverage/validation)

## Files Modified

- `installer/main.go`: Updated `cmdStatus()` next steps logic (lines 715-750)
  - Checks `layerCounts` map for actual spec file presence
  - Progressive suggestions based on phase and layer state
  - Prevents premature implementation suggestions

## Testing Results

All test scenarios passed:
- ✓ Empty project → `/smaqit.business`
- ✓ Business only → `/smaqit.functional`
- ✓ Business + Functional → `/smaqit.stack`
- ✓ All Phase 1 specs → `/smaqit.development`
- ✓ Phase 1 done, no infra → `/smaqit.infrastructure`
- ✓ Phase 1 done, has infra → `/smaqit.deployment`
