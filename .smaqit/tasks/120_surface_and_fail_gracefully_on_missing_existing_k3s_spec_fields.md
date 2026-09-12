---
status: Not Started
created: "2026-09-12"
---

# Surface and Fail Gracefully on Missing existing-k3s Infrastructure-Spec Fields

## Description

Found reviewing the k3s app-deployment skill family (tasks 116-118) from a downstream infra
repo's side, immediately after task 119. `smaqit.infrastructure-request-k3s-onboarding`'s own
Pre-conditions require the Infrastructure spec to declare, for the target environment: the
platform repository (a `## Constraints` table `Platform Repo` row), the registry-file path, and
the target machine-slug. Deploying afterward also needs the ingress class, `ClusterIssuer` name,
and quota/limit numbers reflected somewhere in the spec.

Checked every upstream point that could plausibly elicit, validate, or write these fields, and
none of them do:

- `smaqit.input-infrastructure` (the gate before Infrastructure spec generation) has no mention
  of k3s, `existing-k3s`, or any of these fields — its questions are generic ("where will this
  run", "what platform").
- `smaqit.input-deployment` names and explains `existing-k3s` as a `provisioning_mode` value, but
  its own Readiness Condition states explicitly "no required sections — defaults are always
  valid." It never elicits these specific values.
- The Infrastructure spec-writing agent itself (`smaqit-infrastructure`, ~286 lines) has zero
  references to k3s or any provisioning-mode-specific field, for any mode.
- `smaqit.new-greenfield-project`'s own orchestration text says to invoke
  `smaqit.infrastructure-request-k3s-onboarding` "for the test environment's registered
  machine-slug," silently assuming that slug (and its siblings) are already sitting in the spec,
  without ever saying who puts them there or when.
- `smaqit.infrastructure-request-k3s-onboarding`'s own Failure Handling table has no row for "the
  Infrastructure spec doesn't declare these fields at all" — every other precondition failure
  (missing credential, missing `entry_content`, nonexistent platform repo) is handled explicitly;
  this one silently falls through to undefined behavior.

**Explicitly out of scope:** smaqit inventing, inferring, or defaulting any platform-specific
value. Only the platform team that owns the target k3s cluster's own onboarding repo can supply
the actual `Platform Repo`, registry path, machine slug, ingress class, `ClusterIssuer` name, or
quota numbers — publishing that is the platform side's documentation job (tracked, for one
downstream platform repo, as that repo's own task). This task is only about smaqit failing loudly,
early, and specifically when these fields are missing, instead of assuming they exist.

## Design Decisions

- **Fail-fast, not fail-smart.** No new elicitation intelligence that guesses or infers
  platform-specific facts. The fix is entirely about surfacing a known requirement earlier and
  producing an actionable error instead of an undefined one later.
- **Three touch points, in order of where a gap-caused failure should be caught:**
  1. At mode-resolution time (`smaqit.input-deployment`), tell the operator that resolving to
     `existing-k3s` means the Infrastructure spec's Constraints table must already carry these
     fields, sourced from the platform repository's own onboarding documentation — not invented
     locally. Stays a soft prompt, consistent with this skill's existing "defaults are always
     valid" posture; it does not need to become a hard block.
  2. At spec-authoring time, the Infrastructure spec-writing agent's own instructions should name
     these fields for an `existing-k3s` target, the same way it would for any other
     provisioning-mode-specific fact a spec needs to carry.
  3. At the point of actual failure (`smaqit.infrastructure-request-k3s-onboarding`'s Failure
     Handling table), add the missing row: spec doesn't declare `Platform Repo` /
     `registry_file_path` / `machine_slug` → stop, name exactly which field is missing, and point
     at the platform repository's own onboarding documentation. Never fabricate a value or guess a
     path.
- **Check `smaqit.feature-new` for the same gap.** It also threads `existing-k3s`; confirm whether
  it duplicates `smaqit.new-greenfield-project`'s assumption or correctly delegates to it, and
  apply the same fix if it duplicates.

## Acceptance Criteria

- [ ] `smaqit.infrastructure-request-k3s-onboarding`'s Failure Handling table has an explicit row
      for a spec missing `Platform Repo`/`registry_file_path`/`machine_slug`, naming the exact
      fields and pointing at the platform's own documentation — never fabricating a value
- [ ] `smaqit.input-deployment`'s `existing-k3s` value description tells the operator, at
      mode-resolution time, that these fields must already exist in the spec and where they come
      from
- [ ] The Infrastructure spec-writing agent's instructions name these fields as required spec
      content for an `existing-k3s` target
- [ ] `smaqit.feature-new` checked for the same silent assumption; fixed if it duplicates rather
      than delegates
- [ ] No component introduced by this task infers, guesses, or defaults any platform-specific
      value — verified by inspection of every new instruction added

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

## Notes

- Origin: 2026-09-12, found while assessing whether smaqit has everything needed to run an
  `existing-k3s` deployment end to end for a downstream infra repo. The onboarding-request and
  deploy skills themselves (tasks 117/118, fixed further in task 119) are sound; this gap sits
  entirely upstream of them, in spec authoring.
- Companion to task 119 (same review, landed as v3.5.1's fixes) and to a downstream infra repo's
  own task 020, which is publishing the actual values this task's fixes would point the operator
  toward — that task is a separate repo's own documentation work, not a smaqit dependency.
