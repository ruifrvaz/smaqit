# Claude Code Migration

**Date:** 2026-07-16
**Session Focus:** Add Claude Code as a second supported platform alongside GitHub Copilot
**Tasks Referenced:** None (no task file created for this session)
**Tasks Completed:** N/A

---

## Actions Taken

### Review and Planning
- Reviewed the full product surface (9 Copilot agents, 23 skills, Go installer, framework docs) to scope the migration
- Confirmed scope via user decision: add Claude Code as a second target, keep GitHub Copilot working unchanged (not a replacement)
- Entered plan mode; produced and got approval for a migration plan covering agent translation, skill fixes, installer changes, docs, and hook parity

### Agent/Command Architecture (iterated 3 times based on feedback)
- First attempt duplicated full agent bodies into `agents-claude/` — rejected as pure duplication
- Second attempt used a single source with `{{PLACEHOLDER}}` tokens, but generated output was committed at repo root — rejected as generated files should never be committed
- Final architecture: `agents/<name>.md` (shared body, root, committed) + `.smaqit/definitions/agents/<name>.frontmatter.yaml` (per-platform frontmatter metadata + placeholder resolutions, committed) → compiled by `scripts/generate-agents.py` into `installer/agents-copilot/` and `installer/agents-claude/` (both gitignored, build artifacts only)
- Migrated all 9 agents to this pattern; verified compiled Copilot output is byte-identical (or semantically identical, differing only in YAML flow-style quoting) to the pre-migration hand-written files
- Fixed a pre-existing bug while migrating `qa`: frontmatter `name: qa` → `name: smaqit.qa`, matching the `smaqit.[name]` convention every other agent follows (confirmed compatible/desired by user)
- Added 4 Claude Code slash commands (`commands/smaqit.{development,deployment,validation,qa}.md`) that Task-delegate to the matching subagent — the 5 specification agents (business/functional/stack/infrastructure/coverage) intentionally get no command file, reproducing Copilot's `user-invocable: false` boundary since Claude Code has no equivalent frontmatter flag

### Skills Fixes
- Introduced `[SMAQIT_SKILLS_DIR]` placeholder replacing ~15 files' hardcoded `.github/skills/...` self-references
- Changed `copilot-instructions.md`-reading skills/scripts (vault-loader, cicd-generate, repo-config, deploy-rsync, provision-cyso) to check `CLAUDE.md` first, falling back to `.github/copilot-instructions.md`
- Reworded `@smaqit.X` Copilot chat-mention syntax to `/smaqit.X` in `smaqit.new-greenfield-project/SKILL.md` (works identically on both platforms)
- Extended `scripts/generate-agents.py` to compile skills the same way as agents: `skills/<name>/**` (root, shared source) → `installer/skills-copilot/` and `installer/skills-claude/`, with `[SMAQIT_SKILLS_DIR]` resolved at compile time — this replaced an earlier design that resolved the placeholder at install time via Go string-substitution in `installer/main.go`

### Installer Naming Symmetry
- User requested every compiled output directory carry an explicit platform suffix; renamed `installer/agents` → `installer/agents-copilot` for symmetry with `installer/agents-claude`, and added `installer/skills-copilot`/`installer/skills-claude` (previously a single reused `installer/skills`)
- Removed the now-obsolete runtime `[SMAQIT_SKILLS_DIR]` substitution logic from `copyEmbeddedDir` in `installer/main.go`

### Installer Wiring (`installer/main.go`, `installer/Makefile`)
- Added `go:embed` directives for `agents-copilot`, `agents-claude`, `commands-claude`, `skills-copilot`, `skills-claude`
- `cmdInit` now copies all five into `.claude/agents`, `.claude/commands`, `.claude/skills` alongside the existing `.github/agents`, `.github/skills`, `.github/workflows`
- `detectConflicts`, `cmdUninstall`, `cmdValidate`, `cmdHelp`, and the post-init message all updated to cover `.claude/*` consistently with `.github/*`
- `installer/Makefile`'s `prepare` target no longer copies `skills/` or `agents/` directly — `scripts/generate-agents.py` reads `agents/`, `commands/`, `skills/`, and `.smaqit/definitions/agents/` directly from the repo root and writes all five compiled trees itself

