# User Testing Report

**Date:** 2026-08-18
**Repository:** ruifrvaz/smaqit
**Branch:** main
**Commit:** b2d10d9
**OS/Arch:** Linux/x86_64
**Duration:** ~20 minutes

## Scope
- Test file: `.smaqit/user-testing/tests/111_design-sequence-diagram-enforcement.md`
- Commands executed:
   - `cd installer && make build`
   - `cd installer && make test`
   - `go test ./... -run '...'` (named confirmation of the four new regression tests)
   - `smaqit version` (before/after)
   - `smaqit update`
   - `smaqit init` / `smaqit status` (in `/tmp/smaqit-111-test`)
   - `smaqit design render` (×3 invocations across two designs)
   - `smaqit design attest` (×2 designs)
   - `smaqit design validate` (×4 invocations, across draft/implemented/linked/unlinked states)
   - `rm -rf /tmp/smaqit-111-test`

## Checklist
- [x] Test command discovered and confirmed (`installer/Makefile`, `AGENTS.md` quick commands)
- [x] Dependencies installed (Node.js 22+ and Go toolchain already present; global `smaqit` upgraded via `smaqit update`)
- [x] Test suite executed (`make test` + named-test confirmation)
- [x] Results captured (pass/fail + key errors) — see Execution Log
- [x] Evidence collected (per test file — playbook 111 updated in place with checked-off steps and captured output)

## Execution Log (Timestamped)
- T+0m — `make build` exits 0; `make test` exits 0 (`ok github.com/ruifrvaz/smaqit 18.138s`); all four new tests confirmed `--- PASS` by name
- T+2m — `smaqit version` → v3.1.1 (baseline)
- T+3m — `smaqit update` → `Updated from v3.1.1 to v3.2.0`; `smaqit version` → v3.2.0
- T+4m — `smaqit init` / `smaqit status` in scratch project, both exit 0
- T+6m — Design-sequence fixture without `hide footbox`: `smaqit design render` **rejected** with `DESIGN-VISUAL-INVALID: design-sequence diagrams must include \`hide footbox\`` — fires on `render`, not only `validate` (stronger than the playbook assumed)
- T+8m — Added `hide footbox`; render succeeds; opened rendered PNG — **visually confirmed no duplicated participant boxes at the bottom**
- T+9m — Attest blocked pending the `realizes` target (system-sequence design) existing — created and attested it first, then attested the design-sequence diagram cleanly
- T+11m — Draft-spec exemption check: moved the not-yet-linked design-sequence file aside (an unrelated, pre-existing bidirectional-link check would otherwise mask the exemption check), then `smaqit design validate` passed cleanly at `status: draft`
- T+13m — Flipped spec and its system-sequence design to `status: implemented`; `smaqit design validate` **failed** with `DESIGN-ARTIFACT-MISSING: implemented spec has no grounding design-sequence diagram for DSG-FUN-ORDER-SYSTEM-SEQUENCE (see smaqit.feature-new Development phase requirement)` — exact match to the task's design decision
- T+15m — Restored the design-sequence file and linked it from the spec's Design References (the two-link convention); `smaqit design validate` → `✓ Validated 2 canonical design(s)` — fully clean
- T+18m — Cleanup: `rm -rf /tmp/smaqit-111-test`; global `smaqit` left at v3.2.0 intentionally

## Results
- Overall: **PASS**
- Summary:
   - All three fixes (two-link convention, missing-grounding detection, footbox suppression) confirmed working against the real released v3.2.0 binary, driven through the real PlantUML render/attest pipeline — not just the merged Go unit tests
   - Both intentional-failure scenarios produced the exact expected error messages
   - Real self-update path (`smaqit update`) confirmed functional, landing the correct version
   - Two playbook-ordering issues surfaced and were worked around live (see Pain Points) — both are test-fixture sequencing artifacts, not defects in the shipped fixes

## Pain Points
- Blockers:
   - None
- Issues:
   - The original playbook draft assumed `smaqit design render` would succeed on a footbox-missing design-sequence diagram and only `smaqit design validate` would reject it. In reality `render` calls `parseDesign`, which validates metadata upfront, so `render` itself already rejects it. Not a defect — just means the playbook's Scenario A step ordering under-stated how early the check fires. Corrected in the playbook.
   - The original playbook authored the design-sequence diagram before its spec linked it back, then ran a general `smaqit design validate` sweep in between (to test the draft-exemption case). This tripped an unrelated, pre-existing check (`validateDesignReferences`' bidirectional spec↔design link requirement) that has nothing to do with task 111's fixes, producing a misleading failure. Worked around by temporarily moving the unlinked design file out of `docs/designs/` for that one check. A real project would never hit this, since a Design Sequence Diagram is only ever authored *after* implementation and is linked immediately — but it's a real trap for anyone hand-authoring test fixtures in this exact order.
- UX Friction:
   - None beyond the above — once the corrected ordering (SSD → attest → spec → flip status → author DSD → attest → link) was followed, every command behaved exactly as documented.
- Performance:
   - `make test` (full suite, Node-gated tests included): ~18s. `smaqit update`: a few seconds. No performance concerns.

## Recommendations
- Update `installer/design.go`'s doc comment or `framework/ARTIFACTS.md` to note explicitly that `hide footbox` (and other `validateDesignMetadata` checks) fire on `smaqit design render` too, not only `smaqit design attest`/`validate` — the current framing in a few places implies render is purely a syntax/rendering step.
- No code change needed for the bidirectional-link-check ordering trap — it's a hand-authoring pitfall, not a tool defect. Worth a one-line callout in `framework/ARTIFACTS.md`'s Reference Rules section: a design's declared `specifications` back-reference must exist before the general sweep will validate cleanly, even if the design itself is otherwise perfect.
- Playbook 111 has been corrected in place with the real command ordering; future re-runs (e.g. regression-testing a later change to this same area) can follow it directly without re-discovering these two traps.
