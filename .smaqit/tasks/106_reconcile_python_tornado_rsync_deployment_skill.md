# Reconcile Python/Tornado Rsync Deployment Skill Into Canonical smaqit

**Status:** Not Started
**Created:** 2026-08-13

## Description

A downstream project validated a deployment skill for a Python/Tornado monolith with SQLite persistence, nginx, systemd, and rsync. The authored definition has been contributed to canonical smaqit at `.smaqit/definitions/skills/smaqit.infrastructure-deploy-rsync-python-tornado.md`; this task makes it a supported product capability rather than a downstream-only artifact.

The product implementation must retain the safety-critical deployment behavior while generalizing project-specific provenance and examples. It must integrate with the global installation model and the Dynamic Stack Detection / on-the-fly deploy-skill synthesis flow from Task 087, so matching projects select this maintained skill rather than synthesizing an equivalent copy.

## Design Decisions

- **Deployment profile:** Python/Tornado, single-process, no Docker or frontend build, SQLite persistence, nginx, and systemd are the matching stack characteristics.
- **Data safety:** Rsync excludes SQLite databases and local virtual environments; the deployed database remains on the target's persistent storage.
- **Reusable dependencies:** nginx vhost generation and deploy stamps reuse the existing shared capabilities rather than duplicating their implementations.
- **Product integration:** The canonical skill is built and installed globally through the normal smaqit release pipeline; it is not copied back into project-local framework mirrors.

## Implementation Steps

1. Assess the contributed definition against the canonical deployment-skill conventions and generalize any project-specific provenance, paths, or examples.
2. Compile the definition into `skills/smaqit.infrastructure-deploy-rsync-python-tornado/`, including any required bundled references or scripts, and validate its metadata and self-references.
3. Update stack-detection and no-match/synthesis routing so the Python/Tornado/systemd/SQLite profile selects the maintained skill.
4. Include the skill in global payload generation and installation for every supported agent platform.
5. Add focused automated coverage for matching, non-matching, and safety-critical rsync exclusions; run the relevant installer and product test suites.
6. Update user-facing documentation and release notes for the new supported deployment profile.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] Canonical smaqit contains a reviewed, product-generic definition and compiled `smaqit.infrastructure-deploy-rsync-python-tornado` skill.
- [ ] Dynamic deploy-skill selection recognizes the Python/Tornado, SQLite, nginx, systemd, no-Docker profile and selects the canonical skill without synthesis.
- [ ] The skill preserves database and virtual-environment rsync exclusions, uses the shared nginx-vhost and deploy-stamp conventions, and does not introduce Terraform or Docker steps.
- [ ] The released global payload makes the skill available through the supported global skill roots.
- [ ] Automated tests cover profile routing and the safety-critical deployment behavior, and relevant product/installer tests pass.
- [ ] Product documentation records the supported deployment profile and its operational prerequisites.

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| `.smaqit/definitions/skills/smaqit.infrastructure-deploy-rsync-python-tornado.md` | Create — contributed downstream definition |
| `skills/smaqit.infrastructure-deploy-rsync-python-tornado/` | Create — compiled canonical skill and any supporting assets |
| `skills/smaqit.new-greenfield-project/` and routing implementation | Modify — select the maintained skill for the matching stack profile |
| Global installer/payload generation and tests | Modify — distribute and verify the new skill |
| User documentation and release notes | Modify — document supported profile |

## Notes

The source definition was synthesized and validated in a downstream project on 2026-07-21. It deliberately differs from the existing Node/React Docker Compose and FastAPI/Next.js Docker Compose deploy skills: the target is a direct Python/Tornado rsync deployment with a persistent SQLite database and systemd supervision.
