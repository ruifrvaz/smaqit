---
status: Completed
created: "2026-10-06"
completed: "2026-10-11"
---

# Design Tools: Skip the Runtime Install Lock When the Bundle Is Already Valid

## Description

Every `smaqit mcp plantuml` start took the PlantUML runtime installation lock
(`installer/design_tools.go`, `acquireDesignToolLock`), even when the bundle was already materialized
and valid. A lock left by a process killed mid-start (an editor reload killing MCP servers) blocked
every new start for up to `designToolLockStale` (5 min), while each start gave up after
`designToolLockWait` (60 s) — well past Claude Code's 30 s MCP connect timeout, so the server showed
as "failed to connect". Reproduced in a downstream project on 2026-10-06, twice after editor reloads.

A fix exists on local branch `fix/design-tools-lock-fast-path` (commit 8fd2e93, worktree
`../smaqit-wt-fix-design-tools-lock-fast-path`): validate the destination lock-free first (it is only
ever replaced by an atomic rename), take the lock only to install, re-check under the lock. New test
`TestMaterializeDesignToolsIgnoresLockWhenBundleIsValid` fails at 60.4 s without the fix and passes
in 0.4 s; the full installer suite passes.

## Issue Triage Context

**Mode:** Skip
**Technologies:** None
**Platforms/Environments:** None
**Features/Integrations:** None
**Versions/Constraints:** Go installer, MCP stdio server startup

## Design Decisions

- **Lock-free fast path for a valid bundle**; the lock protects installation only.
- Further hardening (owner PID in the lock for faster stale detection) is optional, not required.

## Implementation Steps

1. Push `fix/design-tools-lock-fast-path` and open a PR (operator: the agent's token cannot write
   to this repository).
2. Release; reinstall in downstream projects.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [x] The fix is merged and released
- [x] A leftover lock no longer delays an MCP server start with a valid bundle (the new test)

## Findings

**Implementation approach:**
- Fix merged via PR #92 (fix branch, no task worktree); released separately via PR #93 (`Prepare release v3.6.1`), which carried only the CHANGELOG section
- Regression test fails at 60.4 s against v3.6.0 code and passes in 0.4 s; a build of the merged code answered MCP `initialize` in under 1 s with an orphaned lock present, while the installed v3.6.0 binary was still blocked at 35 s

**Decisions made:**
- Released as PATCH v3.6.1 from a dedicated `release/v3.6.1` branch cut from `origin/main`, keeping local-only bookkeeping commits out of the release PR
- Owner-PID lock hardening left out of scope, per this task's Design Decisions

**Blockers encountered:**
- PR #92's title did not match `Prepare release vX.Y.Z`, so `post-merge-release.yml` skipped and no release was published until PR #93
- Branch push initially failed: the HTTPS token was read-only, and the desktop SSH agent refused to sign with the account key; the push succeeded over HTTPS once write access was restored

**Follow-up identified:**
- The install path still takes the lock: a start killed mid-install (fresh worktree, or first start after a bundle-version bump) can still leave a lock that blocks starts for up to 60 s against a 30 s MCP connect timeout; owner-PID liveness or a shorter wait would close it
- `ensureDesignTools` re-validates the bundle right after `materializeDesignTools` already did, hashing about 3,900 files twice on every start

## Files to Create / Modify

| File | Action |
|------|--------|
| installer/design_tools.go, installer/design_test.go | Modify (done on branch) |

## Notes

- Workaround until released: when `smaqit-plantuml` fails to connect and no `smaqit` process runs,
  remove `.smaqit/tools/plantuml/.<bundle>.lock` and reconnect.
