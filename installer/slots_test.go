package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// slotsTestProject creates an initialized project and an isolated home whose
// shared global skill tree holds the real embedded skills.
func slotsTestProject(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, "claude"))
	shared := filepath.Join(home, ".agents", "skills")
	err := fs.WalkDir(skillFilesShared, "skills-shared", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Base(path) != "SKILL.md" {
			return err
		}
		content, err := skillFilesShared.ReadFile(path)
		if err != nil {
			return err
		}
		dst := filepath.Join(shared, strings.TrimPrefix(path, "skills-shared/"))
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		return os.WriteFile(dst, content, 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".smaqit"), 0755); err != nil {
		t.Fatal(err)
	}
	// existing-k3s plans need a complete Environments row for every environment
	// the tests plan.
	rows := append([]k3sEnvRow(nil), k3sFixtureRows...)
	staging := k3sFixtureRows[0]
	staging[0] = "staging"
	writeInfraSpec(t, root, "platform.md", k3sSpec(append(rows, staging), k3sFixtureConstraints))
	return root
}

func writeTestSkill(t *testing.T, dir, name, frontmatterExtra string) {
	t.Helper()
	content := "---\nname: " + name + "\ndescription: Test filler.\n" + frontmatterExtra + "---\n\n# " + name + "\n"
	path := filepath.Join(dir, name, "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeSlotsConfig(t *testing.T, root, slots string) {
	t.Helper()
	content := "schemaVersion: 1\nphase: deploy\nslots:\n" + slots
	if err := os.WriteFile(filepath.Join(root, slotsConfigPath), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func planSummary(plan *slotPlan) map[string]string {
	summary := map[string]string{}
	for _, entry := range plan.Slots {
		value := entry.Source
		switch entry.Kind {
		case fillerSkill:
			value += " " + entry.Skill + ":" + entry.Operation
		case fillerFamily:
			value += " " + entry.Family + "*:" + entry.Operation
		case fillerBuiltin:
			value += " builtin:" + entry.Builtin
		}
		if entry.Executor != "" {
			value += " @" + entry.Executor
		}
		summary[entry.Slot] = value
	}
	return summary
}

func mustResolve(t *testing.T, root string, opts slotPlanOptions) *slotPlan {
	t.Helper()
	plan, errs := resolveSlotPlan(root, opts)
	if len(errs) > 0 {
		t.Fatalf("resolveSlotPlan(%+v) errors: %v", opts, errs)
	}
	return plan
}

func requireSlotError(t *testing.T, errs []error, code string) {
	t.Helper()
	for _, err := range errs {
		if strings.HasPrefix(err.Error(), code) {
			return
		}
	}
	t.Fatalf("expected an error with code %s, got %v", code, errs)
}

// With no slots.yml, the standard full sequence per mode reproduces the
// current project-new dev-sweep wiring.
func TestSlotPlanDefaultsMatchCurrentBehaviour(t *testing.T) {
	root := slotsTestProject(t)
	vm := map[string]string{
		"platform-prereqs":    "none",
		"access-credentials":  "profile smaqit.infrastructure-vault-loader:access @local",
		"platform-onboarding": "none",
		"artifact-build":      "none",
		"ci-generate":         "profile smaqit.infrastructure-cicd-generate:generate @local",
		"host-bootstrap":      "profile smaqit.infrastructure-vm-bootstrap:bootstrap @local",
		"environment-config":  "none",
		"app-deploy":          "profile smaqit.infrastructure-deploy-rsync*:deploy @local",
		"verify":              "profile smaqit.infrastructure-deploy-verify:verify @local",
		"tls":                 "none",
	}
	owned := map[string]string{
		"target-credentials": "none",
		"iac-generate":       "profile builtin:generate-iac @local",
		"provision":          "profile smaqit.infrastructure-provision-cyso:apply @local",
	}
	shared := map[string]string{
		"target-credentials": "profile smaqit.infrastructure-vault-loader:target @local",
		"iac-generate":       "none",
		"provision":          "none",
	}
	k3s := map[string]string{
		"platform-prereqs":    "profile smaqit.infrastructure-deploy-k3s-app:mesh-check @local",
		"access-credentials":  "profile smaqit.infrastructure-vault-loader:access @local",
		"platform-onboarding": "profile smaqit.infrastructure-request-k3s-onboarding:request @local",
		"target-credentials":  "profile smaqit.infrastructure-vault-loader:target @local",
		"artifact-build":      "profile smaqit.infrastructure-image-build-push*:build @local",
		"iac-generate":        "none",
		"ci-generate":         "profile smaqit.infrastructure-cicd-generate:generate @local",
		"provision":           "none",
		"host-bootstrap":      "none",
		"environment-config":  "none",
		"app-deploy":          "profile smaqit.infrastructure-deploy-k3s-app*:deploy @local",
		"verify":              "profile smaqit.infrastructure-deploy-verify:verify @local",
		"tls":                 "none",
	}
	merge := func(base, extra map[string]string) map[string]string {
		out := map[string]string{}
		for k, v := range base {
			out[k] = v
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	want := map[string]map[string]string{
		"provision":          merge(vm, owned),
		"existing-owned":     merge(vm, owned),
		"existing-shared":    merge(vm, shared),
		"existing-unmanaged": merge(vm, shared),
		"existing-k3s":       k3s,
	}
	for mode, expected := range want {
		plan := mustResolve(t, root, slotPlanOptions{Mode: mode, Environment: "dev", Path: "standard"})
		if plan.Config != "" {
			t.Fatalf("%s: expected no config, got %q", mode, plan.Config)
		}
		if len(plan.Slots) != 13 {
			t.Fatalf("%s: expected 13 slots in a full sequence, got %d", mode, len(plan.Slots))
		}
		got := planSummary(plan)
		for slot, value := range expected {
			if got[slot] != value {
				t.Errorf("%s: slot %s = %q, want %q", mode, slot, got[slot], value)
			}
		}
		if plan.Slots[0].Slot != "platform-prereqs" || plan.Slots[12].Slot != "tls" {
			t.Errorf("%s: unexpected slot order %v", mode, plan.Slots)
		}
	}
}

func TestSlotPlanFamilyCandidates(t *testing.T) {
	root := slotsTestProject(t)
	plan := mustResolve(t, root, slotPlanOptions{Mode: "provision", Environment: "dev", Path: "standard"})
	for _, entry := range plan.Slots {
		if entry.Slot != "app-deploy" {
			continue
		}
		var names []string
		for _, candidate := range entry.Candidates {
			names = append(names, candidate.Skill)
		}
		want := "smaqit.infrastructure-deploy-rsync,smaqit.infrastructure-deploy-rsync-python-nextjs,smaqit.infrastructure-deploy-rsync-python-tornado"
		if strings.Join(names, ",") != want {
			t.Fatalf("app-deploy candidates = %v, want %s", names, want)
		}
		return
	}
	t.Fatal("app-deploy slot missing")
}

func TestSlotPlanDeploymentPaths(t *testing.T) {
	root := slotsTestProject(t)
	pr := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-owned", Environment: "prod", Path: "existing-cicd-pr"}))
	for _, slot := range []string{"iac-generate", "ci-generate", "provision", "host-bootstrap"} {
		if pr[slot] != "refused" {
			t.Errorf("existing-cicd-pr %s = %q, want refused", slot, pr[slot])
		}
	}
	if pr["app-deploy"] != "profile smaqit.infrastructure-deploy-rsync*:deploy @ci" {
		t.Errorf("existing-cicd-pr app-deploy = %q", pr["app-deploy"])
	}
	if pr["environment-config"] != "profile smaqit.infrastructure-repo-config:sync @local" {
		t.Errorf("existing-cicd-pr environment-config = %q", pr["environment-config"])
	}

	push := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "provision", Environment: "prod", Path: "ci-push"}))
	if push["provision"] != "profile smaqit.infrastructure-provision-cyso:apply @ci" {
		t.Errorf("ci-push provision = %q", push["provision"])
	}
	for _, slot := range []string{"iac-generate", "ci-generate", "host-bootstrap", "tls"} {
		if push[slot] != "none" {
			t.Errorf("ci-push %s = %q, want none", slot, push[slot])
		}
	}
	if push["environment-config"] != "profile smaqit.infrastructure-repo-config:sync @local" {
		t.Errorf("ci-push environment-config = %q", push["environment-config"])
	}
	k3sPush := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-k3s", Environment: "prod", Path: "ci-push"}))
	if k3sPush["platform-onboarding"] != "profile smaqit.infrastructure-request-k3s-onboarding:request @local" ||
		k3sPush["artifact-build"] != "profile smaqit.infrastructure-image-build-push*:build @ci" {
		t.Errorf("k3s ci-push onboarding/artifact-build = %q/%q", k3sPush["platform-onboarding"], k3sPush["artifact-build"])
	}

	// existing-k3s on a CI path: the deploy run joins the mesh and verifies on its
	// own, so nothing reaches the target from this machine, whatever the row.
	for _, path := range []string{"existing-cicd-pr", "ci-push"} {
		for _, env := range []string{"test", "prod", "dev"} { // private-mesh, platform-owned, public own-tls
			got := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-k3s", Environment: env, Path: path}))
			if got["platform-prereqs"] != "none" {
				t.Errorf("k3s %s %s platform-prereqs = %q, want none", path, env, got["platform-prereqs"])
			}
			if got["verify"] != "profile smaqit.infrastructure-deploy-verify:ci-evidence @local" {
				t.Errorf("k3s %s %s verify = %q, want ci-evidence", path, env, got["verify"])
			}
		}
		// VM modes keep their local HTTP verify on the CI paths.
		vm := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-owned", Environment: "prod", Path: path}))
		if vm["verify"] != "profile smaqit.infrastructure-deploy-verify:verify @local" || vm["platform-prereqs"] != "none" {
			t.Errorf("existing-owned %s verify/platform-prereqs = %q/%q", path, vm["verify"], vm["platform-prereqs"])
		}
	}
	// A local deploy still checks the mesh and verifies from this machine.
	std := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-k3s", Environment: "test", Path: "standard"}))
	if std["platform-prereqs"] != "profile smaqit.infrastructure-deploy-k3s-app:mesh-check @local" ||
		std["verify"] != "profile smaqit.infrastructure-deploy-verify:verify @local" {
		t.Errorf("k3s standard platform-prereqs/verify = %q/%q", std["platform-prereqs"], std["verify"])
	}
}

func TestSlotPlanSingleSlotRuns(t *testing.T) {
	root := slotsTestProject(t)
	teardown := mustResolve(t, root, slotPlanOptions{Mode: "provision", Environment: "dev", Path: "standard", Slot: "teardown"})
	if len(teardown.Slots) != 1 || teardown.Slots[0].Builtin != "iac-destroy" {
		t.Fatalf("provision teardown = %+v", teardown.Slots)
	}
	owned := mustResolve(t, root, slotPlanOptions{Mode: "existing-owned", Environment: "dev", Path: "standard", Slot: "teardown"})
	if owned.Slots[0].Kind != fillerNone {
		t.Fatalf("existing-owned teardown must default to none, got %+v", owned.Slots[0])
	}
	// A single-slot run ignores the path's full-sequence default-none.
	tls := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-shared", Environment: "prod", Path: "standard", Slot: "tls"}))
	if tls["tls"] != "profile smaqit.infrastructure-domain-tls:issue @local" {
		t.Fatalf("single-slot tls = %q", tls["tls"])
	}
	k3sTLS := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-k3s", Environment: "prod", Path: "standard", Slot: "tls"}))
	if k3sTLS["tls"] != "none" {
		t.Fatalf("k3s single-slot tls = %q, want none", k3sTLS["tls"])
	}
	// A path's per-mode default still applies to a single-slot run, and a
	// single-slot platform-prereqs on a CI path is an explicit request.
	for _, path := range []string{"existing-cicd-pr", "ci-push"} {
		verify := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-k3s", Environment: "prod", Path: path, Slot: "verify"}))
		if verify["verify"] != "profile smaqit.infrastructure-deploy-verify:ci-evidence @local" {
			t.Fatalf("k3s %s single-slot verify = %q, want ci-evidence", path, verify["verify"])
		}
		prereqs := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-k3s", Environment: "test", Path: path, Slot: "platform-prereqs"}))
		if prereqs["platform-prereqs"] != "profile smaqit.infrastructure-deploy-k3s-app:mesh-check @local" {
			t.Fatalf("k3s %s single-slot platform-prereqs = %q, want mesh-check", path, prereqs["platform-prereqs"])
		}
	}
	// Day-two slots never appear in a full sequence.
	full := mustResolve(t, root, slotPlanOptions{Mode: "provision", Environment: "dev", Path: "standard"})
	for _, entry := range full.Slots {
		if entry.Slot == "teardown" || entry.Slot == "replace" {
			t.Fatalf("day-two slot %s in full sequence", entry.Slot)
		}
	}
}

func TestSlotPlanBindings(t *testing.T) {
	root := slotsTestProject(t)
	local := filepath.Join(root, ".agents", "skills")
	writeTestSkill(t, local, "platform.machine-provision", "metadata:\n  smaqit-slots: \"provision=apply, replace=replace\"\n")
	writeTestSkill(t, local, "platform.mesh-verify", "metadata:\n  smaqit-slots: \"verify=check\"\n")
	writeSlotsConfig(t, root, `  provision:
    skill: platform.machine-provision
    environments: [prod]
    params:
      flavor: s5.small
  verify:
    skill: platform.mesh-verify
  tls: none
  ci-generate: default
`)
	prod := mustResolve(t, root, slotPlanOptions{Mode: "provision", Environment: "prod", Path: "standard"})
	got := planSummary(prod)
	if got["provision"] != "bound platform.machine-provision:apply @local" {
		t.Errorf("bound provision = %q", got["provision"])
	}
	if got["verify"] != "bound platform.mesh-verify:check @local" {
		t.Errorf("bound verify = %q", got["verify"])
	}
	if got["ci-generate"] != "profile smaqit.infrastructure-cicd-generate:generate @local" {
		t.Errorf("default binding = %q", got["ci-generate"])
	}
	if prod.Config != slotsConfigPath {
		t.Errorf("config = %q", prod.Config)
	}
	for _, entry := range prod.Slots {
		if entry.Slot == "provision" {
			if entry.Params["flavor"] != "s5.small" || !strings.HasPrefix(entry.SkillPath, local) {
				t.Errorf("provision entry = %+v", entry)
			}
		}
	}
	// The environment filter falls back to the profile default elsewhere.
	dev := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "provision", Environment: "dev", Path: "standard"}))
	if dev["provision"] != "profile smaqit.infrastructure-provision-cyso:apply @local" {
		t.Errorf("dev provision = %q", dev["provision"])
	}
	// An explicit none skips a slot that has a default.
	shared := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-shared", Environment: "prod", Path: "standard", Slot: "tls"}))
	if shared["tls"] != "none" {
		t.Errorf("tls bound none = %q", shared["tls"])
	}
	// A binding still wins over a path's per-mode default (test: the provision
	// binding above covers prod only, which existing-cicd-pr would refuse).
	k3sPR := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-k3s", Environment: "test", Path: "existing-cicd-pr"}))
	if k3sPR["verify"] != "bound platform.mesh-verify:check @local" {
		t.Errorf("bound verify on existing-cicd-pr = %q", k3sPR["verify"])
	}
	if errs := validateSlotBindings(root); len(errs) > 0 {
		t.Fatalf("validateSlotBindings: %v", errs)
	}
}

