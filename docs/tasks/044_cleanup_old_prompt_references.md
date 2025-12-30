# Clean Up Old Prompt Name References

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2025-12-30

## Description

Installation code still references old prompt names from before task 029 renaming:
- `smaqit.develop` → should be `smaqit.development`
- `smaqit.deploy` → should be `smaqit.deployment`
- `smaqit.validate` → should be `smaqit.validation`

Need to update all references in installer code, documentation, and any other locations to use the new consistent naming convention.

## Acceptance Criteria

- [x] All references to `smaqit.develop` updated to `smaqit.development`
- [x] All references to `smaqit.deploy` updated to `smaqit.deployment`
- [x] All references to `smaqit.validate` updated to `smaqit.validation`
- [x] Installer builds successfully with updated references
- [x] `smaqit init` creates correct prompt files with new names
- [x] No remaining references to old names in codebase (verified via grep)

## Notes

**Actual locations updated:**
- `installer/main.go` - Updated all help text, init messages, and status next steps (lines 176-178, 188, 267)
- `docs/tasks/001_create_smaq_commands_file.md` - Added historical notes about Task 029 rename
- `docs/tasks/023_implement_installer_cli.md` - Added notes about prompt name changes
- `docs/tasks/029_simplify_implementation_prompts.md` - Updated example code with final names
- `docs/history/009_installer_refinements_2025-12-19.md` - Added note about subsequent Task 029 rename
- `docs/history/010_prompts_and_testing_agent_2025-12-20.md` - Added notes about renaming

**Locations NOT requiring changes:**
- `README.md` - CLI commands (`smaqit develop`, `smaqit deploy`, `smaqit validate`) are correct—they're shell commands, not prompt invocations
- Framework files - Already used correct names (`.development`, `.deployment`, `.validation`)

**Name mapping:**
| Old Name | New Name | Reason |
|----------|----------|--------|
| `smaqit.develop` | `smaqit.development` | Consistency with agent name |
| `smaqit.deploy` | `smaqit.deployment` | Consistency with agent name |
| `smaqit.validate` | `smaqit.validation` | Consistency with agent name |

This naming was established in task 029 for consistency between prompts and agents.

## Completion Summary

All old prompt name references have been updated. The installer now consistently displays the correct names (`.development`, `.deployment`, `.validation`) in help text, init messages, and status output. Historical documentation files updated with appropriate notes explaining the Task 029 rename for future reference.
