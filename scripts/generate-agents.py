#!/usr/bin/env python3
"""Generate platform-specific agent, command, and skill files for the installer.

Every artifact type follows the same rule: source lives at the repo root (or in
.smaqit/definitions/ for metadata that only makes sense per-platform), and compiled,
platform-resolved output exists ONLY inside installer/ (gitignored, rebuilt by
`make -C installer prepare`). Nothing here is ever written back to the repo root.

Agents — split across two committed locations:
  agents/<name>.md                                  - canonical body (no frontmatter), with
                                                       {{PLACEHOLDER}} tokens for the small
                                                       number of phrases that differ per platform
  .smaqit/definitions/agents/<name>.frontmatter.yaml - per-platform frontmatter metadata
                                                       (name/description/tools/...) and the
                                                       resolved value of each {{PLACEHOLDER}}
  -> installer/agents-copilot/<name>.agent.md   (GitHub Copilot custom agent)
  -> installer/agents-claude/<name>.md          (Claude Code subagent)

Commands — Claude Code-only (Copilot invokes an agent by its own `name:` directly, so it
needs no separate command file). No metadata split needed; source is already a complete
Claude Code command file (frontmatter + body):
  commands/<name>.md
  -> installer/commands-claude/<name>.md   (copied verbatim)

Skills — one shared source tree, since SKILL.md frontmatter needs no per-platform variance.
The only platform-specific bit is each skill's own install path, referenced in a few
usage comments via the [SMAQIT_SKILLS_DIR] placeholder, resolved here (not at install time):
  skills/<name>/**
  -> installer/skills-copilot/<name>/**   ([SMAQIT_SKILLS_DIR] -> .github/skills)
  -> installer/skills-claude/<name>/**    ([SMAQIT_SKILLS_DIR] -> .claude/skills)

Run via `make -C installer prepare`, or directly after editing agents/, commands/, skills/,
or .smaqit/definitions/agents/:
  python3 scripts/generate-agents.py
"""
import re
import shutil
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent

AGENTS_BODY_SRC_DIR = ROOT / "agents"
AGENTS_METADATA_SRC_DIR = ROOT / ".smaqit" / "definitions" / "agents"
AGENTS_COPILOT_OUT_DIR = ROOT / "installer" / "agents-copilot"
AGENTS_CLAUDE_OUT_DIR = ROOT / "installer" / "agents-claude"

COMMANDS_SRC_DIR = ROOT / "commands"
COMMANDS_OUT_DIR = ROOT / "installer" / "commands-claude"

SKILLS_SRC_DIR = ROOT / "skills"
SKILLS_COPILOT_OUT_DIR = ROOT / "installer" / "skills-copilot"
SKILLS_CLAUDE_OUT_DIR = ROOT / "installer" / "skills-claude"
SKILLS_DIR_BY_PLATFORM = {"copilot": ".github/skills", "claude": ".claude/skills"}

PLACEHOLDER_RE = re.compile(r"\{\{([A-Z0-9_]+)\}\}")


class FlowList(list):
    """A list that always renders in YAML flow style, e.g. [a, b, c]."""


def _represent_flow_list(dumper: yaml.Dumper, data: "FlowList"):
    return dumper.represent_sequence("tag:yaml.org,2002:seq", data, flow_style=True)


yaml.SafeDumper.add_representer(FlowList, _represent_flow_list)


def dump_frontmatter(data: dict) -> str:
    data = dict(data)
    if isinstance(data.get("tools"), list):
        data["tools"] = FlowList(data["tools"])
    return yaml.safe_dump(data, sort_keys=False, default_flow_style=False, width=1000).strip()


def resolve_placeholders(body: str, values: dict, agent_name: str, platform: str) -> str:
    def replace(match: re.Match) -> str:
        key = match.group(1)
        if key not in values:
            raise ValueError(
                f"{agent_name}: placeholder {{{{{key}}}}} used in body but not defined "
                f"for platform '{platform}' in .smaqit/definitions/agents/{agent_name}.frontmatter.yaml"
            )
        return values[key]

    return PLACEHOLDER_RE.sub(replace, body)