// deploy-verify declares two verify operations, so a binding to it must name one.
func TestSlotPlanDeployVerifyBindingNeedsOperation(t *testing.T) {
	root := slotsTestProject(t)
	writeSlotsConfig(t, root, "  verify:\n    skill: smaqit.infrastructure-deploy-verify\n")
	_, errs := resolveSlotPlan(root, slotPlanOptions{Mode: "existing-owned", Environment: "prod", Path: "standard"})
	requireSlotError(t, errs, "SLOTS-OPERATION-INVALID")
	writeSlotsConfig(t, root, "  verify:\n    skill: smaqit.infrastructure-deploy-verify\n    operation: ci-evidence\n")
	got := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-k3s", Environment: "prod", Path: "standard"}))
	if got["verify"] != "bound smaqit.infrastructure-deploy-verify:ci-evidence @local" {
		t.Errorf("bound verify with operation = %q", got["verify"])
	}
}

// A platform repository binds the shipped machine operations. Both skills declare
// a second operation for the slot, so the binding must name it; their defaults
// keep resolving to the original operation.
func TestSlotPlanPlatformRepositoryBindings(t *testing.T) {
	root := slotsTestProject(t)
	writeSlotsConfig(t, root, `  target-credentials:
    skill: smaqit.infrastructure-vault-loader
    operation: register
  environment-config:
    skill: smaqit.infrastructure-repo-config
    operation: sync-machine
`)
	got := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "provision", Environment: "prod", Path: "standard"}))
	if got["target-credentials"] != "bound smaqit.infrastructure-vault-loader:register @local" {
		t.Errorf("target-credentials = %q", got["target-credentials"])
	}
	if got["environment-config"] != "bound smaqit.infrastructure-repo-config:sync-machine @local" {
		t.Errorf("environment-config = %q", got["environment-config"])
	}
	if errs := validateSlotBindings(root); len(errs) > 0 {
		t.Fatalf("validateSlotBindings: %v", errs)
	}

	defaults := slotsTestProject(t)
	k3s := planSummary(mustResolve(t, defaults, slotPlanOptions{Mode: "existing-k3s", Environment: "prod", Path: "ci-push"}))
	if k3s["target-credentials"] != "profile smaqit.infrastructure-vault-loader:target @local" {
		t.Errorf("default target-credentials = %q", k3s["target-credentials"])
	}
	if k3s["environment-config"] != "profile smaqit.infrastructure-repo-config:sync @local" {
		t.Errorf("default environment-config = %q", k3s["environment-config"])
	}

	unnamed := slotsTestProject(t)
	writeSlotsConfig(t, unnamed, `  target-credentials:
    skill: smaqit.infrastructure-vault-loader
  environment-config:
    skill: smaqit.infrastructure-repo-config
`)
	_, errs := resolveSlotPlan(unnamed, slotPlanOptions{Mode: "provision", Environment: "prod", Path: "standard"})
	requireSlotError(t, errs, "SLOTS-OPERATION-INVALID")
	if len(errs) != 2 {
		t.Fatalf("expected both unnamed bindings refused, got %d: %v", len(errs), errs)
	}
}

