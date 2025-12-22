# Separate framework instructions from human rationale

**Status:** Completed  
**Created:** 2025-12-21

## Description

Framework files (`framework/`) should contain only pure LLM execution instructions. Meta-rationale, design explanations, and human context should be separated into `docs/wiki/` for human readers.

This separation serves two purposes:
1. **Framework files become pure agent instructions** — No distractions, no external references, no "why" explanations
2. **Wiki becomes human learning resource** — Rationale, examples, workflows, patterns

## Acceptance Criteria

- [x] Framework files stripped of meta-rationale (SMAQIT.md, PROMPTS.md, LAYERS.md)
- [x] Wiki directory structure created (`docs/wiki/{concepts,design-decisions,patterns,workflows}`)
- [x] Eight wiki documents created with moved content:
  - [x] `concepts/prompts-as-input-records.md`
  - [x] `design-decisions/free-style-prompts.md`
  - [x] `design-decisions/layer-references-upstream.md`
  - [x] `patterns/html-comment-convention.md`
  - [x] `patterns/validation-messages.md`
  - [x] `workflows/amending-requirements.md`
  - [x] `patterns/prompt-evolution.md`
  - [x] `patterns/archiving-prompts.md`
- [x] README.md updated with "Documentation Structure" section
- [x] Framework Evolution moved to `docs/history/011_framework_evolution_meta.md`
- [x] "See Also" sections centralized in SMAQIT.md, removed from other framework files
- [x] Content guidelines added to `.github/copilot-instructions.md`
- [x] Task created documenting refactoring work

## Notes

**Removed from SMAQIT.md:**
- "Iteration Through Experimentation" section (~40 lines)
- Moved to `docs/history/011_framework_evolution_meta.md`
- "Framework Files" renamed to "See Also" with "When to Consult" guidance

**Removed from PROMPTS.md:**
- "Why free-style?" rationale bullets
- Validation message examples with ✅❌ formatting
- 5-step amendment workflow detail
- Amendment vs New Specs examples
- Archiving suggestions and patterns
- Template structure examples (replaced with references)
- "See Also" section
- Reduced from ~280 lines to ~160 lines

**Removed from LAYERS.md:**
- "Why Layers Reference Upstream" framing (kept instructional table, removed rationale)
- "See Also" section

**Removed from other framework files:**
- "See Also" sections from TEMPLATES.md, AGENTS.md, ARTIFACTS.md, PHASES.md

**Created wiki structure:**
- `concepts/` — Core concepts explained
- `design-decisions/` — Why we chose these patterns (8 documents)
- `patterns/` — Common usage patterns
- `workflows/` — Step-by-step processes

**Updated copilot-instructions.md:**
- Added 6 new content guidelines (3 Do's, 3 Don'ts)
- Enforces separation: framework/templates/agents = pure instructions, wiki = rationale

**Documentation audience separation:**
- **Framework files**: For LLM agents (instructions only)
- **Wiki**: For human developers (context and rationale)
- **README**: For users and contributors (project overview)
