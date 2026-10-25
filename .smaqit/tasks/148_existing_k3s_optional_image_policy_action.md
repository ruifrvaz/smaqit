---
status: Not Started
created: "2026-10-25"
---

# existing-k3s: Optional Image Policy Action Around the Image Push

## Description

A platform that hosts existing-k3s apps may enforce an image policy, for example "images only from
the registries this app is allowed to use" and "the pushed image must not be anonymously pullable".
It can ship that policy as a composite GitHub Action that needs no token, Environment, OIDC identity
or mesh access, but smaqit's generated `deploy.yml` has no place to call it: the `build` job does
build and push in one "Build and push image" step.

This task adds an optional variation point: an `Image Policy Action` Constraints row in the
Infrastructure spec. When present, the generated `build` job splits into build → policy check
(`mode: pre-push`) → push → policy check (`mode: post-push`); when absent, the generated workflow is
unchanged. Requested by a downstream platform repository that has built such an action (a pinned
composite Action with inputs `mode` = `pre-push|post-push`, `machine`, `slug`, `image`).

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub Actions, Docker, GHCR, Go, Bash, jq
**Platforms/Environments:** None
**Features/Integrations:** existing-k3s spec parser (Image Policy Action), deploy.yml build job, render-k3s-deploy.sh, golden files
**Versions/Constraints:** smaqit v3.21.0, starts after task 137 merges, optional row (absent means byte-identical output), no input or variable name containing REGISTRY_, pinned uses: reference only

## Design Decisions

- **Optional row, parsed like `Cluster Connect`:** `Image Policy Action` holds a pinned `uses:`
  reference (`actionRefPattern`); an unpinned value is `SLOTS-SPEC-FIELD-INVALID`. It appears in
  `smaqit spec environments --json`.
- **Absent row → unchanged workflow:** an optional template block (the `smaqit:mesh-join` pattern),
  dropped by the renderer when the row is absent, so existing goldens stay byte-identical.
- **Inputs:** `mode` (`pre-push` then `post-push`), `machine` (the dispatched environment's machine
  slug), `slug` (the app slug), `image` (the pushed reference). No input or variable name contains
  `REGISTRY_`, which the render test's leak list refuses.
- **Sequenced after task 137:** both edit `installer/infra_spec.go`, `deploy.yml.k3s.template`,
  `render-k3s-deploy.sh` and the k3s golden files; starting after 137 merges avoids golden-file
  conflicts.

## Implementation Steps

1. `installer/infra_spec.go`: parse and validate the optional `Image Policy Action` row; expose it in
   `--json`; tests in `installer/infra_spec_test.go`.
2. `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template`: split "Build and push
   image" into build, policy `pre-push`, push, policy `post-push`, inside an optional block. Resolve
   the machine in the `build` job from the dispatched `environment` input (the `build` job runs before
   `deploy`'s "Resolve target").
3. `render-k3s-deploy.sh`: keep or drop the block, substitute the action reference, extend the
   leftover-token regex and add a survival check.
4. Tests: golden cases for row present and absent (bump the case count), step order, and the
   `REGISTRY_` leak list.
5. `templates/specs/infrastructure.template.md`, `agents/infrastructure.md` (populate only from the
   platform team's documentation), `docs/wiki/workflows/deployment-slots.md`, `.smaqit/compendium.md`,
   CHANGELOG; release.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] An optional `Image Policy Action` row (pinned `uses:` reference) is parsed, validated with
      `SLOTS-SPEC-*` errors, and shown in `smaqit spec environments --json`
- [ ] With the row present, the `build` job runs build → policy `pre-push` → push → policy
      `post-push` with `mode`, `machine`, `slug` and `image`; with it absent, the rendered workflow is
      byte-identical to before
- [ ] No input or variable name contains `REGISTRY_`
- [ ] Golden cases for both, and the step order is tested
- [ ] Template, agent, wiki, compendium and CHANGELOG updated; no downstream names in canonical text;
      released

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
| installer/infra_spec.go, installer/infra_spec_test.go | Modify |
| skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template, scripts/render-k3s-deploy.sh | Modify |
| installer/k3s_render_test.go, installer/testdata/k3s-deploy/* | Modify / Create |
| templates/specs/infrastructure.template.md, agents/infrastructure.md | Modify |
| docs/wiki/workflows/deployment-slots.md, .smaqit/compendium.md, CHANGELOG.md | Modify |

## Notes

- Requested 2026-10-25 by a peer session in a downstream platform repository; its action was
  verified live from a private repository with `permissions: {}`. A downstream hub task will set the
  row for that platform.
- Start after task 137 merges.
