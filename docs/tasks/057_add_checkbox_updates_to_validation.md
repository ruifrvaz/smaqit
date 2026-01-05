# Task 057: Add Checkbox Updates to Validation Agent

**Status:** new  
**Priority:** Low  
**Created:** 2026-01-05  
**Related:** Task 048 (E2E Testing), Issue 8

## Problem

After successful validation, acceptance criteria checkboxes in specs remain unchecked `- [ ]` instead of marked `- [x]` (pass) or `- [!]` (fail/untestable).

**Impact:**
- Loss of visual verification indicator—cannot see which criteria passed at file level
- Validation report shows all tests passed, but specs don't reflect this
- Manual review burden—must cross-reference report to specs to determine pass/fail state
- Undermines "spec as living document" principle

**Severity:** Medium—reduces usability but validation report provides same information.

## Objective

Decide on approach for checkbox updates (automated vs manual vs CLI command), then implement chosen solution.

## Decision Required

**Three options:**

1. **Option A: Automated (Validation Agent)**
   - Validation agent updates checkboxes during validation execution
   - Pros: Specs immediately reflect validation state
   - Cons: Adds complexity to agent, requires parsing/editing multiple spec files

2. **Option B: Manual (Documentation)**
   - Document that checkbox updates are manual user action after reviewing validation report
   - Pros: Simplest, no implementation needed
   - Cons: Manual burden, specs don't auto-reflect validation state

3. **Option C: CLI Command (Future)**
   - Add `smaqit update-checkboxes` command that parses validation report and updates specs
   - Pros: Automation without agent complexity, user controls when to update
   - Cons: Requires CLI development, additional command to remember

**Recommendation:** Option C (CLI command) in future release (v0.6.0). Option B (documentation) for v0.5.0.

## Acceptance Criteria (Option B - Documentation)

- [ ] Documented in `README.md` or wiki that checkbox updates are manual
- [ ] Provided guidance on how to interpret validation report and update checkboxes
- [ ] Explained checkbox meanings: `[x]` = pass, `[!]` = fail/untestable, `[ ]` = not yet validated
- [ ] Mentioned future automation (v0.6.0) as CLI command

## Acceptance Criteria (Option C - CLI Command, future)

- [ ] Implemented `smaqit update-checkboxes` command
- [ ] Command reads validation report from `.smaqit/reports/validation-phase-report-*.md`
- [ ] Command parses requirement IDs and pass/fail results
- [ ] Command updates all spec files with appropriate checkbox states
- [ ] Command handles untestable criteria (marks as `[!]`)
- [ ] Command provides summary of updates made
- [ ] Command is idempotent (can be run multiple times safely)

## Implementation Plan (Option B - Documentation, v0.5.0)

1. **Add documentation to README or wiki:**
   - Section: "Working with Validation Results"
   - Content:
     - Validation report shows pass/fail per requirement
     - Checkbox updates are manual user action
     - How to update: Review report, mark `[x]` for pass, `[!]` for fail/untestable
     - Future: v0.6.0 will add `smaqit update-checkboxes` for automation

2. **Optional: Update Validation agent output:**
   - Add note to validation report: "To update spec checkboxes, manually review this report and mark criteria as [x] (pass) or [!] (fail). Future releases will automate this."

## Implementation Plan (Option C - CLI Command, v0.6.0)

1. **Design CLI command:**
   - Command: `smaqit update-checkboxes [--report PATH] [--dry-run]`
   - Default report: Most recent validation report in `.smaqit/reports/`
   - Dry-run: Show what would be updated without modifying files

2. **Implement parser:**
   - Parse validation report to extract requirement IDs and results
   - Map: `BUS-GREETING-001` → `PASS` or `FAIL` or `UNTESTABLE`

3. **Implement spec updater:**
   - Scan all specs in `specs/`
   - For each requirement ID in validation results:
     - Locate checkbox line in spec file
     - Update checkbox: `- [ ]` → `- [x]` (pass) or `- [!]` (fail)
   - Preserve all other content unchanged

4. **Add to CLI:**
   - Add command handler in `installer/main.go`
   - Add to help output

## Files to Modify (Option B - Documentation)

- `README.md` or `docs/wiki/working-with-validation.md` (add documentation)
- Optional: `agents/smaqit.validation.agent.md` (add note to report output)

## Files to Modify (Option C - CLI Command, future)

- `installer/main.go` (add command handler)
- `installer/spec.go` (add checkbox update logic)
- `README.md` (document command usage)

## Testing

**Option B (Documentation):**
1. Read documentation
2. Verify instructions are clear and actionable

**Option C (CLI Command):**
1. Run validation on test project
2. Run `smaqit update-checkboxes`
3. Verify all spec checkboxes updated correctly
4. Verify pass/fail/untestable states correct
5. Run command again (idempotence test)
6. Verify no duplicate updates or errors

## Estimated Effort

- **Option B (Documentation):** 30 minutes
- **Option C (CLI Command):** 2-4 hours

## Dependencies

- Task 053 (Validation frontmatter updates) should be completed first
- Option C depends on validation report format remaining stable

## Blocks

None (this is a usability enhancement, not blocking release)

## Related Tasks

- Task 053: Fix Validation Frontmatter Updates (related validation agent fix)
- Task 048: E2E Agent Workflow Testing (discovered this issue)

## Notes

**Validation report is sufficient:** Validation report already provides complete pass/fail information. Checkbox updates are UX enhancement for "spec as living document" principle, not functional requirement.

**Tradeoff consideration:** Automated checkbox updates improve usability but add complexity. Validation agent already has many responsibilities (execute tests, generate report, update frontmatter). Adding checkbox updates may be scope creep. CLI command is cleaner separation of concerns.

**Expected checkbox states:**

```markdown
## Acceptance Criteria

- [x] BUS-GREETING-001: Application displays ASCII art (validated via COV-GREETING-APP-001)
- [x] BUS-GREETING-002: Application shows catchphrase (validated via COV-GREETING-APP-002)
- [!] BUS-GREETING-008: Users recognize Mario (untestable - manual review required)
```

**User workflow with CLI command:**
```bash
# Run validation
# (Validation agent creates report)

# Update checkboxes based on validation results
smaqit update-checkboxes

# Inspect specs - checkboxes now reflect validation state
cat specs/business/uc1-greeting.md
```

**Recommendation rationale:** Option C (CLI command) provides automation without agent complexity. User controls when to update (after reviewing report). Aligns with smaqit principle of CLI as primary interface for project state management.

**v0.5.0 decision:** Implement Option B (documentation) for immediate release. Schedule Option C (CLI command) for v0.6.0 as enhancement.