func TestSlotPlanRefusesInvalidBindings(t *testing.T) {
	root := slotsTestProject(t)
	local := filepath.Join(root, ".agents", "skills")
	writeTestSkill(t, local, "platform.undeclared", "metadata:\n  version: \"1.0.0\"\n")
	writeTestSkill(t, local, "platform.multi", "metadata:\n  smaqit-slots: \"app-deploy=deploy, app-deploy=redeploy\"\n")
	writeTestSkill(t, local, "platform.provision", "metadata:\n  smaqit-slots: \"provision=apply\"\n")
	writeSlotsConfig(t, root, `  verify:
    skill: platform.undeclared
  tls:
    skill: platform.not-installed
  app-deploy:
    skill: platform.multi
  provision:
    skill: platform.provision
`)
	_, errs := resolveSlotPlan(root, slotPlanOptions{Mode: "provision", Environment: "prod", Path: "existing-cicd-pr"})
	requireSlotError(t, errs, "SLOTS-SLOT-UNDECLARED")
	requireSlotError(t, errs, "SLOTS-SKILL-MISSING")
	requireSlotError(t, errs, "SLOTS-OPERATION-INVALID")
	requireSlotError(t, errs, "SLOTS-BINDING-REFUSED")
	if len(errs) != 4 {
		t.Fatalf("expected every failure reported together, got %d: %v", len(errs), errs)
	}
	requireSlotError(t, validateSlotBindings(root), "SLOTS-SLOT-UNDECLARED")
}

