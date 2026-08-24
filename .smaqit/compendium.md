# Project Compendium

## Installation

**Where does smaqit install agents and skills, and what does `smaqit init` create?**

The shell installer and `smaqit update` install the shared framework payload once per user: Copilot agents go to `~/.copilot/agents/` (or `$COPILOT_HOME/agents/`), Claude Code agents, commands, and skills go under `~/.claude/` (or `$CLAUDE_CONFIG_DIR/`), Codex agents go to `~/.codex/agents/` (or `$CODEX_HOME/agents/`), and Copilot/Codex share skills in `~/.agents/skills/`. The hidden bootstrap used by those entry points is not a public installation command.

`smaqit init` is project scaffolding only. It creates and maintains project state such as `.smaqit/`, specifications, design artifacts, MCP configuration, instruction-file integration, and the create-if-absent Copilot setup workflow; it does not create `.github/agents/`, `.github/skills/`, `.claude/`, `.codex/agents/`, or `.agents/skills/` mirrors in the repository, and it performs no legacy-mirror migration — global installation is the only agent/skill distribution path, and an outdated project-level installation is replaced by reinstalling cleanly.

---

**How do I confirm a fresh `curl install.sh | bash` run matches a given release?**

Check the platform agent directories directly rather than trusting the script's own log output: `~/.copilot/agents/`, `~/.claude/{agents,commands,skills}/`, `~/.codex/agents/`, and the shared `~/.agents/skills/` should all show the 9 canonical agents and the current shipped skill count with today's mtime. The shared skill and agent directories are also used by the separate `smaqit-extensions` tool, so they legitimately contain more entries than smaqit's own `skills/`/`agents/` source trees (e.g. `smaqit.session-*`, `smaqit.task-*`, `smaqit.release-*`, `smaqit.project-*`, `smaqit.utils.*` skills, and extra Codex/Claude release/session/testing agents) — diff the installed directory listing against this repo's own `skills/`/`agents/` directories rather than assuming every installed entry originates from smaqit. Files with an older mtime than the install run are smaqit-extensions content correctly left untouched, not evidence of a partial install.

---

**How do I clean up installer build artifacts?**

There is no `make clean` target. Run `make uninstall` from `installer/`; it removes the dev binary and then prompts to also clean the build artifacts (`dist/`, `framework/`, `templates/`, `tools/`, `agents-*` platform trees, `commands-claude/`, `skills-shared/`, `skills-claude/`, `test/`, `.test-venv/`). Answer `y` at the prompt to remove them.

---

**Does the installer smoke test install into the real home directory?**

No. `scripts/smoke-test-installer.sh` redirects `HOME`, `COPILOT_HOME`, `CLAUDE_CONFIG_DIR`, and `CODEX_HOME` into a `mktemp -d` tree before invoking the binary, so every global install destination resolves inside the temporary directory. The tree is deleted by an `EXIT` trap; set `KEEP_SMOKE_DIR=1` to preserve it for debugging.

---

**What does `make smoke-test` check beyond `go test`?**

`make smoke-test` (target in `installer/Makefile`) builds the real `smaqit-dev` binary and runs `scripts/smoke-test-installer.sh` end-to-end against a temporary project: full global install, `smaqit init`, and a real `design render`/`design attest`/`design validate` cycle against the script's own inline PlantUML fixture. This exercises the actual compiled binary and its embedded templates, which `go test`'s in-process unit tests (built from their own separate fixture strings in `installer/*_test.go`) do not — a structural rule enforced only in Go validation code (e.g. a new required PlantUML directive) can pass every `go test` case while still failing `make smoke-test` if the script's own fixture wasn't updated to match. `make test` runs `go test`+`go vet` only; `smoke-test` is a separate target and must be run explicitly to catch this class of gap. See also: does the installer smoke test install into the real home directory?

---

## Self-Update

**Why does smaqit install thousands of files under `.smaqit/tools/`, and should they be committed?**

They are smaqit-managed PlantUML MCP/rendering dependencies, vendored per project so visual-design rendering is reproducible and does not require the consumer to install npm packages separately. They are generated installation content, not project source: `smaqit init` and `smaqit update` ensure the root `.gitignore` contains the single narrow rule `.smaqit/tools/`, while preserving user rules and tracked files. Keep canonical PlantUML sources and rendered PNG artifacts under `docs/designs/` in version control; the ignore rule must not cover that directory.

The current runtime is materialized only in the checkout where `init` or `update` ran. A new Git worktree therefore lacks the ignored runtime and current design/MCP commands fail until it is bootstrapped; this is a tracked upstream reliability gap, not a reason to commit or manually copy the runtime.

---

## Architecture

**Why can `smaqit-plantuml` be absent from ToolSearch even when its project MCP registrations exist?**

