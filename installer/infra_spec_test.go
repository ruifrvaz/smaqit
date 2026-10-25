package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// k3sEnvRow is one Environments row, in environmentColumns order.
type k3sEnvRow [10]string

var k3sFixtureRows = []k3sEnvRow{
	{"dev", "pub-01", "dev.example.com", "public", "own-tls", "traefik", "letsencrypt", "/health.json", "—", "—"},
	{"test", "edge-01", "app.mesh.internal", "private-mesh", "platform-default-tls", "traefik", "—", "`/health.json`", "MESH_SETUP_KEY", "edge-01.mesh.internal"},
	{"prod", "pub-02", "—", "public", "platform-owned", "—", "—", "/healthz", "", ""},
}

// k3sConnectConstraints are the Cluster Connect rows every existing-k3s spec declares.
var k3sConnectConstraints = map[string]string{
	"Cluster Connect Action":          "`example-org/platform/.github/actions/k3s-connect@v1.0.0`",
	"Cluster Connect Machine Input":   "machine",
	"Cluster Connect Namespace Input": "namespace",
}

var k3sFixtureConstraints = map[string]string{
	"Mesh Join Action":                "`example-org/mesh-join@v1.4.2`",
	"Mesh Setup-Key Input":            "setup-key",
	"Mesh Wait-For Input":             "wait-for",
	"Probe Client":                    "wget",
	"Cluster Connect Action":          k3sConnectConstraints["Cluster Connect Action"],
	"Cluster Connect Machine Input":   k3sConnectConstraints["Cluster Connect Machine Input"],
	"Cluster Connect Namespace Input": k3sConnectConstraints["Cluster Connect Namespace Input"],
}

// k3sSpec renders an Infrastructure spec with an Environments table and the
// given Constraints rows.
func k3sSpec(rows []k3sEnvRow, constraints map[string]string) string {
	var b strings.Builder
	b.WriteString("---\nid: INF-PLATFORM\nstatus: draft\ncreated: 2026-10-16\n---\n\n# Platform\n\n")
	b.WriteString("## Environments\n\n| " + strings.Join(environmentColumns, " | ") + " |\n|")
	b.WriteString(strings.Repeat("---|", len(environmentColumns)) + "\n")
	for _, row := range rows {
		b.WriteString("| " + strings.Join(row[:], " | ") + " |\n")
	}
	b.WriteString("\n## Constraints\n\n| Constraint | Description | Impact |\n|---|---|---|\n")
	b.WriteString("| Platform Repo | example-org/platform | onboarding PR target |\n")
	for _, name := range []string{"Cluster Connect Action", "Cluster Connect Machine Input", "Cluster Connect Namespace Input",
		"Mesh Join Action", "Mesh Setup-Key Input", "Mesh Wait-For Input", "Probe Client", "Allowed Ingress Modes", "Deployer Exec"} {
		if value, ok := constraints[name]; ok {
			b.WriteString("| " + name + " | " + value + " | used by the k3s fillers |\n")
		}
	}
	b.WriteString("\n## Acceptance Criteria\n\n- [ ] INF-PLATFORM-001: deployed\n")
	return b.String()
}