### Verification
- Rebuilt the installer from a clean state and ran a full scratch `smaqit init` multiple times across the redesign iterations, confirming `.claude/agents`, `.claude/commands`, `.claude/skills` (23 dirs) and `.github/agents`, `.github/skills` (23 dirs) all populate correctly
- Confirmed `smaqit status`, `smaqit plan`, `smaqit validate`, conflict detection, and `smaqit uninstall` all work correctly against the new `.claude/` tree
- Diffed `installer/skills-copilot/` vs `installer/skills-claude/`: confirmed only the 15 files using `[SMAQIT_SKILLS_DIR]` differ, everything else byte-identical

### Claude Code Hook Parity (repo dev tooling, not shipped)
- Added `.claude/settings.json` with `PreToolUse`/`PostToolUse` hooks and `.claude/hooks/scripts/test-inject.sh`, mirroring `.github/hooks/smaqit-test-hook.json`'s purpose (confirm hooks fire, dump git status) but using Claude Code's documented hook contract
- Confirmed live during the session — hook fired on every subsequent tool call
- Deliberately not added to the installer's embedded payload — this is a repo-local dev diagnostic, not a shipped product feature

### Documentation
- `README.md`: added Claude Code to the compatibility table; noted which agents are direct slash commands vs. Task-delegated subagents on Claude Code
- `framework/SKILLS.md`: documented the dual `.github/skills/` + `.claude/skills/` install locations and the `[SMAQIT_SKILLS_DIR]` placeholder
- `framework/TEMPLATES.md`: replaced the Copilot-only "Agent Definition Format" section with a description of the `agents/` + `.smaqit/definitions/agents/` → dual-compiled-format architecture
- `docs/wiki/agent-tools-reference.md`: added a Copilot→Claude Code tool-name mapping table for maintainers authoring/editing agents

### Commit
- Staged explicitly by path (not `git add -A`) after discovering `git add -A` would have swept in pre-existing untracked task-083 work (`smaqit.test-e2e-playwright/`, vault-loader `assets/`/`install-vault.sh`/`setup-vault.sh`) that predated this session
- Flagged this to the user; user chose to include it in the same commit since this session's `[SMAQIT_SKILLS_DIR]` fix had already touched some of those files
- Ran a secret-safety scan over the staged diff (clean — only matched removed lines referencing env-var indirection, no literal credentials)
- Committed as `11be9be feat: add Claude Code support alongside GitHub Copilot (dual-target)` (52 files, 1384 insertions, 361 deletions)

---

## Problems Solved

- **Duplication in first agent-migration attempt**: Full 250-line agent bodies were being copy-pasted per platform. Resolved by splitting into a shared body (with a handful of `{{PLACEHOLDER}}` tokens for platform-varying phrases) plus a small per-platform frontmatter YAML.
- **Generated files committed to repo root**: Second attempt wrote compiled output to `agents-claude/` at repo root and it got tracked. Resolved by generating exclusively into gitignored `installer/*` directories, matching the existing convention already used for `installer/skills`, `installer/framework`, etc.
- **Source of truth in the wrong place**: Third attempt put both body and frontmatter under `.smaqit/definitions/agents/`, leaving the repo-root `agents/` folder empty. User expected `agents/` to remain the visible source of truth. Resolved by moving body files back to `agents/`, keeping only per-platform metadata (frontmatter YAML) under `.smaqit/definitions/agents/`.
- **Skills placeholder resolved at the wrong stage**: `[SMAQIT_SKILLS_DIR]` was originally resolved at install time via Go string-replacement, inconsistent with how agents resolve their placeholders at compile time. Resolved by moving skill compilation into `scripts/generate-agents.py`, producing pre-resolved `installer/skills-copilot/` and `installer/skills-claude/` trees.
- **`qa` agent's Copilot tool list mismatch**: Q&A agent's `web/fetch`-only design meant it needed `Read` added on the Claude Code side (to read local `framework/*.md`/`docs/wiki/**` files) even though the original Copilot tool list had no explicit read tool — a necessary correction, not scope creep.
- **`git add -A` risk**: Would have bundled pre-existing unrelated uncommitted work into this session's commit with a misleading attribution. Caught by reviewing `git status` before staging and asking the user how to proceed, per the repo's own safety conventions.

