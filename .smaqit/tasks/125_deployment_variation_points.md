---
status: In Progress
mode: Assisted
started: "2026-10-11"
created: "2026-10-11"
---

# Deployment Variation Points: Slot-Driven Deployment Orchestration With Project-Bound Fillers

## Description

smaqit's deployment flow is not orchestrated by the deployment agent. `agents/deployment.md` is a
generic phase shell ("Transform consolidated specifications into IaC…", "Deploy… via trusted execution
layer"); the actual skill wiring — credentials → CI generation → provision → host bootstrap → deploy →
verify → TLS — is hardcoded per `provisioning_mode` inside two app-centric orchestrator skills,
`smaqit.new-greenfield-project` (Phases 4–6) and `smaqit.feature-new` (Phase 3). The L0 principle
behind it is explicit: "Deterministic routing — The invocation sequence is hardcoded per phase, not
derived at runtime" (`framework/AGENTS.md`, Phase Orchestration).

That leaves a platform repository — one that owns machines rather than an app — with no way to run its
provisioning through smaqit. A downstream platform repository (Cyso VMs running single-node k3s,
mesh-only access, gated `workflow_dispatch` applies, one GitHub Environment per machine, Vault only
through project scripts) found every canonical infrastructure skill misfitting
(`provision-cyso`, `vm-bootstrap`, `repo-config`, `vault-loader`, `deploy-verify`, `domain-tls` all
assume an app-owned VM over public SSH) and built a parallel provisioning path in its own project
skills, which smaqit's agents never reach. Tasks 114 (Finding 5: "a mode of `smaqit-deployment` or a
distinct machine-deployment agent", TBD), 115, 122 and 123 each patch one axis of the same problem by
adding branches to the closed `provisioning_mode` enum.

Make `smaqit-deployment` the authoritative deployment orchestrator for app and platform repositories
alike: each phase declares a fixed, ordered sequence of **slots** (variation points). Each slot has a
contract (purpose, inputs, handoff, completion signal) and a canonical default filler skill. A project
binds any slot to its own skill (and operation) in `.smaqit/slots.yml`, referenced from its
`AGENTS.md`; an unbound slot uses the default; `none` skips the slot. Fillers declare the slots they
fill in their own metadata, and every filler keeps the phase's invariants (guarded plan/apply, no
out-of-band fixes, credential references only). Task 116's thin-dispatcher precedent ("each
infra-owning repo owns its own contract") is the philosophical basis.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Claude Code, GitHub Copilot, Codex CLI, smaqit-adk, Go, gopkg.in/yaml.v3
**Platforms/Environments:** Claude Code, GitHub Copilot, Codex CLI
**Features/Integrations:** subagents (Task tool), Agent Skills, AGENTS.md project instructions, YAML configuration
**Versions/Constraints:** smaqit 3.6.1, installer skill-count test, AGENTS.md managed block is never rewritten by the installer, `smaqit update` re-runs init (skeleton must be create-if-absent)

## Design Decisions

- **Orchestrator = the deployment agent** (user, 2026-10-10): `smaqit-deployment` owns the slot
  sequence and invokes the bound filler per slot; `new-greenfield-project` and `feature-new` stop
  hardcoding infrastructure skills and route through it. Resolves task 114 Finding 5 as "a mode of
  `smaqit-deployment`, through slots" — no separate machine-deployment agent.
- **Configuration = a separate `.smaqit/slots.yml`, referenced from `AGENTS.md`** (user, 2026-10-10):
  machine-parseable, project-owned; `AGENTS.md` carries the pointer (outside the managed
  `smaqit:instructions` block, which `installInstructionsFile` never rewrites once present).
- **Unbound slot = default filler from the `provisioning_mode` profile; `none` = skip** (user,
  2026-10-10; refined 2026-10-11): existing app projects keep working unchanged with no configuration.
- **Lands first and re-scopes 114/115/122/123** (user, 2026-10-10): 114 Finding 5 resolved here; 115
  becomes optional default fillers and contracts (the platform repo owns its fillers); 122's
  reachability class becomes slot parameters of `verify`/`tls`/`ci-generate`; 123's "Vault through a
  project-provided script interface" becomes the `credentials` slot contract.
