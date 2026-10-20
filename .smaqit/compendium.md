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

**Why can `smaqit-plantuml` be absent from ToolSearch, or fail to connect, even when its project MCP registrations exist?**

`smaqit init` and `smaqit update` register the stdio server in `.vscode/mcp.json` (VS Code), root `.mcp.json` (Claude Code), and trusted `.codex/config.toml` (Codex). Registration is not host activation: load the project, trust and start the server, then refresh the agent session. A valid Codex registration can still be ignored by known host defects, so missing tools remain `DESIGN-TOOLCHAIN-UNAVAILABLE`; inspect the host MCP-server list and restart or refresh rather than treating configuration presence as proof that tools are reachable.

A `CONNECT_TIMEOUT` on startup has a separate, local cause: the design-tools runtime install lock. It is a directory, `.smaqit/tools/plantuml/.<bundle-version>.lock`, and it stays behind if a `smaqit mcp plantuml` process is killed while holding it, for example by an editor reload. A start that needs the lock waits up to 60 s and treats it as stale only after 5 minutes, while Claude Code stops waiting after 30 s. A start whose bundle is already installed and valid validates it without taking the lock, so an orphaned lock blocks only starts that must install: a fresh worktree, or the first start after a bundle-version change. If the server times out and no `smaqit mcp plantuml` process is running for that project, delete the lock directory and reconnect (`/mcp` in Claude Code).

---

**Why must a design's `status` match its linked specification status?**

A design is a lifecycle-coupled sidecar, not an independent deliverable. For active links, `smaqit design validate` requires its rank to equal the least-advanced linked specification rank; otherwise it reports `DESIGN-ARTIFACT-STALE`. Update status through the synchronized specification-status workflow so the specification and its canonical design remain at the same lifecycle point.

---

**Does `smaqit design validate` stop at the first failure, or report every problem at once?**

It reports independent failures together. `validateDesigns` (`installer/design.go`) checks every design under `docs/designs/` and collects one line per failing design (`<file>: <error>`). Without a path argument, it also adds every active specification that lacks a valid design pair. Everything comes back in a single `DESIGN-VALIDATION-FAILED: N issue(s):` report. Within one design, checks still stop at that design's first failure. The order is parse/metadata/profile, then PNG presence, hashes, visual attestation, and references. The PlantUML syntax check runs separately and can add a second line for the same design. Fixing a design's reported failure can therefore reveal the next one in that same design on a rerun. Only prerequisites stop the whole run immediately: no initialized project root, or a PlantUML runtime session that can't be opened (for example, Node 22+ is missing). Those prevent every further check. Passing a single design path validates only that design and skips the active-specification walk. Designs with `status: deprecated` skip the syntax and artifact checks. See also: how does `smaqit plan`'s phase design-readiness gate decide which specs need a design pair?

---

**How does `smaqit update` refresh newly shipped content?**

The self-update flow re-execs the newly downloaded binary after replacing the file on disk, because Go's `//go:embed` content is fixed in the running process. The fresh binary refreshes the global agent/skill payload through its internal bootstrap and re-scaffolds project-local assets only when the current directory is already a smaqit project. Updaters older than v3.0.0 predate global-payload bootstrapping: after such an updater first replaces itself with v3+, run `smaqit update` once more or rerun the shell installer to install the global payload.

Already being on the latest release doesn't make `smaqit update` a no-op. It still reinstalls the global payload, which also removes anything an older release installed and this one no longer ships, then re-initializes project assets if the current directory has `.smaqit/`.

That re-initialization asks before overwriting `.smaqit/templates/`. In a non-interactive shell the question gets no answer and init prints "Installation cancelled". The global step has already run by then, and only the project step is skipped. Answer it explicitly with `printf 'y\n' | smaqit update`, or run `echo y | smaqit init .` afterwards.

To refresh only the global agents and skills, without touching any project, run `smaqit update` from a directory that has no `.smaqit/`. See also: does `smaqit update` remove a skill that a release retires from an already-installed global directory?
---

**Does `smaqit update` remove a skill that a release retires from an already-installed global directory?**

Yes, since the release that renamed `smaqit.new-greenfield-project` to `smaqit.project-new` (task 141). Global install, which `smaqit update` and the shell installer both run, works like this:

- **It records what it installed.** It writes a manifest, `.smaqit-installed.json`, inside each global destination: `~/.agents/skills/`, and the Claude, Copilot and Codex agent, command and skill directories, honouring `CLAUDE_CONFIG_DIR`, `COPILOT_HOME` and `CODEX_HOME`.
- **It removes what was dropped.** On the next install, it deletes every file the previous manifest lists that the current release no longer ships, then removes any directory that leaves empty. A file the manifest never listed is never touched: user-added files inside a smaqit skill directory, and everything smaqit-extensions or smaqit-adk installs into the same directories.
- **It covers releases older than the manifest** with a seed list, `retiredSkills` in `installer/global_manifest.go`. A seeded directory is removed only when its `SKILL.md` still carries that exact name. The seed names only `smaqit.`-prefixed skills that a v3 release installed globally, and a test checks that none of them is still shipped.
- **It refuses an untrusted manifest.** An unreadable or tampered manifest (bad JSON, an unknown schema, or a path that leaves the destination) stops the install before anything is removed.
- **Uninstall cleans up too.** It also removes the manifest's leftovers and the manifest itself.
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

This scoping means the gate can never block on legacy, unrelated specs in a project that has used the design-pair convention from the start; it matters specifically for a project retrofitting the convention onto specs written before it existed, since those pre-existing specs sit at a post-draft status with no design pair and would otherwise be re-litigated on every unrelated feature's plan run. One case this scoping does not resolve: a project convention that deliberately never gives one layer (e.g. Coverage) a design pair — a draft spec in that layer is still in-cycle and still fails the gate, since that is a distinct per-layer policy question, not a scoping question. See also: does `smaqit design validate` stop at the first failure, or report every problem at once? (a different, sibling command with its own reporting behavior).

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

**How do smaqit agents ask the user questions when they run as subagents?**

Mostly, they don't ask; they return. Claude Code removes `AskUserQuestion` from every subagent, and a Claude subagent whose `tools:` list omits `Skill` cannot invoke skills at all. A spec agent running under a phase agent is two levels deep, so it can neither run its input skill on demand nor reach the user. smaqit therefore uses two mechanisms:

- **The skills are preloaded.** Each agent's `claude:` metadata lists its own `smaqit.input-*` skill under `skills:`, which injects the skill's content at startup. `scripts/generate-agents.py` fails `make prepare` when an agent body names an input skill its Claude metadata can't reach.
- **Signals are returned, not asked.** An agent that can't ask returns `NEEDS-INPUT`, listing every open question and stating that nothing was written. In Assisted mode, a phase agent returns `REVIEW-REQUESTED` after each specification layer. Orchestrating agents pass these signals up unchanged.

The main conversation owns the loop. That is the `/smaqit.<phase>` command body, or an orchestrating skill like `smaqit.feature-new`. It applies the phase input skill itself, passes the agent an explicit `PHASE-PARAMETERS` block, asks the user whatever a signal lists, and resumes the same subagent by sending its agent ID the `answers`. If resuming isn't possible, it invokes the agent again with `answers` plus `approved_spec_paths`. Approved specs count as current even while `draft`, so a fresh run never regenerates them.

On GitHub Copilot, all eight non-QA agents have `vscode/askQuestions` and can ask directly; a `runSubagent` child that can't ask returns the signal to its parent. Before task 131 this was broken on Claude Code: the input gates never ran, Assisted mode was unreachable on every platform because nothing set the mode, and project-new re-ran every specification agent after its own specification phase. See also: what agent orchestration pattern does smaqit's phase orchestration match? When are the `smaqit.input-*` skills invoked?

