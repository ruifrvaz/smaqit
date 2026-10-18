---
status: Completed
created: "2026-10-18"
mode: Assisted
started: "2026-10-18"
completed: "2026-10-18"
---

# Orchestrator Skills Default by Invocation; Rename new-greenfield-project to project-new; Prune Retired Skills

## Description

`smaqit.feature-new` and `smaqit.new-greenfield-project` always default to `autonomous`, even when the
user types their slash command. Each skill's Phase 0 says "Execution mode is **Autonomous** unless the
operator asks for assisted mode", and checks nothing about how it was started. The resolved mode then
spreads to every phase and specification agent, so a typed `/smaqit.feature-new` runs all five phases
without a review pause.

The phase commands already follow the intended rule: a typed `/smaqit.development` is `assisted`, and a
model-routed one is `autonomous`. Task 138 (v3.17.0) recorded the orchestrators as autonomous by
design. That design is now superseded. The rule is: explicit user invocation means `assisted`, and the
model choosing the skill itself means `autonomous`. The managed `AGENTS.md` section ships the old rule
to every downstream repository (spotted in a downstream review of the v3.17.2 upgrade).

The greenfield skill is also renamed `smaqit.project-new`, everywhere the text means that skill, with
no alias.

Renaming exposes a long-standing installer gap: global install only writes files, so a retired or
renamed skill's directory stays installed forever. On this machine `smaqit.new-greenfield-project` is
present in both `~/.claude/skills/` and `~/.agents/skills/`. After the rename both names would be listed
with near-identical descriptions. The installer gains an install manifest plus a seed list, so retired
skills are removed from machines.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Claude Code, GitHub Copilot, OpenAI Codex, Go, Bash, Python
**Platforms/Environments:** None
**Features/Integrations:** skill slash invocation, Codex $ mention and /skills picker, execution mode defaults, global skill install and uninstall, retired-skill pruning, AGENTS.md managed section
**Versions/Constraints:** no backwards compatibility (no alias), skill descriptions at most 400 characters, never remove smaqit-extensions or user files, historical records unchanged

## Design Decisions

Resolved by `task.plan` on 2026-10-18.

- **Mode trigger: explicit invocation syntax only.** These count as user-started (`assisted`):
  - a typed `/smaqit.project-new` or `/smaqit.feature-new`;
  - a Codex `$` mention;
  - a pick from the `/skills` menu.

  Claude choosing the skill itself to handle a broader request counts as model-started (`autonomous`).
  So does a user naming the skill in plain words. A mode the user states always wins, and sub-phases
  inherit the resolved mode.
- **Phase agents keep "picked or named" as user-started.** In Codex, asking it to spawn an agent by name
  is the only way to start a phase agent. The docs note the asymmetry.
- **Full rename, every mention that means this skill or its phases.** For example, "greenfield Phase 4"
  becomes "project-new Phase 4". This includes the H1, the stale `/zero.to.prod` example, and the
  `design_test.go` test name. Released CHANGELOG entries, history files and task records stay unchanged
  as historical facts.
- **No alias for the old name** (no backwards compatibility). The skill version becomes 3.0.0.
- **Distinguish the description from smaqit-extensions' `smaqit.project-init`.** The `smaqit.project-new`
  name is free; `project-init` is a different skill. The new description must say this skill goes from
  raw assets to production, whereas `project-init` only sets up instruction files and base directories.
  It must stay within 400 characters.
- **Old installs are removed by an install manifest plus a seed list.**
  - Global install records every file it installs, per destination, honouring `CLAUDE_CONFIG_DIR`,
    `COPILOT_HOME` and `CODEX_HOME`.
  - The next install removes manifest entries the current release no longer ships: files first, then
    directories left empty.
  - The seed covers names retired before the manifest existed: `smaqit.new-greenfield-project`,
    `smaqit.project-zero-to-prod`, `smaqit.feature-deploy`, `smaqit.infrastructure-onboard-k3s-app`, plus
    any others the CHANGELOG's renamed or removed entries show. A seeded directory is removed only when
    its `SKILL.md` `name:` matches.
  - smaqit-extensions files and user files are never removed. Uninstall uses the manifest.
- **No new live e2e scenario.** The orchestrators need smaqit-extensions' task lifecycle and a lot of
  model usage, so the rule stays at text level, like the phase commands.
