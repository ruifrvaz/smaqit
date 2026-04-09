# Task 043 Complete: Refine Copilot Instructions Example Usage

**Date:** 2025-12-30  
**Session Type:** Framework refinement and cleanup  
**Tasks Completed:** 043

## Overview

Strengthened example usage guidance in copilot instructions and cleaned all specific examples from framework files and agents. Replaced specific examples (BUS-LOGIN-001, FUN-AUTH-001, JWT, authentication, login) with generic placeholders to maximize framework reusability and prevent example pollution.

## What Was Done

### Phase 1: Strengthen Copilot Instructions

**Added "Example Usage Rules" section to `.github/copilot-instructions.md`:**

Created comprehensive guidance with three main components:

1. **Prohibited Examples Table**

| Category | Prohibited | Use Instead |
|----------|-----------|-------------|
| Requirement IDs | BUS-LOGIN-001, FUN-AUTH-001 | [LAYER_PREFIX]-[CONCEPT]-[NNN] |
| Technologies | JWT, React, AWS, PostgreSQL | [Technology], [Framework] |
| Features/Domains | login, authentication, checkout | [Feature name], [Concept] |
| Architecture | microservices, REST API | [Pattern], [Architecture style] |
| Data/Entities | User, Order, Product | [Entity], [Data model] |

2. **Allowed Examples Table**

| Location | Context | Format | Example |
|----------|---------|--------|---------|
| Prompt files | User guidance | HTML comments | `<!-- Example: "..." -->` |
| Wiki docs | Human explanation | Plain text | Documentation |
| History files | Session docs | Plain text | This file |
| Test cases | Demonstrations | Plain text | mario-hello.md |

3. **Validation Checklist**

Before committing framework/templates/agents:
- No specific requirement IDs
- No specific technology names
- No specific business domains
- All examples use [BRACKETS]
- HTML comments for guidance only

**Rationale:** Previous guidance ("prefer abstract categories") was too vague. Specific examples were still appearing because developers lacked clear rules about what constitutes example pollution and where it's prohibited.

### Phase 2: Clean Framework Files

**ARTIFACTS.md (6 cleanups):**

1. **Requirement ID examples table** (lines 32-40)
   - Before: `BUS-LOGIN-001 | User can authenticate with valid credentials`
   - After: `BUS-[CONCEPT]-001 | [Use case or actor goal description]`

2. **Traceability references** (lines 127-134)
   - Before: `[BUS-LOGIN](../business/uc1-login.md) — Implements login use case`
   - After: `[BUS-[CONCEPT]-NNN](../business/[filename].md) — Implements [use case description]`

3. **Traceability matrix** (line 163)
   - Before: `BUS-LOGIN-001 | FUN-AUTH-001 | STK-JWT-001`
   - After: `BUS-[CONCEPT]-001 | FUN-[CONCEPT]-001 | STK-[CONCEPT]-001`

4. **Coverage translation example** (lines 171-183)
   - Before: `FUN-AUTH-001: User receives JWT token upon successful login`
   - After: `FUN-[CONCEPT]-001: [Behavior description]`
   - Before Gherkin: `Successful login returns JWT token`
   - After Gherkin: `[Scenario description]` with `[precondition]`, `[action]`, `[expected outcome]`

5. **Implementation traceability** (lines 274-279)
   - Before: `Authenticates user and returns JWT token. Implements: FUN-AUTH-001`
   - After: `[Method description]. Implements: [REQ-ID-001]`

6. **Validation report examples** (lines 350-357)
   - Before: `BUS-UX-002`, `COV-AUTH-005 | FUN-AUTH-003`
   - After: `[REQ-ID]`, `[TEST-ID] | [REQ-ID]`

**Result:** ARTIFACTS.md now provides format structure without prescribing specific requirements or technologies.

### Phase 3: Clean Agent Files

**4 agents updated:**

1. **smaqit.business.agent.md** (lines 130-141)
   - Removed: `**Example:** BUS-LOGIN-001: User can authenticate with valid credentials`
   - Added: `**Format:** BUS-[CONCEPT]-[NNN]: [Use case or actor goal description]`
   - Removed specific CONCEPT examples (LOGIN, CHECKOUT, USER-REGISTRATION)
   - Added: "uppercase with hyphens" description

2. **smaqit.functional.agent.md** (2 cleanups)
   - Lines 112-123: Removed `FUN-AUTH-001: JWT token expires after 24 hours` example
   - Added: `FUN-[CONCEPT]-[NNN]: [Behavior or data model requirement]` format
   - Lines 170-179: Removed reference examples (`BUS-LOGIN`, `BUS-CHECKOUT`, `BUS-PROFILE`)
   - Added: Generic placeholder format for Implements/Enables references