---

**When are the `smaqit.input-*` skills invoked?**

There are two families, and each runs at a different point:

- **The five layer gates** (`input-business`, `-functional`, `-stack`, `-infrastructure`, `-coverage`) run inside their specification agent, before it writes anything. A specification agent MUST apply its gate (`framework/AGENTS.md`, Session Context). The gate checks the layer's required sections (for Business: Actors, Use Cases, Success Metrics), flags contradictions, and checks whether to add new specs or revise existing ones. It asks the user directly where it can, and otherwise returns `NEEDS-INPUT` with stable question ids.
- **The three phase skills** (`input-development`, `-deployment`, `-validation`) resolve execution parameters for a phase: `mode` (assisted when the user started the phase or invoked the orchestrating skill running it, autonomous when the model started it), the user's request verbatim as `user_request`, the `specification_mode` the caller sets, the phase's own preferences and, for deployment, `provisioning_mode`, environment, deployment path and the slot plan from `smaqit slots plan`.
  - Every parameter has a default, so these skills ask only about blocking ambiguity.
  - They output a `PHASE-PARAMETERS` block.
  - The caller runs them in the main conversation and passes the block in: the `/smaqit.<phase>` command body, `smaqit.feature-new` Phase 3, or project-new Phases 4–6.
  - A phase agent applies its own preloaded copy only when it was invoked without a block.

On Claude Code every agent preloads exactly its own input skill through `skills:`. That's the only way it can reach the skill, since agent `tools:` lists omit `Skill`. Their "Invoke automatically when…" descriptions matter only in the main conversation, where Claude can choose a skill by its description. See also: how do smaqit agents ask the user questions when they run as subagents?

---

**What mode does a smaqit phase run in, and what counts as a requirement?**

It depends on what started the phase. A phase the user starts (a typed `/smaqit.<phase>`, or a Copilot/Codex phase agent they picked) defaults to `assisted`, and so does a phase agent invoked with no parameter block. A phase the model starts on its own defaults to `autonomous`. The orchestrating skills `smaqit.project-new` and `smaqit.feature-new` resolve one mode for every phase they run: `assisted` when the user invoked them (slash command, Codex `$` mention, or skill-menu pick), `autonomous` when the model chose them, including when the user only named them in plain words. A mode the user states always wins. The phase input skills resolve this; they never ask about it.

- **Assisted:** every specification layer works only from requirements the user stated and pauses with `REVIEW-REQUESTED`, and the phase ends by asking the user to review it before the next one.
- **Autonomous:** best effort end to end. Upstream specifications count as requirement sources, and the run stops only for `NEEDS-INPUT` and human operations such as a PR merge, a credential prompt, or the release. The orchestrators' review and sign-off gates are assisted-only.

In both modes the layer gates count only what the user stated: the request, passed verbatim as `user_request`, their `answers`, their own messages, and files they pointed at such as `.smaqit/requirements-extract.md`. A project or directory name, the model's own knowledge, and a caller's paraphrase never satisfy a required section, and no agent may fill a required section with an assumption; it returns `NEEDS-INPUT` instead. A blank project titled "mario-app" therefore stops at the Business gate rather than producing a Mario-themed app. An empty request blocks only when a layer actually has to be generated. Earlier releases defaulted every direct phase run to autonomous, and their gates accepted inferred requirements. See also: when are the `smaqit.input-*` skills invoked?

---

**What is the ownership model for an end-to-end post-MVP feature workflow?**

`smaqit.feature-new` is the single top-level workflow for specification revalidation, development, deployment, validation, and release. Phase 1 owns one incremental specification pass and records a durable exact-path handoff. Development, Deployment, and Validation consume that handoff in an explicit prevalidated-spec mode, skipping repeated specification generation while retaining consolidation and `smaqit plan` processing; direct phase-agent calls keep orchestration-first behavior.

Deployment remains part of the feature cycle. A feature PR is the human gate: an unmerged PR is active work awaiting approval, not a separate deferred deployment state. The Deployment agent owns the contiguous existing-CI/CD operation—artifact preparation, PR creation, merge pause, exact workflow monitoring, deployed-revision verification, spec state, and report—while Feature New owns phase task state, preflight decisions, and evidence validation. See also: what subagents and skills does each phase of `smaqit.feature-new` invoke?

---

**What subagents and skills does each phase of `smaqit.feature-new` invoke?**

Phase 0 (Task Creation) uses `smaqit.task-create` and `smaqit.task-start` to set up the shared feature-cycle parent and its five phase children. Phase 1 (Spec Revalidation) invokes the `smaqit.business`, `smaqit.functional`, `smaqit.stack`, `smaqit.infrastructure`, and `smaqit.coverage` specification agents as needed (by name; they have no slash command on Claude Code), plus `smaqit.design-validate` and `smaqit.spec-status-update` for status-only bumps. Phase 2 (Development) invokes `/smaqit.development` with `specification_mode: prevalidated`. Phase 3 (Deployment) invokes `smaqit.input-deployment` to resolve `provisioning_mode` and the deployment slot plan for the `existing-cicd-pr` path, then `/smaqit.deployment` with that plan; the agent runs the credential and repository-configuration slots (by default `smaqit.infrastructure-vault-loader` and `smaqit.infrastructure-repo-config`), opens and monitors the deploy PR's CI run, and runs the `verify` slot (by default `smaqit.infrastructure-deploy-verify`). Phase 3 never calls an infrastructure skill directly. Phase 4 (Validation) invokes `/smaqit.validation` with `specification_mode: prevalidated`. Phase 5 (Close-out) runs the release chain — `smaqit.release-analysis` → `smaqit.release-approval` → `smaqit.release-prepare-files` → `smaqit.release-git-pr` — then completes the parent task. Every phase brackets its work with `smaqit.task-start`/`smaqit.task-complete` for that phase's child task. See also: what is the ownership model for an end-to-end post-MVP feature workflow?

---

**Does the Deployment agent itself know about `provisioning_mode` values like `existing-k3s`, `existing-shared`, or `existing-unmanaged`?**

Only through the slot plan it executes. `agents/deployment.md` contains no per-mode branches: `provisioning_mode` selects the default filler for each deployment slot, `smaqit slots plan` resolves that into an ordered plan (applying the project's `.smaqit/slots.yml` bindings and the deployment path's rules), and the agent walks the plan — reading each filler's `SKILL.md` from its resolved path and following the named operation. The orchestrating skills (`smaqit.project-new` Phases 4–6, `smaqit.feature-new` Phase 3) resolve the mode and the plan through `smaqit.input-deployment` and invoke the agent; they never call infrastructure skills directly. Adding a `provisioning_mode` value or a deploy-skill family therefore means adding its defaults to the catalogue in `installer/slots.go` and having its skills declare `metadata.smaqit-slots` — not editing the agent or the orchestrators' step lists. The agent's own orthogonal knobs remain `specification_mode` (`orchestrate`/`prevalidated`) and `deployment_path` (`standard`, `existing-cicd-pr`, `ci-push`). See also: how does a project bind its own skill to a deployment step?

---

**How does a project bind its own skill to a deployment step?**

Through deployment slots. The deploy phase runs a fixed, ordered sequence of slots per environment — `platform-prereqs`, `access-credentials`, `platform-onboarding`, `target-credentials`, `environment-config`, `artifact-build`, `iac-generate`, `ci-generate`, `provision`, `host-bootstrap`, `app-deploy`, `verify`, `tls` — plus day-two `replace` and `teardown`, which run only as single-slot runs. Each slot has a default filler chosen by `provisioning_mode`. A project overrides a slot in its own `.smaqit/slots.yml` (created once by `smaqit init`, never overwritten, kept by `smaqit uninstall`) with `none` (skip), `default`, or `{skill, operation, params, environments}`. The bound skill must declare the slot in its frontmatter as `metadata.smaqit-slots: "slot=operation, ..."` — a single string, because the Agent Skills spec only allows string `metadata` values. Project-local skill directories (`.agents/skills/`, `.claude/skills/`, `.github/skills/`) are searched before the global ones.

