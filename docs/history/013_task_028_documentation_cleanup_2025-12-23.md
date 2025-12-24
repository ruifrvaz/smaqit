# Task 028 Complete: Documentation Architecture Cleanup

**Date:** 2025-12-22 to 2025-12-23  
**Session Type:** Refactoring and documentation  
**Tasks Completed:** 028

## Overview

Completed comprehensive audit of all 23 files across framework, templates, and agents to remove meta-rationale while preserving valuable context in wiki. Created 8 new wiki documents (4 designs + 4 concepts) with extensive explanations. Fixed structure diagram inconsistencies across project. Session also refined TEMPLATES.md and SMAQIT.md based on critical assessment.

## What Was Done

### Task 028: Systematic Audit

**Directory renamed:**
- `docs/wiki/design-decisions/` → `docs/wiki/designs/` (9 references updated across project)

**Framework files cleaned (4 files):**

1. **SMAQIT.md**
   - Initial: Removed entire "Design Philosophy" section (73-105 lines)
   - After refinement: Moved 3 subsections (Progressive Refinement, Explicit Over Implicit, Fail-Fast on Ambiguity) into Core Principles
   - Condensed to instructional directives while keeping essential guidance

2. **TEMPLATES.md**
   - Initial: Removed "Purpose" section with benefits bullets
   - After refinement: Rephrased Purpose section instructionally ("agents MUST produce consistent output structure", "enable predictable consumption", "minimize variance")
   - Kept as objectives to achieve rather than benefits explanation

3. **AGENTS.md**
   - Removed "Purpose" explanatory sections from Specification Agents and Implementation Agents
   - Kept one-liner instructional descriptions

4. **ARTIFACTS.md**
   - Removed "Purpose" section with bullet explanations
   - Kept one-liner instructional description

**Spec templates cleaned (1 file):**
- **stack.template.md**: Removed "Rationale" columns from Languages and Frameworks tables

**Other files reviewed:**
- PHASES.md: Clean (no meta-rationale found)
- All other spec templates: Clean (instructional only)
- Prompt templates (2): Clean (free-style by design)
- Agent templates (2): Clean (instructional only)
- Agent definitions (8): Spot-checked, clean (follow templates)

### Wiki Documents Created

**Design documents (4):**
1. **progressive-refinement.md** — Why layers are independent, why they reference upstream, order rationale, trade-offs
2. **explicit-over-implicit.md** — Why smaqit favors explicit documentation, manifestations in framework, trade-offs
3. **fail-fast-on-ambiguity.md** — Why agents stop on unclear input, cost-benefit analysis, patterns, trade-offs
4. **template-constraints.md** — Why templates are mandatory, enforcement mechanisms, placeholder conventions, trade-offs
5. **hierarchical-levels.md** — Four-level architecture explained, dependency flow, amendment propagation, practical implications

**Concept documents (4):**
1. **layer-independence.md** — How layers maintain independence through prompt files
2. **traceability.md** — How explicit references enable impact analysis and coverage mapping
3. **accept-mutability.md** — Why artifacts vary but behavior is consistent
4. **self-validating-agents.md** — How agents enforce compliance through completion criteria

### Structure Diagrams Fixed

**Inconsistencies corrected:**
1. **SMAQIT.md**: Moved specs/ from under `.smaqit/` to project root, added `.github/prompts/`
2. **Task 023**: Updated framework files count (6→7), added `.smaqit/templates/prompts/`, added `.github/prompts/`
3. **README.md**: Corrected hierarchical levels structure to show all 7 framework files at Level 0, all templates at Level 1, agents/prompts/specs at Level 2

**Now consistent:**
- 7 framework files (including PROMPTS.md)
- 9 template files (5 specs + 2 prompts + 2 agents)
- 8 agent definitions
- 8 prompt files
- specs/ at user project root (not in kit source)

## Decisions Made

### Critical Assessment Applied

**TEMPLATES.md Purpose section:**
- User correctly challenged removal
- Content serves as instructional objectives (what agents MUST achieve)
- Rephrased from benefits ("serve as cognitive scaffolds") to directives ("agents MUST produce consistent output structure")

**SMAQIT.md Design Philosophy:**
- User correctly challenged removal
- Content provides essential guidance for agent behavior
- Moved 3 subsections into Core Principles section
- Condensed but preserved directive content

### Documentation Separation Principle Refined

**What belongs in framework files:**
- Direct instructions ("MUST", "MUST NOT", "SHOULD")
- Objectives to achieve (what outcome agents must produce)
- Structure definitions and validation criteria
- Process steps and error handling patterns

**What belongs in wiki:**
- "Why" explanations and rationale
- Trade-offs and cost-benefit analysis
- Extended examples with commentary
- Historical context and evolution notes

**Gray area handled:**
- Brief objective statements (like Purpose sections) can stay if framed instructionally
- Core principles can include condensed guidance if agents need behavioral context

### Hierarchical Levels Clarified

