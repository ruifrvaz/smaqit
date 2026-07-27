# Changelog

All notable changes to smaqit will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Removed
- `smaqit.feature-deploy` — standalone post-MVP deployment skill retired. Post-MVP feature deployment is consolidated into `smaqit.feature-new` as the sole workflow entry point. Deployment is mandatory and uses a pull request as the human approval gate.

### Changed
- `smaqit.feature-new` — consolidated as the single post-MVP feature/deployment workflow. Removed deploy-now/defer branching; deployment is mandatory via PR-gated CI/CD. Added durable Phase 1 spec handoff and `specification_mode: prevalidated` for Phases 2–4 so specification agents run once per feature. Added deterministic trigger decision table, vault-loader/repo-config preflight, and own amendment scanner.
- Development, Deployment, and Validation agents — added `specification_mode: orchestrate|prevalidated` parameter (default `orchestrate` preserves existing behavior). Deployment agent additionally supports `deployment_path: standard|existing-cicd-pr`.
- `smaqit.infrastructure-deploy-verify` — added optional `--expected-sha` parameter; defaults to `git rev-parse HEAD` (backward-compatible).

### Fixed
- Duplicate specification orchestration: `smaqit.feature-new` Phase 1 now records a durable exact-path handoff; Development, Deployment, and Validation agents consume it in `prevalidated` mode.
- Unsafe direct-push assumption: trigger decision table inspects committed workflow files before PR creation and rejects ambiguous or duplicate trigger layouts.

## [1.9.0] - 2026-07-24

### Added
- `smaqit.feature-deploy` — standalone post-MVP deployment skill. Extracts greenfield's Phase 4/5 provisioning_mode branching (`provision`, `existing-owned`, `existing-shared`) into a self-contained 3-phase workflow: pre-phase feature identification (reads stack and infrastructure specs), Phase 1 infrastructure readiness (vault-loader, provision-cyso, vm-bootstrap), Phase 2 CI/CD-triggered deploy (push to main, monitor pipeline, deploy-verify, amendment gate), Phase 3 close-out with optional release tagging. Deploys exclusively through existing CI/CD pipelines — no dev-VM sweep, no deploy-rsync skills. Ships its own `scripts/check-amendments.sh` (self-contained, no cross-skill file references).

### Changed
- Bumped shipped skill count from 25 to 26 in installer test assertions (`installer/main_test.go`, `scripts/smoke-test-installer.sh`).

### Fixed
- Removed external project references from history files, task files, CHANGELOG, and shipped skill documentation.

## [1.8.0] - 2026-07-23

### Added
- New Vault credential namespace: `secret/apps/<app-slug>/*` (`ssh`, `github`, `machine`) and `secret/machines/<machine-slug>/*` (`base-ssh`, `cyso`, `tfstate`, `metadata`), giving a provisioned machine an identity of its own, separate from any one app deployed onto it. Every app now gets its own distinct SSH keypair — no exceptions, including the project that originally provisions the machine — bootstrapped against the machine's `base-ssh` credential rather than sharing or copying another app's key.
- `smaqit.infrastructure-vault-loader` script `bootstrap-app-to-machine.sh <app-slug> <machine-slug>`: idempotently registers a machine (generating its `base-ssh` credential on first use) and bootstraps an app's own keypair onto it, authorizing the new key via the machine's base credential and verifying it authenticates before reporting success.

### Changed
- `smaqit.infrastructure-vault-loader`'s `load-credentials.sh` and `rotate-credential.sh` auto-detect the new `apps/`+`machines/` scheme vs. the legacy flat `secret/<project-slug>/*` scheme per invocation, so unmigrated projects using the old flat scheme keep working unchanged. `rotate-credential.sh` gains scheme-specific handling for `apps/<app-slug>/{ssh,github}` and `machines/<machine-slug>/{cyso,tfstate,base-ssh}`, including a bespoke generate/install-with-old-key/retire flow for rotating a machine's `base-ssh` without touching any already-bootstrapped app's access.
- `smaqit.infrastructure-provision-cyso`: Step 2 now resolves a machine slug, generates and stores `secret/machines/<machine-slug>/base-ssh` on first use, and sources the Terraform SSH public key from it; a new Step 7.5 registers machine metadata and bootstraps the provisioning project's own app-specific key via `bootstrap-app-to-machine.sh` — the provisioning project gets no shortcut around having its own distinct key.
- `smaqit.infrastructure-repo-config`'s `scripts/sync-secrets.sh` and preconditions are now scheme-aware: `ssh`/`github` always read from the app root, `tfstate`/`cyso` always read from the machine root (the same path as the app root on the legacy scheme).
- `smaqit.new-greenfield-project` acceptance criteria and `existing-shared` deployment guidance updated to reference the app/machine bootstrap flow instead of the legacy shared-SSH-key mechanisms.

### Deprecated
- Nothing to add.

### Removed
- Nothing to add.