`smaqit slots plan --phase=deploy --mode=<mode> --env=<env> [--path=standard|existing-cicd-pr|ci-push] [--slot=<slot>] [--json]` shows the resolved plan and refuses invalid bindings with `SLOTS-*` diagnostics, all reported together: a missing or undeclared skill, an invalid operation, a binding the path forbids (`existing-cicd-pr` refuses `iac-generate`, `ci-generate`, `provision`, `host-bootstrap`), or a local binding that would split a CI run. Paths also change defaults: on both CI paths (`existing-cicd-pr`, `ci-push`) `platform-prereqs` defaults to none, and `existing-k3s`'s `verify` defaults to deploy-verify's `ci-evidence`, which decides from the deploy run's own record. `smaqit slots validate` and `smaqit validate` check the file. Every filler runs under the Deployment agent's invariants (guarded IaC plan/apply, no out-of-band changes, credential references only). Full reference: `docs/wiki/workflows/deployment-slots.md`. See also: does the Deployment agent itself know about `provisioning_mode` values?

---

## Claude Code Support

**Can smaqit tell deterministically whether the user typed a slash command or the model invoked a skill?**

No. Commands and orchestrator skills tell the model which mode to use: a typed `/smaqit.<phase>`, `/smaqit.feature-new` or `/smaqit.project-new` (or a Codex `$` mention, or a skill-menu pick) means `assisted`, and the model choosing it means `autonomous`. The model then judges from the conversation which case it is in. Live Claude Code runs have shown this working for both the typed and the self-invoked paths.

Claude Code does offer deterministic signals:

- a `UserPromptExpansion` hook fires when a user-typed command expands, with its `command_name` and arguments, and `PreToolUse` doesn't fire on that path;
- a `PreToolUse` hook matching the `Skill` tool fires only when the model invokes a skill;
- both can add `additionalContext` that the model reads.

smaqit doesn't use them, for three reasons:

- they would have to be registered in the user's `~/.claude/settings.json` or in a plugin, because hooks in a skill's own frontmatter register only after that skill is invoked;
- GitHub Copilot and Codex have no verified equivalent;
- splitting each skill into a user-only (`disable-model-invocation: true`) entry and a model entry would double every listing.

See also: what mode does a smaqit phase run in, and what counts as a requirement?

---

**How do I start a new project with smaqit on Claude Code, and how do I run only the Develop phase?**

There are two levels of setup:

1. **Once per machine,** `install.sh` puts the `smaqit` binary on `PATH` and installs agents, commands and skills globally: `~/.claude/agents/`, `~/.claude/commands/`, `~/.claude/skills/`.
2. **Once per project,** `smaqit init` scaffolds:
   - `.smaqit/` (templates, state, and the git-ignored PlantUML runtime under `.smaqit/tools/`)
   - `specs/` and `docs/designs/` for the five layers
   - the `smaqit-plantuml` MCP registration in `.mcp.json`
   - an `AGENTS.md` smaqit block plus a `CLAUDE.md` that imports it
   - the Copilot setup workflow

