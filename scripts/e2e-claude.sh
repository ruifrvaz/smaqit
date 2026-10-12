#!/usr/bin/env bash
# Claude Code end-to-end test: install smaqit into an isolated scratch HOME and
# CLAUDE_CONFIG_DIR, then drive headless `claude -p` through scripted scenarios
# that build the hello-site fixture through smaqit's Development phase.
#
# Every run spends real model usage. Credentials, in order of precedence:
#   1. CLAUDE_CODE_OAUTH_TOKEN or ANTHROPIC_API_KEY from the environment
#   2. otherwise, the access token of your current claude.ai login (subscription).
#      Only `claudeAiOauth.accessToken` is read from your real credentials file and passed
#      to the run as CLAUDE_CODE_OAUTH_TOKEN. Nothing is written to the scratch folder, and
#      the refresh token is never used, so the run cannot rotate your live login.
#
# Usage: scripts/e2e-claude.sh [--binary PATH] [--scenarios LIST]
#   SCENARIOS        comma-separated: happy-path,needs-input,assisted,self-invocation (default all)
#   E2E_BUDGET_USD   spend cap per claude turn (default 10)
#   E2E_TIMEOUT      wall-clock seconds per claude turn (default 3600)
#   E2E_MAX_TURNS    main-thread turn cap per claude turn (default 200)
#   E2E_PERMISSION_MODE claude --permission-mode for every turn (default auto: Claude Code's
#                    classifier allows ordinary commands and denies risky ones; nothing is bypassed)
#   E2E_ALLOWED_TOOLS   override the pre-approved tool list (space-separated)
#   KEEP_E2E_DIR=1   keep the scratch folder after the run
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
fixture="$script_dir/e2e/hello-site"
log_check="$script_dir/e2e/log_check.py"

binary_input="$repo_root/installer/dist/smaqit-dev"
scenarios="${SCENARIOS:-happy-path,needs-input,assisted,self-invocation}"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --binary) binary_input="$2"; shift 2 ;;
    --scenarios) scenarios="$2"; shift 2 ;;
    *) echo "[ERROR] Unknown argument: $1" >&2; exit 2 ;;
  esac
done

budget_usd="${E2E_BUDGET_USD:-10}"
turn_timeout="${E2E_TIMEOUT:-3600}"
max_turns="${E2E_MAX_TURNS:-200}"
permission_mode="${E2E_PERMISSION_MODE:-auto}"
min_claude_version="2.1.268"
default_allowed_tools=(
  Read Write Edit Glob Grep TodoWrite Task Agent Skill SendMessage
  "Bash(smaqit:*)" "Bash(node:*)" "Bash(git:*)" "Bash(ls:*)" "Bash(mkdir:*)" "Bash(cat:*)"
  "Bash(find:*)" "Bash(grep:*)" "Bash(head:*)" "Bash(tail:*)" "Bash(wc:*)" "Bash(test:*)" "Bash(pwd)"
  mcp__smaqit-plantuml__check_syntax mcp__smaqit-plantuml__render_diagram
)
if [[ -n "${E2E_ALLOWED_TOOLS:-}" ]]; then
  read -r -a allowed_tools <<< "$E2E_ALLOWED_TOOLS"
else
  allowed_tools=("${default_allowed_tools[@]}")
fi

fail() { echo "[ERROR] $*" >&2; exit 1; }
info() { echo "[INFO] $*"; }

# --- Preflight (before touching HOME) --------------------------------------------

