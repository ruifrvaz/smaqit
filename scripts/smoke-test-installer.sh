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

git init -q "$smoke_root"

node_failure_target="$smoke_root/node-prerequisite-failure"
if PATH=/smaqit-node-intentionally-unavailable "$binary" init "$node_failure_target" > "$smoke_root/node-failure.txt" 2>&1; then
  echo "[ERROR] Initialization succeeded without the mandatory Node prerequisite" >&2
  exit 1
fi
grep -Fq 'DESIGN-TOOLCHAIN-UNAVAILABLE' "$smoke_root/node-failure.txt"
if [[ -e "$node_failure_target" ]]; then
  echo "[ERROR] Failed Node preflight partially created the target project" >&2
  exit 1
fi
rm -- "$smoke_root/node-failure.txt"

# Seed unrelated Codex content before first init. These shared namespaces must survive
# install, reinstallation, and uninstall byte-for-byte.
mkdir -p "$smoke_root/.codex/agents" "$smoke_root/.agents/skills/custom-skill" "$smoke_root/.agents/skills/smaqit.input-business" "$smoke_root/.vscode" "$smoke_root/docs/designs/business"
printf '%s\n' 'custom_config = true' > "$smoke_root/.codex/config.toml"
printf '%s\n' 'name = "custom-agent"' > "$smoke_root/.codex/agents/custom-agent.toml"
printf '%s\n' '---' 'name: custom-skill' 'description: Unrelated sentinel skill.' '---' > "$smoke_root/.agents/skills/custom-skill/SKILL.md"
printf '%s\n' 'custom neighbor inside a smaqit-named skill directory' > "$smoke_root/.agents/skills/smaqit.input-business/custom-note.txt"
printf '%s\n' 'user-owned-rule/' > "$smoke_root/.gitignore"
printf '%s\n' '{' '  // unrelated MCP configuration must survive' '  "servers": {"custom-server": {"command": "custom"}}' '}' > "$smoke_root/.vscode/mcp.json"
printf '%s\n' 'user-owned-design-sentinel' > "$smoke_root/docs/designs/business/sentinel.txt"

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
first_init_output="$(printf 'y\n' | "$binary" init "$smoke_root")"
grep -Fq '✓ PlantUML MCP configuration and stdio transport are ready' <<< "$first_init_output"
grep -Fq '✓ Verified PlantUML MCP configuration and local stdio transport' <<< "$first_init_output"
assert_file_equal "$first_owned_agent" "$smoke_root/.codex/agents/$first_owned_name" "confirmed first-install conflict"

assert_owned_tree_matches "$repo_root/installer/agents-copilot" "$smoke_root/.github/agents" "Copilot agents"
assert_owned_tree_matches "$repo_root/installer/skills-copilot" "$smoke_root/.github/skills" "Copilot skills"
assert_owned_tree_matches "$repo_root/installer/agents-claude" "$smoke_root/.claude/agents" "Claude agents"
assert_owned_tree_matches "$repo_root/installer/commands-claude" "$smoke_root/.claude/commands" "Claude commands"
assert_owned_tree_matches "$repo_root/installer/skills-claude" "$smoke_root/.claude/skills" "Claude skills"
assert_owned_tree_matches "$repo_root/installer/agents-codex" "$smoke_root/.codex/agents" "Codex agents"
assert_owned_tree_matches "$repo_root/installer/skills-codex" "$smoke_root/.agents/skills" "Codex skills"
test -f "$smoke_root/.smaqit/templates/designs/business.template.md"
test -f "$smoke_root/.smaqit/tools/plantuml/plantuml-mcp-js-0.2.0_resvg-wasm-2.6.2_noto-sans-5.3.0_opaque-png-1/node_modules/@plantuml/mcp-js/server.js"
grep -Fxq 'user-owned-rule/' "$smoke_root/.gitignore"
test "$(grep -Fxc '.smaqit/tools/' "$smoke_root/.gitignore")" -eq 1
git -C "$smoke_root" check-ignore -q .smaqit/tools/plantuml/plantuml-mcp-js-0.2.0_resvg-wasm-2.6.2_noto-sans-5.3.0_opaque-png-1/node_modules/@plantuml/mcp-js/server.js
if git -C "$smoke_root" check-ignore -q docs/designs/business/sentinel.txt; then
  echo "[ERROR] Canonical design artifacts must not be ignored" >&2
  exit 1
