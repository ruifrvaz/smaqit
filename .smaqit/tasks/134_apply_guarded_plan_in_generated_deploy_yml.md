---
status: Not Started
created: "2026-10-12"
---

# Apply the Guarded Plan in the Generated deploy.yml (full Mode)

## Description

Follow-up from task 115 (released in v3.11.0). `plan-guard.sh` gained `--out <planfile>`, which
keeps the exact plan the guard inspected so the caller applies that plan instead of planning again.
`smaqit.infrastructure-provision-cyso` now uses it locally (`plan-guard.sh . --out tfplan`, then
`terraform apply tfplan`), but the CI template `smaqit.infrastructure-cicd-generate` generates for
`full` mode still runs the guard and then a bare `terraform apply -auto-approve`
(`assets/deploy.yml.full.template`, provision job). That second command plans again, so CI can
apply a plan the guard never saw — for example, if the state or a remote resource changed between
the two steps. Priority: Medium.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Terraform, GitHub Actions, Bash
**Platforms/Environments:** GitHub-hosted runners (ubuntu-latest)
**Features/Integrations:** terraform saved plan files, terraform apply, plan-guard.sh --out, smaqit cicd-generate templates
**Versions/Constraints:** Terraform 1.14.0 (template pin), smaqit 3.11.0, plan-guard.sh --out resolves relative paths against the caller's directory

## Design Decisions

- **Only the `full` template's provision job changes.** `provision.yml.template` only plans (its
  guard verdict is posted to the PR) and never applies; the `deploy-only` and `k3s` templates run
  no Terraform.
- **The plan file lives in the Terraform directory:** `--out deployment/terraform/tfplan` (the
  guard resolves it against the repository root, where the step runs), and the apply step runs
  `terraform apply tfplan` with `working-directory: deployment/terraform`. A saved plan applies
  without a confirmation prompt, so `-auto-approve` is no longer needed.
- **No version skew:** cicd-generate vendors `plan-guard.sh` together with the templates on every
  run, so a regenerated `deploy.yml` always ships with a guard that supports `--out`.
- **No replacement input** is added to the push-triggered `deploy.yml`; deliberate replacement
  stays a gated, operator-dispatched operation (documented in cicd-generate's gotchas).

## Implementation Steps

1. `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.full.template`, provision job:
   run the guard with `--out deployment/terraform/tfplan` and apply with `terraform apply tfplan`;
   update the comment above the guard to say the apply uses the guarded plan.
2. `skills/smaqit.infrastructure-cicd-generate/SKILL.md`: update the gotcha that says never to
   remove the guard steps (the apply now consumes the guard's plan, so the two steps are coupled),
   note the change in the vendoring step if relevant, and bump `metadata.version`.
3. Add a test that pins the template's guard and apply lines (the template is embedded through
   `skills-shared`), so a regression back to a bare apply fails `make test`.
4. Run `make -C installer prepare test smoke-test`.
5. CHANGELOG via `task.complete`.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] The generated `full`-mode `deploy.yml` runs `plan-guard.sh deployment/terraform --out deployment/terraform/tfplan` and applies with `terraform apply tfplan`; no `terraform apply -auto-approve` remains in any cicd-generate template.
- [ ] `provision.yml.template`, `deploy.yml.deploy-only.template`, and `deploy.yml.k3s.template` are unchanged.
- [ ] cicd-generate's SKILL.md describes the guard and apply steps as coupled, and its version is bumped.
- [ ] A test fails if the `full` template's provision job applies anything other than the guarded plan file.
- [ ] `make -C installer test` and `make -C installer smoke-test` pass; `CHANGELOG.md` is updated.

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
| `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.full.template` | Modify |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md` | Modify |
| `installer/*_test.go` | Modify (template assertion) |
| `CHANGELOG.md` | Modify |

## Notes

- Origin: task 115 Findings (follow-up identified), 2026-10-12.
