package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// These tests run the machine-registration and machine-Environment scripts
// shipped in ../skills against stub vault and gh binaries placed first on PATH.
// The stubs keep their state in files and log every argument vector, so the
// tests can prove idempotency and that no secret value ever reaches a command
// line or the terminal. No real Vault or GitHub is contacted.

const stubVault = `#!/usr/bin/env bash
set -u
store="$STUB_STATE/vault"
printf '%s\0' "$@" >> "$STUB_STATE/vault.argv"
printf '\n' >> "$STUB_STATE/vault.argv"
case "$1 ${2:-}" in
  "token lookup") exit 0 ;;
  "status -format=json") echo '{"sealed": false, "initialized": true}'; exit 0 ;;
  "secrets list")
    if [ -f "$STUB_STATE/kv-enabled" ]; then echo '{"secret/": {"type": "kv"}}'; else echo '{}'; fi
    exit 0 ;;
  "secrets enable") touch "$STUB_STATE/kv-enabled"; echo "$*" >> "$STUB_STATE/secrets-enable.log"; exit 0 ;;
esac
[ "$1" = "status" ] && exit 0
[ "$1" = "kv" ] || exit 2
op="$2"; shift 2
field=""
case "${1:-}" in -field=*) field="${1#-field=}"; shift ;; esac
path="$1"; shift
dir="$store/$path"
case "$op" in
  get)
    [ -d "$dir" ] || { echo "No value found at $path" >&2; exit 2; }
    if [ -n "$field" ]; then
      [ -f "$dir/$field" ] || { echo "field $field not present" >&2; exit 1; }
      cat "$dir/$field"
    else
      ls "$dir"
    fi ;;
  put|patch)
    if [ "$op" = "patch" ]; then [ -d "$dir" ] || exit 2; else rm -rf "$dir"; fi
    mkdir -p "$dir"
    for kv in "$@"; do
      k="${kv%%=*}"; v="${kv#*=}"
      case "$v" in
        -) cat > "$dir/$k" ;;
        @*) cp "${v#@}" "$dir/$k" ;;
        *) printf '%s' "$v" > "$dir/$k" ;;
      esac
    done ;;
  *) exit 2 ;;
esac
`

const stubGh = `#!/usr/bin/env bash
set -u
gh="$STUB_STATE/gh"
printf '%s\0' "$@" >> "$STUB_STATE/gh.argv"
printf '\n' >> "$STUB_STATE/gh.argv"
case "$1 ${2:-}" in
  "auth status") exit 0 ;;
esac
if [ "$1 ${2:-}" = "run view" ]; then
  # The run record comes from a fixture file; an unset fixture is an unknown run.
  [ -n "${STUB_GH_RUN_JSON:-}" ] || { echo "could not find any workflow run with ID: $3" >&2; exit 1; }
  cat "$STUB_GH_RUN_JSON"; exit 0
fi
if [ "$1" = "api" ]; then
  method=GET; path=""
  shift
  while [ "$#" -gt 0 ]; do
    case "$1" in
      -X) method="$2"; shift 2 ;;
      -i|--silent) shift ;;
      --jq) shift 2 ;;
      *) path="$1"; shift ;;
    esac
  done
  case "$path" in
    */compare/*)
      [ -n "${STUB_GH_COMPARE_STATUS:-}" ] || { echo "gh: Not Found (HTTP 404)" >&2; exit 1; }
      echo "$STUB_GH_COMPARE_STATUS"; exit 0 ;;
  esac
  env="${path#*/environments/}"; rest=""
  case "$env" in */*) rest="${env#*/}"; env="${env%%/*}" ;; esac
  if [ "$method" = "PUT" ]; then
    [ "${STUB_GH_FORBID_CREATE:-}" = 1 ] && { echo "gh: Resource not accessible by personal access token (HTTP 403)" >&2; exit 1; }
    mkdir -p "$gh/envs/$env"; exit 0
  fi
  if [ ! -d "$gh/envs/$env" ]; then echo "HTTP/2.0 404 Not Found"; echo "gh: Not Found (HTTP 404)" >&2; exit 1; fi
  if [ "$rest" = "variables" ] && [ "${STUB_GH_FORBID_VARIABLES:-}" = 1 ]; then
    echo "HTTP/2.0 403 Forbidden"; echo "gh: Resource not accessible by personal access token (HTTP 403)" >&2; exit 1
  fi
  echo "HTTP/2.0 200 OK"; echo; echo "{}"; exit 0
fi
if [ "$1 ${2:-}" = "workflow run" ]; then
  # Like the real gh: without --repo/-R or GH_REPO it infers the repository from git.
  case " $* " in *" --repo "*|*" -R "*) exit 0 ;; esac
  [ -n "${GH_REPO:-}" ] && exit 0
  git rev-parse --git-dir >/dev/null 2>&1 || { echo "failed to run git: fatal: not a git repository (or any of the parent directories): .git" >&2; exit 1; }
  exit 0
fi
kind="$1"; verb="$2"; shift 2
name=""; env=""; body=""; has_body=0
while [ "$#" -gt 0 ]; do
  case "$1" in
    --env) env="$2"; shift 2 ;;
    -R) shift 2 ;;
    --body) body="$2"; has_body=1; shift 2 ;;
    --json|--jq) shift 2 ;;
    *) name="$1"; shift ;;
  esac
done
dir="$gh/envs/$env/$kind"
case "$verb" in
  set)
    [ -d "$gh/envs/$env" ] || { echo "HTTP 404" >&2; exit 1; }
    mkdir -p "$dir"
    if [ "$has_body" = 1 ]; then printf '%s' "$body" > "$dir/$name"; else cat > "$dir/$name"; fi ;;
  list) [ -d "$dir" ] && ls "$dir" ;;
esac
exit 0
`

const stubTerraform = `#!/usr/bin/env bash
set -u
printf '%s\0' "$@" >> "$STUB_STATE/terraform.argv"
printf '\n' >> "$STUB_STATE/terraform.argv"
case "$1" in
  plan)
    for a in "$@"; do
      case "$a" in -out=*) printf 'stub-plan' > "${a#-out=}" ;; esac
    done
    pwd > "$STUB_STATE/terraform.pwd"
    exit "${STUB_TF_PLAN_EXIT:-2}" ;;
  show)
    [ "$(cat "$3" 2>/dev/null)" = stub-plan ] || { echo "no plan at $3" >&2; exit 1; }
    cat "$STUB_TF_PLAN_JSON" ;;
esac
`

