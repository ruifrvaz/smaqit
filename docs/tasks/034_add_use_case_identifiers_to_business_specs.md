# Add Use Case Identifiers to Business Specs

**Status:** Not Started  
**Created:** 2025-12-28  
**Source:** User Testing Report Issue #3 (2025-12-27)

## Description

Business specs create acceptance criteria with proper IDs (BUS-GREETING-001, etc.) but Use Case sections lack identifiers. Add use case ID format to Business spec template.

## Acceptance Criteria

- [ ] Business spec template (`templates/specs/business.template.md`) includes use case ID pattern
- [ ] Pattern format: `## Use Case: [UC_ID]` (e.g., `## Use Case: UC1-GREETING`)
- [ ] Business agent (`agents/smaqit.business.agent.md`) updated with use case ID format guidance
- [ ] Use case IDs follow consistent format: `UC[N]-[CONCEPT]` where N is sequential number
- [ ] Template placeholder: `## Use Case: UC[N]-[CONCEPT]`
- [ ] File names should also contain the UC identifier

## Impact

**Severity:** Low  
**User Impact:** Reduces traceability; use cases cannot be easily referenced by ID in downstream specs

## Notes

Low-priority enhancement for traceability. Use cases should be referenceable like acceptance criteria.
