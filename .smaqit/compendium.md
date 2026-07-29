# Project Compendium

## Self-Update

**Why does `smaqit update` silently skip new skills/scripts after downloading a new release?**

The self-update flow must re-exec the newly downloaded binary after replacing the file on disk. Go's `//go:embed` content is fixed in the running process, so an in-process reinitialization would still use the old agents, skills, and templates. The no-replacement paths (already up to date, or local newer than remote) can safely reinitialize in-process. See `installer/update.go`'s `reinitWithBinary()`.

---

**How must a release retire a previously shipped skill from existing projects?**

Deleting canonical source and regenerating installer staging removes a skill from new binaries, but it does not remove copies already installed in consumer projects: `cmdInit` overlays the new embedded tree without pruning paths absent from the new release. A skill retirement therefore needs a persistent installer tombstone listing the exact formerly-owned files. After the init conflict/approval gate, cleanup removes only those files from the Copilot, Claude, and Codex skill directories and prunes directories only when empty; uninstall applies the same legacy cleanup where normal embedded-file enumeration can no longer see the retired package. User-added files inside or beside the retired package must survive.

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

**What is the ownership model for an end-to-end post-MVP feature workflow?**

`smaqit.feature-new` is the single top-level workflow for specification revalidation, development, deployment, validation, and release. Phase 1 owns one incremental specification pass and records a durable exact-path handoff. Development, Deployment, and Validation consume that handoff in an explicit prevalidated-spec mode, skipping repeated specification generation while retaining consolidation and `smaqit plan` processing; direct phase-agent calls keep orchestration-first behavior.

Deployment remains part of the feature cycle. A feature PR is the human gate: an unmerged PR is active work awaiting approval, not a separate deferred deployment state. The Deployment agent owns the contiguous existing-CI/CD operation—artifact preparation, PR creation, merge pause, exact workflow monitoring, deployed-revision verification, spec state, and report—while Feature New owns phase task state, preflight decisions, and evidence validation.

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

**What Vault namespace convention does smaqit use for machine credentials vs. app credentials?**

Credentials split across two namespaces by what they belong to. `secret/machines/<machine-slug>/*` holds everything scoped to a provisioned VM regardless of which app runs on it: `base-ssh` (the bootstrap-only credential Terraform installs at provision time — never used for routine deploys), `cyso` and `tfstate` (cloud-provisioning credentials, since provisioning is a property of the machine, not any one app), and `metadata` (non-secret host/provider/owner-project info). `secret/apps/<app-slug>/*` holds everything scoped to an individual app: `ssh` (a distinct keypair per app, no exceptions — even the project that originally provisioned the machine gets its own, bootstrapped rather than reused), `github`, and `machine` (a pointer recording which machine-slug this app is bootstrapped against, used by `rotate-credential.sh` to know where to re-authorize a rotated key). An app's `ssh` credential is always populated via `smaqit.infrastructure-vault-loader/scripts/bootstrap-app-to-machine.sh <app-slug> <machine-slug>` — idempotent, never by copying another app's key material or letting Terraform install it directly. Projects predating this convention still use the older flat `secret/<project-slug>/{cyso,ssh,tfstate,github}` scheme with no machine-level namespace at all; `load-credentials.sh` supports both, auto-detecting which applies per invocation, and migrating an existing flat-scheme project onto `apps/`+`machines/` is a manual, project-by-project decision rather than something any smaqit skill does automatically.

---

**How does `load-credentials.sh` decide whether to prompt for installing an SSH public key manually, versus assuming it gets installed automatically?**

The decision is per `provisioning_mode`, independent of which credential scheme (legacy flat vs. `apps/`+`machines/`) is in play. `provision`/`existing-owned` generate the deploy keypair silently and store it in Vault, because a Terraform `apply` is expected to push the public key onto the VM via cloud-init — no manual step is needed. `existing-shared` and `existing-unmanaged` both skip Terraform for this project entirely, so neither has any automated path to install the key onto the VM; both print explicit "append this public key to `~/.ssh/authorized_keys` yourself" instructions after generating it. The two differ only in *how* the key gets there: `existing-shared` additionally offers copying the already-trusted keypair from the project whose Terraform state owns the VM (a real option, since that project already has access); `existing-unmanaged` has no such option, since by definition no project's Terraform manages the VM — every credential load for that mode generates a fresh keypair and always needs the manual-install step. `existing-shared` and `existing-unmanaged` also both skip prompting for `cyso`/`tfstate` credentials in the legacy flat-scheme path, for the same underlying reason (no Terraform run for this project against this VM either way).

