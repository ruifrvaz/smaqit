# Task 056: Document Context Pollution Workaround

**Status:** new  
**Priority:** Low  
**Created:** 2026-01-05  
**Related:** Task 048 (E2E Testing), Issue 1

## Problem

When switching between layer agents in the same Copilot session (e.g., `/smaqit.business` → `/smaqit.functional` → `/smaqit.stack`), agents retain context from previous layer invocations and may not fully recognize they are operating in a new agent mode.

**Impact:**
- Agents may reference previous layer's context inappropriately
- Mode confusion could lead to cross-layer contamination
- User must manually verify agent is operating in correct mode
- Adds UX friction to multi-layer workflows

**Root Cause:** GitHub Copilot maintains session state across prompt invocations, causing context carryover.

**Severity:** Medium—doesn't block functionality but affects user experience.

## Objective

1. **Improve agent awareness**: Update all specification agents to explicitly state their layer at execution start
2. **Document workaround**: Provide clear guidance for users on when to start fresh Copilot sessions

## Acceptance Criteria

### Agent Improvements
- [ ] Updated all 5 specification agents to explicitly state layer at start of execution
- [ ] Added directive: "Begin response by stating: 'I am the [LAYER_NAME] Agent, operating in [LAYER] layer mode.'"
- [ ] Agents explicitly ignore context from other layers in their response
- [ ] Pattern consistent across Business, Functional, Stack, Infrastructure, Coverage agents

### Documentation
- [ ] Created troubleshooting section in README.md or docs/wiki/troubleshooting.md
- [ ] Documented context pollution issue with clear description
- [ ] Documented workaround: Start fresh Copilot chat session between layers
- [ ] Provided guidance on when fresh session is recommended vs required
- [ ] Mentioned future solution: Orchestrator agent pattern (v0.6.0)
- [ ] Optional: Added note to agent prompt files reminding users of this limitation

## Implementation Plan

### Phase 1: Agent Improvements

1. **Update all specification agents** (Business, Functional, Stack, Infrastructure, Coverage):
   
   Add to each agent's "Role" section or create new "Agent Awareness" section:
   
   ```markdown
   ## Agent Awareness
   
   **Layer Identity:** This agent operates in the [LAYER_NAME] layer.
   
   **MUST at start of every response:**
   - State: "I am the [LAYER_NAME] Agent, operating in [LAYER] layer mode."
   - Acknowledge only context relevant to this layer
   - Explicitly ignore context from other layers if carried over from previous session
   
   **Example opening:**
   "I am the Business Agent, operating in business layer mode. I will generate business specifications based on the business prompt file and will not reference functional, stack, or infrastructure concerns from any previous context."
   ```

2. **Files to modify:**
   - `agents/smaqit.business.agent.md`
   - `agents/smaqit.functional.agent.md`
   - `agents/smaqit.stack.agent.md`
   - `agents/smaqit.infrastructure.agent.md`
   - `agents/smaqit.coverage.agent.md`

3. **Directive additions:**
   - Add to MUST section: "State layer identity at the start of every response"
   - Add to MUST NOT section: "Reference or carry over context from other layer agents executed in the same session"

### Phase 2: Documentation

4. **Create or update troubleshooting documentation:**
   - Location options:
     - Add "Troubleshooting" section to `README.md` (if space permits)
     - Create `docs/wiki/troubleshooting.md` (preferred for detail)
   - Content structure:
     - Issue description
     - When it occurs
     - Observed symptoms
     - Workaround steps
     - Future solution mention
## Files to Modify

### Agent Files
- `agents/smaqit.business.agent.md` (add agent awareness section)
- `agents/smaqit.functional.agent.md` (add agent awareness section)
- `agents/smaqit.stack.agent.md` (add agent awareness section)
- `agents/smaqit.infrastructure.agent.md` (add agent awareness section)
- `agents/smaqit.coverage.agent.md` (add agent awareness section)

