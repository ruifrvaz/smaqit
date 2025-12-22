# Audit all smaqit levels for meta-rationale (Part 2 of 027)

**Status:** Not Started  
**Created:** 2025-12-22

## Description

Systematically audit all files across all smaqit levels (framework, templates, agents) to ensure they follow the separation principle: pure instructions in level files, rationale in wiki.

This is Part 2 of task 027, expanding the cleanup to templates and agent definitions.

## Scope

**Level 0 - Framework files (`framework/`):**
- [x] SMAQIT.md - Already cleaned
- [x] PROMPTS.md - Already cleaned
- [x] LAYERS.md - Already cleaned
- [ ] PHASES.md - Need to review
- [ ] TEMPLATES.md - Need to review
- [ ] AGENTS.md - Need to review
- [ ] ARTIFACTS.md - Need to review

**Level 1 - Templates (`templates/`):**
- [ ] templates/specs/*.template.md (5 files)
- [ ] templates/prompts/*.template.md (2 files)
- [ ] templates/agents/*.template.md (2 files)

**Level 2 - Agent definitions (`agents/`):**
- [ ] agents/*.agent.md (8 files)

## Acceptance Criteria

- [ ] All framework files reviewed and stripped of meta-rationale
- [ ] All spec templates reviewed (business, functional, stack, infrastructure, coverage)
- [ ] All prompt templates reviewed (specification-prompt, phase-prompt)
- [ ] All agent templates reviewed (specification-agent, implementation-agent)
- [ ] All agent definitions reviewed (8 agents)
- [ ] Any found meta-rationale moved to appropriate wiki documents
- [ ] Content guidelines compliance verified across all levels

## What to Look For

**Red flags (move to wiki):**
- "Why" sections or explanations
- "Benefits" or "Trade-offs" sections
- Extended examples with commentary
- Historical context or evolution notes
- Design rationale or philosophy
- "This is because..." explanations

**Green flags (keep):**
- Direct instructions ("MUST", "MUST NOT", "SHOULD")
- Structure definitions (templates, formats)
- Process steps (workflows)
- Reference tables
- Validation criteria
- Error handling patterns

## Strategy

1. **Review framework files** (PHASES, TEMPLATES, AGENTS, ARTIFACTS)
2. **Review spec templates** for meta-commentary
3. **Review prompt templates** for rationale
4. **Review agent templates** for explanatory content
5. **Review agent definitions** for "why" explanations
6. **Create wiki documents** for any extracted rationale
7. **Update copilot-instructions** if new patterns emerge

## Notes

This is systematic cleanup. We already did emergency cleanup on SMAQIT, PROMPTS, LAYERS. Now we do thorough audit of everything else.
