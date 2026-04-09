# Add state.json Validation to Validate Command

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2025-12-28  
**Source:** User Testing Report Issue #7 (2025-12-27)

## Description

`smaqit validate` command checks directory structure and spec files, but does not validate `.smaqit/state.json`. Add validation to catch structural issues in state file.

## Acceptance Criteria

- [x] Validate command checks `.smaqit/state.json` file exists
- [x] Validates JSON structure is valid (parseable)
- [x] Verifies phase keys are present (develop, deploy, validate)
- [x] Verifies phase keys are correctly ordered in JSON
- [x] Validates each phase object has required "completed" boolean field
- [x] Validates "version" field is present at root level
- [x] Reports specific validation errors with actionable messages
- [x] Does not fail on additional fields (forward compatibility)

## Impact

**Severity:** Medium  
**User Impact:** Corrupted or malformed state.json can cause status command failures; validate should catch structural issues

## Implementation

Added comprehensive state.json validation to `smaqit validate` command in `installer/main.go`:

**New Functions:**
- `validateStateFile()` - Main validation function (138 lines)
  - File existence check with recovery guidance
  - JSON parseability with syntax error details
  - Version field validation (presence, type, non-empty)
  - Phases object validation (presence, type)
  - Phase keys validation (develop, deploy, validate)
  - Phase structure validation (completed boolean field)
  - Forward compatibility (allows additional fields)
- `validatePhaseOrdering()` - Semantic ordering check
  - Warns if phases not in workflow order
  - Warning only (not an error - functionality unaffected)

**Error Messages:**
All validation errors include actionable guidance prefixed with →:
- File missing: "→ Run 'smaqit init' to create state.json"
- Invalid JSON: "→ Fix JSON syntax or regenerate with 'smaqit init' (after backup)"
- Missing field: "→ Add: \"field\": value"
- Wrong type: "→ Change to: expected type"

**Testing:**
Manually tested 10 scenarios:
1. ✅ Valid state.json (passes)
2. ✅ Missing file (error with guidance)
3. ✅ Invalid JSON syntax (error with parse details)
4. ✅ Missing version field (error with fix)
5. ✅ Missing phases object (error with guidance)
6. ✅ Missing phase keys (error per missing phase)
7. ✅ Missing completed field (error per phase)
8. ✅ Wrong type for completed (error with type guidance)
9. ✅ Additional fields (passes - forward compatible)
10. ✅ Wrong phase ordering (warning only)

## Notes

Defensive validation. Helps users catch state file corruption early. Phase ordering check implemented as warning (semantic issue) rather than error (functional issue), aligning with graceful degradation principle.

Related: Task #033 will fix the phase ordering issue in `initStateFile()`.
