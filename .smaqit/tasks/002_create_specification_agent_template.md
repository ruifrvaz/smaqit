# Task: Create Specification Agent Template

**ID**: 002
**Status**: completed

## Context

Create a meta-template that can be used to generate specification agents (business, functional, stack, infrastructure, validation). All specification agents should follow a consistent structure, making it easier to create new spec agents and maintain existing ones.

Currently, each agent (`.agent.md`) is hand-crafted. A specification agent template would:
- Define the common structure all spec agents must follow
- Enable an agent to generate new spec agents from this template
- Ensure consistency across all specification layer agents

## Acceptance Criteria

- [x] Create `agents/specification.agent.template.md` (or similar location)
- [x] Template defines: YAML frontmatter structure, Role, Input, Output, Constraints sections
- [x] Template includes placeholders for layer-specific customization
- [x] Document usage in copilot-instructions or SMAQIT.md

## Notes

- Consider which agents are "specification agents" vs other types (e.g., coverage, deployment, development)
- Template should be usable by both humans and AI agents to create new spec agents
