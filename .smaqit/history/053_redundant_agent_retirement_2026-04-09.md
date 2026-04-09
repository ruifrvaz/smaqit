# Redundant Agent Retirement

## Metadata

- **Date:** 2026-04-09
- **Focus:** Agent inventory cleanup — identify and retire redundant agents
- **Tasks referenced:** None (housekeeping)

## Actions Taken

### 1. Release agent redundancy audit

Compared all three release agents:
- `smaqit.release` — monolithic 7-step agent with all workflow logic embedded
- `smaqit.release.local` — skill-delegating orchestrator (interactive, direct git)
- `smaqit.release.pr` — skill-delegating orchestrator (CI/CD, auto-confirm, PR-based)

Conclusion: `smaqit.release` was superseded wholesale when the release workflow was decomposed into 4 skills (`release-analysis`, `release-approval`, `release-prepare-files`, `release-git-local/pr`). The `.local` and `.pr` variants are correct and non-redundant.

**Retired:** `smaqit.release.agent.md`
**Updated:** `copilot-instructions.md` quick commands to reference `.local` and `.pr` with context on when to use each.

### 2. doc-helper vs smaqit.qa audit

Compared the two read-only documentation Q&A agents:
- `doc-helper` — generic project docs reader, user-provided URLs, no backing task
- `smaqit.qa` — framework-aware, hard-coded smaqit GitHub URLs, full redirect map to all layer/phase agents

Both serve the same function (read-only doc Q&A). `smaqit.qa` is the more complete and framework-specific variant. `doc-helper` was a one-shot creation with no task, no history entry.

**Bug found:** `smaqit.qa` had `edit` in its `tools` list despite declaring `MUST NOT: Modify any files`.
**Fixed:** `smaqit.qa` tools → `['read', 'search', 'fetch']` (consistent with `doc-helper` and its own directives).
**Retired:** `agents/doc-helper.agent.md`

## Decisions Made

- **Retire over merge:** Both cases opted for retirement rather than merging, since the surviving agent already covered the full scope.
- **Skill decomposition validates agent reduction:** The release skills architecture is the reason `smaqit.release` became obsolete — the split was correct by design.
- **`smaqit.qa` is the canonical doc agent for user projects** — covers both smaqit framework docs and general project docs via local file reads.

## Problems Solved

- `copilot-instructions.md` referenced the deleted `smaqit.release.prompt.md` and the monolithic release agent — now updated to the current `.local`/`.pr` pair.
- `smaqit.qa` tools list was inconsistent with its own directives (security/hygiene issue: edit tool available to a read-only agent).

## Files Modified

| File | Change |
|------|--------|
| `.github/agents/smaqit.release.agent.md` | Deleted (retired) |
| `.github/copilot-instructions.md` | Updated quick commands to `.local`/`.pr` |
| `agents/smaqit.qa.agent.md` | Fixed tools: `edit` → `read` |
| `agents/doc-helper.agent.md` | Deleted (retired) |

## Commits This Session

| SHA | Message |
|-----|---------|
| `9b818f9` | `chore(meta): retire smaqit.release agent, superseded by .local/.pr variants` |
| `488a0f6` | `fix(agents): correct smaqit.qa tools list to read-only` |
| `42db7c2` | `chore(agents): retire doc-helper agent` |

## Next Steps

- **Commit and push pending file:** `.github/workflows/copilot-setup-steps.yml` (fix for `/releases/latest` returning 404 on pre-releases — applied last session, still uncommitted)
- **Push all commits to origin:** 8 commits ahead of `origin/main` as of session end (5 from prior session + 3 from this session)
- **Decide release version:** Promote to `0.9.0` stable or `0.8.3-beta`? Then populate `CHANGELOG.md [Unreleased]`, bump version, tag, push

## Session Metrics

- Duration: ~1 session
- Agents retired: 2 (`smaqit.release`, `doc-helper`)
- Bugs fixed: 1 (`smaqit.qa` tools list)
- Files modified: 4
- Commits created: 3
