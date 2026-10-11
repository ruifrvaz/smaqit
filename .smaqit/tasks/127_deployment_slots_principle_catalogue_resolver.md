---
status: Completed
completed: "2026-10-11"
mode: Assisted
started: "2026-10-11"
created: "2026-10-11"
parent: "125"
---

# Deployment Slots A: Principle, Catalogue, Schema, and `smaqit slots plan` Resolver

## Description

Child A of task 125. Lays the foundation for slot-driven deployment orchestration without changing any
runtime behaviour: the L0 principle, the revised deploy-phase slot catalogue (as a Go data table and a
wiki page), the project-owned `.smaqit/slots.yml` schema and its create-if-absent init skeleton, the
deterministic `smaqit slots plan` resolver, and `metadata.smaqit-slots` declarations on every default
filler skill. Children B (128) and C (129) consume this; see task 125's Design Decisions for the full
rationale.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Go, gopkg.in/yaml.v3, Agent Skills
**Platforms/Environments:** Claude Code, GitHub Copilot, Codex CLI
**Features/Integrations:** YAML configuration, SKILL.md frontmatter metadata, installer init scaffolding
**Versions/Constraints:** smaqit 3.6.1, `smaqit update` re-runs init (skeleton must be create-if-absent), installer skill-count test stays at 30

## Design Decisions