func writeInfraSpec(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Join(root, "specs", "infrastructure")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func copyConstraints(overrides map[string]string, drop ...string) map[string]string {
	out := map[string]string{}
	for k, v := range k3sFixtureConstraints {
		out[k] = v
	}
	for k, v := range overrides {
		out[k] = v
	}
	for _, k := range drop {
		delete(out, k)
	}
	return out
}

func TestInfraEnvironmentsParseTheTable(t *testing.T) {
	root := t.TempDir()
	writeInfraSpec(t, root, "platform.md", k3sSpec(k3sFixtureRows, k3sFixtureConstraints))
	writeInfraSpec(t, root, "other.md", "---\nid: INF-OTHER\nstatus: draft\n---\n\n# Other\n")
	set, errs := loadInfraEnvironments(root)
	if len(errs) > 0 {
		t.Fatalf("load: %v", errs)
	}
	if errs := validateInfraEnvironments(set, ""); len(errs) > 0 {
		t.Fatalf("validate: %v", errs)
	}
	if set.Spec != "specs/infrastructure/platform.md" || len(set.Environments) != 3 {
		t.Fatalf("set = %+v", set)
	}
	test := set.find("test")
	if test == nil || test.HealthPath != "/health.json" || test.ClusterIssuer != "" || test.MeshWaitForHost != "edge-01.mesh.internal" {
		t.Errorf("test row = %+v (code spans unwrapped, em dash empty)", test)
	}
	if prod := set.find("prod"); prod.Host != "" || prod.IngressClass != "" || !prod.verifiesInCluster() {
		t.Errorf("prod row = %+v", prod)
	}
	if set.Mesh == nil || set.Mesh.Action != "example-org/mesh-join@v1.4.2" || set.ProbeClient != "wget" {
		t.Errorf("constraints = %+v probe=%q", set.Mesh, set.ProbeClient)
	}
	if set.find("dev").verifiesInCluster() {
		t.Error("a public own-tls environment must verify externally")
	}
	if set.DeployerExec != "denied" {
		t.Errorf("absent Deployer Exec row = %q, want denied", set.DeployerExec)
	}
	if c := set.Connect; c == nil || c.Action != "example-org/platform/.github/actions/k3s-connect@v1.0.0" ||
		c.MachineInput != "machine" || c.NamespaceInput != "namespace" {
		t.Errorf("connect = %+v", set.Connect)
	}
	if strings.Join(set.AllowedIngressModes, ",") != "own-tls,platform-default-tls,platform-owned" {
		t.Errorf("absent Allowed Ingress Modes row = %v, want every mode", set.AllowedIngressModes)
	}
	if !set.find("test").joinsMesh() || set.find("dev").joinsMesh() {
		t.Error("only a row naming a mesh secret joins the mesh")
	}

	// A public row may join the mesh too: the platform keeps the API mesh-only
	// even where the app is public. Reachability still picks the verify mode.
	rows := append([]k3sEnvRow(nil), k3sFixtureRows...)
	rows[0][8], rows[0][9] = "MESH_DEPLOY_KEY_PUB_01", "pub-01.mesh.internal"
	writeInfraSpec(t, root, "platform.md", k3sSpec(rows, copyConstraints(map[string]string{
		"Deployer Exec": "allowed",
		"Probe Client":  "python3",
	})))
	set, errs = loadInfraEnvironments(root)
	if errs = append(errs, validateInfraEnvironments(set, "")...); len(errs) > 0 {
		t.Fatalf("public row joining the mesh: %v", errs)
	}
	if dev := set.find("dev"); !dev.joinsMesh() || dev.verifiesInCluster() || set.DeployerExec != "allowed" || set.ProbeClient != "python3" {
		t.Errorf("dev row = %+v deployerExec=%q probe=%q", dev, set.DeployerExec, set.ProbeClient)
	}
}

func TestInfraEnvironmentsRejectIncompleteRows(t *testing.T) {
	mutate := func(env string, column int, value string) []k3sEnvRow {
		rows := append([]k3sEnvRow(nil), k3sFixtureRows...)
		for i := range rows {
			if rows[i][0] == env {
				rows[i][column] = value
			}
		}
		return rows
	}
	cases := []struct {
		name        string
		rows        []k3sEnvRow
		constraints map[string]string
		env         string
		want        string
	}{
		{"missing reachability", mutate("dev", 3, ""), nil, "dev", `SLOTS-SPEC-FIELD-MISSING: environment "dev" Reachability`},
		{"missing ingress mode", mutate("dev", 4, "—"), nil, "dev", `SLOTS-SPEC-FIELD-MISSING: environment "dev" Ingress Mode`},
		{"unknown reachability", mutate("dev", 3, "vpn"), nil, "dev", `SLOTS-SPEC-FIELD-INVALID: environment "dev" Reachability "vpn"`},
		{"unknown ingress mode", mutate("dev", 4, "self"), nil, "dev", `SLOTS-SPEC-FIELD-INVALID: environment "dev" Ingress Mode "self"`},
		{"own-tls without issuer", mutate("dev", 6, ""), nil, "dev", `environment "dev" ClusterIssuer is required when Ingress Mode is own-tls`},
		{"issuer without own-tls", mutate("test", 6, "letsencrypt"), nil, "test", `environment "test" ClusterIssuer must be empty`},
		{"host missing for an owned Ingress", mutate("dev", 2, ""), nil, "dev", `environment "dev" Host is required`},
		{"ingress class on platform-owned", mutate("prod", 5, "traefik"), nil, "prod", `environment "prod" Ingress Class must be empty`},
		{"private-mesh without setup-key secret", mutate("test", 8, ""), nil, "test", `environment "test" Mesh Setup-Key Secret is required`},
		{"lowercase setup-key secret", mutate("test", 8, "mesh_key"), nil, "test", `environment "test" Mesh Setup-Key Secret "mesh_key"`},
		{"wait-for host without a mesh secret", mutate("dev", 9, "x.mesh.internal"), nil, "dev", `environment "dev" Mesh Setup-Key Secret is required when Mesh Wait-For Host is set`},
		{"mesh secret without a wait-for host", mutate("dev", 8, "MESH_KEY"), nil, "dev", `environment "dev" Mesh Wait-For Host is required when the row names a Mesh Setup-Key Secret`},
		{"bad machine slug", mutate("dev", 1, "Pub_01"), nil, "dev", `environment "dev" Machine Slug "Pub_01"`},
		{"relative health path", mutate("dev", 7, "health"), nil, "dev", `environment "dev" Health Path "health"`},
		{"no mesh action", k3sFixtureRows, copyConstraints(nil, "Mesh Join Action"), "test", `Constraints row "Mesh Join Action" is required`},
		{"unpinned mesh action", k3sFixtureRows, copyConstraints(map[string]string{"Mesh Join Action": "example-org/mesh-join"}), "test", `"Mesh Join Action" "example-org/mesh-join" must be a pinned`},
		{"no probe client for in-cluster verify", k3sFixtureRows, copyConstraints(nil, "Probe Client"), "prod", `Constraints row "Probe Client"`},
		{"unknown probe client", k3sFixtureRows, copyConstraints(map[string]string{"Probe Client": "httpie"}), "", `"Probe Client" "httpie" must be one of`},
		{"no cluster connect action", k3sFixtureRows, copyConstraints(nil, "Cluster Connect Action"), "dev", `Constraints row "Cluster Connect Action" is required`},
		{"no cluster connect inputs", k3sFixtureRows, copyConstraints(nil, "Cluster Connect Namespace Input"), "dev", `Constraints row "Cluster Connect Namespace Input" is required`},
		{"unpinned cluster connect action", k3sFixtureRows, copyConstraints(map[string]string{"Cluster Connect Action": "example-org/platform"}), "", `"Cluster Connect Action" "example-org/platform" must be a pinned`},
		{"bad cluster connect input", k3sFixtureRows, copyConstraints(map[string]string{"Cluster Connect Machine Input": "the machine"}), "", `"Cluster Connect Machine Input" "the machine" must be an action input name`},
		{"ingress mode the platform doesn't allow", k3sFixtureRows, copyConstraints(map[string]string{"Allowed Ingress Modes": "platform-owned"}), "dev", `environment "dev" Ingress Mode "own-tls" is not one the platform allows`},
		{"unknown allowed ingress mode", k3sFixtureRows, copyConstraints(map[string]string{"Allowed Ingress Modes": "platform-owned, ingress-nginx"}), "", `"Allowed Ingress Modes" "platform-owned, ingress-nginx" lists "ingress-nginx"`},
		{"unknown deployer exec", k3sFixtureRows, copyConstraints(map[string]string{"Deployer Exec": "maybe"}), "", `"Deployer Exec" "maybe" must be one of denied, allowed`},
		{"no row for the environment", k3sFixtureRows, nil, "staging", `SLOTS-SPEC-MISSING: specs/infrastructure/platform.md "## Environments" table has no row for environment "staging"`},
		{"duplicate environment", append(append([]k3sEnvRow(nil), k3sFixtureRows...), k3sFixtureRows[0]), nil, "", `environment "dev" is declared more than once`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			constraints := c.constraints
			if constraints == nil {
				constraints = k3sFixtureConstraints
			}
			root := t.TempDir()
			writeInfraSpec(t, root, "platform.md", k3sSpec(c.rows, constraints))
			set, errs := loadInfraEnvironments(root)
			if len(errs) == 0 {
				errs = validateInfraEnvironments(set, c.env)
			}
			for _, err := range errs {
				if strings.Contains(err.Error(), c.want) {
					return
				}
			}
			t.Fatalf("want an error containing %q, got %v", c.want, errs)
		})
	}

	// A platform that allows only platform-owned exposure accepts a row that
	// uses it.
	owned := []k3sEnvRow{k3sFixtureRows[2]}
	root0 := t.TempDir()
	writeInfraSpec(t, root0, "platform.md", k3sSpec(owned, copyConstraints(map[string]string{"Allowed Ingress Modes": "`platform-owned`"})))
	set0, errs0 := loadInfraEnvironments(root0)
	if errs0 = append(errs0, validateInfraEnvironments(set0, "")...); len(errs0) > 0 || strings.Join(set0.AllowedIngressModes, ",") != "platform-owned" {
		t.Fatalf("platform-owned only: %v %v", errs0, set0.AllowedIngressModes)
	}

	// A public-only spec needs neither mesh constraints nor a probe client, only
	// the Cluster Connect rows.
	root := t.TempDir()
	writeInfraSpec(t, root, "platform.md", k3sSpec(k3sFixtureRows[:1], k3sConnectConstraints))
	set, errs := loadInfraEnvironments(root)
	if errs = append(errs, validateInfraEnvironments(set, "")...); len(errs) > 0 {
		t.Fatalf("public-only spec: %v", errs)
	}
}

