package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The namespace guard fails closed: an unreachable API, a rejected credential,
// or a check with no clear answer is a FAIL, never a PASS.
func TestNamespaceGuardFailsClosed(t *testing.T) {
	h := newScriptHarness(t)
	kubeconfig := filepath.Join(h.work, "kubeconfig")
	if err := os.WriteFile(kubeconfig, []byte("apiVersion: v1\nkind: Config\n"), 0600); err != nil {
		t.Fatal(err)
	}
	script := skillScript("smaqit.infrastructure-deploy-k3s-app", "namespace-guard.sh")
	cases := []struct {
		env  []string
		code int
		want string
	}{
		{nil, 0, "namespace-guard: PASS"},
		{[]string{"STUB_KUBECTL_MODE=unreachable"}, 1, "unreachable or rejected this credential"},
		{[]string{"STUB_KUBECTL_MODE=rejected"}, 1, "Unauthorized"},
		{[]string{"STUB_KUBECTL_MODE=allow-nodes"}, 1, "permits 'get' on cluster-scoped resource 'nodes'"},
		{[]string{"STUB_KUBECTL_MODE=flaky"}, 1, "could not determine whether"},
		{[]string{"STUB_KUBECTL_NAMESPACE=app-other"}, 1, "does not match"},
		{[]string{"STUB_KUBECTL_MODE=exec-yes"}, 1, "permits 'create' on pods/exec"},
		{[]string{"STUB_KUBECTL_MODE=exec-get-yes"}, 1, "permits 'get' on pods/exec"},
		{[]string{"STUB_KUBECTL_MODE=exec-unclear"}, 1, "could not determine whether 'create' on pods/exec"},
	}
	for _, c := range cases {
		out, err := h.run(c.env, script, kubeconfig, "app-shop")
		if scriptExitCode(err) != c.code || !strings.Contains(out, c.want) {
			t.Errorf("%v: exit %d, want %d mentioning %q\n%s", c.env, scriptExitCode(err), c.code, c.want, out)
		}
		if c.code != 0 && strings.Contains(out, "PASS") {
			t.Errorf("%v: a refusal printed PASS\n%s", c.env, out)
		}
	}

	// The exec checks name the subresource with --subresource, never as pods/exec
	// (kubectl would read that as a pod named exec).
	calls, _ := os.ReadFile(filepath.Join(h.state, "kubectl.calls"))
	for _, verb := range []string{"create", "get"} {
		if !strings.Contains(string(calls), "auth can-i "+verb+" pods --subresource=exec -n app-shop") {
			t.Errorf("no %q pods --subresource=exec check:\n%s", verb, calls)
		}
	}
	if strings.Contains(string(calls), "pods/exec") {
		t.Errorf("a can-i call used the pods/exec TYPE/NAME form:\n%s", calls)
	}

	// Deployer Exec: allowed turns a permitted exec into a warning, but an
	// unclear answer still fails closed.
	allowed := []struct {
		env  []string
		code int
		want string
	}{
		{[]string{"STUB_KUBECTL_MODE=exec-yes"}, 0, "WARNING — credential permits 'create' on pods/exec"},
		{[]string{"STUB_KUBECTL_MODE=exec-yes", "GITHUB_ACTIONS=true"}, 0, "::warning::namespace-guard"},
		{[]string{"STUB_KUBECTL_MODE=exec-unclear"}, 1, "could not determine whether"},
		{nil, 0, "no cluster-scoped verb or pods/exec permitted"},
	}
	for _, c := range allowed {
		out, err := h.run(c.env, script, kubeconfig, "app-shop", "--deployer-exec", "allowed")
		if scriptExitCode(err) != c.code || !strings.Contains(out, c.want) {
			t.Errorf("allowed %v: exit %d, want %d mentioning %q\n%s", c.env, scriptExitCode(err), c.code, c.want, out)
		}
	}
	for _, args := range [][]string{{kubeconfig}, {kubeconfig, "app-shop", "--deployer-exec"}, {kubeconfig, "app-shop", "--deployer-exec", "maybe"}, {kubeconfig, "app-shop", "--bogus"}} {
		if out, err := h.run(nil, script, args...); scriptExitCode(err) != 2 {
			t.Errorf("args %q: exit %d, want 2\n%s", args, scriptExitCode(err), out)
		}
	}
}

