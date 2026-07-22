# Add First-Class Codex Compatibility

**Status:** Completed
**Created:** 2026-07-22
**Mode:** Assisted
**Started:** 2026-07-22
**Completed:** 2026-07-22

## Description

Extend smaqit's canonical artifact generator and Go installer with Codex as a third supported platform. Compile all core agents and skills into Codex-native project locations, integrate them throughout install, update, validation, and uninstall, and verify safe coexistence with unrelated Codex content.

The implementation must adapt the committed Codex architecture from `smaqit-extensions` commit `71a90d2` without introducing a runtime dependency or copying extension-specific artifacts. Canonical content remains in `agents/` and `skills/`; Codex output is generated build staging only.

## Design Decisions

- **Complete product surface:** Install all 9 core agents and all 24 product skills for Codex.
- **Canonical sources:** Preserve `agents/` and `skills/` as the only canonical content sources; generated Codex staging is ephemeral.
- **Native discovery:** Install agents under `.codex/agents/` and skills under `.agents/skills/`; generate no Codex command files, `.codex/config.toml`, or `skills.config` registrations.
- **Platform-neutral shared content:** Resolve explicit platform placeholders during compilation and rewrite shared agent/skill instructions that otherwise assume only Copilot or Claude.
- **Ownership-safe uninstall:** Remove only Codex artifacts enumerated by the embedded smaqit payload and preserve unrelated agents, skills, and `.codex/config.toml`.
- **No root dogfooding mirror:** Do not add core product artifacts to the repository's root `.codex/` or `.agents/` mirrors because those paths currently have a separate generator and owner.
- **Backward compatibility:** Preserve existing Copilot and Claude formats and behavior. Refactoring their broad uninstall behavior is outside this task.
- **Reference baseline:** Use committed `smaqit-extensions` commit `71a90d2`; ignore unrelated working-tree changes in that repository.

## Implementation Steps

1. Extend `scripts/generate-agents.py` with a data-driven Codex platform, TOML rendering, Codex skill generation, staging cleanup, and complete-platform validation.
2. Add Codex metadata and platform-placeholder values to all nine agent definition manifests.
3. Make shared agent instructions platform-neutral and resolve Codex delegation to named custom subagents.
4. Update affected skills and scripts for `AGENTS.md` project discovery and valid YAML frontmatter.
5. Add Codex embeds, installation mappings, exact conflict handling, validation, help, plan, and status guidance to the Go installer.
6. Implement embedded-allowlist-based Codex uninstall and empty-parent pruning while preserving unrelated shared-directory content.
7. Update build preparation, cleanup, ignored staging paths, and supported build entrypoints.
8. Add an isolated temporary-project installer smoke test and CI/local entrypoint.
9. Update user, framework, installed-project, and testing documentation for Codex.
10. Run the complete verification matrix: generation counts and parsing, Go tests and vet, installer build, smoke test, cross-compilation, and generated-tree consistency checks.

## Known Issues Triage

**Triaged:** 2026-07-22
**Tools searched:** OpenAI Codex, Go, PyYAML
**Result:** Blocking