func TestInfraEnvironmentsStructuralErrors(t *testing.T) {
	root := t.TempDir()
	if set, errs := loadInfraEnvironments(root); set != nil || len(errs) > 0 {
		t.Fatalf("no specs dir: %+v %v", set, errs)
	}
	if errs := validateInfraEnvironments(nil, "prod"); len(errs) != 1 || !strings.HasPrefix(errs[0].Error(), "SLOTS-SPEC-MISSING") {
		t.Fatalf("no table: %v", errs)
	}

	missingColumn := strings.Replace(k3sSpec(k3sFixtureRows, k3sFixtureConstraints), "| Ingress Mode |", "| Mode |", 1)
	writeInfraSpec(t, root, "platform.md", missingColumn)
	if _, errs := loadInfraEnvironments(root); len(errs) != 1 || !strings.Contains(errs[0].Error(), `no "Ingress Mode" column`) {
		t.Fatalf("missing column: %v", errs)
	}

	writeInfraSpec(t, root, "platform.md", k3sSpec(k3sFixtureRows, k3sFixtureConstraints))
	writeInfraSpec(t, root, "second.md", k3sSpec(k3sFixtureRows[:1], nil))
	if _, errs := loadInfraEnvironments(root); len(errs) != 1 || !strings.HasPrefix(errs[0].Error(), "SLOTS-SPEC-AMBIGUOUS") {
		t.Fatalf("two tables: %v", errs)
	}
}