Approve the `smaqit-plantuml` server in Claude Code (`/mcp` shows whether it's connected); Node 22+ is the only prerequisite.

Nothing forces Claude to use smaqit. The `AGENTS.md` block makes Claude aware of the specs, designs and commands, but smaqit runs only when you invoke a phase command or orchestrating skill. Because agents and skills are global, they are also present in projects without `smaqit init`, where they fail for lack of `specs/`, templates and the MCP server.

To run only Develop, use `/smaqit.development <request>`:
1. The command applies `smaqit.input-development` in the main conversation and invokes the `smaqit-development` subagent with the parameter block.
2. That subagent runs the business, functional and stack specification agents for any spec that is missing, `draft` or `failed` (and not already approved).
3. It consolidates the specs and runs `smaqit plan --phase=develop`, which also checks the design gate.
4. It implements, builds and tests the application, and generates design-sequence diagrams.

Deploy runs only when you invoke `/smaqit.deployment`. The specification agents have no slash command on Claude Code, but you can ask Claude to delegate to one by name (for example `smaqit-business`). The full lifecycle runs through the `smaqit.project-new` and `smaqit.feature-new` skills, which also need smaqit-extensions for task lifecycle. See also: when are the `smaqit.input-*` skills invoked?

---

**What's the Claude Code equivalent of `copilot-setup-steps.yml`?**

There is no direct equivalent because Claude Code has no magic-filename bootstrap convention. The closest local mechanism is a `SessionStart` hook in `.claude/settings.json`; cloud workflows must add their setup steps explicitly. `CLAUDE.md` is project context, not a bootstrap mechanism.

---

**What is `installer/commands-claude/`?**

It is the gitignored compiled-output directory holding Claude Code slash-command files. `scripts/generate-agents.py` generates it from root `commands/*.md`, and the global installer copies it to `~/.claude/commands/` (or `$CLAUDE_CONFIG_DIR/commands/`). Only development, deployment, validation, and QA receive direct commands; specification agents remain delegation-only.

---

**Does `smaqit init` install `copilot-instructions.md`, `CLAUDE.md`, or `AGENTS.md`?**

It installs `AGENTS.md` plus a thin `CLAUDE.md` that imports it. Copilot and Codex read `AGENTS.md`; Claude Code reads `CLAUDE.md`. Only smaqit's section, delimited by `<!-- smaqit:instructions:begin -->` and `<!-- smaqit:instructions:end -->`, is ever written. It is appended when absent, and replaced with the current template when present, so `smaqit update` (which re-runs `init`) delivers template changes to existing projects. Content outside the markers is preserved byte for byte, an already-current section is not rewritten, and a CRLF file keeps CRLF line endings. A section with malformed markers (begin without end, reversed, duplicated) is left untouched and `init` prints a warning instead of guessing where it ends. Before v3.9.0 the section was never refreshed once present. See `installInstructionsFile()` in `installer/main.go`.

---

**How does a coding-agent session know to use smaqit on its own, without the user typing a slash command?**

Several layers work together, and none is enough alone. The managed `AGENTS.md` section, loaded in every session on every platform (Claude Code reads it through `CLAUDE.md`'s `@AGENTS.md` import), has a "When to use smaqit" section. It sets the policy: it routes feature, build, deploy, test, and framework-question requests to the matching smaqit entry point, puts specifications first, lists the trivial changes that may skip the workflow, and states the return-signal rule. Projects don't add this by hand: `smaqit init` and `smaqit update` refresh the section between its markers, so `smaqit update` delivers it to existing projects.

Descriptions are the routing signal. On Claude Code, the phase commands are skills the model can invoke itself, each with a "Use when…" `description` plus `when_to_use` trigger phrases. The development, deployment, validation, and QA subagents say "Use proactively when…", the wording Anthropic's subagent docs recommend for encouraging automatic delegation. Copilot and Codex have no "proactively" convention, so their entry-agent descriptions just lead with "Use when…" (v3.10.1), and `AGENTS.md` remains their main signal. The instructions alone leave Claude to infer which agent fits a request, and the descriptions alone can't carry the specs-first rule, the exemptions, or the return-signal rule.

The five specification agents are deliberately never proactive: delegating to one directly would skip its phase's parameter resolution and ordering. The commands stay the preferred path, because they resolve phase parameters before delegating. Direct delegation to a proactive phase agent is safe for two reasons: an agent invoked without a `PHASE-PARAMETERS` block applies its preloaded defaults (including `mode: assisted`, the safe direction if a self-delegation forgets its block), and the `AGENTS.md` rule makes the main conversation put every `NEEDS-INPUT`/`REVIEW-REQUESTED` to the user, never answer it itself, and resume the agent with `answers`. Skill and command descriptions stay at 400 characters or fewer, because Claude Code drops descriptions from the least-invoked skills first when its listing exceeds its budget; `make prepare` enforces the limit. See also: does `smaqit init` install `copilot-instructions.md`, `CLAUDE.md`, or `AGENTS.md`?

---

**Why don't Claude Code slash commands show up when typing `/` in the VS Code extension chat?**

In the smaqit source repository, product commands exist only as generated installer staging and are not installed into the repository itself. After global installation, reload the VS Code window so it discovers `~/.claude/commands/` (or `$CLAUDE_CONFIG_DIR/commands/`). If the Linux or WSL extension still misses them, use the Claude CLI in an integrated terminal; the CLI surface can discover commands even when the extension panel does not.

---

## Infrastructure Skills

**How does `smaqit.project-new` pick a deploy skill when a project's stack matches none of the existing `smaqit.infrastructure-deploy-rsync*` family?**

Phase 4 reads the declared stack from the stack specification, compares it generically against installed deploy skills, and synthesizes a new skill when no match exists. The synthesized skill uses the existing deploy-rsync family as exemplars and preserves shared conventions such as the `__APP_DIR__` token, `write-vhost.sh`, deploy stamps, and guard-script reuse. A human checkpoint is required before first invocation unless prior autonomous approval exists.

---

**Why must every `terraform plan` or `apply` against already-provisioned infrastructure go through `plan-guard.sh`, even for diagnosis?**

The risk of a destructive delete or replace plan is the same whether the operator intends to apply it or only inspect it. Routing every invocation through `plan-guard.sh` ensures direct interactive commands receive the same protection as generated CI workflows. See also: where may a `plan-guard.sh` allow-flag address come from?

---

**Where may a `plan-guard.sh` allow-flag address come from, and can the guard stop an agent from allowing the wrong resource?**

Only from a human, before the plan ran: the request, the reviewed pull request, or a gated dispatch input. Never from the guard's own BLOCKED output. `--allow-delete` and `--allow-replace` make the guard accept a destroy or replacement of exactly the listed addresses. The guard checks that the plan matches the flags, not that the flags match anyone's intent. So an agent that copied a blocked address into a flag and re-ran would pass. A blocked delete is usually a side effect nobody reviewed. For example, removing one element of a `count` list plans a delete of the *last* index and updates the others.

The guard therefore says what to do in its BLOCKED message, the text an agent reads at that moment: stop, report the plan, and wait for a human decision. The provision-cyso failure table, cicd-generate's gated inputs, and the Deployment agent's guard rule state the same rule. This guidance is wording only. smaqit ships no human-confirmation gate (terminal confirmation, break-glass procedure): a downstream project that needs one enforces it on its own apply path. See also: does a destroy through `plan-guard.sh` always require a pull request?

---

**Does a destroy through `plan-guard.sh` always require a pull request?**

No. The guard inspects only the plan, never git state. A local run can edit Terraform files in a working copy, then run `plan-guard.sh --allow-delete <address> --out tfplan` and `terraform apply tfplan` against shared remote state, with nothing committed or reviewed. A plan can also destroy a resource without any file change, when an input changes (for example `count = var.enable_x` with `TF_VAR_enable_x=false` in the environment).

A pull request is required only where the repository enforces it: branch protection requiring PRs on `main`, plus applies that run only from `main` in CI through an Environment branch policy. The generated `deploy.yml` (full mode) runs the guard with no allow flags, so it never applies a destroy. It also has a `workflow_dispatch` trigger, so without an Environment branch policy it can be dispatched from an unmerged branch. See also: can smaqit prevent local applies to GitHub Environment targets?

---

**Can smaqit prevent local applies or deploys to targets that have a GitHub Environment?**

Not today; local applies are supported by design. `terraform-env.sh` exports a machine's cloud and state-backend credentials into a local shell. `credential-file.sh` hands out an app's deploy SSH key and a machine's kubeconfig as local files. provision-cyso performs its first `terraform apply` locally.

A policy limiting local runs to test targets could only be enforced where the credentials live:
- the scripts that hand credentials to a local shell refuse Environment-target credentials, which only the Environment sync scripts read;
- test targets use separate state and credentials;
- GitHub OIDC short-lived credentials would remove long-lived secrets entirely, if the cloud provider supports federation.

Agents reach Vault only through scripts, so a script-level refusal would stop an agent. A human with direct Vault access still could, deliberately. A `GITHUB_ACTIONS` check inside scripts can be set by anyone, so it only produces clear errors. Bootstrapping (loading credentials into Vault and syncing them into the Environment) stays local either way; it configures rather than deploys.

---

**What causes SSH keys stored in Vault to fail with `error in libcrypto`?**

Shell command substitution strips the private key's required trailing newline. Store the key with Vault's `@file` syntax instead of `private_key="$(cat file)"`; file syntax preserves the exact bytes. See the Gotchas in `smaqit.infrastructure-vault-loader`.

---

**What Vault namespace convention does smaqit use for machine credentials vs. app credentials?**

Credentials split across two namespaces by what they belong to. `secret/machines/<machine-slug>/*` holds everything scoped to a provisioned VM regardless of which app runs on it: `base-ssh` (the bootstrap-only credential Terraform installs at provision time — never used for routine deploys), `cyso` and `tfstate` (cloud-provisioning credentials, since provisioning is a property of the machine, not any one app), and `metadata` (non-secret host/provider/owner-project info). `secret/apps/<app-slug>/*` holds everything scoped to an individual app: `ssh` (a distinct keypair per app, no exceptions — even the project that originally provisioned the machine gets its own, bootstrapped rather than reused), `github`, and `machine` (a pointer recording which machine-slug this app is bootstrapped against, used by `rotate-credential.sh` to know where to re-authorize a rotated key). An app's `ssh` credential is always populated via `smaqit.infrastructure-vault-loader/scripts/bootstrap-app-to-machine.sh <app-slug> <machine-slug>` — idempotent, never by copying another app's key material or letting Terraform install it directly. Projects predating this convention still use the older flat `secret/<project-slug>/{cyso,ssh,tfstate,github}` scheme with no machine-level namespace at all; `load-credentials.sh` supports both, auto-detecting which applies per invocation, and migrating an existing flat-scheme project onto `apps/`+`machines/` is a manual, project-by-project decision rather than something any smaqit skill does automatically. See also: what Vault credentials does `provisioning_mode: existing-k3s` use?

---

**How does `load-credentials.sh` decide whether to prompt for installing an SSH public key manually, versus assuming it gets installed automatically?**

The decision is per `provisioning_mode`, independent of which credential scheme (legacy flat vs. `apps/`+`machines/`) is in play. `provision`/`existing-owned` generate the deploy keypair silently and store it in Vault, because a Terraform `apply` is expected to push the public key onto the VM via cloud-init — no manual step is needed. `existing-shared` and `existing-unmanaged` both skip Terraform for this project entirely, so neither has any automated path to install the key onto the VM; both print explicit "append this public key to `~/.ssh/authorized_keys` yourself" instructions after generating it. The two differ only in *how* the key gets there: `existing-shared` additionally offers copying the already-trusted keypair from the project whose Terraform state owns the VM (a real option, since that project already has access); `existing-unmanaged` has no such option, since by definition no project's Terraform manages the VM — every credential load for that mode generates a fresh keypair and always needs the manual-install step. `existing-shared` and `existing-unmanaged` also both skip prompting for `cyso`/`tfstate` credentials in the legacy flat-scheme path, for the same underlying reason (no Terraform run for this project against this VM either way).

---

**How does `load-credentials.sh`/`rotate-credential.sh` derive a project's Vault slug?**

Both scripts source a shared `derive_project_slug()` (`skills/smaqit.infrastructure-vault-loader/scripts/lib-project-slug.sh`): the basename of `git remote get-url origin` (`.git` suffix stripped), falling back to the current working directory's basename if there is no git remote, falling back to a manual prompt only if both are empty. Neither script parses `AGENTS.md`'s human-readable "Project Name" field for this purpose — a project's Vault identity is always its repo/directory identity, never a doc title, since a title has no guaranteed relationship to the technical slug already used throughout the rest of this framework's Vault conventions (see the Vault namespace convention entry above). `check-no-ad-hoc-secret-reads.sh` in the same directory is a separate, unrelated static check (secret-read `/dev/tty` safety, not slug derivation).

---

**What's the difference between app onboarding, requesting onboarding, and app deployment in smaqit's k3s infrastructure-skill family, and why does it matter for skill design?**

Three distinct concerns, but only two have a smaqit skill — the third is deliberately not one:

1. **Onboarding** — granting an app a tenant slot on an already-provisioned k3s cluster (Namespace, RBAC, Pod Security, NetworkPolicy, quota, kubeconfig issuance). This is entirely the cluster-owning infra repo's own concern: that repo defines its own onboarding contract, and the contract legitimately varies per repo. smaqit does not ship a skill for it — an earlier stub, `smaqit.infrastructure-onboard-k3s-app`, existed only as prose deferring to "whatever the infra repo's own process is," had no caller anywhere in the skill tree, and was removed (task 119) once skill 2 below made it redundant.
2. **Requesting onboarding** (`smaqit.infrastructure-request-k3s-onboarding`) — the app repo's own means of asking for a tenant slot without needing direct write/dispatch access to a repo it doesn't own: opens a PR against the platform repo's registry file and gates on merge, a human-review handoff.
3. **Deployment** (`smaqit.infrastructure-deploy-k3s-app`) — the app's own CI applying its manifests into the Namespace using the onboarding-issued kubeconfig, once onboarding (concerns 1-2) has already happened.

This split has a direct design consequence: concern 1 lives inside an infra repo smaqit doesn't control the internals of, so smaqit cannot own conventions for it beyond recognizing the hand-off point — its own Pre-conditions already state plainly that a target repo whose onboarding process isn't PR-based is out of scope, which fully covers what the removed stub used to say. Skills 2 and 3 live inside the smaqit-managed **app** repo instead, which smaqit fully owns the conventions for, so they can and do declare one fixed, opinionated contract each (a PR-to-a-registry-file request model; a Pod-Security-`restricted`/quota-guardrail deploy model) rather than deferring to unknown mechanics on their own side. Using `existing-k3s` with skill 2 implicitly requires the target infra repo's onboarding process to accept requests via a PR-mergeable registry file — an infra repo whose onboarding works some other way is simply not compatible with that skill, a scoping boundary rather than something the skill adapts around.

---

**What Vault credentials does `provisioning_mode: existing-k3s` use, and how do they differ from the VM-based conventions?**

Three app-scoped fields (four for a `private-mesh` machine), all under `secret/apps/<app-slug>/*`; the app keeps nothing under `secret/machines/*`, since from its side a k3s machine has no SSH and no Terraform state (the platform repository that owns the machine may still register it there with `register-machine.sh`):

- `secret/apps/<app-slug>/github` — same as every other mode, scoped to this project's own repository.
- `secret/apps/<app-slug>/<machine-slug>/kubeconfig` — a single `value` field per registered machine-slug. "Environment" is expressed entirely through which machine-slug is targeted: test and prod get distinct machine-slugs even when they happen to share one physical k3s server, rather than one path split into `test`/`prod` fields — a flat, environment-keyed structure can't represent an app onboarded onto genuinely different machines per environment. Always populated out-of-band (an operator paste of the platform's own onboarding hand-off, never a live cluster call); rotation re-prompts for a freshly platform-reissued value rather than regenerating locally — how the platform rotates (in place with the old token revoked, or only by re-onboarding) is the platform's decision. After rotating, re-sync the Environment and redeploy with a fresh dispatch, not a re-run.
- `secret/apps/<app-slug>/platform-repo` — a fine-grained PAT scoped only to `contents:write` + `pull_requests:write` on the platform-owned infrastructure repository, distinct from `github`. Used by `smaqit.infrastructure-request-k3s-onboarding` to open a PR there; never has `workflow_dispatch` scope, since triggering convergence is entirely the platform repo's own concern once the PR merges. Unlike `kubeconfig`, this follows the standard delete-and-repopulate rotation shape, since it's a smaqit-managed PAT the operator can freely regenerate, not a platform-issued artifact.

- `secret/apps/<app-slug>/<machine-slug>/mesh-setup-key` — a single `value` field, only for a machine whose Environments row is `private-mesh`: the platform-issued setup key the app's CI peer joins the mesh with. Loaded hidden with `MESH_SETUP_KEY=1`, replaced with `rotate-credential.sh apps/<app>/<machine>/mesh-setup-key`, and synced into the machine's Environment under the spec's `Mesh Setup-Key Secret` name by `sync-k3s-env.sh --mesh-secret <name>`.

`load-credentials.sh` requires `MACHINE_SLUG` to be set for this mode and populates one machine's kubeconfig per invocation — run it again with a different `MACHINE_SLUG` to populate a second machine. A merged onboarding request doesn't produce the kubeconfig: the platform's own hand-off delivers it, and the `target-credentials` slot stops the deployment until it's loaded. `REGISTRY_ORG=<org>` additionally loads the org-scoped `secret/organizations/<org>/github-package-read` pull credential when the image is private. In the app's repository each target machine gets a GitHub Environment named after its machine slug, synced with `smaqit.infrastructure-repo-config`'s `sync-k3s-env.sh`.

---

**How does an `existing-k3s` app deploy to a private-mesh or platform-owned target?**

Each environment declares it in one row of the Infrastructure spec's `## Environments` table: `Reachability` (`public` or `private-mesh`) and `Ingress Mode` (`own-tls`, `platform-default-tls`, or `platform-owned`), plus its machine slug, host, Ingress class, ClusterIssuer (`own-tls` only), health path, and for `private-mesh` the mesh setup-key secret name and the API's mesh name. Constraints rows add the mesh-join action (a pinned `uses:` reference and its input names) and the `Probe Client`. There is no default: `installer/infra_spec.go` parses the table, `smaqit slots plan --mode=existing-k3s` refuses an incomplete row with `SLOTS-SPEC-*` lines, and `smaqit spec environments --json` prints the rows the fillers read.

The default fillers branch on the row. On a local deploy (`standard` path), `platform-prereqs` runs `deploy-k3s-app`'s `mesh-check`, which confirms the API answers over the mesh from this machine; the CI paths skip it. `ci-generate` renders a single dispatch-only `deploy.yml` (`render-k3s-deploy.sh`) for every environment. It carries a per-machine `K3S_TARGETS` map, and its steps branch with `if:` on the dispatched machine: mesh join first for `private-mesh`; `ingress.yaml`, `ingress-default-tls.yaml`, or no Ingress; the `Certificate` wait only for `own-tls`. Verification runs in-cluster whenever the runner can't reach the app (`private-mesh` or `platform-owned`): a probe Job runs the app's own image against the Service over cluster DNS, and its verdict is the Job's status, so no `pods/log` grant is needed. Every run also checks that the Deployment's image tag is the commit SHA. After a CI deploy, the `verify` slot doesn't reach the target again: deploy-verify's `ci-evidence` reads `gh run view <id> --json headSha,status,conclusion,jobs` and passes only when the run succeeded, its head is the intended commit (merge or pushed SHA) or contains it (`gh api` compare `identical` or `ahead`), and its "Verify revision" and the row's verify step succeeded. The operator's machine needs no mesh access, kubeconfig, or Vault for a CI deploy. Golden files under `installer/testdata/k3s-deploy/` cover every combination.

**Exec checks.** The workflow fails when the deployer can exec into pods, and it checks this twice:
- before any apply, `namespace-guard.sh` runs `kubectl auth can-i create|get pods --subresource=exec`;
- after the rollout, `exec-refused.sh` makes a real `kubectl exec`, which must be refused as Forbidden.

Never write `can-i … pods/exec`: kubectl reads that as a pod named `exec`. A `Deployer Exec: allowed` Constraints row turns a deliberate exec grant into a warning.

**Spec-wide Constraints rows.** `Probe Client` may be `curl`, `wget`, or `python3`. A private `Container Registry` (`<host>, private`) renders the pull-Secret step.

**Secrets.** Every secret reaches a step through `env:`, never inside a `run:` script or on a command line. The kubeconfig is written owner-only to `$RUNNER_TEMP` and removed by an `always()` step. Concurrency allows one deploy per machine at a time. See also: what Vault credentials does `provisioning_mode: existing-k3s` use?

---

**Do the `kubectl auth can-i` TYPE/NAME trap and the `bash -e` trap affect smaqit's k3s deploys, and who guards against them?**

Both traps are kubectl and GitHub Actions behaviour that smaqit can't remove:

- **The TYPE/NAME trap.** `kubectl auth can-i create pods/exec` reads `pods/exec` as "a pod named exec", so any deployer allowed to create pods gets `yes`. The correct form is `--subresource=exec`. `can-i` doesn't validate subresource names, so a misspelled one silently answers `no`.
- **The `bash -e` trap.** GitHub runs every `run:` step with `bash -e`, so a step that captures a command expected to exit non-zero ends before it reads the answer. That includes a `can-i` "no" and a `kubectl exec` that must be refused.

smaqit's own checks are immune to both:

- `namespace-guard.sh` asks with `--subresource=exec`.
- `exec-refused.sh` brackets its real exec attempt with `set +e` and decides from the captured output.
- The generated `deploy.yml` runs both, and tests pin these forms and run the rendered steps under `bash -e`.

For steps a project adds itself, smaqit gives guidance, not enforcement. The `smaqit.infrastructure-deploy-k3s-app` and `smaqit.infrastructure-cicd-generate` skills describe both traps as gotchas, and their failure-handling tables say what to do when a project's own check answers misleadingly or ends on its expected failure. A static checker for project files was considered and rejected as too brittle. See also: how does an `existing-k3s` app deploy to a private-mesh or platform-owned target?

---

**How do smaqit skills read or write Vault without running `vault` commands?**

Through purpose-named scripts; skill text never tells an agent or operator to run `vault` directly, and a Go test (`installer/skill_text_test.go`) fails if a run-able `vault` command or a `Bash(vault:*)` grant appears in any shipped skill document. Scripts may call `vault` internally. The interface:

- **Load or populate:** `load-credentials.sh` (unseals, authenticates, enables kv-v2 on first run, prompts only for missing paths), `register-machine.sh` (a machine's own credentials), `bootstrap-app-to-machine.sh` (an app's SSH key on a machine), `rotate-credential.sh` (replace an existing path).
- **Use:** `credential-file.sh <path>` writes one allowlisted credential (an app SSH key, a kubeconfig, the platform-repo token) to a 0600 file in a private temp directory and prints only the path; `source terraform-env.sh` exports Terraform's provider and backend credentials into the shell (`--unset` clears them).
- **Sync to GitHub:** repo-config's `sync-secrets.sh` (VM modes), `sync-k3s-env.sh` (an app's machine-slug Environment), `sync-machine-env.sh` (a platform repository's machine Environment).

Secrets reach `vault` and `gh` on stdin, never as command arguments. Operator tooling outside smaqit may call `load-credentials.sh` (no arguments, legacy scheme) and `bootstrap-app-to-machine.sh <app> <machine>` directly, so those interfaces and the Vault path and field conventions stay stable. If an operation has no script, add one rather than typing a raw `vault` command.

---

**How does a platform repository (one that owns machines, not an app) register a machine and configure its GitHub Environment?**

Through two shipped slot operations, bound in the repository's `.smaqit/slots.yml`: `target-credentials: {skill: smaqit.infrastructure-vault-loader, operation: register}` and `environment-config: {skill: smaqit.infrastructure-repo-config, operation: sync-machine}`. Both skills also keep their original operation as the default (`target`, `sync`), so a binding to either must name the operation or `smaqit slots plan` refuses it with `SLOTS-OPERATION-INVALID`.

`register` runs `register-machine.sh <machine-slug>`, the single entrypoint for first-time population of `secret/machines/<machine-slug>/{base-ssh,cyso,tfstate,metadata}`. The slug is a required argument, never derived from the working directory, and is restricted to lowercase letters, digits, and inner hyphens because it also names the GitHub Environment. It keeps every path that already exists, prompts on the terminal only for missing ones (secrets hidden, handed to Vault on stdin), and refuses empty answers. After apply, `register-machine.sh <machine-slug> --host <ip>` backfills `metadata.host`, which `rotate-credential.sh machines/<slug>/base-ssh` requires. `rotate-credential.sh` is rotation-only. `smaqit.infrastructure-provision-cyso` Step 7.5 uses the same script for an app-owned VM.

`sync-machine` runs `sync-machine-env.sh <owner>/<repo> <machine-slug>`: it creates or verifies an Environment named after the machine slug, preflights the token with `check-pat-permissions.sh`, and pipes six secrets (`BASE_SSH_*`, `CYSO_APPLICATION_CREDENTIAL_*`, `CYSO_S3_*`) plus the variables `RECONCILE_SSH_USER` and `FIXED_IP` into it. Because `environment-config` precedes `provision` in the slot order, the first run happens before the machine has an address and skips `FIXED_IP`; re-run the slot alone after the `--host` backfill. A fine-grained PAT needs **organization-level** Secrets and Variables read/write for Environment writes — repository-level grants alone return 403.

Everything else — provisioning, host baseline, machine verification, replacement, and the tenancy contract — stays the platform repository's own project skills, bound to `provision`, `host-bootstrap`, `verify`, and `replace`; smaqit deliberately ships no canonical copies (task 115 found the downstream mechanisms had changed wholesale within weeks). A deliberate instance replacement still goes through smaqit's guard: `plan-guard.sh <dir> --allow-replace <address> --out tfplan`, then `terraform apply tfplan` — it plans with `-replace=`, passes only a true replacement of the named address (either action order), blocks everything else, and keeps the exact plan it inspected. A reviewed removal of one resource (closing a firewall opening, retiring a port) uses `--allow-delete <address>` the same way: it adds nothing to the plan and passes only a pure destroy of that address. An allowed address must come from a human before the plan ran (the request, the reviewed PR, a gated dispatch input), never from the guard's BLOCKED output; when the guard blocks, stop and report. `docs/wiki/workflows/platform-repositories.md` has the pattern, a machine-verification checklist, and the live-found gotchas. See also: how does a project bind its own skill to a deployment step?

---

## Codex Support

**How does smaqit provide first-class Codex compatibility?**

`scripts/generate-agents.py` compiles canonical agent bodies and platform metadata into `installer/agents-codex/*.toml`, and renders the shared product skills into `installer/skills-shared/` for the global `~/.agents/skills/` root. The shell installer and updater install Codex agents to `~/.codex/agents/` (or `$CODEX_HOME/agents/`) and shared Copilot/Codex skills to `~/.agents/skills/`; `smaqit init` does not create project agent or skill mirrors. It still registers the exact `smaqit-plantuml` server table in the initialized project’s trusted `.codex/config.toml`, together with the other project-local MCP configuration. Uninstall removes only exact SmaQit-owned global entries while preserving unrelated or nested custom content.

---

**What must be updated when adding a new canonical skill?**

The total shipped skill count is asserted in `installer/main_test.go` in two tests that must be bumped together: `TestRemoveEmbeddedSkillDirsPreservesUnownedSharedContent` (a hardcoded `removed != N` check against the `skills-shared` embed) and `TestSharedSkillsServeCopilotAndCodex` (a hardcoded top-level skill-directory count, plus placeholder/path invariant checks). `scripts/smoke-test-installer.sh` asserts one representative skill file's presence rather than a count. Missing either Go count fails `go test` with a count mismatch, not a useful error about which skill changed.

---

## Release Workflow

**How do I push to, or open a PR on, a repository in another GitHub organisation from this workstation?**

The default git credential on this workstation only reaches the user's own repositories, so a fetch or push to another organisation's repository fails with HTTP 403. Load that organisation's token through the Vault scripts, never with `vault` directly:

- `bash ~/projects/scripts/vault-gh-token.sh --list` shows the available token references (names and paths only).
- `source ~/projects/scripts/vault-gh-token.sh organizations/<org>` exports `GH_TOKEN` into the current shell without printing it. Each tool call is a fresh shell, so source it in every command that needs it.

`gh` reads `GH_TOKEN` directly. For `git`, set `gh` as the credential helper for one command only, so the user's git config never changes:

`git -c credential.helper= -c 'credential.helper=!gh auth git-credential' push -u origin <branch>`

Organisation tokens may lack scopes that some `gh` subcommands need, such as reading check runs. `gh pr view --json body` can also fail on GitHub's retired Projects (classic) API; reading and patching the PR through `gh api repos/<owner>/<repo>/pulls/<n>` works instead.

---

**Why can a release worktree remain after its release PR has merged and published?**

Publishing a release does not automatically remove the local `release/vX.Y.Z` worktree or branch: Git keeps it registered until an explicit cleanup removes the worktree, deletes the merged local branch, and refreshes the project workspace. Confirm the PR, tag, and release workflow first; then clean up the local release workspace. This separates publication verification from a recoverable local cleanup step.

---

**How can a local smaqit release authenticate an encrypted SSH key from WSL2 or another interactive desktop Linux session?**

Preferred path: `smaqit.release-git-local`'s "Desktop Linux SSH Agent Recovery" procedure discovers an already-running desktop SSH agent socket (GNOME Keyring's `gcr-ssh-agent`, an alternate keyring path, GnuPG's agent socket, the current `$SSH_AUTH_SOCK`, or the systemd user session's) by testing each candidate with `ssh-add -l`, then scopes `SSH_AUTH_SOCK` to a single retry of the exact failed git command — no persistent config changes, no new agent started, no identity loaded or removed. This works whenever the user's desktop keyring is already unlocked (commonly true for the whole desktop session once logged in, e.g. GNOME's Login keyring auto-unlocking with the account password) — no popup or passphrase prompt is needed in that case, since the key material is already available.