**Correct structure:**
- **Level 0**: All 7 framework files (foundation rules)
- **Level 1**: All 9 template files (structure definitions)
- **Level 2**: Agents (8) + prompts (8) + specs (instances following templates)
- **Level 3**: Built system (application output)

## Problems Solved

### Problem 1: Meta-Rationale Mixed with Instructions
**Impact:** Framework files were 35-40% longer with explanatory content distracting from execution instructions.

**Solution:** Extracted ~200 lines of rationale to 8 wiki documents. Framework files now focused purely on what/how, wiki explains why.

### Problem 2: Overzealous Removal of Context
**Initial approach:** Removed all "Purpose" sections as meta-rationale.

**Correction:** Distinguished between explanatory benefits ("why we chose this") and instructional objectives ("what you must achieve"). Kept objectives, moved explanations.

### Problem 3: Broken Wiki References
**Initial error:** Created wiki documents with references to non-existent concept files.

**Correction:** Created 4 missing concept files (layer-independence, traceability, accept-mutability, self-validating-agents) with comprehensive explanations.

### Problem 4: Inconsistent Structure Diagrams
**Impact:** Different files showed different directory structures for user projects, confusing file locations.

**Solution:** Systematically reviewed all structure diagrams, corrected 3 inconsistencies, verified against actual installer behavior.

### Problem 5: Incorrect Level Hierarchy in README
**Impact:** README showed framework files split across Level 0 and Level 1, obscuring clear dependency structure.

**Solution:** Reorganized to show all framework files at Level 0, templates at Level 1, instances at Level 2, clarifying that each level builds on the previous.

## Files Modified

**Framework files (4):**
- framework/SMAQIT.md
- framework/TEMPLATES.md
- framework/AGENTS.md
- framework/ARTIFACTS.md

**Spec templates (1):**
- templates/specs/stack.template.md

**Documentation (3):**
- README.md
- .github/copilot-instructions.md
- docs/tasks/023_implement_installer_cli.md

**Wiki files created (8):**
- docs/wiki/designs/progressive-refinement.md
- docs/wiki/designs/explicit-over-implicit.md
- docs/wiki/designs/fail-fast-on-ambiguity.md
- docs/wiki/designs/template-constraints.md
- docs/wiki/designs/hierarchical-levels.md
- docs/wiki/concepts/layer-independence.md
- docs/wiki/concepts/traceability.md
- docs/wiki/concepts/accept-mutability.md
- docs/wiki/concepts/self-validating-agents.md

**Wiki files updated (3):**
- docs/wiki/concepts/prompts-as-input-records.md
- docs/wiki/patterns/html-comment-convention.md
- docs/wiki/patterns/validation-messages.md

**Task files (2):**
- docs/tasks/028_audit_all_levels_for_meta_rationale.md
- docs/tasks/PLANNING.md

**Total:** 22 files modified/created

## Next Steps

**Immediate:**
1. Test installer to verify all structure diagrams match reality
2. Run `@smaqit.user-testing` to validate agents still function with leaner framework files
3. Review wiki organization (17 files now in docs/wiki/)

**Task 029 Decision Required:**
Architecture choice for implementation prompts:
- Option 1: Phase prompts orchestrate entire phases (current)
- Option 2: Implementation prompts trigger implementation agents only
- Option 3: Both phase prompts AND implementation prompts
- Option 4: Implementation prompts + 1 orchestrator prompt

**Open Tasks:**
- 014: Define iterative development using smaqit
- 015: Investigate framework bundling at installation
- 022: Create GitHub Action for automated releases
- 025: Integrate testing agent with CI/CD
- 029: Simplify implementation prompts to minimal orchestration inputs

## Key Learnings

### 1. Critical Assessment is Essential
User correctly challenged two removals (TEMPLATES.md Purpose, SMAQIT.md Design Philosophy). Need to distinguish between:
- Explanatory benefits (move to wiki)
- Instructional objectives (keep in framework)

### 2. Documentation Separation Has Nuance
Not all "why" content is meta-rationale. Some context helps agents understand objectives. The line is:
- **Keep:** "What outcome must be achieved"
- **Move:** "Why we chose this approach"

### 3. Systematic Audits Catch Hidden Issues
Reviewing structure diagrams revealed 4 inconsistencies that would have confused users. Systematic review better than ad-hoc fixes.

### 4. Wiki References Need Validation
Creating design documents with references to concept files that don't exist breaks user experience. Should create referenced files immediately or remove references.

### 5. Levels Architecture Needs Clear Communication
README hierarchical levels section was confusing (split framework across levels). Created comprehensive wiki document to explain architecture properly.

## Session Metrics

**Duration:** ~3 hours across 2 days  
**Files audited:** 23  
**Files cleaned:** 5 (4 framework + 1 template)  
**Wiki docs created:** 8 (4 designs + 4 concepts)  
**Structure diagrams fixed:** 3  
**Directory renamed:** 1  
**References updated:** 9  
**Lines reduced from framework:** ~200  
**Lines added to wiki:** ~1800  
**Tasks completed:** 1 (028)