### Blocking Issues
- [#26408 Project-scoped custom subagent in `.codex/agents` is advertised but cannot be spawned](https://github.com/openai/codex/issues/26408) — `openai/codex` — opened 2026-06-04 — bug, CLI, app, subagent, config
- [#14785 Custom skills do not show up](https://github.com/openai/codex/issues/14785) — `openai/codex` — opened 2026-03-16 — bug, extension, skills

### Advisory Issues
- [#18823 Custom agent requests are easy to misroute to skills instead of `.codex/agents/*.toml` custom agents](https://github.com/openai/codex/issues/18823) — `openai/codex` — opened 2026-04-21 — enhancement, skills, subagent

### Historical (Closed)
- [#25651 Repo-root `AGENTS.md` and `.agents/skills` are not loaded on session start](https://github.com/openai/codex/issues/25651) — `openai/codex` — closed 2026-06-05

## Acceptance Criteria

- [x] Generator supports `copilot`, `claude`, and `codex` as explicit platforms.
- [x] Preparation produces exactly 9 valid Codex agent TOMLs containing non-empty `name`, `description`, and `developer_instructions`.
- [x] All 9 agent manifests define Codex metadata and resolve every platform-specific placeholder.
- [x] Shared agent instructions contain no Claude-only slash-command or Copilot-only delegation guidance in Codex output.
- [x] Preparation produces all 24 skills under `installer/skills-codex/`, with `[SMAQIT_SKILLS_DIR]` resolved to `.agents/skills`.
- [x] Installed Codex skills have parseable frontmatter and recognize `AGENTS.md` wherever project instructions are discovered.
- [x] `smaqit init` installs agents to `.codex/agents/` and skills to `.agents/skills/`.
- [x] Exact Codex destination conflicts are detected before both first installation and reinstallation.
- [x] `smaqit update` refreshes Codex artifacts through the existing fresh-binary reinitialization path.
- [x] `smaqit validate` checks the Codex installation directories.
- [x] `smaqit uninstall` removes only embedded smaqit Codex artifacts and preserves unrelated agents, skills, and `.codex/config.toml`.
- [x] The installer never generates or modifies `.codex/config.toml`.
- [x] CLI help, README, installed `AGENTS.md`, framework documentation, and testing guidance describe Codex accurately.
- [x] A temporary-project smoke test validates generated-tree parity, TOML and skill parsing, conflict handling, reinstallation, validation, and ownership-safe uninstall.
- [x] Go tests, `go vet`, installer build, cross-compilation, and smoke testing pass without regressing Copilot or Claude outputs.

## Findings

**Implementation approach:**
- Extended the canonical generator with a third data-driven platform that emits Codex TOML agents and repository skill staging from existing shared sources.
- Integrated Codex into installer initialization, conflict detection, update reinitialization, validation, help, status, and ownership-safe uninstall behavior.
- Added unit, lifecycle smoke, parsing, CI, and five-target cross-build verification before publishing v1.6.0.

**Decisions made:**
- Used native `.codex/agents/` and `.agents/skills/` discovery without generating or modifying `.codex/config.toml`.
- Kept generated Codex staging ephemeral and left repository-root dogfooding artifacts under their existing separate ownership.
- Removed exact embedded files during uninstall and pruned only empty directories so unrelated and nested custom content survives.

**Blockers encountered:**
- Upstream Codex discovery reports initially blocked task start; the operator approved proceeding with documented formats.
- Local and automated validation passed, and the operator confirmed successful installation and operation in another project.

**Follow-up identified:**
- Monitor upstream Codex project-agent and IDE skill-discovery behavior for regressions; no task-specific implementation work remains.

## Files to Create / Modify

| File | Action |
|------|--------|
| `scripts/generate-agents.py` | Modify |
| `.smaqit/definitions/agents/*.frontmatter.yaml` | Modify |
| `agents/*.md` | Modify where platform-neutral invocation wording is required |
| `skills/smaqit.infrastructure-vault-loader/**` | Modify |
| Affected skill `SKILL.md` files | Modify for valid frontmatter and platform-neutral guidance |
| `installer/main.go` | Modify |
| `installer/update.go` | Verify; modify documentation/tests only if needed |
| `installer/Makefile` | Modify |
| `.gitignore` | Modify |
| `scripts/smoke-test-installer.sh` | Create |
| CI workflow for installer verification | Create or modify |
| `README.md` | Modify |
| `templates/AGENTS.md.template` | Modify |
| `framework/*.md` and relevant `docs/wiki/**` files | Modify |

## Notes

- Official Codex project discovery paths are `.codex/agents/*.toml` for custom agents and `.agents/skills/*/SKILL.md` for repository skills.
- Existing root `.codex/agents/` and `.agents/skills/` artifacts belong to the `smaqit-extensions` dogfooding sync and are intentionally outside this task's generated product payload.
- Strict discovery found three existing skill descriptions requiring YAML quoting or folded form and two credential scripts requiring `AGENTS.md` lookup support.
- Upstream Codex issues #26408 and #14785 were acknowledged by the operator on 2026-07-22 with direction to proceed using the documented artifact formats and retain explicit runtime-limit verification.