fi
tracked_runtime="$smoke_root/.smaqit/tools/plantuml/plantuml-mcp-js-0.2.0_resvg-wasm-2.6.2_noto-sans-5.3.0_opaque-png-1/node_modules/@plantuml/mcp-js/server.js"
git -C "$smoke_root" add -f "$tracked_runtime"
git -C "$smoke_root" ls-files --error-unmatch "$tracked_runtime" >/dev/null
for layer in business functional stack infrastructure coverage; do
  test -d "$smoke_root/docs/designs/$layer"
done
grep -Fq 'custom-server' "$smoke_root/.vscode/mcp.json"
grep -Fq 'smaqit-plantuml' "$smoke_root/.vscode/mcp.json"

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
design_authors = {"business", "functional", "stack", "infrastructure", "coverage"}
phase_agents = {"development", "deployment", "validation"}
checked_agents = design_authors | phase_agents
for expected in expected_agents:
    installed = root / ".codex" / "agents" / expected.name
    data = tomllib.loads(installed.read_text())
    missing = {key for key in required_agent_fields if not data.get(key)}
    if missing:
        raise SystemExit(f"{installed}: missing required non-empty fields: {sorted(missing)}")
    short_name = expected.name.removeprefix("smaqit.").removesuffix(".toml")
    if short_name in checked_agents:
        instructions = data["developer_instructions"]
        if short_name in design_authors:
            if data.get("tools", {}).get("view_image") is not True:
                raise SystemExit(f"{installed}: design author does not enable view_image")
            if "## Mandatory Visual Design" in instructions:
                raise SystemExit(f"{installed}: visual design must not be a peer-level agent section")
            required_contract = {"**Visual Design:**", "Invoke `smaqit.design-validate`", "Visual design gate failure"}
            if not all(marker in instructions for marker in required_contract):
                raise SystemExit(f"{installed}: distributed visual design contract is incomplete")
        else:
            required_contract = {"automatic readiness gate", "linked PlantUML source", "Author, render, visually review, attest, or repair designs"}
            if not all(marker in instructions for marker in required_contract):
                raise SystemExit(f"{installed}: phase design-consumption contract is incomplete")
            if data.get("tools", {}).get("view_image") is not None:
                raise SystemExit(f"{installed}: implementation agent must not enable view_image")
        server = data.get("mcp_servers", {}).get("smaqit-plantuml")
        if short_name in design_authors:
            if not server or server.get("command") != "smaqit" or server.get("args") != ["mcp", "plantuml"] or server.get("required") is not True:
                raise SystemExit(f"{installed}: mandatory PlantUML MCP configuration is missing")
        elif server is not None:
            raise SystemExit(f"{installed}: implementation agent must not receive PlantUML MCP authoring access")

for platform, directory, suffix in (
    ("Copilot", root / ".github" / "agents", ".agent.md"),
    ("Claude", root / ".claude" / "agents", ".md"),
):
    for short_name in sorted(checked_agents):
        path = directory / f"smaqit.{short_name}{suffix}"
        text = path.read_text()
        if short_name in design_authors:
            if "## Mandatory Visual Design" in text:
                raise SystemExit(f"{path}: visual design must not be a peer-level agent section")
            required_contract = {"**Visual Design:**", "Invoke `smaqit.design-validate`", "Visual design gate failure"}
            if not all(marker in text for marker in required_contract):
                raise SystemExit(f"{path}: distributed visual design contract is incomplete")
        else:
            required_contract = {"automatic readiness gate", "linked PlantUML source", "Author, render, visually review, attest, or repair designs"}
            if not all(marker in text for marker in required_contract):
                raise SystemExit(f"{path}: phase design-consumption contract is incomplete")
        _, frontmatter, _ = text.split("---", 2)
        metadata = yaml.safe_load(frontmatter)
        tools = metadata.get("tools", [])
        if platform == "Copilot":
            if short_name in design_authors:
                required = {"read/viewImage", "smaqit-plantuml/check_syntax", "smaqit-plantuml/render_diagram"}
                if not required.issubset(set(tools)):
                    raise SystemExit(f"{path}: mandatory image/MCP tools are missing")
            server = metadata.get("mcp-servers", {}).get("smaqit-plantuml")
            if short_name in design_authors:
                if not server or server.get("type") != "local" or server.get("command") != "smaqit" or server.get("args") != ["mcp", "plantuml"]:
                    raise SystemExit(f"{path}: Copilot cloud PlantUML MCP configuration is missing")
            elif server is not None or {"read/viewImage", "smaqit-plantuml/check_syntax", "smaqit-plantuml/render_diagram"} & set(tools):
                raise SystemExit(f"{path}: implementation agent must not receive image/MCP design access")
        else:
            servers = metadata.get("mcpServers", [])
            server = next((entry.get("smaqit-plantuml") for entry in servers if isinstance(entry, dict) and "smaqit-plantuml" in entry), None)
            authoring_tools = {"mcp__smaqit-plantuml__check_syntax", "mcp__smaqit-plantuml__render_diagram"}
            if short_name in design_authors:
                if not server or server.get("command") != "smaqit" or server.get("args") != ["mcp", "plantuml"] or not authoring_tools.issubset(set(tools)):
                    raise SystemExit(f"{path}: mandatory PlantUML MCP configuration is missing")
            elif server is not None or authoring_tools & set(tools):
                raise SystemExit(f"{path}: implementation agent must not receive PlantUML MCP authoring access")

