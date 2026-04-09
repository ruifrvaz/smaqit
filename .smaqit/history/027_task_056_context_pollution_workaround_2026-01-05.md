# Session 020: Task 056 - Document Context Pollution Workaround

**Date:** 2026-01-05  
**Task:** 056 - Document context pollution workaround  
**Status:** Completed

## Session Overview

Implemented two-pronged mitigation strategy for context pollution in multi-agent workflows. Added explicit layer identity statements to all specification agents and created comprehensive troubleshooting documentation to guide users when context carryover occurs.

## Problem Statement

When users invoke multiple layer agents in the same GitHub Copilot session (e.g., `/smaqit.business` → `/smaqit.functional` → `/smaqit.stack`), agents may retain context from previous invocations, causing mode confusion where agents reference inappropriate context from other layers.

**Root Cause:** GitHub Copilot maintains session state across prompt invocations, causing context carryover—a platform limitation, not a smaqit bug.

**Impact:**
- Agents reference previous layer's context inappropriately
- Mode confusion could lead to cross-layer contamination
- Users must manually verify agent is operating in correct mode
- Adds UX friction to multi-layer workflows

**Severity:** Low—doesn't block functionality but affects user experience.

## Solution Architecture

### Two-Pronged Mitigation Strategy

**1. Agent-level mitigation:** Explicit layer identity statements reduce confusion
- Each agent states its layer at response start
- Agents explicitly ignore context from other layers
- Creates defensive programming layer within agents

**2. User-level mitigation:** Documentation provides workaround when needed
- Clear guidance on when to start fresh sessions
- Symptoms to watch for
- Future solution roadmap

**Why both approaches:** Agent awareness reduces confusion and provides explicit layer context, while documentation gives users control when confusion still occurs. Combined approach provides defense-in-depth.

## Implementation

### Phase 1: Agent Improvements

**Files modified (5):**
- `agents/smaqit.business.agent.md`
- `agents/smaqit.functional.agent.md`
- `agents/smaqit.stack.agent.md`
- `agents/smaqit.infrastructure.agent.md`
- `agents/smaqit.coverage.agent.md`

**Changes per agent:**

1. **Added Agent Awareness section** after Role section:
   ```markdown
   ## Agent Awareness
   
   **Layer Identity:** This agent operates in the **[LAYER_NAME]** layer.
   
   **MUST at start of every response:**
   - State: "I am the [LAYER_NAME] Agent, operating in [LAYER_NAME] layer mode."
   - Acknowledge only context relevant to this layer
   - Explicitly ignore context from other layers if carried over from previous session
   
   **Example opening:**
   "I am the [LAYER_NAME] Agent, operating in [LAYER_NAME] layer mode. I will generate [layer] specifications based on the [layer] prompt file and [upstream context]. I will not reference [downstream layers] concerns from any previous context."
   ```

2. **Updated MUST directives** to include layer identity statement requirement:
   - Added: "State layer identity at the start of every response: 'I am the [LAYER_NAME] Agent, operating in [LAYER_NAME] layer mode.'"

3. **Updated MUST NOT directives** to explicitly forbid context carryover:
   - Added: "Reference or carry over context from other layer agents executed in the same session"

**Pattern Consistency:** All 5 agents follow identical structure with layer-specific content substitution. This ensures uniform behavior across the specification agent suite.

### Phase 2: Documentation

**File created (1):**
- `docs/wiki/troubleshooting.md`

**Content structure:**

1. **Issue Description**
   - Clear problem statement
   - When it occurs
   - Observable symptoms
   - Root cause explanation

2. **Workaround**
   - Recommended approach (fresh chat session)
   - When fresh session is recommended
   - When fresh session is required
   - Step-by-step instructions

3. **Agent Awareness Feature**
   - Explanation of v0.5.0 agent improvements
   - Expected agent behavior (layer identity statements)
   - How this reduces confusion

4. **Future Solution**
   - Orchestrator agent pattern (v0.6.0)
   - How it will solve the issue properly
   - Context isolation approach

5. **Getting Help**
   - Links to issues tracker
   - Reference to main documentation
   - How to report problems

**Documentation philosophy:** Written for human users (not agents), provides context and rationale, includes actionable guidance, mentions future improvements without overpromising.

## Key Decisions

### 1. Agent Awareness vs Context Clearing

**Decision:** Add agent awareness statements, don't attempt programmatic context clearing.

**Alternatives considered:**
- Programmatic context clearing via API
- Agent awareness only
- User guidance only

**Rationale:**
- GitHub Copilot API doesn't support programmatic context clearing
- Agent awareness provides defensive layer without relying on unavailable features
- Combined with user guidance, addresses issue within current platform constraints

### 2. Troubleshooting Location

**Decision:** Create `docs/wiki/troubleshooting.md` as standalone file.