func TestSlotPlanRefusesLocalSlotInsideCIRun(t *testing.T) {
	root := slotsTestProject(t)
	local := filepath.Join(root, ".agents", "skills")
	writeTestSkill(t, local, "platform.baseline", "metadata:\n  smaqit-slots: \"host-bootstrap=bootstrap\"\n")
	writeSlotsConfig(t, root, "  host-bootstrap:\n    skill: platform.baseline\n")
	// provision (ci) → host-bootstrap (local) → app-deploy (ci) would split the run.
	_, errs := resolveSlotPlan(root, slotPlanOptions{Mode: "provision", Environment: "prod", Path: "ci-push"})
	requireSlotError(t, errs, "SLOTS-BINDING-REFUSED")
	// Without CI provisioning, the local slot precedes every CI slot and is allowed.
	got := planSummary(mustResolve(t, root, slotPlanOptions{Mode: "existing-shared", Environment: "prod", Path: "ci-push"}))
	if got["host-bootstrap"] != "bound platform.baseline:bootstrap @local" {
		t.Fatalf("shared ci-push host-bootstrap = %q", got["host-bootstrap"])
	}
}

// On CI paths every local preparation slot precedes the first CI slot.
func TestSlotPlanLocalSlotsPrecedeCI(t *testing.T) {
	root := slotsTestProject(t)
	for _, mode := range provisioningModes {
		for _, path := range []string{"existing-cicd-pr", "ci-push"} {
			plan := mustResolve(t, root, slotPlanOptions{Mode: mode, Environment: "prod", Path: path})
			seenCI := false
			for _, entry := range plan.Slots {
				if entry.Executor == "ci" {
					seenCI = true
				} else if entry.Executor == "local" && seenCI && entry.Slot != "verify" {
					t.Errorf("%s/%s: local slot %s runs after CI started", mode, path, entry.Slot)
				}
			}
		}
	}
}

