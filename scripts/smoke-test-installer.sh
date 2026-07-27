#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
binary_input="${1:-$repo_root/installer/dist/smaqit-dev}"

if [[ "$binary_input" = /* ]]; then
  binary="$binary_input"
else
  binary="$(cd "$(dirname "$binary_input")" && pwd)/$(basename "$binary_input")"
fi

if [[ ! -x "$binary" ]]; then
  echo "[ERROR] Installer binary is missing or not executable: $binary" >&2
  exit 1
fi

smoke_root="$(mktemp -d "${TMPDIR:-/tmp}/smaqit-installer-smoke.XXXXXX")"

cleanup() {
  if [[ "${KEEP_SMOKE_DIR:-0}" == "1" ]]; then
    echo "[INFO] Preserving temporary project: $smoke_root"
    return
  fi
  case "$smoke_root" in
    */smaqit-installer-smoke.*)
      rm -rf -- "$smoke_root"
      ;;
    *)
      echo "[WARN] Refusing to remove unexpected temporary path: $smoke_root" >&2
      ;;
  esac
}
trap cleanup EXIT

assert_file_equal() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  if ! cmp -s "$expected" "$actual"; then
    echo "[ERROR] $label differs: $actual" >&2
    exit 1
  fi
}

assert_owned_tree_matches() {
  local expected_root="$1"
  local actual_root="$2"
  local label="$3"
  while IFS= read -r expected; do
    local rel="${expected#"$expected_root"/}"
    assert_file_equal "$expected" "$actual_root/$rel" "$label"
  done < <(find "$expected_root" -type f | sort)
  echo "[OK] $label matches generated staging"
}

echo "[INFO] Installer: $binary"
echo "[INFO] Temporary project: $smoke_root"

# Seed unrelated Codex content before first init. These shared namespaces must survive
# install, reinstallation, and uninstall byte-for-byte.
mkdir -p "$smoke_root/.codex/agents" "$smoke_root/.agents/skills/custom-skill" "$smoke_root/.agents/skills/smaqit.input-business"
printf '%s\n' 'custom_config = true' > "$smoke_root/.codex/config.toml"
printf '%s\n' 'name = "custom-agent"' > "$smoke_root/.codex/agents/custom-agent.toml"
printf '%s\n' '---' 'name: custom-skill' 'description: Unrelated sentinel skill.' '---' > "$smoke_root/.agents/skills/custom-skill/SKILL.md"
printf '%s\n' 'custom neighbor inside a smaqit-named skill directory' > "$smoke_root/.agents/skills/smaqit.input-business/custom-note.txt"

# An exact owned destination must prompt even before .smaqit/ exists. Cancellation
# preserves it; confirmation replaces it with the generated artifact.
first_owned_agent="$(find "$repo_root/installer/agents-codex" -maxdepth 1 -type f | sort | head -1)"
first_owned_name="$(basename "$first_owned_agent")"
printf '%s\n' 'first-install-conflict-sentinel' > "$smoke_root/.codex/agents/$first_owned_name"
first_cancel_output="$(printf 'n\n' | "$binary" init "$smoke_root")"
grep -Fq 'Existing files conflict with the smaqit installation.' <<< "$first_cancel_output"
grep -Fq 'Installation cancelled' <<< "$first_cancel_output"
grep -Fq 'first-install-conflict-sentinel' "$smoke_root/.codex/agents/$first_owned_name"
if [[ -e "$smoke_root/.smaqit" ]]; then
  echo "[ERROR] Cancelled first installation created .smaqit/" >&2
  exit 1
fi
printf 'y\n' | "$binary" init "$smoke_root"
assert_file_equal "$first_owned_agent" "$smoke_root/.codex/agents/$first_owned_name" "confirmed first-install conflict"

assert_owned_tree_matches "$repo_root/installer/agents-copilot" "$smoke_root/.github/agents" "Copilot agents"
assert_owned_tree_matches "$repo_root/installer/skills-copilot" "$smoke_root/.github/skills" "Copilot skills"
assert_owned_tree_matches "$repo_root/installer/agents-claude" "$smoke_root/.claude/agents" "Claude agents"
assert_owned_tree_matches "$repo_root/installer/commands-claude" "$smoke_root/.claude/commands" "Claude commands"
assert_owned_tree_matches "$repo_root/installer/skills-claude" "$smoke_root/.claude/skills" "Claude skills"
assert_owned_tree_matches "$repo_root/installer/agents-codex" "$smoke_root/.codex/agents" "Codex agents"
assert_owned_tree_matches "$repo_root/installer/skills-codex" "$smoke_root/.agents/skills" "Codex skills"

python3 - "$smoke_root" "$repo_root" <<'PY'
import pathlib
import sys
import tomllib