**Alternatives considered:**
- Add section to README.md
- Add notes to individual prompt files
- Create separate troubleshooting guide per issue

**Rationale:**
- Keeps README.md focused on getting started
- Centralized troubleshooting benefits all users
- Expandable structure for future common issues
- Aligns with wiki structure (concepts, workflows, patterns, troubleshooting)

### 3. Directive Placement

**Decision:** Add layer identity statement to MUST section, context carryover prohibition to MUST NOT section.

**Alternatives considered:**
- Add to Agent Awareness section only
- Add to Role section as narrative
- Create separate "Context Management" section

**Rationale:**
- MUST/MUST NOT sections are actionable directives agents follow strictly
- Agent Awareness section provides context and example
- Both are needed: explanatory section + enforcement directives
- Aligns with existing agent structure patterns

## Files Modified Summary

**Agent files (5 modified):**
- Added Agent Awareness section with layer identity and context isolation directives
- Updated MUST directives to require layer identity statement
- Updated MUST NOT directives to forbid context carryover

**Documentation (1 created):**
- Created comprehensive troubleshooting guide with context pollution workaround

**Task tracking (2 modified):**
- `docs/tasks/056_document_context_pollution_workaround.md` — Marked completed with summary
- `docs/tasks/PLANNING.md` — Moved task 056 from Active to Completed

**Total: 8 files modified/created**

## Validation

### Agent Consistency Check

Verified all 5 agents follow identical pattern:
- ✓ Agent Awareness section present after Role
- ✓ Layer-specific identity statement included
- ✓ Example opening demonstrates expected behavior
- ✓ MUST directive includes identity statement requirement
- ✓ MUST NOT directive forbids context carryover

### Documentation Quality Check

Verified troubleshooting documentation:
- ✓ Clear problem description with symptoms
- ✓ Actionable workaround steps
- ✓ Guidance on when workaround is needed
- ✓ Reference to agent awareness feature
- ✓ Future solution mentioned (v0.6.0 orchestrator)
- ✓ Getting help section with links

### Level Hierarchy Respect

Confirmed changes respect smaqit levels:
- **Level 0 (Framework):** No changes needed—principles already address boundaries
- **Level 1 (Templates):** No changes needed—agent behavior, not template structure
- **Level 2 (Agents):** Updated 5 specification agents ✓
- **Documentation:** Created troubleshooting guide ✓

### Task Completion

- ✓ All acceptance criteria met
- ✓ Task file marked completed with summary
- ✓ PLANNING.md updated (task moved to Completed)
- ✓ Session history documented

## Outcome

Successfully implemented defense-in-depth approach to context pollution:

1. **Agent-level:** Explicit layer identity statements at response start reduce confusion
2. **User-level:** Clear troubleshooting guidance provides workaround when needed

**User benefit:** Improved agent clarity and explicit guidance when context issues occur.

**Platform limitation acknowledged:** Cannot fully solve GitHub Copilot's context persistence, but mitigations reduce impact.

**Future path:** Orchestrator agent pattern (v0.6.0) will properly address with isolated contexts.

## Lessons Learned

### 1. Platform Limitations Require Creative Solutions

When platform features (like context clearing) aren't available, combination of defensive programming (agent awareness) and user guidance (documentation) can effectively mitigate issues.

### 2. Defense-in-Depth for UX

Addressing user experience issues at multiple levels (agent behavior + documentation) provides better coverage than single-layer solutions. Users benefit from both automated prevention and manual workarounds.

### 3. Documentation as First-Class Artifact

Troubleshooting documentation has value beyond immediate issue. Creates foundation for future common issues, improves discoverability, and demonstrates commitment to user support.

### 4. Consistent Patterns Across Agents

Using identical structure for agent awareness across all 5 specification agents ensures uniform behavior and reduces maintenance burden. Template-like consistency is valuable even at agent level.

### 5. Explicit Over Implicit in Agent Behavior

Making agents explicitly state their layer at response start may seem verbose, but clarity trumps brevity when preventing user confusion. The extra 2-3 lines per agent provide significant defensive value.

## Related Tasks

- **Task 048:** E2E Agent Workflow Testing (discovered this issue during testing)
- **Future:** Orchestrator agent pattern (v0.6.0) will eliminate need for workaround

## Session Metrics

- **Duration:** ~45 minutes
- **Tasks completed:** 1 (Task 056)
- **Files modified:** 5 agent files + 2 task files
- **Files created:** 1 troubleshooting guide + 1 session history
- **Total changes:** 8 files
- **Agent behavior changes:** 5 specification agents with awareness sections
- **Documentation additions:** First troubleshooting guide in wiki

## Next Steps

Task 056 complete. No follow-up work required unless users report issues with the mitigation strategy or request additional troubleshooting topics.

**Future enhancement (v0.6.0):** Implement orchestrator agent pattern for proper context isolation.
