# Task: Investigate Framework Bundling at Installation

**ID**: 015
**Status**: Complete

## Decision: Embed Framework Content in Agents

After comprehensive assessment, determined that framework files should NOT be bundled in user installations. Instead, framework content will be embedded directly in agent definitions at Level 1 (templates) and Level 2 (agents).

**Rationale:**
- Saves ~20K tokens per workflow (no file reads needed)
- Faster agent execution (no startup reads)
- Cleaner user workspace (70% less content)
- More reliable (no risk of file deletion)
- Consistent with Level 0→1→2 compilation model

**Implementation approach:** 3-phase sequential update preserving foundational sequence.

---

## Phase 1: Update Level 1 (Templates) - COMPLETE ✓

### Changes Made

**1. Specification Agent Template** (`templates/agents/specification-agent.template.md`):
- Added **Layer-Specific Rules** section with placeholders
- Added **Requirement ID Format** section (format, components, examples, rules)
- Added **Acceptance Criteria Format** section (testability requirements, untestable pattern)
- Added **File Organization** section (one spec per concept, naming conventions)
- Added `[TRACEABILITY_FORMAT]` placeholder for Implements/Enables
- Removed reference to ARTIFACTS.md

**2. Implementation Agent Template** (`templates/agents/implementation-agent.template.md`):
- Added **State Tracking** section (JSON format, atomic writes)
- Added state.json writing to MUST directives
- Removed reference to Anchoring Principle in ARTIFACTS.md

**Review checkpoint passed:** Template structure ready for Phase 2 content population

---

## Phase 2: Update Level 2 (Agents) - COMPLETE ✓

### All 8 agents regenerated with embedded content:

**Specification agents (5):**
- smaqit.business.agent.md — Embedded Business layer rules, BUS prefix format
- smaqit.functional.agent.md — Embedded Functional rules, FUN prefix, Foundation vs Feature pattern
- smaqit.stack.agent.md — Embedded Stack rules, STK prefix format
- smaqit.infrastructure.agent.md — Embedded Infrastructure rules, INF prefix format
- smaqit.coverage.agent.md — Embedded Coverage rules, COV prefix, coverage translation pattern

**Implementation agents (3):**
- smaqit.development.agent.md — Embedded State Tracking, industry standards directives
- smaqit.deployment.agent.md — Embedded State Tracking, Isolation Principle (secrets), industry standards directives
- smaqit.validation.agent.md — Embedded State Tracking, Validation Report Format

### Refinements applied:
- Replaced specific examples with generic placeholders (no false requirements)
- Streamlined verbose principle sections to concise actionable directives
- Aligned security rules with Level 0 Isolation Principle
- Restored "structurally recognizable and behaviorally equivalent" directive

**Result:** All agents now self-contained, no ARTIFACTS.md or LAYERS.md references

---

## Phase 3: Update Level 3 (Installer) - COMPLETE ✓

### Installer Changes (installer/main.go):
- ✅ Removed `//go:embed framework/*.md` directive
- ✅ Removed `frameworkFiles` variable
- ✅ Removed `.smaqit/framework` from directory creation
- ✅ Removed framework file copying logic
- ✅ Removed "✓ Copied framework files" output message
- ✅ Removed `.smaqit/framework` from validation checks
- ✅ Removed framework files existence validation

### Testing:
- ✅ Installer builds successfully
- ✅ `smaqit init` creates clean structure without framework/
- ✅ `smaqit validate` passes without framework directory
- ✅ `smaqit status` works correctly
- ✅ Agent files copied correctly to `.github/agents/`

### Documentation:
- ✅ README.md accurate (describes source repo structure)
- ✅ Wiki files accurate (references are about source, not bundled files)

---

## Final Outcome

**User projects now receive:**
- `.smaqit/templates/specs/` — 5 specification templates
- `.smaqit/state.json` — Phase tracking
- `.github/agents/` — 8 self-contained agent definitions
- `.github/prompts/` — 9 prompt files
- `specs/` — 5 layer directories (empty initially)

**Benefits achieved:**
- 70% less content in user workspace
- ~20K tokens saved per workflow
- Faster agent execution (no startup file reads)
- Cleaner, simpler project structure
- More reliable (no framework file dependencies)

## Related Tasks

- Task 027: Separate framework instructions from human rationale (prerequisite)
- Task 028: Audit levels for meta-rationale (prerequisite)
- ✅ State tracking section (JSON format, atomic writes)
- ✅ Anchoring Principle (from ARTIFACTS.md Implementation Artifacts section)
- ✅ Three Dimensions model (Behavior/Structure/Internals)

**smaqit.deployment.agent.md:**
- ✅ State tracking section (JSON format, atomic writes)
- ✅ Anchoring Principle (from ARTIFACTS.md Implementation Artifacts section)
- ✅ Isolation Principle (from ARTIFACTS.md Implementation Artifacts section)
- ✅ Three Dimensions model

**smaqit.validation.agent.md:**
- ✅ State tracking section (JSON format, atomic writes)
- ✅ Validation report format (from ARTIFACTS.md)
- ✅ Spec coverage calculation rules

**Next steps:** Regenerate all 8 agents with embedded content from framework files

---

## Work Completed (Side Tasks)

While investigating task 015, implemented related improvements:

1. **Version tracking via spec templates** - Each generated spec includes version footer (eliminated VERSION file)
2. **Phase state tracking via state.json** - Structured JSON tracking phase completion with timestamps
3. **Status command redesigned** - Shows spec counts per layer and phase completion based on state.json
4. **Prompt templates removed from installer** - Level 1 structure stays in source, not distributed

These improvements don't resolve the core investigation question below.

## Core Investigation (Still Open)

**Original question:** Should Level 2 users receive condensed framework instructions instead of full Level 0 docs?

**Current state:** Framework files (7 markdown files) are copied to `.smaqit/framework/` during installation. These are Level 0 development docs.

**Decision needed:** Keep full framework, create condensed guide, embed in agents, or hybrid approach?

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

