# Task: Create Implementation Agent Template

**ID**: 003
**Status**: completed

## Context

Create a meta-template for generating implementation agents (development, deployment, validation). These agents differ from specification agents—they consume specs and produce artifacts (code, infrastructure, test results) rather than producing specs.

A consistent template for implementation agents would:
- Define the common structure all implementation agents must follow
- Enable an agent to generate new implementation agents from this template
- Ensure consistency across all implementation layer agents
- Clarify the distinction between spec agents and implementation agents

## Acceptance Criteria

- [x] Create `agents/implementation.agent.template.md` (or similar location)
- [x] Template defines: YAML frontmatter structure, Role, Input, Output, Constraints sections
- [x] Template includes placeholders for implementation-specific customization
- [x] Document usage in copilot-instructions or SMAQIT.md

## Notes

- Implementation agents consume specs as input and produce artifacts as output
- Consider relationship with Task 002 (specification agent template) — shared base vs distinct templates?
- Agents to consider: development, deployment, validation (coverage agent may be hybrid?)
