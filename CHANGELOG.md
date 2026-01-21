# Changelog

All notable changes to smaqit will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.2-beta] - 2026-01-21

### Added
- Compilation file architecture: 8 compilation files (5 layers + 3 phases) for L0→L1 transformation (Task 068)
- Wiki documentation: `why-no-system-actor.md` and `why-non-functional-requirements.md`
- Source L0 Principles table in all compilation files for traceability

### Changed
- **BREAKING:** System Actor pattern removed from Business layer (Task 068)
- Business layer boundary enforcement strengthened with concept-based directives
- Compilation files now contain pure L1 directives without inline L0 Source citations
- Business layer MUST NOT directive refined: "Describe HOW features work" (replaces word blacklist)
- NFR terminology changed from "system property advocates" to "non-functional requirement stakeholders"
- Agent-L1 updated to enforce clean directive compilation
- Specification agent template simplified (Actor Concept placeholder removed)

### Fixed
- Business specs no longer leak functional/stack concerns through System Actor pattern
- Behavioral verbs directive now focuses on concept boundary (HOW vs WHAT) instead of vocabulary blacklist

## [0.6.0-beta] - 2026-01-17

### Added
- Troubleshooting documentation for multi-agent workflow context pollution (Task 056)

### Changed
- Specification agents now reset acceptance criteria checkboxes when modifying requirements (Task 060)
- Implementation agents update all referenced spec frontmatter, not just target layer (Tasks 061, 063)
- Validation agent generates executable test artifacts for CI/CD automation (Task 062)
- Implementation agents execute CLI as first action to determine which specs to process (Tasks 049, 051, 052)
- Coverage layer redesigned with dual-input model: test requirements + upstream criteria (Task 050)
- Agent Role sections refined to 3-component structure for clarity (Task 056)
- Foundation Reference pattern unified across all specification layers (Task 055)

### Fixed
- Validation agent frontmatter updates now apply to all validated specs across all layers (Task 053)
- CLI directive ambiguity resolved in all implementation agents (Tasks 049, 051, 052)

### Removed
- Context pollution verbal statements pattern (replaced with structured Role sections) (Task 056)

## [0.5.0-beta] - 2026-01-03

### Added
- **Stateful Specifications** (Task 014)
  - YAML frontmatter state tracking in all specs (id, status, created, timestamps, prompt_version)
  - Spec lifecycle states: draft → implemented → deployed → validated → failed/deprecated
  - Acceptance criteria checkbox updates: `[ ]` → `[x]` (passed) or `[!]` (failed)
  - Phase reports generated in `.smaqit/reports/` directory
  - Stale spec detection via prompt_version tracking (git commit hash)
  - Wiki documentation for stateful specifications and stale management workflows

- **Incremental Processing** (Task 047)
  - `smaqit plan --phase=[develop|deploy|validate]` command
    - Outputs spec file paths requiring processing (one per line)
    - Default: returns only specs with `status: draft` or `status: failed`
    - `--regen` flag: returns all specs regardless of status
  - Implementation agents now skip already-completed specs
  - Frontmatter as single source of truth (removed dual state system)
  - CLI scans specs on-demand and aggregates status
  - Strict phase completion rules: requires ALL layers present + ALL specs at target status

### Changed
- Agents refactored to directive-based instructions (pure MUST/MUST NOT rules)
  - Removed procedural "State-Based Processing" workflows
  - Simplified agent instructions for better LLM interpretation
- Phase completion detection now requires all layers present
  - Develop: business + functional + stack specs required
  - Deploy: infrastructure specs required
  - Validate: coverage specs required
- Framework documentation updated to remove example pollution
  - Generic placeholders ([ID], [CONCEPT]) replace specific examples
  - Templates remain abstract and reusable

### Removed
- state.json aggregate state file (replaced by on-demand CLI scanning)
- Dual state system complexity (frontmatter is sole source of truth)
- Example pollution from templates (BUS-LOGIN-001, etc.)

## [0.4.2-beta] - 2026-01-02

### Added
- Intelligent next-step suggestions in `smaqit status` command
  - Progressive guidance based on actual spec file presence
  - Suggests missing spec layers before implementation (business → functional → stack)
  - Only suggests `/smaqit.development` when all Phase 1 specs exist
  - Phase-aware suggestions for infrastructure and coverage layers

### Fixed
- State.json phase ordering now consistent (develop → deploy → validate)
- Status command no longer suggests premature implementation steps

## [0.4.1-beta] - 2026-01-02

### Added
- One-liner installation script (`install.sh`)
  - Platform detection for Linux, macOS, and Windows
  - Automatic installation to `~/.local/bin`
  - Version selection via `SMAQIT_VERSION` environment variable (latest/prerelease/vX.Y.Z)
  - Installation verification and PATH checking
