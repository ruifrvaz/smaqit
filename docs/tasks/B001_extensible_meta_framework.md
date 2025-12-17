# Extensible Meta-Framework

**Status:** Backlog  
**Created:** 2025-12-17

## Insight

smaqit may be more than a software development tool — it could be an extensible spec-driven development framework where the domain itself is configurable.

## Potential

- **Custom layers** — e.g., "Compliance" for regulated industries, "Content" for media projects
- **Custom phases** — e.g., "Migrate" for legacy modernization, "Audit" for security reviews
- **Domain-specific agent kits** — Built on smaqit primitives for non-software domains

## Current Architecture Alignment

The Level 0 → Level 1 → Level 2 hierarchy already implies this:
- Level 0 (Framework) defines the rules
- Level 1 (Templates) provides structure
- Level 2 (Agents/Specs) are instances

Extension would operate at Level 1 — creating new layer templates and agents.

## Considerations

1. **Scope creep** — Current 5-layer, 3-phase model is already complex
2. **Evidence needed** — Per SMAQIT.md: "New layer: Multiple projects demonstrate a missing concern"
3. **Timing** — Base framework not yet proven on real projects
4. **Alternative** — Forkability vs extensibility

## Promotion Criteria

Move to Active when:
- [ ] Base kit is complete (all templates, installer working)
- [ ] smaqit has been applied to 2+ real projects
- [ ] Projects demonstrate consistent need for layers/phases that don't exist
- [ ] Concrete extension requirements emerge from evidence

## Notes

This task exists to preserve the insight for future consideration, not to act on it now.
