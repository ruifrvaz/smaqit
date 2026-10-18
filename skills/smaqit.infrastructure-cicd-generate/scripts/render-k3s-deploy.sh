#!/usr/bin/env bash
# Renders the existing-k3s deploy workflow for smaqit.infrastructure-cicd-generate.
#
# Input is the Infrastructure spec's Environments table as `smaqit spec environments --json`
# prints it (validated there). Output is one deploy.yml for every environment: the dispatch
# options list each machine slug, K3S_TARGETS carries each machine's row, and steps branch on
# the dispatched machine's row at run time. The mesh-join step is kept, with the spec's pinned
# action and input names, only when some environment is private-mesh. The imagePullSecret step is
# kept, with the registry host, only when the spec's Container Registry is private. The spec's
# Deployer Exec row (default denied) becomes the deploy job's DEPLOYER_EXEC.
#
# Usage: render-k3s-deploy.sh <environments.json> <app-slug> [--default-env <environment>] [--template <path>]
#   <environments.json>  output of `smaqit spec environments --json` ("-" reads stdin)
#   <app-slug>           the app slug; the Namespace is app-<app-slug>
#   --default-env        the environment post-merge-deploy.yml deploys (default: prod, or the
#                        only environment when there is one)
#   --template           defaults to ../assets/deploy.yml.k3s.template next to this script
#   Writes the workflow to stdout.
#   Exit 0 — rendered. Exit 1 — the input cannot be rendered (named on stderr). Exit 2 — usage.

set -euo pipefail

usage() {
  echo "Usage: render-k3s-deploy.sh <environments.json> <app-slug> [--default-env <environment>] [--template <path>]" >&2
  exit 2
}
fail() {
  echo "render-k3s-deploy: $*" >&2
  exit 1
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE="${SCRIPT_DIR}/../assets/deploy.yml.k3s.template"
DEFAULT_ENV=""
POSITIONAL=()
while [ "$#" -gt 0 ]; do
  case "$1" in
    --default-env)
      [ "$#" -ge 2 ] && [ -n "$2" ] || usage
      DEFAULT_ENV="$2"; shift 2 ;;
    --template)
      [ "$#" -ge 2 ] && [ -n "$2" ] || usage
      TEMPLATE="$2"; shift 2 ;;
    -) POSITIONAL+=("$1"); shift ;;
    -*) usage ;;
    *) POSITIONAL+=("$1"); shift ;;
  esac
done
[ "${#POSITIONAL[@]}" -eq 2 ] || usage
INPUT="${POSITIONAL[0]}"
APP_SLUG="${POSITIONAL[1]}"

command -v jq >/dev/null 2>&1 || fail "jq not found on PATH"
[[ "$APP_SLUG" =~ ^[a-z0-9]([a-z0-9-]*[a-z0-9])?$ ]] || fail "app slug '${APP_SLUG}' must be lowercase letters, digits, and inner hyphens"
[ -f "$TEMPLATE" ] || fail "template not found: ${TEMPLATE}"

if [ "$INPUT" = "-" ]; then
  SPEC="$(cat)"
else
  [ -f "$INPUT" ] || fail "input not found: ${INPUT}"
  SPEC="$(cat "$INPUT")"
fi
jq -e '.schemaVersion == 1 and (.environments | type == "array" and length > 0)' >/dev/null 2>&1 <<<"$SPEC" \
  || fail "input is not 'smaqit spec environments --json' output with at least one environment"

DUPLICATE="$(jq -r '[.environments[].machineSlug] | group_by(.) | map(select(length > 1) | .[0]) | join(", ")' <<<"$SPEC")"
[ -z "$DUPLICATE" ] || fail "machine slug(s) ${DUPLICATE} back more than one environment; each environment needs its own machine slug (its own GitHub Environment)"

if [ -z "$DEFAULT_ENV" ]; then
  if jq -e '.environments | any(.environment == "prod")' >/dev/null <<<"$SPEC"; then
    DEFAULT_ENV=prod
  elif [ "$(jq '.environments | length' <<<"$SPEC")" -eq 1 ]; then
    DEFAULT_ENV="$(jq -r '.environments[0].environment' <<<"$SPEC")"
  else
    fail "no 'prod' environment; pass --default-env <environment> (one of: $(jq -r '[.environments[].environment] | join(", ")' <<<"$SPEC"))"
  fi
fi
DEFAULT_MACHINE="$(jq -r --arg e "$DEFAULT_ENV" '.environments[] | select(.environment == $e) | .machineSlug' <<<"$SPEC")"
[ -n "$DEFAULT_MACHINE" ] || fail "--default-env '${DEFAULT_ENV}' is not an environment in the spec"

OPTIONS="$(jq -r '.environments[].machineSlug' <<<"$SPEC")"
TARGETS="$(jq '[.environments[] | {key: .machineSlug, value: {environment, reachability, ingressMode, host, ingressClass, clusterIssuer, healthPath, meshSetupKeySecret, meshWaitForHost}}] | from_entries' <<<"$SPEC")"

DEPLOYER_EXEC="$(jq -r '.deployerExec // "denied" | if . == "" then "denied" else . end' <<<"$SPEC")"
case "$DEPLOYER_EXEC" in
  denied|allowed) ;;
  *) fail "deployerExec '${DEPLOYER_EXEC}' must be denied or allowed" ;;
esac