### Documentation Files
- `docs/wiki/troubleshooting.md` (create new file, preferred) OR
- `README.md` (add troubleshooting section, if concise enough)
- Optional: All 5 specification prompt files in `.github/prompts/` (add HTML comment note)
   **Description:** When invoking multiple layer agents in the same Copilot session (e.g., `/smaqit.business` → `/smaqit.functional`), agents may retain context from previous invocations. This can cause mode confusion where agents reference inappropriate context from other layers.
   
   **When it occurs:** Multi-layer specification workflows where user invokes different agents sequentially in same Copilot chat session.
   
   **Symptoms:**
   - Agent mentions awareness of previous layer's execution
   - Agent references context from different layer inappropriately
   - Uncertainty about which agent mode is active
   
   **Workaround:**
   - **Recommended:** Start fresh Copilot chat session between layer invocations
   - Click "New Chat" in Copilot panel before invoking next layer agent
   - Context clearing ensures each agent operates in clean mode
   
   **When fresh session recommended:**
   - Always recommended between specification layers (Business → Functional → Stack)
   - Especially important when switching between phases (Phase 1 → Phase 2 → Phase 3)
   
   **When fresh session required:**
   - If agent exhibits confusion about current mode
   - If agent references inappropriate context from previous invocation
   
   **Future solution:** v0.6.0 will introduce orchestrator agent pattern (`/smaqit.orchestrator`) that invokes sub-agents with isolated contexts, eliminating this issue.

3. **Optional: Update prompt files:**
   - Add comment to specification prompt files (business, functional, stack, infrastructure, coverage)
## Testing

### Agent Improvements Testing

**Manual verification:**
1. Read updated agent definitions
2. Confirm layer identity statement is clear
3. Verify directive consistency across all 5 agents

**Agent behavior testing:**
1. Start Copilot session, invoke `/smaqit.business`
2. Verify agent starts response with "I am the Business Agent, operating in business layer mode"
## Estimated Effort

- **Agent improvements:** 1 hour (5 agents × ~12 minutes each)
- **Documentation:** 30 minutes
- **Total:** 1.5 hours if agent explicitly ignores Business layer context or references it appropriately
6. Confirm reduced confusion about active layer

### Documentation Testing

**Manual verification:**
1. Read troubleshooting documentation
2. Confirm issue description is clear
3. Confirm workaround steps are actionable
4. Verify future solution mention provides context

**User testing (optional):**
1. Share documentation with test users
2. Collect feedback on clarity and usefulness
1. Read documentation
2. Confirm issue description is clear
3. Confirm workaround steps are actionable
4. Verify future solution mention provides context

**User testing (optional):**
1. Share documentation with test users
2. Collect feedback on clarity and usefulness

## Estimated Effort

30 minutes

## Dependencies

**Approaches implemented:**
1. **Agent awareness:** Update agents to explicitly state "I am now operating in [LAYER] mode"—**NOW IMPLEMENTING** as primary mitigation
2. **User guidance:** Document workaround with fresh session recommendation—**IMPLEMENTING** as secondary mitigation
3. **Context clearing:** Investigate if agents can programmatically clear session context—not possible with current Copilot API (deferred)

**Why both approaches:** Agent awareness reduces confusion and provides explicit layer context, while documentation gives users control when confusion still occurs. Combined approach provides defense-in-depth.
None (documentation-only task)

## Related Tasks

- Task 048: E2E Agent Workflow Testing (discovered this issue)
**Mitigation strategy:** This is a GitHub Copilot platform limitation, not a smaqit bug. We're implementing a two-pronged approach:
1. **Agent-level mitigation:** Explicit layer identity statements reduce confusion
2. **User-level mitigation:** Documentation provides workaround when needed

## Notes

**Why document rather than fix:** This is a GitHub Copilot platform limitation, not a smaqit bug. Documenting workaround is pragmatic solution until orchestrator pattern is implemented.

**User impact:** Low friction—users can work around by starting fresh sessions. Does not block workflows, just requires extra click.

**Future solution:** Orchestrator agent pattern (planned for v0.6.0) will enable single invocation that coordinates all sub-agents with isolated contexts. This is proper architectural solution but requires significant development effort.

**Alternative workarounds considered:**
1. **Agent awareness:** Update agents to explicitly state "I am now operating in [LAYER] mode"—rejected as incomplete solution
2. **Context clearing:** Investigate if agents can programmatically clear session context—not possible with current Copilot API
3. **User guidance only:** Document workaround (current approach)—pragmatic and sufficient for v0.5.0

**Troubleshooting section value:** Adding troubleshooting documentation benefits all users, not just for context pollution. Can include other common issues in future (spec frontmatter errors, YAML parsing, etc.).
