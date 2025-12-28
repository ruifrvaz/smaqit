# Restrict Agents to Their Layer/Phase

**Status:** Not Started  
**Created:** 2025-12-28

## Description

Agents should be explicitly forbidden from executing work belonging to other layers or phases. Currently, agent definitions lack clear boundary enforcement rules, which could lead to agents attempting to generate specs or run implementations outside their designated responsibility.

## Acceptance Criteria

**Level 0 (AGENTS.md):**
- [ ] Add "Scope Boundaries" section with generic MUST NOT rules using placeholders
- [ ] Add "Boundary Enforcement" pattern: Stop → Respond → Suggest
- [ ] Use placeholders: `[PHASE]`, `[OTHER_PHASES]`, `[OTHER_AGENTS]`, `[LAYER]`, `[OTHER_LAYERS]`

**Level 1 (Agent Templates):**
- [ ] Specification agent template includes MUST NOT section with Develop/Deploy/Validate phase prohibition
- [ ] Specification agent template includes other layer prohibition with `[OTHER_LAYERS]` placeholder
- [ ] Implementation agent template includes MUST NOT section appropriate for impl agents
- [ ] Both templates include enforcement pattern with populated placeholders

**Level 2 (Agent Definitions):**
- [ ] All 5 specification agents have layer-specific MUST NOT rules (Business, Functional, Stack, Infrastructure, Coverage)
- [ ] All 3 implementation agents have phase-specific MUST NOT rules (Development, Deployment, Validation)
- [ ] Each agent explicitly lists what phases/layers they cannot execute
- [ ] Enforcement responses are contextual to each agent's scope

**Validation:**
- [ ] No placeholders remain in Level 2 agents
- [ ] Each agent clearly states its single responsibility
- [ ] Boundary violations result in helpful redirection, not execution

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
