---
status: Not Started
created: "2026-10-11"
---

# Design Tools: Install Path Still Holds the Runtime Lock Past MCP Connect Timeouts

## Description

Follow-up from task 124 (released in v3.6.1). Task 124 stopped normal `smaqit mcp plantuml` starts
from taking the PlantUML runtime installation lock when the bundle at
`.smaqit/tools/plantuml/<designBundleVersion>/` is already valid. A start no longer waits behind a
lock orphaned by a killed process, which was the common failure: an editor reload killing an MCP
server mid-start.

The install path itself was left unchanged (`materializeDesignTools` →
`acquireDesignToolLock` in `installer/design_tools.go`). It still runs whenever the bundle is
missing or invalid: in a new Git worktree, which deliberately doesn't inherit the ignored runtime,
or on the first start after a release changes `designBundleVersion`. Installing extracts and hashes
about 3,900 files (67 MB) while holding the lock, which is the longest window in which a client
can kill the process. The lock is an `os.Mkdir` directory released by `defer`, so a killed
installer leaves it behind. Each later start then waits up to `designToolLockWait` (60 s) and only
breaks the lock once it is older than `designToolLockStale` (5 min). Claude Code gives up
connecting after 30 s, so the server reports `CONNECT_TIMEOUT` for up to 5 minutes after any
interrupted install.

Secondary, same code path: `ensureDesignTools` calls `validateMaterializedDesignTools` again right
after `materializeDesignTools` has already validated the same directory. The full integrity check
(manifest, package-lock hash, every file's hash, and an inventory walk) therefore runs twice on
every start.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Go, Model Context Protocol, Node.js
**Platforms/Environments:** Linux, WSL2, macOS, Windows
**Features/Integrations:** PlantUML MCP server startup, design-tools runtime installation, Claude Code MCP connect timeout
**Versions/Constraints:** smaqit v3.6.1, Claude Code 30 s MCP connect timeout, Go 1.25

## Design Decisions

TBD — to be confirmed during assessment. Candidate approaches:
- **Owner liveness in the lock:** write the owning PID (and a start timestamp, to guard against PID reuse) into the lock directory; a waiter breaks the lock immediately when the owner is no longer alive. Needs a portable liveness check (signal 0 on Unix; Windows differs).
- **Bounded wait below common MCP timeouts:** shorten `designToolLockWait` so a blocked start fails fast with a clear `DESIGN-TOOLCHAIN-UNAVAILABLE` message instead of being killed silently by the client. Doesn't by itself shorten the up-to-5-minute stale window.
- **Shrink the critical section:** extract into the private temporary directory without the lock, and take the lock only for the final rename swap, so a kill during extraction can't orphan it.
- **Single validation per start:** drop the redundant `validateMaterializedDesignTools` call in `ensureDesignTools`, or have `materializeDesignTools` report that it has already validated.

## Implementation Steps

1. Assess the candidate approaches above (they can combine); record the chosen design under Design Decisions.
2. Implement the change in `installer/design_tools.go`, keeping the atomic-rename guarantee that a reader without the lock never sees a half-written bundle.
3. Add a regression test where an install-path lock is orphaned (no bundle installed, lock present, owner gone), and the next start installs and returns well within 30 s.
4. Keep `TestMaterializeDesignToolsIsConcurrentSafe` and `TestMaterializeDesignToolsIgnoresLockWhenBundleIsValid` passing.
5. Remove the duplicate validation, and confirm that a start with a valid bundle hashes the bundle once.
6. Run `go test`/`go vet` and `make smoke-test`; then, in a fresh worktree, kill an MCP start mid-install and confirm the next start connects.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] A lock left by an installer killed mid-install no longer blocks the next `smaqit mcp plantuml` start beyond Claude Code's 30 s connect timeout
- [ ] A live concurrent installer is still respected: two simultaneous starts never both extract over the same destination, and neither sees a half-written bundle
- [ ] A regression test covers the orphaned install-path lock case
- [ ] A start with an already-valid bundle runs the full integrity check once, not twice
- [ ] Existing design-tools tests, `go vet`, and `make smoke-test` pass

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
| installer/design_tools.go | Modify — lock acquisition/ownership, redundant validation |
| installer/design_test.go | Modify — orphaned install-path lock regression test |
| CHANGELOG.md | Modify — version entry |

## Notes

- Workaround until fixed: if `smaqit-plantuml` fails to connect and no `smaqit mcp plantuml` process is running for the project, remove `.smaqit/tools/plantuml/.<designBundleVersion>.lock` and reconnect.
- Release PR titles must match `Prepare release vX.Y.Z`, or `post-merge-release.yml` skips (task 124 needed a second PR for this).