---

## Decisions Made

- **Dual-target, not a replacement**: GitHub Copilot support is fully preserved; Claude Code is additive. Confirmed via explicit user choice early in the session.
- **Single source of truth per artifact type, split by concern**: shared content (`agents/`, `commands/`, `skills/`) lives at repo root; per-platform-only metadata lives in `.smaqit/definitions/`; all compiled/resolved output lives only inside gitignored `installer/*` directories. No content is ever duplicated by hand across platforms.
- **No Claude Code equivalent for `user-invocable: false`**: reproduced structurally by omitting a `commands/<name>.md` for agents that shouldn't be user-invocable, rather than adding a new frontmatter convention.
- **`copilot-setup-steps.yml` and the `test-e2e` Makefile target stay Copilot-only, permanently**: both are scoped specifically to GitHub Actions / Copilot's cloud coding agent and Copilot-SDK-based installer testing — not part of the product surface that ships via `smaqit init`, and GitHub Actions deployments will continue using Copilot. Explicitly out of scope, not a gap to fill later.
- **Installer output directories are fully symmetric by platform suffix**: `agents-copilot`/`agents-claude`, `skills-copilot`/`skills-claude`, `commands-claude` (no `commands-copilot` — not needed, Copilot invokes agents by their own name).

---

## Files Modified

### New (agents/commands architecture)
- `agents/{business,functional,stack,infrastructure,coverage,development,deployment,validation,qa}.md` (renamed from `agents/smaqit.*.agent.md`, content adjusted for shared-body placeholders)
- `.smaqit/definitions/agents/*.frontmatter.yaml` (9 files)
- `commands/smaqit.{development,deployment,validation,qa}.md`
- `scripts/generate-agents.py`

### New (Claude Code hook parity)
- `.claude/settings.json`
- `.claude/hooks/scripts/test-inject.sh`

### Modified (installer)
- `installer/main.go` — new embeds, dual `.claude/`+`.github/` copy/uninstall/validate logic, removed runtime skills-path substitution
- `installer/Makefile` — `prepare` target simplified to rely on `scripts/generate-agents.py`
- `.gitignore` — updated for renamed/added `installer/*-copilot`, `installer/*-claude` directories

### Modified (docs)
- `README.md`, `framework/SKILLS.md`, `framework/TEMPLATES.md`, `docs/wiki/agent-tools-reference.md`

### Modified (skills — `[SMAQIT_SKILLS_DIR]` / `CLAUDE.md` fallback / `/` invocation)
- `skills/smaqit.infrastructure-{cicd-generate,deploy-rsync,hook-pre-commit-validate,provision-cyso,repo-config,vault-loader}/SKILL.md` + associated scripts
- `skills/smaqit.new-greenfield-project/SKILL.md`

### New (pre-existing task-083 work, included per user decision)
- `skills/smaqit.test-e2e-playwright/**` (SKILL.md, scripts, templates)
- `skills/smaqit.infrastructure-vault-loader/{assets/vault.hcl.template,scripts/install-vault.sh,scripts/rotate-credential.sh,scripts/setup-vault.sh}`

---

## Next Steps

- No outstanding work from this session — `copilot-setup-steps.yml` and `test-e2e` were both explicitly ruled out of scope by the user, not deferred
- If Claude Code test coverage is ever wanted (parallel to the Copilot-SDK-based `test-e2e` Makefile target), that would be new scope, not a gap in this migration
- Consider whether `.claude/` should be added to the repo's own `.gitignore`-adjacent documentation (e.g. `.github/copilot-instructions.md`) the way `installer/` subdirectories already are, for consistency — not yet done

---

## Session Metrics

- **Date:** 2026-07-16
- **Platforms added:** 1 (Claude Code, alongside existing GitHub Copilot)
- **Agents migrated:** 9
- **Slash commands added:** 4
- **Skills fixed for dual-target:** 6 (content changes) + 23 (compiled into two resolved trees)
- **Architecture iterations before landing:** 3 (duplication → committed-generated-output → definitions-only → final root+metadata split)
- **Commits:** 1 (`11be9be`)
- **Files changed:** 52 (1384 insertions, 361 deletions)
