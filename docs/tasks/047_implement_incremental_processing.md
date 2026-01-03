# Implement Incremental Processing in Implementation Agents

**Status:** New  
**Created:** 2026-01-03  
**Related:** Task 014 (Stateful Specifications), Task 045 (Validation/Testing), Task 046 (Workflow Docs)  
**Depends On:** Task 046 (workflow documentation defines requirements)

## Preliminary Assessment

A preliminary assessment was conducted by the user-testing agent (2026-01-03). The testing agent stepped outside its intended role to analyze this task and propose an implementation approach. While the analysis is useful, it should be considered advisory guidance only.

**Key Findings:**
- Identified 6 files requiring modification (3 agents, 2 framework files, 1 template)
- Proposed State Checking workflow pattern: scan → read → categorize → report → process
- Defined processing modes: incremental (default), force (explicit)
- Created detailed implementation guidance

**Assessment Note:** The preliminary work provides a solid foundation but should be critically reviewed before implementation. Apply the standard critical assessment process to verify the approach aligns with smaqit principles and user needs.

## Description

Extend implementation agents (Development, Deployment, Validation) to support incremental processing by checking existing spec status and selectively processing only specs that need work. This completes the iterative development capability introduced by Task 014's stateful specifications infrastructure.

## Context

**Current State (Task 014):**
- ✅ Specs have YAML frontmatter with `status` field
- ✅ Agents update state after processing
- ✅ state.json tracks aggregate counts
- ❌ Agents process ALL specs regardless of current status

**Gap Identified (Task 045 Phase 1):**
Agents lack directives to:
- Read existing spec status before processing
- Skip specs with `status: implemented|deployed|validated`
- Process only specs with `status: draft` or `status: failed`

**User Value:**
Without incremental processing, users cannot:
- Add new features without regenerating everything
- Fix failed implementations selectively
- Resume work across sessions efficiently

## Scope

### In Scope

**1. Agent Behavior Updates**

Add state checking logic to 3 implementation agents:
- `smaqit.development.agent.md`
- `smaqit.deployment.agent.md`
- `smaqit.validation.agent.md`

**2. Processing Modes**

Define three processing modes:

| Mode | Behavior | Use Case |
|------|----------|----------|
| **Incremental** (default) | Process only `status: draft` or `status: failed` specs | Adding features, fixing failures |
| **Force** (explicit) | Process all specs, update existing | Refactoring, major changes |
| **Resume** (automatic) | Detect incomplete work, continue | Session interrupted, retry failed |

**3. State Checking Workflow**

Before processing specs, agents MUST:
1. Scan spec directories for existing files
2. Read YAML frontmatter from each spec
3. Check `status` field value
4. Categorize: draft, failed (process) vs implemented, deployed, validated (skip)
5. Report: "Processing X new specs, skipping Y completed specs"

**4. Framework Updates**

Update framework files with incremental processing directives:
- `framework/AGENTS.md` - Add state checking requirements
- `framework/PHASES.md` - Document processing modes

### Out of Scope

- CLI flags for mode selection (future enhancement)
- Parallel processing of specs (optimization, not core feature)
- Incremental validation within single spec (all-or-nothing per spec)
- Dependency resolution between specs (assumes independence)

## Acceptance Criteria

### Agent Updates

- [ ] Development agent has "State Checking" section before "State Tracking"
- [ ] Development agent directives: read frontmatter, check status, skip implemented
- [ ] Deployment agent has state checking directives
- [ ] Validation agent has state checking directives
- [ ] All agents document default behavior: incremental processing
- [ ] All agents document force mode: process all specs (explicit user request)

### Framework Updates

- [ ] `framework/AGENTS.md` "Implementation Agents" section updated with:
  - State checking requirements (MUST read before processing)
  - Processing mode definitions (incremental, force, resume)
  - Skip logic (which statuses to skip vs process)
- [ ] `framework/PHASES.md` updated with:
  - Incremental processing workflow description
  - When agents skip vs process specs
  - Force mode usage guidance

