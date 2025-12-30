# Clean Up Old Prompt Name References

**Status:** Not Started  
**Created:** 2025-12-28

## Description

Installation code still references old prompt names from before task 029 renaming:
- `smaqit.develop` → should be `smaqit.development`
- `smaqit.deploy` → should be `smaqit.deployment`
- `smaqit.validate` → should be `smaqit.validation`

Need to update all references in installer code, documentation, and any other locations to use the new consistent naming convention.

## Acceptance Criteria

- [ ] All references to `smaqit.develop` updated to `smaqit.development`
- [ ] All references to `smaqit.deploy` updated to `smaqit.deployment`
- [ ] All references to `smaqit.validate` updated to `smaqit.validation`
- [ ] Installer builds successfully with updated references
- [ ] `smaqit init` creates correct prompt files with new names
- [ ] No remaining references to old names in codebase (verified via grep)

## Notes

**Likely locations:**
- `installer/main.go` - Prompt file copying logic
- `README.md` - Usage examples
- Framework files - Any command references
- Documentation files - Examples or instructions

**Name mapping:**
| Old Name | New Name | Reason |
|----------|----------|--------|
| `smaqit.develop` | `smaqit.development` | Consistency with agent name |
| `smaqit.deploy` | `smaqit.deployment` | Consistency with agent name |
| `smaqit.validate` | `smaqit.validation` | Consistency with agent name |

This naming was established in task 029 for consistency between prompts and agents.
