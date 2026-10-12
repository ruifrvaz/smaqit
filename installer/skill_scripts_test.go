package main

import (
	"bytes"
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
if [ "$1" = "api" ]; then
  method=GET; path=""
  shift
  while [ "$#" -gt 0 ]; do
    case "$1" in
      -X) method="$2"; shift 2 ;;
      -i|--silent) shift ;;
      *) path="$1"; shift ;;
    esac
  done
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
	for name, body := range map[string]string{"vault": stubVault, "gh": stubGh, "terraform": stubTerraform} {
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

func TestPlanGuardRejectsBadUsage(t *testing.T) {
	h := newScriptHarness(t)
	for _, args := range [][]string{
		{"--bogus"},
		{"--allow-replace"},
		{"--out"},
		{"extra-dir"},
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
	const kubeconfig = "apiVersion: v1\nkind: Config\nusers: [{name: deployer, user: {token: KUBE-TOKEN-91a2}}]\n"
	const sshKey = "-----BEGIN OPENSSH PRIVATE KEY-----\nAPP-KEY-77c1\n-----END OPENSSH PRIVATE KEY-----\n"
	h.seedVault("secret/apps/shop/edge-01/kubeconfig", map[string]string{"value": kubeconfig})
	h.seedVault("secret/apps/shop/ssh", map[string]string{"private_key": sshKey, "public_key": "ssh-ed25519 AAAA"})
	h.seedVault("secret/apps/shop/platform-repo", map[string]string{"token": "PLATFORM-PAT-3f9e"})
	h.seedVault("secret/legacy-app/ssh", map[string]string{"private_key": sshKey})

	for path, want := range map[string]string{
		"apps/shop/edge-01/kubeconfig": kubeconfig,
		"apps/shop/ssh":                sshKey,
		"apps/shop/platform-repo":      "PLATFORM-PAT-3f9e",
		"legacy-app/ssh":               sshKey,
	} {
		out, err := h.run(nil, script, path)
		if err != nil {
			t.Fatalf("%s: %v\n%s", path, err, out)
		}
		file := strings.TrimSpace(out)
		h.requireAbsent("stdout of "+path, out, "KUBE-TOKEN-91a2", "APP-KEY-77c1", "PLATFORM-PAT-3f9e")
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

	for _, path := range []string{"machines/edge-01/base-ssh", "apps/shop/github", "organizations/acme/github-package-read", "apps/ssh", "apps/Shop/ssh", "apps/shop/edge-01/kubeconfig/extra"} {
		if out, err := h.run(nil, script, path); err == nil {
			t.Errorf("%s: handed out a credential outside the allowlist:\n%s", path, out)
		}
	}

	out, err := h.run(nil, script, "apps/shop/edge-02/kubeconfig")
	if err == nil || !strings.Contains(out, "hand-off") || strings.Contains(out, "/credential") {
		t.Errorf("missing kubeconfig: %v\n%s", err, out)
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

func TestSyncK3sEnvUsesTheMachineSlugEnvironment(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-repo-config", "sync-k3s-env.sh")
	const kubeconfig = "apiVersion: v1\nusers: [{user: {token: KUBE-TOKEN-5e5e}}]\n"
	h.seedVault("secret/apps/shop/edge-01/kubeconfig", map[string]string{"value": kubeconfig})
	h.seedVault("secret/organizations/acme/github-package-read", map[string]string{"username": "bot", "token": "PKG-TOKEN-7a7a"})

	out, err := h.run(nil, script, "acme/shop", "shop", "edge-01", "--host", "shop.example.com", "--registry-org", "acme")
	if err != nil {
		t.Fatalf("sync failed: %v\n%s", err, out)
	}
	for name, want := range map[string]string{"KUBECONFIG": kubeconfig, "REGISTRY_USERNAME": "bot", "REGISTRY_TOKEN": "PKG-TOKEN-7a7a"} {
		if got, ok := h.ghValue("edge-01", "secret", name); !ok || got != want {
			t.Errorf("secret %s = %q (present %v), want %q", name, got, ok, want)
		}
	}
	if got, _ := h.ghValue("edge-01", "variable", "APP_HOST"); got != "shop.example.com" {
		t.Errorf("APP_HOST = %q", got)
	}
	for _, env := range []string{"test", "prod"} {
		if _, err := os.Stat(filepath.Join(h.state, "gh", "envs", env)); err == nil {
			t.Errorf("created a %q Environment instead of the machine slug", env)
		}
	}
	h.requireAbsent("gh argv", h.argv("gh"), "KUBE-TOKEN-5e5e", "PKG-TOKEN-7a7a")
	h.requireAbsent("output", out, "KUBE-TOKEN-5e5e", "PKG-TOKEN-7a7a")

	// Without the platform's hand-off there is nothing to sync, and nothing is written.
	h2 := newScriptHarness(t)
	out, err = h2.run(nil, script, "acme/shop", "shop", "edge-01")
	if err == nil || !strings.Contains(out, "hand-off") {
		t.Fatalf("missing kubeconfig accepted: %v\n%s", err, out)
	}
	if strings.Contains(h2.argv("gh"), "PUT") || strings.Contains(h2.argv("gh"), "secret\x00set") {
		t.Error("wrote to GitHub before the kubeconfig existed")
	}
	if _, err := h2.run(nil, script, "acme/shop", "shop", "Edge_01"); err == nil {
		t.Error("accepted a machine slug that needs URL encoding")
	}
}

func TestLoadCredentialsEnablesKvAndLoadsTheRegistryToken(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-vault-loader", "load-credentials.sh")
	const (
		ghToken   = "GH-PAT-1d1d"
		platToken = "PLAT-PAT-2e2e"
		pkgToken  = "PKG-PAT-3f3f"
	)
	cmd := "PROVISIONING_MODE=existing-k3s CREDENTIAL_SCOPE=access REGISTRY_ORG=acme bash " + shellQuote(mustAbs(t, script))
	out, err := h.runInteractiveCommand(cmd, [][2]string{
		{"github_token:", ghToken},
		{"platform_repo_token:", platToken},
		{"username:", "bot"},
		{"token:", pkgToken},
	})
	if err != nil {
		t.Fatalf("access load failed: %v\n%s", err, out)
	}
	slug := filepath.Base(h.work)
	for path, want := range map[string]string{
		"secret/apps/" + slug + "/github/token":                  ghToken,
		"secret/apps/" + slug + "/platform-repo/token":           platToken,
		"secret/organizations/acme/github-package-read/username": "bot",
		"secret/organizations/acme/github-package-read/token":    pkgToken,
	} {
		if got := h.vaultValue(filepath.Dir(path), filepath.Base(path)); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
	h.requireAbsent("vault argv", h.argv("vault"), ghToken, platToken, pkgToken)
	h.requireAbsent("terminal output", out, ghToken, platToken, pkgToken)
	enable, _ := os.ReadFile(filepath.Join(h.state, "secrets-enable.log"))
	if strings.Count(string(enable), "kv-v2") != 1 {
		t.Errorf("kv-v2 enable calls: %q, want exactly one", enable)
	}

	// A second run skips everything already present and doesn't re-enable kv-v2.
	if out, err := h.runInteractiveCommand(cmd, nil); err != nil || strings.Count(out, "SKIP") != 3 {
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
