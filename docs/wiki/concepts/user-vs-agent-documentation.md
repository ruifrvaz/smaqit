# User vs Agent Documentation

## Definition

User vs agent documentation distinction is the principle that agent-facing specifications contain ONLY execution instructions, while user-facing documentation contains context, rationale, examples, and explanations. This separation enables LLM agents to focus on "what to do" without noise from "why we're doing it."

## The Two Audiences

smaqit documentation serves two fundamentally different audiences with different needs:

### Agent Specifications (Execution Instructions)

**Audience:** LLM agents executing workflows

**Files:**
- `framework/*.md` — Core principles, layer definitions, phase workflows
- `specs/**/*.md` — Specification documents produced by spec agents
- `agents/*.agent.md` — Agent definitions (behavior, directives, completion criteria)
- `templates/**/*.template.md` — Structure templates for specs, agents, prompts

**Purpose:** Enable agents to execute tasks without human context

**Content includes:**
- What must be done
- How to structure output
- What rules to follow (MUST/MUST NOT/SHOULD)
- What to validate before completion
- Where to find input
- Where to write output

**Content excludes:**
- Why decisions were made
- Historical context ("we tried X but it didn't work")
- Human stakeholder names
- Business politics or organizational dynamics
- Delivery dates or project timelines
- Extended examples showing multiple alternatives
- Design rationale or trade-off analysis
- References to past projects or prior art

### User Documentation (Context and Rationale)

**Audience:** Human developers, contributors, users

**Files:**
- `README.md` — Project overview, installation, usage
- `docs/wiki/` — Concepts, designs, patterns, workflows
- `docs/history/` — Session logs documenting work
- `docs/tasks/` — Work items with context and decisions

**Purpose:** Help humans understand, contribute, and make informed decisions

**Content includes:**
- Why the framework is designed this way
- Trade-offs between alternatives
- Historical context and evolution
- Examples with multiple scenarios
- Design rationale and thought process
- Related concepts and references
- Usage patterns and best practices
- Business context when relevant

**Content excludes:**
- Nothing — user docs can contain any information useful to humans

## How It Works

### The Separation in Practice

**Agent spec example (framework/AGENTS.md):**
```markdown
**Specification agents MUST:**
- Produce one specification file per distinct concept
- Include testable acceptance criteria in every specification
- Reference context specs used for coherence and traceability

**Specification agents MUST NOT:**
- Include implementation details
- Create inconsistencies with context layer specifications
- Produce specs for layers outside their scope
```

This tells agents WHAT to do. No explanation of WHY these rules exist.

**User documentation example (docs/wiki/concepts/bounded-agents.md):**
```markdown
## Why Bounded Agents Matter

**Prevents scope creep:**
- Business agent won't start writing functional specs
- Development agent won't start deploying infrastructure
- Each agent stays focused on its responsibility

**Enforces separation of concerns:**
- Business intent separate from functional behavior
- Implementation separate from specification
- Deploy separate from develop
```

This explains WHY the principle matters. Agents don't need this — humans do.

### The Split That Established This Principle

Session 012 (Framework Split) recognized this distinction:

> "Split monolithic SMAQIT.md into focused framework files. This major restructuring improves maintainability and allows agents to load only the context they need."

Then session 027 (Separate Framework from Human Rationale) made it explicit:

> "Create wiki structure (concepts, designs, patterns, workflows) to hold explanations. Keep framework files purely instructional."

This established the pattern we now codify as principle.

## Why This Distinction Matters

### For LLM Agents

**Reduces token count:**
- Agents load only instructions, not explanations
- More tokens available for actual work
- Faster processing and lower costs

**Improves focus:**
- No distraction from human context
- Clear directives without ambiguity
- Reduced risk of agents inferring requirements from examples

**Prevents contamination:**
- Examples in specs become false requirements
- Stakeholder names become hardcoded strings
- Historical context influences current decisions inappropriately

### For Human Contributors

**Clearer contribution model:**
- Framework changes = changing what agents do
- Wiki changes = explaining why it works this way
- No confusion about where information belongs

