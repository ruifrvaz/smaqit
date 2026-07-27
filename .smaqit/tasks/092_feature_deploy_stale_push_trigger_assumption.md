# `smaqit.feature-deploy` — Stale "Push to Main Triggers Deploy" Assumption

**Status:** Abandoned
**Created:** 2026-07-26

**Abandoned:** 2026-07-27
**Reason:** Superseded by Task 093, which removes the redundant `smaqit.feature-deploy` skill and consolidates its necessary safeguards into `smaqit.feature-new`.

## Description

`smaqit.feature-deploy`'s Phase 2 (Deploy via CI/CD) documents a single, unconditional deploy trigger mechanism:

> Push to main: `git push origin main` — triggers the existing `deploy.yml` workflow

This assumes every project's `deploy.yml` (generated during its original greenfield run) triggers on `push` to `main`. That assumption does not hold universally. Discovered on `iodis-crm-poc` (2026-07-26, during a `smaqit.feature-new` cycle for its SendGrid email integration feature): that project's `deploy.yml` trigger is `workflow_dispatch`-only — a direct push to `main` fires nothing. Its actual automatic-deploy path is a second workflow, `post-merge-deploy.yml`, which fires on `pull_request: closed` gated by `github.event.pull_request.merged == true && contains(github.event.pull_request.body, 'smaqit:deploy')`, and dispatches `deploy.yml` via the GitHub API. That PR-gated pattern exists specifically so a human reviews and merges the deploy-triggering change — a direct push bypasses that review entirely, and on `iodis-crm-poc` it would silently deploy nothing at all (worse than bypassing review: it fails invisibly, with no error, since nothing observes the push).

The user had to explicitly catch this by asking "confirm if the deploy skill does that (the PR)" before the deploy task proceeded — `smaqit.feature-deploy` was not the source of that catch, it was the source of the wrong assumption. The workaround (PR + `smaqit:deploy` marker + merge) was applied directly on `iodis-crm-poc`, documented inline in that project's own task file, and never fed back to this skill — this task closes that loop.

## Design Decisions

TBD — to be confirmed during assessment/planning (`smaqit.task-plan 092`). Open questions:
- Detection mechanism: should the skill read the target project's actual `deploy.yml` (and any sibling `post-merge-*.yml`/`*-gate.yml` workflows) to determine the real trigger mechanism before choosing how to initiate a deploy, rather than assuming `push` unconditionally?
- If a PR-gated pattern is detected (a workflow that dispatches the deploy workflow conditional on a merged PR with a specific body marker), should the skill open a PR with that marker and wait for user merge, analogous to how `smaqit.feature-new`'s Phase 3 already documents "Push (or open a PR, per this project's CI/CD convention)" — i.e., should `feature-deploy` adopt the same convention-detection language `feature-new` already has, rather than `feature-deploy`'s current unconditional "push to main"?
- Should this also cover the inverse case — a project whose `deploy.yml` *does* trigger on push, where opening a PR-and-wait would be unnecessarily conservative — i.e., the fix needs to handle both directions correctly, not just flip the default.

## Implementation Steps

TBD — to be planned via `smaqit.task-plan 092` before implementation. Expected to touch: `skills/smaqit.feature-deploy/SKILL.md`'s Phase 2 step 3 (the `git push origin main` line) and its Gotchas section (add this as a named gotcha, following the pattern of its existing "provisioning_mode default-override lives here" and "CI/CD must already exist" gotchas).

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `smaqit.feature-deploy`'s Phase 2 no longer unconditionally assumes `git push origin main` triggers a deploy — it first determines the target project's actual deploy-trigger mechanism from its committed workflow files
- [ ] When a PR-gated deploy pattern is detected (a merge-triggered workflow conditional on a body marker, e.g. `smaqit:deploy`), the skill opens a PR with that marker and waits for user merge, rather than pushing directly to `main`
- [ ] When a direct-push-triggered `deploy.yml` is detected, the skill's existing direct-push behavior is preserved (no regression for projects that actually work this way)
- [ ] A new Gotcha is added documenting this failure mode, referencing `iodis-crm-poc` as the concrete case that surfaced it

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
| `skills/smaqit.feature-deploy/SKILL.md` | Modify (Phase 2 step 3, Gotchas) |

## Notes

Discovered on `iodis-crm-poc` during task 010 (Deployment phase of a `smaqit.feature-new` cycle for its SendGrid email integration feature). That project applied the correct workaround directly (PR + `smaqit:deploy` marker), documented only in its own task file — this task is the feedback loop back into the framework, following the same precedent as task 089 (a different `smaqit.feature-new` gap found and tracked from the same downstream project). Task 091 (this skill's original build task, completed 2026-07-23) is where the stale assumption was first introduced — not a duplicate of that task, since 091 already shipped and this is a correction to what it shipped.