// existing-k3s plans are refused until the spec row is complete; the refusal
// is reported together with binding errors, and other modes ignore the spec.
func TestSlotPlanExistingK3sGate(t *testing.T) {
	root := slotsTestProject(t)
	os.RemoveAll(filepath.Join(root, "specs"))
	_, errs := resolveSlotPlan(root, slotPlanOptions{Mode: "existing-k3s", Environment: "prod", Path: "ci-bootstrap"})
	requireSlotError(t, errs, "SLOTS-SPEC-MISSING")
	mustResolve(t, root, slotPlanOptions{Mode: "provision", Environment: "prod", Path: "standard"})

	rows := append([]k3sEnvRow(nil), k3sFixtureRows...)
	rows[2][3], rows[2][4] = "", "" // prod: no reachability, no ingress mode
	writeInfraSpec(t, root, "platform.md", k3sSpec(rows, k3sFixtureConstraints))
	writeSlotsConfig(t, root, "  verify:\n    skill: platform.not-installed\n")
	_, errs = resolveSlotPlan(root, slotPlanOptions{Mode: "existing-k3s", Environment: "prod", Path: "ci-bootstrap"})
	var fieldErrs int
	for _, err := range errs {
		if strings.HasPrefix(err.Error(), "SLOTS-SPEC-FIELD-MISSING") {
			fieldErrs++
		}
	}
	if fieldErrs != 2 {
		t.Fatalf("want both missing fields reported, got %v", errs)
	}
	requireSlotError(t, errs, "SLOTS-SKILL-MISSING")

	// Another environment's row is complete, so its plan resolves.
	os.Remove(filepath.Join(root, slotsConfigPath))
	plan := mustResolve(t, root, slotPlanOptions{Mode: "existing-k3s", Environment: "test", Path: "ci-bootstrap"})
	if plan.Target == nil || plan.Target.MachineSlug != "edge-01" || plan.Target.Reachability != "private-mesh" {
		t.Fatalf("plan target = %+v", plan.Target)
	}
}
