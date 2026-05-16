# Research: Microsoft AI Agent Orchestration Patterns

**Source:** [AI agent orchestration patterns — Azure Architecture Center](https://learn.microsoft.com/en-us/azure/architecture/ai-ml/guide/ai-agent-design-patterns)  
**Published:** February 2026  
**Reviewed:** 2026-05-16  
**Purpose:** Validate smaqit orchestration architecture against industry standards; inform Task 082 design decisions.

---

## Patterns Defined

The guide defines five orchestration patterns for multi-agent systems:

### 1. Sequential Orchestration

Agents process work in a predefined linear order. Each agent receives the output of the previous agent.

Also known as: pipeline, prompt chaining, linear delegation.

**Use when:**
- Multistage processes with clear linear dependencies
- Data transformation pipelines where each stage builds on the previous
- Progressive refinement requirements (draft → review → polish)
- Stages cannot be parallelized
- Workflow progression is predictable

**Avoid when:**
- Stages are embarrassingly parallel
- Only a few stages that a single agent can handle
- Early-stage failures propagate and prevent later stages from recovering
- Agents need to collaborate rather than hand off
- Workflow requires backtracking or dynamic routing

**Key constraint:** The choice of which agent gets invoked next is **deterministically defined** as part of the workflow — it is not a choice given to agents in the process.

---

### 2. Concurrent Orchestration

Multiple agents work on the same input simultaneously. Results are aggregated.

Also known as: parallel, fan-out/fan-in, scatter-gather, map-reduce.

**Use when:**
- Tasks can run in parallel without quality compromise or shared-state contention
- Multiple independent perspectives needed (brainstorming, ensemble reasoning, voting)
- Time-sensitive scenarios where parallel processing reduces latency

**Avoid when:**
- Agents need to build on each other's work (cumulative context required)
- Specific order of operations needed
- No clear conflict resolution strategy for contradictory results
- Resource constraints make parallel processing inefficient

---

### 3. Group Chat Orchestration

Multiple agents participate in a shared conversation thread. A chat manager coordinates turn order.

Also known as: roundtable, collaborative, multiagent debate, council.

**Use when:**
- Collaborative brainstorming where agents build on each other's contributions
- Decision-making that benefits from debate and consensus
- Iterative quality gates (maker-checker loops)
- Human-in-the-loop (HITL) scenarios with real-time oversight

**Avoid when:**
- Basic task delegation or linear pipeline is sufficient
- Real-time processing makes discussion overhead unacceptable
- Clear hierarchical decision-making without discussion is more appropriate
- Chat manager has no objective way to determine task completion

**Practical limit:** Group chat becomes difficult to control with more than 3 agents.

#### Maker-Checker Loop (sub-pattern)

One agent (maker) creates output; another agent or human (checker) evaluates it against defined criteria. If the checker identifies gaps, it returns feedback to the maker for revision. Repeats until approved or iteration cap reached.

Also known as: evaluator-optimizer, generator-verifier, critic loop, reflection loop.

**Requirements:**
- Clear acceptance criteria for the checker
- Explicit iteration cap to prevent infinite refinement
- Defined fallback when cap is reached (escalate to human, or return best-effort with warning)

---

### 4. Handoff Orchestration

Dynamic delegation between specialized agents. Each agent decides whether to handle the task directly or transfer it to a more appropriate agent.

Also known as: routing, triage, transfer, dispatch, delegation.

**Use when:**
- The right agent for a task isn't known upfront
- Expertise requirements emerge during processing
- Multiple-domain problems requiring different specialists, one at a time

**Avoid when:**
- The appropriate agent or sequence is identifiable from initial input (use deterministic routing instead)
- Task routing is deterministic and rule-based
- Multiple operations should run concurrently
- Preventing infinite handoff loops is challenging

---

### 5. Magentic Orchestration

A manager agent dynamically builds and refines a task ledger, invoking specialized agents to gather information and execute tasks. The solution path is not predetermined.

Also known as: dynamic orchestration, task-ledger-based orchestration, adaptive planning.

**Use when:**
- Complex or open-ended problems with no predetermined solution path
- Multiple specialized agents needed to develop a valid plan
- A documented, reviewable plan of approach is required
- Agents interact with external systems (tools that induce changes)

**Avoid when:**
- Solution path is deterministic
- No requirement for a ledger
- Low-complexity task
- Time-sensitive (the pattern prioritizes plan quality over speed)

---

## Pattern Selection Guide

| Pattern | Control | Best for | Key risks |
|---------|---------|----------|-----------|
| Sequential | Deterministic, predefined order | Step-by-step refinement with clear stage dependencies | Failures in early stages propagate; no parallelism |
| Concurrent | Deterministic or dynamic agent selection | Independent analysis from multiple perspectives; latency-sensitive | Conflict resolution when results contradict; resource-intensive |
| Group chat | Chat manager controls turn order | Consensus-building, brainstorming, maker-checker validation | Conversation loops; difficult to control with many agents |
| Handoff | Agents decide when to transfer control | Tasks where the right specialist emerges during processing | Infinite handoff loops; unpredictable routing |
| Magentic | Manager agent assigns and reorders dynamically | Open-ended problems without a predetermined solution path | Slow to converge; stalls on ambiguous goals |

---

## Implementation Considerations

### Start with the Lowest Complexity

Before adopting multiagent orchestration, verify it's necessary:

| Level | Description | Use when |
|-------|-------------|----------|
| Direct model call | Single LLM call, no tools, no agent logic | Classification, summarization, single-step tasks |
| Single agent with tools | One agent with dynamic tool access, can loop | Varied queries within a single domain |
| Multiagent orchestration | Multiple specialized agents coordinating | Cross-functional problems, security boundaries, tasks that benefit from parallel specialization |

Multiagent complexity is justified when a single agent can't reliably handle the task due to **prompt complexity, tool overload, or security requirements**.

### Deterministic vs. Nondeterministic Routing

- Use **deterministic routing** when the sequence of agents is known upfront — hardcode it, don't let agents decide
- Use **nondeterministic** (handoff, magentic) only when the optimal agent can't be identified from initial input
- Antipattern: using nondeterministic patterns for workflows that are inherently deterministic

### Context and State Management

- Context windows grow rapidly in multi-agent orchestrations as each agent adds reasoning, tool results, and intermediate outputs
- Decide what context each downstream agent needs: full raw context, or a scoped/compacted version
- Pass **scoped context** (summary of prior outputs, specific inputs needed) rather than full accumulated context unless the next agent genuinely needs everything
- For long-running tasks, persist shared state externally rather than relying on in-memory context
- Apply compaction (summarization, selective pruning) between agents to stay within model limits

### Reliability

- Implement timeout and retry mechanisms
- Include graceful degradation for agent failures
- Surface errors explicitly — don't hide them from downstream agents or orchestrators
- **Validate agent output before passing to the next agent** — malformed or low-quality responses cascade through pipelines
- Design agents to be isolated from each other with single points of failure not shared
- Use checkpoint features to recover from interrupted orchestrations

### Human Participation (HITL)

- Identify which points require human input: optional observer, mandatory approval gate, or feedback provider
- **Mandatory gates make the orchestration synchronous at that step** — persist state at these checkpoints
- HITL gates can be scoped to specific tool invocations (sensitive operations only) rather than full agent outputs
- In maker-checker loops: human acts as checker; approval advances workflow, feedback loops back to maker

### Common Antipatterns

- Creating coordination complexity when sequential orchestration would suffice
- Adding agents that don't provide meaningful specialization
- Overlooking latency impacts of multi-hop communication
- Sharing mutable state between concurrent agents (transactional inconsistency)
- Using deterministic patterns for inherently nondeterministic workflows (and vice versa)
- Consuming excessive model resources as context windows accumulate

---

## smaqit Alignment Analysis

### Pattern Used: Sequential Orchestration

smaqit's phase agents use sequential orchestration for spec layer coordination:
- Development: business → functional → stack (fixed dependency order)
- Deployment: infrastructure (single agent)
- Validation: coverage (single agent)

**Verdict: Correct pattern.** The guide's "when to use" criteria match exactly: clear linear dependencies (functional needs business for context; stack needs both), progressive refinement (each layer refines requirements into more concrete form), predictable workflow progression.

**Concurrent is explicitly wrong here:** The guide says avoid concurrent when "agents need to build on each other's work or require cumulative context in a specific sequence." Business → Functional → Stack has this exact dependency.

### Assisted Mode: Maker-Checker Loop

smaqit's assisted mode maps to the maker-checker sub-pattern of group chat orchestration:
- Spec agent = maker (produces spec output)
- User = checker (reviews, approves, or provides feedback)
- Loop repeats per spec layer until approved

**Verdict: Correct pattern.** The guide requires explicit iteration caps and defined fallback behavior — this is a gap in the current implementation.

### Gaps Found (relevant to Task 082)

| Gap | Guideline violated | Required fix |
|-----|--------------------|--------------|
| Spec generation routing derived from `smaqit plan` CLI output at runtime | "Routing is deterministically defined, not a choice given to agents" | Hardcode the sequence; use file/status checks to determine which layers to skip |
| Pre-Orchestration Validation halts on missing specs | "Validate input sufficiency, not artifact presence" | Replace artifact-presence check with session context sufficiency check |
| No iteration cap on spec generation retries or assisted-mode review loops | "Set iteration caps; define fallback when cap is reached" | Add max 3-iteration cap per spec layer with escalate-to-user fallback |
| Context passed to subagents is unspecified ("pass session context and layer context") | "Decide what context the next agent requires; scoped context preferred over full accumulation" | Specify: user requirements from session + relevant upstream specs only; not full accumulated phase context |

### Alignment Summary

| smaqit mechanism | Pattern | Alignment |
|---|---|---|
| Business → Functional → Stack → implement | Sequential orchestration | ✅ Correct |
| Develop → Deploy → Validate (phase progression) | Sequential orchestration | ✅ Correct |
| Bounded Agents principle | Specialization | ✅ Correct |
| Assisted mode with user review | Maker-checker loop | ✅ Correct in structure; ⚠️ missing iteration cap |
| Spec routing via `smaqit plan` parsing | Should be deterministic | ❌ Wrong mechanism — must be hardcoded sequence |
| Pre-Orchestration Validation blocks on missing specs | Gate misplacement | ❌ Wrong gate — context sufficiency, not artifact presence |
| Self-Validation before completion | Output validation before downstream handoff | ✅ Direction correct |
