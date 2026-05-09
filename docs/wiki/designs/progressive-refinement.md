# Progressive Refinement

## Overview

smaqit architecture is built on progressive refinement—each layer addresses a distinct concern while maintaining coherence with adjacent layers. This design decision ensures separation of concerns while enabling traceability.

## The Five-Layer Model

```
Business (intent) | Functional (behavior) | Stack (tools) | Infrastructure (environment) | Coverage (verification)
```

Each layer answers a specific question:
- **Business**: Why are we building this? (stakeholder value)
- **Functional**: What behaviors does it have? (system contracts)
- **Stack**: With what tools do we build it? (technology choices)
- **Infrastructure**: Where does it run? (operational environment)
- **Coverage**: How do we verify it works? (testing strategy)

## Why Layers Are Independent

**Layer independence prevents false dependency chains.** If Functional derived requirements from Business, changes to Business would cascade automatically to Functional. This creates confusion about what actually changed:

- Did the business requirement change?
- Or did someone's interpretation of the business requirement change?

With independent layers (each receiving requirements from its own agent invocation in session context), changes are explicit. When Business changes, Functional doesn't automatically change—users must explicitly invoke the Functional agent with updated requirements if they want different behaviors.

## Why Layers Reference Upstream

Despite independence, layers reference upstream specs for:

1. **Coherence validation**: Implementation agents check that specs across layers are compatible before execution
2. **Traceability**: Coverage can trace requirements through all layers to ensure complete verification
3. **Context**: Agents can understand the full picture while respecting layer boundaries

References provide context, not requirements. Session context remains the sole source of requirements for each layer.

## Why This Order

The layer order (Business → Functional → Stack → Infrastructure → Coverage) reflects:

1. **Stakeholder-first**: Start with why (business value) before how (implementation)
2. **Abstraction cascade**: Each layer adds implementation detail
3. **Natural dependencies**: Technologies should satisfy behaviors, not drive them
4. **Team alignment**: Maps to typical Agile roles (PO → Engineer → Developer → DevOps → Tester)

This order is not arbitrary—it reflects how successful software projects naturally flow from intent to implementation.

## Trade-offs

**Benefits:**
- Clear separation of concerns
- Explicit requirement sources
- Reduced cascading changes
- Team role boundaries respected

**Costs:**
- More upfront structure (5 separate agent invocations vs 1)
- Requires discipline to maintain layer boundaries
- Potential for inconsistency if layers diverge

The cost is intentional—making layer boundaries explicit prevents the hidden costs of tangled requirements.

## Related

- [Layer Independence](../concepts/layer-independence.md) — How layers maintain independence
- [Layer References Upstream](layer-references-upstream.md) — Why layers reference despite independence
