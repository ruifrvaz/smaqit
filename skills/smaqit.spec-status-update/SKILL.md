---
name: smaqit.spec-status-update
description: Use when implementation or deployment is confirmed and spec files need to be brought in sync with the live codebase. Updates frontmatter fields (`status`, `deployed`, `updated`) and flips acceptance criteria checkboxes (`[ ]→[x]` or `[ ]→[!]`) without running a full spec agent. Also use when the user asks to mark a spec as deployed, update spec status after a release, or record which acceptance criteria are met.
metadata:
  version: "1.0.1"
---

# Spec Status Update

## Steps

1. **Identify target files.** If no file is specified, ask the user which spec or phase to update (e.g., `"all functional specs"`, `"specs/functional/diagnostic-flow.md"`).

2. **For each spec file:**

   a. **Update frontmatter fields:**
      - `status` → set to the phase-appropriate value: `draft`, `implemented`, or `deployed`
      - `updated` → set to today's date in `YYYY-MM-DD` format
      - `deployed` → set to ISO 8601 datetime (e.g., `2026-05-21T00:00:00Z`) only when transitioning to `deployed`; omit otherwise

   b. **Update acceptance criteria checkboxes:**
      - `- [ ]` → `- [x]` for each criterion confirmed as met
      - `- [ ]` → `- [!]` for each criterion that is explicitly deferred or untestable — add a brief inline note (e.g., `- [!] FUN-HOME-ANIM-007 — aesthetic quality, requires manual review`)
      - Leave uncertain criteria as `- [ ]`; do not change them without confirmation

   c. **Synchronize linked designs:**
      - Resolve every pair from `## Design References`; missing or broken pairs stop with `DESIGN-ARTIFACT-MISSING`.
      - Set each design `status` to the least-advanced status of all active linked specs. Do not change PlantUML, PNG, hashes, or visual attestation for a status-only transition.
      - Run `smaqit design validate <design.md>`. A lifecycle mismatch, stale hash, or failed visual attestation blocks the spec status update.

3. **Report** — for each file: old status → new status, number of checkboxes changed to `[x]`, number changed to `[!]`, and any left uncertain as `[ ]`.

## Output

- Updated spec file(s) and lifecycle-synchronized linked design metadata
- Inline report: file list, status transitions, and checkbox counts

## Scope

- Does NOT write new acceptance criteria or modify spec prose beyond frontmatter and checkboxes.
- Does NOT validate whether criteria are actually met — the caller is responsible for confirming this before invoking.
- Does NOT create spec files. Use the appropriate specification agent for new specs.
- Does NOT author, render, or visually attest designs. Use `smaqit.design-validate` and the owning specification agent when design content is stale or invalid.

## Examples

**Trigger:** After deploying the absenteeism feature, user runs `/spec.status absence-api deployed`.

**Output:** `specs/functional/absence-api.md` updated — `status: deployed`, `deployed: 2026-05-21T00:00:00Z`, 6 checkboxes `[ ]→[x]`, 1 checkbox left as `[ ]` (unverified).

## Gotchas

- Some spec files carry `status: draft` even after partial implementation. Verify all criteria under the new status are actually met before updating. (Observed in a past session: `smaqit plan` returned empty because a stack spec was still `implemented` when new criteria had not been addressed.)
- `deployed` is a **datetime**, not a date. Always include the time component (`T00:00:00Z` is acceptable if the exact time is unknown).
- The `[!]` marker is a project convention for deferred or untestable criteria. Always add a brief inline note when using it.

## Completion

- [ ] Target spec file(s) identified
- [ ] Frontmatter fields updated (`status`, `updated`, `deployed` as applicable)
- [ ] Acceptance criteria checkboxes updated with confirmation
- [ ] Linked design lifecycle statuses synchronized and `smaqit design validate` passed
- [ ] Report delivered (file list + transitions + checkbox counts)

## Failure Handling

| Situation | Action |
|-----------|--------|
| No file specified | Ask the user for the target spec or phase before proceeding |
| Criterion status is uncertain | Leave checkbox as `- [ ]`; report it as uncertain in the output |
| File does not have frontmatter | Add a minimal frontmatter block before the first heading |
| Linked design is missing, stale, or visually invalid | Stop with the canonical design failure; do not advance the spec alone |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
