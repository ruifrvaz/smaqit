# Task 069: Strengthen Bounded Agents Principle

**Status:** New  
**Priority:** Medium  
**Created:** 2026-01-19  
**Context:** Discovered during Task 068 when Agent-L0 violated scope boundaries

## Problem Statement

During Task 068 implementation, Agent-L0 violated its scope boundaries by modifying Level 1 templates and Level 2 agents despite having explicit "MUST NOT" directives against this. The agent's directives said "Stop immediately" but external task framing (Task 068 checklist listing framework, agent, and template changes together) created pressure to complete all grouped work.

**Root cause:** The existing "Bounded Agents" principle doesn't explicitly address what happens when external framing (task specifications, documentation, user requests) implies or lists work that spans multiple agent scopes.

## Current State

**Existing "Bounded Agents" principle (SMAQIT.md lines 61-66):**

> "**Bounded Agents:** Agents execute only their designated layer or phase. Unbounded agents lose accountability. Each agent has a single responsibility. Agents decline out-of-scope requests with clear redirection to the appropriate agent. This enforces separation of concerns and prevents scope creep across workflow boundaries."

**Gap:** Doesn't explicitly state that scope enforcement is self-governing and cannot be overridden by external framing.

## Proposed Solution

Strengthen the existing "Bounded Agents" principle to make enforcement implications explicit about external pressure to violate scope.

**Enhanced principle (Level 0 form):**

> "**Bounded Agents:** Agents execute only their designated layer or phase. Unbounded agents lose accountability. Each agent has a single responsibility. Agents decline out-of-scope requests with clear redirection to the appropriate agent. This boundary enforcement is self-governing—external framing, task specifications, or grouped work descriptions cannot override an agent's scope. When requests span boundaries, agents stop at their limit and redirect rather than expand authority. This enforces separation of concerns and prevents scope creep across workflow boundaries."

## Implementation Checklist

### 1. Update Framework Principle

- [ ] Strengthen "Bounded Agents" in `framework/SMAQIT.md` (lines 61-66)
- [ ] Add clarification about self-governing scope enforcement
- [ ] Add statement about external framing not overriding scope
- [ ] Add statement about stopping at boundaries when work spans scopes

### 2. Verify Agent Directives

- [ ] Review all agent scope boundary sections
- [ ] Ensure "Stop immediately" directives are present
- [ ] Verify boundary enforcement language is consistent with strengthened principle

### 3. Update Documentation

- [ ] Add wiki page explaining principle strengthening rationale (optional)
- [ ] Update session history if significant change
- [ ] Note in CHANGELOG if impacts agent behavior expectations

## Alternative Approaches Considered

**Option A: New "Scope Over Specification" principle**
- Pro: Very explicit about external specs not overriding scope
- Con: Creates principle proliferation (13 → 14 principles)
- Con: Concept already covered by "Bounded Agents"

**Option B: New "Bounded Authority" principle**
- Pro: Emphasizes authority limitations
- Con: Similar to "Bounded Agents" conceptually
- Con: Adds complexity without adding new concept

**Option C: Strengthen existing "Bounded Agents" (RECOMMENDED)**
- Pro: Keeps related concepts together
- Pro: Avoids principle proliferation
- Pro: Makes enforcement implications explicit
- Con: Slightly longer principle statement

## Success Criteria

- [ ] "Bounded Agents" principle explicitly addresses external framing
- [ ] Principle states scope is self-governing and not externally negotiable
- [ ] Principle clarifies agents stop at boundaries when work spans scopes
- [ ] Level 0 principle form maintained (philosophical, no directives)
- [ ] Consistent with existing agent scope boundary implementations

## Related Issues

- **Task 068:** Remove System Actor from Business Layer (where scope violation occurred)
- **Session 021:** Agent Scope Boundaries (2025-12-28) - original scope boundary implementation
- **Task 041:** Restrict agents to their layer/phase - scope boundaries in agents

## Notes

This is a principle refinement, not a new principle. The concept of bounded agents exists; this strengthens it to address a discovered edge case where external framing creates pressure to violate scope.

The agent-level directives already exist (Task 041). This task ensures the Level 0 principle explicitly supports those directives.
