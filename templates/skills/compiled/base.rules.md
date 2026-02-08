---
target: .github/skills/[skill-name]/SKILL.md
sources:
  - framework/SKILLS.md
created: 2026-02-06
---

# Base Skill Rules

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| framework/SKILLS.md | Skills as Invocable Capabilities |
| framework/SKILLS.md | Condition Detection |
| framework/SKILLS.md | Skill Invocation Context |
| framework/AGENTS.md | Fail-Fast on Ambiguity |

## L1 Directive Compilation

### Frontmatter

MUST include required fields:
- name: 1-64 characters, lowercase alphanumeric and hyphens only
- description: 1-1024 characters describing what the skill does and when to use it

SHOULD include keywords in description that help agents identify relevant trigger conditions.

MUST match name field to parent directory name.

### Body Content

MUST provide clear, step-by-step instructions for the workflow.

MUST define output format so agents know what the skill returns.

SHOULD include examples of inputs and expected outputs.

SHOULD document edge cases and error handling.

### Progressive Disclosure

MUST keep main SKILL.md under 500 lines.

SHOULD move detailed reference material to separate files in references/ directory.

MUST structure content for efficient context usage:
- Metadata (~100 tokens): name and description for startup loading
- Instructions (<5000 tokens): full SKILL.md body when activated
- Resources (as needed): additional files loaded on demand

## Compilation Guidance for Agent-L2

When compiling a skill from base template + base rules + specific rules:

1. Read base-skill.template.md for structure
2. Read base.rules.md for foundation directives
3. Read [skill-specific].rules.md for specialized directives
4. Replace [SKILL_NAME] with skill name (lowercase with hyphens)
5. Replace [SKILL_DESCRIPTION] with description containing trigger keywords
6. Replace [SKILL_TITLE] with title case skill name
7. Replace [SKILL_BODY] with introduction and purpose
8. Replace [WORKFLOW_STEPS] with numbered step-by-step instructions
9. Replace [OUTPUT_FORMAT] with structured output specification
10. Validate: no placeholders remain, all required frontmatter fields present, name matches directory
