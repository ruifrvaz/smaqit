# Task 038 Complete: Add state.json Validation to Validate Command

**Date:** 2025-12-28  
**Session Type:** Feature implementation  
**Tasks Completed:** 038

## Overview

Implemented comprehensive state.json validation in `smaqit validate` command. The validate command now checks for file existence, JSON validity, required fields, correct types, and semantic phase ordering, providing actionable error messages for all validation failures.

## What Was Done

### Task 038: State.json Validation

**Implementation:**

Added two new functions to `installer/main.go`:

1. **`validateStateFile(path string) int`** (138 lines)
   - Checks file existence with recovery guidance
   - Validates JSON parseability with syntax error details
   - Verifies version field (presence, type, non-empty)
   - Verifies phases object (presence, type)
   - Checks all required phase keys (develop, deploy, validate)
   - Validates each phase structure (completed boolean field)
   - Supports forward compatibility (allows additional fields)
   - Returns error count for aggregate reporting

2. **`validatePhaseOrdering(jsonContent []byte, path string) error`**
   - Checks if phases appear in workflow order (develop → deploy → validate)
   - Returns warning only (semantic issue, not functional)
   - Integrated into validateStateFile as non-fatal check

**Integration:**

Modified `cmdValidate()` to call `validateStateFile()` and aggregate errors from directory structure and state file validation.

**Error Messages:**

All validation errors provide actionable guidance with → prefix:
- Missing file: "→ Run 'smaqit init' to create state.json"
- Invalid JSON: "→ Fix JSON syntax or regenerate with 'smaqit init' (after backup)"
- Missing version: "→ Add: \"version\": \"1.0\""
- Missing phases: "→ Add phases object with develop, deploy, validate keys"
- Missing phase: "→ Add: \"[phase]\": {\"completed\": false}"
- Missing completed: "→ Add 'completed' boolean to [phase] phase"
- Wrong type: "→ Change 'completed' to true or false in [phase] phase"

**Testing:**

Manually tested 10 scenarios:

| Scenario | Expected | Result |
|----------|----------|--------|
| Valid state.json | Pass | ✅ Pass |
| Missing file | Error with guidance | ✅ Error |
| Invalid JSON syntax | Error with parse details | ✅ Error |
| Missing version field | Error with fix | ✅ Error |
| Empty version field | Error with fix | ✅ Error |
| Wrong version type | Error with type guidance | ✅ Error |
| Missing phases object | Error with guidance | ✅ Error |
| Wrong phases type | Error with type guidance | ✅ Error |
| Missing phase keys | Error per missing phase | ✅ Error |
| Missing completed field | Error per phase | ✅ Error |
| Wrong completed type | Error with type guidance | ✅ Error |
| Additional fields | Pass (forward compatible) | ✅ Pass |
| Wrong phase ordering | Warning only | ✅ Warning |

### Acceptance Criteria Met

- [x] Validate command checks `.smaqit/state.json` file exists
- [x] Validates JSON structure is valid (parseable)
- [x] Verifies phase keys are present (develop, deploy, validate)
- [x] Verifies phase keys are correctly ordered in JSON
- [x] Validates each phase object has required "completed" boolean field
- [x] Validates "version" field is present at root level
- [x] Reports specific validation errors with actionable messages
- [x] Does not fail on additional fields (forward compatibility)

## Decisions Made

### Phase Ordering as Warning, Not Error

**Decision:** Phase ordering check reports a warning, not an error.

**Rationale:** 
- Phase ordering in JSON is a semantic/readability issue
- Does not affect functionality (JSON objects are unordered by spec)
- Go's map iteration is non-deterministic anyway
- `readStateFile()` validates presence, not order
- Graceful degradation: warn but don't block validation

**Implementation:** `validatePhaseOrdering()` returns error type but is treated as warning in output.

### Forward Compatibility

**Decision:** Allow additional fields at both root and phase object levels.

**Rationale:**
- Future versions may add new fields (e.g., "last_updated", "user_metadata")
- Strict validation would break older CLI with newer state files
- Only validate required fields; ignore extras
- Aligns with JSON schema extensibility best practices

