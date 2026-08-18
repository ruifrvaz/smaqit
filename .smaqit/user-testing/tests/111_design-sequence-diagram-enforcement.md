# Design Sequence Diagram Has No Deterministic Enforcement, and a Layer-Mismatch Bug — E2E Test Playbook

**Test ID:** 111
**Title:** Design Sequence Diagram Has No Deterministic Enforcement, and a Layer-Mismatch Bug Rejects the Established Convention That Would Link One
**Date:** 2026-08-18
**Tester:** User Testing Agent
**Task:** 111 (released as v3.2.0)
**Executed:** 2026-08-18 — **PASS**

## Objectives

Validate the **actual released v3.2.0 artifact** — not a local dev build — end-to-end: download/install it through the same path a real consumer would (`smaqit update` / `install.sh`), then in a fresh scratch project driven entirely through the globally-installed `smaqit` CLI, confirm the three fixes behave correctly: (1) a Functional spec may link both its `system-sequence` design and its `design-sequence` companion without a layer-mismatch rejection, (2) `smaqit design validate`'s general sweep reports `DESIGN-ARTIFACT-MISSING` for an implemented Functional spec with no grounding Design Sequence Diagram, and (3) a Design Sequence Diagram without `hide footbox` is rejected, and the rendered PNG shows no duplicated participant boxes once it is present. This exercises the real PlantUML render/attest pipeline and the real self-update mechanism, neither of which the merged Go unit tests exercise (they bypass Node via fixture hashing and never touch the release/update path at all).

## Prerequisites

- Node.js 22+ available on PATH (mandatory consumer prerequisite for `smaqit design render`)
- `smaqit` already installed globally (this machine had v3.1.1 at `~/.local/bin/smaqit`, confirmed via `smaqit version`) — Step 2 upgrades this in place
- **Step 2 mutates the tester's real, shared global `smaqit` installation** (`~/.claude/{agents,commands,skills}/`, `~/.local/bin/smaqit`, etc.) — it is not sandboxed like `scripts/smoke-test-installer.sh`'s isolated-HOME approach, by design: the objective is to prove the real consumer upgrade path works, not a synthetic one. Acceptable here since v3.2.0 is already the latest published release and every other local project on this machine should be running it anyway.
- No running services required — `smaqit` is a CLI tool, not a daemon

## Test Steps

### Step 1 — Build & Unit Test Gate (source sanity check)
- [x] `cd installer && make build` exits 0
- [x] `cd installer && make test` exits 0 (zero failures) — `go test ./...` + `go vet ./...`, 18.1s. All four new regression tests confirmed passing by name: `TestSpecDesignReadyAcceptsDesignSequenceCompanionLink`, `TestSpecDesignReadyStillRejectsUnrelatedLayerLink`, `TestCollectActiveSpecDesignDiagnosticsRequiresGroundingOnceImplemented`, `TestDesignSequenceRequiresFootboxHidden`

### Step 2 — Download and Update to the Released Version
- [x] `smaqit version` → `smaqit v3.1.1` (baseline confirmed)
- [x] `smaqit update` exits 0 → `Updated from v3.1.1 to v3.2.0`, refreshed Copilot/Claude/Codex agent and skill payloads
- [x] `smaqit version` → `smaqit v3.2.0` confirmed

### Step 3 — Scaffold a Scratch Project
- [x] `mkdir -p /tmp/smaqit-111-test && cd /tmp/smaqit-111-test`
- [x] `smaqit init` exits 0 — `✓ Initialized smaqit v3.2.0`
- [x] `smaqit status` exits 0 — clean, 0 specs, Phase 1/2/3 all "Not started" as expected for a fresh project

### Step 4 — Additional Validation: the three shipped fixes — **all confirmed on the real v3.2.0 binary**

