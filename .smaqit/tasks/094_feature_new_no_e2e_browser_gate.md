# `smaqit.feature-new` — No Mandatory Browser/E2E Gate for Frontend-Touching Features

**Status:** Not Started
**Created:** 2026-07-27

## Description

`smaqit.feature-new`'s Phase 2 (Development) gate reads:

> **Gate:** Build passes. All this feature's acceptance criteria met. Development agent sets touched specs to `status: implemented`.

This ties Development's completion to no specific testing methodology — a feature that changes rendered frontend output (templates, static assets, routing/handler code affecting what a browser actually shows) can pass this gate on HTTP-level checks alone (status codes, redirect targets, response-body substring matches), without ever rendering a page, executing its JavaScript, or clicking through it.

Discovered on a downstream project (2026-07-27, task 016/018 — a `smaqit.feature-new` cycle removing that project's A/B frontend-variant system, keeping a single "Lightning" UI). Development shipped a PR backed only by a curl-driven smoke script and a Python `http.client`-based E2E test suite (86/86 passing) — both verify HTTP semantics, never real rendering. The user caught this after the deployment PR was already open, asking directly whether local Playwright/browser testing had been done at all. It hadn't. The gap was closed retroactively with a real interactive Playwright session (login, dashboard, sidebar navigation, an order-detail action button, a full form submission) before the PR merged, catching nothing broken in that case — but the catch depended entirely on the user noticing, not on the workflow requiring it.

This is not a missing-capability gap: the tooling already exists in this ecosystem (Playwright MCP browser tools, the `smaqit.test-e2e-playwright` skill, `smaqit.test-create`'s own "live-service E2E validation where the task touches live services" language). It's a missing *requirement* — nothing in Feature New's own Phase 2 gate, or in `smaqit.task-create`'s task-file template, obligates a browser pass for a feature that demonstrably changes what renders in a browser.

## Design Decisions

TBD — to be confirmed during assessment/planning (`smaqit.task-plan 094`). Open questions:

- **Detection mechanism:** should the gate inspect the *actual Development diff* for frontend-rendering-affecting paths (templates, static assets, routing/handler code that changes rendered output), rather than relying on the task file's own predicted `Files to Create/Modify` (which can be incomplete or wrong at planning time, as nearly happened on the downstream project that surfaced this — the task's own step list didn't call out a browser pass until the user asked)?
- **Verification form:** should the gate require a real, interactive Playwright session (as was performed ad hoc on the downstream project that surfaced this), or push toward a *committed, automatable* E2E test file that survives past this one Development cycle and runs in CI going forward — the latter is more durable but a bigger lift; worth deciding which this gate should actually mandate.
- **Overlap with existing skills:** is this genuinely a Phase-2-gate gap, or is `smaqit.test-e2e-playwright`/`smaqit.test-create` already supposed to cover this and simply isn't being invoked at the right point in Feature New's own flow? If the latter, the fix may be "invoke the existing skill at Phase 2's gate," not "build new detection logic."
- **Scope relative to Task 093:** Task 093 ("Consolidate Post-MVP Feature Deployment in `smaqit.feature-new`") is actively rewriting Phase 2's gate mechanics and surrounding orchestration in the same file, in progress as of this task's creation. This task must be planned and implemented against Phase 2's shape *after* Task 093 lands, not its current wording — the two should not be worked concurrently against the same section.

## Implementation Steps

TBD — to be planned via `smaqit.task-plan 094` before implementation, after Task 093 completes. Expected to touch `skills/smaqit.feature-new/SKILL.md`'s Phase 2 section (exact location depends on how Task 093 restructures it) and possibly `smaqit.task-create`'s task-file template if the fix includes a template-level prompt/checklist item.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `smaqit.feature-new`'s Development-phase gate detects when a feature's actual changes touch frontend-rendering code (templates, static assets, routing/handler code affecting rendered output) and requires a browser-based (Playwright) verification pass before the phase can close — not satisfied by HTTP-level (curl/`http.client`) checks alone
- [ ] The detection mechanism is based on the real Development diff, not solely the task file's own predicted file list, to avoid the same near-miss that occurred on the downstream project
- [ ] Backend-only features (no frontend-rendering-affecting changes) are unaffected — no new gate requirement is triggered for them
- [ ] A new Gotcha or note references the downstream project's task 016 as the concrete case that surfaced this gap, following the same documentation precedent as Task 092

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
| `skills/smaqit.feature-new/SKILL.md` | Modify (Phase 2 gate — exact location depends on Task 093's restructuring) |

## Notes

Discovered on a downstream project during task 016/018 (Decommission Frontend Variant A / Development phase), 2026-07-27, in the same session that also surfaced a second, unrelated framework gap (`smaqit.utils.worktree`'s sparse-checkout wrongly excluding `.github/workflows/` — real project CI/CD, not smaqit scaffolding — alongside actual scaffolding paths; not yet filed as its own task as of this writing). Mirrors the Task 089/092 precedent of tracking real framework gaps found on downstream projects rather than fixing them silently in place and losing the feedback loop back to the framework.

**Depends on Task 093** landing first — 093 is actively rewriting Phase 2's gate mechanics and surrounding orchestration in `skills/smaqit.feature-new/SKILL.md` (in progress, Phase A/13 complete as of this task's creation). Planning and implementing this task against the pre-093 file would conflict with that in-flight work.
