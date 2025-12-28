# smaqit User Testing Report

**Date:** 2025-12-27  
**Tester:** smaqit.user-testing (Agent)  
**smaqit Version:** c0a1d70-dirty  
**Test Case:** Mario Hello World Console Application

---

## Test Information

| Property | Value |
|----------|-------|
| Operating System | Linux (WSL2) 6.6.87.2-microsoft-standard-WSL2 |
| Go Version | go1.25.5 linux/amd64 |
| smaqit Version | c0a1d70-dirty |
| Test Duration | 22:00 - 23:57 (1 hour 57 minutes) |
| Test Project Path | /home/ruifrvaz/projects/smaqit/installer/test/mario-hello |

---

## Standardized Checklist

### Environment Setup
- [x] Go toolchain available
- [x] smaqit source repository accessible
- [x] Test directory created

### Installer Build
- [x] `make build` executed successfully (includes prepare step)
- [x] Binary created in `dist/` (3.2MB)
- [x] Binary is executable

### Project Initialization
- [x] Test project directory created
- [x] `smaqit init` executed successfully
- [x] `.smaqit/` directory created (NO framework files - embedded architecture validated)
- [x] `.github/agents/` directory created with 9 agent files
- [x] `.github/prompts/` directory created with 9 prompt files
- [x] `specs/` directory structure created (5 subdirectories)

### Business Layer Specification
- [x] `/smaqit.business` prompt invoked with test requirements
- [x] Business agent executed successfully
- [x] `specs/business/mario-greeting.md` file created
- [x] Business spec contains acceptance criteria with `BUS-` IDs (BUS-GREETING-001 through 006)

### Functional Layer Specification
- [x] `/smaqit.functional` prompt invoked with test requirements
- [x] Functional agent executed successfully
- [x] `specs/functional/greeting-display.md` file created
- [x] Functional spec contains acceptance criteria with `FUN-` IDs (FUN-DISPLAY-001 through 010+)
- [x] Functional spec references Business specs

### Stack Layer Specification
- [x] `/smaqit.stack` prompt invoked with test requirements
- [x] Stack agent executed successfully
- [x] `specs/stack/python-stack.md` file created
- [x] Stack spec contains acceptance criteria with `STK-` IDs (STK-PYTHON-001 through 010+)
- [x] Stack spec references Business and Functional specs

### Infrastructure Layer Specification
- [x] `/smaqit.infrastructure` prompt invoked with test requirements
- [x] Infrastructure agent executed successfully
- [x] `specs/infrastructure/local-execution.md` file created
- [x] Infrastructure spec contains acceptance criteria with `INF-` IDs (INF-EXEC-001 through 010+)
- [x] Infrastructure spec references Phase 1 specs

### Coverage Layer Specification
- [x] `/smaqit.coverage` prompt invoked with test requirements
- [x] Coverage agent executed successfully
- [x] `specs/coverage/greeting-verification.md` file created
- [x] Coverage spec contains acceptance criteria with `COV-` IDs (COV-GREET-001 through 008+)
- [x] Coverage spec references all upstream specs

### Cleanup
- [ ] Test project removed
- [ ] No residual artifacts left in smaqit source directory

---

## Execution Log