import yaml

root = pathlib.Path(sys.argv[1])
repo = pathlib.Path(sys.argv[2])

expected_agents = sorted((repo / "installer" / "agents-codex").glob("*.toml"))
if len(expected_agents) != 9:
    raise SystemExit(f"expected 9 generated Codex agents, found {len(expected_agents)}")

required_agent_fields = {"name", "description", "developer_instructions"}
for expected in expected_agents:
    installed = root / ".codex" / "agents" / expected.name
    data = tomllib.loads(installed.read_text())
    missing = {key for key in required_agent_fields if not data.get(key)}
    if missing:
        raise SystemExit(f"{installed}: missing required non-empty fields: {sorted(missing)}")

expected_skills = {
    path.name for path in (repo / "installer" / "skills-codex").iterdir() if path.is_dir()
}
if len(expected_skills) != 25:
    raise SystemExit(f"expected 25 generated Codex skills, found {len(expected_skills)}")

for skill_name in sorted(expected_skills):
    skill_file = root / ".agents" / "skills" / skill_name / "SKILL.md"
    text = skill_file.read_text()
    if not text.startswith("---\n"):
        raise SystemExit(f"{skill_file}: missing YAML frontmatter")
    _, frontmatter, _ = text.split("---", 2)
    metadata = yaml.safe_load(frontmatter)
    if not isinstance(metadata, dict) or not metadata.get("name") or not metadata.get("description"):
        raise SystemExit(f"{skill_file}: missing name or description")

print("[OK] 9 Codex agents and 25 Codex skills parsed successfully")
PY

if grep -R -E '\{\{(WEB_TOOL|DELEGATE_SPEC_AGENT|DELEGATE_INFRASTRUCTURE|DELEGATE_COVERAGE)\}\}|runSubagent|Task tool|/smaqit\.' "$smoke_root/.codex/agents"; then
  echo "[ERROR] Platform-incompatible agent content found in Codex output" >&2
  exit 1
fi

if grep -R -F '[SMAQIT_SKILLS_DIR]' "$smoke_root/.agents/skills"; then
  echo "[ERROR] Unresolved Codex skill-directory placeholder found" >&2
  exit 1
fi

grep -Fq 'custom_config = true' "$smoke_root/.codex/config.toml"
grep -Fq 'name = "custom-agent"' "$smoke_root/.codex/agents/custom-agent.toml"
grep -Fq 'name: custom-skill' "$smoke_root/.agents/skills/custom-skill/SKILL.md"
grep -Fq 'custom neighbor' "$smoke_root/.agents/skills/smaqit.input-business/custom-note.txt"

(
  cd "$smoke_root"
  "$binary" validate
)

# Reinstallation must detect exact owned conflicts. Cancellation preserves the modified
# file; confirmation restores the generated artifact.
owned_agent="$(find "$repo_root/installer/agents-codex" -maxdepth 1 -type f | sort | head -1)"
owned_agent_name="$(basename "$owned_agent")"
printf '%s\n' 'cancel-sentinel' > "$smoke_root/.codex/agents/$owned_agent_name"
printf 'n\n' | "$binary" init "$smoke_root"
grep -Fq 'cancel-sentinel' "$smoke_root/.codex/agents/$owned_agent_name"
printf 'y\n' | "$binary" init "$smoke_root"
assert_file_equal "$owned_agent" "$smoke_root/.codex/agents/$owned_agent_name" "confirmed reinstall"

(
  cd "$smoke_root"
  printf 'y\nn\n' | "$binary" uninstall
)

while IFS= read -r expected; do
  installed="$smoke_root/.codex/agents/$(basename "$expected")"
  if [[ -e "$installed" ]]; then
    echo "[ERROR] Owned Codex agent survived uninstall: $installed" >&2
    exit 1
  fi
done < <(find "$repo_root/installer/agents-codex" -maxdepth 1 -type f | sort)

while IFS= read -r expected; do
  relative="${expected#"$repo_root/installer/skills-codex/"}"
  installed="$smoke_root/.agents/skills/$relative"
  if [[ -e "$installed" ]]; then
    echo "[ERROR] Owned Codex skill file survived uninstall: $installed" >&2
    exit 1
  fi
done < <(find "$repo_root/installer/skills-codex" -type f | sort)

grep -Fq 'custom_config = true' "$smoke_root/.codex/config.toml"
grep -Fq 'name = "custom-agent"' "$smoke_root/.codex/agents/custom-agent.toml"
grep -Fq 'name: custom-skill' "$smoke_root/.agents/skills/custom-skill/SKILL.md"
grep -Fq 'custom neighbor' "$smoke_root/.agents/skills/smaqit.input-business/custom-note.txt"

echo "[PASS] Installer smoke test completed successfully"
