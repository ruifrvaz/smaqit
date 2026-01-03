# Document Implementation Workflows

**Status:** New  
**Created:** 2026-01-03  
**Related:** Task 014, Task 045, Task 047

## Description

Create comprehensive documentation for implementation agent workflows, covering how agents process specifications, handle state, and support iterative development patterns. This documentation should live in the wiki for human understanding of agent behaviors.

## Context

Task 045 testing revealed a gap: while stateful specifications infrastructure exists, incremental processing workflows aren't fully documented or implemented. Implementation agents need clear workflows for:

1. **Initial implementation** - Processing fresh `status: draft` specs
2. **Incremental addition** - Handling new specs while skipping implemented ones
3. **Refinement/fixes** - Reprocessing `status: failed` specs
4. **Full regeneration** - When to regenerate all specs vs partial updates
5. **State transitions** - How specs move through lifecycle states

## Scope

### In Scope

**Wiki Articles to Create:**

1. **`docs/wiki/workflows/implementation-agent-lifecycle.md`**
   - How implementation agents process specs
   - State checking logic (what agents should do)
   - State update patterns (frontmatter + state.json)
   - Atomic write patterns for state.json

2. **`docs/wiki/workflows/incremental-development.md`**
   - Adding new features (new specs)
   - Fixing failed implementations
   - Refining existing features
   - When to regenerate vs extend
   - Progressive implementation patterns

3. **`docs/wiki/workflows/spec-lifecycle.md`**
   - State transitions (draft → implemented → deployed → validated)
   - Triggers for state changes
   - Failed state handling
   - Deprecated state handling
   - State verification

**Framework Updates:**

4. **`framework/PHASES.md` - Add "Implementation Workflows" section**
   - Link to wiki articles for details
   - High-level workflow overview
   - State handling requirements for agents

5. **`framework/AGENTS.md` - Enhance "Implementation Agents" section**
   - Add state checking directives
   - Define incremental processing behavior
   - Clarify when to skip vs process specs

### Out of Scope

- Implementation of incremental processing logic in agents (separate task)
- Testing workflows (covered by Task 045)
- CLI workflow documentation (user-facing, separate from implementation internals)

## Acceptance Criteria

- [ ] Three new wiki articles created (`implementation-agent-lifecycle.md`, `incremental-development.md`, `spec-lifecycle.md`)
- [ ] `framework/PHASES.md` updated with implementation workflows section
- [ ] `framework/AGENTS.md` updated with state checking directives
- [ ] Workflows cover all five scenarios (initial, incremental, refinement, regeneration, transitions)
- [ ] Documentation distinguishes between "what exists" and "what should exist" (if gap remains)
- [ ] Cross-references between framework and wiki articles
- [ ] Examples use generic placeholders (no specific requirement IDs per Task 043 guidelines)

## Implementation Steps

### Step 1: Research Current State

1. Read all 3 implementation agent files (`smaqit.development.agent.md`, `smaqit.deployment.agent.md`, `smaqit.validation.agent.md`)
2. Identify existing state handling logic
3. Identify gaps in incremental processing
4. Document "as implemented" vs "as designed"

### Step 2: Define Workflows

1. Create state transition diagrams
2. Define decision trees (when to process vs skip)
3. Specify state checking algorithms
4. Define atomic write patterns

### Step 3: Write Wiki Articles

1. Draft `implementation-agent-lifecycle.md` (what agents do)
2. Draft `incremental-development.md` (how users leverage it)
3. Draft `spec-lifecycle.md` (state machine documentation)
4. Review for clarity and completeness

### Step 4: Update Framework

1. Add "Implementation Workflows" to PHASES.md
2. Add state checking directives to AGENTS.md
3. Maintain Level 0 abstraction (directives, not rationale)

### Step 5: Validation

1. Verify documentation matches Task 014 implementation
2. Check cross-references are valid
3. Ensure no example pollution (Task 043 compliance)
4. Get user review before marking complete

## Notes

**Critical Distinction:**

- **Framework files** (AGENTS.md, PHASES.md): Directive-only, what agents MUST/SHOULD do
- **Wiki articles**: Explanatory, why workflows exist, how they work, design rationale

**Design Philosophy:**

Incremental development is core to smaqit's value proposition. Users should be able to:
- Start small (minimal viable specs)
- Add features incrementally (new prompt content → new specs → partial implementation)
- Fix issues without full regeneration (failed state → reprocess → success)
- Maintain progressive validation (validate incrementally added features)

**Related Decisions:**

- Task 014: Chose frontmatter for state (colocation with specs)
- Task 014: Chose state.json for aggregate (CLI efficiency)
- Task 045: Split testing into infrastructure vs incremental phases

**Open Questions:**

1. Should agents always check state, or only in "incremental mode"?
2. How do agents detect "full regeneration requested" vs "incremental update"?
3. Should state checking be mandatory or optional (user preference)?

**These questions should be answered in documentation or flagged for future design decisions.**

## Dependencies

**Blocks:**
- Task 045 Phase 2 (incremental workflow testing needs documented workflows)

**Blocked By:**
- Task 045 Phase 1 (infrastructure testing reveals what's implemented)

**Related:**
- Task 014 (implemented stateful specs, defined requirements)
- Task 031 (review implementation artifacts, may inform workflows)

## Success Metrics

- Developers understand how implementation agents process specs
- Users understand how to leverage incremental development
- Gaps between documentation and implementation are explicit
- Framework files provide clear directives for state handling
- Wiki provides context and rationale for workflow design
