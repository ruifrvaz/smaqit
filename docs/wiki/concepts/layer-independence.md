# Layer Independence

## Definition

Layer independence is the principle that each specification layer receives its requirements from its own prompt file, not from upstream layers. Upstream layers provide context for coherence and traceability, but they do not define requirements for downstream layers.

## How It Works

Each layer has a dedicated prompt file in `.github/prompts/`:

| Layer | Prompt File | Requirements Source |
|-------|-------------|---------------------|
| Business | `smaqit.business.prompt.md` | Stakeholder input |
| Functional | `smaqit.functional.prompt.md` | Experience requirements |
| Stack | `smaqit.stack.prompt.md` | Technology preferences |
| Infrastructure | `smaqit.infrastructure.prompt.md` | Deployment requirements |
| Coverage | `smaqit.coverage.prompt.md` | Verification requirements |

The prompt file is the **sole source of requirements** for each layer. Upstream specs are read for context, not copied as requirements.

## Why Independence Matters

**Prevents false derivation chains:**
- Business requirement change doesn't automatically cascade to Functional
- Users must explicitly update each layer's prompt when requirements change
- Changes are traceable to specific layer modifications

**Enables clear change tracking:**
- Git diff shows which layer's prompt changed
- No ambiguity about whether business changed or someone's interpretation changed
- Explicit user intent at every layer

**Supports layer swapping:**
- Can change Stack without touching Business or Functional
- Can add new Infrastructure without modifying application layers
- Each layer can evolve independently when requirements for that layer change

## Example

**Scenario:** Stakeholder adds new business requirement

**With layer independence:**
1. User updates `smaqit.business.prompt.md`
2. Business agent regenerates business spec
3. Functional spec unchanged (user didn't update functional prompt)
4. If functional changes needed, user explicitly updates `smaqit.functional.prompt.md`

**Without layer independence (hypothetical):**
1. User updates business requirements somewhere
2. Functional agent "derives" new functional requirements
3. Unclear if functional change was intended or agent interpretation
4. Silent cascading changes through all layers

## Coherence vs Requirements

Layers still reference upstream specs, but for different purposes:

| Purpose | How It Works |
|---------|--------------|
| **Requirements** | From prompt file only |
| **Coherence** | Check compatibility with upstream specs |
| **Traceability** | Link coverage back through all layers |

Implementation agents consolidate specs from multiple layers and validate coherence before execution.

## Related

- [Progressive Refinement](../designs/progressive-refinement.md) — Why layers are structured this way
- [Layer References Upstream](../designs/layer-references-upstream.md) — Why layers reference despite independence
