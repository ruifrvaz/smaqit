# Feature Workflow Consolidation

**Date:** 2026-07-27
**Session Focus:** Reassess the overlap between `smaqit.feature-new` and `smaqit.feature-deploy`, choose a single post-MVP workflow, and create a clean-session-ready implementation task.
**Tasks Referenced:** 092, 093
**Tasks Completed:** None

---

## Actions Taken

### Session Initialization
- Loaded the project README, framework instructions, current planning state, latest history, compendium, installer entry points, and source areas for the next active task.
- Confirmed v1.9.0 had introduced `smaqit.feature-deploy` and that Task 092 originally targeted its unconditional direct-push assumption.

### Feature Workflow Reassessment
- Compared `smaqit.feature-new` and `smaqit.feature-deploy` end to end.
- Confirmed that Feature New already owns the full post-MVP lifecycle: specification revalidation, development, deployment, validation, and release close-out.
- Identified that the standalone deploy skill duplicated Feature New's deployment phase and split a single pragmatic user journey across two entry points.
- Agreed that PR merge is the deployment approval gate; an unmerged PR is active work awaiting approval, so a separate deploy-now/defer model is unnecessary.

### Task Replacement and Planning
- Abandoned Task 092 as superseded while retaining its original trigger incident and analysis.
- Created Task 093 to retire `smaqit.feature-deploy` and consolidate post-MVP execution into `smaqit.feature-new`.
- Expanded Task 093 from a narrow deployment cleanup into a standalone implementation specification with 13 ordered steps and 16 measurable acceptance criteria.

### Clean-Session Audit
- Audited Task 093 as if it were opened in a new session with no conversation context.
- Added a mandatory first-step migrate/reject inventory of Feature Deploy behavior before deletion.
- Detected deterministic duplicate specification generation: Feature New Phase 1 leaves touched specs `draft`, then Development, Deployment, and Validation re-invoke their specification agents.
- Defined a durable Phase 1 exact-path handoff and `specification_mode: prevalidated` for all three phase agents, while preserving orchestration-first behavior for direct agent invocation.
- Defined `deployment_path: existing-cicd-pr` so the Deployment agent owns artifact preparation, PR creation, human merge, exact CI observation, verification, spec state, and report generation as one operation.
- Added a deterministic trigger table covering direct push, PR-dispatch sentinel, duplicate-trigger rejection, and unsupported workflow layouts.
- Replaced local-HEAD assumptions with exact workflow-run SHA verification through a backward-compatible `--expected-sha`.
- Defined a self-contained amendment scanner contract and a terminal amendment lifecycle.
- Found that deleting embedded source would not remove the retired skill from existing initialized projects; specified a persistent exact-file installer tombstone that preserves custom content during reinit and uninstall.

### Project Research Refresh
- Refreshed the seven-day-old project research map.
- Re-read project manifests and instructions, verified 16 official documentation URLs, and removed the stale Task 084 research annotation because no task is currently in progress.

## Problems Solved

- **Redundant post-MVP entry points:** Resolved the architectural ambiguity by choosing Feature New as the sole top-level feature workflow.
- **False direct-push assumption:** Captured workflow inspection rules so a successful git push cannot be mistaken for a deployment trigger.
- **Duplicate deployment risk:** Specified that the PR sentinel is omitted when merge-to-main already triggers deployment.
- **Duplicate specification orchestration:** Added a durable prevalidated-spec handoff so each specification agent runs once per feature unless Phase 1 is explicitly reopened.
- **Contradictory deployment ownership:** Assigned the complete PR/CI operation to one explicit Deployment-agent mode rather than splitting it between the skill and agent.
- **Incorrect deploy revision verification:** Bound verification to the selected workflow run's actual `headSha`.
- **Non-terminal amendment handling:** Required each amendment to be incorporated, recorded, removed, and rescanned successfully.
- **Incomplete skill retirement:** Added upgrade and uninstall cleanup requirements for already-installed Feature Deploy packages, with custom-content preservation.
- **Stale project research:** Rebuilt the project map from current manifests, instructions, and session-relevant tooling.

## Decisions Made

- `smaqit.feature-new` will be the only post-MVP feature/deployment entry point.
- `smaqit.feature-deploy` will be retired rather than delegated to or retained alongside Feature New.
- Feature deployment is mandatory within the end-to-end feature cycle; the PR merge is the human approval gate.
- Feature New Phase 1 owns specification revalidation once and records a durable exact-path handoff.
- Development, Deployment, and Validation retain their orchestration-first default but gain an explicit prevalidated-spec mode for Feature New.
- Deployment gains a distinct existing-CI/CD PR mode that owns the complete execution and evidence chain.
- Trigger selection is derived from committed workflow files and must reject ambiguous or duplicate layouts.
- Production verification uses the selected deployment run's SHA, not an assumed local revision.
- Shipped skills own their operational scripts; Feature New receives its own corrected amendment scanner.
- Retired embedded assets require a permanent exact-file tombstone because installer reinitialization overlays rather than prunes.
- Historical release, history, compilation, and completed/abandoned task records remain unchanged when current product content is retired.

## Files Modified

- `.smaqit/tasks/092_feature_deploy_stale_push_trigger_assumption.md` — marked abandoned with the superseding-task reason.
- `.smaqit/tasks/093_consolidate_post_mvp_feature_deployment.md` — created and refined into a clean-session-ready execution specification.
- `.smaqit/tasks/PLANNING.md` — moved Task 092 to Abandoned and added Task 093 as the next active task.
- `.smaqit/references/project-research.md` — refreshed project tooling and verified documentation topology.
- `.smaqit/compendium.md` — updated with durable workflow and retirement conventions from this session.
- `.smaqit/history/067_feature_workflow_consolidation_2026-07-27.md` — created as the complete session record.

## Next Steps

- Start Task 093 with `task.start 093`.
- Execute its first-step Feature Deploy migrate/reject inventory before changing source.
- Implement and verify the consolidated Feature New workflow, phase-agent modes, deployment SHA handling, amendment scanner, and installer retirement tombstone.
- After Task 093, reassess the related direct-push/PR-marker ambiguity in `smaqit.new-greenfield-project` as separate follow-up work.

## Session Metrics

- **Duration:** Not recorded
- **Tasks created:** 1
- **Tasks abandoned:** 1
- **Tasks completed:** 0
- **Task 093 implementation steps:** 13
- **Task 093 acceptance criteria:** 16
- **Project documentation URLs verified:** 16
- **Files created:** 2
- **Files modified:** 4
- **Source implementation changes:** 0