expected_skills = {
    path.name for path in (repo / "installer" / "skills-codex").iterdir() if path.is_dir()
}
if len(expected_skills) != 26:
    raise SystemExit(f"expected 26 generated Codex skills, found {len(expected_skills)}")

for skill_name in sorted(expected_skills):
    skill_file = root / ".agents" / "skills" / skill_name / "SKILL.md"
    text = skill_file.read_text()
    if not text.startswith("---\n"):
        raise SystemExit(f"{skill_file}: missing YAML frontmatter")
    _, frontmatter, _ = text.split("---", 2)
    metadata = yaml.safe_load(frontmatter)
    if not isinstance(metadata, dict) or not metadata.get("name") or not metadata.get("description"):
        raise SystemExit(f"{skill_file}: missing name or description")

print("[OK] 9 Codex agents and 26 Codex skills parsed successfully")
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
grep -Fq 'custom-server' "$smoke_root/.vscode/mcp.json"
grep -Fq 'smaqit-plantuml' "$smoke_root/.vscode/mcp.json"

(
  cd "$smoke_root"
  "$binary" validate
)

# Existing active specifications without canonical designs must fail validation
# and the automatic phase-readiness gate.
migration_spec="$smoke_root/specs/business/migration-check.md"
printf '%s\n' '---' 'id: BUS-MIGRATION' 'status: implemented' 'created: 2026-08-03' '---' '' '# UC1-MIGRATION: Migration check' '' '## Acceptance Criteria' '' '- **BUS-MIGRATION-001**: The existing spec requires a visual design.' > "$migration_spec"
if (cd "$smoke_root" && "$binary" validate) > "$smoke_root/migration-validation.txt" 2>&1; then
  echo "[ERROR] Active specification without a design passed strict migration validation" >&2
  exit 1
fi
grep -Fq 'DESIGN-ARTIFACT-MISSING' "$smoke_root/migration-validation.txt"
if (cd "$smoke_root" && "$binary" plan --phase=develop) > "$smoke_root/migration-plan.txt" 2>&1; then
  echo "[ERROR] Design-blocked phase plan passed its automatic readiness gate" >&2
  exit 1
fi
grep -Fq 'Phase design readiness failed' "$smoke_root/migration-plan.txt"
grep -Fq 'DESIGN-ARTIFACT-MISSING' "$smoke_root/migration-plan.txt"
status_output="$(cd "$smoke_root" && "$binary" status)"
grep -Fq 'Design gates: 1 active specification(s)' <<< "$status_output"
rm -- "$migration_spec" "$smoke_root/migration-validation.txt" "$smoke_root/migration-plan.txt"

# Exercise the released CLI against only its materialized embedded runtime.
design_spec="$smoke_root/specs/business/smoke-design.md"
design_source="$smoke_root/docs/designs/business/dsg-bus-smoke-use-case.md"
design_image="$smoke_root/docs/designs/business/dsg-bus-smoke-use-case.png"
printf '%s\n' '---' 'id: BUS-SMOKE' 'status: draft' 'created: 2026-08-03' '---' '' '# UC1-SMOKE: Design smoke test' '' '## Design References' '' '- [DSG-BUS-SMOKE-USE-CASE](../../docs/designs/business/dsg-bus-smoke-use-case.md) · [Image](../../docs/designs/business/dsg-bus-smoke-use-case.png)' '' '## Acceptance Criteria' '' '- **BUS-SMOKE-001**: The embedded runtime renders this design.' > "$design_spec"
printf '%s\n' '---' 'id: DSG-BUS-SMOKE-USE-CASE' 'status: draft' 'created: 2026-08-03' 'layer: business' 'diagram_type: use-case' 'notation: plantuml' 'specifications:' '  - ../../../specs/business/smoke-design.md' 'requirements:' '  - BUS-SMOKE-001' 'source_sha256: ""' 'image_sha256: ""' 'visual_validation:' '  status: pending' '  validated_at: null' '  source_sha256: null' '  image_sha256: null' '---' '' '```plantuml' '@startuml' 'actor User' 'rectangle System {' '  usecase Smoke' '}' 'User --> Smoke' '@enduml' '```' > "$design_source"
(
  cd "$smoke_root"
  "$binary" design render docs/designs/business/dsg-bus-smoke-use-case.md
  test -s "$design_image"
  "$binary" design attest docs/designs/business/dsg-bus-smoke-use-case.md
  "$binary" design validate docs/designs/business/dsg-bus-smoke-use-case.md
  "$binary" mcp verify
  plan_output="$("$binary" plan --phase=develop)"
  if [[ "$plan_output" != 'specs/business/smoke-design.md' ]]; then
    echo "[ERROR] Ready plan output changed its path-only contract: $plan_output" >&2
    exit 1
  fi
)
rm -- "$design_spec" "$design_source" "$design_image"