3. **smaqit.stack.agent.md** (lines 148-155)
   - Removed: `[FUN-AUTH](../functional/authentication.md) — Framework supports auth patterns`
   - Removed: `[FUN-DATA](../functional/data-model.md) — ORM supports data relationships`
   - Added: `[FUN-[CONCEPT]-NNN](../functional/[filename].md) — Framework supports [capability]`

4. **smaqit.coverage.agent.md** (lines 100-111)
   - Removed: `COV-LOGIN-001: Test case for BUS-LOGIN-001`
   - Added: `COV-[CONCEPT]-[NNN]: Test case for [upstream requirement ID]`
   - Removed specific CONCEPT examples (LOGIN, AUTH, API)

**Result:** All agents now demonstrate format structure with generic placeholders, not specific examples that could be misinterpreted as actual requirements.

### Phase 4: Validation

**Comprehensive testing:**

1. **Example cleanup verification**
   ```bash
   grep -rn "BUS-LOGIN|FUN-AUTH|COV-LOGIN|STK-JWT" framework/ templates/ agents/
   # Result: Clean (no specific requirement IDs found)
   ```

2. **Installer build**
   ```bash
   cd installer && make build
   # Result: Success (Built: dist/smaqit)
   ```

3. **Installation test**
   ```bash
   mkdir test-clean && cd test-clean
   ../dist/smaqit init
   # Result: Success (all files copied, agents installed)
   ```

4. **CLI validation**
   ```bash
   ../dist/smaqit validate
   # Result: ✓ Validation passed
   ../dist/smaqit status
   # Result: Correct status display
   ```

5. **Installed agents verification**
   ```bash
   grep -rn "BUS-LOGIN|FUN-AUTH|COV-LOGIN" .github/agents/
   # Result: ✓ Installed agents clean
   ```

**Remaining allowed examples:**
- `framework/ARTIFACTS.md`: File naming examples (`login.md`, `user-login.md`) — Contextually appropriate (demonstrating naming conventions, not requirements)
- `framework/PHASES.md`: `${secrets.AWS_ACCESS_KEY}` — Demonstrating Isolation Principle placeholder format, not prescribing AWS

Both are format demonstrations with clear context, not prescriptive examples.

## Decisions Made

### 1. Explicit Tables vs Vague Guidance

**Decision:** Create explicit prohibited/allowed examples tables instead of general "prefer abstract categories" guidance.

**Rationale:**
- Previous guidance was too vague
- Developers needed concrete examples of what to avoid
- Tables provide quick reference during development
- Clear prohibited list prevents ambiguity

**Alternative rejected:** Keep existing vague guidance and rely on code review to catch violations. This failed in practice (specific examples still appeared).

### 2. Replace All Specific Examples vs Keep Some for Clarity

**Decision:** Replace ALL specific examples in framework/agents with generic placeholders, even if contextually clear.

**Rationale:**
- Maximizes framework reusability across all project types
- Prevents copy-paste contamination (developers using examples as templates)
- Aligns with Level 0→1→2 compilation model (Level 0 should be abstract)
- Consistent enforcement easier than case-by-case judgment

**Alternative rejected:** Keep some specific examples when context makes it clear they're demonstrations. Risk: developers might not always recognize the context, leading to contamination.

### 3. Format Demonstration vs Placeholder Format

**Decision:** Change from `**Example:** [specific]` to `**Format:** [placeholder]` in agents.

**Rationale:**
- "Format" signals structure demonstration, not example to copy
- Placeholder format (`[BRACKETS]`) clearly indicates "fill this in"
- Reduces ambiguity about whether example should be copied
- Matches established convention from framework files

**Alternative rejected:** Keep "Example:" label but use placeholders. Still ambiguous whether it's an example to emulate vs a format to follow.

### 4. Validation Checklist Placement

**Decision:** Add validation checklist directly in copilot-instructions.md, not in separate document.

**Rationale:**
- Developers read copilot-instructions.md during work
- Checklist integrated where it's needed
- Reduces navigation overhead
- Higher visibility = better compliance

**Alternative rejected:** Create separate validation document. Lower visibility, developers might not consult it.

## Problems Solved

### Problem 1: Example Pollution in Framework Files

**Symptom:** Specific examples (BUS-LOGIN-001, JWT, authentication) appeared in framework files despite cleanup tasks 027 and 028.

**Root Cause:** Copilot instructions had vague guidance ("prefer abstract categories") without explicit rules or examples of violations.

**Solution:** Created explicit prohibited/allowed examples tables with concrete cases.

**Impact:** Clear rules prevent future contamination. Framework now maximally reusable.

### Problem 2: Copy-Paste Risk from Examples

**Symptom:** Developers might copy specific examples (BUS-LOGIN-001) as templates when creating new specs.

