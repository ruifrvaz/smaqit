# Team Alignment

smaqit layers align with Agile/Scrum team roles, enabling specialists to work in their domain.

## Role-to-Layer Mapping

| Role | Layer | Focus |
|------|-------|-------|
| Stakeholders | Input | Requirements and business needs |
| Product Owner | Business | Why, for whom, success criteria |
| Engineers | Functional | What behaviors, contracts, data models |
| Software Developer | Stack | With what languages, frameworks, tools |
| DevOps | Infrastructure | Where and how it runs |
| Testers | Coverage | How we verify it works |

## Layer Boundaries Respect Role Boundaries

- **Product Owners** define *what* success looks like, not *how* to build it
- **Engineers** translate business goals into system behaviors
- **Developers** choose technologies that satisfy functional requirements
- **DevOps** specifies runtime environment without changing application logic
- **Testers** verify against requirements without inventing new ones

Each role focuses on their expertise without immediate cross-concerns. This separation:

1. **Reduces cognitive load** — Each role works within familiar concepts
2. **Enables parallel work** — Different roles can fill their prompts simultaneously
3. **Creates clear handoffs** — Specs at each layer are contracts for downstream work
4. **Maintains accountability** — When something fails, the layer boundary identifies ownership

## Example: Feature Request Flow

When a new feature is requested:

1. **Stakeholders** describe the business need
2. **Product Owner** captures use cases and success criteria in Business prompt
3. **Engineers** define behaviors and data models in Functional prompt
4. **Developers** select technologies in Stack prompt
5. **DevOps** specifies deployment in Infrastructure prompt
6. **Testers** define verification in Coverage prompt

Each role owns their layer's prompt. Agents generate specs from those prompts, maintaining the separation throughout implementation.
