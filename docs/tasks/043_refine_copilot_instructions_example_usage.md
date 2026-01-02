# Refine Copilot Instructions Regarding Example Usage

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2025-12-30

## Description

Copilot instructions contain clear rules about example usage in smaqit framework content:
- **DO**: "Prefer abstract categories over specific examples"
- **DON'T**: "Put examples or extended explanations in template/agent files"

These rules are not being consistently respected, leading to specific examples appearing in framework files, templates, and agents where they should use abstract categories and generic placeholders instead.

Need to strengthen and clarify these rules in `.github/copilot-instructions.md` to prevent example pollution in framework artifacts.

## Acceptance Criteria

- [x] Updated copilot instructions with strengthened guidance on example usage
- [x] Clear distinction between allowed examples (HTML comments in prompts) and prohibited examples (framework files, templates, agents)
- [x] Explicit rule about generic placeholders (`[PLACEHOLDER]`) vs specific examples (BUS-LOGIN-001, authentication, etc.)
- [x] Cleanup of existing specific examples in framework files, templates, and agents
- [x] Exception documented: HTML comment examples in prompts (like mario-hello test case) are allowed and intentional
- [x] Validation that framework/, templates/, and agents/ directories use only abstract categories and generic placeholders

## Implementation Summary

### Phase 1: Strengthen Copilot Instructions

Added comprehensive "Example Usage Rules" section to `.github/copilot-instructions.md`:

**Prohibited examples table:**
- Requirement IDs: Use `[LAYER_PREFIX]-[CONCEPT]-[NNN]` instead of `BUS-LOGIN-001`
- Technologies: Use `[Technology]` instead of `JWT`, `React`, `AWS`
- Features/Domains: Use `[Feature name]` instead of `login`, `authentication`, `checkout`
- Architecture: Use `[Pattern]` instead of `microservices`, `REST API`
- Data/Entities: Use `[Entity]` instead of `User`, `Order`, `Product`

**Allowed examples table:**
- Prompt files (HTML comments): `<!-- Example: "..." -->`
- Wiki docs (human explanation)
- History files (session documentation)
- Test cases (demonstration scenarios)

**Validation checklist:**
- No specific requirement IDs
- No specific technology names
- No specific business domains
- All examples use `[BRACKETS]`
- HTML comments for guidance only

### Phase 2: Clean Framework Files

**ARTIFACTS.md (6 cleanups):**
1. Requirement ID examples table → Generic format patterns
2. Traceability references → Placeholder format (`[BUS-[CONCEPT]-NNN]`)
3. Traceability matrix → Generic IDs (`BUS-[CONCEPT]-001`)
4. Coverage translation example → Placeholder Gherkin (`[CONCEPT]`, `[Feature Name]`)
5. Implementation traceability → Placeholder code comments (`[REQ-ID-001]`)
6. Validation report examples → Placeholder format (`[REQ-ID]`, `[TEST-ID]`)

### Phase 3: Clean Agent Files

**4 agents updated:**
1. `smaqit.business.agent.md` — Removed `BUS-LOGIN-001` example, added format description
2. `smaqit.functional.agent.md` — Removed `FUN-AUTH-001` example and reference examples (`BUS-LOGIN`, `BUS-CHECKOUT`, `BUS-PROFILE`)
3. `smaqit.stack.agent.md` — Removed `FUN-AUTH` and `FUN-DATA` reference examples
4. `smaqit.coverage.agent.md` — Removed `COV-LOGIN-001` example

All replaced with generic placeholders maintaining format structure.

### Phase 4: Validation

**Comprehensive validation:**
- ✅ framework/ — Clean (no specific requirement IDs or technologies)
- ✅ templates/ — Clean (no specific examples)
- ✅ agents/ — Clean (no specific examples)
- ✅ Installer builds successfully
- ✅ `smaqit init` creates clean installation
- ✅ `smaqit validate` passes
- ✅ `smaqit status` works correctly
- ✅ Installed agents have no specific examples

**Remaining allowed examples:**
- File naming patterns in ARTIFACTS.md: `login.md`, `user-login.md` (demonstrating naming conventions, not requirements)
- Placeholder format in PHASES.md: `${secrets.AWS_ACCESS_KEY}` (demonstrating Isolation Principle, not prescribing AWS)

Both are contextually appropriate and don't pollute framework with false requirements.

## Files Modified

1. `.github/copilot-instructions.md` — Added "Example Usage Rules" section (~40 lines)
2. `framework/ARTIFACTS.md` — 6 cleanups replacing specific examples with generic placeholders
3. `agents/smaqit.business.agent.md` — Requirement ID format cleanup
4. `agents/smaqit.functional.agent.md` — Requirement ID format and references cleanup
5. `agents/smaqit.stack.agent.md` — Reference examples cleanup
6. `agents/smaqit.coverage.agent.md` — Requirement ID format cleanup

**Total:** 6 files modified

## Impact

**Before:**
- Specific examples scattered across framework and agents could be misinterpreted as actual requirements
- Copy-paste risk: developers might use `BUS-LOGIN-001`, `JWT`, `authentication` as templates

**After:**
- All framework content uses generic placeholders: `[CONCEPT]`, `[Technology]`, `[Feature name]`
- Clear rules prevent future example pollution
- Framework maximally reusable across all project types
- Aligns with Level 0→1→2 compilation model (Level 0 should be abstract, not prescriptive)

## Notes

**Current problems:**
- Framework files may contain specific technology/architecture examples that should be abstract
- Templates may have leftover specific examples instead of placeholders
- Agents may contain example pollution from copy-paste refinements

**Allowed examples:**
- HTML comments in prompts: `<!-- Example: "Mario Fan - Users who love Nintendo's Mario franchise" -->`
- Test cases: `docs/test-cases/mario-hello.md` as demonstration scenario
- Session history: Examples in history files for documentation purposes

**Prohibited examples:**
- Specific requirement IDs in framework/template/agent files (except when demonstrating format structure)
- Technology-specific examples (JWT, React, AWS, etc.) in layer definitions
- Architecture patterns (microservices, containers, etc.) in abstract guidance
- Business domain examples (login, checkout, etc.) in reusable templates

The goal is to ensure framework content remains maximally reusable across all project types without prescribing specific solutions.