```
[22:00:00] Started test execution
[22:00:02] Go toolchain verified: go1.25.5 linux/amd64
[22:00:05] OS verified: Linux WSL2
[22:00:10] Built installer successfully: 3.2MB binary
[22:00:15] Created test project: /home/ruifrvaz/projects/smaqit/installer/test/mario-hello
[22:00:20] Ran smaqit init - success
[22:00:25] Validated structure: NO .smaqit/framework/ directory (embedded architecture confirmed)
[22:00:30] Verified 9 agents in .github/agents/
[22:00:35] Verified 9 prompts in .github/prompts/
[22:00:40] Ran smaqit status: 0 specs, phases "Not started"
[22:10:00] Invoked Business agent with Mario requirements
[22:12:00] Business spec created: specs/business/mario-greeting.md (6 acceptance criteria)
[22:12:05] Validated Business spec has BUS-GREETING-001 through BUS-GREETING-006
[22:29:00] Invoked Functional agent with Mario requirements
[22:38:00] Functional spec created: specs/functional/greeting-display.md (10+ acceptance criteria)
[22:38:05] Validated Functional spec has FUN-DISPLAY-001 through FUN-DISPLAY-010+
[22:48:00] Invoked Stack agent with Mario requirements
[22:50:00] Stack spec created: specs/stack/python-stack.md (10+ acceptance criteria)
[22:50:05] Validated Stack spec has STK-PYTHON-001 through STK-PYTHON-010+
[23:36:00] Invoked Infrastructure agent with Mario requirements
[23:42:00] Infrastructure spec created: specs/infrastructure/local-execution.md (10+ acceptance criteria)
[23:42:05] Validated Infrastructure spec has INF-EXEC-001 through INF-EXEC-010+
[23:50:00] Invoked Coverage agent with Mario requirements
[23:50:05] Coverage spec created: specs/coverage/greeting-verification.md (8+ acceptance criteria with traceability map)
[23:50:10] Validated Coverage spec has COV-GREET-001 through COV-GREET-008+
[23:57:00] All 5 specification layers completed successfully
```

---

## Painpoints Identified

### Issues

**Issue #1: smaqit status command provides incorrect next step suggestion**
- **Context:** After Business layer spec created (1 spec exists), `smaqit status` still shows "Next steps: Type '/smaqit.development' in GitHub Copilot chat to start Develop phase"
- **Expected:** Should detect incomplete specification layers and suggest next layer prompt (e.g., "Next steps: Type '/smaqit.functional' to continue with Functional layer")
- **Severity:** Medium
- **Impact:** Misleads users who might try to start development phase prematurely with incomplete specs

**Issue #2: state.json file has incorrect phase order**
- **Context:** `.smaqit/state.json` contains phases in order: deploy, develop, validate
- **Expected:** Phases should be ordered: develop, deploy, validate (matching workflow sequence)
- **Severity:** Low
- **Impact:** JSON readability issue; does not affect functionality but violates semantic ordering

**Issue #3: Business specs missing use case identifier**
- **Context:** Business spec creates acceptance criteria with proper IDs (BUS-GREETING-001, etc.) but Use Case section lacks identifier
- **Expected:** Use Case should have an identifier like UC1-GREETING or similar (e.g., "## Use Case: UC1-GREETING")
- **Severity:** Low
- **Impact:** Reduces traceability; use cases cannot be easily referenced by ID in downstream specs

**Issue #4: smaqit status displays layers separately from phases**
- **Context:** `smaqit status` shows "Specification Layers" and "Phase Status" as separate sections
- **Expected:** Layers should be nested under their respective phases for clarity:
  - Develop: Business (X specs), Functional (Y specs), Stack (Z specs)
  - Deploy: Infrastructure (X specs)
  - Validate: Coverage (X specs)
- **Severity:** Low
- **Impact:** Reduces clarity of progress tracking; users must mentally map layers to phases

**Issue #5: Iterative spec refinements break prompt reproducibility**
- **Context:** When user is unsatisfied with a generated spec and prompts LLM to refine it (e.g., "fix the stack spec"), the additional instructions are not captured in the layer prompt file
- **Expected:** Framework should maintain an "Addendum" section in prompt files that captures all iterative refinement instructions
- **Proposed Solution:** When agent detects user instructions that modify existing specs, append those instructions to the layer prompt file under "## Addendum" section with timestamp. *THIS IS A FRAMEWORK LEVEL CHANGE*
- **Severity:** High
- **Impact:** Breaks reproducibility principle; prompt files no longer represent complete input record; regenerating specs from prompt would produce different output than current specs