// exec-refused passes only when a real exec is refused as Forbidden. A
// permitted exec fails under denied and warns under allowed; anything else
// fails closed under both.
func TestExecRefusedRequiresForbidden(t *testing.T) {
	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-deploy-k3s-app", "exec-refused.sh")
	cases := []struct {
		exec   string
		policy string
		code   int
		want   string
	}{
		{"forbidden", "denied", 0, "exec-refused: PASS"},
		{"forbidden", "allowed", 0, "exec-refused: PASS"},
		{"runs", "denied", 1, "could exec into deployment/shop (the command ran)"},
		{"runs", "allowed", 0, "WARNING — the deployer could exec"},
		{"command-fails", "denied", 1, "the exec was accepted"},
		{"command-fails", "allowed", 0, "WARNING — the deployer could exec"},
		{"notfound", "denied", 1, "neither refused as Forbidden nor permitted"},
		{"notfound", "allowed", 1, "neither refused as Forbidden nor permitted"},
		{"hang", "allowed", 1, "timed out"},
	}
	for _, c := range cases {
		env := []string{"STUB_KUBECTL_EXEC=" + c.exec, "EXEC_REFUSED_TIMEOUT=2"}
		out, err := h.run(env, script, "app-shop", "shop", "--deployer-exec", c.policy)
		if scriptExitCode(err) != c.code || !strings.Contains(out, c.want) {
			t.Errorf("%s/%s: exit %d, want %d mentioning %q\n%s", c.exec, c.policy, scriptExitCode(err), c.code, c.want, out)
		}
		if c.code != 0 && strings.Contains(out, "PASS") {
			t.Errorf("%s/%s: a refusal printed PASS\n%s", c.exec, c.policy, out)
		}
	}
	calls, _ := os.ReadFile(filepath.Join(h.state, "kubectl.calls"))
	if !strings.Contains(string(calls), "exec -n app-shop deployment/shop --request-timeout=15s -- true") {
		t.Errorf("unexpected exec call:\n%s", calls)
	}
	for _, args := range [][]string{{"app-shop"}, {"app-shop", "shop", "--deployer-exec", "maybe"}} {
		if out, err := h.run(nil, script, args...); scriptExitCode(err) != 2 {
			t.Errorf("args %q: exit %d, want 2\n%s", args, scriptExitCode(err), out)
		}
	}
}

