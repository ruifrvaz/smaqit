---
status: Not Started
created: "2026-10-23"
---

# VM Modes: CI-Only Deploys (Close the Workflow Gaps, Then Enforce the Task 144 Rule)

## Description

Task 144 made "a deploy to a target runs only in the CI/CD pipeline; bootstrap and configuration run
locally" a framework rule and enforced it for `existing-k3s` in the slot resolver (`ciOnlyModes`,
`SLOTS-LOCAL-DEPLOY-REFUSED`, the `ci-bootstrap` path). The four VM modes (`provision`,
`existing-owned`, `existing-shared`, `existing-unmanaged`) were left on the local `standard` path as a
declared transitional exception, because their generated workflows can't yet do what project-new's
Phase 4 dev sweep does from the operator's machine.

This task closes those workflow gaps, then adds the VM modes to `ciOnlyModes` so they follow the same
rule: project-new Phase 4 deploys a VM dev/test environment through `ci-bootstrap` (or an equivalent
path that also provisions), and no target-acting slot runs locally.

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub Actions, GitHub CLI, Terraform, OpenStack, Docker Compose, nginx, Certbot, Go
**Platforms/Environments:** Cyso Cloud, Ubuntu 24.04
**Features/Integrations:** multi-environment VM deploy.yml, GitHub Environments, nginx vhost in CI, docker-compose and .env in CI, vm-bootstrap workflow, TLS workflow, Python stack templates, first terraform apply in CI, plan-guard, day-two teardown and replace dispatches, deploy-verify from CI evidence
**Versions/Constraints:** builds on task 144 (ci-bootstrap, ciOnlyModes, ci-evidence), no backwards compatibility, MINOR with action-required note, workflow_dispatch requires the workflow on the default branch, every Terraform apply through plan-guard

## Design Decisions

TBD — to be confirmed during assessment. Starting points carried from task 144's planning (2026-10-20):
- **Same rule, same seam:** the ci-monitor contract (trigger, select the run by commit, wait, read
  verify evidence) documented in task 144 applies unchanged; GitHub Actions through `gh`.
- **No break-glass path**, as for existing-k3s.

## Implementation Steps

1. Multi-environment VM `deploy.yml`: an `environment` input, and per-environment GitHub Environments
   instead of the repository-level `VM_HOST` variable.
2. Write the nginx vhost (`write-vhost.sh`), `docker-compose.yml`, and `.env` in CI rather than from
   the operator's machine.
3. `vm-bootstrap` and certbot/TLS (`smaqit.infrastructure-domain-tls`) as gated workflows.
4. Workflow templates for the Python stacks (FastAPI + Next.js, Tornado), matching their rsync skills.
5. App SSH key bootstrap (`bootstrap-app-to-machine.sh`) without a local SSH step.
6. The first `terraform apply` in CI through `plan-guard.sh`, resolving the base-key vs app-key
   keypair mismatch between `terraform-env.sh` and the CI `VM_SSH_PUBLIC_KEY`.
7. Day-two `teardown` and `replace` as gated `workflow_dispatch` runs.
8. VM verify from CI evidence (an HTTP verify step recorded in the deploy run, read like
   `ci-evidence`), or keep the local HTTP `verify` if it doesn't act on the target.
9. Add the VM modes to `ciOnlyModes`; extend `ci-bootstrap` (or add a path) for modes that provision;
   update `smaqit.input-deployment`'s path defaults, project-new Phase 4, the Deployment agent, the
   deployment-slots wiki, the compendium, and the README.
10. Tests and deliberate breaks, mirroring task 144's; CHANGELOG with an action-required note.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] The generated VM workflows deploy any declared environment (dev/test and prod), each through its own GitHub Environment
- [ ] Host bootstrap, the nginx vhost, `docker-compose.yml`/`.env`, TLS, the first provision, and day-two teardown/replace all run in CI, every Terraform apply through `plan-guard.sh`
- [ ] The Python stacks have workflow templates equivalent to their rsync skills
- [ ] VM modes are CI-only in the resolver: a local target-acting slot is refused with `SLOTS-LOCAL-DEPLOY-REFUSED`, and project-new Phase 4 deploys VM targets through CI
- [ ] Docs (Deployment agent, input-deployment, project-new, wiki, compendium, README) no longer describe a transitional local path; tests and deliberate breaks pass; released with an action-required note

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
| installer/slots.go, installer/slots_test.go | Modify |
| skills/smaqit.infrastructure-cicd-generate/assets/*.template, scripts/ | Modify / Create |
| skills/smaqit.infrastructure-deploy-rsync*/SKILL.md | Modify |
| skills/smaqit.infrastructure-vm-bootstrap/SKILL.md, skills/smaqit.infrastructure-domain-tls/SKILL.md | Modify |
| skills/smaqit.infrastructure-provision-cyso/SKILL.md, skills/smaqit.infrastructure-repo-config/SKILL.md | Modify |
| skills/smaqit.input-deployment/SKILL.md, skills/smaqit.project-new/SKILL.md, agents/deployment.md | Modify |
| docs/wiki/workflows/deployment-slots.md, .smaqit/compendium.md, README.md, CHANGELOG.md | Modify |

## Notes

- Split from task 144 (its Phase D step 12), which listed these gaps during `task.plan` on 2026-10-20.
- Start after task 144 merges.
