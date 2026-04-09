# Restrict Agents to Their Layer/Phase

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2025-12-28

## Description

Agents should be explicitly forbidden from executing work belonging to other layers or phases. Currently, agent definitions lack clear boundary enforcement rules, which could lead to agents attempting to generate specs or run implementations outside their designated responsibility.

## Acceptance Criteria

**Level 0 (AGENTS.md):**
- [x] Add "Scope Boundaries" section with generic MUST NOT rules using placeholders
- [x] Add "Boundary Enforcement" pattern: Stop → Respond → Suggest
- [x] Use placeholders: `[PHASE]`, `[OTHER_PHASES]`, `[OTHER_AGENTS]`, `[LAYER]`, `[OTHER_LAYERS]`

**Level 1 (Agent Templates):**
- [x] Specification agent template includes MUST NOT section with Develop/Deploy/Validate phase prohibition
- [x] Specification agent template includes other layer prohibition with `[OTHER_LAYERS]` placeholder
- [x] Implementation agent template includes MUST NOT section appropriate for impl agents
- [x] Both templates include enforcement pattern with populated placeholders

**Level 2 (Agent Definitions):**
- [x] All 5 specification agents have layer-specific MUST NOT rules (Business, Functional, Stack, Infrastructure, Coverage)
- [x] All 3 implementation agents have phase-specific MUST NOT rules (Development, Deployment, Validation)
- [x] Each agent explicitly lists what phases/layers they cannot execute
- [x] Enforcement responses are contextual to each agent's scope

**Validation:**
- [x] No placeholders remain in Level 2 agents
- [x] Each agent clearly states its single responsibility
- [x] Boundary violations result in helpful redirection, not execution

## Impact

**Severity:** Medium  
**User Impact:** Prevents agents from overstepping boundaries; improves workflow clarity; enforces separation of concerns

## Approach

**Level 0 (AGENTS.md) — Generic principles:**
```markdown
## Scope Boundaries

### MUST NOT
- **Execute work assigned to [OTHER_PHASES]**
- **Execute work assigned to [OTHER_AGENTS]**

## Boundary Enforcement

When user requests out-of-scope work:
- **Stop immediately** — Do not plan, create todos, or execute
- **Respond clearly**: "[PHASE] is [STATUS]. To proceed with [REQUESTED_WORK], invoke [TARGET_AGENT]."
- **Suggest next step**: Provide prompt file or agent invocation
```

**Level 1 (specification-agent.template.md) — Template with placeholders:**
```markdown
### MUST NOT
- **Execute work assigned to Development, Deploy, or Validate phases**
- **Execute work assigned to other specification layers ([OTHER_LAYERS])**

## Boundary Enforcement

When user requests implementation or other layer specs:
- **Stop immediately** — Do not plan, create todos, or execute
- **Respond**: "[Layer] specification is complete. To proceed with [requested work], invoke [target agent]."
```

**Level 2 (smaqit.business.agent.md) — Populated example:**
```markdown
### MUST NOT
- **Execute work assigned to Development, Deploy, or Validate phases**
- **Execute work assigned to Functional, Stack, Infrastructure, or Coverage specification layers**

## Boundary Enforcement

When user requests implementation or other specs:
- **Stop immediately** — Do not plan, create todos, or execute
- **Respond**: "Business specification is complete. To proceed with [requested work], invoke the appropriate agent."
```

## Notes

**Refinement considerations:**
- Placeholders enable template inheritance from L0 → L1 → L2
- Each level distills from generic (L0) to template (L1) to specific (L2)
- Enforcement pattern preserves UX (helpful redirection, not silent failure)
- Aligns with task 039 (agent handover guidance)
- Preserves separation of concerns across all 4 levels while remaining maintainable

## Implementation Summary

**Files Modified:**
- **Level 0 (1 file):** `framework/AGENTS.md` - Added Scope Boundaries section after Self-Validation Before Completion
- **Level 1 (2 files):** 
  - `templates/agents/specification-agent.template.md` - Added Scope Boundaries with placeholders for other layers
  - `templates/agents/implementation-agent.template.md` - Added Scope Boundaries with placeholders for other phases
- **Level 2 (8 files):** All specification agents (business, functional, stack, infrastructure, coverage) and implementation agents (development, deployment, validation)

**Key Features:**
- Each agent now has a "Scope Boundaries" section defining what work is out of scope
- 3-step enforcement pattern: Stop → Respond → Suggest
- Contextual responses include agent status and suggest appropriate next agent
- Specification agents prohibited from: implementation phases + other layer specs
- Implementation agents prohibited from: other phases + specification work

**Validation:**
- ✓ Installer builds successfully (version f0a295f)
- ✓ CLI commands work (init, status, validate)
- ✓ Agents installed with scope boundaries intact
- ✓ No placeholders remain in Level 2 agents (grep verified)