// stubKubectl answers the namespace guard's and exec-refused's calls.
// STUB_KUBECTL_MODE: ok (every can-i is "no"), unreachable, unreachable-twice
// (the first two --list calls cannot connect, then the API answers), rejected
// (Unauthorized), forbidden (the identity is not bound yet), allow-nodes (get nodes is "yes"), flaky (one can-i call errors after the API
// answered), exec-yes (both pods/exec checks are "yes"), exec-get-yes (only
// get pods/exec is "yes"), or exec-unclear (the pods/exec checks error).
// STUB_KUBECTL_EXEC sets a real exec's result: forbidden (default), runs,
// command-fails, notfound, or hang.
const stubKubectl = `#!/usr/bin/env bash
set -u
mode="${STUB_KUBECTL_MODE:-ok}"
printf '%s\n' "$*" >> "$STUB_STATE/kubectl.calls"
if [ "$1" = config ] && [ "$2" = view ]; then
  printf '%s' "${STUB_KUBECTL_NAMESPACE:-app-shop}"; exit 0
fi
if [ "$1" = auth ] && [ "$2" = can-i ]; then
  case "$mode" in
    unreachable) echo "Unable to connect to the server: dial tcp 100.64.0.7:6443: i/o timeout" >&2; exit 1 ;;
    rejected) echo "error: You must be logged in to the server (Unauthorized)" >&2; exit 1 ;;
    forbidden) echo 'Error from server (Forbidden): selfsubjectrulesreviews.authorization.k8s.io is forbidden: User "gha:123:edge-01" cannot create resource "selfsubjectrulesreviews"' >&2; exit 1 ;;
    unreachable-twice)
      if [ "$3" = --list ]; then
        n=$(( $(cat "$STUB_STATE/kubectl.reach" 2>/dev/null || echo 0) + 1 )); echo "$n" > "$STUB_STATE/kubectl.reach"
        if [ "$n" -le 2 ]; then echo "Unable to connect to the server: dial tcp 100.64.0.7:6443: connect: connection refused" >&2; exit 1; fi
      fi ;;
  esac
  if [ "$3" = --list ]; then echo "Resources   Non-Resource URLs   Resource Names   Verbs"; exit 0; fi
  case "$mode:$3:$4:${5:-}" in
    allow-nodes:get:nodes:*) echo yes; exit 0 ;;
    flaky:*:storageclasses:*) echo "error: the server was unable to return a response in the time allotted" >&2; exit 1 ;;
    exec-yes:*:pods:--subresource=exec|exec-get-yes:get:pods:--subresource=exec) echo yes; exit 0 ;;
    exec-unclear:*:pods:--subresource=exec) echo "error: context deadline exceeded" >&2; exit 1 ;;
  esac
  echo no; exit 1
fi
if [ "$1" = apply ]; then
  cat > "$STUB_STATE/kubectl.applied"; exit 0
fi
if [ "$1" = exec ]; then
  case "${STUB_KUBECTL_EXEC:-forbidden}" in
    runs) exit 0 ;;
    command-fails) echo 'error: Internal error occurred: error executing command in container: failed to exec in container: OCI runtime exec failed: exec: "true": executable file not found in $PATH: unknown' >&2; exit 1 ;;
    notfound) echo 'Error from server (NotFound): deployments.apps "shop" not found' >&2; exit 1 ;;
    hang) sleep 30; exit 0 ;;
    *) echo 'Error from server (Forbidden): pods "shop-7d9f-abcde" is forbidden: User "system:serviceaccount:app-shop:deployer" cannot create resource "pods/exec" in API group "" in the namespace "app-shop"' >&2; exit 1 ;;
  esac
fi
echo "stub kubectl: unexpected call: $*" >&2
exit 3
`

type scriptHarness struct {
	t     *testing.T
	state string
	work  string
	env   []string
}

func newScriptHarness(t *testing.T) *scriptHarness {
	t.Helper()
	for _, tool := range []string{"bash", "ssh-keygen", "script", "setsid"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not available: %v", tool, err)
		}
	}
	bin := t.TempDir()
	for name, body := range map[string]string{"vault": stubVault, "gh": stubGh, "terraform": stubTerraform, "kubectl": stubKubectl} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(body), 0755); err != nil {
			t.Fatal(err)
		}
	}
	state := t.TempDir()
	return &scriptHarness{
		t:     t,
		state: state,
		work:  t.TempDir(),
		env: append(os.Environ(),
			"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
			"STUB_STATE="+state,
			"VAULT_ADDR=http://127.0.0.1:1",
			"GH_TOKEN=",
		),
	}
}

func skillScript(skill, name string) string {
	return filepath.Join("..", "skills", skill, "scripts", name)
}