# A linked Git worktree receives tracked smaqit configuration but not the ignored
# runtime. Every local design command must bootstrap its own verified bundle.
worktree_source="$smoke_root/worktree-source"
worktree_child="$smoke_root/worktree-child"
git init -q "$worktree_source"
git -C "$worktree_source" config user.email "smoke@example.test"
git -C "$worktree_source" config user.name "Smoke Test"
"$binary" init "$worktree_source" >/dev/null
for dir in specs/business specs/functional specs/stack specs/infrastructure specs/coverage docs/designs/business docs/designs/functional docs/designs/stack docs/designs/infrastructure docs/designs/coverage; do
  touch "$worktree_source/$dir/.gitkeep"
done
git -C "$worktree_source" add -A
git -C "$worktree_source" commit -qm "Initialize smaqit"
git -C "$worktree_source" worktree add -q -b toolchain-child "$worktree_child"
if [[ -e "$worktree_child/.smaqit/tools" ]]; then
  echo "[ERROR] Ignored runtime unexpectedly appeared in a fresh Git worktree" >&2
  exit 1
fi
(
  cd "$worktree_child"
  "$binary" validate
  "$binary" mcp verify
)
test -f "$worktree_child/.smaqit/tools/plantuml/plantuml-mcp-js-0.2.0_resvg-wasm-2.6.2_noto-sans-5.3.0_opaque-png-1/node_modules/@plantuml/mcp-js/server.js"
git -C "$worktree_child" check-ignore -q .smaqit/tools/plantuml/plantuml-mcp-js-0.2.0_resvg-wasm-2.6.2_noto-sans-5.3.0_opaque-png-1/node_modules/@plantuml/mcp-js/server.js
git -C "$worktree_source" worktree remove --force "$worktree_child"
git -C "$worktree_source" branch -D toolchain-child >/dev/null

# Reinstallation must detect exact owned conflicts. Cancellation preserves the modified
# file; confirmation restores the generated artifact.
owned_agent="$(find "$repo_root/installer/agents-codex" -maxdepth 1 -type f | sort | head -1)"
owned_agent_name="$(basename "$owned_agent")"
printf '%s\n' 'cancel-sentinel' > "$smoke_root/.codex/agents/$owned_agent_name"
runtime_lock="$smoke_root/.smaqit/tools/plantuml/plantuml-mcp-js-0.2.0_resvg-wasm-2.6.2_noto-sans-5.3.0_opaque-png-1/package-lock.json"
printf '%s\n' 'corrupt-runtime-sentinel' > "$runtime_lock"
printf 'n\n' | "$binary" init "$smoke_root"
grep -Fq 'cancel-sentinel' "$smoke_root/.codex/agents/$owned_agent_name"
grep -Fq 'corrupt-runtime-sentinel' "$runtime_lock"
printf 'y\n' | "$binary" init "$smoke_root"
assert_file_equal "$owned_agent" "$smoke_root/.codex/agents/$owned_agent_name" "confirmed reinstall"
grep -Fxq 'user-owned-rule/' "$smoke_root/.gitignore"
test "$(grep -Fxc '.smaqit/tools/' "$smoke_root/.gitignore")" -eq 1
git -C "$smoke_root" ls-files --error-unmatch "$tracked_runtime" >/dev/null
(
  cd "$smoke_root"
  "$binary" validate
)

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
grep -Fq 'custom-server' "$smoke_root/.vscode/mcp.json"
if grep -Fq 'smaqit-plantuml' "$smoke_root/.vscode/mcp.json"; then
  echo "[ERROR] Owned MCP configuration survived uninstall" >&2
  exit 1
fi
grep -Fq 'user-owned-design-sentinel' "$smoke_root/docs/designs/business/sentinel.txt"

echo "[PASS] Installer smoke test completed successfully"