KEEP_PULL=0
REGISTRY_HOST=""
REGISTRY_VISIBILITY="$(jq -r '.registry.visibility // "public"' <<<"$SPEC")"
case "$REGISTRY_VISIBILITY" in
  public) ;;
  private)
    KEEP_PULL=1
    REGISTRY_HOST="$(jq -r '.registry.host // ""' <<<"$SPEC")"
    [[ "$REGISTRY_HOST" =~ ^[a-z0-9]([a-z0-9.-]*[a-z0-9])?(:[0-9]{1,5})?$ ]] \
      || fail "a private Container Registry needs a registry host (got '${REGISTRY_HOST}')" ;;
  *) fail "Container Registry visibility '${REGISTRY_VISIBILITY}' must be public or private" ;;
esac

KEEP_MESH=0
MESH_ACTION="" MESH_KEY_INPUT="" MESH_WAIT_INPUT=""
if jq -e '.environments | any(.reachability == "private-mesh")' >/dev/null <<<"$SPEC"; then
  KEEP_MESH=1
  MESH_ACTION="$(jq -r '.mesh.action // ""' <<<"$SPEC")"
  MESH_KEY_INPUT="$(jq -r '.mesh.setupKeyInput // ""' <<<"$SPEC")"
  MESH_WAIT_INPUT="$(jq -r '.mesh.waitForInput // ""' <<<"$SPEC")"
  [ -n "$MESH_ACTION" ] && [ -n "$MESH_KEY_INPUT" ] && [ -n "$MESH_WAIT_INPUT" ] \
    || fail "a private-mesh environment needs the spec's Mesh Join Action, Mesh Setup-Key Input, and Mesh Wait-For Input"
fi

OUTPUT="$(OPTIONS="$OPTIONS" TARGETS="$TARGETS" APP_SLUG="$APP_SLUG" DEFAULT_MACHINE="$DEFAULT_MACHINE" \
  KEEP_MESH="$KEEP_MESH" MESH_ACTION="$MESH_ACTION" MESH_KEY_INPUT="$MESH_KEY_INPUT" MESH_WAIT_INPUT="$MESH_WAIT_INPUT" \
  KEEP_PULL="$KEEP_PULL" REGISTRY_HOST="$REGISTRY_HOST" DEPLOYER_EXEC="$DEPLOYER_EXEC" \
  awk '
  # Literal (non-regex) token replacement, so no value is ever read as a pattern.
  function repl(s, token, value,    out, i) {
    out = ""
    while ((i = index(s, token)) > 0) {
      out = out substr(s, 1, i - 1) value
      s = substr(s, i + length(token))
    }
    return out s
  }
  function block(text, line,    n, parts, k, indent) {
    indent = substr(line, 1, index(line, "__") - 1)
    n = split(text, parts, "\n")
    for (k = 1; k <= n; k++) print indent parts[k]
  }
  # A dropped block also drops the blank line after it, so no double blank is left behind.
  dropped && $0 == "" { dropped = 0; next }
  { dropped = 0 }
  /# smaqit:mesh-join:begin/ { in_mesh = 1; next }
  /# smaqit:mesh-join:end/ { in_mesh = 0; dropped = ENVIRON["KEEP_MESH"] != "1"; next }
  in_mesh && ENVIRON["KEEP_MESH"] != "1" { next }
  /# smaqit:pull-secret:begin/ { in_pull = 1; next }
  /# smaqit:pull-secret:end/ { in_pull = 0; dropped = ENVIRON["KEEP_PULL"] != "1"; next }
  in_pull && ENVIRON["KEEP_PULL"] != "1" { next }
  index($0, "__MACHINE_SLUG_OPTIONS__") {
    n = split(ENVIRON["OPTIONS"], slugs, "\n")
    indent = substr($0, 1, index($0, "__") - 1)
    for (k = 1; k <= n; k++) print indent "- " slugs[k]
    next
  }
  index($0, "__K3S_TARGETS__") { block(ENVIRON["TARGETS"], $0); next }
  {
    line = $0
    line = repl(line, "__APP_SLUG__", ENVIRON["APP_SLUG"])
    line = repl(line, "__DEFAULT_MACHINE_SLUG__", ENVIRON["DEFAULT_MACHINE"])
    line = repl(line, "__MESH_JOIN_ACTION__", ENVIRON["MESH_ACTION"])
    line = repl(line, "__MESH_SETUP_KEY_INPUT__", ENVIRON["MESH_KEY_INPUT"])
    line = repl(line, "__MESH_WAIT_FOR_INPUT__", ENVIRON["MESH_WAIT_INPUT"])
    line = repl(line, "__REGISTRY_HOST__", ENVIRON["REGISTRY_HOST"])
    line = repl(line, "__DEPLOYER_EXEC__", ENVIRON["DEPLOYER_EXEC"])
    print line
  }
' "$TEMPLATE")"

# Every render-time token must be gone; the remaining __TOKEN__s are stamped at run time.
LEFT="$(grep -oE '__(APP_SLUG|DEFAULT_MACHINE_SLUG|MACHINE_SLUG_OPTIONS|K3S_TARGETS|MESH_[A-Z_]+|REGISTRY_HOST|DEPLOYER_EXEC)__' <<<"$OUTPUT" | sort -u | tr '\n' ' ' || true)"
[ -z "$LEFT" ] || fail "unrendered token(s) left in the workflow: ${LEFT}"
if [ "$KEEP_MESH" != "1" ] && grep -q 'smaqit:mesh-join' "$TEMPLATE" && grep -q 'Join the mesh' <<<"$OUTPUT"; then
  fail "the mesh-join step survived an all-public render"
fi
if [ "$KEEP_PULL" != "1" ] && grep -q 'smaqit:pull-secret' "$TEMPLATE" && grep -q 'Create imagePullSecret' <<<"$OUTPUT"; then
  fail "the imagePullSecret step survived a public-registry render"
fi
printf '%s\n' "$OUTPUT"