if [[ "$binary_input" = /* ]]; then binary="$binary_input"; else binary="$(cd "$(dirname "$binary_input")" && pwd)/$(basename "$binary_input")"; fi
[[ -x "$binary" ]] || fail "smaqit binary is missing or not executable: $binary (run 'make -C installer build')"

claude_bin="$(command -v claude || true)"
node_bin="$(command -v node || true)"
python_bin="$(command -v python3 || true)"
git_bin="$(command -v git || true)"
[[ -n "$claude_bin" ]] || fail "Claude Code CLI ('claude') not found on PATH"
[[ -n "$node_bin" ]] || fail "Node.js not found on PATH (Node 22+ required)"
[[ -n "$python_bin" ]] || fail "python3 not found on PATH (needed for session-log checks)"
[[ -n "$git_bin" ]] || fail "git not found on PATH"

node_major="$("$node_bin" -p 'process.versions.node.split(".")[0]')"
(( node_major >= 22 )) || fail "Node.js 22+ required, found $("$node_bin" --version)"

claude_version="$("$claude_bin" --version 2>/dev/null | awk '{print $1}')"
if [[ "$(printf '%s\n%s\n' "$min_claude_version" "$claude_version" | sort -V | head -1)" != "$min_claude_version" ]]; then
  fail "Claude Code $min_claude_version+ required (for --permission-prompts none and 'auth status' configDirectory), found ${claude_version:-unknown}"
fi

IFS=',' read -r -a selected <<< "$scenarios"
for scenario in "${selected[@]}"; do
  case "$scenario" in
    happy-path|needs-input|assisted|self-invocation) ;;
    *) fail "Unknown scenario: '$scenario' (choose from happy-path, needs-input, assisted, self-invocation)" ;;
  esac
done

auth_source="environment"
if [[ -z "${CLAUDE_CODE_OAUTH_TOKEN:-}" && -z "${ANTHROPIC_API_KEY:-}" ]]; then
  # Reuse the current claude.ai login: read only the short-lived access token, before HOME changes.
  real_credentials="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/.credentials.json"
  [[ -f "$real_credentials" ]] || fail "No credentials: log in to Claude Code with claude.ai, or set CLAUDE_CODE_OAUTH_TOKEN ('claude setup-token') or ANTHROPIC_API_KEY"
  token_minutes="$("$python_bin" -c 'import json,sys,time; o=json.load(open(sys.argv[1])).get("claudeAiOauth") or {}; print(int((o.get("expiresAt",0)/1000-time.time())//60) if o.get("accessToken") else -1)' "$real_credentials")"
  (( token_minutes >= 0 )) || fail "The current login has no claude.ai access token; set CLAUDE_CODE_OAUTH_TOKEN ('claude setup-token') instead"
  (( token_minutes >= 30 )) || fail "The current claude.ai access token expires in ${token_minutes} min; start any interactive 'claude' session to refresh it, then rerun"
  CLAUDE_CODE_OAUTH_TOKEN="$("$python_bin" -c 'import json,sys; print(json.load(open(sys.argv[1]))["claudeAiOauth"]["accessToken"])' "$real_credentials")"
  export CLAUDE_CODE_OAUTH_TOKEN
  auth_source="current claude.ai login (access token valid ${token_minutes} min)"
  if (( token_minutes * 60 < turn_timeout )); then
    echo "[WARN] The access token expires in ${token_minutes} min, before E2E_TIMEOUT; a long turn may fail with 401" >&2
  fi
fi

# --- Isolation --------------------------------------------------------------------

e2e_root="$(mktemp -d "${TMPDIR:-/tmp}/smaqit-e2e-claude.XXXXXX")"
evidence_root="$repo_root/installer/test/e2e-claude/$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -p "$evidence_root"
cleanup() {
  if [[ "${KEEP_E2E_DIR:-0}" == "1" ]]; then
    info "Preserving scratch folder: $e2e_root"
  else
    rm -rf -- "$e2e_root"
  fi
}
trap cleanup EXIT

export HOME="$e2e_root/home"
export CLAUDE_CONFIG_DIR="$e2e_root/claude"   # absolute: claude-code#78988
export COPILOT_HOME="$e2e_root/copilot"
export CODEX_HOME="$e2e_root/codex"
export CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS="$(( turn_timeout * 1000 ))"   # wait for background work: claude-code#85066
unset CLAUDE_CODE_SUBPROCESS_ENV_SCRUB                                    # strips CLAUDE_CONFIG_DIR: claude-code#91020
mkdir -p "$HOME" "$CLAUDE_CONFIG_DIR" "$e2e_root/bin"
ln -s "$binary" "$e2e_root/bin/smaqit"
scratch_path="$e2e_root/bin:$(dirname "$node_bin"):$(dirname "$claude_bin"):$(dirname "$git_bin"):/usr/local/bin:/usr/bin:/bin"
export PATH="$scratch_path"

# Isolation must hold before any model call: Claude Code reads its config from the scratch folder.
auth_json="$("$claude_bin" auth status 2>&1 || true)"   # auth errors can land on stdout: claude-code#73067
printf '%s\n' "$auth_json" > "$evidence_root/auth-status.json"
config_dir="$("$python_bin" -c 'import json,sys; print(json.loads(sys.stdin.read()).get("configDirectory",""))' <<< "$auth_json" 2>/dev/null || true)"
auth_usable="$("$python_bin" -c 'import json,sys; s=json.loads(sys.stdin.read()); print("yes" if s.get("loggedIn") or s.get("authMethod") not in (None, "", "none") else "no")' <<< "$auth_json" 2>/dev/null || true)"
[[ "$config_dir" == "$CLAUDE_CONFIG_DIR" ]] || fail "Claude Code is not isolated: configDirectory '$config_dir' != '$CLAUDE_CONFIG_DIR'"
[[ "$auth_usable" == "yes" ]] || fail "Claude Code reports no usable login with the provided token (see $evidence_root/auth-status.json)"
info "Isolated Claude Code $claude_version in $CLAUDE_CONFIG_DIR (credentials: $auth_source)"

"$binary" --install-global > "$evidence_root/install.log"
for path in agents/smaqit.development.md agents/smaqit.business.md commands/smaqit.development.md skills/smaqit.input-business/SKILL.md; do
  [[ -f "$CLAUDE_CONFIG_DIR/$path" ]] || fail "Install did not place $path under the scratch CLAUDE_CONFIG_DIR"
done

# --- Checks -----------------------------------------------------------------------

failures=0
scenario_failures=0
check() {  # check <evidence-file> <label> <command...>
  local file="$1" label="$2"; shift 2
  if "$@" >/dev/null 2>&1; then
    echo "PASS  $label" | tee -a "$file"
  else
    echo "FAIL  $label" | tee -a "$file"
    scenario_failures=$((scenario_failures + 1))
  fi
}

spec_count() { find "$1/specs/$2" -name '*.md' -type f 2>/dev/null | wc -l; }
has_specs() { [[ "$(spec_count "$1" "$2")" -gt 0 ]]; }
no_specs() { [[ -z "$(find "$1/specs" -name '*.md' -type f 2>/dev/null)" ]]; }
plan_empty() { local out; out="$(cd "$1" && smaqit plan --phase=develop)" && [[ -z "$out" ]]; }
status_complete() { (cd "$1" && smaqit status) | grep -Fq 'Phase 1 (Develop): ✓ Complete'; }
site_built() { [[ -n "$(find "$1" -path "$1/.smaqit" -prune -o -name index.html -type f -print)" ]]; }
tests_pass() {  # at least one test ran and none failed; node --test exits 0 when it finds no tests
  local out; out="$(cd "$1" && node --test 2>&1)" || return 1
  grep -Eq '^# pass [1-9]' <<< "$out" && grep -Fxq '# fail 0' <<< "$out"
}
report_exists() { compgen -G "$1/.smaqit/reports/development-phase-report-*.md" >/dev/null; }
validate_ok() { (cd "$1" && smaqit validate); }

develop_checks() {  # develop_checks <project> <checks-file>
  check "$2" "business spec written" has_specs "$1" business
  check "$2" "functional spec written" has_specs "$1" functional
  check "$2" "stack spec written" has_specs "$1" stack
  check "$2" "smaqit validate passes" validate_ok "$1"
  check "$2" "smaqit plan --phase=develop has nothing pending" plan_empty "$1"
  check "$2" "smaqit status shows Develop complete" status_complete "$1"
  check "$2" "index.html built" site_built "$1"
  check "$2" "node --test runs at least one test and none fail" tests_pass "$1"
  check "$2" "development report written" report_exists "$1"
}

# --- Turns ------------------------------------------------------------------------

last_session=""
run_turn() {  # run_turn <project> <evidence-dir> <turn-no> <prompt> [resume-session-id]
  local project="$1" dir="$2" turn="$3" prompt="$4" resume="${5:-}" code=0
  local log="$dir/turn-$turn.jsonl"
  local args=(-p "$prompt" --output-format stream-json --verbose
    --permission-mode "$permission_mode" --permission-prompts none --disallowedTools AskUserQuestion
    --allowedTools "${allowed_tools[@]}"
    --max-turns "$max_turns" --max-budget-usd "$budget_usd")
  [[ -n "$resume" ]] && args+=(--resume "$resume")
  info "  turn $turn${resume:+ (resume $resume)}"
  (cd "$project" && timeout "$turn_timeout" "$claude_bin" "${args[@]}") > "$log" 2> "$dir/turn-$turn.stderr" || code=$?
  echo "$code" > "$dir/turn-$turn.exit"
  "$python_bin" "$log_check" summary "$log" > "$dir/turn-$turn.summary.json" || true
  last_session="$("$python_bin" "$log_check" session-id "$log" || true)"
  if "$python_bin" "$log_check" errored "$log" > "$dir/turn-$turn.error" 2>/dev/null; then
    echo "WARN  turn $turn ended with an error: $(cat "$dir/turn-$turn.error")" | tee -a "$dir/checks.txt"
  fi
  [[ -n "$last_session" ]] || echo "WARN  turn $turn produced no session id (exit $code)" | tee -a "$dir/checks.txt"
}

signal_in() { "$python_bin" "$log_check" signal "$1" "$2"; }
delegated_in() { "$python_bin" "$log_check" delegated "$1" "$2"; }
# A spec counts as regenerated only if its content changed. Completing the phase legitimately
# rewrites the frontmatter (status: implemented), ticks acceptance criteria ([ ] -> [x]), and
# links the design-sequence diagram the Development agent generates after implementation, so
# the digest ignores the frontmatter block, checkbox state, and design-sequence reference lines.
spec_digest() {
  "$python_bin" -c 'import hashlib,re,sys
text = open(sys.argv[1], encoding="utf-8").read()
text = re.sub(r"\A---\n.*?\n---\n", "", text, flags=re.S)
text = re.sub(r"^(\s*[-*] )\[[ xX!]\]", r"\1[ ]", text, flags=re.M)
text = re.sub(r"^.*docs/designs/design-sequence/.*\n?", "", text, flags=re.M)
print(hashlib.sha256(text.encode()).hexdigest())' "$1"
}
hash_specs() {  # hash_specs <project>: one "<digest> <path>" line per spec
  local path
  while IFS= read -r -d '' path; do
    echo "$(spec_digest "$1/$path") $path"
  done < <(cd "$1" && find specs -name '*.md' -type f -print0 | sort -z)
}
hashes_unchanged() {  # hashes_unchanged <project> <snapshot>
  local digest path
  while read -r digest path; do
    [[ -f "$1/$path" && "$(spec_digest "$1/$path")" == "$digest" ]] || return 1
  done < "$2"
}

new_project() {  # new_project <scenario> -> prints project path
  local project="$e2e_root/projects/$1"
  mkdir -p "$project"
  git -C "$project" init -q
  (cd "$project" && smaqit init) > "$evidence_root/$1/init.log"
  grep -Fxq '## When to use smaqit' "$project/AGENTS.md" || fail "$1: AGENTS.md lacks the 'When to use smaqit' block"
  echo "$project"
}

finish_scenario() {  # finish_scenario <scenario> <project>
  local dir="$evidence_root/$1"
  [[ -d "$CLAUDE_CONFIG_DIR/projects" ]] && cp -r "$CLAUDE_CONFIG_DIR/projects" "$dir/transcripts" 2>/dev/null || true
  (cd "$2" && find . -path ./.smaqit/tools -prune -o -path ./.git -prune -o -type f -print | sort) > "$dir/tree.txt"
  if (( scenario_failures > 0 )); then
    echo "FAIL  $1 ($scenario_failures check(s) failed)" | tee -a "$evidence_root/summary.txt"
    failures=$((failures + 1))
  else
    echo "PASS  $1" | tee -a "$evidence_root/summary.txt"
  fi
}

# --- Scenarios --------------------------------------------------------------------

scenario_happy_path() {
  local dir="$evidence_root/happy-path" project
  project="$(new_project happy-path)" || fail "happy-path: project setup failed"
  run_turn "$project" "$dir" 1 "/smaqit.development $(cat "$fixture/requirements.md")"
  develop_checks "$project" "$dir/checks.txt"
  finish_scenario happy-path "$project"
}

scenario_needs_input() {
  local dir="$evidence_root/needs-input" project
  project="$(new_project needs-input)" || fail "needs-input: project setup failed"
  run_turn "$project" "$dir" 1 "/smaqit.development $(cat "$fixture/thin-requirements.md")"
  check "$dir/checks.txt" "turn 1 wrote no specification" no_specs "$project"
  check "$dir/checks.txt" "turn 1 returned NEEDS-INPUT" signal_in "$dir/turn-1.jsonl" NEEDS-INPUT
  if [[ -n "$last_session" ]]; then
    run_turn "$project" "$dir" 2 "$(cat "$fixture/answers.md")" "$last_session"
  fi
  develop_checks "$project" "$dir/checks.txt"
  finish_scenario needs-input "$project"
}

scenario_assisted() {
  local dir="$evidence_root/assisted" project turn=1 approved="$e2e_root/assisted-approved.sha256"
  project="$(new_project assisted)" || fail "assisted: project setup failed"
  : > "$approved"
  run_turn "$project" "$dir" 1 "/smaqit.development Use assisted mode: I want to review and approve each specification layer before you continue. $(cat "$fixture/requirements.md")"
  check "$dir/checks.txt" "turn 1 returned REVIEW-REQUESTED" signal_in "$dir/turn-1.jsonl" REVIEW-REQUESTED
  check "$dir/checks.txt" "turn 1 stopped after the business layer" has_specs "$project" business
  check "$dir/checks.txt" "turn 1 wrote no functional spec yet" test "$(spec_count "$project" functional)" -eq 0
  while signal_in "$dir/turn-$turn.jsonl" REVIEW-REQUESTED && [[ -n "$last_session" ]] && (( turn < 6 )); do
    hash_specs "$project" > "$approved"
    rm -rf "$dir/approved-before-turn-$((turn + 1))" && cp -r "$project/specs" "$dir/approved-before-turn-$((turn + 1))"
    turn=$((turn + 1))
    run_turn "$project" "$dir" "$turn" "$(cat "$fixture/approve.md")" "$last_session"
    check "$dir/checks.txt" "turn $turn left approved specs unchanged" hashes_unchanged "$project" "$approved"
    diff -ru "$dir/approved-before-turn-$turn" "$project/specs" > "$dir/approved-specs-turn-$turn.diff" || true
  done
  develop_checks "$project" "$dir/checks.txt"
  finish_scenario assisted "$project"
}

scenario_self_invocation() {
  local dir="$evidence_root/self-invocation" project
  project="$(new_project self-invocation)" || fail "self-invocation: project setup failed"
  run_turn "$project" "$dir" 1 "$(cat "$fixture/plain-request.md")"
  check "$dir/checks.txt" "plain request delegated to smaqit-development" delegated_in "$dir/turn-1.jsonl" smaqit-development
  develop_checks "$project" "$dir/checks.txt"
  finish_scenario self-invocation "$project"
}

for scenario in "${selected[@]}"; do
  mkdir -p "$evidence_root/$scenario"
  scenario_failures=0
  info "Scenario: $scenario"
  case "$scenario" in
    happy-path) scenario_happy_path ;;
    needs-input) scenario_needs_input ;;
    assisted) scenario_assisted ;;
    self-invocation) scenario_self_invocation ;;
    *) fail "Unknown scenario: $scenario" ;;
  esac
done

[[ ! -e "$CLAUDE_CONFIG_DIR/.credentials.json" ]] || { echo "FAIL  a credential file was written to the scratch config dir" | tee -a "$evidence_root/summary.txt"; failures=$((failures + 1)); }

echo
info "Evidence: $evidence_root"
cat "$evidence_root/summary.txt"
if (( failures > 0 )); then
  echo "[FAIL] $failures scenario check group(s) failed" >&2
  exit 1
fi
echo "[PASS] Claude Code e2e completed successfully"
