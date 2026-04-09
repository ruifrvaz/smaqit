# SDK Template Extraction

**Date:** 2026-01-25  
**Session Focus:** Level 1 template cleanup through systematic directive extraction, SDK extensibility preservation  
**Tasks Completed:** Task 065 (Clean Up Level 1 Templates)  
**Tasks Updated:** Task 071 (Create Q&A Agent) - aligned with SDK architecture

## Session Overview

Session began with goal to prepare for Task 071 (Q&A Agent creation) but discovered fundamental L1 contamination requiring systematic cleanup. Executed complete extraction of directive content from agent templates to compilation files, establishing pure template + compilation file pattern that became the foundation of smaqit's SDK capability.

## Actions Taken

### Phase 1: Planning and Assessment (Todos 1-6)

Created systematic 6-step extraction plan:
1. Extract specification-agent directives to `specification.rules.md`
2. Extract implementation-agent directives to `implementation.rules.md`
3. Refine `base.rules.md` to foundation directives only
4. Refactor `specification-agent.template.md` to pure placeholders
5. Refactor `implementation-agent.template.md` to pure placeholders
6. Update Agent-L2 with 4-way merge instructions

### Phase 2: Specification Template Extraction (Todo 1)

Created `templates/agents/compiled/specification.rules.md` (276 lines):
- **Source L0 Principles table** documenting transformation sources
- **L1 Directive Compilation** with 10 content structure sections:
  - Role Content Structure (Agent Identity + Goal + Context)
  - Input Content Structure (User Input + Upstream Specifications + Conflict Resolution)
  - Output Content Structure (Location + Template + Format)
  - Specification-Extension MUST Directives (9 items)
  - Specification-Extension MUST NOT Directives (6 items)
  - Specification-Extension SHOULD Directives (5 items)
  - Scope Boundary Enforcement (3-step pattern)
  - Requirement ID Format Rules
  - Acceptance Criteria Format Rules (testability table)
  - File Organization Rules
  - Incremental Spec Updates (decision table)
  - Completion Criteria Content (6 spec-specific checks)
  - Workflow Handover Content
- **Compilation Guidance for Agent-L2** with 9 merging subsections

### Phase 3: Implementation Template Extraction (Todo 2)

Created `templates/agents/compiled/implementation.rules.md` (432 lines):
- Same structure as specification.rules.md but for implementation workflow
- 10 content structure sections including:
  - Cross-Layer Consolidation Content (4-step workflow)
  - Phase-Specific Rules Content (placeholder with compilation instructions)
  - State Tracking Content (spec frontmatter + upstream spec updates)
  - Failure Handling Content (5-row table + stop conditions)
- Implementation-Extension MUST Directives (16 items)
- Implementation-Extension MUST NOT Directives (8 items)
- Implementation-Extension SHOULD Directives (6 items)
- Compilation Guidance for Agent-L2 with 13 merging subsections

### Phase 4: Base Rules Refinement (Todo 3)

Refined `templates/agents/compiled/base.rules.md`:
- Removed 3 workflow-specific directives that belonged in spec/impl.rules
- Result: 9 MUST + 9 MUST NOT pure foundation directives
- Foundation only: self-validation, scope boundaries, output validation, clarity, fail-fast

### Phase 5: Template Refactoring (Todos 4-5)

**Specification template** (`specification-agent.template.md`):
- Reduced to 78 lines of pure placeholder structure
- All directive content removed
- Sections replaced with placeholders: `[ROLE_CONTENT]`, `[BASE_MUST_DIRECTIVES]`, `[SPECIFICATION_MUST_DIRECTIVES]`, `[LAYER_MUST_DIRECTIVES]`, etc.
- Zero HTML comments

**Implementation template** (`implementation-agent.template.md`):
- Reduced to 74 lines of pure placeholder structure
- Same pattern as specification template
- All directive content extracted to implementation.rules.md

**Information gap analysis performed:**
- Scanned templates for content not yet in compilation files
- Added missing content structures (Role, Input, Output, Completion Criteria, Workflow Handover, Failure Handling)
- Moved meta-guidance (Purpose/Structure) to Compilation Guidance sections
- Final verification: no information loss, no duplication, clean structure

### Phase 6: Agent-L2 Update (Todo 6)