### Fixed
- Installer's Codex skill-count assertions (`installer/main_test.go`, `scripts/smoke-test-installer.sh`) updated from a stale hardcoded 24 to 25, matching `smaqit.feature-new`'s addition as the 25th shipped skill. This mismatch was blocking CI on the v1.7.0 release commit.

### Security
- Nothing to add.

### Chore
- Nothing to add.

## [1.7.0] - 2026-07-23

### Added
- `smaqit.feature-new` skill: a task-per-phase workflow for post-MVP iterative feature cycles (Spec Revalidation, Development, Deployment, Validation, Close-out), closing the gap `smaqit.new-greenfield-project` explicitly leaves for post-MVP work. Applies the same task-per-phase and amendment-gate discipline as greenfield, without requirements extraction, from-scratch spec generation, or a dev-VM sweep. Deployment defaults to the project's existing target (only provisioning a new one when no deployed Infrastructure spec exists) and supports an explicit deploy-now/defer choice; the amendment gate (`check-amendments.sh`, referenced not duplicated) blocks Deployment-phase completion on unresolved amendments in both paths.

### Changed
- Nothing to add.

### Deprecated
- Nothing to add.

### Removed
- Nothing to add.

### Fixed
- Nothing to add.

### Security
- Nothing to add.

### Chore
- Nothing to add.

## [1.6.0] - 2026-07-22

### Added
- First-class OpenAI Codex compatibility in the smaqit installer:
  - All 9 canonical agents compile to Codex project-agent TOML under `.codex/agents/` with non-empty `name`, `description`, and `developer_instructions`.
  - All 24 canonical product skills install under `.agents/skills/`, with platform paths resolved during generation.
  - `smaqit init`, reinstallation, update reinitialization, validation, status/help guidance, and uninstall now include Codex alongside GitHub Copilot and Claude Code.
  - Codex uninstall removes exact smaqit-owned files while preserving unrelated agents, skills, nested custom content, and `.codex/config.toml`.
  - Go unit tests, a temporary-project installer lifecycle smoke test, CI coverage, and Linux arm64 cross-compilation were added.
- Repository-local Codex workflow support synced from `smaqit-extensions`: 28 project/session/task/release/testing utility skills plus `smaqit.release.local`, `smaqit.release.pr`, and `smaqit.user-testing` custom agents.

### Changed
- Shared agent, skill, framework, installed-project, and user documentation now use platform-neutral invocation guidance and recognize `AGENTS.md` when discovering project instructions.
- Installed specification-template footers now identify smaqit v1.5.1 instead of the former beta version.

### Deprecated
- Nothing to add.

### Removed
- Nothing to add.

### Fixed
- Three existing skill descriptions now use strict, parseable YAML frontmatter so all installed Codex skills pass metadata validation.

### Security
- Nothing to add.

### Chore
- Added session history and compendium documentation for the v1.5.1 self-update reinitialization fix.

## [1.5.1] - 2026-07-21

### Fixed
- `smaqit update` no longer silently skips new skills/scripts added in the release just downloaded. It replaced the binary on disk but then re-initialized project assets in the same still-running process — `go:embed` content is baked in at compile time, so the old process only had the previous release's embedded content in memory even though the file on disk was already the new version. Reinit now re-execs the freshly-downloaded binary as a subprocess, so it always reflects what was actually just installed.

### Chore
- Closed Task 084 (deploy-target provisioning-mode branching) findings, documenting the one knowingly-accepted gap (no live `existing-shared` co-hosting walkthrough performed yet).

## [1.5.0] - 2026-07-21