Fallback when no usable socket is found: load the key into a temporary `ssh-agent` and use `SSH_ASKPASS_REQUIRE=force` with a WSLg password dialog so the passphrase never passes through chat, command arguments, or logs. This requires a GUI askpass binary (`zenity`, `ssh-askpass-gnome`, or similar) already installed — none ships by default, and installing one via `sudo apt` is blocked by the Claude Code auto-mode permission classifier.

In both cases: confirm authentication with `ssh -T git@github.com`, push `main` and the annotated tag separately, verify both remote refs. The `gh` CLI's HTTPS token is not a usable fallback for either path: it authenticates read operations but returns `403` (no write scope) on push. Note that the Claude Code auto-mode permission classifier may independently deny a push (particularly a tag push) or even a read-only `ls-remote`, with no further reasoning than a generic denial — this is unrelated to SSH/credential state and requires either a Bash permission rule or the user pushing that specific step from their own shell. If no agent socket is usable and no askpass tool is available, the default expectation is a local commit + annotated tag + verified build, with the actual push left to the user's own shell.

---

**Does a later per-task release always contain an earlier one's code?**

Yes, guaranteed by construction: each task's PR is built on whatever `origin/main` tip existed when its branch was created, and PRs merge in review order — so a task that merges later always has the then-current `main` (including every already-merged earlier release) as an ancestor of its own release tag. This holds even when two tasks' version *numbers* land out of strict numeric order (see the boundary-vs-baseline distinction below) — ancestry and version numbering are independent guarantees. Verify directly rather than assuming from version numbers alone: `git merge-base --is-ancestor v<earlier> v<later>` (exit 0 confirms it) or inspect a specific file's content at either tag with `git show v<tag>:<path>`.