**Implementation:** Parse into `map[string]interface{}` and check only required keys.

### Actionable Error Messages

**Decision:** Every validation error includes → prefixed guidance on how to fix.

**Rationale:**
- Users shouldn't need to consult docs for common fixes
- Self-documenting error messages reduce support burden
- Consistent format (✗ error + → action) improves UX
- Follows established pattern from existing validation

**Implementation:** Each error printf includes two lines: problem + solution.

## Problems Solved

### Problem 1: Corrupted state.json Causes Silent Failures

**Before:** `smaqit status` gracefully handled corrupted state.json by returning default state, masking corruption.

**After:** `smaqit validate` catches corruption early with specific error messages, helping users fix issues before they impact workflows.

### Problem 2: No Validation of State File Structure

**Before:** Only `readStateFile()` performed defensive parsing; no proactive validation.

**After:** Comprehensive pre-emptive validation in `validate` command catches structural issues.

### Problem 3: Generic Error Messages

**Before:** Validation errors were generic (e.g., "state.json is corrupted").

**After:** Specific, actionable errors guide users to exact fixes.

## Files Modified

**Implementation (1 file):**
- `installer/main.go`
  - Added `validateStateFile()` function (118 lines)
  - Added `validatePhaseOrdering()` function (20 lines)
  - Modified `cmdValidate()` to integrate state validation

**Documentation (2 files):**
- `docs/tasks/038_add_state_json_validation_to_validate_command.md`
  - Updated status to Completed
  - Added implementation details
  - Documented testing results
- `docs/tasks/PLANNING.md`
  - Moved task 038 from Active to Completed

**History (1 file):**
- `docs/history/014_task_038_state_json_validation_2025-12-28.md` (this file)

**Total:** 4 files modified/created

## Next Steps

### Immediate

1. **Task 033:** Fix state.json phase ordering in `initStateFile()`
   - Related issue discovered during testing
   - Current init creates phases in wrong order (deploy, develop, validate)
   - Should be: develop, deploy, validate

2. **Build and Release:** Rebuild installer with validation changes
   - Run `make build-all` for all platforms
   - Update release notes with validation feature

### Suggested

1. **Task 032:** Status command intelligent next step logic
   - Now that state validation is robust, enhance status suggestions
   - Could leverage validated state structure for smarter recommendations

2. **Integration Testing:** Add automated tests for state.json validation
   - Currently manually tested (10 scenarios)
   - Could add Go unit tests for validateStateFile()

## Key Learnings

### 1. Validation Should Be Defensive and Graceful

Phase ordering implemented as warning, not error, because:
- It's a semantic issue (readability), not functional
- JSON spec doesn't guarantee object key order
- Graceful degradation improves UX

### 2. Actionable Errors Reduce Support Burden

Every error message includes how to fix it:
- Users self-service common issues
- Reduces documentation lookups
- Consistent format (✗ + →) creates predictable UX

### 3. Forward Compatibility Matters

Allowing additional fields prevents future breaking changes:
- New CLI versions can add fields
- Old CLI versions won't break on new files
- Only validate required fields

### 4. Manual Testing Validates Edge Cases

Tested 10 scenarios systematically:
- Caught all edge cases (missing fields, wrong types, ordering)
- Verified error messages are clear and actionable
- Confirmed forward compatibility works

### 5. Related Issues Surface During Implementation

Discovered phase ordering issue (task 033) while testing:
- `initStateFile()` generates wrong order
- Validation catches it as warning
- Validates that defensive validation catches real issues

## Session Metrics

**Duration:** ~1 hour  
**Lines added:** ~140 (2 new functions)  
**Test scenarios:** 10  
**Files modified:** 4  
**Tasks completed:** 1 (038)  
**Related tasks identified:** 1 (033)

## Code Quality

**Strengths:**
- Comprehensive validation coverage
- Clear, actionable error messages
- Forward compatibility support
- Well-tested (10 scenarios)
- Follows existing patterns

**Potential Improvements:**
- Could add Go unit tests
- Could extract error message constants
- Could add JSON schema validation library (overkill for current needs)