Updated `.github/agents/smaqit.L2.agent.md` (464 lines):
- Changed from 3-way merge to **4-way merge** documentation:
  - Base template + base.rules + specification/implementation.rules + layer/phase.rules → product agent
- **Hierarchy explanation:**
  - Foundation (base) → Universal agent behaviors
  - Extension type (spec/impl) → Workflow family shared behaviors
  - Specific role (layer/phase) → Unique behaviors per agent
- Added compilation files to Input section: `specification.rules.md`, `implementation.rules.md`
- Documented section-level compilation for all sections (Role, Input, Output, Directives, etc.)
- Added 4-way merge example showing hierarchical directive ordering

### Phase 7: Opportunistic Cleanup

**HTML comment contamination fix:**
- Discovered HTML comment in `implementation.rules.md` Phase-Specific Rules Content section
- Contained Agent-L2 instructions (L1 contamination violating principle that "L1 instructions belong in compilation files")
- Removed HTML comment from content structure
- Expanded "Merging Phase-Specific Rules Content" subsection in Compilation Guidance with detailed 3-step compilation process
- Aligned with specification.rules.md pattern (zero HTML comments in content structures)

**Orchestrator cleanup:**
- Removed obsolete `orchestrator.template.md` and `smaqit.orchestrator.agent.md` references from Agent-L2 Input section

### Phase 8: SDK Extensibility Recovery

**Problem identified:**
User flagged critical misalignment: "base template has indeed been extended by the implementation and specification types of agents, but these are specific extensions of the base. for other agent types, like Q&A and helper, they will still extend the base."

Agent-L2 had become over-specialized for specification/implementation workflows only, losing SDK capability to compile Q&A agents, helper agents, and other simple agents through base template.

**Solution implemented:**
Added **SDK Pattern 1 (Base Agents)** to Agent-L2 Compilation Architecture:

**3 Compilation Patterns documented:**
1. **Pattern 1: Base Agents (2-way merge)** - base-agent.template.md + base.rules.md → Q&A, helper, orchestrator, custom utilities
2. **Pattern 2: Specification Agents (4-way merge)** - specification template + base + specification.rules + layer.rules
3. **Pattern 3: Implementation Agents (4-way merge)** - implementation template + base + implementation.rules + phase.rules

**Pattern 1 compilation process (6 steps):**
1. Read base template for pure structure
2. Read base rules for foundation directives (9 MUST, 9 MUST NOT)
3. Merge both with agent-specific content
4. Validate (no placeholders, self-contained)

**Section-level compilation updated:**
- Added "Role section (Base Agents)" pattern
- Split Directives section into Base Agents (2-way) vs Spec/Impl Agents (4-way)
- Added 2-way merge example (Q&A agent)

**SDK extensibility preserved:**
- Base template remains foundation for ALL agent types
- Specification/Implementation templates are EXTENSIONS for specific workflows
- Task 071 Q&A agent can now use Pattern 1 (base only)

### Phase 9: Task Management

**Task 065 marked completed:**
- Moved from Active to Completed in `docs/tasks/PLANNING.md`
- Updated `docs/tasks/065_clean_up_level_1_templates.md` with completion summary
- Documented systematic extraction approach and results achieved

**Task 071 updated for SDK alignment:**
- Replaced "use specification agent template" with SDK Pattern 1 (Base Agents) approach
- Added compilation workflow (Agent-L2 invocation with base template + base rules)
- Documented Q&A-specific customizations (role, input, output, scope, extension directives)
- Updated acceptance criteria to include compilation validation
- Added implementation notes explaining 2-way merge pattern
- Marked template question as resolved

## Problems Solved

### Problem 1: Directive Form Contamination

**Issue:** MUST NOT directives mixed in MUST section using "NOT" prefix (e.g., "NOT include specific technologies")

**Solution:** Separated all negations to proper MUST NOT sections in both `specification.rules.md` and `base.rules.md`

**Lesson:** "Never mix positive and negative directives using 'NOT' prefix within MUST section. Extract negations to proper MUST NOT section."

### Problem 2: Directive Form Guidance Over-Complication

**Issue:** Verbose table format for directive form guidance when single directive sufficient

**Solution:** Simplified Agent-L1 directive form guidance to single directive

**Updated:** `.github/agents/smaqit.L1.agent.md` with simplified Directive Form Standards section

### Problem 3: Merge Errors and Formatting Issues