---

**Why does pushing to `main` not trigger a new release build?**

`post-merge-release.yml` triggers only on a `v*` tag push or a merged pull request to `main` — a plain branch push never fires it, regardless of what changed. Push the annotated release tag separately (`git push origin vX.Y.Z`) to trigger it. A newly triggered run can take a couple of minutes to appear in `gh run list` or the Actions UI; check via `gh api repos/<owner>/<repo>/actions/runs` (unfiltered by status) rather than assuming a run that isn't immediately visible never started.

Under the per-task PR-gated release model, every task's own PR *is* its release: merging it is normally sufficient, with no separate manual tag or release step needed. The one hard requirement is the PR title — it must exactly match `Prepare release vX.Y.Z` (or `Release vX.Y.Z`), since the workflow matches on title and silently skips every job on a mismatch (e.g. extra words, wrong prefix, wrong casing).

A PR merged under any other title (for example a `fix(...)` title) lands its code on `main` but publishes no tag, binaries, or GitHub Release, so `smaqit update` has nothing new to install. To ship such a change, cut a branch from `origin/main` that adds only that version's `## [X.Y.Z]` `CHANGELOG.md` section, and merge it as a `Prepare release vX.Y.Z` PR. The workflow extracts the release notes from that section. Branching from `origin/main` rather than local `main` keeps unpushed local task bookkeeping commits out of the release PR.

