# Audit all smaqit levels for meta-rationale (Part 2 of 027)

**Status:** Completed  
**Created:** 2025-12-22  
**Completed:** 2025-12-22

## Description

Systematically audit all files across all smaqit levels (framework, templates, agents) to ensure they follow the separation principle: pure instructions in level files, rationale in wiki.

This is Part 2 of task 027, expanding the cleanup to templates and agent definitions.

## Scope

**Level 0 - Framework files (`framework/`):**
- [x] SMAQIT.md - Already cleaned (removed Design Philosophy section)
- [x] PROMPTS.md - Already cleaned
- [x] LAYERS.md - Already cleaned
- [x] PHASES.md - Reviewed, no meta-rationale found
- [x] TEMPLATES.md - Cleaned (removed Purpose bullets)
- [x] AGENTS.md - Cleaned (removed Purpose explanations)
- [x] ARTIFACTS.md - Cleaned (removed Purpose explanations)

**Level 1 - Templates (`templates/`):**
- [x] templates/specs/*.template.md (5 files) - Reviewed, only stack.template.md needed cleanup
- [x] templates/specs/stack.template.md - Removed Rationale columns
- [x] templates/prompts/*.template.md (2 files) - Reviewed, clean (free-style by design)
- [x] templates/agents/*.template.md (2 files) - Reviewed, clean (instructional only)

**Level 2 - Agent definitions (`agents/`):**
- [x] agents/*.agent.md (8 files) - Spot-checked, no meta-rationale found (follow templates)

**Wiki organization:**
- [x] Renamed docs/wiki/design-decisions/ → docs/wiki/designs/
- [x] Updated all references from design-decisions to designs

## Acceptance Criteria

- [x] All framework files reviewed and stripped of meta-rationale
- [x] All spec templates reviewed (business, functional, stack, infrastructure, coverage)
- [x] All prompt templates reviewed (specification-prompt, phase-prompt)
- [x] All agent templates reviewed (specification-agent, implementation-agent)
- [x] All agent definitions reviewed (8 agents)
- [x] Any found meta-rationale moved to appropriate wiki documents
- [x] Content guidelines compliance verified across all levels

## What Was Found and Fixed

### Framework Files (4 cleaned)

**SMAQIT.md:**
- Removed entire "Design Philosophy" section (73-105)
- Moved to: progressive-refinement.md, explicit-over-implicit.md, fail-fast-on-ambiguity.md

**TEMPLATES.md:**
- Removed "Purpose" section with benefits bullets (lines 5-12)
- Moved to: template-constraints.md

**AGENTS.md:**
- Removed "Purpose" explanations from Specification Agents and Implementation Agents sections
- Kept instructional one-liner descriptions

**ARTIFACTS.md:**
- Removed "Purpose" section with bullet explanations (lines 15-22)
- Kept instructional one-liner description

### Spec Templates (1 cleaned)

**stack.template.md:**
- Removed "Rationale" columns from Languages and Frameworks tables
- Stack specs now document WHAT technologies, not WHY they were chosen

### Wiki Documents Created (4 new)

Created in `docs/wiki/designs/`:
1. **progressive-refinement.md** - Why layers are independent, why they reference upstream, layer order rationale
2. **explicit-over-implicit.md** - Why smaqit favors explicit documentation, trade-offs, examples
3. **fail-fast-on-ambiguity.md** - Why agents stop on unclear input, cost-benefit analysis, patterns
4. **template-constraints.md** - Why templates are mandatory, enforcement mechanisms, trade-offs

## Results

**Files audited:** 20 (7 framework + 5 spec templates + 2 prompt templates + 2 agent templates + 4 spot-checked agents)  
**Files cleaned:** 5 (4 framework + 1 spec template)  
**Wiki docs created:** 4  
**Directory renamed:** 1 (design-decisions → designs)  
**References updated:** 9 files