### Template Updates

- [ ] Implementation agent template includes State Checking section

### Validation

- [ ] Task 045 Phase 2 test scenarios pass (after implementation)
- [ ] Add new feature workflow works (new specs processed, existing skipped)
- [ ] Fix failed spec workflow works (failed spec reprocessed, others skipped)
- [ ] Force mode regenerates all specs when requested

## Implementation Steps

### Step 1: Define State Checking Directives (Framework)

**Update `framework/AGENTS.md`:**

Add after "Self-Validation Before Completion" section, before "Implementation Agents" section:

```markdown
### State-Based Processing

Implementation agents MUST check existing spec state before processing.

**Default behavior: Incremental processing**
- Read YAML frontmatter from all specs in target directories
- Process only specs with `status: draft` or `status: failed`
- Skip specs with `status: implemented`, `status: deployed`, or `status: validated`
- Report: "Processing X specs (Y draft, Z failed), skipping W completed"

**Force mode: Full regeneration**
- Process all specs regardless of status
- Update existing implementations
- User must explicitly request force mode

**Resume mode: Automatic recovery**
- Agents detect force mode is NOT requested
- Default incremental behavior resumes work
```

**Update `framework/PHASES.md`:**

Add "Incremental Processing" section to each phase description:

```markdown
### Incremental Processing

### Step 2: Update Implementation Agents

For each of the 3 implementation agents, add "State Checking" section before "State Tracking":

**Template for agent updates:**

```markdown
## State Checking

Before processing specs, [Agent Name] MUST:

1. **Scan spec directories:**
   - [List of directories for this agent, e.g., specs/business/, specs/functional/, specs/stack/]

2. **Read spec frontmatter:**
   - Parse YAML frontmatter from each .md file
   - Extract `status` field value

3. **Categorize specs:**
   - **Process:** `status: draft` or `status: failed`
   - **Skip:** `status: implemented` (for Development), `status: deployed` (for Deployment), `status: validated` (for Validation)

4. **Report processing plan:**
   - "Processing X specs: Y draft, Z failed"
   - "Skipping W specs: already [implemented|deployed|validated]"
   - If no specs to process: "All specs already processed. Use force mode to regenerate."

5. **Handle missing status:**
   - Specs without frontmatter or missing `status` field: treat as `status: draft` (process)
   - Log warning: "Spec [filename] missing status field, treating as draft"

**Default mode: Incremental**
Process only draft/failed specs.

**Force mode: Full regeneration**
When user explicitly requests, process all specs regardless of status. Update frontmatter to reflect reprocessing.
```

**Agent-specific directories (from preliminary assessment):**

| Agent | Directories to Scan | Skip Status |
|-------|-------------------|-------------|
| Development | `specs/business/`, `specs/functional/`, `specs/stack/` | `implemented` |
| Deployment | `specs/infrastructure/` | `deployed` |
| Validation | `specs/coverage/` | `validated` |

**Force mode: Full regeneration**
When user explicitly requests, process all specs regardless of status. Update frontmatter to reflect reprocessing.
```

**Agent-specific directories:**

| Agent | Directories to Scan |
|-------|-------------------|
| Development | `specs/business/`, `specs/functional/`, `specs/stack/` |
| Deployment | `specs/infrastructure/` |
| Validation | `specs/coverage/` |

### Step 3: Update Agent Templates

Update `templates/agents/implementation-agent.template.md` to include state checking section:

```markdown
## State Checking

[PHASE_NAME] agent MUST check existing spec state before processing.

1. Scan [SPEC_DIRECTORIES] for existing specs
2. Read YAML frontmatter and extract `status` field
3. Process only specs with `status: draft` or `status: failed`
4. Skip specs with `status: [PREVIOUS_PHASE_STATUS]`
5. Report processing plan to user

Default behavior: Incremental processing (skip already-processed specs).
```

### Step 4: Testing Validation

