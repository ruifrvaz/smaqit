# Task 058: Implementation Agents Should Update Acceptance Criteria Checkboxes

**Status:** Completed (2026-01-05)  
**Priority:** High  
**Effort:** 3-4 hours (actual: ~2 hours)  
**Related Issues:** E2E Test Issue 8 (Validation doesn't update checkboxes), E2E Test Issue 9 (Development has no checkbox directive)  
**Discovered:** 2026-01-05 during E2E test review  
**Context:** Session 030 E2E Testing Mario Hello

---

## Problem Statement

Implementation agents (Development, Deployment, Validation) do not update acceptance criteria checkboxes in specs after satisfying requirements. This creates a disconnect between implementation state and spec documentation.

**Current Behavior:**
- Development agent implements features, updates frontmatter to `status: implemented`, but leaves checkboxes unchecked `[ ]`
- Deployment agent deploys application, updates frontmatter to `status: deployed`, but leaves checkboxes unchecked `[ ]`
- Validation agent validates requirements, updates frontmatter to `status: validated`, but leaves checkboxes unchecked `[ ]`

**Expected Behavior:**
- Each implementation agent should update checkboxes for acceptance criteria it satisfies
- Checkboxes should reflect implementation progress: `[ ]` → `[x]` (pass) or `[!]` (fail/untestable)
- Specs become living documents showing which requirements are satisfied

---

## Root Cause Analysis

### Level 0: Framework (PHASES.md)

**Develop Phase Completion Criteria** (lines 78-89):
```markdown
- [ ] Spec frontmatter updated: `status: implemented`, `implemented: [ISO8601_TIMESTAMP]`
```
❌ **Missing:** No mention of checkbox updates

**Validate Phase Completion Criteria** (lines 226-234):
```markdown
- [ ] Spec frontmatter updated: `status: validated`, `validated: [ISO8601_TIMESTAMP]`
- [ ] Acceptance criteria checkboxes updated: `[ ]` → `[x]` or `[!]`
```
✅ **Present:** Checkbox updates mentioned, but unclear whose checkboxes (all specs or just Coverage?)

**Inconsistency:** Framework only mentions checkbox updates in Validate phase, but doesn't clarify if Validation agent updates ALL spec checkboxes (Phase 1 + Phase 2 + Coverage) or only Coverage checkboxes.

### Level 2: Agents

**Development Agent** (`agents/smaqit.development.agent.md` line 157):
```markdown
- [ ] Spec frontmatter updated: `status: implemented`, `implemented: YYYY-MM-DDTHH:MM:SSZ`
```
❌ **Missing:** No checkbox update directive

**Conflict:** Line 60 states "MUST NOT: Modify specifications (request changes through proper channels)"
- This could be interpreted as preventing checkbox updates
- Needs clarification that checkbox updates are part of implementation tracking, not spec modification

**Validation Agent** (`agents/smaqit.validation.agent.md`):
- Need to check if it has checkbox update directive
- If present, need to clarify: update Coverage checkboxes only, or all specs?

---

## Proposed Solution

### Design Decision: Each Phase Updates Its Own Checkboxes

**Principle:** Implementation agents update checkboxes for specs they work on as part of self-validation.

| Phase | Agent | Updates Checkboxes In | Rationale |
|-------|-------|----------------------|-----------|
| Develop | Development | Business, Functional, Stack specs | Agent implements these requirements, can confirm satisfaction |
| Deploy | Deployment | Infrastructure specs | Agent deploys to environment, can confirm infrastructure requirements met |
| Validate | Validation | Coverage specs | Agent executes tests, can confirm test cases passed/failed |

**Benefits:**
- Clear responsibility per phase
- Checkboxes updated immediately after implementation
- Specs reflect real-time progress
- No confusion about which agent updates what

**Coverage Spec Handling:**
- Coverage spec contains COV-* checkboxes that reference upstream requirements
- Validation agent updates Coverage checkboxes: `[x]` if test passed, `[!]` if test failed
- Validation agent does NOT update upstream spec checkboxes (Business, Functional, Stack, Infrastructure)
- Those were already updated by Development/Deployment agents in prior phases

---

## Implementation Plan

### Step 1: Update Level 0 Framework (PHASES.md)

**File:** `framework/PHASES.md`

**Develop Phase Completion Criteria** (add after frontmatter line):
```markdown
- [ ] Spec frontmatter updated: `status: implemented`, `implemented: [ISO8601_TIMESTAMP]`
- [ ] Acceptance criteria checkboxes updated in Business, Functional, Stack specs: `[ ]` → `[x]` or `[!]`
```

**Deploy Phase Completion Criteria** (add after frontmatter line):
```markdown
- [ ] Spec frontmatter updated: `status: deployed`, `deployed: [ISO8601_TIMESTAMP]`
- [ ] Acceptance criteria checkboxes updated in Infrastructure specs: `[ ]` → `[x]` or `[!]`
```

**Validate Phase Completion Criteria** (clarify existing line):
```markdown
- [ ] Spec frontmatter updated: `status: validated`, `validated: [ISO8601_TIMESTAMP]`
- [ ] Acceptance criteria checkboxes updated in Coverage specs: `[ ]` → `[x]` or `[!]`
```

**Rationale Section** (add new section explaining checkbox philosophy):
```markdown
## Acceptance Criteria Checkboxes

Each implementation agent updates checkboxes in the specs it processes:

- **Development agent:** Updates Business, Functional, Stack spec checkboxes
- **Deployment agent:** Updates Infrastructure spec checkboxes  
- **Validation agent:** Updates Coverage spec checkboxes

**Checkbox States:**
- `[ ]` — Not yet implemented/validated
- `[x]` — Satisfied (implementation complete or test passed)
- `[!]` — Failed, untestable, or not satisfied

**Self-Validation:** Checkbox updates are part of the implementation agent's self-validation process, confirming that requirements were addressed during execution.
```

### Step 2: Update Level 2 Agents

**File:** `agents/smaqit.development.agent.md`

**Completion Criteria** (line 157, add after):
```markdown
- [ ] Spec frontmatter updated: `status: implemented`, `implemented: YYYY-MM-DDTHH:MM:SSZ`
- [ ] Acceptance criteria checkboxes updated in all processed specs: `[ ]` → `[x]` (satisfied) or `[!]` (not satisfied/untestable)
```

**MUST NOT section** (line 60, clarify):
```markdown
- Modify specifications (request changes through proper channels)
  - **Exception:** Updating frontmatter status and acceptance criteria checkboxes is part of implementation tracking, not spec modification
```

**OR** reword to:
```markdown
- Modify specification requirements or structure (request changes through proper channels)
  - **Note:** Updating frontmatter and checkboxes for tracking purposes is expected
```

**File:** `agents/smaqit.deployment.agent.md`

**Completion Criteria** (add):
```markdown
- [ ] Spec frontmatter updated: `status: deployed`, `deployed: YYYY-MM-DDTHH:MM:SSZ`
- [ ] Acceptance criteria checkboxes updated in Infrastructure specs: `[ ]` → `[x]` (satisfied) or `[!]` (not satisfied)
```

**File:** `agents/smaqit.validation.agent.md`

**Completion Criteria** (clarify scope):
```markdown
- [ ] Spec frontmatter updated: `status: validated`, `validated: YYYY-MM-DDTHH:MM:SSZ`
- [ ] Acceptance criteria checkboxes updated in Coverage specs: `[ ]` → `[x]` (test passed) or `[!]` (test failed)
```

**Add clarification:**
```markdown
**Note:** Validation agent updates ONLY Coverage spec checkboxes. Phase 1 and Phase 2 spec checkboxes were updated by their respective implementation agents (Development, Deployment).
```

### Step 3: Update Agent Template (Level 1)

**File:** `templates/agents/implementation-agent.template.md`

**Completion Criteria section** (add standard checkbox directive):
```markdown
- [ ] Spec frontmatter updated: `status: [PHASE_STATUS]`, `[PHASE_STATUS]: [ISO8601_TIMESTAMP]`
- [ ] Acceptance criteria checkboxes updated in processed specs: `[ ]` → `[x]` (satisfied) or `[!]` (not satisfied/untestable)
```

### Step 4: Validate Consistency

**Check these files for any conflicting guidance:**
- `framework/ARTIFACTS.md` - Does it mention checkbox updates?
- `framework/TEMPLATES.md` - Does spec template include checkbox format?
- `templates/specs/*.template.md` - Do templates show checkbox format correctly?

**Ensure consistent terminology:**
- "Acceptance criteria checkboxes" (not just "checkboxes")
- "`[ ]` → `[x]` or `[!]`" (standard notation)
- "Self-validation" (process name)

---

## Testing Strategy

After implementing changes:

1. **Run Development agent** on test project
   - Verify Business, Functional, Stack spec checkboxes updated
   - Verify frontmatter shows `status: implemented`
   - Verify agent didn't error on "MUST NOT modify specifications"

2. **Run Deployment agent** on test project (if Infrastructure specs exist)
   - Verify Infrastructure spec checkboxes updated
   - Verify frontmatter shows `status: deployed`

3. **Run Validation agent** on test project
   - Verify Coverage spec checkboxes updated
   - Verify Phase 1/2 spec checkboxes NOT re-updated (stay as-is from prior phases)
   - Verify frontmatter shows `status: validated`

4. **Verify checkbox states**
   - Satisfied criteria: `[x]`
   - Untestable/failed criteria: `[!]`
   - Not addressed criteria: `[ ]` (should be rare after implementation)

---

## Success Criteria

- [x] PHASES.md includes checkbox update requirements for all three implementation phases
- [x] PHASES.md clarifies which agent updates which spec checkboxes (new "Acceptance Criteria Checkboxes" section)
- [x] Development agent has checkbox update directive in Completion Criteria
- [x] Deployment agent has checkbox update directive in Completion Criteria
- [x] Validation agent clarifies it only updates Coverage checkboxes
- [x] "MUST NOT modify specifications" clarified to exclude frontmatter/checkbox updates (all implementation agents)
- [x] Agent template includes checkbox update as standard completion criterion
- [x] All documentation uses consistent checkbox notation: `[ ]`, `[x]`, `[!]`
- [x] ARTIFACTS.md updated to reflect per-phase checkbox updates
- [x] Build verification successful (installer compiles)
- [ ] Test execution confirms agents update checkboxes correctly (requires E2E testing with agents)
- [x] No conflicts or ambiguities in framework/agent directives

---

## Implementation Summary

**Status:** Completed (2026-01-05)

**Files Modified:**

1. **framework/PHASES.md**
   - Added checkbox update to Develop phase completion criteria
   - Added checkbox update to Deploy phase completion criteria
   - Clarified Validate phase updates Coverage specs only
   - Added new "Acceptance Criteria Checkboxes" section explaining philosophy and responsibility

2. **framework/ARTIFACTS.md**
   - Updated "Acceptance Criteria State" to reflect all phases update checkboxes
   - Added checkbox updates to Develop phase artifacts
   - Added checkbox updates to Deploy phase artifacts
   - Clarified Validate phase updates Coverage specs only

3. **agents/smaqit.development.agent.md**
   - Updated "MUST NOT" to clarify frontmatter/checkbox updates are tracking, not modification
   - Added checkbox update directive to Completion Criteria

4. **agents/smaqit.deployment.agent.md**
   - Updated "MUST NOT" to clarify frontmatter/checkbox updates are tracking, not modification
   - Added checkbox update directive to Completion Criteria

5. **agents/smaqit.validation.agent.md**
   - Updated State Tracking section to clarify only Coverage specs updated
   - Removed language about updating all layers
   - Updated Completion Criteria to specify Coverage specs only

6. **templates/agents/implementation-agent.template.md**
   - Updated "MUST NOT" template with clarification note
   - Added checkbox update to standard completion criteria

**Key Changes:**

- Established clear responsibility: Each phase updates checkboxes for specs it processes
- Development → Business, Functional, Stack specs
- Deployment → Infrastructure specs
- Validation → Coverage specs only
- Clarified that checkbox/frontmatter updates are "implementation tracking" not "spec modification"
- Added comprehensive "Acceptance Criteria Checkboxes" section in PHASES.md
- Consistent checkbox notation across all documentation: `[ ]`, `[x]`, `[!]`

**Testing:**

- ✅ Installer builds successfully with no syntax errors
- ⏸️ Runtime E2E testing with actual agents deferred (requires agents to be invoked)

**Next Steps:**

- Run E2E testing with actual agent invocations to verify checkbox updates work correctly
- Monitor agent behavior to ensure checkboxes are updated as designed

---

## Notes

**Key Insight:** Checkbox updates are **implementation tracking**, not **spec modification**. They reflect work done, not changes to requirements.

**Design Philosophy:** Each implementation agent "signs off" on the requirements it addressed by updating checkboxes. This creates an audit trail of which phase satisfied which requirements.

**Alternative Considered (Rejected):** Have Validation agent update ALL checkboxes (Phase 1 + Phase 2 + Coverage).
- **Rejected because:** Validation agent doesn't have context of implementation decisions in prior phases. It can only confirm tests pass, not that implementation satisfies original intent.
- **Better approach:** Each agent updates checkboxes for specs it works on, based on its direct knowledge.

**Related to Issue 8:** E2E test discovered Validation agent doesn't update checkboxes. This task extends the fix beyond Validation to all implementation agents and establishes the pattern at framework level.

**Related to Issue 9:** E2E test review discovered Development agent has no checkbox directive and conflicting "MUST NOT modify specifications" rule.