func TestParseSlotsConfigRejectsSchemaViolations(t *testing.T) {
	cases := map[string]struct {
		content string
		code    string
	}{
		"invalid yaml":     {"schemaVersion: [1\n", "SLOTS-CONFIG-INVALID"},
		"schema version":   {"schemaVersion: 2\nphase: deploy\n", "SLOTS-CONFIG-INVALID"},
		"phase":            {"schemaVersion: 1\nphase: develop\n", "SLOTS-CONFIG-INVALID"},
		"missing phase":    {"schemaVersion: 1\n", "SLOTS-CONFIG-INVALID"},
		"unknown field":    {"schemaVersion: 1\nphase: deploy\nextra: true\n", "SLOTS-CONFIG-INVALID"},
		"unknown slot":     {"schemaVersion: 1\nphase: deploy\nslots:\n  deploy: none\n", "SLOTS-UNKNOWN-SLOT"},
		"bad scalar":       {"schemaVersion: 1\nphase: deploy\nslots:\n  tls: skip\n", "SLOTS-CONFIG-INVALID"},
		"missing skill":    {"schemaVersion: 1\nphase: deploy\nslots:\n  tls:\n    operation: issue\n", "SLOTS-CONFIG-INVALID"},
		"binding field":    {"schemaVersion: 1\nphase: deploy\nslots:\n  tls:\n    skill: a.b\n    when: always\n", "SLOTS-CONFIG-INVALID"},
		"empty envs":       {"schemaVersion: 1\nphase: deploy\nslots:\n  tls:\n    skill: a.b\n    environments: []\n", "SLOTS-CONFIG-INVALID"},
		"nested params":    {"schemaVersion: 1\nphase: deploy\nslots:\n  tls:\n    skill: a.b\n    params:\n      a: [1]\n", "SLOTS-CONFIG-INVALID"},
		"slots not a map":  {"schemaVersion: 1\nphase: deploy\nslots: [tls]\n", "SLOTS-CONFIG-INVALID"},
		"bad skill name":   {"schemaVersion: 1\nphase: deploy\nslots:\n  tls:\n    skill: Bad Name\n", "SLOTS-CONFIG-INVALID"},
		"bad operation":    {"schemaVersion: 1\nphase: deploy\nslots:\n  tls:\n    skill: a.b\n    operation: Issue\n", "SLOTS-CONFIG-INVALID"},
		"top-level a list": {"- 1\n", "SLOTS-CONFIG-INVALID"},
	}
	for name, tc := range cases {
		_, errs := parseSlotsConfig(slotsConfigPath, []byte(tc.content))
		if len(errs) == 0 {
			t.Errorf("%s: expected %s", name, tc.code)
			continue
		}
		requireSlotError(t, errs, tc.code)
	}
	cfg, errs := parseSlotsConfig(slotsConfigPath, []byte("schemaVersion: 1\nphase: deploy\nslots:\n"))
	if len(errs) > 0 || len(cfg.Bindings) != 0 {
		t.Fatalf("empty slots: cfg=%+v errs=%v", cfg, errs)
	}
}

