# Level Agent Extraction Cleanup

## Metadata

- **Date:** 2026-04-09
- **Focus:** Remove Level Agent artifacts (L0/L1/L2) from smaqit repo after migration to smaqit-adk
- **Tasks referenced:** 064, 066 (abandoned)

## Actions Taken

- Explored smaqit-adk repository to understand what was migrated and what remains smaqit-specific
- Confirmed installer embeds only `installer/agents/` (9 product agents) — framework and agent templates are never shipped
- Confirmed product agents are fully compiled and self-contained with zero runtime dependency on framework or templates
- Deleted 3 level agent files (L0, L1, L2) from `.github/agents/`
- Deleted 3 generic agent template files from `templates/agents/`
- Deleted 3 generic compiled rules files (base, specification, implementation) from `templates/agents/compiled/`
- Cleaned `.github/copilot-instructions.md` — removed Level 0 Content Model section, all L0/L1/L2 invocation guidance blocks, Level Agents section, Level Contamination Awareness section; fixed Source/Artifacts structure and Installer description to reflect actual behavior
- Cleaned `agents/smaqit.qa.agent.md` — removed 2 stale Agent-L2 redirect references
- Abandoned tasks 064 and 066 in PLANNING.md (superseded by smaqit-adk extraction)
- Kept 8 smaqit-specific compiled rules: `business`, `functional`, `stack`, `infrastructure`, `coverage`, `develop`, `deploy`, `validate` (not in smaqit-adk)
- Kept all `framework/` files (LAYERS.md, PHASES.md, PROMPTS.md, ARTIFACTS.md, etc. — smaqit-product-specific, not in smaqit-adk)

## Problems Solved

- **Ambiguity over what to keep:** The key distinction — smaqit-adk owns the generic ADK compilation chain (L0/L1/L2 agents, base/spec/impl templates and rules); smaqit owns the product domain layer/phase definitions (layers, phases, compiled rules) — resolved through direct examination of smaqit-adk repo structure
- **Stale instructions:** copilot-instructions.md contained several sections describing workflows that no longer apply (L0/L1/L2 agent invocations, level contamination cleanup, incorrect installer artifact description)

## Decisions Made

- **Keep framework/ entirely** — All 8 files are smaqit-product-specific (layers, phases, prompts, artifacts, agents, templates, skills, smaqit index). smaqit-adk's framework/ covers only generic ADK principles, not product domain definitions.
- **Keep 8 layer/phase compiled rules** — `business.rules.md` through `validate.rules.md` are smaqit-specific domain content; smaqit-adk only has the generic `base`, `specification`, `implementation` rules.
- **Delete 3 agent templates + 3 generic rules** — These are fully equivalent to what smaqit-adk ships; no smaqit-specific content.
- **Abandon 064 and 066** — Both tasks targeted Level Layer purity cleanup; the extraction to smaqit-adk supersedes that work entirely.

## Files Modified

| File | Action |
|------|--------|
| `.github/agents/smaqit.L0.agent.md` | Deleted |
| `.github/agents/smaqit.L1.agent.md` | Deleted |
| `.github/agents/smaqit.L2.agent.md` | Deleted |
| `templates/agents/base-agent.template.md` | Deleted |
| `templates/agents/specification-agent.template.md` | Deleted |
| `templates/agents/implementation-agent.template.md` | Deleted |
| `templates/agents/compiled/base.rules.md` | Deleted |
| `templates/agents/compiled/specification.rules.md` | Deleted |
| `templates/agents/compiled/implementation.rules.md` | Deleted |
| `.github/copilot-instructions.md` | Removed stale L0/L1/L2 sections; updated Kit Components, Source/Artifacts tree, Installer section |
| `agents/smaqit.qa.agent.md` | Removed 2 Agent-L2 redirect references |
| `docs/tasks/PLANNING.md` | Abandoned tasks 064 and 066 |

## Next Steps

- Consider a release to publish the cleanup (smaqit-adk extraction is a notable structural change)
- Task 070 (E2E Boundary Enforcement Validation) — high priority, no blockers
- Task 071 (Q&A Agent and GitHub Skill for Wiki Documentation) — medium priority

## Session Metrics

- Files deleted: 9
- Files edited: 3
- Tasks abandoned: 2 (064, 066)
- Tasks completed: 0 (cleanup work, no feature tasks)
