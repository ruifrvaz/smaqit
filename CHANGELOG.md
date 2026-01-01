# Changelog

All notable changes to smaqit will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GitHub Actions workflow for automated releases (Task 022)
- Manual workflow dispatch for releases via GitHub UI
- Changelog management system using session history

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

[Unreleased]: https://github.com/ruifrvaz/smaqit/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/ruifrvaz/smaqit/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ruifrvaz/smaqit/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ruifrvaz/smaqit/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/ruifrvaz/smaqit/releases/tag/v0.0.1