func TestParseSkillSlots(t *testing.T) {
	slots, err := parseSkillSlots("x", []byte("---\nname: x\ndescription: d\nmetadata:\n  smaqit-slots: \"access-credentials=access, target-credentials=target\"\n---\n"))
	if err != nil || slots["access-credentials"][0] != "access" || slots["target-credentials"][0] != "target" {
		t.Fatalf("slots=%v err=%v", slots, err)
	}
	none, err := parseSkillSlots("x", []byte("---\nname: x\ndescription: d\n---\n"))
	if err != nil || len(none) != 0 {
		t.Fatalf("no declaration: %v %v", none, err)
	}
	crlf, err := parseSkillSlots("x", []byte("---\r\nname: x\r\nmetadata:\r\n  smaqit-slots: \"verify=verify\"\r\n---\r\n"))
	if err != nil || crlf["verify"][0] != "verify" {
		t.Fatalf("crlf: %v %v", crlf, err)
	}
	for name, front := range map[string]string{
		"nested map":   "metadata:\n  smaqit-slots:\n    verify: verify\n",
		"missing op":   "metadata:\n  smaqit-slots: \"verify\"\n",
		"unknown slot": "metadata:\n  smaqit-slots: \"deploy=deploy\"\n",
		"invalid yaml": "description: mode: k3s\n",
	} {
		if _, err := parseSkillSlots("x", []byte("---\nname: x\n"+front+"---\n")); err == nil || !strings.HasPrefix(err.Error(), "SLOTS-SKILL-INVALID") {
			t.Errorf("%s: expected SLOTS-SKILL-INVALID, got %v", name, err)
		}
	}
}

