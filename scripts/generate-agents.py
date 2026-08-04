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
  -> installer/agents-codex/<name>.toml         (Codex project custom agent)

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
  -> installer/skills-codex/<name>/**     ([SMAQIT_SKILLS_DIR] -> .agents/skills)

Run via `make -C installer prepare`, or directly after editing agents/, commands/, skills/,
or .smaqit/definitions/agents/:
  python3 scripts/generate-agents.py
"""
import json
import copy
import re
import shutil
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent

PLATFORMS = ("copilot", "claude", "codex")

AGENTS_BODY_SRC_DIR = ROOT / "agents"
AGENTS_METADATA_SRC_DIR = ROOT / ".smaqit" / "definitions" / "agents"
AGENTS_OUT_DIR_BY_PLATFORM = {
    "copilot": ROOT / "installer" / "agents-copilot",
    "claude": ROOT / "installer" / "agents-claude",
    "codex": ROOT / "installer" / "agents-codex",
}
AGENT_OUT_SUFFIX_BY_PLATFORM = {
    "copilot": ".agent.md",
    "claude": ".md",
    "codex": ".toml",
}

COMMANDS_SRC_DIR = ROOT / "commands"
COMMANDS_OUT_DIR = ROOT / "installer" / "commands-claude"

SKILLS_SRC_DIR = ROOT / "skills"
SKILLS_OUT_DIR_BY_PLATFORM = {
    "copilot": ROOT / "installer" / "skills-copilot",
    "claude": ROOT / "installer" / "skills-claude",
    "codex": ROOT / "installer" / "skills-codex",
}
SKILLS_DIR_BY_PLATFORM = {
    "copilot": ".github/skills",
    "claude": ".claude/skills",
    "codex": ".agents/skills",
}

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


def render_codex_agent(metadata: dict, body: str, source_name: str) -> str:
    """Render a Codex project custom agent as a standalone TOML file."""
    required = ("name", "description")
    missing = [key for key in required if not metadata.get(key)]
    if missing:
        raise ValueError(f"{source_name}: Codex metadata missing: {', '.join(missing)}")
    if "'''" in body:
        raise ValueError(
            f"{source_name}: agent body contains triple single quotes and cannot be "
            "rendered as a TOML literal string"
        )

    content = (
        f"name = {json.dumps(metadata['name'], ensure_ascii=False)}\n"
        f"description = {json.dumps(metadata['description'], ensure_ascii=False)}\n"
        "developer_instructions = '''\n\n"
        f"{body.rstrip()}\n"
        "'''\n"
    )
    tools = metadata.get("tools", {})
    if tools:
        content += "\n[tools]\n"
        for key, value in tools.items():
            content += f"{key} = {json.dumps(value)}\n"
    for server_name, server in metadata.get("mcp_servers", {}).items():
        content += f'\n[mcp_servers.{json.dumps(server_name)}]\n'
        for key, value in server.items():
            content += f"{key} = {json.dumps(value)}\n"
    return content


def design_metadata(metadata: dict, platform: str, role: str | None) -> dict:
    """Add role-appropriate visual design capabilities to an agent."""
    metadata = copy.deepcopy(metadata)
    if role not in (None, "author"):
        raise ValueError(f"unsupported design role: {role}")
    if role is None:
        return metadata
    if platform == "copilot":
        tools = metadata.setdefault("tools", [])
        for tool in ("read/viewImage",):
            if tool not in tools:
                tools.append(tool)
        for tool in ("smaqit-plantuml/check_syntax", "smaqit-plantuml/render_diagram"):
            if tool not in tools:
                tools.append(tool)
        metadata["mcp-servers"] = {
            "smaqit-plantuml": {
                "type": "local",
                "command": "smaqit",
                "args": ["mcp", "plantuml"],
                "tools": ["check_syntax", "render_diagram"],
            }
        }
    elif platform == "claude":
        tools = metadata.setdefault("tools", [])
        if isinstance(tools, str):
            tools = [item.strip() for item in tools.split(",")]
            metadata["tools"] = tools
        for tool in (
            "mcp__smaqit-plantuml__check_syntax",
            "mcp__smaqit-plantuml__render_diagram",
        ):
            if tool not in tools:
                tools.append(tool)
        metadata["mcpServers"] = {
            "smaqit-plantuml": {
                "type": "stdio",
                "command": "smaqit",
                "args": ["mcp", "plantuml"],
            }
        }
    elif platform == "codex":
        metadata["tools"] = {"view_image": True}
        metadata["mcp_servers"] = {
            "smaqit-plantuml": {
                "command": "smaqit",
                "args": ["mcp", "plantuml"],
                "required": True,
            }
        }
    return metadata


def generate_agent(name: str) -> None:
    body_path = AGENTS_BODY_SRC_DIR / f"{name}.md"
    fm_path = AGENTS_METADATA_SRC_DIR / f"{name}.frontmatter.yaml"

    body = body_path.read_text()
    manifest = yaml.safe_load(fm_path.read_text())
    placeholders = manifest.get("placeholders", {})

    missing_platforms = [platform for platform in PLATFORMS if platform not in manifest]
    if missing_platforms:
        raise ValueError(f"{name}: metadata missing for: {', '.join(missing_platforms)}")

    for platform in PLATFORMS:
        out_dir = AGENTS_OUT_DIR_BY_PLATFORM[platform]
        out_suffix = AGENT_OUT_SUFFIX_BY_PLATFORM[platform]
        frontmatter = design_metadata(manifest[platform], platform, manifest.get("design-role"))
        missing_values = [key for key, values in placeholders.items() if platform not in values]
        if missing_values:
            raise ValueError(
                f"{name}: placeholders missing values for platform '{platform}': "
                f"{', '.join(missing_values)}"
            )
        values = {key: platform_values[platform] for key, platform_values in placeholders.items()}
        resolved_body = resolve_placeholders(body, values, name, platform)

        out_name = f"smaqit.{name}{out_suffix}"
        out_path = out_dir / out_name
        if platform == "codex":
            content = render_codex_agent(frontmatter, resolved_body, name)
        else:
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

    orphan_metadata = sorted(
        path.stem.removesuffix(".frontmatter")
        for path in AGENTS_METADATA_SRC_DIR.glob("*.frontmatter.yaml")
        if path.stem.removesuffix(".frontmatter") not in names
    )
    if orphan_metadata:
        print(
            f"orphan .smaqit/definitions/agents metadata for: {', '.join(orphan_metadata)}",
            file=sys.stderr,
        )
        sys.exit(1)

    for out_dir in AGENTS_OUT_DIR_BY_PLATFORM.values():
        if out_dir.exists():
            shutil.rmtree(out_dir)
        out_dir.mkdir(parents=True, exist_ok=True)

    for name in names:
        generate_agent(name)


def copy_commands() -> None:
    if not COMMANDS_SRC_DIR.exists():
        return
    if COMMANDS_OUT_DIR.exists():
        shutil.rmtree(COMMANDS_OUT_DIR)
    COMMANDS_OUT_DIR.mkdir(parents=True, exist_ok=True)
    for src in sorted(COMMANDS_SRC_DIR.glob("*.md")):
        out_path = COMMANDS_OUT_DIR / src.name
        out_path.write_text(src.read_text())
        print(f"wrote {out_path.relative_to(ROOT)}")


def generate_skills() -> None:
    if not SKILLS_SRC_DIR.exists():
        print(f"no source directory at {SKILLS_SRC_DIR}", file=sys.stderr)
        sys.exit(1)

    for platform, out_dir in SKILLS_OUT_DIR_BY_PLATFORM.items():
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