**Issues encountered:**
- Failure Handling table mixed into Completion Criteria section
- Unclosed code fence in Workflow Handover Content section
- Duplicate "Compilation Guidance for Agent-L2" section

**Solutions:**
- Fixed section separation (moved Failure Handling table to proper section with header)
- Fixed unclosed code fence, removed malformed merge artifacts
- Removed duplicate section, consolidated all merging instructions in proper order

### Problem 4: Information Loss Risk

**Issue:** Converting mixed placeholder/directive content to pure placeholders risked losing content

**Solution:** Comprehensive information gap analysis workflow:
1. Scan template for all content
2. Compare with compilation file
3. Identify missing structures
4. Add with proper formatting
5. Verify no duplication or gaps

**Results:** Added Role Content Structure, Input Content Structure, Output Content Structure, Completion Criteria Content, Workflow Handover Content, Failure Handling Content to both specification.rules.md and implementation.rules.md

### Problem 5: Purpose vs Content Structure Confusion

**Issue:** Purpose and Structure subsections mixed with content structure definitions

**Solution:** Clarified Purpose/Structure are meta-guidance, not content structure. Moved to Compilation Guidance section as meta-instructions for Agent-L2.

### Problem 6: HTML Comment Contamination

**Issue:** HTML comment in implementation.rules.md Phase-Specific Rules Content section violated principle that "L1 instructions belong in compilation files, not templates"

**Solution:** Removed HTML comment from content structure, expanded Compilation Guidance subsection with detailed instructions. Aligned with specification.rules.md pattern (zero HTML comments).

### Problem 7: SDK Extensibility Loss

**Issue:** Agent-L2 over-specialized for spec/impl agents only, losing capability to compile Q&A/helper agents through base template (2-way merge pattern)

**Root cause:** Incorrectly assumed specification/implementation templates REPLACED base template when they actually EXTEND it

**Solution:** Added SDK Pattern 1 (Base Agents) to Agent-L2, documenting 2-way merge for simple agents. Preserved base template as foundation for ALL agent types.

## Decisions Made

### Decision 1: Pure Template Pattern

**Choice:** Extract ALL directive content to compilation files, refactor templates to pure placeholders

**Rationale:**
- Eliminates duplication between templates
- Centralizes directives for easier maintenance
- Enables systematic compilation through Agent-L2
- Establishes clear L0→L1→L2 boundaries

**Result:** Templates are pure structure (78-74 lines), compilation files comprehensive (276-432 lines)

### Decision 2: 4-Way Merge Hierarchy

**Choice:** Foundation (base) → Workflow Extension (spec/impl) → Role-Specific (layer/phase)

**Rationale:**
- Clarifies directive inheritance and precedence
- Separates universal behaviors from workflow-specific from role-specific
- Enables Task 071 Q&A agent (foundation only, no workflow)
- Proves SDK extensibility capability

**Result:** Agent-L2 documents 3 compilation patterns (base agents, spec agents, impl agents)

### Decision 3: Content Structures in Compilation Files

**Choice:** Add Role/Input/Output/Completion/Workflow/Failure content structures to compilation files

**Rationale:**
- Prevents information loss during template refactoring
- Provides complete compilation guidance for Agent-L2
- Documents Purpose and Structure for each section
- Enables pure placeholder templates

**Result:** Compilation files include both directives AND content structures for complete agent generation

### Decision 4: Directive Form Standards

**Choice:** Single directive "Never mix positive and negative directives using 'NOT' prefix within MUST section"

**Rationale:**
- Simpler than verbose table
- Clear and actionable
- Sufficient for enforcement
- Easier to maintain

**Result:** Agent-L1 updated with simplified Directive Form Standards section

### Decision 5: HTML Comment Removal from Content Structures

**Choice:** Remove all HTML comments from content structure sections, move instructions to Compilation Guidance

**Rationale:**
- L1 instructions belong in compilation files (Compilation Guidance section), not in content structures
- Aligns with specification.rules.md pattern (zero HTML comments)
- Eliminates duplication between content structure and Compilation Guidance
- Centralizes all Agent-L2 instructions in proper section

**Result:** implementation.rules.md clean with zero HTML comments in content structures

### Decision 6: SDK Pattern 1 (Base Agents) Documentation

**Choice:** Document base agent 2-way merge as equal pattern to spec/impl 4-way merge