- **Fillers declare their slots** in `metadata.smaqit-slots` (slot → operation); `smaqit slots plan`
  refuses a binding whose skill doesn't exist or doesn't declare the slot.
- **Invariants are slot-independent:** the deployment agent's guard-script MUST, the no-out-of-band
  MUST NOT and credential-references-only apply to every filler; a filler that cannot honour them is
  rejected at plan time, not tolerated.
- **Day-2 runs are single-slot runs** (e.g. `replace`), through the same agent and invariants.
- **No downstream project is named** in this task's shipped artifacts (CONTRIBUTING rule).
- **Revised slot catalogue** (user, 2026-10-11, from `task.plan` discovery): per-environment fixed
  order `platform-prereqs` → `access-credentials` → `platform-onboarding` → `target-credentials` →
  `artifact-build` → `iac-generate` → `ci-generate` → `provision` → `host-bootstrap` →
  `environment-config` → `app-deploy` → `verify` → `tls`; day-2 single-slot operations `replace`,
  `teardown`. Replaces the original 13-slot list, which contradicted current behaviour: k3s
  credentials are two-stage (the kubeconfig exists only after onboarding merges), IaC + CI workflows
  are generated before provisioning (greenfield P4.3), TLS runs after verify (greenfield Phase 6), and
  image/manifest authoring had no slot.
- **`provisioning_mode` becomes the default-profile selector** (user, 2026-10-11): each slot's default
  filler is a function of the mode (e.g. `provision` → `none` for `existing-k3s`); `slots.yml`
  overrides per slot. No more per-mode branching in orchestrator skills.
- **Resolution in the Go CLI** (user, 2026-10-11): `smaqit slots plan --phase=deploy --mode=<m>
  --env=<e> [--path=<deployment_path>] [--slot=<s>] [--json]` parses `.smaqit/slots.yml` strictly
  (yaml.v3), resolves bound skills project-local (`.agents/skills`, `.claude/skills`, `.github/skills`)
  then global (`resolveGlobalDir`), checks `metadata.smaqit-slots`, and emits the plan (filler,
  operation, absolute SKILL.md path, executor `local`/`ci`, source `bound`/`profile`/`none`) or a
  nonzero `SLOTS-*` diagnostic. `smaqit.input-deployment` wraps it.
- **Invocation = read SKILL.md from the resolved path** (user, 2026-10-11): platform-uniform; no
  native Skill tool. The Copilot deployment agent gains a read-file tool (it currently has none).
- **Delivery = parent + children A/B/C, one branch, one PR, release 3.7.0** (user, 2026-10-11): A =
  principle, catalogue, schema, resolver, init skeleton, default fillers declare slots; B = agent +
  input-deployment execute the plan, feature-new Phase 3 routed; C = greenfield Phases 4–6 routed.
- **Principle edits are direct** (2026-10-11): consistent with feature-work practice (`a2efc1f`,
  `160af09`); the smaqit-adk base/extension migration (history 080) was deliberately not done, so
  `smaqit.new-principle` scope detection is not applied here.
- **AGENTS.md pointer is informational** (2026-10-11): the whole template lives inside the managed
  block, so existing projects never receive it; the resolver always reads the fixed path
  `.smaqit/slots.yml`. The template also notes `slots.yml` as a project-owned exception to the
  `.smaqit/` "scaffolding to ignore" bullet.
- **`existing-cicd-pr` refuses generation/provision bindings** (2026-10-11): its profile defaults
  `iac-generate`, `ci-generate`, `provision`, `host-bootstrap` to `none`, and a binding to them under
  that path is refused — preserving the agent's existing "must not regenerate workflows / directly
  provision" prohibition.
- **New `deployment_path: ci-push`** (2026-10-11): greenfield Phase 5's push-to-main + correlated run
  watch + verify + `status: deployed`, owned by the agent; `provision`/`host-bootstrap`/`app-deploy`
  are `ci` executors. Closes the existing gap where the Phase 5 gate says the agent sets
  `status: deployed` but no agent is invoked.
