# Why Layers Reference Upstream

## Question

If layers are independent and each gets requirements from session context, why do specs reference upstream layers at all?

## Answer

**Layers reference upstream for coherence checking and traceability, not for requirements derivation.**

## Two Purposes

### 1. Coherence Validation

Implementation agents consolidate specs from multiple layers before execution. References ensure specs are compatible across layers.

**Example conflict:**
- Business spec says: "real-time updates required"
- Stack spec says: "batch processing system"

When Development agent consolidates these, it detects the conflict and flags it. Without references, the agent wouldn't know which specs to check for compatibility.

### 2. Traceability Chains

Coverage layer maps requirements through all layers to ensure complete specification coverage.

**Traceability chain:**
```
Session Context → Business Spec → Functional Spec → Stack Spec → Implementation → Tests
```

Each link in the chain is explicit. This enables:
- Impact analysis (what code is affected by a requirement change?)
- Gap detection (which requirements lack tests?)
- Audit trails (where did this feature come from?)

## Key Distinction

**Requirements flow:** Session Context → Spec (one direction only)

**References flow:** Spec → Upstream Spec (for validation and traceability)

Requirements come FROM session context. References point TO upstream specs for context.

## Benefits

- **Early conflict detection**: Agents catch incompatibilities before implementation
- **Complete coverage**: Nothing falls through the cracks
- **Explicit traceability**: Every requirement has a clear path from session input to code to tests
- **Independent evolution**: Update one layer's requirements without affecting others

## Trade-offs

**Cost**: Agents must read upstream specs even though requirements come from session context.

**Benefit**: Worth it for coherence validation and traceability guarantees.

## Related

- [Layer Independence Principle](../concepts/layer-independence.md) — How layers stay independent despite references
- [Traceability Across Layers](../concepts/traceability.md) — Complete traceability chains
