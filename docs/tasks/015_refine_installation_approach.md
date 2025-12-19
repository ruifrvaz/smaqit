# Task: Investigate Framework Bundling at Installation

**ID**: 015
**Status**: new

## Context

Currently, `smaqit init` installs the full framework directory (`.smaqit/framework/`) containing all Level 0 development files (SMAQIT.md, LAYERS.md, PHASES.md, TEMPLATES.md, AGENTS.md, ARTIFACTS.md). These files were foundational for building Level 1 (templates) and Level 2 (agents/prompts) but may not be necessary for end users.

**Current installation:**
- `.smaqit/framework/` — Full framework docs (6 markdown files)
- `.smaqit/templates/specs/` — Spec templates (5 files)
- `.github/agents/` — Agent definitions (8 files)
- `.github/prompts/` — Copilot prompts (8 files, when implemented)

**Question:** Should Level 2 users receive condensed framework instructions instead of full Level 0 docs?

## Investigation Areas

### 1. Framework Reference Analysis

- [ ] Identify which framework files are referenced by agents and prompts
- [ ] Determine if agents/prompts need direct access to framework files
- [ ] Check if framework files are used during runtime vs. build-time only

### 2. Condensed Instructions Approach

- [ ] Prototype a single `FRAMEWORK_GUIDE.md` or embedded instructions in agents
- [ ] Evaluate if core principles (Layer Independence, Traceability, etc.) can be condensed
- [ ] Determine if user-facing docs need full framework depth or just essentials

### 3. Documentation Layering

Establish what belongs at each level:

| Level | Audience | Content | Location |
|-------|----------|---------|----------|
| Level 0 | smaqit developers | Full framework specs, design rationale | This repo only |
| Level 1 | Template creators | Template structure rules | `.smaqit/templates/` (or embedded) |
| Level 2 | End users | Agent definitions, prompts, condensed framework guide | `.github/`, `specs/` |

### 4. Agent Self-Containment

- [ ] Investigate embedding framework rules directly in agent definitions
- [ ] Determine if agents can be fully self-contained (no external framework refs)
- [ ] Evaluate trade-offs: agent file size vs. framework file duplication

### 5. User Experience Impact

- [ ] Survey what users actually need to understand to use smaqit effectively
- [ ] Identify minimal viable documentation for spec-driven development
- [ ] Determine if full framework is helpful or overwhelming for new users

## Acceptance Criteria

- [ ] Document current framework file usage by agents/prompts
- [ ] Prototype at least one alternative bundling approach
- [ ] Provide recommendation with evidence from investigation
- [ ] Update installer if recommendation differs from current approach

## Potential Outcomes

**Option 1: Keep Current Approach**
- Ship full `.smaqit/framework/` directory
- Users have complete reference docs
- Maintains current agent references

**Option 2: Condensed Framework Guide**
- Create single `FRAMEWORK_GUIDE.md` with essentials
- Remove individual framework files
- Update agent references to condensed guide

**Option 3: Self-Contained Agents**
- Embed all framework rules in agent definitions
- No separate framework directory needed
- Agents are fully standalone

**Option 4: Hybrid Approach**
- Ship minimal reference (e.g., just SMAQIT.md)
- Embed layer-specific rules in agents
- Point users to online docs for deep dive

## Notes

**Current agent references:**
All agents include:
```markdown
## Framework Reference

- [SMAQIT](../framework/SMAQIT.md) — Core principles
- [LAYERS](../framework/LAYERS.md) — Layer definitions
- [TEMPLATES](../framework/TEMPLATES.md) — Template rules
- [AGENTS](../framework/AGENTS.md) — Agent behaviors
- [ARTIFACTS](../framework/ARTIFACTS.md) — Artifact rules
```

**Question:** Are these references informational (users can ignore) or operational (agents need them at runtime)?

**Evidence needed:**
- Do users consult framework files when using smaqit?
- Are framework files cluttering the project?
- Would condensed docs reduce barrier to entry?

## Related Tasks

- Task 001: Copilot prompt files (also part of Level 2)