**Better maintainability:**
- Rationale preserved even when instructions change
- Can update instructions without rewriting explanations
- Historical decisions documented separately from current directives

**Effective onboarding:**
- New contributors read wiki to understand WHY
- Existing contributors read framework to remember WHAT
- Clear separation accelerates learning

## Examples of Inappropriate Content in Specs

### Example 1: Stakeholder Names

**Inappropriate (in spec):**
```markdown
## BUS-LOGIN-001: Login Feature

Sarah from Marketing needs users to log in so we can track engagement metrics 
for the Q4 board presentation. John approved this in the November sprint planning.
```

**Appropriate (in spec):**
```markdown
## BUS-LOGIN-001: User Authentication

**Goal:** Enable registered users to access protected features

**Actors:** Registered User, System

**Success Metric:** 95% of login attempts by valid users succeed within 2 seconds
```

**Human context belongs in:** Task file or session history documenting the requirement source

---

### Example 2: Historical Context

**Inappropriate (in spec):**
```markdown
## FUN-AUTH-001: JWT Token Authentication

We tried OAuth 2.0 in the previous project but it was too complex for our users.
After three failed attempts with Auth0, we decided to roll our own JWT implementation.
This approach worked well in the 2023 product launch.
```

**Appropriate (in spec):**
```markdown
## FUN-AUTH-001: JWT Token Authentication

**Contract:**
- Input: Valid username/password
- Output: JWT token with 24-hour expiration
- Error: 401 Unauthorized if credentials invalid

**Acceptance Criteria:**
- [ ] FUN-AUTH-001: Token expires after 24 hours
- [ ] FUN-AUTH-002: Token includes user ID and role claims
```

**Historical context belongs in:** `docs/wiki/designs/authentication-choice.md` explaining why JWT was selected

---

### Example 3: Delivery Dates and Business Politics

**Inappropriate (in spec):**
```markdown
## INF-DEPLOY-001: Deployment Architecture

Legal requires this deployed by March 15 for GDPR compliance. The CEO wants 
cost under $500/month because the board is concerned about burn rate. 
Marketing needs EU region support for the Germany launch.
```

**Appropriate (in spec):**
```markdown
## INF-DEPLOY-001: Deployment Architecture

**Compute:** Containerized application on managed Kubernetes

**Regions:** EU-West (primary), US-East (secondary)

**Constraints:**
- Data residency: EU user data MUST remain in EU region
- Cost target: Optimize for <$500/month operational cost
- Availability: 99.9% uptime SLA
```

**Business context belongs in:** Task file documenting the requirement source and timeline

---

### Example 4: Extended Examples as False Requirements

**Inappropriate (in template):**
```markdown
## Actors

<!-- Example: "Mario Fan - Users who love Nintendo's Mario franchise" -->

[List your actors here]
```

If agents don't ignore HTML comments, "Mario Fan" becomes a requirement.

**Appropriate (in template):**
```markdown
## Actors

<!-- Example: "[Role name] - [Brief description of their goals]" -->

[List your actors here]
```

Generic placeholder format prevents contamination.

---

### Example 5: "Why" Explanations in Directives

**Inappropriate (in agent definition):**
```markdown
**Stack agents MUST:**
- Document technology choices with rationale

This is important because future maintainers need to understand why React 
was chosen over Vue. In past projects, undocumented choices led to poor 
technology decisions when the original team left.
```

**Appropriate (in agent definition):**
```markdown
**Stack agents MUST:**
- Document technology choices with rationale
- Define language versions and framework versions
- Specify libraries and their purposes
```

**Explanation belongs in:** `docs/wiki/concepts/stack-layer-rationale.md`

## Validation Patterns

### For Agent Specifications

When reviewing framework, template, or agent files, ask:

1. **Is this an instruction?** → Keep it
2. **Is this an explanation of why?** → Move to wiki
3. **Does this include human names/dates/politics?** → Remove or move to task file
4. **Would an agent misinterpret this as a requirement?** → Clarify or remove

