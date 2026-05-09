# Layer Independence

## Definition

Layer independence is the principle that each specification layer receives its requirements from session context (user input in chat, compacted context blocks, or open tasks), not from upstream layers. Upstream layers provide context for coherence and traceability, but they do not define requirements for downstream layers.

## How It Works

When a layer's agent is invoked, requirements come from the current session context:

| Layer | Requirements Source |
|-------|---------------------|
| Business | Session context: stakeholder goals, use cases, success criteria |
| Functional | Session context: experience shape, behaviors, interactions |
| Stack | Session context: technology preferences, constraints, expertise |
| Infrastructure | Session context: deployment requirements, hosting, scaling |
| Coverage | Session context: test scope, environment, integration points, thresholds |

Session context is the **sole source of requirements** for each layer. Upstream specs are read for context, not copied as requirements.

## Why Independence Matters

**Prevents false derivation chains:**
- Business requirement change doesn't automatically cascade to Functional
- Users must explicitly invoke each layer's agent when requirements change
- Changes are traceable to specific layer modifications

**Enables clear change tracking:**
- Each agent invocation captures explicit user intent for that layer
- No ambiguity about whether business changed or someone's interpretation changed
- Explicit user intent at every layer

**Supports layer swapping:**
- Can change Stack without touching Business or Functional
- Can add new Infrastructure without modifying application layers
- Each layer can evolve independently when requirements for that layer change

## Example

**Scenario:** Stakeholder adds new business requirement

**With layer independence:**
1. User invokes Business agent with updated requirements in chat
2. Business agent regenerates business spec
3. Functional spec unchanged (user hasn't invoked functional agent with changes)
4. If functional changes needed, user explicitly invokes Functional agent with updated requirements

**Without layer independence (hypothetical):**
1. User updates business requirements somewhere
2. Functional agent "derives" new functional requirements
3. Unclear if functional change was intended or agent interpretation
4. Silent cascading changes through all layers

## Coherence vs Requirements

Layers still reference upstream specs, but for different purposes:

| Purpose | How It Works |
|---------|--------------|
| **Requirements** | From session context only |
| **Coherence** | Check compatibility with upstream specs |
| **Traceability** | Link coverage back through all layers |

Implementation agents consolidate specs from multiple layers and validate coherence before execution.

## Related

- [Progressive Refinement](../designs/progressive-refinement.md) — Why layers are structured this way
- [Layer References Upstream](../designs/layer-references-upstream.md) — Why layers reference despite independence