// run executes a script with no controlling terminal and closed stdin.
func (h *scriptHarness) run(extraEnv []string, script string, args ...string) (string, error) {
	h.t.Helper()
	abs, err := filepath.Abs(script)
	if err != nil {
		h.t.Fatal(err)
	}
	cmd := exec.Command("setsid", append([]string{"-w", "bash", abs}, args...)...)
	cmd.Dir = h.work
	cmd.Env = append(append([]string{}, h.env...), extraEnv...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// runInteractive executes a script on a pseudo-terminal via script(1),
// answering each prompt only after it appears so hidden reads never echo.
func (h *scriptHarness) runInteractive(script string, answers [][2]string, args ...string) (string, error) {
	h.t.Helper()
	abs, err := filepath.Abs(script)
	if err != nil {
		h.t.Fatal(err)
	}
	line := "bash " + shellQuote(abs)
	for _, a := range args {
		line += " " + shellQuote(a)
	}
	return h.runInteractiveCommand(line, answers)
}

// runInteractiveCommand runs a shell command line on a pseudo-terminal via
// script(1), answering each prompt only after it appears.
func (h *scriptHarness) runInteractiveCommand(line string, answers [][2]string) (string, error) {
	h.t.Helper()
	transcript := filepath.Join(h.t.TempDir(), "typescript")
	cmd := exec.Command("script", "-qfec", line, transcript)
	cmd.Dir = h.work
	cmd.Env = h.env
	stdin, err := cmd.StdinPipe()
	if err != nil {
		h.t.Fatal(err)
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		h.t.Fatal(err)
	}
	go func() {
		defer stdin.Close()
		seen := 0
		for _, a := range answers {
			deadline := time.Now().Add(10 * time.Second)
			for {
				data, _ := os.ReadFile(transcript)
				if idx := bytes.Index(data[seen:], []byte(a[0])); idx >= 0 {
					seen += idx + len(a[0])
					break
				}
				if time.Now().After(deadline) {
					return
				}
				time.Sleep(20 * time.Millisecond)
			}
			time.Sleep(200 * time.Millisecond) // let read -s disable echo
			io.WriteString(stdin, a[1]+"\n")
		}
	}()
	err = cmd.Wait()
	out, _ := os.ReadFile(transcript)
	return string(out), err
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func (h *scriptHarness) vaultValue(path, field string) string {
	h.t.Helper()
	data, err := os.ReadFile(filepath.Join(h.state, "vault", path, field))
	if err != nil {
		return ""
	}
	return string(data)
}

func (h *scriptHarness) seedVault(path string, fields map[string]string) {
	h.t.Helper()
	dir := filepath.Join(h.state, "vault", path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.t.Fatal(err)
	}
	for k, v := range fields {
		if err := os.WriteFile(filepath.Join(dir, k), []byte(v), 0644); err != nil {
			h.t.Fatal(err)
		}
	}
}

func (h *scriptHarness) argv(tool string) string {
	data, _ := os.ReadFile(filepath.Join(h.state, tool+".argv"))
	return string(data)
}

func (h *scriptHarness) requireAbsent(where, text string, secrets ...string) {
	h.t.Helper()
	for _, s := range secrets {
		if strings.Contains(text, s) {
			h.t.Errorf("secret value %q leaked into %s", s, where)
		}
	}
}

func TestRegisterMachineRequiresValidSlug(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-vault-loader", "register-machine.sh")
	for _, args := range [][]string{{}, {"Bad_Slug"}, {"--host", "1.2.3.4"}, {"edge-01", "--bogus"}} {
		if out, err := h.run(nil, script, args...); err == nil {
			t.Errorf("args %q: expected failure, got success:\n%s", args, out)
		}
	}
	if _, err := os.Stat(filepath.Join(h.state, "vault", "secret")); err == nil {
		t.Error("an invalid invocation wrote to Vault")
	}
}

func TestRegisterMachineRegistersThenBackfillsIdempotently(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-vault-loader", "register-machine.sh")
	const (
		cysoID     = "cred-id-visible"
		cysoSecret = "CYSO-SECRET-7f3a"
		tfKey      = "s3-access-visible"
		tfSecret   = "S3-SECRET-91bc"
	)
	out, err := h.runInteractive(script, [][2]string{
		{"app_credential_id:", cysoID},
		{"app_credential_secret:", cysoSecret},
		{"access_key:", tfKey},
		{"secret_key:", tfSecret},
		{"provider [cyso]:", ""},
		{"owner_project", "platform-infra"},
	}, "edge-01")
	if err != nil {
		t.Fatalf("registration failed: %v\n%s", err, out)
	}
	m := "secret/machines/edge-01"
	for path, want := range map[string]string{
		m + "/cyso/app_credential_id":     cysoID,
		m + "/cyso/app_credential_secret": cysoSecret,
		m + "/tfstate/access_key":         tfKey,
		m + "/tfstate/secret_key":         tfSecret,
		m + "/metadata/provider":          "cyso",
		m + "/metadata/owner_project":     "platform-infra",
	} {
		if got := h.vaultValue(filepath.Dir(path), filepath.Base(path)); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
	privateKey := h.vaultValue(m+"/base-ssh", "private_key")
	if !strings.HasPrefix(privateKey, "-----BEGIN OPENSSH PRIVATE KEY-----") || !strings.HasSuffix(privateKey, "\n") {
		t.Errorf("base-ssh private_key not stored byte-exact with its trailing newline")
	}
	if pub := h.vaultValue(m+"/base-ssh", "public_key"); !strings.HasPrefix(pub, "ssh-ed25519 ") || strings.HasSuffix(pub, "\n") {
		t.Errorf("base-ssh public_key = %q", pub)
	}
	if h.vaultValue(m+"/metadata", "host") != "" {
		t.Error("host written before it was supplied")
	}
	if !strings.Contains(out, "--host <fixed_ip>") {
		t.Errorf("missing host backfill hint:\n%s", out)
	}
	h.requireAbsent("vault argv", h.argv("vault"), cysoSecret, tfSecret, privateKey)
	h.requireAbsent("terminal output", out, cysoSecret, tfSecret)

	// Post-apply backfill: no prompts (no terminal at all), every path kept.
	before := h.argv("vault")
	out, err = h.run(nil, script, "edge-01", "--host", "203.0.113.7")
	if err != nil {
		t.Fatalf("backfill failed: %v\n%s", err, out)
	}
	if got := h.vaultValue(m+"/metadata", "host"); got != "203.0.113.7" {
		t.Errorf("host = %q", got)
	}
	if h.vaultValue(m+"/metadata", "owner_project") != "platform-infra" ||
		h.vaultValue(m+"/base-ssh", "private_key") != privateKey ||
		h.vaultValue(m+"/cyso", "app_credential_secret") != cysoSecret {
		t.Error("backfill changed existing fields")
	}
	rerun := strings.TrimPrefix(h.argv("vault"), before)
	if strings.Contains(rerun, "put\x00") {
		t.Errorf("idempotent re-run issued a put:\n%q", rerun)
	}
	if !strings.Contains(rerun, "patch\x00"+m+"/metadata\x00host=203.0.113.7") {
		t.Errorf("backfill did not patch host:\n%q", rerun)
	}
}

func TestRegisterMachineRefusesEmptySecretsAndNonInteractiveRuns(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-vault-loader", "register-machine.sh")
	out, err := h.runInteractive(script, [][2]string{
		{"app_credential_id:", "cred-id"},
		{"app_credential_secret:", ""},
	}, "edge-02")
	if err == nil || !strings.Contains(out, "refusing to write a placeholder secret") {
		t.Fatalf("empty secret accepted: %v\n%s", err, out)
	}
	if h.vaultValue("secret/machines/edge-02/cyso", "app_credential_id") != "" {
		t.Error("cyso written despite an empty secret")
	}

	// No terminal: the first prompt fails instead of writing an empty value.
	h2 := newScriptHarness(t)
	if out, err := h2.run(nil, script, "edge-03"); err == nil {
		t.Fatalf("non-interactive registration succeeded:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(h2.state, "vault", "secret/machines/edge-03/cyso")); err == nil {
		t.Error("cyso written without a terminal")
	}
}

func seedRegisteredMachine(h *scriptHarness, slug, host string) map[string]string {
	m := "secret/machines/" + slug
	values := map[string]string{
		"BASE_SSH_PUBLIC_KEY":                "ssh-ed25519 AAAAPUBLIC base",
		"BASE_SSH_PRIVATE_KEY":               "-----BEGIN OPENSSH PRIVATE KEY-----\nPRIVATE-5d2e\n-----END OPENSSH PRIVATE KEY-----\n",
		"CYSO_APPLICATION_CREDENTIAL_ID":     "cred-id-77",
		"CYSO_APPLICATION_CREDENTIAL_SECRET": "CYSO-SECRET-c4d1",
		"CYSO_S3_ACCESS_KEY":                 "s3-access-12",
		"CYSO_S3_SECRET_KEY":                 "S3-SECRET-e8f0",
	}
	h.seedVault(m+"/base-ssh", map[string]string{"public_key": values["BASE_SSH_PUBLIC_KEY"], "private_key": values["BASE_SSH_PRIVATE_KEY"]})
	h.seedVault(m+"/cyso", map[string]string{"app_credential_id": values["CYSO_APPLICATION_CREDENTIAL_ID"], "app_credential_secret": values["CYSO_APPLICATION_CREDENTIAL_SECRET"]})
	h.seedVault(m+"/tfstate", map[string]string{"access_key": values["CYSO_S3_ACCESS_KEY"], "secret_key": values["CYSO_S3_SECRET_KEY"]})
	meta := map[string]string{"provider": "cyso", "owner_project": "platform-infra"}
	if host != "" {
		meta["host"] = host
	}
	h.seedVault(m+"/metadata", meta)
	return values
}

func (h *scriptHarness) ghValue(env, kind, name string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(h.state, "gh", "envs", env, kind, name))
	return string(data), err == nil
}

func TestSyncMachineEnvMapsMachineCredentials(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-repo-config", "sync-machine-env.sh")
	values := seedRegisteredMachine(h, "edge-01", "203.0.113.7")
	out, err := h.run(nil, script, "acme/platform", "edge-01")
	if err != nil {
		t.Fatalf("sync failed: %v\n%s", err, out)
	}
	for name, want := range values {
		if got, ok := h.ghValue("edge-01", "secret", name); !ok || got != want {
			t.Errorf("secret %s = %q (present %v), want %q", name, got, ok, want)
		}
	}
	for name, want := range map[string]string{"RECONCILE_SSH_USER": "ubuntu", "FIXED_IP": "203.0.113.7"} {
		if got, _ := h.ghValue("edge-01", "variable", name); got != want {
			t.Errorf("variable %s = %q, want %q", name, got, want)
		}
	}
	secrets := []string{values["BASE_SSH_PRIVATE_KEY"], values["CYSO_APPLICATION_CREDENTIAL_SECRET"], values["CYSO_S3_SECRET_KEY"]}
	h.requireAbsent("gh argv", h.argv("gh"), secrets...)
	h.requireAbsent("output", out, secrets...)

	// Idempotent: a second run converges to the same state.
	if out, err := h.run(nil, script, "acme/platform", "edge-01", "--ssh-user", "ops"); err != nil {
		t.Fatalf("re-run failed: %v\n%s", err, out)
	}
	if got, _ := h.ghValue("edge-01", "variable", "RECONCILE_SSH_USER"); got != "ops" {
		t.Errorf("RECONCILE_SSH_USER after re-run = %q", got)
	}
}

func TestSyncMachineEnvBeforeProvisioningSkipsFixedIP(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-repo-config", "sync-machine-env.sh")
	seedRegisteredMachine(h, "edge-01", "")
	out, err := h.run(nil, script, "acme/platform", "edge-01")
	if err != nil {
		t.Fatalf("pre-provision sync failed: %v\n%s", err, out)
	}
	if _, ok := h.ghValue("edge-01", "variable", "FIXED_IP"); ok {
		t.Error("FIXED_IP set without a host")
	}
	if _, ok := h.ghValue("edge-01", "secret", "CYSO_S3_SECRET_KEY"); !ok {
		t.Error("secrets not synced before provisioning")
	}
	if !strings.Contains(out, "FIXED_IP skipped") {
		t.Errorf("missing FIXED_IP notice:\n%s", out)
	}
}

func TestSyncMachineEnvRefusesUnregisteredMachineAndMissingGrants(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-repo-config", "sync-machine-env.sh")
	h.seedVault("secret/machines/edge-01/base-ssh", map[string]string{"public_key": "k", "private_key": "k"})
	out, err := h.run(nil, script, "acme/platform", "edge-01")
	if err == nil || !strings.Contains(out, "register-machine.sh edge-01") {
		t.Fatalf("unregistered machine accepted: %v\n%s", err, out)
	}
	if strings.Contains(h.argv("gh"), "secret\x00set") || strings.Contains(h.argv("gh"), "PUT") {
		t.Error("wrote to GitHub before the Vault preconditions passed")
	}
	if _, err := h.run(nil, script, "acme/platform", "Edge_01"); err == nil {
		t.Error("accepted a slug that needs URL encoding")
	}

	h2 := newScriptHarness(t)
	seedRegisteredMachine(h2, "edge-01", "203.0.113.7")
	out, err = h2.run([]string{"STUB_GH_FORBID_CREATE=1"}, script, "acme/platform", "edge-01")
	if err == nil || !strings.Contains(out, "organization-level") {
		t.Fatalf("403 on Environment create not mapped: %v\n%s", err, out)
	}

	h3 := newScriptHarness(t)
	seedRegisteredMachine(h3, "edge-01", "203.0.113.7")
	out, err = h3.run([]string{"STUB_GH_FORBID_VARIABLES=1"}, script, "acme/platform", "edge-01")
	if err == nil {
		t.Fatalf("missing variables grant not caught:\n%s", out)
	}
	if strings.Contains(h3.argv("gh"), "secret\x00set") {
		t.Error("wrote secrets after the token preflight failed")
	}
}

func TestCheckPatPermissionsMapsSignatures(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-repo-config", "check-pat-permissions.sh")
	exitCode := func(err error) int {
		if err == nil {
			return 0
		}
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		return -1
	}
	out, err := h.run(nil, script, "acme/app", "test")
	if exitCode(err) != 2 || !strings.Contains(out, "environments/test -X PUT") {
		t.Errorf("missing Environment: exit %d\n%s", exitCode(err), out)
	}
	if err := os.MkdirAll(filepath.Join(h.state, "gh", "envs", "test"), 0755); err != nil {
		t.Fatal(err)
	}
	if out, err := h.run(nil, script, "acme/app", "test"); exitCode(err) != 0 {
		t.Errorf("accessible Environment: exit %d\n%s", exitCode(err), out)
	}
	out, err = h.run([]string{"STUB_GH_FORBID_VARIABLES=1"}, script, "acme/app", "test")
	if exitCode(err) != 3 || !strings.Contains(out, "organization-level Secrets and Variables") {
		t.Errorf("forbidden variables: exit %d\n%s", exitCode(err), out)
	}
}

// planChange is one resource_changes entry in the stub plan's JSON.
type planChange struct {
	address string
	actions []string
}

// runPlanGuard runs plan-guard.sh against the stub terraform in a "tf"
// subdirectory, with the plan exiting planExit and reporting changes.
func runPlanGuard(t *testing.T, h *scriptHarness, planExit int, changes []planChange, args ...string) (string, int) {
	t.Helper()
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skipf("jq not available: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(h.work, "tf"), 0755); err != nil {
		t.Fatal(err)
	}
	var entries []string
	for _, c := range changes {
		entries = append(entries, `{"address":"`+c.address+`","change":{"actions":["`+strings.Join(c.actions, `","`)+`"]}}`)
	}
	planJSON := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(planJSON, []byte(`{"resource_changes":[`+strings.Join(entries, ",")+`]}`), 0644); err != nil {
		t.Fatal(err)
	}
	os.Remove(filepath.Join(h.state, "terraform.argv"))
	script := skillScript("smaqit.infrastructure-provision-cyso", "plan-guard.sh")
	out, err := h.run([]string{
		"STUB_TF_PLAN_EXIT=" + strconv.Itoa(planExit),
		"STUB_TF_PLAN_JSON=" + planJSON,
	}, script, append([]string{"tf"}, args...)...)
	code := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("plan-guard did not run: %v\n%s", err, out)
		}
		code = ee.ExitCode()
	}
	return out, code
}

const instance = "module.machine.openstack_compute_instance_v2.this"

func TestPlanGuardWithoutFlagsKeepsBlockingEveryDelete(t *testing.T) {
	h := newScriptHarness(t)
	if out, code := runPlanGuard(t, h, 0, nil); code != 0 || !strings.Contains(out, "no changes") {
		t.Errorf("no changes: exit %d\n%s", code, out)
	}
	if pwd, _ := os.ReadFile(filepath.Join(h.state, "terraform.pwd")); !strings.HasSuffix(strings.TrimSpace(string(pwd)), "/tf") {
		t.Errorf("terraform did not run in the working directory argument: %q", pwd)
	}
	out, code := runPlanGuard(t, h, 2, []planChange{{"openstack_networking_secgroup_v2.app", []string{"update"}}})
	if code != 0 || !strings.Contains(out, "additive/update-only") {
		t.Errorf("update-only: exit %d\n%s", code, out)
	}
	for _, actions := range [][]string{{"delete", "create"}, {"create", "delete"}, {"delete"}} {
		out, code := runPlanGuard(t, h, 2, []planChange{{instance, actions}})
		if code != 1 || !strings.Contains(out, instance) {
			t.Errorf("unflagged %v: exit %d\n%s", actions, code, out)
		}
	}
	if out, code := runPlanGuard(t, h, 1, nil); code != 1 || !strings.Contains(out, "plan failed") {
		t.Errorf("failed plan: exit %d\n%s", code, out)
	}
	if strings.Contains(h.argv("terraform"), "-replace=") {
		t.Error("passed -replace= without --allow-replace")
	}
}

func TestPlanGuardAllowsOnlyTheNamedReplacement(t *testing.T) {
	h := newScriptHarness(t)
	// Both action orders are a replacement (create-before-destroy plans create first).
	for _, actions := range [][]string{{"delete", "create"}, {"create", "delete"}} {
		out, code := runPlanGuard(t, h, 2, []planChange{
			{instance, actions},
			{"module.machine.openstack_compute_volume_attach_v2.data", []string{"update"}},
		}, "--allow-replace", instance)
		if code != 0 || !strings.Contains(out, "replaces only the explicitly allowed") {
			t.Errorf("allowed replacement %v: exit %d\n%s", actions, code, out)
		}
		if !strings.Contains(h.argv("terraform"), "-replace="+instance+"\x00") {
			t.Errorf("-replace not passed to terraform plan: %q", h.argv("terraform"))
		}
	}

	cases := []struct {
		name    string
		exit    int
		changes []planChange
		want    string
	}{
		{"pure destroy of the allowed address", 2, []planChange{{instance, []string{"delete"}}}, instance + ": delete"},
		{"allowed address not replaced", 2, []planChange{{instance, []string{"update"}}}, "does not replace"},
		{"allowed address with no changes", 0, nil, "nothing was replaced"},
		{"another delete beside the allowed replacement", 2, []planChange{
			{instance, []string{"delete", "create"}},
			{"module.machine.openstack_blockstorage_volume_v3.data", []string{"delete", "create"}},
		}, "openstack_blockstorage_volume_v3.data"},
	}
	for _, c := range cases {
		out, code := runPlanGuard(t, h, c.exit, c.changes, "--allow-replace", instance)
		if code != 1 || !strings.Contains(out, c.want) {
			t.Errorf("%s: exit %d, want 1 mentioning %q\n%s", c.name, code, c.want, out)
		}
	}

	// Two listed addresses: both must be replaced.
	second := "module.machine.openstack_compute_keypair_v2.this"
	if out, code := runPlanGuard(t, h, 2, []planChange{{instance, []string{"delete", "create"}}},
		"--allow-replace", instance, "--allow-replace", second); code != 1 || !strings.Contains(out, second) {
		t.Errorf("second allowed address unreplaced: exit %d\n%s", code, out)
	}
}

const breakGlassRule = "openstack_networking_secgroup_rule_v2.break_glass"

func TestPlanGuardAllowsOnlyTheNamedDestroy(t *testing.T) {
	h := newScriptHarness(t)
	out, code := runPlanGuard(t, h, 2, []planChange{
		{breakGlassRule, []string{"delete"}},
		{"openstack_networking_secgroup_v2.app", []string{"update"}},
	}, "--allow-delete", breakGlassRule)
	if code != 0 || !strings.Contains(out, "destroys only the explicitly allowed") || !strings.Contains(out, breakGlassRule) {
		t.Errorf("allowed destroy: exit %d\n%s", code, out)
	}
	if strings.Contains(h.argv("terraform"), "-replace=") {
		t.Errorf("--allow-delete changed the plan arguments: %q", h.argv("terraform"))
	}

	cases := []struct {
		name    string
		exit    int
		changes []planChange
		want    []string
	}{
		{"another delete beside the allowed destroy", 2, []planChange{
			{breakGlassRule, []string{"delete"}},
			{"openstack_networking_secgroup_rule_v2.https", []string{"delete"}},
		}, []string{"openstack_networking_secgroup_rule_v2.https: delete"}},
		{"allowed address replaced instead", 2, []planChange{{breakGlassRule, []string{"create", "delete"}}},
			[]string{"does not purely destroy", breakGlassRule + " (planned: create,delete)"}},
		{"allowed address only updated", 2, []planChange{{breakGlassRule, []string{"update"}}},
			[]string{"does not purely destroy", breakGlassRule + " (planned: update)"}},
		{"allowed address absent from the plan", 2, []planChange{{"openstack_networking_secgroup_v2.app", []string{"update"}}},
			[]string{"does not purely destroy", breakGlassRule + " (planned: not in the plan)"}},
		{"allowed address with no changes", 0, nil, []string{"nothing was replaced or destroyed", breakGlassRule + " (--allow-delete)"}},
	}
	for _, c := range cases {
		out, code := runPlanGuard(t, h, c.exit, c.changes, "--allow-delete", breakGlassRule)
		if code != 1 {
			t.Errorf("%s: exit %d, want 1\n%s", c.name, code, out)
		}
		for _, want := range c.want {
			if !strings.Contains(out, want) {
				t.Errorf("%s: output does not mention %q\n%s", c.name, want, out)
			}
		}
	}

	// Repeatable: every listed address must be destroyed.
	second := "openstack_networking_secgroup_rule_v2.public_port"
	if out, code := runPlanGuard(t, h, 2, []planChange{{breakGlassRule, []string{"delete"}}, {second, []string{"delete"}}},
		"--allow-delete", breakGlassRule, "--allow-delete", second); code != 0 {
		t.Errorf("two allowed destroys: exit %d\n%s", code, out)
	}
	if out, code := runPlanGuard(t, h, 2, []planChange{{breakGlassRule, []string{"delete"}}},
		"--allow-delete", breakGlassRule, "--allow-delete", second); code != 1 || !strings.Contains(out, second+" (planned: not in the plan)") {
		t.Errorf("second allowed address not destroyed: exit %d\n%s", code, out)
	}

	// A pure destroy allowed only as a replacement stays blocked.
	if out, code := runPlanGuard(t, h, 2, []planChange{{breakGlassRule, []string{"delete"}}},
		"--allow-replace", breakGlassRule); code != 1 || !strings.Contains(out, breakGlassRule+": delete") {
		t.Errorf("destroy allowed only as a replacement: exit %d\n%s", code, out)
	}
}

func TestPlanGuardCombinesAllowedDestroyReplaceAndOut(t *testing.T) {
	h := newScriptHarness(t)
	kept := filepath.Join(h.work, "guarded.tfplan")
	out, code := runPlanGuard(t, h, 2, []planChange{
		{instance, []string{"create", "delete"}},
		{breakGlassRule, []string{"delete"}},
	}, "--allow-replace", instance, "--allow-delete", breakGlassRule, "--out", "guarded.tfplan")
	if code != 0 {
		t.Fatalf("combined: exit %d\n%s", code, out)
	}
	for _, want := range []string{"replaces only the explicitly allowed", "destroys only the explicitly allowed", instance, breakGlassRule, "terraform apply " + kept} {
		if !strings.Contains(out, want) {
			t.Errorf("combined PASS does not mention %q\n%s", want, out)
		}
	}
	argv := h.argv("terraform")
	if !strings.Contains(argv, "-replace="+instance+"\x00") || strings.Contains(argv, "-replace="+breakGlassRule) {
		t.Errorf("plan arguments: %q", argv)
	}
	if _, err := os.Stat(kept); err != nil {
		t.Fatalf("plan not kept: %v", err)
	}

	// The allowed replacement does not excuse an unlisted destroy, and the kept plan is discarded.
	out, code = runPlanGuard(t, h, 2, []planChange{
		{instance, []string{"create", "delete"}},
		{breakGlassRule, []string{"delete"}},
	}, "--allow-replace", instance, "--out", "guarded.tfplan")
	if code != 1 || !strings.Contains(out, breakGlassRule+": delete") {
		t.Errorf("unlisted destroy beside an allowed replacement: exit %d\n%s", code, out)
	}
	if _, err := os.Stat(kept); err == nil {
		t.Error("blocked plan left on disk")
	}
}

func TestPlanGuardOutKeepsOnlyAGuardedPlan(t *testing.T) {
	h := newScriptHarness(t)
	kept := filepath.Join(h.work, "guarded.tfplan") // relative to the caller, not the tf directory
	out, code := runPlanGuard(t, h, 2, []planChange{{instance, []string{"delete", "create"}}},
		"--allow-replace", instance, "--out", "guarded.tfplan")
	if code != 0 || !strings.Contains(out, "terraform apply "+kept) {
		t.Fatalf("guarded plan: exit %d\n%s", code, out)
	}
	if data, err := os.ReadFile(kept); err != nil || string(data) != "stub-plan" {
		t.Fatalf("plan not kept at %s: %v", kept, err)
	}
	if _, err := os.Stat(filepath.Join(h.work, "tf", "guarded.tfplan")); err == nil {
		t.Error("--out resolved against the terraform directory instead of the caller's")
	}

	// A blocked plan never survives, even one kept by an earlier run.
	out, code = runPlanGuard(t, h, 2, []planChange{{instance, []string{"delete"}}}, "--out", "guarded.tfplan")
	if code != 1 {
		t.Fatalf("destroy with --out: exit %d\n%s", code, out)
	}
	if _, err := os.Stat(kept); err == nil {
		t.Error("blocked plan left on disk")
	}
}

// The BLOCKED message is what an agent reads when it decides what to do next, so it must
// say stop, and must never invite taking an allow-flag address from the blocked list.
func TestPlanGuardBlockedOutputSaysStopAndReport(t *testing.T) {
	h := newScriptHarness(t)
	for name, run := range map[string]func() (string, int){
		"unlisted delete": func() (string, int) {
			return runPlanGuard(t, h, 2, []planChange{{breakGlassRule, []string{"delete"}}})
		},
		"listed address not destroyed": func() (string, int) {
			return runPlanGuard(t, h, 2, []planChange{{breakGlassRule, []string{"update"}}}, "--allow-delete", breakGlassRule)
		},
	} {
		out, code := run()
		if code != 1 {
			t.Fatalf("%s: exit %d, want 1\n%s", name, code, out)
		}
		for _, want := range []string{"STOP. Do not apply this plan", "address taken from this output", "a human named before this plan", "wait for a human decision"} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: BLOCKED output lacks %q\n%s", name, want, out)
			}
		}
		if strings.Contains(out, "outside of this guarded path") {
			t.Errorf("%s: BLOCKED output still suggests applying outside the guard\n%s", name, out)
		}
	}
}

