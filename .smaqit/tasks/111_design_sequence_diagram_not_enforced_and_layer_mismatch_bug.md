# Design Sequence Diagram Has No Deterministic Enforcement, and a Layer-Mismatch Bug Rejects the Established Convention That Would Link One

**Status:** Not Started
**Created:** 2026-08-18
**Mode:** Assisted

## Description

The Development phase (`smaqit.feature-new` Phase 2, Step 3: "Invoke `/smaqit.development`...") is documented as owning a mandatory artifact — a `design-sequence`-layer diagram that grounds a Functional spec's black-box `system-sequence` design in real implementation, via `' impl: <path>:<line>` citations, `realizes:` pointing at the system-sequence design it grounds. This is a real, code-enforced artifact type (`installer/design.go`'s `designProfiles["design-sequence"]`, `validateDesignSequenceGrounding` at `design.go:1019`, which correctly requires `impl:` citations and requires every operation promised by the `realizes:` target's message labels to be represented — both working as intended once a diagram exists).

The problem: **nothing forces one to exist.** Found live in a downstream project (2026-08-18, `iodis-crm-poc`) — an orchestrating session ran a `smaqit.feature-new` Phase 2 (Development) child task, implemented the feature's code directly instead of invoking `/smaqit.development`, bumped the Functional/Business specs to `status: implemented`, and moved on. No tool anywhere flagged the missing Design Sequence Diagram. A second, independent instance of the identical failure mode already happened in this same downstream project (see task 062 there: "first Development-agent invocation wrongly skipped `docs/designs/`, missing its separate mandatory Design Sequence Diagram requirement") — two occurrences is a real pattern, not a one-off.

Root-caused to two distinct issues in `installer/design.go`:

### Issue 1 — the phase-readiness gate structurally cannot enforce this artifact

`specDesignReady` (`design.go:746`) is the function behind `smaqit plan --phase=develop`'s gate, invoked via `getPhaseDesignGateSpecs` (`installer/spec.go:242`). Since task 109, that gate is deliberately scoped to specs with `status: draft`/`failed` only (`isCycleRelevant`, `spec.go:188`) — specs already `implemented` are out of its scope by design, because task 109's fix was about not re-litigating specs the current cycle didn't touch. But a Design Sequence Diagram, by definition, is only ever produced *after* a spec moves past `draft` — so the one gate that runs during/around Development is structurally incapable of ever checking for it. `design.go:35-36`'s own comment confirms the `design-sequence` layer is "never subject to the active-spec-needs-a-design completeness walk" at all, anywhere. Running `smaqit plan --phase=develop` any number of times gives zero signal about this artifact's existence.

### Issue 2 — the general design-validate sweep actively rejects the established two-link convention

`smaqit design validate` (no path — the general sweep) walks every active spec via `collectActiveSpecDesignDiagnostics` (`design.go:871`), calling `specDesignReady(path, layer)` for each. Inside `specDesignReady`, the per-link loop (`design.go:787`) fails the instant it encounters a linked design whose `Front.Layer != layer` — but the *already-established, currently-shipped* convention for a spec that has a Design Sequence Diagram is to link **both** the `functional`-layer system-sequence design **and** the `design-sequence`-layer diagram from the same Functional spec's `## Design References` section (confirmed via two real files in the downstream project: `specs/functional/price-override.md` and `specs/functional/settings.md`, both listing a `dsg-fun-*-system-sequence.md` link immediately followed by a `dsg-dsd-*-design-sequence.md` link). `specDesignReady`'s layer check treats the second link as an error.

This bug is currently **dormant** for both existing examples, purely by accident: both `price-override.md`'s and `settings.md`'s *first*-listed link (their own functional system-sequence design) already fails `validateDesignArtifact` on an unrelated, newer rule ("system-sequence diagrams must declare exactly participant \"System\" as System") and short-circuits the loop before it ever reaches the second (design-sequence) link. Fix that unrelated participant-naming issue on either file, and this layer-mismatch bug immediately surfaces and breaks the phase-readiness/general-sweep check for an already-shipped, already-`validated` feature.

## Design Decisions