- **Release:** MINOR, with an action-required note.

## Implementation Steps

*Phase A — mode by invocation*
1. `skills/smaqit.feature-new/SKILL.md` (Phase 0, around lines 31–35) and the greenfield skill's Phase 0
   (around lines 56–60): replace the hardcoded autonomous default with the invocation rule. Reuse the
   phase commands' wording ("the user typed `/…`, or you invoked it yourself"), extended with Codex `$`
   and `/skills`. Keep the assisted-only gates. The synthesized-skill checkpoint follows the resolved
   mode.
2. `skills/smaqit.input-{development,deployment,validation}/SKILL.md`: the bullet that lists the
   orchestrators under autonomous becomes "an orchestrating skill passes the mode it resolved".
3. `templates/AGENTS.md.template` (the line-21 bullet and the line-13 routing row), `framework/AGENTS.md`
   (lines 284–285), `README.md:53`, `docs/wiki/agent-tools-reference.md` (the mode table, lines 186–197),
   `docs/wiki/designs/fail-fast-on-ambiguity.md:93` and `.smaqit/compendium.md` (lines 171, 183 and
   others): state one rule.

*Phase B — rename (parallel with Phase A)*

4. `git mv skills/smaqit.new-greenfield-project skills/smaqit.project-new`. Then update:
   - the `name:` field, the H1 and the self-path at `SKILL.md:177`;
   - `references/diagrams.md:3` and the stale `/zero.to.prod` example;
   - the version, to 3.0.0;
   - the description, to distinguish it from `project-init`.
5. Rename every other live mention:
   - `feature-new`, `infrastructure-cicd-generate` (its description is at the 400-character limit),
     `infrastructure-deploy-rsync`, `infrastructure-deploy-rsync-python-tornado`,
     `infrastructure-provision-cyso` and its `scripts/ownership-guard.sh` (stderr lines 76 and 108),
     `input-deployment`, `input-development`, `input-validation`;
   - "greenfield Phase N" in the `deploy-k3s-app`, `deploy-rsync-python-tornado` and
     `request-k3s-onboarding` descriptions;
   - `agents/development.md:162`;
   - `.smaqit/definitions/skills/*`;
   - `docs/wiki/workflows/deployment-slots.md`;
   - `installer/design_test.go` (line 367's path and line 353's test name);
   - the comments at `installer/slots.go:208` and `slots_test.go:109`.

*Phase C — pruning retired skills (installer)*

6. `installer/main.go`: `cmdInstallGlobal` writes a per-destination manifest of installed files. On the
   next install, it removes entries the current embed no longer ships (files, then empty directories),
   the way `removeEmbeddedSkillDirs` works. `cmdUninstall` uses the manifest.
7. A seed list of retired names, applied with a check that the `SKILL.md` `name:` matches.

*Phase D — tests and release*

8. Go tests with a temporary `HOME`:
   - the manifest is written;
   - an entry a later release drops is removed;
   - smaqit-extensions and user files survive;
   - the seed removes only matching names;
   - uninstall follows the manifest.
9. A guard test: no shipped file contains `new-greenfield-project`.
10. `scripts/smoke-test-installer.sh`: pre-create the old skill directory in the temporary home, install,
    and assert that only `smaqit.project-new` remains.
11. CHANGELOG with an action-required note, `make test`, `make smoke-test`, and release.