**Scenario A — Fix 3 (footbox)**
- [x] Created the design-sequence diagram **without** `hide footbox`
- [x] `smaqit design render` → **rejected immediately** with `DESIGN-VISUAL-INVALID: design-sequence diagrams must include \`hide footbox\`` (exit 1). **Deviation from the written playbook:** the check fires on `render`, not only `validate` — `render` calls `parseDesign`, which validates metadata upfront. Stronger confirmation than planned: the check is enforced uniformly across render/attest/validate, not just validate.
- [x] Added `hide footbox` → `smaqit design render` exits 0, PNG produced
- [x] Opened the rendered PNG — **visually confirmed no duplicated participant box row at the bottom** (the exact reported symptom)
- [x] `smaqit design attest` — required the paired system-sequence design to exist first for its completeness check (`realizes` target lookup); created that fixture, then attest succeeded (0)

**Scenario B — Fix 1 (two-link convention) and Fix 2 (missing-grounding detection)**
- [x] Created and attested the system-sequence design (0)
- [x] Opened its rendered PNG — visually confirmed correct black-box shape (one actor, one `System` participant, no footbox)
- [x] Created `specs/functional/order.md`, `status: draft`, linking only the system-sequence design
- [x] `smaqit design validate` **passed cleanly** at `draft` — confirms the grounding requirement is exempt pre-implementation. **Deviation:** to test this cleanly, the design-sequence file (already authored in Scenario A) had to be moved out of `docs/designs/` first — with it present but not yet linked from the spec, the general sweep's unrelated, pre-existing bidirectional-reference check (`validateDesignReferences`) fails on its own, masking the exemption check. Restored after this step. Real projects never hit this ordering artifact since a DSD is only ever authored *after* implementation, by which point it's immediately linked.
- [x] Flipped spec **and** its system-sequence design to `status: implemented` (design's lifecycle rank must track its spec's rank — an unrelated, pre-existing rule)
- [x] `smaqit design validate` **failed** with exactly `DESIGN-ARTIFACT-MISSING: implemented spec has no grounding design-sequence diagram for DSG-FUN-ORDER-SYSTEM-SEQUENCE (see smaqit.feature-new Development phase requirement)` — **Fix 2 confirmed live**, message matches the task's design decision verbatim
- [x] Restored the design-sequence file and added its link to `## Design References`, immediately below the system-sequence link — the two-link convention
- [x] `smaqit design validate` → `✓ Validated 2 canonical design(s)` — **fully clean pass, confirms Fix 1 and Fix 2 together**

### Step 5 — Cleanup
- [x] `rm -rf /tmp/smaqit-111-test`
- [x] Global `smaqit` remains updated to v3.2.0 — intentional, not rolled back (see Prerequisites)

## Pass/Fail Criteria

**PASS** — All checkboxes are checked, all `exits 0` commands genuinely exit 0, both intentional-failure commands genuinely fail with the exact stated message substring, and the rendered design-sequence PNG shows no duplicated footbox row.
**FAIL** — Any checkbox unchecked, any unexpected exit code, a failure message that doesn't match, or a footbox row still visible in the rendered PNG.

**Result: PASS.** Every checkbox verified against the real v3.2.0 release. Two playbook ordering issues were found and worked around during execution (noted inline above) — both are artifacts of testing a synthetic fixture out of the order a real project reaches these states in, not defects in the shipped fixes themselves.

## Evidence to Capture

- `make test`: `ok github.com/ruifrvaz/smaqit 18.138s`, all four new tests individually confirmed `--- PASS`
- `smaqit version`: `v3.1.1` → (`smaqit update` → `Updated from v3.1.1 to v3.2.0`) → `v3.2.0`
- Footbox rejection: `DESIGN-VISUAL-INVALID: design-sequence diagrams must include \`hide footbox\`` (on `render`)
- Missing-grounding rejection: `DESIGN-ARTIFACT-MISSING: implemented spec has no grounding design-sequence diagram for DSG-FUN-ORDER-SYSTEM-SEQUENCE (see smaqit.feature-new Development phase requirement)`
- Rendered `dsg-dsd-order-design-sequence.png` before/after `hide footbox` — visually confirmed no bottom participant-box duplication in the "after" render
- Final `smaqit design validate`: `✓ Validated 2 canonical design(s)` (clean)