After implementation:
- Execute Task 045 Phase 2 test scenarios
- Verify incremental processing works as documented
- Update test report with results

### Step 5: Documentation

- Link Task 046 workflow documentation to new agent behavior
- Update wiki articles with incremental processing examples
- Add "Incremental Development" section to README.md

## Design Decisions

### Decision 1: Default to Incremental

**Chosen:** Incremental processing is default behavior (no flag needed)

**Rationale:**
- Aligns with stateful specifications value proposition
- Users expect agents to respect existing state
- Safer: avoids accidental regeneration
- Force mode requires explicit action (safer default)

**Alternative rejected:** Default to full processing, require flag for incremental. This inverts the primary use case.

### Decision 2: Missing Status = Draft

**Chosen:** Specs without `status` field treated as `status: draft` (process)

**Rationale:**
- Backward compatibility (specs created before Task 014)
- Fail-safe: ensures work isn't silently skipped
- Explicit state is agent responsibility, not blocker

**Alternative rejected:** Error on missing status. Too strict, breaks backward compatibility.

### Decision 3: No Dependency Resolution

**Chosen:** Specs processed independently, no dependency graph

**Rationale:**
- Specs are designed to be independent within layer
- Cross-layer dependencies handled by consolidation phase
- Simplifies implementation
- Avoids complex dependency resolution logic

**Alternative rejected:** Analyze references and process in dependency order. Adds complexity without clear benefit given current spec structure.

### Decision 4: All-or-Nothing Per Spec

**Chosen:** Each spec is fully processed or fully skipped (no partial processing)

**Rationale:**
- Atomic state transitions preserve consistency
- Partial implementation is harder to track
- Simpler mental model for users

**Alternative rejected:** Incremental validation within specs (e.g., check individual acceptance criteria). Too granular, complex state management.

## Open Questions

1. **How do agents detect force mode?**
   - Option A: User deletes all implementations → agent sees no existing state
   - Option B: Future CLI flag: `smaqit develop --force`
   - **Recommendation:** Option A for now (Task 047), Option B as enhancement

2. **Should validation agent skip validated specs?**
   - Yes: Consistent with other agents, skip `status: validated`
   - No: Validation should always re-run to catch regressions
   - **Recommendation:** Yes (skip validated), add "force" mode for regression testing

3. **What if spec changed but status is implemented?**
   - Agent has no way to detect content changes
   - Rely on `prompt_version` field (Task 014 stale detection, manual)
   - **Recommendation:** Document as user responsibility, future enhancement for auto-detection

## Dependencies

**Blocked By:**
- Task 046 (workflow documentation defines requirements) - **Should complete first**

**Blocks:**
- Task 045 Phase 2 (incremental workflow testing needs implementation)

**Related:**
- Task 014 (provided state tracking infrastructure)
- Task 031 (review implementation artifacts - may inform design)

## Success Metrics

**Quantitative:**
- All 3 agents updated with state checking
- Framework updated with incremental processing directives
- Task 045 Phase 2 tests pass

**Qualitative:**
- Users can add features without full regeneration
- Failed specs can be fixed selectively
- State tracking delivers iterative development value
- Documentation clearly explains incremental behavior

## Notes

**This task completes the iterative development vision from Task 014.** Infrastructure alone provides visibility but not workflow efficiency. Incremental processing delivers the core value: fast iteration without expensive regeneration.

**Target Release:** v0.5.0 (user decision to include in initial stateful specifications release)

**Preliminary Assessment Impact:**
The user-testing agent's preliminary analysis identified the implementation approach and created detailed guidance. While stepping outside its role, this work provides:
- Comprehensive file modification checklist
- Specific workflow pattern (scan → read → categorize → report → process)
- Processing mode definitions (incremental default, force explicit)
- Agent-specific directory mappings

This assessment should inform but not prescribe the actual implementation. Apply critical review to verify alignment with framework principles, user workflows, and technical feasibility.
