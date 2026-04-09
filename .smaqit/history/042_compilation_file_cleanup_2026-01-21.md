# Compilation File Cleanup

**Date:** 2026-01-21  
**Session Focus:** Remove L0 Source citations from L1 Directive Compilation sections, update documentation, complete Task 068 L0/L1 work  
**Tasks Referenced:** Task 068 (Remove System Actor from Business Layer)

## Session Overview

This session completed the L1 compilation file cleanup by removing redundant L0 Source citations from directive sections across all 8 compilation files, updated documentation to reflect the changes, and completed the L0/L1 work for Task 068 (System Actor removal).

## Actions Taken

### 1. L0 Source Citation Removal (Phase Files)

Removed L0 Source citations from the remaining 3 phase compilation files:
- `templates/agents/compiled/develop.rules.md`
- `templates/agents/compiled/deploy.rules.md`
- `templates/agents/compiled/validate.rules.md`

**Pattern:** Removed `**L0 Source:** PHASES.md § X` lines and `**Compile to [PLACEHOLDER]:**` wrapper phrases, keeping only pure directives.

**Result:** All 8 compilation files (5 layers + 3 phases) now have clean L1 Directive Compilation sections without inline L0 Source citations. Source traceability preserved in Source L0 Principles table at top of each file.

### 2. Documentation Updates

Updated 4 files to reflect the L0 Source citation removal:

**templates/agents/compiled/validate.rules.md:**
- Removed outdated instruction: "Preserve traceability: Keep L0 source citations as HTML comments for debugging"

**.github/copilot-instructions.md:**
- Updated L1 Directive Compilation description to: "Pure directives without L0 Source citations (transformation already documented in table above)"

**.github/agents/smaqit.L1.agent.md:**
- Updated Compilation File Characteristics: "Pure L1 Directive Compilation section (no L0 Source citations within)"
- Updated Completion Criteria: "L1 Directive Compilation contains pure directives (no L0 Source citations)"

**docs/wiki/designs/level-up-compilation.md:**
- Updated example to show current clean format with direct MUST directives instead of "**Compile to MUST directives:**" wrapper

### 3. Task 068 L1 Work - Business Layer Actor Concept

**Critical Assessment:** Identified that `business.rules.md` contained Actor Concept section with no corresponding placeholder in `specification-agent.template.md`, breaking L1→L2 compilation.

**Resolution:** Removed Actor Concept section from business.rules.md per user direction. Rationale:
- Actor concept is universally understood
- Spec template already has actor guidance comments
- Compilation file should contain directives, not conceptual documentation
- Directive "Capture actor diversity: interactive participants AND non-functional requirement stakeholders" is sufficient

### 4. Task 068 L1 Work - NFR Terminology

**Critical Assessment:** User identified that "system property advocates" terminology obscured the fact that non-functional requirements belong in Business layer.

**Resolution:** Added explicit NFR guidance section to business.rules.md and changed directive to "non-functional requirement stakeholders" for clarity.

### 5. Task 068 L2 Compilation - Business Agent

**Critical Assessment:** Before compilation, verified that:
- All 5 layer compilation files exist (created 2026-01-19)
- Only Business agent needed recompilation (other 4 agents not affected by System Actor removal)
- Single-agent compilation was appropriate scope

**Compilation executed:** Applied `business.rules.md` directives to `agents/smaqit.business.agent.md`:
- Removed System Actor section (lines 113-121)
- Inserted strengthened MUST NOT directives
- Inserted NFR stakeholder diversity directive
- Inserted layer-specific MUST directives

### 6. Behavioral Verbs Directive Refinement

**Critical Assessment:** User identified contradiction:
- MUST NOT directive said: "Use behavioral verbs (display, render, output...)"
- Agent's own example used: "Error message **is displayed**" (marked as good example)

**Analysis:** The issue isn't specific words, but describing HOW (behaviors/mechanisms) vs WHAT (observable outcomes).
- ❌ "System displays colorful console output" (describes mechanism)
- ✅ "Error message is displayed" (describes observable outcome)

**Resolution (Option A):** Changed directive from word blacklist to concept boundary:
- From: "Use behavioral verbs (display, render, output, execute, process, detect, handle, parse, format)"
- To: "Describe system behaviors or mechanisms (how features work internally)"

