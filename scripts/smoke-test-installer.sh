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
  rm -rf -- "$smoke_root"
}
trap cleanup EXIT

export HOME="$smoke_root/home"
export COPILOT_HOME="$smoke_root/copilot"
export CLAUDE_CONFIG_DIR="$smoke_root/claude"
export CODEX_HOME="$smoke_root/codex"
project="$smoke_root/project"
mkdir -p "$HOME" "$project"

assert_file() {
  if [[ ! -f "$1" ]]; then
    echo "[ERROR] Missing $2: $1" >&2
    exit 1
  fi
}

assert_absent() {
  if [[ -e "$1" ]]; then
    echo "[ERROR] Unexpected project-local artifact: $1" >&2
    exit 1
  fi
}

echo "[INFO] Installer: $binary"
echo "[INFO] Temporary project: $smoke_root"

"$binary" --install-global
assert_file "$COPILOT_HOME/agents/smaqit.business.agent.md" "global Copilot agent"
assert_file "$CLAUDE_CONFIG_DIR/agents/smaqit.business.md" "global Claude agent"
assert_file "$CLAUDE_CONFIG_DIR/commands/smaqit.development.md" "global Claude command"
assert_file "$CLAUDE_CONFIG_DIR/skills/smaqit.feature-new/SKILL.md" "global Claude skill"
assert_file "$CODEX_HOME/agents/smaqit.business.toml" "global Codex agent"
assert_file "$HOME/.agents/skills/smaqit.feature-new/SKILL.md" "shared global skill"
grep -Fq 'Author, render, visually review, attest, or repair designs' "$CODEX_HOME/agents/smaqit.development.toml"
grep -Fq '${HOME}/.agents/skills' "$HOME/.agents/skills/smaqit.feature-new/SKILL.md"
grep -Fq '${CLAUDE_CONFIG_DIR:-$HOME/.claude}/skills' "$CLAUDE_CONFIG_DIR/skills/smaqit.feature-new/SKILL.md"

mkdir -p "$HOME/.agents/skills/custom-skill"
printf '%s\n' 'user-owned content' > "$HOME/.agents/skills/custom-skill/SKILL.md"

git init -q "$project"
init_output="$(cd "$project" && "$binary" init)"
grep -Fq "Initializing smaqit project in $project..." <<< "$init_output"
assert_file "$project/.smaqit/templates/specs/business.template.md" "project template"
assert_file "$project/.smaqit/tools/plantuml/plantuml-mcp-js-0.2.0_resvg-wasm-2.6.2_noto-sans-5.3.0_opaque-png-1/node_modules/@plantuml/mcp-js/server.js" "project PlantUML runtime"
for path in .github/agents .github/skills .claude .codex/agents .agents/skills; do
  assert_absent "$project/$path"
done
(cd "$project" && "$binary" validate >/dev/null && "$binary" mcp verify >/dev/null)

# Retain the core project-local design lifecycle gate alongside global payload
# assertions: render, visual attestation, and phase readiness must still work.
design_spec="$project/specs/business/smoke-design.md"
design_source="$project/docs/designs/business/dsg-bus-smoke-use-case.md"
printf '%s\n' \
  '---' 'id: BUS-SMOKE-001' 'status: draft' 'created: 2026-08-11' '---' '' \
  '# Smoke design' '' '## Design References' '' \
  '- [DSG-BUS-SMOKE-USE-CASE](../../docs/designs/business/dsg-bus-smoke-use-case.md) · [Image](../../docs/designs/business/dsg-bus-smoke-use-case.png)' '' \
  '## Acceptance Criteria' '' '- **BUS-SMOKE-001**: The embedded runtime renders this design.' > "$design_spec"
printf '%s\n' \
  '---' 'id: DSG-BUS-SMOKE-USE-CASE' 'status: draft' 'created: 2026-08-11' \
  'layer: business' 'diagram_type: use-case' 'notation: plantuml' 'specifications:' \
  '  - ../../../specs/business/smoke-design.md' 'requirements:' '  - BUS-SMOKE-001' \
  'source_sha256: ""' 'image_sha256: ""' 'visual_validation:' '  status: pending' \
  '  validated_at: null' '  source_sha256: null' '  image_sha256: null' '---' '' \
  '```plantuml' '@startuml' 'actor User' 'rectangle System {' '  usecase Smoke' '}' \
  'User --> Smoke' '@enduml' '```' > "$design_source"
(cd "$project" && "$binary" design render "${design_source#$project/}" && "$binary" design attest "${design_source#$project/}" && "$binary" design validate "${design_source#$project/}")

if "$binary" >/dev/null 2>&1; then
  :
fi

(cd "$project" && printf 'y\nn\n' | "$binary" uninstall >/dev/null)
assert_absent "$COPILOT_HOME/agents/smaqit.business.agent.md"
assert_absent "$HOME/.agents/skills/smaqit.feature-new/SKILL.md"
assert_file "$HOME/.agents/skills/custom-skill/SKILL.md" "user-owned global skill"

echo "[PASS] Global installer smoke test completed successfully"
