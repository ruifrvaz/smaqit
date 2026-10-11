---
status: Not Started
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
- **Catalogue as code + docs:** the Go table in `installer/slots.go` is the machine source; `docs/wiki/workflows/deployment-slots.md` is the human explanation; a Go test guards that every default filler named in the table declares its slot.

## Implementation Steps

1. **L0 (direct edit)** — `framework/AGENTS.md` Phase Orchestration: replace "Deterministic routing… hardcoded per phase" with the slot-sequence principle. `framework/SKILLS.md`: "Filler Skills (Variation Points)" section after "Skill Invocation Context" defining `metadata.smaqit-slots`. `framework/SMAQIT.md` "Extensible Through Templates": extend to variation points.
2. **Slot catalogue** — per-environment order `platform-prereqs` → `access-credentials` → `platform-onboarding` → `target-credentials` → `artifact-build` → `iac-generate` → `ci-generate` → `provision` → `host-bootstrap` → `environment-config` → `app-deploy` → `verify` → `tls`; day-2 `replace`, `teardown`. Each slot: purpose, inputs, handoff, completion signal, per-`provisioning_mode` default filler, default executor per `deployment_path`. Go table in `installer/slots.go`; wiki page `docs/wiki/workflows/deployment-slots.md` linked from `docs/wiki/workflows/extending-smaqit.md` and `README.md`.
3. **Schema + template** — `templates/slots.yml.template` (`schemaVersion: 1`, `phase: deploy`, `slots` map; entries `default`/`none` or `{skill, operation, params, environments}`), fully commented. `installer/Makefile` `prepare` + `//go:embed`; `cmdInit` writes `.smaqit/slots.yml` only if absent (not via `copyEmbeddedDir`/`detectConflicts`); `cmdUninstall` preserves it with a notice; `cmdValidate` validates it when present. `templates/AGENTS.md.template`: pointer line + `.smaqit/` exception note.
4. **Resolver** — `smaqit slots plan --phase=deploy --mode=<m> --env=<e> [--path=<deployment_path>] [--slot=<s>] [--json]`: strict parsing (`validateDesignMetadataKeys` pattern), frontmatter fence reading (`parseSpecFrontmatter` pattern), skill lookup project-local (`.agents/skills`, `.claude/skills`, `.github/skills`) then `resolveGlobalDir`; emits filler, operation, absolute SKILL.md path, executor, source; `SLOTS-*` diagnostics. Help text + README command table.
5. **Default fillers** — `metadata.smaqit-slots` on vault-loader, request-k3s-onboarding, image-build-push-static, cicd-generate, provision-cyso, vm-bootstrap, repo-config, deploy-rsync ×3, deploy-k3s-app, deploy-verify, domain-tls; quote invalid descriptions in deploy-k3s-app, repo-config, request-k3s-onboarding; bump `metadata.version`.
6. **Tests** — resolver cases (bound, profile default, `none`, undeclared, missing skill, unknown slot, schema errors, path refusals); skeleton helper (create / preserve / re-init); uninstall preserves; default-filler declaration guard over the `skills-shared` embed. Smoke test asserts the skeleton and a preserving re-init.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `framework/AGENTS.md`, `framework/SKILLS.md`, `framework/SMAQIT.md` carry the slot-sequence principle, `metadata.smaqit-slots`, and variation points respectively
- [ ] The revised catalogue exists as a Go table and a wiki page with every slot's purpose, inputs, handoff, completion signal, and per-mode default filler; a Go test asserts every default filler declares its slot
- [ ] `smaqit init` creates `.smaqit/slots.yml` when absent and preserves it on re-init/update; `smaqit uninstall` preserves it; `smaqit validate` checks it; the AGENTS.md template points to it
- [ ] `smaqit slots plan` resolves deterministically and refuses with `SLOTS-*` diagnostics; with no `slots.yml`, each mode's resolved sequence matches current behaviour
- [ ] Every `skills/*/SKILL.md` frontmatter parses as YAML; `make -C installer test` and `make -C installer smoke-test` pass

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
| `framework/AGENTS.md`, `framework/SKILLS.md`, `framework/SMAQIT.md` | Modify |
| `installer/slots.go`, `installer/slots_test.go` | Create |
| `installer/main.go`, `installer/main_test.go`, `installer/Makefile` | Modify |
| `scripts/smoke-test-installer.sh` | Modify |
| `templates/slots.yml.template`, `templates/AGENTS.md.template` | Create / Modify |
| 13 default-filler `skills/*/SKILL.md` | Modify |
| `docs/wiki/workflows/deployment-slots.md`, `docs/wiki/workflows/extending-smaqit.md`, `README.md` | Create / Modify |

## Notes

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