- **Default `teardown` only for `provision`** (2026-10-11): greenfield Phase 4 step 12 currently has no
  `existing-owned` callout, a latent hazard; `existing-owned` defaults to `none`.
- **Fix three invalid-YAML SKILL.md frontmatters** (2026-10-11): `deploy-k3s-app`, `repo-config`,
  `request-k3s-onboarding` — unquoted `provisioning_mode: existing-k3s` inside a plain-scalar
  `description` breaks `yaml.safe_load` / `yaml.v3`.
- **`smaqit uninstall` preserves `.smaqit/slots.yml`** (2026-10-11): it is project-owned configuration;
  `cmdUninstall` currently does `os.RemoveAll(".smaqit")`.

## Implementation Steps

**Child A — Principle, catalogue, schema, resolver (no behaviour change)**

1. **L0 (direct edit)** — `framework/AGENTS.md` Phase Orchestration: replace "Deterministic routing…
   hardcoded per phase" with the slot-sequence principle (fixed order, binding resolution, default
   profile fallback, resolution recorded, invariants bind every filler). `framework/SKILLS.md`: new
   "Filler Skills (Variation Points)" section after "Skill Invocation Context" defining
   `metadata.smaqit-slots`. `framework/SMAQIT.md` "Extensible Through Templates": extend to variation
   points. Record the direct-edit decision in Findings.
2. **Slot catalogue** — the revised order plus day-2 slots, each with purpose, inputs, handoff,
   completion signal, and per-`provisioning_mode` default filler: as a Go data table in
   `installer/slots.go`, and for humans in `docs/wiki/workflows/deployment-slots.md` (linked from
   `docs/wiki/workflows/extending-smaqit.md` and `README.md`).
3. **Schema + template** — `templates/slots.yml.template` (`schemaVersion: 1`, `phase: deploy`, `slots`
   map; each entry `default`/`none` or `{skill, operation, params, environments}`), fully commented so
   an unedited file resolves every slot to its default. Wire through `installer/Makefile` `prepare` and
   a `//go:embed`; `cmdInit` writes `.smaqit/slots.yml` only if absent (not via `copyEmbeddedDir` /
   `detectConflicts`); `cmdUninstall` preserves it with a notice; `cmdValidate` validates it when
   present. `templates/AGENTS.md.template`: pointer line + `.smaqit/` exception note.
4. **Resolver** — `smaqit slots plan` in `installer/slots.go`: strict parsing following
   `validateDesignMetadataKeys` (`installer/design.go`), frontmatter fence reading following
   `parseSpecFrontmatter` (`installer/spec.go`), skill lookup project-local then `resolveGlobalDir`;
   `SLOTS-*` diagnostics (missing skill, undeclared slot, unknown slot, schema violation, refused
   binding for the deployment path). Register in help text and the README command table.
5. **Default fillers declare slots** — add `metadata.smaqit-slots` to vault-loader,
   request-k3s-onboarding, image-build-push-static, cicd-generate, provision-cyso, vm-bootstrap,
   repo-config, deploy-rsync family (3), deploy-k3s-app, deploy-verify, domain-tls; quote the invalid
   frontmatter descriptions in deploy-k3s-app, repo-config, request-k3s-onboarding; bump each changed
   skill's `metadata.version`.
6. **Go tests** — resolver (bound, unbound → profile default, `none`, undeclared → refused, missing
   skill, schema errors, path refusals); skeleton helper (create / preserve / re-init); uninstall
   preserves; every default filler named in the Go table declares its slot in the `skills-shared`
   embed. Smoke test: `.smaqit/slots.yml` exists after init and survives a re-init.

**Child B — Agent executes the slot plan; feature-new routed (depends on A)**