**Further Refinement:** User identified redundancy between two directives:
1. "Describe system behaviors or mechanisms (how features work internally)"
2. "Describe HOW features work (behaviors belong in Functional layer)"

**Resolution (Option B - Merge):** Combined into single clear directive:
- "Describe HOW features work (behaviors and mechanisms belong in Functional layer)"

Applied across all 3 levels: L0 framework, L1 compilation, L2 agent.

### 7. Test Case Decision

**Assessment:** mario-hello.md Business Layer Input contains deliberate scope violations (Client Organizations actor, "sees Mario greeting with character representation").

**User Decision:** Keep the incorrect instructions as a "curve ball" to test whether Business Agent respects boundaries when given requirements that leak into Functional/Stack layers. This provides validation that the boundary enforcement actually works in practice.

## Problems Solved

### Problem 1: Outdated Instruction in validate.rules.md

**Issue:** Compilation Guidance contained "Preserve traceability: Keep L0 source citations as HTML comments for debugging" which contradicted the L0 Source citation removal.

**Solution:** Removed the outdated instruction.

### Problem 2: Non-existent Placeholder Referenced

**Issue:** business.rules.md Compilation Guidance step 10 said "Insert Actor guidance" but specification-agent.template.md had no `[ACTOR_GUIDANCE]` placeholder, breaking L1→L2 compilation.

**Solution:** Removed Actor Concept section and corresponding compilation step. Agent sees actor examples in spec template, and directive ensures actor diversity captured.

### Problem 3: NFR Guidance Obscurity

**Issue:** "System property advocates" terminology didn't clearly communicate that non-functional requirements belong in Business layer.

**Solution:** Changed to "non-functional requirement stakeholders" for clarity.

### Problem 4: Behavioral Verbs Directive Too Broad

**Issue:** Word blacklist approach (banning "display", "render", etc.) was too broad and contradicted agent's own examples. The real issue is describing HOW (mechanisms) not the words themselves.

**Solution:** Changed to concept-based directive focusing on the actual boundary violation: describing how features work internally.

### Problem 5: Redundant Directives

**Issue:** Two directives essentially said the same thing:
- "Describe system behaviors or mechanisms (how features work internally)"
- "Describe HOW features work (behaviors belong in Functional layer)"

**Solution:** Merged into single directive emphasizing "HOW" with layer guidance: "Describe HOW features work (behaviors and mechanisms belong in Functional layer)".

## Decisions Made

### Decision 1: Remove Actor Concept from business.rules.md (Option A)

**Options:**
- A: Remove Actor Concept section (CHOSEN)
- B: Add `[ACTOR_GUIDANCE]` placeholder to template
- C: Move to `[LAYER_SPECIFIC_PATTERNS]`

**Rationale:** Actor concept is universally understood. Compilation file should contain directives, not conceptual explanations. Template already has actor guidance comments. No template change needed.

### Decision 2: Add NFR Guidance Section

**Options:**
- A: Keep current state (implicit guidance)
- B: Add explicit NFR guidance section (CHOSEN)
- C: Add to MUST directives

**Rationale:** Makes it explicit that NFRs belong in Business layer and should be expressed through named stakeholder actors (not generic "System" actor) with measurable criteria.

### Decision 3: Behavioral Verbs - Concept Boundary (Option A)

**Options:**
- A: Clarify directive to focus on behavior descriptions (CHOSEN initially)
- B: Keep word list with clarification about passive voice
- C: Remove problematic example from agent

**Rationale:** Real boundary violation is describing HOW things work, not specific words. Focus on concept, not vocabulary.

### Decision 4: Merge Redundant Directives (Option B)

**Options:**
- A: Remove old directive
- B: Merge both into single directive (CHOSEN)
- C: Keep both with different emphasis

**Rationale:** Single clear directive is more concise. Emphasizes "HOW" (clearest boundary word) while mentioning what it covers (behaviors and mechanisms) and where they belong (Functional layer).

### Decision 5: Business Agent Only Compilation

**Options:**
- A: Compile only Business agent (CHOSEN)
- B: Compile all 5 specification agents

**Rationale:** Task 068 scope is Business layer System Actor removal. Single-agent compilation allows focused verification. Other 4 agents not affected by this change. Can assess batch compilation separately after Business succeeds.

