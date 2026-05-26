---
name: smaqit.spec-status-update
description: Use when implementation or deployment is confirmed and spec files need to be brought in sync with the live codebase. Updates frontmatter fields (`status`, `deployed`, `updated`) and flips acceptance criteria checkboxes (`[ ]→[x]` or `[ ]→[!]`) without running a full spec agent. Also use when the user asks to mark a spec as deployed, update spec status after a release, or record which acceptance criteria are met.
metadata:
  version: "1.0.0"
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

3. **Report** — for each file: old status → new status, number of checkboxes changed to `[x]`, number changed to `[!]`, and any left uncertain as `[ ]`.

## Output

- Updated spec file(s) with revised frontmatter and accurate acceptance criteria checkboxes
- Inline report: file list, status transitions, and checkbox counts

## Scope

- Does NOT write new acceptance criteria or modify spec prose beyond frontmatter and checkboxes.
- Does NOT validate whether criteria are actually met — the caller is responsible for confirming this before invoking.
- Does NOT create spec files. Use the appropriate specification agent for new specs.

## Examples

**Trigger:** After deploying the absenteeism feature, user runs `/spec.status absence-api deployed`.

**Output:** `specs/functional/absence-api.md` updated — `status: deployed`, `deployed: 2026-05-21T00:00:00Z`, 6 checkboxes `[ ]→[x]`, 1 checkbox left as `[ ]` (unverified).

## Gotchas

- Some spec files carry `status: draft` even after partial implementation. Verify all criteria under the new status are actually met before updating. (Observed in HIM Corporate session 005: `smaqit plan` returned empty because `STK-BACKEND` was still `implemented` when new criteria had not been addressed.)
- `deployed` is a **datetime**, not a date. Always include the time component (`T00:00:00Z` is acceptable if the exact time is unknown).
- The `[!]` marker is a project convention for deferred or untestable criteria. Always add a brief inline note when using it.

## Completion

- [ ] Target spec file(s) identified
- [ ] Frontmatter fields updated (`status`, `updated`, `deployed` as applicable)
- [ ] Acceptance criteria checkboxes updated with confirmation
- [ ] Report delivered (file list + transitions + checkbox counts)

## Failure Handling

| Situation | Action |
|-----------|--------|
| No file specified | Ask the user for the target spec or phase before proceeding |
| Criterion status is uncertain | Leave checkbox as `- [ ]`; report it as uncertain in the output |
| File does not have frontmatter | Add a minimal frontmatter block before the first heading |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
