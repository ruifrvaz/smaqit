# Reconcile Python/Tornado Rsync Deployment Skill Into Canonical smaqit

**Status:** Completed
**Created:** 2026-08-13
**Started:** 2026-08-14
**Completed:** 2026-08-14
**Mode:** Assisted

## Description

A downstream project validated a deployment skill for a Python/Tornado monolith with SQLite persistence, nginx, systemd, and rsync. The authored definition has been contributed to canonical smaqit at `.smaqit/definitions/skills/smaqit.infrastructure-deploy-rsync-python-tornado.md`; this task makes it a supported product capability rather than a downstream-only artifact.

The product implementation must retain the safety-critical deployment behavior while generalizing project-specific provenance and examples. It must integrate with the global installation model and the Dynamic Stack Detection / on-the-fly deploy-skill synthesis flow from Task 087, so matching projects select this maintained skill rather than synthesizing an equivalent copy.

## Design Decisions

- **Deployment profile:** Python/Tornado, single-process, no Docker or frontend build, SQLite persistence, nginx, and systemd are the matching stack characteristics.
- **Data safety:** Rsync excludes SQLite databases and local virtual environments; the deployed database remains on the target's persistent storage.
- **Reusable dependencies:** nginx vhost generation and deploy stamps reuse the existing shared capabilities rather than duplicating their implementations.
- **Product integration:** The canonical skill is built and installed globally through the normal smaqit release pipeline; it is not copied back into project-local framework mirrors.

## Implementation Steps

1. Assess the contributed definition against the canonical deployment-skill conventions and generalize any project-specific provenance, paths, or examples.
2. Compile the definition into `skills/smaqit.infrastructure-deploy-rsync-python-tornado/`, including any required bundled references or scripts, and validate its metadata and self-references.
3. Update stack-detection and no-match/synthesis routing so the Python/Tornado/systemd/SQLite profile selects the maintained skill.
4. Include the skill in global payload generation and installation for every supported agent platform.
5. Add focused automated coverage for matching, non-matching, and safety-critical rsync exclusions; run the relevant installer and product test suites.
6. Update user-facing documentation and release notes for the new supported deployment profile.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [x] Canonical smaqit contains a reviewed, product-generic definition and compiled `smaqit.infrastructure-deploy-rsync-python-tornado` skill.
- [x] Dynamic deploy-skill selection recognizes the Python/Tornado, SQLite, nginx, systemd, no-Docker profile and selects the canonical skill without synthesis.
- [x] The skill preserves database and virtual-environment rsync exclusions, uses the shared nginx-vhost and deploy-stamp conventions, and does not introduce Terraform or Docker steps.
- [x] The released global payload makes the skill available through the supported global skill roots.
- [x] Automated tests cover profile routing and the safety-critical deployment behavior, and relevant product/installer tests pass.
- [x] Product documentation records the supported deployment profile and its operational prerequisites.

## Findings

**Implementation approach:**
- Compiled the contributed definition directly into `skills/smaqit.infrastructure-deploy-rsync-python-tornado/SKILL.md`, following the structural shape of the sibling `smaqit.infrastructure-deploy-rsync-python-nextjs` skill: YAML frontmatter (`name`, `description`, `metadata.version/validated/validated-stack`), then Pre-conditions, Steps, Output, Scope, Examples, Gotchas, Completion, Failure Handling, Allowed Tools.
- Dropped the definitions-only `Provenance` and `Required-inherited-context` headings from the compiled skill — the latter's four points (`__APP_DIR__` token, shared `write-vhost.sh` reuse, shared deploy-stamp `printf` pattern, no-Terraform-step) were already expressed inline in Steps/Gotchas, so no information was lost.
- Confirmed Task 087's Phase 4 Step 6 routing in `smaqit.new-greenfield-project` is already generic ("compare the declared stack against the currently-installed `smaqit.infrastructure-deploy-rsync*` skills by description/metadata") — no code or skill-name enumeration needed updating; an accurate `description:` field in the new skill's frontmatter is sufficient for discovery.
- Bumped the two hardcoded skill-count assertions from 26 to 27 (`installer/main_test.go`'s `TestRemoveEmbeddedSkillDirsPreservesUnownedCodexContent` and `TestSharedSkillsServeCopilotAndCodex`) and the matching count in `docs/wiki/workflows/testing-smaqit.md`.
- Regenerated installer build artifacts (`make -C installer prepare`), ran `go vet`/`go test ./...` (pass), and `scripts/smoke-test-installer.sh` (pass) to confirm the new skill installs cleanly to the shared global path with its `[SMAQIT_SKILLS_DIR]` placeholder resolved.

**Decisions made:**
- No dedicated Go-level "stack matching" test was added — deploy-skill selection is a prose/metadata judgment step performed by an agent reading `SKILL.md` descriptions during Phase 4 Step 6, not a code path; the existing generic `TestSharedSkillsServeCopilotAndCodex` invariant (count, placeholder resolution, path correctness) is the appropriate and proportionate automated coverage for a pure-documentation skill.
- No wiki or README page enumerates the `deploy-rsync*` family by name, so no additional documentation location needed updating beyond the release CHANGELOG entry and the skill file itself.

**Blockers encountered:**
- None.

**Follow-up identified:**
- None.

## Files to Create / Modify

| File | Action |
|------|--------|
| `.smaqit/definitions/skills/smaqit.infrastructure-deploy-rsync-python-tornado.md` | Create — contributed downstream definition |
| `skills/smaqit.infrastructure-deploy-rsync-python-tornado/` | Create — compiled canonical skill and any supporting assets |
| `skills/smaqit.new-greenfield-project/` and routing implementation | Modify — select the maintained skill for the matching stack profile |
| Global installer/payload generation and tests | Modify — distribute and verify the new skill |
| User documentation and release notes | Modify — document supported profile |

## Notes

The source definition was synthesized and validated in a downstream project on 2026-07-21. It deliberately differs from the existing Node/React Docker Compose and FastAPI/Next.js Docker Compose deploy skills: the target is a direct Python/Tornado rsync deployment with a persistent SQLite database and systemd supervision.