7. `skills/smaqit.input-deployment/SKILL.md` — after resolving `provisioning_mode`, call
   `smaqit slots plan` and output the slot plan; stop on `SLOTS-*`; do not re-invoke when the caller
   already passed a plan (removes feature-new's double invocation).
8. `agents/deployment.md` + `.smaqit/definitions/agents/deployment.frontmatter.yaml` — Execution
   Parameters take `slot_plan`; Steps 5–6 walk it in order (read each filler's SKILL.md at its resolved
   path and follow it; skip `none`; `ci`-executor slots are monitored, not run locally); single-slot
   day-2 entry; slot resolution in the deployment report; guard-script / no-out-of-band /
   credential-references invariants restated as binding every filler; add a read-file tool to the
   Copilot tools list.
9. `skills/smaqit.feature-new/SKILL.md` Phase 3 — its vault-loader/repo-config preflight becomes the
   agent's `access-credentials` / `target-credentials` / `environment-config` slots; the
   `existing-cicd-pr` profile and refusals apply.

**Child C — Greenfield Phases 4–6 routed through the agent (depends on B)**

10. `skills/smaqit.new-greenfield-project/SKILL.md` Phase 4 — one agent call, `env=dev`,
    `deployment_path: standard`, all executors local; per-mode callouts collapse into profile
    defaults; k3s credential ordering fixed (`target-credentials` after `platform-onboarding`);
    `teardown` defaults only for `provision`.
11. Phase 5 — one agent call, `env=prod`, `deployment_path: ci-push`; `provision`/`host-bootstrap`/
    `app-deploy` are `ci` executors.
12. Phase 6 — single-slot `tls` run, conditional on a declared domain; k3s profile `none` (TLS lives in
    `app-deploy`). Phase numbering and task bracketing unchanged.

**Close-out (parent)**

13. **Bench** — `.smaqit/bench/agents/smaqit.deployment/bench.yaml`: fixture project with stub fillers
    that write marker files; cases bound → invoked, unbound → default, `none` → skipped, undeclared →
    agent stops on `SLOTS-UNDECLARED`; treatments `agents/deployment.md` + input-deployment; `codex exec`
    process block; every grader declares its `command.environment`.
14. **Re-scope notes** on tasks 114, 115, 122, 123, 074 refreshed with the revised slot names (122's
    reachability class → `verify` / `tls` / `ci-generate` / `app-deploy` parameters).
15. `smaqit.qa` consistency review (L0 ↔ catalogue ↔ CLI ↔ agent ↔ skills), CHANGELOG, release 3.7.0.

## Known Issues Triage
**Triaged:** 2026-10-11
**Tools searched:** Claude Code (`anthropics/claude-code`), Codex CLI (`openai/codex`), GitHub Copilot (`github/copilot-cli`), gopkg.in/yaml.v3 (`go-yaml/yaml`), smaqit-adk (`ruifrvaz/smaqit-adk`)
**Result:** Advisory

No issue findings could be retrieved: every search failed (see Search Warnings), so this is not a Clear result.

### Omitted Tools
- Go — standard library only (`flag`, `embed`); excluded after relevance prioritization and the five-repository limit

### Search Warnings
- `anthropics/claude-code` open/closed searches — HTTP 403, unauthenticated GitHub API rate limit exceeded
- `openai/codex` open/closed searches — HTTP 403, unauthenticated GitHub API rate limit exceeded
- `github/copilot-cli` open/closed searches — HTTP 403, unauthenticated GitHub API rate limit exceeded
- `go-yaml/yaml` open/closed searches — HTTP 403, unauthenticated GitHub API rate limit exceeded
- `ruifrvaz/smaqit-adk` open/closed searches — HTTP 403, unauthenticated GitHub API rate limit exceeded

## Acceptance Criteria

- [ ] `framework/AGENTS.md` Phase Orchestration states the slot-sequence principle (binding resolution, default profile fallback, recorded resolution, invariants for every filler); `framework/SKILLS.md` defines `metadata.smaqit-slots`; `framework/SMAQIT.md` mentions variation points; the direct-edit decision is recorded in Findings
- [ ] The revised deploy-phase slot catalogue (per-environment order + day-2 `replace`/`teardown`) documents every slot's purpose, inputs, handoff, completion signal and per-`provisioning_mode` default filler, as a Go table and a wiki page; a Go test asserts every default filler declares its slot in `metadata.smaqit-slots`
- [ ] `.smaqit/slots.yml` schema/template exists; `smaqit init` creates a commented skeleton when absent and preserves it on re-init/update; `smaqit uninstall` preserves it; `smaqit validate` checks it when present; the AGENTS.md template points to it
- [ ] `smaqit slots plan` resolves a deterministic plan (project-local then global skill lookup; `default`/`none` honoured) and refuses with `SLOTS-*` diagnostics on missing skill, undeclared slot, unknown slot, schema violation, or a binding refused by the deployment path; every `skills/*/SKILL.md` frontmatter parses as YAML
- [ ] `smaqit.input-deployment` outputs the slot plan via the CLI and is invoked once per deployment (not re-invoked when a plan is passed in)
- [ ] `agents/deployment.md` executes the slot plan in Steps 5–6 (reading each filler's SKILL.md from its resolved path; `ci` executors monitored, not run), supports a single-slot day-2 run, records the resolution in the deployment report, and keeps the guard/out-of-band/credential invariants for every filler; the Copilot deployment agent has a read-file tool
- [ ] `feature-new` Phase 3 routes infrastructure steps through the slot plan; under `existing-cicd-pr`, bindings to `iac-generate`/`ci-generate`/`provision`/`host-bootstrap` are refused
- [ ] `new-greenfield-project` Phases 4–6 route through the agent (Phase 4 `standard`, Phase 5 `ci-push`, Phase 6 single-slot `tls`); k3s `target-credentials` runs after `platform-onboarding`; default `teardown` applies only to `provision`; with no `.smaqit/slots.yml`, the resolved sequence per mode matches current behaviour
- [ ] `make -C installer test` and `make -C installer smoke-test` pass; bench cases (bound, unbound, none, undeclared) pass
- [ ] Tasks 114, 115, 122, 123 and 074 carry refreshed re-scope notes; no downstream project is named in shipped artifacts
- [ ] `smaqit.qa` review passes; released as 3.7.0

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
| `framework/AGENTS.md`, `framework/SKILLS.md`, `framework/SMAQIT.md` | Modify (L0) |
| `templates/slots.yml.template`, `templates/AGENTS.md.template` | Create / Modify (L1) |
| `installer/slots.go`, `installer/slots_test.go` | Create (catalogue table, `smaqit slots plan`, tests) |
| `installer/main.go`, `installer/main_test.go`, `installer/Makefile` | Modify (embed, init skeleton, uninstall preserve, validate hook, help) |
| `scripts/smoke-test-installer.sh` | Modify (skeleton + re-init assertions) |
| `skills/smaqit.input-deployment/SKILL.md` | Modify |
| `agents/deployment.md`, `.smaqit/definitions/agents/deployment.frontmatter.yaml` | Modify |
| `skills/smaqit.new-greenfield-project/SKILL.md`, `skills/smaqit.feature-new/SKILL.md` | Modify |
| 13 default-filler `skills/*/SKILL.md` (vault-loader, request-k3s-onboarding, image-build-push-static, cicd-generate, provision-cyso, vm-bootstrap, repo-config, deploy-rsync ×3, deploy-k3s-app, deploy-verify, domain-tls) | Modify (`metadata.smaqit-slots`, version bump; YAML fix in 3) |
| `docs/wiki/workflows/deployment-slots.md`, `docs/wiki/workflows/extending-smaqit.md`, `README.md` | Create / Modify |
| `.smaqit/bench/agents/smaqit.deployment/` | Create |
| `CHANGELOG.md` | Modify |

## Notes

- Insertion points mapped 2026-10-10: Phase Orchestration principle (`framework/AGENTS.md`, the
  "Deterministic routing" line); `input-deployment` mode resolution; deployment agent Execution
  Parameters and Steps 5–6; greenfield Step 6's family lookup is the closest existing slot (deploy
  only, prefix-matched, stack-driven) and is subsumed by the `app-deploy` slot.
- Several canonical skills read the project's AGENTS.md only for identity facts (project slug,
  owner/repo); none reads it for bindings.
- A downstream platform repository is waiting on this release to bind its own fillers.