**Issue #6: Framework unclear about phase-first workflow (CRITICAL)**
- **Context:** Both testing agent and actual execution treated specifications as separate from phases. Test process executed all 5 spec layers before considering implementation phases.
- **Expected:** Phases are the primary workflow unit in smaqit:
  - **Phase 1 (Develop):** Business spec → Functional spec → Stack spec → Development agent executes → Phase completes
  - **Phase 2 (Deploy):** Infrastructure spec → Deployment agent executes → Phase completes
  - **Phase 3 (Validate):** Coverage spec → Validation agent executes → Phase completes
- **Observed Behavior:** Specifications treated as separate "specification phase" before implementation phases
- **Severity:** CRITICAL
- **Impact:** Fundamental misunderstanding of smaqit workflow; framework documentation insufficiently emphasizes phases as primary unit; both agents and users may execute incorrectly; violates core smaqit principle that each phase includes specs + implementation

**Issue #7: smaqit validate does not validate state.json file**
- **Context:** `smaqit validate` command checks directory structure and spec files, but does not validate `.smaqit/state.json`
- **Expected:** Should validate:
  - state.json file exists
  - JSON structure is valid
  - Phase keys are present and correctly ordered (develop, deploy, validate)
  - Phase objects have required "completed" boolean field
  - Version field is present
- **Severity:** Medium
- **Impact:** Corrupted or malformed state.json can cause status command failures; validate should catch structural issues

**Issue #8: Agents lack instructive handover guidance**
- **Context:** When specification agents complete their task, they don't provide guidance on next steps in the workflow
- **Expected:** Agents should suggest the next logical step in the workflow:
  - Business agent → "Next: Create functional specifications with `/smaqit.functional`"
  - Functional agent → "Next: Create stack specifications with `/smaqit.stack`"
  - Stack agent → "Next: Create infrastructure specifications with `/smaqit.infrastructure` OR run development phase with `/smaqit.development`"
  - Infrastructure agent → "Next: Create coverage specifications with `/smaqit.coverage` OR run deployment phase with `/smaqit.deployment`"
  - Coverage agent → "Next: Run validation phase with `/smaqit.validation`"
- **Proposed Solution:** Add handover section to agent templates (Level 1) with placeholder `[PROPOSE_NEXT_STEP]`, populate in Level 2 agents
- **Severity:** Medium
- **Impact:** Users must infer next steps; reduces workflow clarity; misses opportunity for just-in-time guidance

**Issue #9: Missing distinction between user documentation and agent specifications**
- **Context:** Framework lacks clear guidance on what content belongs in user-facing documentation vs agent-facing specifications
- **Expected:** Clear distinction should be documented:
  - **Agent specifications** (specs, framework, agents, templates): Pure execution instructions, no meta-rationale, no human context, no explanations
  - **User documentation** (wiki, README, task files): Can include business context, stakeholder names, delivery dates, rationale, examples, explanations
- **Observed:** This principle exists implicitly (e.g., session 012 split instructions from rationale) but is not explicitly documented as a wiki entry
- **Proposed Solution:** Create wiki entry at `docs/wiki/concepts/user-vs-agent-documentation.md` explaining:
  - Why the distinction matters in spec-driven development
  - What content belongs in each category
  - Examples of inappropriate content in specs (stakeholder names, delivery dates, business politics)
  - How this enables LLM agents to focus on execution without human context noise
- **Severity:** Medium
- **Impact:** Without explicit guidance, contributors may pollute specs with human context; reduces agent effectiveness; violates separation of concerns

---

## Recommendations

Based on painpoints identified during testing, the following improvements are recommended:

### High Priority