**Rationale:**
- Base agents are equally valid as spec/impl agents
- SDK extensibility is core capability, not edge case
- Task 071 proves need (Q&A needs base only, no workflow)
- Clearest documentation: 3 explicit patterns with examples

**Result:** Agent-L2 Compilation Architecture includes Pattern 1 (Base Agents), Pattern 2 (Specification Agents), Pattern 3 (Implementation Agents)

## Files Modified

### Created Files

1. `templates/agents/compiled/specification.rules.md` (276 lines)
   - Specification-extension directives for all specification agents (Business, Functional, Stack, Infrastructure, Coverage)
   - 10 content structure sections
   - 9 Compilation Guidance subsections for Agent-L2

2. `templates/agents/compiled/implementation.rules.md` (432 lines)
   - Implementation-extension directives for all implementation agents (Development, Deployment, Validation)
   - 10 content structure sections including Cross-Layer Consolidation, State Tracking, Failure Handling
   - 13 Compilation Guidance subsections for Agent-L2

### Modified Files

1. `templates/agents/compiled/base.rules.md`
   - Removed 3 workflow-specific directives
   - Result: 9 MUST + 9 MUST NOT pure foundation directives

2. `templates/agents/specification-agent.template.md`
   - Reduced from ~200 lines to 78 lines
   - All directive content removed, replaced with placeholders
   - Pure structure: `[ROLE_CONTENT]`, `[BASE_MUST_DIRECTIVES]`, `[SPECIFICATION_MUST_DIRECTIVES]`, `[LAYER_MUST_DIRECTIVES]`, etc.

3. `templates/agents/implementation-agent.template.md`
   - Reduced from ~197 lines to 74 lines
   - All directive content removed, replaced with placeholders
   - Same pattern as specification template

4. `.github/agents/smaqit.L2.agent.md`
   - Updated from 399 lines to 464 lines
   - Changed from 3-way merge to 4-way merge documentation
   - Added SDK Pattern 1 (Base Agents) with 2-way merge
   - Added specification.rules.md and implementation.rules.md to Input section
   - Documented 3 compilation patterns with examples
   - Updated section-level compilation for all patterns

5. `.github/agents/smaqit.L1.agent.md`
   - Added Directive Form Standards section
   - Simplified directive form guidance to single directive

6. `docs/tasks/PLANNING.md`
   - Moved Task 065 from Active to Completed

7. `docs/tasks/065_clean_up_level_1_templates.md`
   - Updated status to completed
   - Added completion summary with full session details

8. `docs/tasks/071_qa_agent_github_skill.md`
   - Updated deliverable 1 with SDK Pattern 1 (Base Agents) approach
   - Added compilation workflow
   - Updated acceptance criteria
   - Added implementation notes for 2-way merge
   - Marked template question as resolved

## Key Outcomes

### Architectural Achievements

1. **L1 Template Purity Established:**
   - Templates are pure structure with zero directive content
   - Directives consolidated in compilation files
   - Zero duplication between templates
   - Clear L0→L1→L2 boundaries maintained

2. **SDK Extensibility Proven:**
   - 3 compilation patterns documented (base, spec, impl)
   - Base agents (2-way merge) enable Q&A, helper, custom agents
   - Specification agents (4-way merge) for layer-specific workflows
   - Implementation agents (4-way merge) for phase-specific workflows

3. **4-Way Merge Hierarchy:**
   - Foundation (base) → universal behaviors
   - Workflow extension (spec/impl) → family shared behaviors
   - Role-specific (layer/phase) → unique behaviors
   - Hierarchical directive ordering: base → spec/impl → layer/phase

4. **Content Structure Pattern:**
   - Role Content Structure (Agent Identity + Goal + Context/Phase Context)
   - Input Content Structure (User Input + Upstream Specs + Conflict Resolution)
   - Output Content Structure (Location/Artifacts + Template/Format)
   - Completion Criteria Content
   - Workflow Handover Content
   - Failure Handling Content

5. **Compilation Guidance Pattern:**
   - Each compilation file includes "Compilation Guidance for Agent-L2"
   - Step-by-step merge instructions for each section
   - Purpose and Structure meta-guidance
   - Merge order documentation

### Quantitative Results