---

## Codex Support

**How does smaqit provide first-class Codex compatibility?**

`scripts/generate-agents.py` compiles the 9 canonical agent bodies and platform metadata into `installer/agents-codex/*.toml`, and copies all 25 canonical product skills into `installer/skills-codex/` with `.agents/skills` path substitution. `smaqit init` installs agents to `.codex/agents/` and skills to `.agents/skills/` without creating or modifying `.codex/config.toml`. Validation checks both directories, update reinitialization uses the fresh binary's embedded content, and uninstall removes exact embedded files while preserving unrelated or nested custom content.

---

**What must be updated when adding a new canonical skill?**

The total shipped skill count is asserted independently in two places, and both must be bumped together: `installer/main_test.go`'s `TestRemoveEmbeddedSkillDirsPreservesUnownedCodexContent` (a hardcoded `removed != N` check), and `scripts/smoke-test-installer.sh`'s Python verification block (a hardcoded `len(expected_skills) != N` check). Missing either one fails CI (`go test` or `make smoke-test` respectively) with a count mismatch, not a useful error about which skill changed.

---

## Release Workflow

**How can a local smaqit release authenticate an encrypted SSH key from WSL2 or another interactive desktop Linux session?**

Preferred path: `smaqit.release-git-local`'s "Desktop Linux SSH Agent Recovery" procedure discovers an already-running desktop SSH agent socket (GNOME Keyring's `gcr-ssh-agent`, an alternate keyring path, GnuPG's agent socket, the current `$SSH_AUTH_SOCK`, or the systemd user session's) by testing each candidate with `ssh-add -l`, then scopes `SSH_AUTH_SOCK` to a single retry of the exact failed git command — no persistent config changes, no new agent started, no identity loaded or removed. This works whenever the user's desktop keyring is already unlocked (commonly true for the whole desktop session once logged in, e.g. GNOME's Login keyring auto-unlocking with the account password) — no popup or passphrase prompt is needed in that case, since the key material is already available.

Fallback when no usable socket is found: load the key into a temporary `ssh-agent` and use `SSH_ASKPASS_REQUIRE=force` with a WSLg password dialog so the passphrase never passes through chat, command arguments, or logs. This requires a GUI askpass binary (`zenity`, `ssh-askpass-gnome`, or similar) already installed — none ships by default, and installing one via `sudo apt` is blocked by the Claude Code auto-mode permission classifier.

In both cases: confirm authentication with `ssh -T git@github.com`, push `main` and the annotated tag separately, verify both remote refs. The `gh` CLI's HTTPS token is not a usable fallback for either path: it authenticates read operations but returns `403` (no write scope) on push. Note that the Claude Code auto-mode permission classifier may independently deny a push (particularly a tag push) or even a read-only `ls-remote`, with no further reasoning than a generic denial — this is unrelated to SSH/credential state and requires either a Bash permission rule or the user pushing that specific step from their own shell. If no agent socket is usable and no askpass tool is available, the default expectation is a local commit + annotated tag + verified build, with the actual push left to the user's own shell.

---

**Why does pushing to `main` not trigger a new release build?**

`post-merge-release.yml` triggers only on a `v*` tag push or a merged pull request to `main` — a plain branch push never fires it, regardless of what changed. Push the annotated release tag separately (`git push origin vX.Y.Z`) to trigger it. A newly triggered run can take a couple of minutes to appear in `gh run list` or the Actions UI; check via `gh api repos/<owner>/<repo>/actions/runs` (unfiltered by status) rather than assuming a run that isn't immediately visible never started.

---

## Skill Development

**Can a shipped skill reference files under `framework/`?**

No. The installer's `go:embed` manifest in `installer/main.go` only ships `agents-*`, `commands-claude`, `skills-*`, and the `AGENTS.md`/`CLAUDE.md` templates — `framework/` is never installed into a consumer project. It exists only in this canonical repo as agent-facing documentation for developing smaqit itself. A skill that depends on framework-documented mechanics (e.g. the Incremental Spec Updates decision table, spec state transitions) must distill the relevant content into its own `references/` file rather than pointing at `framework/*.md`, the same way `smaqit.new-greenfield-project` stays fully self-contained with zero `framework/` references.

---