Both fixes below are adopted — not left as open questions. A cheaper alternative (only naming the artifact explicitly in `smaqit.feature-new`'s own Gate/Completion checklist text) was considered and explicitly rejected: that is pure textual discipline with no code-level enforcement, and does nothing to prevent the exact failure mode observed twice already. The two fixes below are both real code-level gates, matching how every other design artifact type in this framework is already enforced.

- **Fix 1 (Issue 2 — layer-mismatch bug): `specDesignReady`'s per-link loop must not fail solely because a linked design's layer is `design-sequence`.** At `design.go:787`, change the unconditional `if d.Front.Layer != layer { return false, "...linked design belongs to another layer" }` to permit `design-sequence` as an always-acceptable companion layer for a `business`/`functional` spec: only fail on layer mismatch when `d.Front.Layer != layer && d.Front.Layer != "design-sequence"`. This does not change what counts toward `seenPair` for the spec's *own* layer requirement — a `design-sequence` link should not, by itself, satisfy "this spec has a design pair in its own layer"; it only stops being treated as an outright rejection.

- **Fix 2 (Issue 1 — no enforcement at all): add a dedicated grounding check to the general sweep, not the phase-readiness gate.** The phase-readiness gate (`specDesignReady`/`getPhaseDesignGateSpecs`) is the wrong enforcement point — task 109 correctly scoped it away from already-implemented specs, and a Design Sequence Diagram structurally cannot exist before implementation. The right point is `collectActiveSpecDesignDiagnostics` (`design.go:871`, backing the general `smaqit design validate` sweep, which is already run as a matter of course after any design change — confirmed by this framework's own established downstream-project workflow). Add a new check, scoped to `functional`-layer specs only (per the existing `realizes: DSG-FUN-[CONCEPT]-SYSTEM-SEQUENCE` convention in `templates/designs/design-sequence.template.md` — Business use-case designs are not paired with a Design Sequence Diagram in this framework's established convention and should not be required to have one) with `status` at or beyond `implemented`: at least one `design-sequence`-layer design must exist, linked from the same Design References section, whose `realizes:` field names one of that spec's own linked `system-sequence`-layer design ids, and which itself passes `validateDesignArtifact` (including the existing, correct `validateDesignSequenceGrounding` promised-operations check at `design.go:1116` — no change needed there). Missing → `DESIGN-ARTIFACT-MISSING: implemented spec has no grounding design-sequence diagram for <system-sequence-design-id> (see smaqit.feature-new Development phase requirement)`.

- **Not doing:** no change to `validateDesignSequenceGrounding` itself (`design.go:1019`) — its promised-operations matching (`design.go:1116`) is real, correct, and already working exactly as intended once a diagram exists; this task is only about the two gaps that let a diagram never exist at all, or wrongly reject the two-link convention once it does.

- **No grandfathering, no opt-out, no "only specs implemented after this ships" carve-out — this is intentionally a breaking change.** Fix 2 will immediately fail `smaqit design validate` for every pre-existing `functional` spec, in every downstream project, that reached `status: implemented` before this requirement existed and never got a Design Sequence Diagram (almost certainly most of them — the two real examples checked in this task, `price-override.md`/`settings.md`, are the exception, not the rule). That is the correct, desired outcome, not a migration problem to soften: a downstream project hitting this after a `smaqit update` should fix forward (author the missing diagram) exactly the same way it would for any other spec that never got a required design pair, not be exempted by a version check, a legacy flag, or a "specs implemented before vX.Y only need to warn" downgrade. Do not add any such mechanism as part of implementing this task, even if it seems like a kinder rollout — a downstream project silently missing this artifact forever, just because it predates the fix, defeats the entire point of filing this task.

## Implementation Steps

TBD — sketch, not committed:

1. Fix 1: change the layer-mismatch condition at `design.go:787` as described above. Add a regression test: a `functional` spec whose `## Design References` links both a `system-sequence` design and a `design-sequence` design passes `specDesignReady`; a spec linking a design from an unrelated third layer (e.g. `infrastructure`) still correctly fails.
2. Fix 2: add the new grounding check inside (or alongside) `collectActiveSpecDesignDiagnostics`, scoped to `functional`-layer, `status >= implemented` specs. Reuse the existing Design-References-link-parsing logic already in `specDesignReady` rather than duplicating it — consider extracting that parsing into a shared helper both functions call. Add regression tests: an `implemented` functional spec with a valid, `realizes`-linked design-sequence diagram passes; the same spec with no design-sequence diagram at all fails with the new `DESIGN-ARTIFACT-MISSING` message; a still-`draft` functional spec with no design-sequence diagram does **not** fail (the requirement only applies once implementation has actually happened).
3. Re-verify the two currently-dormant real-world cases this task references (`price-override.md`/`settings.md` in the downstream `iodis-crm-poc` project, or an equivalent local fixture) both pass cleanly under both fixes together, not just individually.
4. Update any doc/help text describing `smaqit design validate`'s general-sweep scope to mention the new grounding check.

## Known Issues Triage

**Triaged:** 2026-08-18
**Tools searched:** none
**Result:** Clear — internal bug/gap in `installer/design.go`, not a third-party dependency issue.

## Acceptance Criteria

- [ ] `specDesignReady` no longer rejects a `business`/`functional` spec's `## Design References` section solely because one of its links is a `design-sequence`-layer design
- [ ] `smaqit design validate` (general sweep) reports `DESIGN-ARTIFACT-MISSING: implemented spec has no grounding design-sequence diagram...` for a `functional` spec at `status: implemented`+ that has no linked, `realizes`-matched, valid `design-sequence` design
- [ ] A still-`draft` functional spec is never blocked by the new check — it only applies once a spec has moved past `draft`
- [ ] Regression tests cover both fixes independently and together
- [ ] The two real, currently-dormant downstream cases (`price-override.md`/`settings.md`-shaped Design References) are confirmed to pass cleanly

## Findings

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
| `installer/design.go` | Modify — `specDesignReady` (layer-mismatch fix), `collectActiveSpecDesignDiagnostics` (new grounding check) |
| `installer/design_test.go` | Modify — regression coverage for both fixes |

## Notes

Found live in `iodis-crm-poc` (downstream project) during that project's task 076 (Phase 2/Development of a `smaqit.feature-new` cycle, 2026-08-18) — an orchestrating session implemented the feature's code directly rather than invoking `/smaqit.development`, and neither `smaqit plan --phase=develop` nor `smaqit design validate` (scoped or general) ever signaled the missing Design Sequence Diagram until a human reviewer asked "shouldn't the development [phase] generate designs?" The diagrams were authored retroactively in that same session and both fixes below were needed just to get them to validate cleanly — the layer-mismatch bug (Issue 2) was hit immediately upon linking the new design-sequence diagram from the Functional spec's Design References, exactly as this task predicts it will eventually hit `price-override.md`/`settings.md` too once their own unrelated participant-naming issue is fixed. That "second bug currently masked by a first, unrelated bug" relationship is real and independently reproducible — not speculative.

This is the same *category* of gap as task 109 (a design-related gate with real, observed downstream impact, found live rather than hypothesized) — recommend similar priority.