func TestPlanGuardRejectsBadUsage(t *testing.T) {
	h := newScriptHarness(t)
	for _, args := range [][]string{
		{"--bogus"},
		{"--allow-replace"},
		{"--allow-delete"},
		{"--out"},
		{"extra-dir"},
		{"--allow-delete", breakGlassRule, "--allow-replace", breakGlassRule},
	} {
		if out, code := runPlanGuard(t, h, 0, nil, args...); code != 2 {
			t.Errorf("args %q: exit %d, want 2\n%s", args, code, out)
		}
	}
	if _, err := os.Stat(filepath.Join(h.state, "terraform.argv")); err == nil {
		t.Error("terraform ran despite a usage error")
	}
}

func TestCredentialFileHandsOutOnlyAllowlistedCredentials(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-vault-loader", "credential-file.sh")
	const sshKey = "-----BEGIN OPENSSH PRIVATE KEY-----\nAPP-KEY-77c1\n-----END OPENSSH PRIVATE KEY-----\n"
	h.seedVault("secret/apps/shop/ssh", map[string]string{"private_key": sshKey, "public_key": "ssh-ed25519 AAAA"})
	h.seedVault("secret/apps/shop/platform-repo", map[string]string{"token": "PLATFORM-PAT-3f9e"})
	h.seedVault("secret/legacy-app/ssh", map[string]string{"private_key": sshKey})

	for path, want := range map[string]string{
		"apps/shop/ssh":           sshKey,
		"apps/shop/platform-repo": "PLATFORM-PAT-3f9e",
		"legacy-app/ssh":          sshKey,
	} {
		out, err := h.run(nil, script, path)
		if err != nil {
			t.Fatalf("%s: %v\n%s", path, err, out)
		}
		file := strings.TrimSpace(out)
		h.requireAbsent("stdout of "+path, out, "APP-KEY-77c1", "PLATFORM-PAT-3f9e")
		data, err := os.ReadFile(file)
		if err != nil || string(data) != want {
			t.Errorf("%s: file %q holds %q (%v), want %q", path, file, data, err, want)
		}
		if info, err := os.Stat(file); err != nil || info.Mode().Perm() != 0600 {
			t.Errorf("%s: file mode %v (%v), want 0600", path, info.Mode().Perm(), err)
		}
		if info, err := os.Stat(filepath.Dir(file)); err != nil || info.Mode().Perm() != 0700 {
			t.Errorf("%s: directory mode %v (%v), want 0700", path, info.Mode().Perm(), err)
		}
		os.RemoveAll(filepath.Dir(file))
	}

	// existing-k3s keeps no kubeconfig: CI connects with its own OIDC identity.
	for _, path := range []string{"machines/edge-01/base-ssh", "apps/shop/github", "apps/shop/edge-01/kubeconfig", "organizations/acme/github-package-read", "apps/ssh", "apps/Shop/ssh"} {
		if out, err := h.run(nil, script, path); err == nil {
			t.Errorf("%s: handed out a credential outside the allowlist:\n%s", path, out)
		}
	}
}