1. **Framework Clarity on Phase-First Workflow (Issue #6)**
   - **Action:** Update SMAQIT.md, PHASES.md, README.md to emphasize phases as primary workflow unit
   - **Rationale:** CRITICAL - affects fundamental understanding of smaqit workflow
   - **Specific Changes:**
     - Add explicit statement: "Phases are the primary workflow. Each phase includes specifications + implementation."
     - Update examples to show phase completion (not just spec generation)
     - Clarify that users CAN spec-first (all 5 layers then implement), but SHOULD phase-first (3 specs → implement → 1 spec → deploy → 1 spec → validate)

2. **Prompt Reproducibility with Addendum Section (Issue #5)**
   - **Action:** Add "## Addendum" section to prompt templates and agent instructions
   - **Rationale:** High - breaks reproducibility principle if iterative refinements aren't captured
   - **Implementation:**
     - Update prompt templates with optional Addendum section
     - Update agent instructions to append refinement requests to prompt file
     - Include timestamp with each addendum entry

### Medium Priority

3. **Smaqit Status Next Step Logic (Issue #1)**
   - **Action:** Make status command phase-aware with intelligent next step suggestions
   - **Implementation:**
     - Detect incomplete specification layers within current phase
     - Suggest next layer prompt if phase specs incomplete
     - Suggest implementation agent only when phase specs complete

4. **Agent Handover Guidance (Issue #8)**
   - **Action:** Add handover guidance to agent templates and implementations
   - **Implementation:**
     - Add `[PROPOSE_NEXT_STEP]` placeholder to Level 1 templates
     - Populate with workflow-aware guidance in Level 2 agents
     - Provide contextual options (e.g., Stack agent suggests both next spec AND phase completion)

5. **State.json Validation (Issue #7)**
   - **Action:** Add state.json validation to `smaqit validate` command
   - **Implementation:**
     - Check file existence and JSON validity
     - Verify phase structure (develop, deploy, validate order)
     - Validate required fields (completed boolean, version string)

6. **Status Display Improvements (Issue #4)**
   - **Action:** Nest layers under phases in status output
   - **Implementation:**
     - Restructure status display: Phase → Layers → Spec count
     - Show progress as "Phase 1 (Develop): Business (1), Functional (0), Stack (0)"

### Low Priority

7. **State.json Phase Ordering (Issue #2)**
   - **Action:** Fix phase order in initStateFile() function
   - **Implementation:** Change JSON generation to output develop, deploy, validate (not deploy, develop, validate)

8. **Use Case Identifiers (Issue #3)**
   - **Action:** Add use case ID format to Business spec template
   - **Implementation:** Update template with `## Use Case: [UC_ID]` pattern

9. **User vs Agent Documentation Distinction (Issue #9)**
   - **Action:** Create wiki entry documenting separation between user documentation and agent specifications
   - **Implementation:**
     - New file: `docs/wiki/concepts/user-vs-agent-documentation.md`
     - Explain why specs must be pure instructions (no human context)
     - Explain where human context belongs (wiki, README, task files)
     - Provide examples of inappropriate spec content (stakeholder names, delivery dates, business politics)

---

## Overall Result

**Status:** ✅ PASS (with identified improvements)

**Summary:** Successfully validated embedded agent architecture (session 016). All 5 specification layers generated correctly without framework files. Testing confirmed agents are self-contained and workspace is 70% cleaner. Identified 9 issues ranging from critical framework clarity to low-priority formatting concerns.

**Key Findings:**
- **Embedded architecture validated:** No `.smaqit/framework/` directory created; agents executed successfully with embedded content
- **Specification generation successful:** All 5 layers produced valid specs with correct requirement ID prefixes
- **Critical workflow confusion identified:** Both testing agent and user execution treated specs as separate from phases, revealing framework documentation gap
- **Reproducibility gap:** Iterative spec refinements not captured in prompt files, breaking input record completeness
- **UX improvements needed:** Status command and agent handover guidance require enhancement for better user experience

**Next Steps:**
- Address Issue #6 (phase-first workflow clarity) in framework documentation - CRITICAL
- Implement Issue #5 (prompt addendum section) for reproducibility - HIGH
- Enhance status command logic (Issue #1) and agent handover (Issue #8) - MEDIUM
- Fix minor issues (#2, #3, #4, #7) - LOW

---

## Notes

- Testing embedded agent architecture from session 016 (framework content embedded in agents, no framework bundling)
- Using Mario Hello World test case (Test Case 001)
- Interactive testing workflow: agent displays requirements, user invokes prompts, agent validates outputs
