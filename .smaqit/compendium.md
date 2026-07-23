# Project Compendium

## Self-Update

**Why does `smaqit update` silently skip new skills/scripts after downloading a new release?**

The self-update flow must re-exec the newly downloaded binary after replacing the file on disk. Go's `//go:embed` content is fixed in the running process, so an in-process reinitialization would still use the old agents, skills, and templates. The no-replacement paths (already up to date, or local newer than remote) can safely reinitialize in-process. See `installer/update.go`'s `reinitWithBinary()`.

---

## Hooks

**Do VS Code Copilot hooks fire for `runSubagent` tool calls?**

No. `SubagentStart` and `SubagentStop` only fire for VS Code native agent delegation, such as `@agent` in chat input. They do not fire when a subagent is invoked via the `runSubagent` tool inside an agent. `PostToolUse` does fire for tool calls including terminal and file operations.

---

**What hook format does VS Code Copilot require?**

VS Code uses PascalCase event names such as `SubagentStart` and `PostToolUse`, the `command` key rather than `bash`, no `version` field, and a `hookSpecificOutput` wrapper object. Inline commands work reliably; external scripts also work once the hook pipeline is active.

---

## Agent Orchestration

**What agent orchestration pattern does smaqit's phase orchestration match?**

It is a sequential workflow: a phase agent orchestrates specification agents in a fixed order. Assisted mode, where the user reviews each specification, is a maker-checker workflow. Individual specification-agent invocations are nested composition. Deterministic routing is hardcoded rather than delegated to agents at runtime.

---

## Claude Code Support

**What's the Claude Code equivalent of `copilot-setup-steps.yml`?**

There is no direct equivalent because Claude Code has no magic-filename bootstrap convention. The closest local mechanism is a `SessionStart` hook in `.claude/settings.json`; cloud workflows must add their setup steps explicitly. `CLAUDE.md` is project context, not a bootstrap mechanism.

---

**What is `installer/commands-claude/`?**

It is the gitignored compiled-output directory holding Claude Code slash-command files. `scripts/generate-agents.py` generates it from root `commands/*.md`, and the installer copies it to `.claude/commands/` in target projects. Only development, deployment, validation, and QA receive direct commands; specification agents remain delegation-only.

---

**Does `smaqit init` install `copilot-instructions.md`, `CLAUDE.md`, or `AGENTS.md`?**

It installs `AGENTS.md` plus a thin `CLAUDE.md` that imports it. Copilot and Codex read `AGENTS.md`; Claude Code reads `CLAUDE.md`. Existing files are never overwritten: the marked smaqit section is appended when absent, and repeated initialization is idempotent. See `installInstructionsFile()` in `installer/main.go`.

---

**Why don't Claude Code slash commands show up when typing `/` in the VS Code extension chat?**

In the smaqit source repository, product commands exist only as generated installer staging and are not installed into the repository itself. In an initialized target project, reload the VS Code window first. If the Linux or WSL extension still misses `.claude/commands/`, use the Claude CLI in an integrated terminal; the CLI surface can discover commands even when the extension panel does not.

---

## Infrastructure Skills

**How does `smaqit.new-greenfield-project` pick a deploy skill when a project's stack matches none of the existing `smaqit.infrastructure-deploy-rsync*` family?**

Phase 4 reads the declared stack from the stack specification, compares it generically against installed deploy skills, and synthesizes a new skill when no match exists. The synthesized skill uses the existing deploy-rsync family as exemplars and preserves shared conventions such as the `__APP_DIR__` token, `write-vhost.sh`, deploy stamps, and guard-script reuse. A human checkpoint is required before first invocation unless prior autonomous approval exists.

---

**Why must every `terraform plan` or `apply` against already-provisioned infrastructure go through `plan-guard.sh`, even for diagnosis?**

The risk of a destructive delete or replace plan is the same whether the operator intends to apply it or only inspect it. Routing every invocation through `plan-guard.sh` ensures direct interactive commands receive the same protection as generated CI workflows.

---

**What causes SSH keys stored in Vault to fail with `error in libcrypto`?**

Shell command substitution strips the private key's required trailing newline. Store the key with Vault's `@file` syntax instead of `private_key="$(cat file)"`; file syntax preserves the exact bytes. See the Gotchas in `smaqit.infrastructure-vault-loader`.

---

## Codex Support

**How does smaqit provide first-class Codex compatibility?**

`scripts/generate-agents.py` compiles the 9 canonical agent bodies and platform metadata into `installer/agents-codex/*.toml`, and copies all 25 canonical product skills into `installer/skills-codex/` with `.agents/skills` path substitution. `smaqit init` installs agents to `.codex/agents/` and skills to `.agents/skills/` without creating or modifying `.codex/config.toml`. Validation checks both directories, update reinitialization uses the fresh binary's embedded content, and uninstall removes exact embedded files while preserving unrelated or nested custom content.

---

**What must be updated when adding a new canonical skill?**

The total shipped skill count is asserted independently in two places, and both must be bumped together: `installer/main_test.go`'s `TestRemoveEmbeddedSkillDirsPreservesUnownedCodexContent` (a hardcoded `removed != N` check), and `scripts/smoke-test-installer.sh`'s Python verification block (a hardcoded `len(expected_skills) != N` check). Missing either one fails CI (`go test` or `make smoke-test` respectively) with a count mismatch, not a useful error about which skill changed.

---

## Release Workflow

**How can a local smaqit release authenticate an encrypted SSH key from WSL2?**

Load the key into a temporary `ssh-agent` and use `SSH_ASKPASS_REQUIRE=force` with a WSLg password dialog so the passphrase never passes through chat, command arguments, or logs. Confirm authentication with `ssh -T git@github.com`, push `main` and the annotated tag separately, verify both remote refs, then terminate the temporary agent. This requires a GUI askpass binary (`zenity`, `ssh-askpass-gnome`, or similar) already installed — none ships by default, and installing one via `sudo apt` is blocked by the Claude Code auto-mode permission classifier. The `gh` CLI's HTTPS token is not a usable fallback: it authenticates read operations but returns `403` (no write scope) on push. When no askpass tool is available, the default expectation is a local commit + annotated tag + verified build, with the actual push left to the user's own shell (same filesystem, working SSH credentials the sandboxed session lacks).

---

**Why does pushing to `main` not trigger a new release build?**

`post-merge-release.yml` triggers only on a `v*` tag push or a merged pull request to `main` — a plain branch push never fires it, regardless of what changed. Push the annotated release tag separately (`git push origin vX.Y.Z`) to trigger it. A newly triggered run can take a couple of minutes to appear in `gh run list` or the Actions UI; check via `gh api repos/<owner>/<repo>/actions/runs` (unfiltered by status) rather than assuming a run that isn't immediately visible never started.

---

## Skill Development

**Can a shipped skill reference files under `framework/`?**

No. The installer's `go:embed` manifest in `installer/main.go` only ships `agents-*`, `commands-claude`, `skills-*`, and the `AGENTS.md`/`CLAUDE.md` templates — `framework/` is never installed into a consumer project. It exists only in this canonical repo as agent-facing documentation for developing smaqit itself. A skill that depends on framework-documented mechanics (e.g. the Incremental Spec Updates decision table, spec state transitions) must distill the relevant content into its own `references/` file rather than pointing at `framework/*.md`, the same way `smaqit.new-greenfield-project` stays fully self-contained with zero `framework/` references.

---