`smaqit init` and `smaqit update` register the stdio server in `.vscode/mcp.json` (VS Code), root `.mcp.json` (Claude Code), and trusted `.codex/config.toml` (Codex). Registration is not host activation: load the project, trust and start the server, then refresh the agent session. A valid Codex registration can still be ignored by known host defects, so missing tools remain `DESIGN-TOOLCHAIN-UNAVAILABLE`; inspect the host MCP-server list and restart or refresh rather than treating configuration presence as proof that tools are reachable.

---

**Why must a design's `status` match its linked specification status?**

A design is a lifecycle-coupled sidecar, not an independent deliverable. For active links, `smaqit design validate` requires its rank to equal the least-advanced linked specification rank; otherwise it reports `DESIGN-ARTIFACT-STALE`. Update status through the synchronized specification-status workflow so the specification and its canonical design remain at the same lifecycle point.

---

**Why does `smaqit design validate` show only one failure?**

The current validator is fail-fast: it stops at the first invalid design, stale reference, or active specification without a valid pair. That protects prerequisite ordering but makes large migrations slow to iterate. Aggregate independent artifact failures into one deterministic report, while retaining fail-fast behavior only when a missing prerequisite such as the PlantUML runtime prevents all further checks.

---

**How does `smaqit update` refresh newly shipped content?**

The self-update flow re-execs the newly downloaded binary after replacing the file on disk, because Go's `//go:embed` content is fixed in the running process. The fresh binary refreshes the global agent/skill payload through its internal bootstrap and re-scaffolds project-local assets only when the current directory is already a smaqit project. Updaters older than v3.0.0 predate global-payload bootstrapping: after such an updater first replaces itself with v3+, run `smaqit update` once more or rerun the shell installer to install the global payload.

---

**How must a release retire a previously shipped skill from existing projects?**

Deleting canonical source and regenerating installer staging removes a skill from new binaries, but it does not remove copies already installed in consumer projects: `cmdInit` overlays the new embedded tree without pruning paths absent from the new release. A skill retirement therefore needs a persistent installer tombstone listing the exact formerly-owned files. After the init conflict/approval gate, cleanup removes only those files from the Copilot, Claude, and Codex skill directories and prunes directories only when empty; uninstall applies the same legacy cleanup where normal embedded-file enumeration can no longer see the retired package. User-added files inside or beside the retired package must survive.

---

**What does the `system-sequence` design profile require?**

A Functional `system-sequence` design is a strict black-box contract: it declares exactly one actor and exactly one `participant "System" as System`; both the visible label and alias are case-insensitive matches for `System`. It must include `hide footbox` and may not contain a PlantUML `footer` directive. Every parsed message endpoint must be either the declared actor or `System`, so PlantUML cannot infer a second participant from an undeclared endpoint. The validator rejects extra actors, participant-family declarations, wrong System labels or aliases, missing footbox suppression, footers, and undeclared endpoints deterministically.

Actor names remain unconstrained. When a specification has multiple actor flows, author one linked system-sequence design per actor or flow. See also: what's the difference between a `system-sequence` and a `design-sequence` design?

---

**Why must a design's PlantUML block include a `title` directive?**