## Known Issues Triage
**Triaged:** 2026-10-18
**Tools searched:** Claude Code (`anthropics/claude-code`), OpenAI Codex (`openai/codex`, from the research map's project table), GitHub Copilot (`microsoft/vscode-copilot-release`)
**Result:** Advisory

### Advisory Issues
- [#95469 Stacked skills: only the first is treated as user-invoked, so a disable-model-invocation skill in position 2+ is refused](https://github.com/anthropics/claude-code/issues/95469) — `anthropics/claude-code` — opened 2026-09-18 — bug, has repro, platform:windows, area:skills. Claude Code already distinguishes user-invoked from model-invoked skills. Only the first skill in a stacked invocation counts as user-invoked, which is an edge case for the mode rule. This task sets no `disable-model-invocation`.
- [#87113 Stacked slash-skill invocations load only the first skill on the SDK / --print path](https://github.com/anthropics/claude-code/issues/87113) — `anthropics/claude-code` — opened 2026-08-16 — bug, has repro, area:cli, area:skills. Relevant only if headless e2e stacks two skills in one prompt.
- [#78523 disable-model-invocation: true breaks explicit /skill-name invocation](https://github.com/anthropics/claude-code/issues/78523) — `anthropics/claude-code` — opened 2026-07-17 — bug, area:skills. A reason not to restrict the orchestrators to user-only invocation: they must stay model-invocable for the autonomous path anyway.
- [#44172 Skills across ~/.codex/skills and ~/.agents/skills appear duplicated due to path-based deduplication](https://github.com/openai/codex/issues/44172) — `openai/codex` — opened 2026-09-09 — bug, windows-os, app, skills. Duplicate skill listings are what the pruning half of this task removes for retired names. smaqit installs Codex skills only under `~/.agents/skills`.
- [#41576 $ opens Skills/Apps popup and cannot be disabled](https://github.com/openai/codex/issues/41576) — `openai/codex` — opened 2026-08-29 — bug, app, skills. Confirms `$` as Codex's explicit skill-invocation syntax.

### Historical (Closed)
- [#19695 Explicit-only skills need a visible, reliable explicit invocation contract](https://github.com/openai/codex/issues/19695) — `openai/codex` — closed 2026-04-26
- [#43727 TUI skill picker uses display_name casing, causing explicit-only skill invocation to fail](https://github.com/openai/codex/issues/43727) — `openai/codex` — closed 2026-10-12
- [#73774 Skill tool re-invoked after slash command already loaded skill](https://github.com/anthropics/claude-code/issues/73774) — `anthropics/claude-code` — closed 2026-08-17
- [#14147 Separate invocation namespace for prompt files vs skills in the slash picker](https://github.com/microsoft/vscode-copilot-release/issues/14147) — `microsoft/vscode-copilot-release` — closed 2026-03-03 (confirms that Copilot lists skills in its `/` picker)

### Unresolvable Tools
- OpenAI Codex — the helper's registry lookup matched an unrelated repository (`steipete/CodexBar`); used `openai/codex` from the research map instead

### Omitted Tools
- Go, Bash, Python — build, test and installer tooling only; no task feature dimension to search

## Acceptance Criteria

- [x] `/smaqit.project-new` and `/smaqit.feature-new` (and Codex `$` or `/skills` picks) default to
      `assisted`; when Claude chooses them they default to `autonomous`; a stated mode wins; sub-phases
      inherit the resolved mode
- [x] The AGENTS.md template, framework, input skills, README, wiki and compendium state one invocation
      rule; none lists the orchestrators as autonomous by default
- [x] `skills/smaqit.project-new` replaces `smaqit.new-greenfield-project`; no shipped file references the
      old name (guarded by a test); historical records unchanged
- [x] The `smaqit.project-new` description separates it from smaqit-extensions' `smaqit.project-init` and
      stays within 400 characters
- [x] Global install writes a manifest and removes entries the current release no longer ships; a seed
      list removes previously retired names only when the `SKILL.md` name matches; smaqit-extensions and
      user files are never removed; uninstall uses the manifest (Go tests with a temporary `HOME`)
- [x] The smoke test proves that installing over a home holding the old skill leaves only
      `smaqit.project-new`
- [x] `make test` and `make smoke-test` pass; CHANGELOG has an action-required note; released

## Findings

**Implementation approach:**
- **Mode:**
  - Both orchestrators' Phase 0 now resolve their mode from how they were invoked. A slash command, a Codex `$` mention or a skill-menu pick means assisted; the model choosing the skill means autonomous. Each run announces its mode in one line.
  - One rule now appears everywhere it is stated: the three input skills, the three phase agents, the `AGENTS.md` template, `framework/AGENTS.md`, the README, the wiki and the compendium.
- **Rename:**
  - `git mv` moved the skill to `smaqit.project-new` (3.0.0). Its new 392-character description sets it apart from `smaqit.project-init`.
  - Every other mention was rewritten one phrase at a time, never with a blanket word replacement: the exact name, then each "greenfield" phrase that refers to this skill. Released history was left as it was.
- **Pruning:**
  - The new `installer/global_manifest.go` writes a manifest called `.smaqit-installed.json` in each global destination. The next install removes the files the previous release installed and this one no longer ships.
  - A seed list covers skills retired before the manifest existed. It only removes a skill whose SKILL.md `name:` still matches.
  - Uninstall also removes the manifest's leftovers.
  - `copyEmbeddedDir` and the new helpers now take `fs.FS`, so tests can use `fstest.MapFS`.
- **Verification:**
  - New Go tests cover pruning, the name-matched seed, untrusted manifests, uninstall, and that the seed never names a skill still shipped. A guard test checks that no shipped file names `new-greenfield-project`.
  - The smoke test upgrades the real binary over a home that already holds the old skill.
  - Three deliberate breaks were each caught by a test.

**Decisions made:**
- **The manifest lives inside each destination,** not in a central `~/.config` file. It then follows `CLAUDE_CONFIG_DIR`, `COPILOT_HOME` and `CODEX_HOME`, stays isolated by the existing tests, and can't leak into the real config through `XDG_CONFIG_HOME`.
- **The seed list comes from git history and release tags,** not the CHANGELOG. It holds only `smaqit.`-prefixed skills that a v3 release installed globally: `smaqit.new-greenfield-project` and `smaqit.infrastructure-onboard-k3s-app`. `assessment`, `feature-deploy` and `project-zero-to-prod` were never installed globally, and no agent has been retired since v3.0.0.
- **An untrusted manifest stops the install before anything is removed.** That covers bad JSON, an unknown schema, and paths that leave the destination.
- **The seed removes a retired skill's whole directory, but only after its SKILL.md name matches,** because pre-manifest releases recorded no file list.
- **Mode detection stays inference-based.** Research confirmed that Claude Code offers deterministic signals: `UserPromptExpansion` fires for a typed command, and `PreToolUse` on the `Skill` tool fires for a model invocation. Both would need hooks in `~/.claude/settings.json` or a plugin, though. The user chose not to pursue that.
- **Phase agents keep "picked or named" as a user start,** because Codex starts agents by name. The asymmetry is documented in `framework/AGENTS.md`.

**Blockers encountered:**
- None.

- **Final name `smaqit.project-new`, not `smaqit.new-project`.** The user changed it in review so it matches `smaqit.feature-new` (noun then verb; diagnostics read "Project New Phase N"). It was renamed again on the same PR (#109). The guard test also rejects the interim name, which was never released, so it needs no seed entry.

**Follow-up identified:**
- A one-line wording fix is needed in smaqit-extensions: `smaqit.project-init`'s description ("start… a smaqit project") overlaps with `smaqit.project-new`.
- Downstream repositories pick up the corrected `AGENTS.md` section on their next `smaqit update`.
- Claude Code hooks for deterministic invocation detection were researched and explicitly declined by the user; there is no task for them.

## Files to Create / Modify

| File | Action |
|------|--------|
| skills/smaqit.new-greenfield-project/** → skills/smaqit.project-new/** | Rename + Modify |
| skills/smaqit.feature-new/SKILL.md | Modify |
| skills/smaqit.input-{development,deployment,validation}/SKILL.md | Modify |
| skills/smaqit.infrastructure-{cicd-generate,deploy-rsync,deploy-rsync-python-tornado,provision-cyso,deploy-k3s-app,request-k3s-onboarding}/** | Modify |
| agents/development.md, framework/AGENTS.md, templates/AGENTS.md.template | Modify |
| .smaqit/definitions/skills/* | Modify / Rename |
| installer/main.go, installer/update.go | Modify |
| installer/main_test.go, installer/design_test.go, new manifest tests | Modify / Create |
| scripts/smoke-test-installer.sh | Modify |
| README.md, docs/wiki/agent-tools-reference.md, docs/wiki/designs/fail-fast-on-ambiguity.md, docs/wiki/workflows/deployment-slots.md, .smaqit/compendium.md, CHANGELOG.md | Modify |

## Notes

- Supersedes task 138's design decision that orchestrator skills default to autonomous. Raised while
  reviewing a downstream repository's v3.17.2 upgrade PR.
- Follow-up outside this repo: smaqit-extensions' `smaqit.project-init` description ("start… a smaqit
  project") overlaps with `project-new`. A one-line wording fix belongs there.
- Downstream repositories pick up the corrected `AGENTS.md` section on their next `smaqit update`.
- Closes the installer retired-skill pruning gap noted in earlier sessions.