**Root Cause:** Examples looked like templates to follow rather than format demonstrations.

**Solution:** Replaced all specific examples with generic placeholders using [BRACKETS] notation.

**Impact:** Impossible to copy-paste specific examples. Developers forced to create their own content.

### Problem 3: Ambiguous "Example" Label

**Symptom:** Unclear whether `**Example:** BUS-LOGIN-001` is something to copy or just a format demonstration.

**Root Cause:** "Example" label suggests concrete instance, not abstract pattern.

**Solution:** Changed to `**Format:**` label with placeholder content.

**Impact:** Clear signal: this is structure to follow, not content to copy.

## Files Modified

**Implementation (6 files):**
1. `.github/copilot-instructions.md` — Added "Example Usage Rules" section (~40 lines)
2. `framework/ARTIFACTS.md` — 6 cleanups replacing specific examples with placeholders
3. `agents/smaqit.business.agent.md` — Requirement ID format cleanup
4. `agents/smaqit.functional.agent.md` — Requirement ID and references cleanup
5. `agents/smaqit.stack.agent.md` — Reference examples cleanup
6. `agents/smaqit.coverage.agent.md` — Requirement ID format cleanup

**Documentation (2 files):**
7. `docs/tasks/043_refine_copilot_instructions_example_usage.md` — Marked completed, added implementation summary
8. `docs/tasks/PLANNING.md` — Moved task 043 to Completed

**History (1 file):**
9. `docs/history/017_task_043_refine_example_usage_2025-12-30.md` — This file

**Total:** 9 files modified/created

## Next Steps

### Immediate

1. **Monitor compliance** — Check future PRs for example pollution using new validation checklist
2. **Update wiki if needed** — Consider adding example usage article to `docs/wiki/patterns/`

### Suggested

1. **Task 044:** Clean up old prompt name references
   - Next logical cleanup task
   - Completes consistency across codebase

2. **Task 042:** Move development phase report to .smaqit/reports
   - Simple organizational improvement
   - Keeps project root clean

3. **Task 041:** Restrict agents to their layer/phase
   - Related to framework integrity
   - Prevents agents from overstepping boundaries

## Key Learnings

### 1. Vague Rules Lead to Inconsistent Application

**Before:** "Prefer abstract categories over specific examples"  
**Problem:** Too open to interpretation, specific examples still appeared  
**After:** Explicit tables of prohibited/allowed examples  
**Learning:** Concrete examples of violations are more effective than abstract principles

### 2. Examples Are Templates by Default

When developers see `BUS-LOGIN-001: User can authenticate`, they treat it as:
- Template to copy (wrong interpretation)
- Format to follow (intended interpretation)

Ambiguity leads to contamination. Solution: Use `[PLACEHOLDER]` format that cannot be copy-pasted.

### 3. Label Choice Matters

- `**Example:**` suggests concrete instance
- `**Format:**` suggests abstract pattern
- Small word choice has big impact on developer behavior

### 4. Validation Checklist Needs High Visibility

Checklists in separate docs get ignored. Checklist in copilot-instructions.md gets used. Put validation where developers work, not in reference docs.

### 5. Progressive Cleanup is Necessary

Tasks 027, 028, and now 043 all addressed example pollution:
- 027: Separated instructions from rationale
- 028: Cleaned meta-rationale from all levels
- 043: Cleaned specific examples, added explicit rules

Each iteration caught issues missed by previous passes. Comprehensive cleanup requires multiple focused passes.

## Session Metrics

**Duration:** ~2 hours (session recap → implementation → validation → documentation)  
**Tasks completed:** 1 (043)  
**Files modified:** 9  
**Cleanups performed:** 12 (6 in ARTIFACTS.md, 6 in agents)  
**Lines added:** ~60 (copilot-instructions.md section)  
**Lines modified:** ~30 (framework + agents)  

**Quality indicators:**
- ✅ Installer builds successfully
- ✅ All CLI commands tested and working
- ✅ Comprehensive validation (framework, templates, agents)
- ✅ Installation test passed
- ✅ Documentation complete and detailed

## Code Quality

**Strengths:**
- Explicit prohibited/allowed examples tables
- Clear validation checklist
- Comprehensive cleanup across all levels
- Thorough testing (grep, build, install, CLI)
- Well-documented decisions and rationale

**Potential improvements:**
- Could add automated linter to check for specific examples in PRs
- Could create pre-commit hook to run validation checklist
- Could add more examples to prohibited list as new patterns emerge

## Reference

This session completes the progression of framework cleanup tasks:
- Task 027: Separate framework instructions from human rationale
- Task 028: Audit all levels for meta-rationale
- **Task 043: Refine copilot instructions example usage** (this task)

Together, these ensure framework content is purely instructional, abstract, and maximally reusable without prescribing specific technologies or requirements.
