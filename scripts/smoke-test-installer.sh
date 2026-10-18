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

# An upgrade from a release that shipped the skill under its old name, beside a
# skill another tool (smaqit-extensions) installed into the same directories.
for skills in "$CLAUDE_CONFIG_DIR/skills" "$HOME/.agents/skills"; do
  mkdir -p "$skills/smaqit.new-greenfield-project/references" "$skills/smaqit.task-start"
  printf -- '---\nname: smaqit.new-greenfield-project\ndescription: old\n---\n' > "$skills/smaqit.new-greenfield-project/SKILL.md"
  printf '%s\n' 'old diagrams' > "$skills/smaqit.new-greenfield-project/references/diagrams.md"
  printf -- '---\nname: smaqit.task-start\ndescription: extension\n---\n' > "$skills/smaqit.task-start/SKILL.md"
done

"$binary" --install-global
for skills in "$CLAUDE_CONFIG_DIR/skills" "$HOME/.agents/skills"; do
  if [[ -e "$skills/smaqit.new-greenfield-project" ]]; then
    echo "[ERROR] Retired skill survived the install: $skills/smaqit.new-greenfield-project" >&2
    exit 1
  fi
  assert_file "$skills/smaqit.project-new/SKILL.md" "renamed global skill"
  assert_file "$skills/smaqit.task-start/SKILL.md" "another tool's global skill"
  assert_file "$skills/.smaqit-installed.json" "global install manifest"
done
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

# Project instructions: both files carry smaqit's marked section, and AGENTS.md
# tells the coding agent when to use smaqit.
for file in AGENTS.md CLAUDE.md; do
  grep -Fxq '<!-- smaqit:instructions:begin -->' "$project/$file"
  grep -Fxq '<!-- smaqit:instructions:end -->' "$project/$file"
done
grep -Fxq '## When to use smaqit' "$project/AGENTS.md"

# Deployment slot bindings: the skeleton resolves every slot to its default,
# and a re-init (the same path `smaqit update` takes) never overwrites it.
assert_file "$project/.smaqit/slots.yml" "project slot bindings skeleton"
(cd "$project" && "$binary" slots validate >/dev/null)
# existing-k3s plans are refused until the Infrastructure spec's Environments row
# for the environment is complete, then resolve with that row as the target.
if (cd "$project" && "$binary" slots plan --phase=deploy --mode=existing-k3s --env=test --json) 2>/dev/null; then
  echo "existing-k3s plan resolved without an Environments table" >&2
  exit 1
fi
k3s_spec="$project/specs/infrastructure/smoke-k3s.md"
printf '%s\n' '---' 'id: INF-SMOKE-K3S' 'status: draft' 'created: 2026-01-01' '---' '' '# Smoke' '' \
  '## Environments' '' \
  '| Environment | Machine Slug | Host | Reachability | Ingress Mode | Ingress Class | ClusterIssuer | Health Path | Mesh Setup-Key Secret | Mesh Wait-For Host |' \
  '|---|---|---|---|---|---|---|---|---|---|' \
  '| test | edge-01 | — | private-mesh | platform-owned | — | — | /health.json | MESH_SETUP_KEY | edge-01.mesh.internal |' '' \
  '## Constraints' '' '| Constraint | Description | Impact |' '|---|---|---|' \
  '| Mesh Join Action | example-org/mesh-join@v1.0.0 | joins the mesh |' \
  '| Mesh Setup-Key Input | setup-key | action input |' \
  '| Mesh Wait-For Input | wait-for | action input |' \
  '| Probe Client | wget | in-cluster probe |' > "$k3s_spec"
(cd "$project" && "$binary" slots plan --phase=deploy --mode=existing-k3s --env=test --json) \
  | grep -Fq '"skill": "smaqit.infrastructure-request-k3s-onboarding"'
(cd "$project" && "$binary" spec environments --env=test --json) | grep -Fq '"reachability": "private-mesh"'
rm -f "$k3s_spec"
printf '%s\n' 'schemaVersion: 1' 'phase: deploy' 'slots:' '  tls: none' > "$project/.smaqit/slots.yml"
# The same re-init refreshes an outdated smaqit section in AGENTS.md while keeping
# the user's own content around it.
printf '%s\n' '# User-owned heading' '<!-- smaqit:instructions:begin -->' 'Outdated smaqit section.' \
  '<!-- smaqit:instructions:end -->' 'User-owned trailing line.' > "$project/AGENTS.md"
reinit_output="$(cd "$project" && printf 'y\n' | "$binary" init)"
grep -Fq '.smaqit/slots.yml preserved' <<< "$reinit_output"
grep -Fxq '  tls: none' "$project/.smaqit/slots.yml"
grep -Fq 'AGENTS.md updated' <<< "$reinit_output"
grep -Fxq '# User-owned heading' "$project/AGENTS.md"
grep -Fxq 'User-owned trailing line.' "$project/AGENTS.md"
grep -Fxq '## When to use smaqit' "$project/AGENTS.md"
if grep -Fq 'Outdated smaqit section.' "$project/AGENTS.md"; then
  echo "[ERROR] Re-init did not refresh the outdated smaqit section in AGENTS.md" >&2
  exit 1
fi

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
  '```plantuml' '@startuml' 'title DSG-BUS-SMOKE-USE-CASE' 'actor User' 'rectangle System {' '  usecase Smoke' '}' \
  'User --> Smoke' '@enduml' '```' > "$design_source"
(cd "$project" && "$binary" design render "${design_source#$project/}" && "$binary" design attest "${design_source#$project/}" && "$binary" design validate "${design_source#$project/}")

if "$binary" >/dev/null 2>&1; then
  :
fi

(cd "$project" && printf 'y\nn\n' | "$binary" uninstall >/dev/null)
assert_file "$project/.smaqit/slots.yml" "project slot bindings kept by uninstall"
assert_absent "$COPILOT_HOME/agents/smaqit.business.agent.md"
assert_absent "$HOME/.agents/skills/smaqit.feature-new/SKILL.md"
assert_file "$HOME/.agents/skills/custom-skill/SKILL.md" "user-owned global skill"
assert_file "$HOME/.agents/skills/smaqit.task-start/SKILL.md" "another tool's global skill kept by uninstall"
assert_absent "$HOME/.agents/skills/.smaqit-installed.json"
assert_absent "$CLAUDE_CONFIG_DIR/skills/.smaqit-installed.json"

echo "[PASS] Global installer smoke test completed successfully"
