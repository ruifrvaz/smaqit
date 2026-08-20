# smaqit's own HarnessBench suite

This directory is smaqit dogfooding the `smaqit-adk bench` engine against its own skills. The engine itself ships in the `smaqit-adk` binary (sibling repo, `~/projects/smaqit-adk`, `src/bench`/`src/benchcli`) — it is not smaqit-owned or installed by this project's own installer. This directory is this repo's local data, in the same sense `.smaqit/tasks/` or `.smaqit/compendium.md` are: state a project keeps about itself, not something the global installer writes into a consumer's project. See `smaqit-adk`'s own `.smaqit/bench/README.md` for the fuller convention this file is adapted from.

## Layout

```
.smaqit/bench/
├── skills/
│   └── <skill-id>/
│       └── bench.yaml       # + any fixture/oracle files it references
├── runs/                    # generated experiment output — gitignored
└── README.md
```

Skill suites live under `skills/<skill-id>/`, matching the skill directory name at the repo root (`skills/<skill-id>/SMAQIT.md`). One `bench.yaml` per target unless it has genuinely distinct scenarios worth splitting into multiple manifests.

## Bench vocabulary

Bench uses **Case** for an evaluation scenario, **Prompt** for the author-supplied `given.prompt`, and **Case brief** for the rendered prompt plus shared-input and variant-treatment paths delivered to a harness. A smaqit **Task** remains a tracked work item and is not Bench terminology.

## With-artifact / without-artifact comparison — and when to skip it

Bench's headline pattern is one Case, two Variants: the target skill staged as the with-artifact variant's `treatment`, an empty baseline treatment for without-artifact. This is the right shape when the question is "does having this skill change agent behavior." It is *not* automatically the right shape for a pure bug-fix regression case — there, an agent without the skill staged wouldn't be running the fixed script at all, so there's no meaningful baseline to compare against. A single with-artifact-only case is a deliberate, documented choice in that situation, not an oversight — say so explicitly in the manifest's own comments when you make it.

## Reusable Codex process-variant block

Copy this verbatim into every manifest that drives `codex exec` — adjust only the prompt/arguments specific to its case. See `smaqit-adk`'s own `.smaqit/bench/README.md` for the full rationale (pinned sandbox mode against `openai/codex#36570`, `--skip-git-repo-check` since Bench's workspace is a plain temp dir, an explicit `timeoutSeconds` against `openai/codex#28476`):

```yaml
process:
  executable: codex
  arguments:
    - exec
    - --sandbox
    - danger-full-access
    - --skip-git-repo-check
    - --cd
    - "{workspace}"
    - "{briefFile}"
  inputMode: argument
```

## Case data planes

- `fixture` is a writable starting project tree shared by every variant; `destination` places it at a safe workspace-relative path.
- Case-level `prepare` performs common deterministic preparation before the baseline snapshot. It can use only `{workspace}` and `{caseId}`, and runs sequentially, waiting for each command to exit.
- `given` inputs are shared, read-only resources available to every variant.
- Variant `treatment` artifacts are read-only resources available only to that variant, staged into a read-only `.smaqit-bench-input/` sidecar (**not** at the artifact's real project-relative path) — prompts must reference the Case brief's treatment table, not a guessed path.

**Known gap — no teardown for a backgrounded `prepare` process.** A process launched in the background from `prepare` (e.g. `vault server -dev &`) survives past a successful run with no engine-provided cleanup — confirmed against `smaqit-adk`'s own `src/bench/run.go`/`process.go`: process-group termination only fires on timeout/cancellation, never on success. Do not background a long-lived process in `prepare` until this is fixed upstream (`smaqit-adk` task 033). A Case needing a live backing service across its full lifecycle should be deferred or scoped down instead, as this repo's first manifest (`skills/smaqit.infrastructure-vault-loader/bench.yaml`) does.

## Command graders and environment

A `command`-type expectation, grader, or Case preparation command runs with **no environment at all** unless the manifest sets `command.environment.inherit`/`.set` — not even `PATH` or `HOME`. Plain POSIX tools (`sh`, `grep`, `test`, `cat`) resolve fine via `sh`'s own default `PATH` fallback when none is set; anything needing a specific `PATH`, credentials, or service address must declare `environment.inherit`/`.set` explicitly.

## Run output

`bench run` (and the suite-level equivalent) writes experiment output under each manifest's `output.directory`. Point every dogfood manifest at a path under `.smaqit/bench/runs/<skill-id>/` so all evidence lands in one gitignored place.