### Added
- **Deploy-target provisioning modes** — `smaqit.input-deployment` gained a new "Provisioning Mode" parameter with three values: `provision` (default, this project provisions its own VM), `existing-owned` (redeploy to a VM this project's own Terraform state already manages), and `existing-shared` (deploy onto a VM a *different* project owns and manages, e.g. co-hosting). `smaqit.new-greenfield-project` Phase 4/5 now branches explicitly on this value at every step where behavior differs.
- `smaqit.infrastructure-provision-cyso/scripts/ownership-guard.sh` — pre-flight guard stopping a direct/manual invocation of the skill from silently provisioning a second VM when a target is already declared (via `VM_HOST`) but not owned by this project's Terraform state — the `existing-shared` case.
- Mode-aware credential loading in `smaqit.infrastructure-vault-loader`: `existing-shared` skips the `cyso`/`tfstate` Vault paths entirely and offers two mechanisms for the SSH key — copy it from the owning project's Vault namespace, or generate a new key and manually append it to the shared VM's `authorized_keys`.
- Mode-aware `smaqit.infrastructure-repo-config`: restricted-mode secret sync for `existing-shared` (only `ssh`/`github`-derived secrets populated; `tfstate`/`cyso` are skipped cleanly, not treated as an error); `VM_HOST` changed from a GitHub Actions secret to a **variable**, since it's non-sensitive and needs to be readable back by `ownership-guard.sh`.
- `full` / `deploy-only` generation modes for `smaqit.infrastructure-cicd-generate`, matching `provisioning_mode` — `deploy-only` skips all Terraform generation (`provision.yml` is not generated at all) for `existing-shared` targets.
- Co-hosted-VM nginx vhost rule in `smaqit.infrastructure-deploy-rsync`: only the first site deployed to a VM may claim `default_server`; every subsequent co-hosted site's vhost must be name-based only, checked against `/etc/nginx/sites-enabled/` before writing the conf.
- Deterministic CI/CD generation (**partial** — see Notes below): real template assets (`deploy.yml.deploy-only.template`, `deploy.yml.full.template`, `post-merge-deploy.yml.template`, `provision.yml.template`) in `smaqit.infrastructure-cicd-generate`, replacing prose-based YAML generation.
- `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh` and `smaqit.infrastructure-repo-config/scripts/sync-secrets.sh` — deterministic scripts vendored into target projects, replacing agent-interpreted prose for the two riskiest judgment calls (vhost `default_server` assignment, and Vault→GitHub-Secrets sync with `existing-shared` skip-if-absent logic).
- New canonical skill `smaqit.infrastructure-deploy-rsync-python-nextjs`, including `scripts/run-migrations.sh` — a real, validated Python/FastAPI + Next.js rsync deploy skill reconciled from a downstream project.
- Deterministic stack-detection + on-the-fly deploy-skill synthesis in `smaqit.new-greenfield-project` Phase 4 Step 6 — reads the declared stack from the stack spec (authoritative), judges it against currently-installed `smaqit.infrastructure-deploy-rsync*` skills generically (no hardcoded enumeration), and on no match synthesizes a new deploy skill via `smaqit.create-skill`/`smaqit.L2` with a mandatory human checkpoint before the synthesized skill is ever invoked. Live-validated end-to-end against a real downstream project on an unmatched stack (Tornado/systemd, no Docker).
- `smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh` — deterministic guard that runs `terraform plan`, inspects the machine-readable output for any `delete`/`replace` action, and exits non-zero naming the specific resource(s) before `terraform apply` is ever reached.
- `smaqit.infrastructure-vm-bootstrap/scripts/remove-default-nginx-site.sh` — idempotent removal of the distro's stock `default` nginx site, so it can never conflict with the application's own `default_server` vhost once a second, co-hosted site is added.
- `smaqit.infrastructure-provider-cyso/references/openstack-forcenew-attributes.md` — documents which `openstack_compute_instance_v2` attributes (`user_data`, `image_id`/`image_name`, `key_pair`) are `ForceNew`, including the gotcha where a comment added inside a `file()`-referenced cloud-init file forces replacement even though the same comment as an HCL comment would not.
- `smaqit.infrastructure-provider-cyso/references/github-provider-import-ids.md` — documents the `integrations/github` provider's `<repository>:<name>` import ID format for `github_actions_variable`/`github_actions_secret`, and the "pick one owner" rule for Terraform-vs-manual resource management.
- New "IaC Drift Prevention" section in `agents/deployment.md` — MUST NOT patch a live IaC-managed resource out-of-band without reconciling IaC config/state in the same change; MUST route every plan/apply against already-provisioned infrastructure through the project's guard script, including diagnostic-only checks with no intent to apply. Matching Gotcha added to `smaqit.infrastructure-provision-cyso`.
- Claude Code skill parity completed: all `.github/skills/*` skills mirrored into `.claude/skills/*` (28 skills), plus 3 new `.claude/agents/*` (`release.local`, `release.pr`, `user-testing`) with corresponding updates to their existing `.github/agents/*` counterparts.
- 7 brand-new skills landed in both `.github/` and `.claude/` simultaneously: `smaqit.parity-assess`, `smaqit.project-compendium`, `smaqit.project-diagnose`, `smaqit.task-plan`, `smaqit.task-refresh`, `smaqit.test-complete`, `smaqit.test-create`.

### Changed
- `smaqit.infrastructure-provision-cyso` — plan review step now requires invoking `plan-guard.sh`, never a bare `terraform plan` eyeballed by the agent; Step 0 runs `ownership-guard.sh` before touching Vault or Terraform; documents the canonical-vs-vendored relationship with `smaqit.infrastructure-cicd-generate` (guard scripts are copied verbatim into a target project at generation time — never hand-edit the vendored copy). (v1.0.0 → v1.4.1)
- `smaqit.infrastructure-provider-cyso` — load-condition table now routes to the two new reference files; description and Gotchas/Failure Handling updated to match. (v1.0.0 → v1.1.1)
- `smaqit.infrastructure-vm-bootstrap` — added a bootstrap step invoking `remove-default-nginx-site.sh`; documented that `systemctl enable nginx` does not start the service immediately, so downstream deploy skills must use `reload-or-restart`, not a bare `reload`. (v1.0.0 → v1.1.1)
- `smaqit.infrastructure-cicd-generate` — generated `deploy.yml`/`provision.yml` now gate every `terraform apply` behind `plan-guard.sh`; the generated nginx reload step uses `reload-or-restart`; generation now branches on `full`/`deploy-only` mode and uses real template assets instead of prose. (v1.0.0 → v2.0.0)
- `smaqit.infrastructure-repo-config` — refactored to call `sync-secrets.sh` instead of inline `vault kv get | gh secret set` steps; restricted-mode handling for `existing-shared`. (v1.0.0 → v1.3.0)
- `smaqit.infrastructure-vault-loader` — mode-aware credential loading for `existing-shared`; gained `~/.vault-token` auto-reuse (skips re-prompting if a valid local token already exists). (v3.0.0 → v3.2.0)
- `smaqit.new-greenfield-project` — Phase 4 pre-conditions and Phase 4/5 steps now branch on `provisioning_mode`; Phase 4 Step 6 rewritten for deterministic stack-detection + skill synthesis (see Added). (v1.0.0 → v1.4.1)

### Fixed
- Two real `smaqit.infrastructure-vault-loader` bugs found during a live first-run: SSH private keys losing their required trailing newline via `$(cat file)` command substitution when written to Vault (`error in libcrypto` on every subsequent use) — fixed to use Vault's `@file` syntax, which preserves exact bytes; a project-slug misdetection where an unfilled `AGENTS.md` template placeholder (`[TODO: add project name]`) was parsed as a real slug, silently misdirecting credential writes to the wrong Vault path.

### Security
- Removed inadvertently-committed project-specific configuration from templates and repository history (11 canonical skill template files genericized: hardcoded IPs, nginx site names, health-check paths, spec filenames, Terraform resource labels, and example domains replaced with generic placeholders; the one commit message that itself contained the leaked name was reworded).

### Chore
- Session history and compendium documentation updates; a follow-up commit-SHA citation fix.

### Notes
- Tasks 084 (provisioning-mode branching), 086 (Python/FastAPI + Next.js deploy skill), and 087 (dynamic stack detection + skill synthesis) are shipped and verified against real live validation runs, though not yet formally closed via `task-complete`.
- Task 085 (deterministic CI/CD templates) ships in this release as **partial**: the template-based generation mechanism itself has not yet been exercised against a live target end-to-end, and only 2 of its 4 new scripts (`write-vhost.sh`, `sync-secrets.sh`) have been live-validated; `plan-guard.sh`/`ownership-guard.sh` were validated separately as part of Task 084/087's live runs.

## [1.4.0] - 2026-07-16

### Added
- `smaqit update` command — self-updates the installed binary to the latest GitHub release. Fetches release metadata, compares semantic versions, downloads the matching platform asset, atomically replaces the running binary, and re-initializes `.smaqit/` project assets if present.

### Changed
- Nothing to add.

### Deprecated
- Nothing to add.

### Removed
- Nothing to add.

### Fixed
- Nothing to add.

### Security
- Nothing to add.

### Chore
- Nothing to add.

## [1.3.1] - 2026-07-16

### Added
- Claude Code support alongside GitHub Copilot — `smaqit init` now installs `.claude/agents/`, `.claude/commands/`, and `.claude/skills/` in addition to the existing `.github/agents/`, `.github/skills/`, `.github/workflows/`
- 4 Claude Code slash commands: `/smaqit.development`, `/smaqit.deployment`, `/smaqit.validation`, `/smaqit.qa`
- `scripts/generate-agents.py` — compiles agents, commands, and skills from a single source (`agents/`, `commands/`, `skills/`, `.smaqit/definitions/agents/`) into platform-specific installer output (`installer/agents-copilot/`, `installer/agents-claude/`, `installer/commands-claude/`, `installer/skills-copilot/`, `installer/skills-claude/`)
- Copilot → Claude Code tool-mapping reference table in `docs/wiki/agent-tools-reference.md`
- `smaqit init` now installs `AGENTS.md` (read natively by GitHub Copilot) and a thin `CLAUDE.md` (`@AGENTS.md` import — Claude Code does not read `AGENTS.md` on its own). Existing files are never overwritten; smaqit's section is appended if not already present.
- `.claude/settings.json` hook parity with `.github/hooks/` for this repo's own development tooling (not shipped to installed projects)

### Changed
- Agent source of truth restructured: `agents/<name>.md` (shared body, with `{{PLACEHOLDER}}` tokens for platform-varying phrasing) + `.smaqit/definitions/agents/<name>.frontmatter.yaml` (per-platform frontmatter/metadata), replacing one hand-written `agents/*.agent.md` file per agent
- Skills' self-referential install paths now use a `[SMAQIT_SKILLS_DIR]` placeholder resolved at compile time, instead of a hardcoded `.github/skills/...` path
- Skills that read `copilot-instructions.md` (vault-loader, cicd-generate, repo-config, deploy-rsync, provision-cyso) now check `CLAUDE.md` first, falling back to `.github/copilot-instructions.md`
- `installer/main.go` and `installer/Makefile` updated to compile and install both platforms' agents, commands, and skills; `installer/agents/` and `installer/skills/` renamed to `installer/agents-copilot/` and `installer/skills-copilot/` for naming symmetry with the new `-claude` output directories

### Fixed
- `qa` agent's frontmatter `name` field corrected from `qa` to `smaqit.qa`, matching the naming convention every other agent follows

### Chore
- Nothing to add.

## [1.2.0] - 2026-05-26

### Added
- 14 new skills migrated from SPECtacular under `smaqit.infrastructure-*` namespace
  - `smaqit.infrastructure-cicd-generate` — generate canonical 3-workflow GitHub Actions CI/CD set
  - `smaqit.infrastructure-deploy-rsync` — deploy Node.js + React app to VM via rsync
  - `smaqit.infrastructure-deploy-verify` — verify deployment health after any deploy
  - `smaqit.infrastructure-domain-tls` — configure custom domain with Let's Encrypt TLS via Certbot
  - `smaqit.infrastructure-hook-post-deploy-stamp` — write deploy stamp files (DEPLOY_SHA, DEPLOY_TIME) to VM
  - `smaqit.infrastructure-hook-pre-commit-validate` — pre-commit validation hook (secrets, draft specs, large files)
  - `smaqit.infrastructure-provider-cyso` — Cyso Cloud (OpenStack) reference and pre-flight checks
  - `smaqit.infrastructure-provision-cyso` — provision cloud infrastructure on Cyso Cloud via Terraform
  - `smaqit.infrastructure-repo-config` — configure GitHub repository secrets and variables for CI/CD
  - `smaqit.infrastructure-vault-loader` — load project secrets into local HashiCorp Vault
  - `smaqit.infrastructure-vm-bootstrap` — bootstrap a fresh VM for application deployment
  - `smaqit.new-greenfield-project` — orchestrate full zero-to-prod workflow for new projects (renamed from `smaqit.project-zero-to-prod`)
  - `smaqit.requirements-extract` — extract structured requirements from raw user input
  - `smaqit.spec-status-update` — update spec file lifecycle status fields

### Changed
- `installer/main.go` embed directive fixed: `skills/**/*.md` → `skills` (recursive embed for nested reference dirs and scripts)
- `installer/skills/` added to `.gitignore` alongside other installer embed dirs; previously tracked `smaqit.input-*` files removed from git index
- `.github/copilot-instructions.md` updated: installer subdirectories documented as gitignored, `make sync` workflow documented

### Deprecated
- Nothing to add.

### Removed
- `installer/skills/smaqit.input-*/SKILL.md` removed from git tracking (files remain on disk, now gitignored)

### Fixed
- Nothing to add.

### Security
- Nothing to add.

### Chore
- Task 083 closed
- Session 056 history documented

## [1.1.0] - 2026-05-17

### Added
- Autonomous and assisted (maker-checker) execution modes added to all 3 phase agents (development, deployment, validation)
  - Autonomous mode: spec agents invoked in sequence without user breaks
  - Assisted mode: maker-checker loop with user as checker; max 3 iterations per spec layer

### Changed
- Phase agents (development, deployment, validation) rewritten with orchestration-first workflow — spec generation is always the primary first step, not a conditional fallback
- Pre-Orchestration Validation replaced: upstream spec presence check removed; context sufficiency check added (requirements present, actionable, no conflicts)
- Spec agent invocation sequence hardcoded per phase with scoped context passing (user requirements + upstream specs only)
- `smaqit plan` scoped to implementation routing only; empty output interpreted as "specs up to date — proceed to implementation"
- All 5 spec agent Role sections updated: added note that session context includes requirements propagated from orchestrating phase agent
- `framework/PHASES.md` updated: orchestration-first as primary workflow, deterministic routing, context scoping documented
- `framework/AGENTS.md` Phase Orchestration section rewritten: orchestration-first, deterministic routing, scoped context, iteration caps, autonomous/assisted modes

### Deprecated
- Nothing to add.

### Removed
- Nothing to add.

### Fixed
- Nothing to add.

### Security
- Nothing to add.

### Chore
- CHANGELOG_TEMPLATE.md added to release-prepare-files skill; 7-section structure locked
- Task 082 closed

## [1.0.0] - 2026-05-16

### Added
- 8 `smaqit.input-*` skills — per-layer and per-phase requirement validation gates (PR #69)
  - `smaqit.input-business`, `smaqit.input-functional`, `smaqit.input-stack`, `smaqit.input-infrastructure`, `smaqit.input-coverage` (spec layers)
  - `smaqit.input-development`, `smaqit.input-deployment`, `smaqit.input-validation` (implementation phases)
  - Each skill validates required inputs, checks existing specs, and detects conflicts before execution begins

### Changed
- All 8 specification and phase agents now invoke their corresponding input skill as a validation gate before execution (PR #69)
- Consolidated release workflows into single `post-merge-release.yml` — handles both tag-push (local) and PR-merge (CI) release paths (PR #69)
- `framework/AGENTS.md` and `framework/SKILLS.md` updated to reflect input skill pattern and removal of assessment skill (PR #69)

### Removed
- Prompts feature fully deprecated — `prompts/`, `framework/PROMPTS.md`, `templates/prompts/` and all prompt files removed (PR #69)
  - Input was always read from session context; prompt files were unused placeholders
- Assessment skill (`skills/assessment/`) deprecated — conflict detection and sufficiency validation absorbed into input skills; residual ambiguity handled inline by agents (PR #69)

### Fixed
- Stray `prompt_version` and prompt file references cleaned from agents, framework, and spec templates (PR #69)

### Chore
- Disabled e2e-test workflow automatic triggers; workflow_dispatch only (PR #69)
- Dev environment `.github/` agent and skill files updated (PR #69)

## [0.9.0] - 2026-04-25

### Added
- Extended tool capabilities for specification agents (PR #60)
  - Specification agents now have broader tool access for richer workflows

### Changed
- Copilot instructions updated: removed stale L0/L1/L2 sections, updated Kit Components and Installer description (Session 054)

### Removed
- Retired `smaqit.release` monolithic agent, superseded by `.local`/`.pr` variants and 4-skill decomposition (Session 053)
- Retired `doc-helper` agent, superseded by `smaqit.qa` (Session 053)
- Deleted L0/L1/L2 agent files after smaqit-adk extraction (Session 054)
- Deleted generic agent templates (`base`, `specification`, `implementation`) after smaqit-adk extraction (Session 054)
- Deleted generic compiled rules (`base.rules.md`, `specification.rules.md`, `implementation.rules.md`) after smaqit-adk extraction (Session 054)

### Fixed
- `smaqit.qa` agent tools list corrected to read-only (`edit` removed) (Session 053)

## [0.8.2-beta] - 2026-04-05

## [0.8.1-beta] - 2026-02-16

### Fixed
- Installer now handles existing `.smaqit` directory gracefully (PR #50)
  - Detects which files would be overwritten before proceeding
  - Preserves user data in specs and custom extensions
  - Prompts for confirmation when conflicts detected
  - Skips installation if only custom files exist
  - Makes reinstallation and version upgrades safe

## [0.8.0-beta] - 2026-02-09

### Added
- Specification lifecycle directives compiled to all specification agents (Task 079)
  - Specs automatically revert to `draft` status when acceptance criteria modified
  - Ensures modified specs proceed through revalidation phases
  - Applied to all 5 specification agents: Business, Functional, Stack, Infrastructure, Coverage
- Assessment skill trigger detection added to frontmatter
  - Skill now responds to explicit keywords: "assess", "assessment", "evaluate", "analyze"
  - Improves automatic invocation when users request critical assessment

### Changed
- Strengthened "Bounded Agents" principle to explicitly state scope enforcement is self-governing (Task 069)
  - Added clarification that external framing, task specifications, or grouped work descriptions cannot override agent scope
  - Emphasized that agents stop at boundaries and redirect when requests span scopes
  - Addresses edge case where external framing creates pressure to violate scope boundaries

## [0.7.0-beta] - 2026-02-08

### Added
- Assessment skill for critical evaluation before execution (Task 078)
  - Automatic invocation when agents detect ambiguous requirements, conflicting inputs, insufficient detail, or complex planning scenarios
  - Five-step workflow: Question premise → Check state → Identify trade-offs → Flag problems → Present assessment
  - Integrated into all 8 product agents with `.github/skills/assessment/` invocable capability
- Phase orchestration in implementation agents (Task 073)
  - Development, Deployment, and Validation agents now coordinate entire phase workflows
  - Agents automatically invoke specification agents in dependency order using `runSubagent` tool
  - Pre-orchestration validation checklist (12 checks)
  - 7-step orchestration workflow
  - Orchestration completion validation checklist (11 checks)
- Multi-format compilation support in Agent-L1
  - Supports 7 format types: directive, checklist, workflow, table, role, structure, frontmatter
  - Format inference from L0 content patterns
- Development binary (`smaqit-dev`) for framework development
- Quickstart guide with Mario Hello tutorial
- Team alignment wiki documentation
- LICENSE (MIT) and CONTRIBUTING.md

### Changed
- **User workflow simplified**: `/smaqit.development` now coordinates spec generation internally instead of requiring manual `/smaqit.business` → `/smaqit.functional` → `/smaqit.stack` sequence (Task 073)
- README restructured from 288 lines to 75 lines for clarity
- Agent-L2 compilation upgraded from 3-way to 4-way merge (base + extension type + specific role)
- L1 templates refactored to pure placeholder structure (Task 065)
  - Created specification.rules.md and implementation.rules.md compilation files
  - Base.rules.md refined to foundation directives only
- User compilation logs moved to `.smaqit/logs/` (was `agents/logs/`)

### Removed
- Orchestrator agent pattern (Task 072)
  - Never documented in user-facing commands, no workflow impact
  - Orchestration capabilities distributed into implementation agents

## [0.6.2-beta] - 2026-01-21

### Added
- Compilation file architecture: 8 compilation files (5 layers + 3 phases) for L0→L1 transformation (Task 068)
- Wiki documentation: `why-no-system-actor.md` and `why-non-functional-requirements.md`
- Source L0 Principles table in all compilation files for traceability

### Changed
- **BREAKING:** System Actor pattern removed from Business layer (Task 068)
- Business layer boundary enforcement strengthened with concept-based directives
- Compilation files now contain pure L1 directives without inline L0 Source citations
- Business layer MUST NOT directive refined: "Describe HOW features work" (replaces word blacklist)
- NFR terminology changed from "system property advocates" to "non-functional requirement stakeholders"
- Agent-L1 updated to enforce clean directive compilation
- Specification agent template simplified (Actor Concept placeholder removed)

### Fixed
- Business specs no longer leak functional/stack concerns through System Actor pattern
- Behavioral verbs directive now focuses on concept boundary (HOW vs WHAT) instead of vocabulary blacklist

## [0.6.0-beta] - 2026-01-17

### Added
- Troubleshooting documentation for multi-agent workflow context pollution (Task 056)

### Changed
- Specification agents now reset acceptance criteria checkboxes when modifying requirements (Task 060)
- Implementation agents update all referenced spec frontmatter, not just target layer (Tasks 061, 063)
- Validation agent generates executable test artifacts for CI/CD automation (Task 062)
- Implementation agents execute CLI as first action to determine which specs to process (Tasks 049, 051, 052)
- Coverage layer redesigned with dual-input model: test requirements + upstream criteria (Task 050)
- Agent Role sections refined to 3-component structure for clarity (Task 056)
- Foundation Reference pattern unified across all specification layers (Task 055)

### Fixed
- Validation agent frontmatter updates now apply to all validated specs across all layers (Task 053)
- CLI directive ambiguity resolved in all implementation agents (Tasks 049, 051, 052)

### Removed
- Context pollution verbal statements pattern (replaced with structured Role sections) (Task 056)

## [0.5.0-beta] - 2026-01-03

### Added
- **Stateful Specifications** (Task 014)
  - YAML frontmatter state tracking in all specs (id, status, created, timestamps, prompt_version)
  - Spec lifecycle states: draft → implemented → deployed → validated → failed/deprecated
  - Acceptance criteria checkbox updates: `[ ]` → `[x]` (passed) or `[!]` (failed)
  - Phase reports generated in `.smaqit/reports/` directory
  - Stale spec detection via prompt_version tracking (git commit hash)
  - Wiki documentation for stateful specifications and stale management workflows

- **Incremental Processing** (Task 047)
  - `smaqit plan --phase=[develop|deploy|validate]` command
    - Outputs spec file paths requiring processing (one per line)
    - Default: returns only specs with `status: draft` or `status: failed`
    - `--regen` flag: returns all specs regardless of status
  - Implementation agents now skip already-completed specs
  - Frontmatter as single source of truth (removed dual state system)
  - CLI scans specs on-demand and aggregates status
  - Strict phase completion rules: requires ALL layers present + ALL specs at target status

### Changed
- Agents refactored to directive-based instructions (pure MUST/MUST NOT rules)
  - Removed procedural "State-Based Processing" workflows
  - Simplified agent instructions for better LLM interpretation
- Phase completion detection now requires all layers present
  - Develop: business + functional + stack specs required
  - Deploy: infrastructure specs required
  - Validate: coverage specs required
- Framework documentation updated to remove example pollution
  - Generic placeholders ([ID], [CONCEPT]) replace specific examples
  - Templates remain abstract and reusable

### Removed
- state.json aggregate state file (replaced by on-demand CLI scanning)
- Dual state system complexity (frontmatter is sole source of truth)
- Example pollution from templates (BUS-LOGIN-001, etc.)

## [0.4.2-beta] - 2026-01-02

### Added
- Intelligent next-step suggestions in `smaqit status` command
  - Progressive guidance based on actual spec file presence
  - Suggests missing spec layers before implementation (business → functional → stack)
  - Only suggests `/smaqit.development` when all Phase 1 specs exist
  - Phase-aware suggestions for infrastructure and coverage layers

### Fixed
- State.json phase ordering now consistent (develop → deploy → validate)
- Status command no longer suggests premature implementation steps

## [0.4.1-beta] - 2026-01-02

### Added
- One-liner installation script (`install.sh`)
  - Platform detection for Linux, macOS, and Windows
  - Automatic installation to `~/.local/bin`
  - Version selection via `SMAQIT_VERSION` environment variable (latest/prerelease/vX.Y.Z)
  - Installation verification and PATH checking
- Standard CLI flag support
  - `--version` and `-v` flags now work alongside `version` subcommand
  - `--help` and `-h` flags work alongside `help` subcommand
  - Consistent with standard tools (go, python, etc.)

### Fixed
- Install script stdout contamination from info messages during download
- Repository visibility (made public for install script access)

## [0.4.0-beta] - 2026-01-01

### Added
- GitHub Actions workflow for automated releases (Task 022)
  - Automatic builds for Linux, macOS (Intel/ARM), Windows
  - SHA256 checksums generation
  - Release notes extracted from CHANGELOG.md
- Manual workflow dispatch for releases via GitHub UI
- Changelog management system using session history
  - `/changelog.update` agent reads `.smaqit/history/` and updates CHANGELOG.md
  - AI-managed changelog following Keep a Changelog format
  - Simplified release process documented in README

## [0.3.0] - 2025-12-28

### Added
- Explicit scope boundaries for all agents (Task 041)
  - Agents now enforce layer/phase boundaries with Stop → Respond → Suggest pattern
  - Prevents agents from executing out-of-scope work
- User vs agent documentation distinction (Task 040)
  - New wiki document explaining separation between agent-facing specs and user-facing docs
  - Guidelines for what content belongs in framework vs wiki
- Agent handover guidance (Task 039)
  - Agents provide clear next steps when completing work
- State.json validation in validate command (Task 038)
  - Defensive validation for phase completion tracking
- Phase-first workflow clarification (Task 037)
  - Updated PHASES.md to emphasize phase-first as recommended approach
- Use case identifiers to business specs (Task 034)
  - Business specs now include UC-XXX identifiers for traceability
- Nested status display (Task 035)
  - `smaqit status` shows layers grouped under phases

### Changed
- Prompt naming consistency (Task 044)
  - Updated all references from `.develop`/`.deploy`/`.validate` to `.development`/`.deployment`/`.validation`
  - Installer messages, help text, and documentation now consistent
- Implementation prompts simplified (Task 029)
  - Renamed for consistency with agent names
  - Reduced to minimal orchestration inputs
- Session and task commands moved to prompts (Task 030)
  - `/session.recap`, `/session.wrap`, `/task.*` now in `.github/prompts/`

### Fixed
- Prompt name references in installer and documentation (Task 044)
- State.json phase ordering corrected (Task 033)

## [0.2.0] - 2025-12-27

### Added
- Framework embedding at installation (Task 015)
  - Framework files now embedded in agents for self-contained execution
  - Removed runtime framework bundling

### Changed
- Documentation architecture refined (Task 028)
  - Separated agent-facing instructions from human-readable rationale
  - Framework files contain only execution instructions
  - Wiki contains context and design decisions

## [0.1.0] - 2025-12-20

### Added
- Prompt architecture and integration (Task 026)
  - Prompts as input records capturing user requirements
  - Free-style natural language with suggested structure
- User testing agent (Task 024)
  - Automated end-to-end testing capability
- Installer CLI implementation (Task 023)
  - `smaqit init`, `smaqit status`, `smaqit validate` commands
  - Cross-platform Go installer with embedded files

### Changed
- Infrastructure layer accepts cross-cutting input (Task 016)
  - Infrastructure specs can reference all Phase 1 specs for coherence

## [0.0.1] - 2025-12-18

### Added
- Complete specification templates (Tasks 017-021)
  - Business, Functional, Stack, Infrastructure, Coverage templates
- Agent templates (Tasks 002-003)
  - Specification agent template
  - Implementation agent template
- Framework documentation split (Task 013)
  - SMAQIT.md split into LAYERS.md, PHASES.md, TEMPLATES.md, AGENTS.md, ARTIFACTS.md
- Cross-platform build system (Task 012)
  - Makefile with `build`, `build-all`, `install`, `uninstall` targets
  - Support for Linux, macOS (Intel/ARM), Windows

### Changed
- Layer independence principle established (Task 007)
  - Each layer's prompt file is sole source of requirements
  - Upstream layers provide context, not requirements

[Unreleased]: https://github.com/ruifrvaz/smaqit/compare/v1.7.0...HEAD
[1.7.0]: https://github.com/ruifrvaz/smaqit/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/ruifrvaz/smaqit/compare/v1.5.1...v1.6.0
[1.5.1]: https://github.com/ruifrvaz/smaqit/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/ruifrvaz/smaqit/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/ruifrvaz/smaqit/compare/v1.3.1...v1.4.0
[1.3.1]: https://github.com/ruifrvaz/smaqit/compare/v1.2.0...v1.3.1
[1.2.0]: https://github.com/ruifrvaz/smaqit/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/ruifrvaz/smaqit/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/ruifrvaz/smaqit/compare/v0.9.1...v1.0.0
[0.9.0]: https://github.com/ruifrvaz/smaqit/compare/v0.8.2-beta...v0.9.0
[0.8.2-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.8.1-beta...v0.8.2-beta
[0.8.1-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.8.0-beta...v0.8.1-beta
[0.8.0-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.7.0-beta...v0.8.0-beta
[0.7.0-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.6.2-beta...v0.7.0-beta
[0.6.2-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.6.0-beta...v0.6.2-beta
[0.6.0-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.5.0-beta...v0.6.0-beta
[0.5.0-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.4.2-beta...v0.5.0-beta
[0.4.2-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.4.1-beta...v0.4.2-beta
[0.4.1-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.4.0-beta...v0.4.1-beta
[0.4.0-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.3.0...v0.4.0-beta
[0.3.2-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.2.0...v0.3.2-beta