---

## Skill Development

**Can a shipped skill reference files under `framework/`?**

No. The installer's `go:embed` manifest in `installer/main.go` only ships `agents-*`, `commands-claude`, `skills-*`, and the `AGENTS.md`/`CLAUDE.md` templates — `framework/` is never installed into a consumer project. It exists only in this canonical repo as agent-facing documentation for developing smaqit itself. A skill that depends on framework-documented mechanics (e.g. the Incremental Spec Updates decision table, spec state transitions) must distill the relevant content into its own `references/` file rather than pointing at `framework/*.md`, the same way `smaqit.project-new` stays fully self-contained with zero `framework/` references.

---

**When should a contributed skill definition be compiled into a shipped product skill immediately, versus held back as definitions-only pending validation?**

Compile immediately when the source mechanism is already hardened through multiple real rounds of production use — withholding compilation in that case serves no purpose. Hold back as definitions-only (a file under `.smaqit/definitions/skills/`, never a compiled `skills/` directory) only when the mechanism is an unproven, single-use synthesis still awaiting real-world validation, matching the precedent set by `smaqit.infrastructure-deploy-rsync-python-tornado`'s initial contribution. A skill definition contributed this way is a candidate for a later, separate reconciliation task once proven — see task 106's precedent for that two-step shape.

Separately, a generic smaqit product skill must never hardcode one specific contributor's own implementation mechanics as if they were a universal contract, regardless of how proven that specific implementation is — see the app-onboarding-vs-deployment entry under Infrastructure Skills for the concrete failure mode and its correct shape (a thin dispatcher, not a mechanism).

