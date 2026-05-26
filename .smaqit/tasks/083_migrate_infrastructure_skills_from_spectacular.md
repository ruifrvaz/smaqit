# Migrate Infrastructure Skills from SPECtacular into smaqit

**Status:** In Progress
**Mode:** Assisted
**Created:** 2026-05-25

## Description

SPECtacular was a proof-of-concept skill library used to validate the `smaqit.project-zero-to-prod` workflow end-to-end on a Hello Mario Node.js + React application deployed to a Cyso Cloud VM. The run was completed successfully and qualified as a success on 2026-05-25.

During that run, a set of new skills was authored and compiled that cover the full infrastructure lifecycle: VM provisioning, deployment, CI/CD generation, domain/TLS, health verification, and cloud provider knowledge. These skills are general enough to belong in `smaqit` and should be migrated here under the `smaqit.infrastructure-*` namespace. The flagship orchestrator skill (`smaqit.project-zero-to-prod`) is renamed to `smaqit.new-greenfield-project` to better reflect intent rather than implementation topology.

Two non-infrastructure skills (requirements extraction, spec status updates) also originated in SPECtacular and belong in `smaqit` without a namespace change. A third skill, `smaqit.hook.session-finish-retro-tasks`, is migrated to `smaqit-extensions` instead (renamed to `smaqit.task-refresh`) — tracked separately.

## Design Decisions

- **All new skills → `smaqit`**: No new repo (smaqit-infrastructure) is created. Adding a namespaced group to an existing markdown-based repo has no meaningful maintenance cost.
- **Naming scheme `smaqit.infrastructure-*`**: Groups all infra-related skills visually and avoids namespace collisions. Cyso-specific skills are included here, not in smaqit-extensions, because they follow the same pattern and are co-dependent with the orchestrator skill.
- **`smaqit.project-zero-to-prod` → `smaqit.new-greenfield-project`**: The new name describes user intent (start a new project from scratch) rather than a pipeline implementation detail.
- **utils stay in SPECtacular**: `smaqit.utils.read-pdf` and `smaqit.utils.triage-issues` are SPECtacular-specific utilities and are not migrated.
- **smaqit-adk skills excluded**: `smaqit.create-skill`, `smaqit.create-agent`, `smaqit.L2` are already owned by smaqit-adk and are not part of this migration.
- **`smaqit.hook.session-finish-retro-tasks` → `smaqit-extensions` as `smaqit.task-refresh`**: This skill is a session-layer hook, not an infrastructure skill. It does not belong in `smaqit` core and is migrated to `smaqit-extensions` under a cleaner name. Tracked separately from this task.

## Implementation Steps

