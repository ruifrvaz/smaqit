# Bounded Agents

## Definition

Bounded agents are agents that execute only their designated layer or phase, declining out-of-scope requests with clear redirection. Each agent has a single responsibility—one specification layer or one implementation phase—and refuses to cross boundaries.

## How It Works

Agents enforce their boundaries through explicit scope definitions and enforcement patterns:

**Specification agents** (Business, Functional, Stack, Infrastructure, Coverage):
- Generate specifications for their layer only
- MUST NOT execute implementation work (coding, deployment, testing)
- MUST NOT generate specifications for other layers
- Decline implementation requests → redirect to implementation agents
- Decline other layer requests → redirect to appropriate specification agent

**Implementation agents** (Development, Deployment, Validation):
- Execute their phase only
- MUST NOT generate specifications
- MUST NOT execute other phases
- Decline specification requests → redirect to specification agents
- Decline other phase requests → redirect to appropriate implementation agent

## The Boundary Enforcement Pattern

```
1. User requests work
2. Agent checks: Is this my layer/phase?
3. If YES → proceed with work
4. If NO → Stop immediately
5. Respond: "[Agent] handles [scope]. To proceed with [requested work], invoke [target agent]."
6. Suggest: Provide prompt file or agent invocation command
```

Agents don't attempt out-of-scope work—they redirect.

## Why Bounded Agents Matter

**Prevents scope creep:**
- Business agent won't start writing functional specs
- Development agent won't start deploying infrastructure
- Each agent stays focused on its responsibility

**Enforces separation of concerns:**
- Business intent separate from functional behavior
- Implementation separate from specification
- Deploy separate from develop

**Maintains workflow clarity:**
- Users know which agent to invoke for which task
- No ambiguity about agent responsibilities
- Clear handoff points between agents

**Enables parallel work:**
- Different teams can work on different layers independently
- Infrastructure team doesn't block functional design
- Testing team can specify coverage before implementation starts

## Example: Business Agent Boundary

User invokes Business agent and requests: "Create business spec AND write the Python code"

**Without boundary enforcement:**
```
❌ Agent proceeds:
   - Generates business spec
   - Generates Python code
   - Mixes concerns (intent + implementation)
   - Violates Specs Before Code principle
```

**With boundary enforcement:**
```
✅ Agent responds:
   "Business agent generates business specifications only.
    
    I've created the business spec at specs/business/uc1-user-login.md
    
    To implement this specification, invoke /smaqit.development after 
    completing the Functional and Stack specifications."
```

Agent stays in bounds, redirects user to correct workflow.

## Example: Development Agent Boundary

User invokes Development agent and requests: "Deploy to production"

**Without boundary enforcement:**
```
❌ Agent proceeds:
   - Attempts to deploy application
   - Skips infrastructure specification
   - Violates phase workflow (Develop → Deploy → Validate)
```

**With boundary enforcement:**
```
✅ Agent responds:
   "Development phase is complete. Application built successfully.
    
    To deploy this application, you must:
    1. Create infrastructure specification with /smaqit.infrastructure
    2. Run deployment phase with /smaqit.deployment
    
    Production deployment requires infrastructure specs."
```

Agent protects workflow integrity by refusing premature deployment.

## The Scope Declaration

Every agent declares its scope explicitly:

**Business Agent:**
```markdown
## Scope

**Layer**: Business  
**Phase**: Develop (specification only)  
**Responsibility**: Generate business requirements and use cases

### MUST NOT
- Execute work assigned to Development, Deploy, or Validate phases
- Execute work assigned to Functional, Stack, Infrastructure, or Coverage layers
```

**Development Agent:**
```markdown
## Scope

**Phase**: Develop (implementation)  
**Responsibility**: Implement application from Business, Functional, Stack specifications

### MUST NOT
- Execute work assigned to Deploy or Validate phases
- Generate specifications (that's specification agents' job)
```

Scope declarations make boundaries explicit, not implicit.

## Boundary Violations Users Might Attempt

**Common scenarios:**

| User Request | Violates Boundary | Correct Response |
|--------------|-------------------|------------------|
| "Business agent: write the code too" | Spec agent → Implementation | Redirect to `/smaqit.development` |
| "Development agent: deploy this" | Develop phase → Deploy phase | Redirect to `/smaqit.deployment` |
| "Stack agent: add these business requirements" | Stack layer → Business layer | Redirect to `/smaqit.business` |
| "Functional agent: run the tests" | Spec agent → Implementation | Redirect to `/smaqit.validation` |

Agents don't accommodate requests—they enforce workflow.

## Why Not Multi-Scope Agents?

**Alternative approach**: Single agent handles multiple layers/phases

**Problems:**
- **Prompt pollution**: Agent definition contains instructions for all scopes, increasing token count unnecessarily
- **Cognitive overhead**: Agent must decide which mode to operate in
- **Reduced specialization**: Each scope gets less context/guidance
- **Workflow confusion**: Unclear when one scope ends and another begins
- **Harder to parallelize**: Can't invoke multiple agents simultaneously

**smaqit choice**: Specialized agents with single responsibility

**Benefits:**
- Clear invocation model (one agent = one scope)
- Smaller, focused agent definitions
- Better prompt engineering (specialized instructions)
- Explicit workflow steps (Business → Functional → Stack)
- Parallelizable (run Stack and Infrastructure specs concurrently)

## Trade-offs

**Benefits:**
- Clear separation of concerns
- Reduced scope creep
- Workflow integrity maintained
- Easier to understand agent responsibilities
- Better error messages (agent knows what it can't do)

**Costs:**
- More agents to invoke (8 agents instead of 1)
- More handoffs between agents
- User must understand which agent to invoke when
- Requires redirection logic in every agent

The cost is intentional—workflow clarity is worth the invocation overhead.

## Relation to Other Principles

**Bounded Agents + Layer Independence:**
- Layer Independence: Each layer's prompt is the sole source of requirements
- Bounded Agents: Each agent generates specifications for one layer only
- Result: Agents can't create false derivation chains by generating multiple layers

**Bounded Agents + Specs Before Code:**
- Specs Before Code: Implementation requires specification
- Bounded Agents: Specification agents can't write code, implementation agents can't skip specs
- Result: Agents enforce the principle through boundaries

**Bounded Agents + Progressive Refinement:**
- Progressive Refinement: Each layer addresses a distinct concern
- Bounded Agents: Each agent addresses one layer/phase only
- Result: Agents embody the layered architecture

Bounded Agents is the enforcement mechanism for smaqit's architectural principles.

## Related

- [Layer Independence](layer-independence.md) — Why layers don't derive from each other
- [Self-Validating Agents](self-validating-agents.md) — What agents validate before completion
- [Phase Workflows](../workflows/phase-workflows.md) — When to invoke which agent
- [Agent Handover](../patterns/agent-handover.md) — How agents redirect users to next step