---

## Task Management

**Where does the task-lifecycle tooling (`task-start`, `task-create`, `task-complete`, `task-list`, `utils.worktree`) live, and does `smaqit init` install it?**

It is owned entirely by the sibling `smaqit-extensions` repository, not by `smaqit`'s own `skills/` directory or its shipped product skills. `smaqit init` does not install it — a consumer project only gets task-lifecycle skills (branch/worktree creation, parent-child task ownership, assisted/autonomous mode) if `smaqit-extensions` is installed separately. Shipped smaqit skills that reference the lifecycle (`smaqit.project-new`, `smaqit.feature-new`) call `smaqit.task-start`/`smaqit.task-create`/`smaqit.task-complete` by name without re-implementing any of their branch, worktree, or merge mechanics — those skills assume the lifecycle tooling is present, they do not provide a fallback if it isn't. In this repo, the copies under `.github/skills/`, `.claude/skills/`, and `.agents/skills/` for task-lifecycle skill names are committed dogfooding install output from `smaqit-extensions`, not canonical smaqit source.

`smaqit-extensions` also ships a parent-owned subtask lifecycle (`Parent: NNN` task metadata): a child task joins its parent's existing branch/worktree and inherits its mode instead of creating its own; only the parent merges and cleans up, and only once every declared child is `Completed`. `smaqit.feature-new` adopts this contract for its own five-phase structure (one shared feature-cycle parent, five phase children) rather than spawning a separate branch per phase.

---

**How do you determine whether an old task file in `.smaqit/tasks/` is still relevant before starting or completing it?**

Task files are written speculatively and can go stale as the framework evolves out from under them — re-reading the task text alone is not sufficient. Check three things against the *current* codebase rather than trusting the file: (1) do the file paths, commands, or skill names it references still exist (e.g. a task written when prompts were still `.github/prompts/*.prompt.md`, or when `PLANNING.md` lived under `docs/tasks/` instead of `.smaqit/tasks/`, is describing a mechanism that may no longer exist); (2) has a later task already superseded or completed the same underlying work, sometimes without ever being run through the task's own file (check `git log` on the files the task claims to touch, not just `PLANNING.md`'s status column); (3) has the concern the task wants to validate already been implicitly exercised by extensive real usage since it was filed, with no issue ever surfacing. A task can fail all three checks and still describe a real, unaddressed gap — in that case the fix is usually much smaller than the original task once rescoped to current architecture, and may not warrant a tracked task on its own.

---

## Bench

**Can smaqit's Bench suite test how Claude Code runs smaqit agents?**

Not soundly yet. The `smaqit-adk bench` engine runs whatever `process.executable` a manifest names, but every existing manifest drives `codex exec`, and three gaps block a valid Claude Code run:

- **Staging and shadowing.** Bench stages the version under test into a read-only `.smaqit-bench-input/` sidecar that Claude Code doesn't read. Copying it into the project's `.claude/` doesn't help, because personal `~/.claude/skills` override project skills with the same name, so an installed smaqit release would mask the version under test.
- **Credentials.** Avoiding that shadowing needs an isolated `CLAUDE_CONFIG_DIR`, which has no login. Credentials must be passed in on purpose and never frozen into evidence; smaqit-adk task 035 found Codex credentials leaking into a workspace that way.
- **Conversation turns.** Headless Claude Code has no user to answer questions, so it covers single-turn Cases only. Testing an ask-and-resume loop needs scripted follow-up turns.

smaqit-adk task 037 (Bench harness selector) would make Claude Code a declared harness that handles all three. Until it lands, the standalone `make -C installer e2e-claude` test does that work outside Bench. It isolates the install itself, takes credentials only from the environment, and scripts follow-up turns with `--resume`. See also: how do I test smaqit end to end on Claude Code?

---

**How do I test smaqit end to end on Claude Code?**

Run `make -C installer e2e-claude`. Credentials come from `CLAUDE_CODE_OAUTH_TOKEN` or `ANTHROPIC_API_KEY` if either is set. Otherwise the script reuses your current claude.ai login: it reads only the access token, passes it as `CLAUDE_CODE_OAUTH_TOKEN`, writes nothing to the scratch folder, and never uses the refresh token. `scripts/e2e-claude.sh` takes these steps:

1. Builds the dev binary.
2. Points `HOME` and `CLAUDE_CONFIG_DIR` at a scratch folder and confirms the isolation with `claude auth status`.
3. Installs smaqit globally there and runs `smaqit init` in a fresh project per scenario.
4. Drives headless `claude -p` with live questions disabled, in Claude Code's `auto` permission mode.

There are six scenarios. Four build the `scripts/e2e/hello-site/` static site:

- `happy-path`: complete requirements and "run autonomously"
- `needs-input`: thin requirements → `NEEDS-INPUT` → a `--resume` turn with answers
- `assisted`: no mode stated, so the typed command defaults to assisted: `REVIEW-REQUESTED` after each layer, and approved specs never regenerated
- `self-invocation`: a plain request with no slash command still reaches `smaqit-development` and runs autonomously

Two build nothing and pass only if nothing is written under `specs/` and the run returns `NEEDS-INPUT`:

- `empty-request`: a bare `/smaqit.development`
- `title-only`: the plain request "Let's build mario-app." in a project named `mario-app` whose README is only its title

Checks use smaqit's own CLI (`validate`, `plan --phase=develop`, `status`), the site's `node --test`, and the stream-json session log, never the exit code alone. Evidence lands in `installer/test/e2e-claude/<timestamp>/`.

Every run spends real model usage, so it is a manual pre-release check. Limit it with `SCENARIOS`, `E2E_BUDGET_USD`, `E2E_TIMEOUT` and `E2E_MAX_TURNS`. It replaces the Copilot-SDK `make test-e2e` path (`e2e-test.yml`, `run_e2e_test.py`), which was retired after its `smaqit.ci-testing` agent was deleted. Full reference: `docs/wiki/workflows/testing-smaqit.md`.

---

**What did the live Claude Code e2e runs show about how smaqit's orchestration behaves?**

It works end to end. On Claude Code 2.1.296, all four `make e2e-claude` scenarios passed. Each cost about $0.70–0.90 in Claude Code's notional cost figure; with claude.ai login that counts against the subscription. The runs also showed several mechanics worth knowing when reading evidence or changing agents:

- **The phase agent usually runs in the background.** `/smaqit.development` applies the input skill in the main conversation, then starts `smaqit-development` as a background subagent and relays its result later. A headless run waits for background work, but the result arrives as `system/task_notification` events and extra `result` events, not as a `tool_result` block. One run therefore emits several result events.
- **Nested subagents can detach** (claude-code issue #75043). When `smaqit-development` started `smaqit-business` without `run_in_background: false`, the child ran detached and the parent polled the filesystem until the spec appeared. It explicitly passed `false` for the next layers. Phase agents should always request foreground specification agents.
- **The return signals work headless.** `NEEDS-INPUT` reached the main conversation and the run wrote nothing, then a `--resume` turn with the answers completed the phase. Assisted mode paused with `REVIEW-REQUESTED` after each layer, and approved specs kept their content.
- **Completing the phase legitimately edits approved specs.** It sets `status: implemented`, ticks acceptance criteria, and links the design-sequence diagram. A "regenerated?" check must ignore exactly those three.
- **Self-invocation works.** With no slash command, Claude routed a feature request to smaqit, citing the `AGENTS.md` "When to use smaqit" block.
- **Use auto permission mode, not per-command Bash rules.** Agents compose shell commands (`ls … && smaqit status | head`), and per-command rules denied them. `--permission-mode auto` gave zero denials without bypassing permissions.

See also: how do I test smaqit end to end on Claude Code?

---