Every design diagram's PlantUML block must open with a single-line `title` directive whose value exactly matches the design's own `id` frontmatter field (e.g. `title DSG-BUS-LOGIN-USE-CASE`), enforced across all diagram types by `installer/design.go`'s `validateDesignMetadata` — a missing or mismatched title fails `DESIGN-VISUAL-INVALID`. The title deliberately uses the design's own `id` rather than the Business spec's `UC[N]-[CONCEPT]` heading label (purely cosmetic — never wired into frontmatter, requirement IDs, or cross-layer references anywhere else in the framework) or a linked specification's identifier (spec-to-design cardinality isn't guaranteed 1:1 — one design may legitimately serve several related specs). The `id` is already unique, stable, and threaded through the design's own frontmatter, filename, and requirement IDs, so it needs no new cross-referencing logic and has a single unambiguous value regardless of cardinality.

---

**Are actor names constrained in `system-sequence` designs?**

No. The actor may use any name, quoted label, or alias — `Customer`, `Employee`, `Visitor`, `actor "Front Desk Clerk" as Clerk`, etc. The validator requires exactly one actor declaration but never constrains its name. The sole system participant is the only fixed identity: its visible label and alias must each be `System`, case-insensitively.

---

**What's the difference between a `system-sequence` and a `design-sequence` design?**

A `system-sequence` design (Functional layer, `docs/designs/functional/`) is a pre-implementation, spec-authored black-box System Sequence Diagram: exactly one actor and one opaque `System` participant, external contract only — no internal collaborators. A `design-sequence` design (`docs/designs/design-sequence/`, layer prefix `DSD`) is its post-implementation counterpart: generated by `smaqit.development` after code and tests pass, showing the real internal collaborators (handlers, services, etc.) that fulfill that contract. Each `design-sequence` design links 1:1 to its paired `system-sequence` design via a `realizes:` frontmatter field.

Two deterministic checks run inside `smaqit design attest` before a `design-sequence` design can pass: grounding (every message must carry a `' impl: <path>:<line>` PlantUML comment citing real code, checked to exist) and completeness (every operation the paired `system-sequence` design promises must be represented). Both are source-level heuristic scans, not semantic verification of correctness.

Storage and lifecycle are deliberately not shared with Design Artifacts: `docs/designs/design-sequence/` is its own sibling tree, not nested under `docs/designs/functional/`, so it stays outside the Bounded Agents ownership boundary that makes `docs/designs/<layer>/` producer-owned by that layer's specification agent, and outside the Design Artifact lifecycle rule that resets a linked spec to `draft` on every semantic edit. A `design-sequence` design is a snapshot generated at Phase 1 completion, not a live view — it is not automatically regenerated when code changes outside a fresh Development pass, and editing it never resets its linked Functional spec's status, since it documents implementation output rather than gating implementation input. See also: what does the `system-sequence` design profile require?

---

**How does `smaqit plan`'s phase design-readiness gate decide which specs need a design pair?**

For `--phase=develop|deploy|validate`, `getPhaseDesignGateSpecs` (`installer/spec.go`) scopes its check to only the specs currently in the incremental cycle — those with `status: draft` or `status: failed` — using the same predicate `filterSpecsByStatus` already applies for pending-work accounting. A spec already `implemented`/`deployed`/`validated` is never re-checked: it already passed this same gate once, as a precondition of leaving `draft` in the first place (the phase workflow runs the gate automatically right after spec generation), so it cannot reach a post-draft status without an already-valid design pair. Editing an already-passed spec reverts its status to `draft`, which naturally brings it back into the gate's scope — nothing currently being touched can slip through. `validatePhaseDesignReadiness` reports every currently-blocking spec at once (aggregate), not just the first.

This scoping means the gate can never block on legacy, unrelated specs in a project that has used the design-pair convention from the start; it matters specifically for a project retrofitting the convention onto specs written before it existed, since those pre-existing specs sit at a post-draft status with no design pair and would otherwise be re-litigated on every unrelated feature's plan run. One case this scoping does not resolve: a project convention that deliberately never gives one layer (e.g. Coverage) a design pair — a draft spec in that layer is still in-cycle and still fails the gate, since that is a distinct per-layer policy question, not a scoping question. See also: why does `smaqit design validate` show only one failure? (a different, sibling command with its own separate fail-fast behavior).

---

**What is smaqit-adk, and how does it relate to this repo?**

smaqit-adk is a separate, generic Agent Development Kit repo that this project was built with: it owns the L0/L1/L2 principle-curation and compilation agents (`smaqit.L0`, `smaqit.L1`, `smaqit.L2`), installed globally, plus the `smaqit.create-agent`/`smaqit.create-skill`/`smaqit.new-principle` skills. This repo's own `agents/`, `commands/`, `skills/`, and `framework/*.md` are product-domain content only — smaqit's five-layer specification system (Layers, Phases, spec/design agents) — compiled independently via `scripts/generate-agents.py`, with no runtime dependency on smaqit-adk's L0/L1/L2 chain. `framework/*.md` in this repo is not generic ADK principle content; it documents this product's own domain concepts and is verified independent of smaqit-adk's own `framework/*.md`. See also: does changing a principle in `framework/*.md` automatically update this repo's agents or skills?

---

**Does changing a principle in framework/*.md automatically update this repo's agents or skills?**

No. There is currently no compiler that reads `framework/*.md` and regenerates `agents/*.md` or `skills/*/SKILL.md` from it — every agent and skill body in this repo is hand-authored. `scripts/generate-agents.py` only renders already-written `agents/*.md` + `commands/*.md` + `skills/*/SKILL.md` + `.smaqit/definitions/agents/*.frontmatter.yaml` into per-platform installer output; it never reads `framework/*.md`. A framework principle change must be manually reflected in whichever agent/skill bodies it affects. See also: what is smaqit-adk, and how does it relate to this repo?

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

It is the gitignored compiled-output directory holding Claude Code slash-command files. `scripts/generate-agents.py` generates it from root `commands/*.md`, and the global installer copies it to `~/.claude/commands/` (or `$CLAUDE_CONFIG_DIR/commands/`). Only development, deployment, validation, and QA receive direct commands; specification agents remain delegation-only.

---

**Does `smaqit init` install `copilot-instructions.md`, `CLAUDE.md`, or `AGENTS.md`?**

It installs `AGENTS.md` plus a thin `CLAUDE.md` that imports it. Copilot and Codex read `AGENTS.md`; Claude Code reads `CLAUDE.md`. Existing files are never overwritten: the marked smaqit section is appended when absent, and repeated initialization is idempotent. See `installInstructionsFile()` in `installer/main.go`.

---

**Why don't Claude Code slash commands show up when typing `/` in the VS Code extension chat?**

In the smaqit source repository, product commands exist only as generated installer staging and are not installed into the repository itself. After global installation, reload the VS Code window so it discovers `~/.claude/commands/` (or `$CLAUDE_CONFIG_DIR/commands/`). If the Linux or WSL extension still misses them, use the Claude CLI in an integrated terminal; the CLI surface can discover commands even when the extension panel does not.

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

`scripts/generate-agents.py` compiles canonical agent bodies and platform metadata into `installer/agents-codex/*.toml`, and renders the shared product skills into `installer/skills-shared/` for the global `~/.agents/skills/` root. The shell installer and updater install Codex agents to `~/.codex/agents/` (or `$CODEX_HOME/agents/`) and shared Copilot/Codex skills to `~/.agents/skills/`; `smaqit init` does not create project agent or skill mirrors. It still registers the exact `smaqit-plantuml` server table in the initialized project’s trusted `.codex/config.toml`, together with the other project-local MCP configuration. Uninstall removes only exact SmaQit-owned global entries while preserving unrelated or nested custom content.

---

**What must be updated when adding a new canonical skill?**

The total shipped skill count is asserted in `installer/main_test.go` in two tests that must be bumped together: `TestRemoveEmbeddedSkillDirsPreservesUnownedSharedContent` (a hardcoded `removed != N` check against the `skills-shared` embed) and `TestSharedSkillsServeCopilotAndCodex` (a hardcoded top-level skill-directory count, plus placeholder/path invariant checks). `scripts/smoke-test-installer.sh` asserts one representative skill file's presence rather than a count. Missing either Go count fails `go test` with a count mismatch, not a useful error about which skill changed.

---

## Release Workflow

**Why can a release worktree remain after its release PR has merged and published?**

Publishing a release does not automatically remove the local `release/vX.Y.Z` worktree or branch: Git keeps it registered until an explicit cleanup removes the worktree, deletes the merged local branch, and refreshes the project workspace. Confirm the PR, tag, and release workflow first; then clean up the local release workspace. This separates publication verification from a recoverable local cleanup step.

---

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

## Task Management

**Where does the task-lifecycle tooling (`task-start`, `task-create`, `task-complete`, `task-list`, `utils.worktree`) live, and does `smaqit init` install it?**

It is owned entirely by the sibling `smaqit-extensions` repository, not by `smaqit`'s own `skills/` directory or its shipped product skills. `smaqit init` does not install it — a consumer project only gets task-lifecycle skills (branch/worktree creation, parent-child task ownership, assisted/autonomous mode) if `smaqit-extensions` is installed separately. Shipped smaqit skills that reference the lifecycle (`smaqit.new-greenfield-project`, `smaqit.feature-new`) call `smaqit.task-start`/`smaqit.task-create`/`smaqit.task-complete` by name without re-implementing any of their branch, worktree, or merge mechanics — those skills assume the lifecycle tooling is present, they do not provide a fallback if it isn't. In this repo, the copies under `.github/skills/`, `.claude/skills/`, and `.agents/skills/` for task-lifecycle skill names are committed dogfooding install output from `smaqit-extensions`, not canonical smaqit source.

`smaqit-extensions` also ships a parent-owned subtask lifecycle (`Parent: NNN` task metadata): a child task joins its parent's existing branch/worktree and inherits its mode instead of creating its own; only the parent merges and cleans up, and only once every declared child is `Completed`. `smaqit.feature-new` adopts this contract for its own five-phase structure (one shared feature-cycle parent, five phase children) rather than spawning a separate branch per phase.

---

**How do you determine whether an old task file in `.smaqit/tasks/` is still relevant before starting or completing it?**

Task files are written speculatively and can go stale as the framework evolves out from under them — re-reading the task text alone is not sufficient. Check three things against the *current* codebase rather than trusting the file: (1) do the file paths, commands, or skill names it references still exist (e.g. a task written when prompts were still `.github/prompts/*.prompt.md`, or when `PLANNING.md` lived under `docs/tasks/` instead of `.smaqit/tasks/`, is describing a mechanism that may no longer exist); (2) has a later task already superseded or completed the same underlying work, sometimes without ever being run through the task's own file (check `git log` on the files the task claims to touch, not just `PLANNING.md`'s status column); (3) has the concern the task wants to validate already been implicitly exercised by extensive real usage since it was filed, with no issue ever surfacing. A task can fail all three checks and still describe a real, unaddressed gap — in that case the fix is usually much smaller than the original task once rescoped to current architecture, and may not warrant a tracked task on its own.

---