def generate_agent(name: str) -> None:
    body_path = AGENTS_BODY_SRC_DIR / f"{name}.md"
    fm_path = AGENTS_METADATA_SRC_DIR / f"{name}.frontmatter.yaml"

    body = body_path.read_text()
    manifest = yaml.safe_load(fm_path.read_text())
    placeholders = manifest.get("placeholders", {})

    for platform, out_dir, out_suffix in (
        ("copilot", AGENTS_COPILOT_OUT_DIR, ".agent.md"),
        ("claude", AGENTS_CLAUDE_OUT_DIR, ".md"),
    ):
        if platform not in manifest:
            continue
        frontmatter = manifest[platform]
        values = {key: platform_values[platform] for key, platform_values in placeholders.items()}
        resolved_body = resolve_placeholders(body, values, name, platform)

        out_name = f"smaqit.{name}{out_suffix}"
        out_path = out_dir / out_name
        content = (
            "---\n"
            f"{dump_frontmatter(frontmatter)}\n"
            "---\n\n"
            f"{resolved_body}"
        )
        out_path.write_text(content)
        print(f"wrote {out_path.relative_to(ROOT)}")


def generate_agents() -> None:
    if not AGENTS_BODY_SRC_DIR.exists():
        print(f"no source directory at {AGENTS_BODY_SRC_DIR}", file=sys.stderr)
        sys.exit(1)

    names = sorted(p.stem for p in AGENTS_BODY_SRC_DIR.glob("*.md"))
    if not names:
        print(f"no *.md sources found in {AGENTS_BODY_SRC_DIR}", file=sys.stderr)
        sys.exit(1)

    missing_metadata = [
        n for n in names if not (AGENTS_METADATA_SRC_DIR / f"{n}.frontmatter.yaml").exists()
    ]
    if missing_metadata:
        print(
            f"missing .smaqit/definitions/agents/*.frontmatter.yaml for: {', '.join(missing_metadata)}",
            file=sys.stderr,
        )
        sys.exit(1)

    AGENTS_COPILOT_OUT_DIR.mkdir(parents=True, exist_ok=True)
    AGENTS_CLAUDE_OUT_DIR.mkdir(parents=True, exist_ok=True)

    for name in names:
        generate_agent(name)


def copy_commands() -> None:
    if not COMMANDS_SRC_DIR.exists():
        return
    COMMANDS_OUT_DIR.mkdir(parents=True, exist_ok=True)
    for src in sorted(COMMANDS_SRC_DIR.glob("*.md")):
        out_path = COMMANDS_OUT_DIR / src.name
        out_path.write_text(src.read_text())
        print(f"wrote {out_path.relative_to(ROOT)}")


def generate_skills() -> None:
    if not SKILLS_SRC_DIR.exists():
        print(f"no source directory at {SKILLS_SRC_DIR}", file=sys.stderr)
        sys.exit(1)

    for platform, out_dir in (
        ("copilot", SKILLS_COPILOT_OUT_DIR),
        ("claude", SKILLS_CLAUDE_OUT_DIR),
    ):
        if out_dir.exists():
            shutil.rmtree(out_dir)
        shutil.copytree(SKILLS_SRC_DIR, out_dir)

        skills_dir = SKILLS_DIR_BY_PLATFORM[platform]
        for path in out_dir.rglob("*"):
            if not path.is_file():
                continue
            try:
                text = path.read_text()
            except UnicodeDecodeError:
                continue  # binary asset, nothing to substitute
            if "[SMAQIT_SKILLS_DIR]" in text:
                path.write_text(text.replace("[SMAQIT_SKILLS_DIR]", skills_dir))

        print(f"wrote {out_dir.relative_to(ROOT)}/ ({sum(1 for _ in out_dir.rglob('*') if _.is_file())} files)")


def main() -> None:
    generate_agents()
    copy_commands()
    generate_skills()


if __name__ == "__main__":
    main()