### For User Documentation

When reviewing wiki, README, or task files, ask:

1. **Does this help humans understand the system?** → Keep it
2. **Would this be useful for future contributors?** → Keep it
3. **Does this belong in agent specs instead?** → Move if it's a directive

## Trade-offs

**Benefits:**
- **Focused agents** — Agents process only necessary instructions
- **Preserved context** — Human rationale documented without polluting specs
- **Clear contribution model** — Contributors know where information belongs
- **Efficient token usage** — Agents don't load unnecessary context
- **Better maintainability** — Instructions and explanations evolve independently

**Costs:**
- **Duplication risk** — Same concept mentioned in both places
- **Synchronization burden** — Framework changes may require wiki updates
- **Finding information** — Users must know where to look (framework vs wiki)
- **Initial learning curve** — Contributors must learn the distinction

The cost is intentional — separation of concerns at the documentation level enables both agent effectiveness and human understanding without compromise.

## How This Enables LLM Effectiveness

### Token Budget Optimization

Every token matters in LLM context. By excluding human context from agent specs:

- **More room for actual work** — Agent loads instructions + current task, not historical rationale
- **Faster processing** — Less input to parse and consider
- **Lower costs** — Fewer tokens per invocation across thousands of calls

### Reduced Hallucination Risk

Human context creates false signals:

- **Examples become requirements** — "Mario Fan" in an example becomes a hardcoded user type
- **Historical decisions influence current work** — "We tried X and failed" biases against X even when appropriate
- **Stakeholder names become literals** — "Sarah's dashboard" becomes a hardcoded label

Pure instructions eliminate these risks.

### Clear Success Criteria

Agent knows when it's done:

- **No ambiguity** — Completion criteria are explicit, not inferred from context
- **Objective validation** — Agent checks objective directives, not subjective rationale
- **Predictable behavior** — Same instructions produce consistent results

## Relation to Other Principles

**User vs Agent Documentation + Explicit Over Implicit:**
- Explicit Over Implicit: Make decisions explicit, not inferred
- User vs Agent Documentation: Make documentation purpose explicit through separation
- Result: Clear audience definition prevents inappropriate content in either location

**User vs Agent Documentation + Bounded Agents:**
- Bounded Agents: Each agent has single responsibility
- User vs Agent Documentation: Each documentation type has single audience
- Result: Agents focus on execution, humans focus on understanding

**User vs Agent Documentation + Template Constraints:**
- Template Constraints: Templates are mandatory structures
- User vs Agent Documentation: Templates contain execution instructions only
- Result: Template compliance ensures pure execution instructions without rationale pollution

## Migration Guide

If you find human context in agent specifications:

### Step 1: Identify the Content Type

- **Execution instruction** → Keep in framework/agent/template
- **Rationale/context** → Move to wiki
- **Business/human context** → Move to task file or remove

### Step 2: Extract to Appropriate Location

Create or update wiki file with the rationale:

```markdown
# Why Stack Must Document Rationale

Stack layer requires technology choice rationale because:

1. **Future maintainability** — New team members understand decisions
2. **Technology debt tracking** — Rationale reveals when assumptions change
3. **Audit trail** — Compliance requirements may need decision records

This is why Stack agents MUST include rationale in their specs.
```

### Step 3: Simplify Agent Specification

Replace explanation with pure directive:

```markdown
**Stack agents MUST:**
- Document technology choices with rationale
```

### Step 4: Link Between Documents

Framework file references wiki for details:

```markdown
See [Stack Layer Rationale](../../docs/wiki/concepts/stack-layer-rationale.md) 
for the reasoning behind this requirement.
```

## Related

- [Explicit Over Implicit](../designs/explicit-over-implicit.md) — Why we make distinctions explicit
- [Template Constraints](../designs/template-constraints.md) — What templates contain (instructions only)
- [Bounded Agents](bounded-agents.md) — How agents maintain single responsibility
- [Hierarchical Levels](../designs/hierarchical-levels.md) — Kit architecture and level boundaries
