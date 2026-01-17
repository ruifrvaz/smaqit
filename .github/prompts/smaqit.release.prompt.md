---
name: smaqit.release
description: Release new smaqit version with automated workflow
agent: smaqit.release
---

# Release Prompt

Orchestrate a new smaqit release: update CHANGELOG.md from session history, sync version strings, commit, tag, and push to trigger automated build/release.

## Release Branch (Optional)

<!-- Default: main -->
<!-- Specify if releasing from a different branch -->

main

## Date Range for Changelog (Optional)

<!-- Limit session history analysis to specific date range -->
<!-- Example: 2025-12-20 to 2026-01-17 -->

[Leave empty to process all sessions since last release]

## Additional Context

<!-- Any special notes about this release -->
<!-- Example: "Breaking changes in CLI, emphasize in changelog" -->

[Optional additional guidance for the agent]