- **2 compilation files created** (specification.rules.md: 276 lines, implementation.rules.md: 432 lines)
- **2 templates refactored** to pure structure (78 lines, 74 lines)
- **1 base rules file refined** (9 MUST, 9 MUST NOT foundation directives)
- **2 agent definition files updated** (Agent-L1, Agent-L2)
- **3 task files updated** (065 completed, 071 aligned with SDK)
- **23 content structure sections** documented across both compilation files
- **22 Compilation Guidance subsections** documented for Agent-L2
- **3 compilation patterns** established (base 2-way, spec 4-way, impl 4-way)
- **Zero information loss** verified through gap analysis
- **Zero HTML comments** in content structures
- **Zero directive duplication** between files

## Next Steps

### Immediate (Next Session - Task 071)

**Primary goal:** Create Q&A Agent using SDK Pattern 1 (Base Agents)

**Workflow:**
1. Invoke Agent-L2 to compile Q&A agent
2. Agent-L2 reads `base-agent.template.md` for structure
3. Agent-L2 reads `base.rules.md` for foundation directives (9 MUST, 9 MUST NOT)
4. Agent-L2 fills template with Q&A-specific content:
   - Role: Q&A agent identity, goal (fetch and answer docs), context (read-only)
   - Input: User questions, wiki URLs, framework files
   - Output: Answers with source references
   - Scope Boundaries: MUST NOT generate code, create specs, perform implementation
   - Extension Directives: MUST fetch from GitHub, MUST provide references, MUST redirect out-of-scope
5. Agent-L2 validates: no placeholders, self-contained, foundation directives embedded
6. Create `.github/skills/smaqit-docs.skill.md` (GitHub skill version)

**Success criteria:** Task 071 acceptance criteria met, Q&A agent demonstrates SDK Pattern 1 capability

### Follow-Up Tasks

1. **Task 066: Clean Up Level 2 Product Agents**
   - Apply compilation files to existing product agents
   - Ensure all agents follow 4-way merge pattern
   - Validate no L1 contamination in L2 agents

2. **Task 064: Complete Level 0 Principle Cleanup**
   - Extract remaining directives from L0 framework files
   - Ensure L0 contains only principles and philosophy
   - Document extracted directives for L1 compilation

3. **SDK Documentation**
   - Document SDK patterns in wiki
   - Create quickstart for custom agent creation
   - Explain 2-way vs 4-way merge use cases

## Session Metrics

- **Duration:** Full session (multiple hours)
- **Tasks Completed:** 1 (Task 065)
- **Tasks Updated:** 1 (Task 071)
- **Files Created:** 2 (specification.rules.md, implementation.rules.md)
- **Files Modified:** 6 (templates, agent definitions, task files)
- **Compilation Files:** 3 total (base.rules.md refined, specification.rules.md created, implementation.rules.md created)
- **Templates Refactored:** 2 (specification, implementation)
- **Compilation Patterns Established:** 3 (base 2-way, spec 4-way, impl 4-way)
- **Lines of Compilation Guidance:** 708 (276 + 432)
- **Directive Extractions:** 40+ directives consolidated from templates to compilation files
- **Information Gap Analyses:** 2 (specification template, implementation template)
- **Opportunistic Cleanups:** 3 (directive form, HTML comments, orchestrator removal)
- **SDK Capability Proven:** Pattern 1 enables Task 071 Q&A agent

## Lessons Learned

1. **Pure templates prevent duplication** - Extracting all directive content to compilation files eliminates inconsistencies and enables systematic compilation

2. **Information gap analysis critical** - Template refactoring requires comprehensive scanning to ensure no content loss

3. **Purpose vs content structure clarity needed** - Meta-guidance (Purpose/Structure) belongs in Compilation Guidance, not content structure sections

4. **SDK extensibility must be explicit** - Base template pattern (2-way merge) is equally important as workflow extensions (4-way merge)

5. **Level contamination vigilance** - HTML comments with L1 instructions violate level boundaries when placed in content structures

6. **Directive form matters** - Separating MUST (positive) from MUST NOT (negative) improves clarity and prevents confusion

7. **Hierarchical merge order** - Foundation → Workflow Extension → Role-Specific provides clear precedence and inheritance

8. **Content structures enable pure templates** - Role/Input/Output/Completion/Workflow/Failure structures in compilation files allow templates to be pure placeholders

9. **Compilation guidance is critical** - Agent-L2 needs explicit step-by-step instructions for each section merge, not just directive lists

10. **SDK proves architecture** - Task 071 requiring base-only agent validates that specification/implementation are extensions, not replacements, of base template
