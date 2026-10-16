package main

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
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
}

// mesh-check passes a public row without probing, and for a private-mesh row
// requires the mesh name to resolve and the API port to answer.
func TestMeshCheckRequiresTheMeshForPrivateTargets(t *testing.T) {
	if _, err := exec.LookPath("timeout"); err != nil {
		t.Skipf("timeout not available: %v", err)
	}
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	openPort := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	closed, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	closedPort := strconv.Itoa(closed.Addr().(*net.TCPAddr).Port)
	closed.Close()

	reachable := k3sRow("test", "edge-01", "private-mesh", "platform-owned")
	reachable[9] = "localhost"
	unresolvable := k3sRow("stage", "edge-02", "private-mesh", "platform-owned")
	unresolvable[9] = "edge-02.mesh.invalid"
	rows := []k3sEnvRow{reachable, unresolvable, k3sRow("prod", "pub-01", "public", "own-tls")}
	input := specEnvironmentsJSON(t, rows, k3sConstraints(rows))

	h := newScriptHarness(t)
	script := skillScript("smaqit.infrastructure-deploy-k3s-app", "mesh-check.sh")
	cases := []struct {
		args []string
		code int
		want string
	}{
		{[]string{input, "prod"}, 0, "'prod' is public; no mesh access needed"},
		{[]string{input, "test", "--port", openPort}, 0, "answers on port " + openPort},
		{[]string{input, "test", "--port", closedPort}, 1, "does not answer from this machine"},
		{[]string{input, "stage", "--port", openPort}, 1, "does not resolve from this machine"},
		{[]string{input, "staging"}, 1, "no Environments row for 'staging'"},
		{[]string{input}, 2, "Usage"},
		{[]string{input, "test", "--port", "api"}, 2, "Usage"},
	}
	for _, c := range cases {
		out, err := h.run(nil, script, c.args...)
		if scriptExitCode(err) != c.code || !strings.Contains(out, c.want) {
			t.Errorf("%v: exit %d, want %d mentioning %q\n%s", c.args[1:], scriptExitCode(err), c.code, c.want, out)
		}
	}
}

// The probe Job and the platform-default-tls Ingress, rendered as
// deploy-k3s-app Step 5 describes, pass manifest-lint; the default-TLS Ingress
// requests no certificate.
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
