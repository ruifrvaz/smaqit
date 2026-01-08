# Session 031: Task 050 - Coverage Layer Clarification (Final)

**Date:** 2026-01-06  
**Task:** 050 - Redesign Coverage Prompt  
**Status:** Completed  
**Previous Attempt:** Session 027 (2026-01-05) - Overly restrictive approach reverted

---

## Objective

Clarify Coverage layer's dual input model: test requirements from prompt + upstream acceptance criteria to verify. This session corrected the first attempt which made Coverage prompt too restrictive ("preferences only").

---

## Problem with First Attempt (Session 027)

**What happened:** Session 027 redesigned Coverage as "pure traceability mapping" with prompt providing only "verification preferences" (tooling, thresholds).

**Why it was wrong:** Coverage layer DOES have its own requirements—just not *functional* requirements. Coverage needs:
- Test requirements (how to test, test strategy, test environment)
- NOT functional requirements (what system does—that comes from upstream)

**The error:** Tried to make Coverage prompt "preferences only" when Coverage actually needs test requirements like:
- Test Scope (integration, E2E types)
- Test Environment (where/how tests run)
- Integration Points (external systems to verify)
- Acceptance Thresholds (pass criteria)

---

## Corrected Architecture

**Coverage has TWO distinct input sources:**

1. **From Prompt (test requirements):** HOW to test
   - Test scope (integration, E2E, acceptance types needed)
   - Test environment (where/how tests run)
   - Integration points (external systems to verify)
   - Acceptance thresholds (coverage goals, pass criteria)

2. **From Upstream Specs (acceptance criteria to verify):** WHAT to test
   - All acceptance criteria from Business, Functional, Stack, Infrastructure layers

**Coverage agent's job:**
- Scan ALL upstream specs for acceptance criteria
- Apply test requirements from prompt to define HOW to verify each criterion
- Generate test cases mapping: `Upstream Requirement ID → Test Case → Expected Outcome`
- Report coverage %

---

## Implementation

### Level 0: Framework (LAYERS.md)

**Updated Coverage Input/Context:**

```markdown
**Input:** User test requirements (test scope, test environment, integration points, acceptance thresholds)

**Context:** All layer specs (Business, Functional, Stack, Infrastructure) — source of upstream acceptance criteria to verify
```

**Updated Coverage MUST directives:**

```markdown
**Coverage specs MUST:**
- Reference every acceptance criterion from upstream specs by ID
- Define a test case for each testable requirement using test requirements from prompt
- Map: Requirement ID → Test Case → Expected Outcome
- Flag untestable requirements explicitly
- Include integration, E2E, and acceptance test definitions per prompt requirements
- Report spec coverage (% of requirements with corresponding tests)

**Coverage specs MUST NOT:**
- Add acceptance criteria not present in upstream specs
- Skip upstream acceptance criteria without justification
- Modify or reinterpret upstream acceptance criteria
- Define unit tests (those are implementation details)
```

**Key change:** "Add functional requirements" → "Add acceptance criteria" (more precise)

### Level 1: Templates