### Decision 6: Keep Test Case Scope Violations

**Rationale:** mario-hello.md deliberately includes Business layer boundary violations to test whether Business Agent properly filters/flags them. Good validation that boundary enforcement works in practice, not just in directives.

## Files Modified

### L1 Compilation Files (3 phase files cleaned)
1. `templates/agents/compiled/develop.rules.md` - Removed L0 Source citations from directives
2. `templates/agents/compiled/deploy.rules.md` - Removed L0 Source citations from directives
3. `templates/agents/compiled/validate.rules.md` - Removed L0 Source citations and outdated instruction

### Documentation Files (4 files updated)
4. `.github/copilot-instructions.md` - Updated L1 Directive Compilation description
5. `.github/agents/smaqit.L1.agent.md` - Updated compilation file characteristics and completion criteria
6. `docs/wiki/designs/level-up-compilation.md` - Updated example with current clean format
7. `templates/agents/compiled/validate.rules.md` - Removed outdated preservation instruction

### Task 068 Files (L0, L1, L2 - Business layer work)
8. `templates/agents/compiled/business.rules.md` - Removed Actor Concept, added NFR guidance, refined directives
9. `agents/smaqit.business.agent.md` - Compiled from L1 sources (System Actor removed, strengthened directives)
10. `framework/LAYERS.md` - Refined behavioral verbs directive to concept boundary
11. `docs/test-cases/mario-hello.md` - User reverted changes (kept scope violations for testing)

**Total: 11 files modified**

## Key Insights

### Compilation File Architecture Maturity

The compilation files now cleanly separate concerns:
- **Source L0 Principles table** - Documents WHERE principles come from (traceability)
- **L1 Directive Compilation** - Pure directives ready for Agent-L2 (no inline citations)
- **Compilation Guidance** - Instructions for Agent-L2 to merge with templates

This structure preserves transformation chain documentation while keeping directive sections clean for compilation consumption.

### Boundary Enforcement Through Concepts, Not Vocabulary

The behavioral verbs directive evolution showed that effective boundary enforcement requires concept-based rules, not word blacklists. "Display" is problematic when describing mechanism ("System displays via console rendering") but acceptable when describing outcome ("Error message is displayed"). Focus on HOW vs WHAT, not specific words.

### Template Minimalism vs Compilation File Specificity

The Actor Concept removal highlighted the proper division between templates (generic structure with placeholders) and compilation files (layer/phase-specific directives). Conceptual explanations belong in L0 framework or as comments in spec templates, not in L1 compilation files meant for Agent-L2 consumption.

### NFR Terminology Matters

Changing "system property advocates" to "non-functional requirement stakeholders" improved clarity. While semantically similar, "NFR stakeholders" immediately signals to practitioners that performance, security, compliance, accessibility concerns belong in Business layer as actor goals with measurable criteria.

## Next Steps

### Task 068 Remaining Work

**L2 verification (not yet done):**
- Test Business agent with mario-hello.md (includes deliberate scope violations)
- Verify agent filters/flags boundary violations correctly
- Verify no System Actor appears in generated spec
- Verify no technical verbs/mechanisms in flows

**Documentation (not yet done):**
- Create `docs/wiki/designs/why-no-system-actor.md` explaining removal rationale
- Update Task 068 status to complete

### Other Pending Work

**Active tasks:**
- Task 064: Complete Level 0 Principle Cleanup
- Task 065: Clean Up Level 1 Templates  
- Task 066: Clean Up Level 2 Product Agents
- Task 069: Strengthen Bounded Agents Principle

## Session Metrics

**Duration:** ~3 hours (across multiple interactions)  
**Tasks Worked:** Task 068 (L0/L1 complete, L2 verification pending)  
**Files Modified:** 11 total
- 3 phase compilation files cleaned (L0 Source citation removal)
- 4 documentation files updated
- 3 Task 068 files (L0 framework, L1 compilation, L2 agent)
- 1 test case (user reverted for testing purposes)

**Key Deliverables:**
- All 8 compilation files now have clean L1 Directive Compilation sections
- Business layer boundary enforcement strengthened (System Actor removed, HOW directive refined)
- Documentation synchronized across all levels
- Business agent ready for boundary violation testing
