# Prompt Deprecation Task Creation

## Metadata

- **Date:** 2026-05-09
- **Focus:** Assess and create a comprehensive task to fully deprecate the prompts feature
- **Tasks referenced:** 081 (created)

## Actions Taken

- Ran session start — loaded project context, history, active tasks
- Identified v0.9.x tags and Copilot agent branches fetched from remote during session
- Performed full surface-area assessment of prompts across the repo: source files, installer, framework, agents, wiki, README, Go code
- Designed the replacement input model: agents read from current session context (including compacted blocks) or open tasks; assessment skill applied when input is ambiguous or insufficient
- Created Task 081 (Deprecate Prompts) — 9-step standalone task covering deletions, framework rewrites, agent rewrites, installer Go changes, wiki cleanup, build verification, and smoke test
- Updated PLANNING.md with task 081 as High priority
- Committed and pushed both files (`942d5cd`)
- Pulled remote changes that were ahead — fetched v0.9.0, v0.9.1 tags and several Copilot agent branches
- Brief discussion on `git pull --rebase` vs `git pull --merge` safety

## Problems Solved

- **Push rejected (exit 128):** SSH passphrase timed out on first attempt; resolved by retrying push after `git pull --rebase`
- **Remote divergence:** Copilot agents had pushed to `main` between session start and push; rebased cleanly, no conflicts

## Decisions Made

- **No backward compatibility** — full deprecation, no migration path, no stray references left behind (except history and tasks)
- **Replacement input model** — session context + assessment skill replaces prompt files; no new file-based input mechanism introduced
- **`prompt_version` removed** — from spec frontmatter struct (`spec.go`), all spec templates, and all framework references
- **`framework/PROMPTS.md` deleted** — not refactored, entirely eliminated
- **Wiki files deleted vs edited** — 8 files deleted outright (exist solely for prompts concept); 14 files edited to clean up references
- **`.github/prompts/` in smaqit repo untouched** — those are session/task management workflow prompts, a completely different category
- **`installer/framework/` kept in sync** — task explicitly instructs copying source files to installer mirror after each edit rather than editing both independently

## Files Modified

| File | Action |
|------|--------|
| `.smaqit/tasks/081_deprecate-prompts.md` | Created |
| `.smaqit/tasks/PLANNING.md` | Added task 081 to Active table |

## Next Steps

- Execute Task 081 (Deprecate Prompts) in a dedicated session — it is standalone and ready to pick up
- Review the Copilot agent branches fetched from remote (`copilot/create-new-release`, `copilot/promote-beta-to-stable`, etc.) — determine if any should be merged or closed
- Task 070 (E2E Boundary Enforcement Validation) remains the highest-priority unblocked task after 081

## Session Metrics

- Tasks created: 1 (081)
- Files created: 1 (task file)
- Files modified: 1 (PLANNING.md)
- Commits: 1 (`942d5cd`)