// The probe Job and the platform-default-tls Ingress, rendered from
// deploy-k3s-app's templates as cicd-generate's k3s mode does, pass
// manifest-lint; the default-TLS Ingress requests no certificate.
func TestK3sManifestTemplatesLintClean(t *testing.T) {
	if err := exec.Command("python3", "-c", "import yaml").Run(); err != nil {
		t.Skipf("python3 with PyYAML not available: %v", err)
	}
	assets := filepath.Join("..", "skills", "smaqit.infrastructure-deploy-k3s-app", "assets")
	render := func(name string, pairs ...string) string {
		content, err := os.ReadFile(filepath.Join(assets, name))
		if err != nil {
			t.Fatal(err)
		}
		out := strings.NewReplacer(pairs...).Replace(string(content))
		path := filepath.Join(t.TempDir(), strings.TrimSuffix(name, ".template"))
		if err := os.WriteFile(path, []byte(out), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	probe := render("probe-job.yaml.template", "__APP_SLUG__", "shop", "__RUN_AS_USER__", "101",
		"__PROBE_FETCH__", "wget -q -O /dev/null", "__IMAGE__", "ghcr.io/example-org/shop:abc", "__HEALTH_PATH__", "/health.json")
	ingress := render("ingress-default-tls.yaml.template", "__APP_SLUG__", "shop", "__INGRESS_CLASS__", "traefik", "__HOST__", "shop.mesh.internal")
	lint := filepath.Join("..", "skills", "smaqit.infrastructure-deploy-k3s-app", "scripts", "manifest-lint.sh")
	if out, err := exec.Command("bash", lint, probe, ingress).CombinedOutput(); err != nil {
		t.Fatalf("manifest-lint rejected the rendered templates: %v\n%s", err, out)
	}
	probeText, _ := os.ReadFile(probe)
	for _, want := range []string{"backoffLimit: 0", "ttlSecondsAfterFinished:", "restartPolicy: Never", "http://shop:80/health.json"} {
		if !strings.Contains(string(probeText), want) {
			t.Errorf("probe Job lacks %q", want)
		}
	}
	ingressText, _ := os.ReadFile(ingress)
	body := string(ingressText)[strings.Index(string(ingressText), "apiVersion:"):]
	if strings.Contains(body, "tls:") || strings.Contains(body, "cert-manager") {
		t.Errorf("platform-default-tls Ingress requests a certificate:\n%s", ingressText)
	}
}

func scriptExitCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	if err != nil {
		return -1
	}
	return 0
}

// A private-mesh machine's setup key loads hidden from the terminal, syncs into
// that machine's Environment under the spec's secret name, and never appears in
// argv or output.
func TestMeshSetupKeyLoadsAndSyncs(t *testing.T) {
	h := newScriptHarness(t)
	slug := filepath.Base(h.work)
	const meshKey = "MESH-KEY-6b6b"
	h.seedVault("secret/apps/"+slug+"/github", map[string]string{"token": "GH"})
	h.seedVault("secret/apps/"+slug+"/platform-repo", map[string]string{"token": "PLAT"})
	h.seedVault("secret/apps/"+slug+"/edge-01/kubeconfig", map[string]string{"value": "apiVersion: v1\n"})
	load := "PROVISIONING_MODE=existing-k3s MACHINE_SLUG=edge-01 MESH_SETUP_KEY=1 bash " +
		shellQuote(mustAbs(t, skillScript("smaqit.infrastructure-vault-loader", "load-credentials.sh")))
	out, err := h.runInteractiveCommand(load, [][2]string{{"mesh_setup_key:", meshKey}})
	if err != nil {
		t.Fatalf("mesh key load failed: %v\n%s", err, out)
	}
	if got := h.vaultValue("secret/apps/"+slug+"/edge-01/mesh-setup-key", "value"); got != meshKey {
		t.Fatalf("mesh-setup-key = %q", got)
	}
	if out, err := h.runInteractiveCommand(load, nil); err != nil || strings.Count(out, "SKIP") != 4 {
		t.Errorf("re-run should skip github, platform-repo, kubeconfig, and the mesh key: %v\n%s", err, out)
	}

	sync := skillScript("smaqit.infrastructure-repo-config", "sync-k3s-env.sh")
	out, err = h.run(nil, sync, "acme/"+slug, slug, "edge-01", "--mesh-secret", "MESH_SETUP_KEY")
	if err != nil {
		t.Fatalf("sync failed: %v\n%s", err, out)
	}
	if got, ok := h.ghValue("edge-01", "secret", "MESH_SETUP_KEY"); !ok || got != meshKey {
		t.Errorf("MESH_SETUP_KEY = %q (present %v)", got, ok)
	}
	h.requireAbsent("gh argv", h.argv("gh"), meshKey)
	h.requireAbsent("vault argv", h.argv("vault"), meshKey)
	h.requireAbsent("output", out, meshKey)

	for _, name := range []string{"mesh_key", "GITHUB_MESH"} {
		if out, err := h.run(nil, sync, "acme/"+slug, slug, "edge-01", "--mesh-secret", name); err == nil {
			t.Errorf("accepted secret name %q\n%s", name, out)
		}
	}
	h2 := newScriptHarness(t)
	h2.seedVault("secret/apps/shop/edge-01/kubeconfig", map[string]string{"value": "apiVersion: v1\n"})
	out, err = h2.run(nil, sync, "acme/shop", "shop", "edge-01", "--mesh-secret", "MESH_SETUP_KEY")
	if err == nil || !strings.Contains(out, "MESH_SETUP_KEY=1") || strings.Contains(h2.argv("gh"), "PUT") {
		t.Errorf("synced without a loaded mesh key: %v\n%s", err, out)
	}
}

// probeFetch is the Probe Client → fetch-command table deploy-k3s-app's
// SKILL.md documents for the probe Job's __PROBE_FETCH__ token.
var probeFetch = map[string]string{
	"curl":    "curl -fsS -o /dev/null",
	"wget":    "wget -q -O /dev/null",
	"python3": "python3 -c 'import sys, urllib.request; urllib.request.urlopen(sys.argv[1], timeout=5)'",
}

// An app image that has python3 but neither curl nor wget (a python:*-slim
// image) verifies in-cluster with Probe Client python3: the probe Job's own
// script answers for a healthy Service and fails for an error status or a
// closed port.
func TestProbeJobPython3Client(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skipf("python3 not available: %v", err)
	}
	skill, err := os.ReadFile(filepath.Join("..", "skills", "smaqit.infrastructure-deploy-k3s-app", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for client, fetch := range probeFetch {
		if !strings.Contains(string(skill), "| `"+client+"` | `"+fetch+"` |") {
			t.Errorf("SKILL.md's fetch-command table has no %s row %q", client, fetch)
		}
	}

	// Only python3 on PATH, as in a slim image without curl or wget.
	bin := t.TempDir()
	if err := os.Symlink(python, filepath.Join(bin, "python3")); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health.json" {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	closed := httptest.NewServer(http.NotFoundHandler())
	closedURL := closed.URL
	closed.Close()

	template, err := os.ReadFile(filepath.Join("..", "skills", "smaqit.infrastructure-deploy-k3s-app", "assets", "probe-job.yaml.template"))
	if err != nil {
		t.Fatal(err)
	}
	rendered := strings.NewReplacer("__APP_SLUG__", "shop", "__RUN_AS_USER__", "10001", "__PROBE_FETCH__", probeFetch["python3"],
		"__IMAGE__", "ghcr.io/example-org/shop:abc", "__HEALTH_PATH__", "/health.json").Replace(string(template))
	var job struct {
		Spec struct {
			Template struct {
				Spec struct {
					Containers []struct {
						Args []string `yaml:"args"`
					} `yaml:"containers"`
				} `yaml:"spec"`
			} `yaml:"template"`
		} `yaml:"spec"`
	}
	if err := yaml.Unmarshal([]byte(rendered), &job); err != nil {
		t.Fatalf("rendered probe Job is not YAML: %v", err)
	}
	if len(job.Spec.Template.Spec.Containers) != 1 || len(job.Spec.Template.Spec.Containers[0].Args) != 1 {
		t.Fatalf("probe Job containers = %+v", job.Spec.Template.Spec.Containers)
	}
	script := strings.ReplaceAll(job.Spec.Template.Spec.Containers[0].Args[0], "http://shop:80", server.URL)
	cmd := exec.Command("/bin/sh", "-c", script)
	cmd.Env = []string{"PATH=" + bin}
	if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), "answered") {
		t.Errorf("probe Job script against a healthy Service: %v\n%s", err, out)
	}

	for name, url := range map[string]string{
		"healthy":      server.URL + "/health.json",
		"error status": server.URL + "/missing",
		"closed port":  closedURL + "/health.json",
	} {
		cmd := exec.Command("/bin/sh", "-c", probeFetch["python3"]+` "$1"`, "probe", url)
		cmd.Env = []string{"PATH=" + bin}
		err := cmd.Run()
		if (err == nil) != (name == "healthy") {
			t.Errorf("%s: python3 fetch exit error %v", name, err)
		}
	}
}