func TestTerraformEnvExportsWithoutPrintingAndUnsets(t *testing.T) {
	h := newScriptHarness(t)
	script, _ := filepath.Abs(skillScript("smaqit.infrastructure-vault-loader", "terraform-env.sh"))
	source := func(env, args, after string) (string, error) {
		cmd := exec.Command("bash", "-c", "source "+shellQuote(script)+" "+args+"; rc=$?; "+after+"; exit $rc")
		cmd.Dir = h.work
		cmd.Env = append(append([]string{}, h.env...), strings.Fields(env)...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	dump := `printf 'VAL %s=%s\n' TF_VAR_app_credential_id "${TF_VAR_app_credential_id-}" TF_VAR_app_credential_secret "${TF_VAR_app_credential_secret-}" AWS_ACCESS_KEY_ID "${AWS_ACCESS_KEY_ID-}" AWS_SECRET_ACCESS_KEY "${AWS_SECRET_ACCESS_KEY-}" TF_VAR_github_token "${TF_VAR_github_token-}" TF_VAR_ssh_public_key "${TF_VAR_ssh_public_key-}" > dump.txt`
	readDump := func() map[string]string {
		data, _ := os.ReadFile(filepath.Join(h.work, "dump.txt"))
		vals := map[string]string{}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if kv, ok := strings.CutPrefix(line, "VAL "); ok {
				k, v, _ := strings.Cut(kv, "=")
				vals[k] = v
			}
		}
		return vals
	}

	// Machine scheme: cyso/tfstate only under machines/, github under apps/.
	h.seedVault("secret/machines/edge-01/cyso", map[string]string{"app_credential_id": "id-1", "app_credential_secret": "CYSO-SECRET-aa"})
	h.seedVault("secret/machines/edge-01/tfstate", map[string]string{"access_key": "ak-1", "secret_key": "S3-SECRET-bb"})
	h.seedVault("secret/machines/edge-01/base-ssh", map[string]string{"public_key": "ssh-ed25519 BASE"})
	h.seedVault("secret/apps/platform/github", map[string]string{"token": "GH-TOKEN-cc"})
	out, err := source("PROJECT_SLUG=platform MACHINE_SLUG=edge-01", "", dump)
	if err != nil {
		t.Fatalf("machine scheme: %v\n%s", err, out)
	}
	h.requireAbsent("terraform-env output", out, "CYSO-SECRET-aa", "S3-SECRET-bb", "GH-TOKEN-cc")
	want := map[string]string{
		"TF_VAR_app_credential_id": "id-1", "TF_VAR_app_credential_secret": "CYSO-SECRET-aa",
		"AWS_ACCESS_KEY_ID": "ak-1", "AWS_SECRET_ACCESS_KEY": "S3-SECRET-bb",
		"TF_VAR_github_token": "GH-TOKEN-cc", "TF_VAR_ssh_public_key": "ssh-ed25519 BASE",
	}
	for k, v := range want {
		if got := readDump()[k]; got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}

	// The project's own paths take precedence, as provision-cyso always read them.
	h.seedVault("secret/platform/cyso", map[string]string{"app_credential_id": "legacy-id", "app_credential_secret": "LEGACY-SECRET"})
	if out, err := source("PROJECT_SLUG=platform MACHINE_SLUG=edge-01", "", dump); err != nil || readDump()["TF_VAR_app_credential_id"] != "legacy-id" {
		t.Errorf("project path precedence: %v\n%s", err, out)
	}

	// --unset clears everything.
	if out, err := source("PROJECT_SLUG=platform MACHINE_SLUG=edge-01", "", "source "+shellQuote(script)+" --unset; "+dump); err != nil {
		t.Fatalf("unset: %v\n%s", err, out)
	}
	for k, v := range readDump() {
		if v != "" {
			t.Errorf("%s still set after --unset", k)
		}
	}

	// Missing credentials export nothing and fail.
	out, err = source("PROJECT_SLUG=other", "", dump)
	if err == nil || !strings.Contains(out, "Nothing exported") {
		t.Errorf("missing credentials: %v\n%s", err, out)
	}
	for k, v := range readDump() {
		if v != "" {
			t.Errorf("%s exported despite missing credentials", k)
		}
	}

	// Executing instead of sourcing is refused.
	if out, err := h.run(nil, skillScript("smaqit.infrastructure-vault-loader", "terraform-env.sh")); err == nil || !strings.Contains(out, "source this file") {
		t.Errorf("direct execution accepted: %v\n%s", err, out)
	}
}

func TestLoadCredentialsEnablesKvAndLoadsThePlatformRepoToken(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-vault-loader", "load-credentials.sh")
	const platToken = "PLAT-PAT-2e2e"
	cmd := "PROVISIONING_MODE=existing-k3s bash " + shellQuote(mustAbs(t, script))
	out, err := h.runInteractiveCommand(cmd, [][2]string{{"platform_repo_token:", platToken}})
	if err != nil {
		t.Fatalf("existing-k3s load failed: %v\n%s", err, out)
	}
	slug := filepath.Base(h.work)
	if got := h.vaultValue("secret/apps/"+slug+"/platform-repo", "token"); got != platToken {
		t.Errorf("platform-repo token = %q, want %q", got, platToken)
	}
	// The app stores no deploy credential: no github PAT, kubeconfig, mesh key or
	// registry credential, and no machine slug is needed.
	for _, path := range []string{"secret/apps/" + slug + "/github", "secret/organizations"} {
		if got := h.vaultValue(path, "token"); got != "" {
			t.Errorf("%s was written: %q", path, got)
		}
	}
	// From the scheme line on: the terminal recorder's banner echoes the script path.
	_, scheme, _ := strings.Cut(out, "==> Scheme:")
	for _, word := range []string{"kubeconfig", "MACHINE_SLUG", "mesh", "registry"} {
		if strings.Contains(strings.ToLower(scheme), strings.ToLower(word)) {
			t.Errorf("existing-k3s load still mentions %q:\n%s", word, out)
		}
	}
	h.requireAbsent("vault argv", h.argv("vault"), platToken)
	h.requireAbsent("terminal output", out, platToken)
	enable, _ := os.ReadFile(filepath.Join(h.state, "secrets-enable.log"))
	if strings.Count(string(enable), "kv-v2") != 1 {
		t.Errorf("kv-v2 enable calls: %q, want exactly one", enable)
	}

	// A second run skips what is already present and doesn't re-enable kv-v2.
	if out, err := h.runInteractiveCommand(cmd, nil); err != nil || strings.Count(out, "SKIP") != 1 {
		t.Errorf("re-run: %v\n%s", err, out)
	}
	enable, _ = os.ReadFile(filepath.Join(h.state, "secrets-enable.log"))
	if strings.Count(string(enable), "kv-v2") != 1 {
		t.Errorf("kv-v2 re-enabled on a second run: %q", enable)
	}
}

func mustAbs(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

// ciEvidenceRun writes a `gh run view --json headSha,status,conclusion,jobs`
// record for a two-job k3s deploy.yml run. deploySteps lists the deploy job's
// named steps and their conclusions, in order; a nil map omits the deploy job.
func ciEvidenceRun(t *testing.T, status, conclusion, headSha string, deploySteps [][2]string) string {
	t.Helper()
	type step struct {
		Name       string `json:"name"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		Number     int    `json:"number"`
	}
	type job struct {
		Name       string `json:"name"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		Steps      []step `json:"steps"`
	}
	implicit := func(names ...string) []step {
		var steps []step
		for i, name := range names {
			steps = append(steps, step{name, "completed", "success", i + 1})
		}
		return steps
	}
	jobs := []job{{"build", "completed", "success",
		implicit("Set up job", "Run actions/checkout@v4", "Log in to GHCR", "Build and push image", "Complete job")}}
	if deploySteps != nil {
		steps := implicit("Set up job", "Run actions/checkout@v4")
		for _, s := range deploySteps {
			steps = append(steps, step{s[0], "completed", s[1], len(steps) + 1})
		}
		steps = append(steps, step{"Complete job", "completed", "success", len(steps) + 1})
		jobs = append(jobs, job{"deploy", status, conclusion, steps})
	}
	out, err := json.Marshal(map[string]any{"headSha": headSha, "status": status, "conclusion": conclusion, "jobs": jobs})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "run.json")
	if err := os.WriteFile(path, out, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// ci-evidence decides an existing-k3s CI deploy from the deploy run's record
// alone: a finished, successful run whose head is the intended commit or
// contains it, with successful "Verify revision" and row-applicable verify
// steps. Every other state fails with its own line, and it touches nothing
// but `gh run view` and the commit comparison.
func TestCIEvidenceDecidesFromTheDeployRun(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skipf("jq not available: %v", err)
	}
	rows := []k3sEnvRow{
		k3sRow("test", "edge-01", "private-mesh", "platform-default-tls"),
		k3sRow("prod", "pub-02", "public", "platform-owned"),
		k3sRow("dev", "pub-01", "public", "own-tls"),
	}
	input := specEnvironmentsJSON(t, rows, k3sConstraints(rows))
	const expected = "0123456789abcdef0123456789abcdef01234567"
	const later = "89abcdef0123456789abcdef0123456789abcdef"
	steps := func(revision, inCluster, https string) [][2]string {
		list := [][2]string{{"Resolve target", "success"}, {"Apply manifests", "success"}, {"Wait for rollout", "success"}}
		if revision != "" {
			list = append(list, [2]string{"Verify revision", revision})
		}
		if inCluster != "" {
			list = append(list, [2]string{"Verify in cluster", inCluster})
		}
		if https != "" {
			list = append(list, [2]string{"Verify over HTTPS", https})
		}
		return append(list, [2]string{"Remove kubeconfig", "success"})
	}
	inCluster := ciEvidenceRun(t, "completed", "success", expected, steps("success", "success", "skipped"))
	external := ciEvidenceRun(t, "completed", "success", expected, steps("success", "skipped", "success"))
	laterRun := ciEvidenceRun(t, "completed", "success", later, steps("success", "success", "skipped"))

	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-deploy-verify", "ci-evidence.sh")
	args := func(env, sha string) []string {
		return []string{input, env, "--run-id", "9876543210", "--expected-sha", sha}
	}
	cases := []struct {
		name    string
		run     string
		compare string
		args    []string
		code    int
		want    string
	}{
		{"in-cluster row", inCluster, "", args("test", expected), 0, "'Verify revision' and 'Verify in cluster' succeeded"},
		{"platform-owned row", inCluster, "", args("prod", expected), 0, "'Verify in cluster' succeeded"},
		{"external row", external, "", args("dev", expected), 0, "'Verify over HTTPS' succeeded"},
		{"short expected SHA", inCluster, "", args("test", expected[:8]), 0, "PASS"},
		{"later head contains the commit", laterRun, "ahead", args("test", expected), 0, "deployed " + later},
		{"later head identical", laterRun, "identical", args("test", expected), 0, "PASS"},
		{"head behind the commit", laterRun, "behind", args("test", expected), 1, "wrong SHA"},
		{"head diverged", laterRun, "diverged", args("test", expected), 1, "compare status 'diverged'"},
		{"compare call fails", laterRun, "", args("test", expected), 1, "could not compare"},
		{"run unfinished", ciEvidenceRun(t, "in_progress", "", expected, steps("success", "", "")), "", args("test", expected), 1, "has not finished"},
		{"run failed", ciEvidenceRun(t, "completed", "failure", expected, steps("success", "failure", "skipped")), "", args("test", expected), 1, "concluded 'failure', not success"},
		{"revision step failed", ciEvidenceRun(t, "completed", "success", expected, steps("failure", "success", "skipped")), "", args("test", expected), 1, "step 'Verify revision' concluded 'failure'"},
		{"revision step missing", ciEvidenceRun(t, "completed", "success", expected, steps("", "success", "skipped")), "", args("test", expected), 1, "step 'Verify revision' is missing"},
		{"applicable step failed", ciEvidenceRun(t, "completed", "success", expected, steps("success", "skipped", "failure")), "", args("dev", expected), 1, "step 'Verify over HTTPS' concluded 'failure'"},
		{"applicable step skipped", external, "", args("test", expected), 1, "step 'Verify in cluster' was skipped"},
		{"applicable step missing", ciEvidenceRun(t, "completed", "success", expected, steps("success", "", "success")), "", args("prod", expected), 1, "step 'Verify in cluster' is missing"},
		{"no deploy job", ciEvidenceRun(t, "completed", "success", expected, nil), "", args("test", expected), 1, "has no 'deploy' job"},
		{"unknown run", "", "", args("test", expected), 1, "could not read run 9876543210"},
		{"unknown environment", inCluster, "", args("staging", expected), 1, "no Environments row for 'staging'"},
		{"no run id", inCluster, "", []string{input, "test", "--expected-sha", expected}, 2, "Usage"},
		{"bad sha", inCluster, "", []string{input, "test", "--run-id", "1", "--expected-sha", "main"}, 2, "Usage"},
		{"bad run id", inCluster, "", []string{input, "test", "--run-id", "latest", "--expected-sha", expected}, 2, "Usage"},
	}
	for _, c := range cases {
		env := []string{"STUB_GH_RUN_JSON=" + c.run, "STUB_GH_COMPARE_STATUS=" + c.compare}
		out, err := h.run(env, script, c.args...)
		if scriptExitCode(err) != c.code || !strings.Contains(out, c.want) {
			t.Errorf("%s: exit %d, want %d mentioning %q\n%s", c.name, scriptExitCode(err), c.code, c.want, out)
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(h.argv("gh")), "\n") {
		if !strings.HasPrefix(line, "run\x00view\x009876543210\x00--json\x00headSha,status,conclusion,jobs\x00") &&
			!strings.HasPrefix(line, "api\x00repos/{owner}/{repo}/compare/"+expected+"..."+later+"\x00") {
			t.Errorf("unexpected gh call: %q", line)
		}
	}
	for _, tool := range []string{"vault", "kubectl", "terraform"} {
		if got := h.argv(tool); got != "" {
			t.Errorf("ci-evidence called %s: %q", tool, got)
		}
	}
}
