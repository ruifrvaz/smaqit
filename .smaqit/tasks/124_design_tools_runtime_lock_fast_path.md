---
status: PR Open
pr: 93
created: "2026-10-06"
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

- [ ] The fix is merged and released
- [ ] A leftover lock no longer delays an MCP server start with a valid bundle (the new test)

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| installer/design_tools.go, installer/design_test.go | Modify (done on branch) |

## Notes

- Workaround until released: when `smaqit-plantuml` fails to connect and no `smaqit` process runs,
  remove `.smaqit/tools/plantuml/.<bundle>.lock` and reconnect.