- Inherits all of task 125's Design Decisions (revised catalogue, `provisioning_mode` as default-profile selector, CLI resolution, direct principle edits, informational AGENTS.md pointer, `existing-cicd-pr` refusals, `teardown` default only for `provision`, uninstall preserves `slots.yml`).
- **`metadata.smaqit-slots` is a single string** (2026-10-11, from triage of agentskills#514): the Agent Skills spec defines `metadata` as string→string, so slots are encoded as comma-separated `slot=operation` pairs, e.g. `smaqit-slots: "app-deploy=deploy, verify=verify"`, not a nested map. The resolver parses this format and rejects malformed pairs or unknown slot names.
- **Catalogue as code + docs:** the Go table in `installer/slots.go` is the machine source; `docs/wiki/workflows/deployment-slots.md` is the human explanation; a Go test guards that every default filler named in the table declares its slot.

## Implementation Steps

1. **L0 (direct edit)** — `framework/AGENTS.md` Phase Orchestration: replace "Deterministic routing… hardcoded per phase" with the slot-sequence principle. `framework/SKILLS.md`: "Filler Skills (Variation Points)" section after "Skill Invocation Context" defining `metadata.smaqit-slots`. `framework/SMAQIT.md` "Extensible Through Templates": extend to variation points.
2. **Slot catalogue** — per-environment order `platform-prereqs` → `access-credentials` → `platform-onboarding` → `target-credentials` → `artifact-build` → `iac-generate` → `ci-generate` → `provision` → `host-bootstrap` → `environment-config` → `app-deploy` → `verify` → `tls`; day-2 `replace`, `teardown`. Each slot: purpose, inputs, handoff, completion signal, per-`provisioning_mode` default filler, default executor per `deployment_path`. Go table in `installer/slots.go`; wiki page `docs/wiki/workflows/deployment-slots.md` linked from `docs/wiki/workflows/extending-smaqit.md` and `README.md`.
3. **Schema + template** — `templates/slots.yml.template` (`schemaVersion: 1`, `phase: deploy`, `slots` map; entries `default`/`none` or `{skill, operation, params, environments}`), fully commented. `installer/Makefile` `prepare` + `//go:embed`; `cmdInit` writes `.smaqit/slots.yml` only if absent (not via `copyEmbeddedDir`/`detectConflicts`); `cmdUninstall` preserves it with a notice; `cmdValidate` validates it when present. `templates/AGENTS.md.template`: pointer line + `.smaqit/` exception note.
4. **Resolver** — `smaqit slots plan --phase=deploy --mode=<m> --env=<e> [--path=<deployment_path>] [--slot=<s>] [--json]`: strict parsing (`validateDesignMetadataKeys` pattern), frontmatter fence reading (`parseSpecFrontmatter` pattern), skill lookup project-local (`.agents/skills`, `.claude/skills`, `.github/skills`) then `resolveGlobalDir`; emits filler, operation, absolute SKILL.md path, executor, source; `SLOTS-*` diagnostics. Help text + README command table.
5. **Default fillers** — `metadata.smaqit-slots` on vault-loader, request-k3s-onboarding, image-build-push-static, cicd-generate, provision-cyso, vm-bootstrap, repo-config, deploy-rsync ×3, deploy-k3s-app, deploy-verify, domain-tls; quote invalid descriptions in deploy-k3s-app, repo-config, request-k3s-onboarding; bump `metadata.version`.
6. **Tests** — resolver cases (bound, profile default, `none`, undeclared, missing skill, unknown slot, schema errors, path refusals); skeleton helper (create / preserve / re-init); uninstall preserves; default-filler declaration guard over the `skills-shared` embed. Smoke test asserts the skeleton and a preserving re-init.

## Known Issues Triage
**Triaged:** 2026-10-11
**Tools searched:** gopkg.in/yaml.v3 (`go-yaml/yaml`), Agent Skills (`agentskills/agentskills`)
**Result:** Advisory

### Advisory Issues
- [#514 metadata is specified as string values, but a shipping runtime stores nested maps and booleans there](https://github.com/agentskills/agentskills/issues/514) — `agentskills/agentskills` — opened 2026-08 — no labels. The spec defines `metadata` as a map from string keys to **string values**; a nested `smaqit-slots` map would violate it. Drove the string-encoding Design Decision below.
- [#769 Can't Unmarshal to `map[string]*yaml.Node`](https://github.com/go-yaml/yaml/issues/769) — `go-yaml/yaml` — opened 2021-07-29 — no labels. Walk `yaml.Node` mappings directly instead of decoding into `map[string]*yaml.Node`.

### Historical (Closed)
- [#784 parse error when strings containing colon-space](https://github.com/go-yaml/yaml/issues/784) — `go-yaml/yaml` — closed 2024-09-23. Expected YAML behaviour; the root cause of the three invalid SKILL.md descriptions this task quotes.
- [#562 DOS-style newlines not allowed in v3](https://github.com/go-yaml/yaml/issues/562) — `go-yaml/yaml` — closed 2020-01-21. Normalize CRLF before parsing (as `parseDesign` already does).

### Omitted Tools
- Go — standard library only (`flag`, `embed`)

## Acceptance Criteria

- [x] `framework/AGENTS.md`, `framework/SKILLS.md`, `framework/SMAQIT.md` carry the slot-sequence principle, `metadata.smaqit-slots`, and variation points respectively
- [x] The revised catalogue exists as a Go table and a wiki page with every slot's purpose, inputs, handoff, completion signal, and per-mode default filler; a Go test asserts every default filler declares its slot
- [x] `smaqit init` creates `.smaqit/slots.yml` when absent and preserves it on re-init/update; `smaqit uninstall` preserves it; `smaqit validate` checks it; the AGENTS.md template points to it
- [x] `smaqit slots plan` resolves deterministically and refuses with `SLOTS-*` diagnostics; with no `slots.yml`, each mode's resolved sequence matches current behaviour
- [x] Every `skills/*/SKILL.md` frontmatter parses as YAML; `make -C installer test` and `make -C installer smoke-test` pass

## Findings

**Implementation approach:**
- `installer/slots.go` holds the catalogue (`deploySlots`, per-mode defaults, `deploymentPathRules`), the strict `slots.yml` parser, skill lookup (project-local, then global shared, then Claude), and `smaqit slots plan|validate`; errors are aggregated, not fail-fast
- Default fillers are skills, stack-selected families (`deploy-rsync*`, `deploy-k3s-app*`, `image-build-push*`, listed as candidates), or agent builtins (`generate-iac`, `iac-destroy`)
- `cmdInit` writes the skeleton create-if-absent; `cmdUninstall` restores `slots.yml` after removing `.smaqit/`; `cmdValidate` checks it when present
- Principle added as an "Implementation Variation Points (Slots)" block in `framework/AGENTS.md`, a "Filler Skills" section in `framework/SKILLS.md`, and a paragraph under "Extensible Through Templates"; edited directly, not via `smaqit.new-principle`
- 13 Go tests, including a per-mode equivalence test against current greenfield wiring and a guard that every default filler ships with its slot declaration; smoke test covers skeleton, re-init preservation, and uninstall preservation

**Decisions made:**
- `metadata.smaqit-slots` is one string of `slot=operation` pairs, because the Agent Skills spec limits `metadata` values to strings (agentskills#514)
- Operation names: vault-loader `access`/`target`, onboarding `request`, image-build `build`, cicd-generate `generate`, provision-cyso `apply`, vm-bootstrap `bootstrap`, repo-config `sync`, deploy families `deploy`, deploy-verify `verify`, domain-tls `issue`
- Added opt-in `CREDENTIAL_SCOPE=access` to the k3s branch of `load-credentials.sh` so access credentials load before onboarding issues a kubeconfig; unset keeps single-pass behaviour
- `standard` path defaults `environment-config` and `tls` to none (greenfield dev sweep runs neither); a single-slot run ignores a path's default-none but never its refusals
- `target-credentials` defaults to none for `provision`/`existing-owned`, matching the current single vault-loader step
- A refused binding is reported as refused only; its skill is not resolved

**Blockers encountered:**
- Issue triage for parent 125 hit the unauthenticated GitHub search rate limit (403); recorded as warnings
- Three SKILL.md descriptions were invalid YAML (unquoted `provisioning_mode: existing-k3s`); quoted as part of this task

**Follow-up identified:**
- Child 128 must make the deployment agent and `smaqit.input-deployment` consume `smaqit slots plan` and document filler operations it relies on
- The triage helper has no token support, so triage degrades to warnings under rate limiting

## Files to Create / Modify

| File | Action |
|------|--------|
| `framework/AGENTS.md`, `framework/SKILLS.md`, `framework/SMAQIT.md` | Modify |
| `installer/slots.go`, `installer/slots_test.go` | Create |
| `installer/main.go`, `installer/main_test.go`, `installer/Makefile` | Modify |
| `scripts/smoke-test-installer.sh` | Modify |
| `templates/slots.yml.template`, `templates/AGENTS.md.template` | Create / Modify |
| 13 default-filler `skills/*/SKILL.md` | Modify |
| `docs/wiki/workflows/deployment-slots.md`, `docs/wiki/workflows/extending-smaqit.md`, `README.md` | Create / Modify |

## Notes

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