func TestSlotPlanArgumentValidation(t *testing.T) {
	root := slotsTestProject(t)
	_, errs := resolveSlotPlan(root, slotPlanOptions{Mode: "vm", Path: "direct", Slot: "deploy"})
	if len(errs) != 4 {
		t.Fatalf("expected mode, path, env, and slot errors, got %v", errs)
	}
	requireSlotError(t, errs, "SLOTS-ARGUMENT-INVALID")
}

// Every default filler named in the catalogue must declare its slot, so the
// Go table and the shipped skills cannot drift apart.
func TestDefaultFillersDeclareTheirSlots(t *testing.T) {
	read := func(name string) map[string][]string {
		content, err := skillFilesShared.ReadFile("skills-shared/" + name + "/SKILL.md")
		if err != nil {
			t.Fatalf("default filler %s is not shipped: %v", name, err)
		}
		slots, err := parseSkillSlots(name, content)
		if err != nil {
			t.Fatal(err)
		}
		return slots
	}
	entries, err := fs.ReadDir(skillFilesShared, "skills-shared")
	if err != nil {
		t.Fatal(err)
	}
	type defaultFiller struct {
		label  string
		filler slotFiller
	}
	for _, def := range deploySlots {
		var fillers []defaultFiller
		for mode, filler := range def.Defaults {
			fillers = append(fillers, defaultFiller{mode, filler})
		}
		for path, modes := range def.PathDefaults {
			for mode, filler := range modes {
				fillers = append(fillers, defaultFiller{path + "/" + mode, filler})
			}
		}
		for _, f := range fillers {
			mode, filler := f.label, f.filler
			switch filler.Kind {
			case fillerSkill:
				if !containsString(read(filler.Name)[def.Name], filler.Operation) {
					t.Errorf("%s default %s for slot %s does not declare %s=%s", mode, filler.Name, def.Name, def.Name, filler.Operation)
				}
			case fillerFamily:
				declared := 0
				for _, entry := range entries {
					name := entry.Name()
					if name != filler.Name && !strings.HasPrefix(name, filler.Name+"-") {
						continue
					}
					if !containsString(read(name)[def.Name], filler.Operation) {
						t.Errorf("family member %s does not declare %s=%s", name, def.Name, filler.Operation)
					}
					declared++
				}
				if declared == 0 {
					t.Errorf("%s family default %s* for slot %s has no shipped member", mode, filler.Name, def.Name)
				}
			}
		}
	}
}

