# Add Use Case Identifiers to Business Specs

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2025-12-28  
**Source:** User Testing Report Issue #3 (2025-12-27)

## Description

Business specs create acceptance criteria with proper IDs (BUS-GREETING-001, etc.) but Use Case sections lack identifiers. Add use case ID format to Business spec template.

## Acceptance Criteria

- [x] Business spec template (`templates/specs/business.template.md`) includes use case ID pattern
- [x] Pattern format: UC ID in document title (e.g., `# UC1-GREETING: Greeting Flow`)
- [x] Business agent (`agents/smaqit.business.agent.md`) updated with use case ID format guidance
- [x] Use case IDs follow consistent format: `UC[N]-[CONCEPT]` where N is sequential number
- [x] Template placeholder: `# UC[N]-[CONCEPT]: [USE_CASE_NAME]`
- [x] File names should also contain the UC identifier

## Implementation Details

### Decision: UC ID in Title (Not Section Header)

After critical assessment, decided to place UC ID in document title rather than section header:
- **Rationale**: The use case IS the primary organizing unit, so title should reflect it
- **Format**: `# UC[N]-[CONCEPT]: [USE_CASE_NAME]`
- **File naming**: `uc1-login.md`, `uc2-checkout.md`
- **Visibility**: More prominent in file browsers and cross-references

### Changes Made

**Level 1 (Templates):**
1. `templates/specs/business.template.md`:
   - Changed title from `# [CONCEPT_NAME]` to `# UC[N]-[CONCEPT]: [USE_CASE_NAME]`
   - Added HTML comment explaining UC ID format with examples
   - Format: `UC[N]-[CONCEPT]` where N is sequential (UC1, UC2, UC3...)

**Level 2 (Agents):**
2. `agents/smaqit.business.agent.md`:
   - Added "Use Case ID Format" section before "Requirement ID Format"
   - Documented components, examples, and rules for UC IDs
   - Updated "File Organization" section with UC ID file naming
   - Updated "Directives > MUST" to include UC ID requirement
   - Updated "Completion Criteria" to verify UC ID format and consistency with requirement IDs
   - Added rule: CONCEPT in UC ID must match CONCEPT in requirement IDs

**Level 0 (Framework) & Supporting Files:**
3. `framework/ARTIFACTS.md`:
   - Updated example references: `../business/login.md` → `../business/uc1-login.md`
   - Updated all three examples (login, checkout, profile) to include UC IDs

4. `agents/smaqit.functional.agent.md`:
   - Updated example references to business specs with UC IDs
   - Maintains consistency for downstream agents referencing business specs

5. Wiki documentation (`docs/wiki/`):
   - `concepts/traceability.md`: Updated file path examples
   - `concepts/bounded-agents.md`: Updated example agent response
   - `designs/explicit-over-implicit.md`: Updated reference example

### Alignment with Existing Patterns

UC ID format mirrors requirement ID format:
- **Requirement IDs**: `[LAYER_PREFIX]-[CONCEPT]-[NNN]` (e.g., `BUS-LOGIN-001`)
- **Use Case IDs**: `UC[N]-[CONCEPT]` (e.g., `UC1-LOGIN`)
- **Shared CONCEPT**: Both use same concept descriptor for traceability
- **Stability rules**: Never reuse, deprecate instead, remain stable over time

## Impact

**Severity:** Low  
**User Impact:** Improves traceability; use cases can now be referenced by ID in downstream specs

## Validation

- [x] Template includes UC ID placeholder with clear guidance
- [x] Agent has complete UC ID documentation
- [x] File naming conventions updated
- [x] Examples throughout codebase updated for consistency
- [x] Alignment with ARTIFACTS.md traceability rules verified
- [x] Level transition workflow followed (Level 1 → Level 2)

## Notes

Low-priority enhancement for traceability. Use cases are now referenceable like acceptance criteria. This creates clearer organizational structure and improves impact analysis when business requirements change.