**Coverage prompt sections (kept from user's simplification):**
- Test Scope
- Test Environment
- Integration Points
- Acceptance Thresholds

**Removed sections:** Performance Benchmarks, Security Requirements (these ARE functional requirements, belong in Business/Infrastructure layers)

### Level 2: Agent (agents/smaqit.coverage.agent.md)

**Updated Role:**

```markdown
Specification agent for the Coverage layer. Enumerates all acceptance criteria from upstream 
specifications and maps each to test cases. Uses prompt file for test requirements (how to test) 
and upstream specs for acceptance criteria to verify (what to test).
```

**Updated Input section:**

```markdown
**Prompt File:** `.github/prompts/smaqit.coverage.prompt.md`

- Read test requirements from prompt file (test scope, environment, integration points, thresholds)
- Ignore all HTML comments to prevent example pollution
- Interpret free-style natural language without rigid structure enforcement
- Validate sufficiency - if content insufficient, request clarification with natural language guidance

**User Input:**
- Test scope (integration, E2E, acceptance types needed)
- Test environment (where/how tests run)
- Integration points (external systems to verify)
- Acceptance thresholds (coverage goals, pass criteria)

**Upstream Specifications (source of acceptance criteria to verify):**
- specs/business/ — Use cases, actors, business goals
- specs/functional/ — Behaviors, contracts, data models
- specs/stack/ — Technology choices, runtime requirements
- specs/infrastructure/ — Deployment topology, scaling policies

**Conflict Resolution:**
When user input conflicts with upstream specs, flag the conflict rather than silently override.
```

**Updated MUST directives:**

```markdown
- Scan ALL upstream specs and map every upstream acceptance criterion by ID to a test case
- Define test case for each testable criterion using test requirements from prompt
- Map format: Upstream Requirement ID → Test Case → Expected Outcome
- Flag untestable upstream acceptance criteria explicitly
- Include integration, E2E, and acceptance test definitions per prompt test requirements
- Report spec coverage (% of upstream acceptance criteria with corresponding tests)
- Calculate coverage: (mapped criteria / total testable criteria) × 100%
```

**Key refinements:**
- Combined "scan" (process) with "map" (output requirement)
- Changed "functional requirements" → "upstream acceptance criteria" (avoids confusion with Functional layer)
- Clarified dual sources explicitly

### Level 2: Prompts

**Coverage prompt (prompts/smaqit.coverage.prompt.md):**
- Kept simplified structure (4 sections: Test Scope, Test Environment, Integration Points, Acceptance Thresholds)
- Description emphasizes test requirements, not functional requirements

**Validation prompt (prompts/smaqit.validation.prompt.md):**
- Renamed "Test Scope" → "Execution Scope" to eliminate ambiguity with Coverage prompt
- Now clear: Coverage = test strategy, Validation = execution preferences

---

## Key Decisions

### 1. Coverage Layer Purpose

**Decision:** Coverage captures test requirements (how to test) while reading upstream acceptance criteria (what to test)

**Alternatives Rejected:**
- Pure preferences only (Session 027) — Too restrictive, Coverage has legitimate test requirements
- Hybrid requirements-adding layer — Would duplicate functional requirements, violate layer independence

**Rationale:**
- Test requirements ARE legitimate Coverage layer concerns (test environment, integration points, etc.)
- Functional requirements come from upstream specs, not Coverage prompt
- This maintains Layer Independence while acknowledging Coverage's unique dual-input nature

### 2. Terminology: "Functional Requirements" → "Upstream Acceptance Criteria"

**Decision:** Use "upstream acceptance criteria" not "functional requirements" when referring to what Coverage tests

**Rationale:**
- "Functional requirements" is ambiguous — could mean Functional layer specifically or functionality generally
- "Upstream acceptance criteria" is precise — all testable criteria from Business, Functional, Stack, Infrastructure layers
- Avoids confusion with Functional layer

### 3. Template Structure vs Coverage Uniqueness

**Decision:** Coverage agent retains Prompt File section despite template using simpler "User Input" structure

**Rationale:**
- Coverage is unique in having dual inputs (prompt + upstream specs)
- Explicit Prompt File section clarifies prompt handling rules (HTML comments, validation)
- Template simplicity is good for typical layers, Coverage needs additional clarity

### 4. Validation Prompt Disambiguation

**Decision:** Rename Validation prompt "Test Scope" → "Execution Scope"

**Rationale:**
- Both Coverage and Validation had "Test Scope" with different meanings
- Coverage: What types of tests to define (strategy)
- Validation: Which tests to run (execution preference)
- Renaming eliminates ambiguity

---

## Files Modified

**Framework (Level 0):**
- `framework/LAYERS.md` - Updated Coverage Input/Context/MUST directives
- `framework/PROMPTS.md` - Updated Validation prompt description
- `framework/ARTIFACTS.md` - Clarified Coverage traceability description

**Templates (Level 1):**
- `templates/prompts/specification-prompt.template.md` - Already simplified by user
- `templates/prompts/implementation-prompt.template.md` - Updated Validation section

**Agents/Prompts (Level 2):**
- `prompts/smaqit.coverage.prompt.md` - Already simplified by user (4 sections)
- `prompts/smaqit.validation.prompt.md` - Renamed "Test Scope" → "Execution Scope"
- `agents/smaqit.coverage.agent.md` - Updated Role, Input, MUST directives with clear terminology

**Total:** 7 files modified (3 framework, 2 templates, 2 agents/prompts)

---

## Testing and Validation

**Build Validation:**
```bash
cd installer && make build
# Result: ✅ Build successful
```

**Terminology Validation:**
```bash
grep -r "functional requirements.*Coverage" framework/ agents/
# Result: No matches (ambiguous terminology removed)
```

**Consistency Check:**
- ✅ Framework establishes "test requirements" vs "upstream acceptance criteria" distinction
- ✅ Templates reflect simplified structure
- ✅ Agent aligned with framework terminology
- ✅ Prompt disambiguation (Coverage vs Validation) complete

---

## Impact Analysis

### User Experience
- **Before:** Confusion about what Coverage prompt should contain
- **After:** Clear - Coverage prompt has test requirements (how to test), upstream specs have what to test
- **Improvement:** Users understand Coverage's role without duplication

### Agent Behavior
- **Before:** Ambiguous whether Coverage uses prompt or upstream for requirements
- **After:** Explicit dual-input model - test requirements from prompt, acceptance criteria from upstream
- **Improvement:** Agent instructions precise and actionable

### Framework Integrity
- **Before:** "Functional requirements" terminology created confusion with Functional layer
- **After:** "Upstream acceptance criteria" terminology is precise
- **Improvement:** No layer name conflicts, clear terminology

### Validation Distinction
- **Before:** Both Coverage and Validation had "Test Scope" sections
- **After:** Coverage has "Test Scope" (strategy), Validation has "Execution Scope" (preferences)
- **Improvement:** No prompt ambiguity

---

## Lessons Learned

### 1. Over-Correction Risk

**Pattern:** When fixing a problem (Coverage adding requirements), first attempt went too far (preferences only).

**This Session:** Coverage DOES have requirements—just not functional requirements. Test requirements (how to test) are legitimate Coverage layer concerns.

**Takeaway:** When simplifying, distinguish between removing wrong content vs removing necessary content. Test requirements ≠ functional requirements.

### 2. Terminology Precision Matters

**Pattern:** "Functional requirements" is ambiguous in context of layers named Business, Functional, Stack, etc.

**This Session:** "Upstream acceptance criteria to verify" is precise and unambiguous.

**Takeaway:** When framework has entities named similar to general terms (Functional layer vs functional requirements), use more specific terminology.

### 3. Template Simplicity vs Layer Uniqueness

**Pattern:** Templates provide structure, but some layers need additional clarity.

**This Session:** Coverage kept Prompt File section despite template simplification because dual-input model needs explicit documentation.

**Takeaway:** Templates are guidelines, not rigid constraints. Justify deviations with layer-specific needs.

### 4. Critical Assessment Prevents Over-Simplification

**Pattern:** User correctly questioned first attempt's approach.

**This Session:** "What is the goal of the Coverage agent?" forced re-evaluation of overly restrictive design.

**Takeaway:** Always validate changes against actual layer purpose before committing. Critical assessment catches over-corrections.

---

## Related Tasks

- **Task 048** (Completed) - E2E Agent Workflow Testing (identified Coverage prompt contradiction)
- **Task 049-053** (Active) - Remaining release blockers

---

## Next Steps

With Task 050 finally complete (correct approach), remaining work:
1. Tasks 049, 051, 052 - Implementation agent CLI directives
2. Task 053 - Validation frontmatter updates
3. v0.5.0 release

---

## Session Metrics

**Duration:** ~1.5 hours (including critical assessment and refinement)

**Tasks Completed:** 1 (Task 050 - corrected approach)

**Files Modified:** 7 (3 framework + 2 templates + 2 agents/prompts)

**Key Outcomes:**
- Coverage layer dual-input model clarified
- Terminology disambiguation ("upstream acceptance criteria" not "functional requirements")
- Validation prompt disambiguation ("Execution Scope" not "Test Scope")
- Framework integrity maintained with precise layer boundaries

**Previous Attempt:** Session 027 (2026-01-05) - "Preferences only" approach was too restrictive

**This Attempt:** Balanced approach - Coverage has test requirements, reads upstream acceptance criteria