func TestSlotsTemplateIsValidAndEmpty(t *testing.T) {
	content, err := slotsTemplate.ReadFile("templates/slots.yml.template")
	if err != nil {
		t.Fatal(err)
	}
	cfg, errs := parseSlotsConfig(slotsConfigPath, content)
	if len(errs) > 0 || len(cfg.Bindings) != 0 {
		t.Fatalf("template must parse with no bindings: cfg=%+v errs=%v", cfg, errs)
	}
}

func TestInstallSlotsSkeletonPreservesExistingFile(t *testing.T) {
	root := t.TempDir()
	status, err := installSlotsSkeleton(root)
	if err != nil || status != "created" {
		t.Fatalf("first install: %q %v", status, err)
	}
	path := filepath.Join(root, slotsConfigPath)
	custom := []byte("schemaVersion: 1\nphase: deploy\nslots:\n  tls: none\n")
	if err := os.WriteFile(path, custom, 0644); err != nil {
		t.Fatal(err)
	}
	status, err = installSlotsSkeleton(root)
	if err != nil || status != "preserved" {
		t.Fatalf("re-init: %q %v", status, err)
	}
	if got, _ := os.ReadFile(path); string(got) != string(custom) {
		t.Fatalf("re-init overwrote project file: %q", got)
	}
}

func TestRemoveSmaqitDirPreservingSlots(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".smaqit", "reports"), 0755); err != nil {
		t.Fatal(err)
	}
	custom := []byte("schemaVersion: 1\nphase: deploy\n")
	if err := os.WriteFile(filepath.Join(root, slotsConfigPath), custom, 0644); err != nil {
		t.Fatal(err)
	}
	kept, err := removeSmaqitDirPreservingSlots(root)
	if err != nil || !kept {
		t.Fatalf("kept=%v err=%v", kept, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".smaqit", "reports")); !os.IsNotExist(err) {
		t.Fatal("other .smaqit content must be removed")
	}
	if got, _ := os.ReadFile(filepath.Join(root, slotsConfigPath)); string(got) != string(custom) {
		t.Fatalf("slots.yml not preserved: %q", got)
	}
	empty := t.TempDir()
	if err := os.MkdirAll(filepath.Join(empty, ".smaqit"), 0755); err != nil {
		t.Fatal(err)
	}
	kept, err = removeSmaqitDirPreservingSlots(empty)
	if err != nil || kept {
		t.Fatalf("without slots.yml: kept=%v err=%v", kept, err)
	}
	if _, err := os.Stat(filepath.Join(empty, ".smaqit")); !os.IsNotExist(err) {
		t.Fatal(".smaqit must be removed entirely when no slots.yml exists")
	}
}
