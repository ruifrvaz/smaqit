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
// or a check with no clear answer is a FAIL, never a PASS. A connection failure
// is retried for a bounded window (a mesh can resolve the API's name before its
// port answers); a credential the API rejects is not.
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
		{[]string{"STUB_KUBECTL_MODE=unreachable"}, 1, "unreachable after 3 attempt(s)"},
		{[]string{"STUB_KUBECTL_MODE=unreachable-twice"}, 0, "API not answering yet (attempt 2/3)"},
		{[]string{"STUB_KUBECTL_MODE=rejected"}, 1, "did not accept this job's identity"},
		{[]string{"STUB_KUBECTL_MODE=forbidden"}, 1, "not bound to the Namespace yet"},
		{[]string{"STUB_KUBECTL_MODE=allow-nodes"}, 1, "permits 'get' on cluster-scoped resource 'nodes'"},
		{[]string{"STUB_KUBECTL_MODE=flaky"}, 1, "could not determine whether"},
		{[]string{"STUB_KUBECTL_NAMESPACE=app-other"}, 1, "does not match"},
		{[]string{"STUB_KUBECTL_MODE=exec-yes"}, 1, "permits 'create' on pods/exec"},
		{[]string{"STUB_KUBECTL_MODE=exec-get-yes"}, 1, "permits 'get' on pods/exec"},
		{[]string{"STUB_KUBECTL_MODE=exec-unclear"}, 1, "could not determine whether 'create' on pods/exec"},
	}
	retry := []string{"NAMESPACE_GUARD_REACH_ATTEMPTS=3", "NAMESPACE_GUARD_REACH_WAIT=0"}
	for _, c := range cases {
		os.Remove(filepath.Join(h.state, "kubectl.reach"))
		before, _ := os.ReadFile(filepath.Join(h.state, "kubectl.calls"))
		out, err := h.run(append(append([]string{}, retry...), c.env...), script, kubeconfig, "app-shop")
		if scriptExitCode(err) != c.code || !strings.Contains(out, c.want) {
			t.Errorf("%v: exit %d, want %d mentioning %q\n%s", c.env, scriptExitCode(err), c.code, c.want, out)
		}
		if c.code != 0 && strings.Contains(out, "PASS") {
			t.Errorf("%v: a refusal printed PASS\n%s", c.env, out)
		}
		after, _ := os.ReadFile(filepath.Join(h.state, "kubectl.calls"))
		lists := strings.Count(string(after[len(before):]), "auth can-i --list")
		want := map[string]int{"STUB_KUBECTL_MODE=unreachable": 3, "STUB_KUBECTL_MODE=unreachable-twice": 3,
			"STUB_KUBECTL_MODE=rejected": 1, "STUB_KUBECTL_MODE=forbidden": 1}[strings.Join(c.env, " ")]
		if want != 0 && lists != want {
			t.Errorf("%v: %d reachability attempts, want %d", c.env, lists, want)
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

// manifest-lint refuses a Service that isn't ClusterIP: exposure goes through
// the Ingress, which a platform-owned exposure model refuses to bypass.
func TestManifestLintRequiresClusterIPServices(t *testing.T) {
	if err := exec.Command("python3", "-c", "import yaml").Run(); err != nil {
		t.Skipf("python3 with PyYAML not available: %v", err)
	}
	lint := filepath.Join("..", "skills", "smaqit.infrastructure-deploy-k3s-app", "scripts", "manifest-lint.sh")
	for serviceType, ok := range map[string]bool{"": true, "ClusterIP": true, "NodePort": false, "LoadBalancer": false} {
		spec := "  selector:\n    app: shop\n  ports:\n    - port: 80\n"
		if serviceType != "" {
			spec = "  type: " + serviceType + "\n" + spec
		}
		path := filepath.Join(t.TempDir(), "service.yaml")
		if err := os.WriteFile(path, []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: shop\nspec:\n"+spec), 0644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command("bash", lint, path).CombinedOutput()
		if (err == nil) != ok {
			t.Errorf("type %q: passed=%v, want %v\n%s", serviceType, err == nil, ok, out)
		}
		if !ok && !strings.Contains(string(out), "Service type '"+serviceType+"' is not permitted") {
			t.Errorf("type %q: unexpected output\n%s", serviceType, out)
		}
	}
}

// Runtime secrets: with RUNTIME_SECRET_NAME set, every name this machine
// receives is referenced by its own secretKeyRef, a reference to another
// environment's secret is optional, and envFrom never pulls the Secret in.
func TestManifestLintRequiresRuntimeSecretRefs(t *testing.T) {
	if err := exec.Command("python3", "-c", "import yaml").Run(); err != nil {
		t.Skipf("python3 with PyYAML not available: %v", err)
	}
	lint := filepath.Join("..", "skills", "smaqit.infrastructure-deploy-k3s-app", "scripts", "manifest-lint.sh")
	deployment := func(env string) string {
		return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: shop
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 101
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: shop
          image: ghcr.io/example-org/shop:sha
` + env + `          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: [ALL]
          resources:
            requests: {cpu: 50m, memory: 32Mi}
            limits: {cpu: 100m, memory: 64Mi}
`
	}
	ref := func(env, key, optional string) string {
		s := "            - name: " + env + "\n              valueFrom:\n                secretKeyRef:\n                  name: shop-runtime\n                  key: " + key + "\n"
		if optional != "" {
			s += "                  optional: " + optional + "\n"
		}
		return s
	}
	session := ref("SESSION_SECRET", "SESSION_SECRET", "")
	adminOptional := ref("ADMIN_PASSWORD", "ADMIN_PASSWORD", "true")
	cases := []struct {
		name, env, secrets, want string
	}{
		{"every name referenced", "          env:\n" + session + adminOptional, "SESSION_SECRET", ""},
		{"prod receives both", "          env:\n" + session + adminOptional, "SESSION_SECRET ADMIN_PASSWORD", ""},
		{"name not referenced", "          env:\n" + session, "SESSION_SECRET ADMIN_PASSWORD", "runtime secret 'ADMIN_PASSWORD' is declared for this environment but no Deployment container references it"},
		{"other environment's secret not optional", "          env:\n" + session + ref("ADMIN_PASSWORD", "ADMIN_PASSWORD", ""), "SESSION_SECRET", "references runtime secret 'ADMIN_PASSWORD', which this environment does not receive"},
		{"env var renamed", "          env:\n" + ref("SESSION", "SESSION_SECRET", ""), "SESSION_SECRET", "env 'SESSION' reads runtime secret key 'SESSION_SECRET'"},
		{"envFrom", "          envFrom:\n            - secretRef:\n                name: shop-runtime\n", "", "pulls the shop-runtime Secret in with envFrom"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "deployment.yaml")
			if err := os.WriteFile(path, []byte(deployment(c.env)), 0644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", lint, path)
			cmd.Env = append(os.Environ(), "RUNTIME_SECRET_NAME=shop-runtime", "RUNTIME_SECRETS="+c.secrets)
			out, err := cmd.CombinedOutput()
			if c.want == "" {
				if err != nil {
					t.Fatalf("want pass, got %v\n%s", err, out)
				}
				return
			}
			if err == nil || !strings.Contains(string(out), c.want) {
				t.Fatalf("want failure containing %q, got %v\n%s", c.want, err, out)
			}
		})
	}

	// Without RUNTIME_SECRET_NAME (a spec with no runtime secrets) the rule is off.
	path := filepath.Join(t.TempDir(), "deployment.yaml")
	if err := os.WriteFile(path, []byte(deployment("")), 0644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("bash", lint, path).CombinedOutput(); err != nil {
		t.Fatalf("no runtime secrets: %v\n%s", err, out)
	}
}
