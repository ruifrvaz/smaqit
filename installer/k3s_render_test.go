package main

import (
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Regenerate on purpose with: go test -run TestK3sDeployGoldenFiles -update
var updateGolden = flag.Bool("update", false, "rewrite installer/testdata golden files")

const k3sGoldenDir = "testdata/k3s-deploy"

// k3sRow builds one Environments row for a reachability class and Ingress mode.
func k3sRow(env, machine, reachability, mode string) k3sEnvRow {
	row := k3sEnvRow{env, machine, "—", reachability, mode, "—", "—", "/health.json", "—", "—"}
	if mode != "platform-owned" {
		row[2], row[5] = env+".example.com", "traefik"
	}
	if mode == "own-tls" {
		row[6] = "letsencrypt"
	}
	if reachability == "private-mesh" {
		row[8], row[9] = "MESH_SETUP_KEY", machine+".mesh.internal"
	}
	return row
}

// k3sConstraints returns exactly the Constraints rows the rows depend on.
func k3sConstraints(rows []k3sEnvRow) map[string]string {
	constraints := map[string]string{}
	for _, row := range rows {
		if row[3] == "private-mesh" {
			constraints["Mesh Join Action"] = "example-org/mesh-join@v1.4.2"
			constraints["Mesh Setup-Key Input"] = "setup-key"
			constraints["Mesh Wait-For Input"] = "wait-for"
		}
		if row[3] == "private-mesh" || row[4] == "platform-owned" {
			constraints["Probe Client"] = "wget"
		}
	}
	return constraints
}

// specEnvironmentsJSON is what `smaqit spec environments --json` prints for the rows.
func specEnvironmentsJSON(t *testing.T, rows []k3sEnvRow, constraints map[string]string) string {
	t.Helper()
	root := t.TempDir()
	writeInfraSpec(t, root, "platform.md", k3sSpec(rows, constraints))
	set, errs := loadInfraEnvironments(root)
	if len(errs) == 0 {
		errs = validateInfraEnvironments(set, "")
	}
	if len(errs) > 0 {
		t.Fatalf("fixture spec is invalid: %v", errs)
	}
	out, err := json.MarshalIndent(set, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "environments.json")
	if err := os.WriteFile(path, out, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func renderK3s(t *testing.T, input string, args ...string) (string, string, int) {
	t.Helper()
	for _, tool := range []string{"bash", "jq", "awk"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not available: %v", tool, err)
		}
	}
	script, err := filepath.Abs(skillScript("smaqit.infrastructure-cicd-generate", "render-k3s-deploy.sh"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", append([]string{script, input}, args...)...)
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return stdout.String(), stderr.String(), code
}

var k3sGoldenCases = func() map[string][]k3sEnvRow {
	cases := map[string][]k3sEnvRow{
		"mixed": {
			k3sRow("test", "edge-01", "private-mesh", "platform-default-tls"),
			k3sRow("prod", "pub-01", "public", "own-tls"),
		},
	}
	for _, reachability := range reachabilityClasses {
		for _, mode := range ingressModes {
			cases[reachability+"-"+mode] = []k3sEnvRow{k3sRow("prod", "edge-01", reachability, mode)}
		}
	}
	return cases
}()

// Every reachability × Ingress-mode pair, plus a mixed public/private pair,
// renders exactly its golden workflow.
func TestK3sDeployGoldenFiles(t *testing.T) {
	if len(k3sGoldenCases) != len(reachabilityClasses)*len(ingressModes)+1 {
		t.Fatalf("golden cases = %d, want every combination plus mixed", len(k3sGoldenCases))
	}
	for name, rows := range k3sGoldenCases {
		t.Run(name, func(t *testing.T) {
			out, stderr, code := renderK3s(t, specEnvironmentsJSON(t, rows, k3sConstraints(rows)), "shop")
			if code != 0 {
				t.Fatalf("render exit %d: %s", code, stderr)
			}
			golden := filepath.Join(k3sGoldenDir, name+".deploy.yml")
			if *updateGolden {
				if err := os.MkdirAll(k3sGoldenDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(golden, []byte(out), 0644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("%v (regenerate with -update)", err)
			}
			if out != string(want) {
				t.Errorf("%s differs from the rendered workflow; review the template change and regenerate with -update", golden)
			}
		})
	}
}

type k3sStep struct {
	Name string            `yaml:"name"`
	If   string            `yaml:"if"`
	Uses string            `yaml:"uses"`
	With map[string]string `yaml:"with"`
	Run  string            `yaml:"run"`
}

type k3sWorkflow struct {
	On   map[string]yaml.Node `yaml:"on"`
	Jobs map[string]struct {
		Environment string            `yaml:"environment"`
		Env         map[string]string `yaml:"env"`
		Steps       []k3sStep         `yaml:"steps"`
	} `yaml:"jobs"`
}

func parseK3sWorkflow(t *testing.T, out string) k3sWorkflow {
	t.Helper()
	var workflow k3sWorkflow
	if err := yaml.Unmarshal([]byte(out), &workflow); err != nil {
		t.Fatalf("rendered workflow is not valid YAML: %v\n%s", err, out)
	}
	return workflow
}

// The rendered workflow is dispatch-only, branches per machine at run time,
// joins the mesh only for private-mesh machines, and verifies in the cluster
// whenever the runner cannot reach the app.
func TestK3sDeployRenderedContract(t *testing.T) {
	rows := k3sGoldenCases["mixed"]
	out, stderr, code := renderK3s(t, specEnvironmentsJSON(t, rows, k3sConstraints(rows)), "shop")
	if code != 0 {
		t.Fatalf("render exit %d: %s", code, stderr)
	}
	workflow := parseK3sWorkflow(t, out)
	if _, ok := workflow.On["push"]; ok || len(workflow.On) != 1 {
		t.Errorf("triggers = %v, want workflow_dispatch only (post-merge-deploy.yml dispatches it)", workflow.On)
	}
	var dispatch struct {
		Inputs map[string]struct {
			Options []string `yaml:"options"`
			Default string   `yaml:"default"`
		} `yaml:"inputs"`
	}
	node := workflow.On["workflow_dispatch"]
	if err := node.Decode(&dispatch); err != nil {
		t.Fatal(err)
	}
	input := dispatch.Inputs["environment"]
	if strings.Join(input.Options, ",") != "edge-01,pub-01" || input.Default != "pub-01" {
		t.Errorf("dispatch input = %+v, want every machine slug with prod's as default", input)
	}
	deploy := workflow.Jobs["deploy"]
	if deploy.Environment != "${{ inputs.environment || 'pub-01' }}" {
		t.Errorf("deploy environment = %q", deploy.Environment)
	}
	var targets map[string]map[string]string
	if err := json.Unmarshal([]byte(deploy.Env["K3S_TARGETS"]), &targets); err != nil {
		t.Fatalf("K3S_TARGETS is not JSON: %v\n%s", err, deploy.Env["K3S_TARGETS"])
	}
	if targets["edge-01"]["reachability"] != "private-mesh" || targets["pub-01"]["ingressMode"] != "own-tls" || targets["pub-01"]["clusterIssuer"] != "letsencrypt" {
		t.Errorf("K3S_TARGETS = %v", targets)
	}

	index := map[string]int{}
	for i, step := range deploy.Steps {
		index[step.Name] = i
		switch step.Name {
		case "Join the mesh":
			if step.If != "env.REACHABILITY == 'private-mesh'" || step.Uses != "example-org/mesh-join@v1.4.2" ||
				step.With["setup-key"] != "${{ secrets[env.MESH_SETUP_KEY_SECRET] }}" || step.With["wait-for"] != "${{ env.MESH_WAIT_FOR_HOST }}" {
				t.Errorf("mesh-join step = %+v", step)
			}
		case "Wait for Certificate Ready":
			if step.If != "env.INGRESS_MODE == 'own-tls'" {
				t.Errorf("Certificate wait runs for every mode: %+v", step)
			}
		case "Verify in cluster":
			if step.If != "env.VERIFY == 'in-cluster'" || !strings.Contains(step.Run, "probe-job.yaml") || strings.Contains(step.Run, "condition=complete") {
				t.Errorf("in-cluster verify = %+v", step)
			}
		case "Verify over HTTPS":
			if step.If != "env.VERIFY == 'external'" || !strings.Contains(step.Run, "seq 1 12") || !strings.Contains(step.Run, "${HEALTH_PATH}") {
				t.Errorf("external verify = %+v", step)
			}
		case "Select and stamp manifests":
			for _, want := range []string{"deployment/k3s/pvc.yaml", "ingress-default-tls.yaml", "platform-owned) INGRESS=\"\"", "__CLUSTER_ISSUER__"} {
				if !strings.Contains(step.Run, want) {
					t.Errorf("stamp step lacks %q", want)
				}
			}
		}
	}
	order := []string{"Resolve target", "Join the mesh", "Install kubectl", "Namespace guard", "Select and stamp manifests",
		"Lint manifests", "Apply manifests", "Wait for rollout", "Wait for Certificate Ready", "Verify revision", "Verify in cluster", "Verify over HTTPS"}
	for i, name := range order {
		if _, ok := index[name]; !ok {
			t.Fatalf("deploy job has no %q step", name)
		}
		if i > 0 && index[order[i-1]] >= index[name] {
			t.Errorf("%q must come before %q", order[i-1], name)
		}
	}
	build := ""
	for _, step := range workflow.Jobs["build"].Steps {
		build += step.Run
	}
	if !strings.Contains(build, `docker build --build-arg DEPLOY_SHA="${{ github.sha }}"`) {
		t.Errorf("build does not pass DEPLOY_SHA: %q", build)
	}

	// An all-public render carries no mesh step and no marker comments.
	public := []k3sEnvRow{k3sRow("prod", "pub-01", "public", "own-tls")}
	out, stderr, code = renderK3s(t, specEnvironmentsJSON(t, public, k3sConstraints(public)), "shop")
	if code != 0 {
		t.Fatalf("public render exit %d: %s", code, stderr)
	}
	if strings.Contains(out, "Join the mesh") || strings.Contains(out, "smaqit:mesh-join") || strings.Contains(out, "__MESH_") {
		t.Error("all-public render kept the mesh-join step")
	}
	parseK3sWorkflow(t, out)
}

func TestK3sDeployRendererRefuses(t *testing.T) {
	valid := specEnvironmentsJSON(t, k3sGoldenCases["mixed"], k3sConstraints(k3sGoldenCases["mixed"]))
	twoNoProd := []k3sEnvRow{k3sRow("test", "edge-01", "public", "own-tls"), k3sRow("stage", "edge-02", "public", "own-tls")}
	shared := []k3sEnvRow{k3sRow("test", "edge-01", "public", "own-tls"), k3sRow("prod", "edge-01", "public", "own-tls")}

	noMesh := filepath.Join(t.TempDir(), "no-mesh.json")
	content, _ := os.ReadFile(valid)
	var set map[string]any
	if err := json.Unmarshal(content, &set); err != nil {
		t.Fatal(err)
	}
	set["mesh"] = nil
	stripped, _ := json.Marshal(set)
	if err := os.WriteFile(noMesh, stripped, 0644); err != nil {
		t.Fatal(err)
	}
	garbage := filepath.Join(t.TempDir(), "garbage.json")
	if err := os.WriteFile(garbage, []byte(`{"environments":[]}`), 0644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		input string
		args  []string
		code  int
		want  string
	}{
		{"no prod and no default", specEnvironmentsJSON(t, twoNoProd, k3sConstraints(twoNoProd)), []string{"shop"}, 1, "pass --default-env"},
		{"unknown default", valid, []string{"shop", "--default-env", "staging"}, 1, "'staging' is not an environment"},
		{"shared machine slug", specEnvironmentsJSON(t, shared, k3sConstraints(shared)), []string{"shop"}, 1, "edge-01 back more than one environment"},
		{"private-mesh without mesh access", noMesh, []string{"shop"}, 1, "Mesh Join Action"},
		{"not spec output", garbage, []string{"shop"}, 1, "smaqit spec environments --json"},
		{"bad app slug", valid, []string{"Shop"}, 1, "app slug"},
		{"missing app slug", valid, nil, 2, "Usage"},
	}
	for _, c := range cases {
		out, stderr, code := renderK3s(t, c.input, c.args...)
		if code != c.code || !strings.Contains(stderr, c.want) || out != "" {
			t.Errorf("%s: exit %d stdout %d bytes, stderr %q; want exit %d mentioning %q", c.name, code, len(out), stderr, c.code, c.want)
		}
	}
	if out, _, code := renderK3s(t, specEnvironmentsJSON(t, twoNoProd, k3sConstraints(twoNoProd)), "shop", "--default-env", "stage"); code != 0 || !strings.Contains(out, "default: edge-02") {
		t.Errorf("explicit --default-env: exit %d", code)
	}
}

var stepCondition = regexp.MustCompile(`^env\.([A-Z_]+) == '([a-z-]+)'$`)

// runsUnder evaluates a rendered step's `if:` against the job environment.
func runsUnder(t *testing.T, step k3sStep, env map[string]string) bool {
	t.Helper()
	if step.If == "" {
		return true
	}
	match := stepCondition.FindStringSubmatch(step.If)
	if match == nil {
		t.Fatalf("step %q has an unexpected condition %q", step.Name, step.If)
	}
	return env[match[1]] == match[2]
}

// runStep executes a rendered step's script the way the runner would, with
// the job environment, and returns what it appended to GITHUB_ENV.
func runStep(t *testing.T, dir string, step k3sStep, env map[string]string) map[string]string {
	t.Helper()
	script := strings.NewReplacer(
		"${{ github.sha }}", "0123456789abcdef0123456789abcdef01234567",
		"${{ needs.build.outputs.image }}", "ghcr.io/example-org/shop:0123456789abcdef0123456789abcdef01234567",
	).Replace(step.Run)
	githubEnv := filepath.Join(t.TempDir(), "github_env")
	cmd := exec.Command("bash", "-c", script)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GITHUB_ENV="+githubEnv)
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("step %q failed: %v\n%s", step.Name, err, out)
	}
	exported := map[string]string{}
	data, _ := os.ReadFile(githubEnv)
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok {
			exported[k] = v
		}
	}
	return exported
}

// For every reachability × Ingress mode, the rendered steps select the right
// manifests, stamp them, and run only the steps that apply to that machine.
func TestK3sDeployStepsFollowTheTarget(t *testing.T) {
	for _, reachability := range reachabilityClasses {
		for _, mode := range ingressModes {
			for _, withPVC := range []bool{false, true} {
				name := reachability + "/" + mode
				if withPVC {
					name += "/pvc"
				}
				t.Run(name, func(t *testing.T) {
					rows := []k3sEnvRow{k3sRow("prod", "edge-01", reachability, mode)}
					out, stderr, code := renderK3s(t, specEnvironmentsJSON(t, rows, k3sConstraints(rows)), "shop")
					if code != 0 {
						t.Fatalf("render: %s", stderr)
					}
					deploy := parseK3sWorkflow(t, out).Jobs["deploy"]
					dir := t.TempDir()
					k3s := filepath.Join(dir, "deployment", "k3s")
					os.MkdirAll(k3s, 0755)
					for _, file := range []string{"deployment.yaml", "service.yaml", "ingress.yaml", "ingress-default-tls.yaml", "probe-job.yaml"} {
						body := "image: __IMAGE__\nhost: __HOST__\nclass: __INGRESS_CLASS__\nissuer: __CLUSTER_ISSUER__\npath: __HEALTH_PATH__\n"
						os.WriteFile(filepath.Join(k3s, file), []byte(body), 0644)
					}
					if withPVC {
						os.WriteFile(filepath.Join(k3s, "pvc.yaml"), []byte("kind: PersistentVolumeClaim\n"), 0644)
					}
					env := map[string]string{"MACHINE": "edge-01", "K3S_TARGETS": deploy.Env["K3S_TARGETS"]}
					var ran []string
					for _, step := range deploy.Steps {
						if !runsUnder(t, step, env) {
							continue
						}
						ran = append(ran, step.Name)
						if step.Name == "Resolve target" || step.Name == "Select and stamp manifests" {
							for k, v := range runStep(t, dir, step, env) {
								env[k] = v
							}
						}
					}
					inCluster := reachability == "private-mesh" || mode == "platform-owned"
					expect := map[string]bool{
						"Join the mesh":              reachability == "private-mesh",
						"Wait for Certificate Ready": mode == "own-tls",
						"Verify in cluster":          inCluster,
						"Verify over HTTPS":          !inCluster,
						"Verify revision":            true,
					}
					for step, want := range expect {
						if got := containsString(ran, step); got != want {
							t.Errorf("step %q ran = %v, want %v (ran %v)", step, got, want, ran)
						}
					}
					manifests := strings.Fields(env["MANIFESTS"])
					wantIngress := map[string]string{"own-tls": "deployment/k3s/ingress.yaml", "platform-default-tls": "deployment/k3s/ingress-default-tls.yaml"}[mode]
					for _, file := range []string{"deployment/k3s/ingress.yaml", "deployment/k3s/ingress-default-tls.yaml"} {
						if containsString(manifests, file) != (file == wantIngress) {
							t.Errorf("MANIFESTS = %v, want Ingress %q only", manifests, wantIngress)
						}
					}
					if withPVC != (len(manifests) > 0 && manifests[0] == "deployment/k3s/pvc.yaml") {
						t.Errorf("MANIFESTS = %v, want pvc.yaml first exactly when committed", manifests)
					}
					if containsString(strings.Fields(env["LINT_MANIFESTS"]), "deployment/k3s/probe-job.yaml") != inCluster {
						t.Errorf("LINT_MANIFESTS = %q, want the probe Job linted only for in-cluster verification", env["LINT_MANIFESTS"])
					}
					if wantIngress != "" {
						stamped, _ := os.ReadFile(filepath.Join(dir, wantIngress))
						if !strings.Contains(string(stamped), "host: prod.example.com") || strings.Contains(string(stamped), "__HOST__") ||
							strings.Contains(string(stamped), "__INGRESS_CLASS__") || strings.Contains(string(stamped), "__CLUSTER_ISSUER__") {
							t.Errorf("Ingress not stamped:\n%s", stamped)
						}
					}
					if inCluster {
						probe, _ := os.ReadFile(filepath.Join(k3s, "probe-job.yaml"))
						if !strings.Contains(string(probe), "path: /health.json") || !strings.Contains(string(probe), "image: ghcr.io/example-org/shop:") {
							t.Errorf("probe Job not stamped:\n%s", probe)
						}
					}
				})
			}
		}
	}
}