- Standard CLI flag support
  - `--version` and `-v` flags now work alongside `version` subcommand
  - `--help` and `-h` flags work alongside `help` subcommand
  - Consistent with standard tools (go, python, etc.)

### Fixed
- Install script stdout contamination from info messages during download
- Repository visibility (made public for install script access)

## [0.4.0-beta] - 2026-01-01

### Added
- GitHub Actions workflow for automated releases (Task 022)
  - Automatic builds for Linux, macOS (Intel/ARM), Windows
  - SHA256 checksums generation
  - Release notes extracted from CHANGELOG.md
- Manual workflow dispatch for releases via GitHub UI
- Changelog management system using session history
  - `/changelog.update` agent reads `docs/history/` and updates CHANGELOG.md
  - AI-managed changelog following Keep a Changelog format
  - Simplified release process documented in README

## [0.3.0] - 2025-12-28

### Added
- Explicit scope boundaries for all agents (Task 041)
  - Agents now enforce layer/phase boundaries with Stop → Respond → Suggest pattern
  - Prevents agents from executing out-of-scope work
- User vs agent documentation distinction (Task 040)
  - New wiki document explaining separation between agent-facing specs and user-facing docs
  - Guidelines for what content belongs in framework vs wiki
- Agent handover guidance (Task 039)
  - Agents provide clear next steps when completing work
- State.json validation in validate command (Task 038)
  - Defensive validation for phase completion tracking
- Phase-first workflow clarification (Task 037)
  - Updated PHASES.md to emphasize phase-first as recommended approach
- Use case identifiers to business specs (Task 034)
  - Business specs now include UC-XXX identifiers for traceability
- Nested status display (Task 035)
  - `smaqit status` shows layers grouped under phases

### Changed
- Prompt naming consistency (Task 044)
  - Updated all references from `.develop`/`.deploy`/`.validate` to `.development`/`.deployment`/`.validation`
  - Installer messages, help text, and documentation now consistent
- Implementation prompts simplified (Task 029)
  - Renamed for consistency with agent names
  - Reduced to minimal orchestration inputs
- Session and task commands moved to prompts (Task 030)
  - `/session.recap`, `/session.wrap`, `/task.*` now in `.github/prompts/`

### Fixed
- Prompt name references in installer and documentation (Task 044)
- State.json phase ordering corrected (Task 033)

## [0.2.0] - 2025-12-27

### Added
- Framework embedding at installation (Task 015)
  - Framework files now embedded in agents for self-contained execution
  - Removed runtime framework bundling

### Changed
- Documentation architecture refined (Task 028)
  - Separated agent-facing instructions from human-readable rationale
  - Framework files contain only execution instructions
  - Wiki contains context and design decisions

## [0.1.0] - 2025-12-20

### Added
- Prompt architecture and integration (Task 026)
  - Prompts as input records capturing user requirements
  - Free-style natural language with suggested structure
- User testing agent (Task 024)
  - Automated end-to-end testing capability
- Installer CLI implementation (Task 023)
  - `smaqit init`, `smaqit status`, `smaqit validate` commands
  - Cross-platform Go installer with embedded files

### Changed
- Infrastructure layer accepts cross-cutting input (Task 016)
  - Infrastructure specs can reference all Phase 1 specs for coherence

## [0.0.1] - 2025-12-18

### Added
- Complete specification templates (Tasks 017-021)
  - Business, Functional, Stack, Infrastructure, Coverage templates
- Agent templates (Tasks 002-003)
  - Specification agent template
  - Implementation agent template
- Framework documentation split (Task 013)
  - SMAQIT.md split into LAYERS.md, PHASES.md, TEMPLATES.md, AGENTS.md, ARTIFACTS.md
- Cross-platform build system (Task 012)
  - Makefile with `build`, `build-all`, `install`, `uninstall` targets
  - Support for Linux, macOS (Intel/ARM), Windows

### Changed
- Layer independence principle established (Task 007)
  - Each layer's prompt file is sole source of requirements
  - Upstream layers provide context, not requirements

[Unreleased]: https://github.com/ruifrvaz/smaqit/compare/v0.6.0-beta...HEAD
[0.6.0-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.5.0-beta...v0.6.0-beta
[0.5.0-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.4.2-beta...v0.5.0-beta
[0.4.2-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.4.1-beta...v0.4.2-beta
[0.4.1-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.4.0-beta...v0.4.1-beta
[0.4.0-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.3.0...v0.4.0-beta
[0.3.2-beta]: https://github.com/ruifrvaz/smaqit/compare/v0.2.0...v0.3.2-beta