1. For each skill in the migration table below, copy the compiled `SKILL.md` from `SPECtacular/.github/skills/<name>/` into `smaqit/skills/<new-name>/SKILL.md`
2. Rename `name:` field in each SKILL.md frontmatter to match the new skill name
3. Update any cross-skill references within SKILL.md files (e.g., `smaqit.deploy-verify` referenced inside `smaqit.project-zero-to-prod` must be updated to `smaqit.infrastructure-deploy-verify`)
4. Copy definition files from `SPECtacular/.smaqit/definitions/skills/<name>.md` into `smaqit/.smaqit/definitions/skills/<new-name>.md` (if definition files exist)
5. Update `smaqit/skills/smaqit.new-greenfield-project/SKILL.md` to reference all renamed skill names throughout the phase steps
6. Register all new skills in `smaqit`'s `copilot-instructions.md` skills list (if applicable)
7. Open a PR against `smaqit` main with the full migration batch
8. After PR merges, open a follow-up task to clean up SPECtacular (remove migrated skills, update SESSION-REPORT.md to note migration)

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] All 14 skills from the migration table exist under `smaqit/skills/` with correct new names
- [ ] SKILL.md frontmatter `name:` field matches directory name for each migrated skill
- [ ] Internal cross-references within SKILL.md files are updated to new names
- [ ] `smaqit.new-greenfield-project` invokes the correct renamed skill names throughout all phase steps
- [ ] No reference to `cyso.knowledge-base` or `smaqit.project-zero-to-prod` remains in any migrated file
- [ ] PR merged to `smaqit` main

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
| `skills/smaqit.new-greenfield-project/SKILL.md` | Create (from `smaqit.project-zero-to-prod`) |
| `skills/smaqit.infrastructure-deploy-rsync/SKILL.md` | Create (from `smaqit.app-deploy-rsync`) |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md` | Create (from `smaqit.cicd-generate`) |
| `skills/smaqit.infrastructure-repo-config/SKILL.md` | Create (from `smaqit.github-repo-config`) |
| `skills/smaqit.infrastructure-deploy-verify/SKILL.md` | Create (from `smaqit.deploy-verify`) |
| `skills/smaqit.infrastructure-domain-tls/SKILL.md` | Create (from `smaqit.domain-tls`) |
| `skills/smaqit.infrastructure-vm-bootstrap/SKILL.md` | Create (from `smaqit.vm-bootstrap`) |
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Create (from `smaqit.utils-vault-loader`) |
| `skills/smaqit.infrastructure-hook-post-deploy-stamp/SKILL.md` | Create (from `smaqit.hook.post-deploy-stamp`) |
| `skills/smaqit.infrastructure-hook-pre-commit-validate/SKILL.md` | Create (from `smaqit.hook.pre-commit-validate`) |
| `skills/smaqit.infrastructure-provision-cyso/SKILL.md` | Create (from `smaqit.provision-target-cyso`) |
| `skills/smaqit.infrastructure-provider-cyso/SKILL.md` | Create (from `cyso.knowledge-base`) |
| `skills/smaqit.requirements-extract/SKILL.md` | Create (from `smaqit.requirements-extract`, no rename) |
| `skills/smaqit.spec-status-update/SKILL.md` | Create (from `smaqit.spec-status-update`, no rename) |

## Notes

Source repository: `ruifrvaz/SPECtacular` — all skills compiled at commit `59fcc1b` (2026-05-25).

**Scope adjustment (session 2026-05-26):** `smaqit.hook.session-finish-retro-tasks` excluded from this migration per user instruction. The 14 remaining skills were migrated; AC updated accordingly.

Migration table (current name → new name):

| Current (SPECtacular) | New (smaqit) |
|---|---|
| `smaqit.project-zero-to-prod` | `smaqit.new-greenfield-project` |
| `smaqit.app-deploy-rsync` | `smaqit.infrastructure-deploy-rsync` |
| `smaqit.cicd-generate` | `smaqit.infrastructure-cicd-generate` |
| `smaqit.github-repo-config` | `smaqit.infrastructure-repo-config` |
| `smaqit.deploy-verify` | `smaqit.infrastructure-deploy-verify` |
| `smaqit.domain-tls` | `smaqit.infrastructure-domain-tls` |
| `smaqit.vm-bootstrap` | `smaqit.infrastructure-vm-bootstrap` |
| `smaqit.utils-vault-loader` | `smaqit.infrastructure-vault-loader` |
| `smaqit.hook.post-deploy-stamp` | `smaqit.infrastructure-hook-post-deploy-stamp` |
| `smaqit.hook.pre-commit-validate` | `smaqit.infrastructure-hook-pre-commit-validate` |
| `smaqit.provision-target-cyso` | `smaqit.infrastructure-provision-cyso` |
| `cyso.knowledge-base` | `smaqit.infrastructure-provider-cyso` |
| `smaqit.requirements-extract` | `smaqit.requirements-extract` |
| `smaqit.spec-status-update` | `smaqit.spec-status-update` |

**Not in this task** (goes to `smaqit-extensions`):

| Current (SPECtacular) | New (smaqit-extensions) |
|---|---|
| `smaqit.hook.session-finish-retro-tasks` | `smaqit.task-refresh` |
