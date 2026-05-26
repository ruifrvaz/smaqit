# Skill Definition: smaqit.spec-status-update

## Identity
- **Name:** smaqit.spec-status-update
- **Version:** 1.0.0
- **Description:** Update spec file frontmatter (`status`, `deployed`, `updated`) and acceptance criteria checkboxes after implementation or deployment is confirmed. Keeps specs in sync with the live codebase without requiring a full spec agent run.

## Steps
1. Read the target spec file(s). If no file is specified, ask the user which spec or phase to update (e.g. "all functional specs", "all infra specs", "specs/functional/diagnostic-flow.md").
2. For each spec file:
   a. Update frontmatter fields:
      - `status` → set to the value appropriate for the phase (`draft`, `implemented`, `deployed`)
      - `updated` → set to today's date in ISO 8601 format (`YYYY-MM-DD`)
      - `deployed` → set to ISO 8601 datetime if transitioning to `deployed` (e.g. `2026-04-06T00:00:00Z`)
   b. Update acceptance criteria checkboxes:
      - `- [ ]` → `- [x]` for criteria confirmed as met
      - `- [ ]` → `- [!]` for criteria that are explicitly deferred or untestable (annotate inline)
      - Do NOT change a checkbox unless the criterion is confirmed. Leave uncertain criteria as `- [ ]`.
3. Report: list each file updated with the old status → new status transition and the number of checkboxes changed.

## Output
- Updated spec file(s) with current frontmatter and accurate checkboxes

## Scope
- Does NOT write new acceptance criteria or modify spec content beyond frontmatter and checkboxes.
- Does NOT validate whether criteria are actually met — the caller is responsible for confirming this before invoking.
- Does NOT create spec files. Use the appropriate specification agent for that.

## Gotchas
- Some spec files use `status: draft` even after partial implementation (seen in HIM Corporate session 005 — `smaqit plan` returned empty because `STK-BACKEND` was still `implemented` when new criteria had not been addressed). Verify that all criteria under the new status are actually met before updating.
- `deployed` is a datetime, not a date. Include the time component (`T00:00:00Z` is acceptable if exact time is unknown).
- The `[!]` marker for deferred/untestable criteria is a project convention. Always add a brief inline note when using it (e.g., `- [!] FUN-HOME-ANIM-007 — aesthetic quality, requires manual review`).

## Completion
- [ ] Target spec file(s) identified
- [ ] Frontmatter fields updated (`status`, `updated`, `deployed` as applicable)
- [ ] Acceptance criteria checkboxes updated with confirmation
- [ ] Report delivered (file list + transitions + checkbox counts)

## Failure Handling
| Situation | Action |
|-----------|--------|
| No file specified | Ask the user for the target spec or phase |
| Criterion status is uncertain | Leave checkbox as `- [ ]`; report it as uncertain in the output |
| File does not have frontmatter | Add a minimal frontmatter block before the first heading |

## Examples
**Input:** After deploying the absenteeism feature, user runs `/spec.status absence-api deployed`.
**Output:** `specs/functional/absence-api.md` updated — `status: deployed`, `deployed: 2026-05-21T00:00:00Z`, 6 checkboxes `[ ]→[x]`, 1 checkpoint left as `[ ]` (unverified).
