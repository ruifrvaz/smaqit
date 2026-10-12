package main

import (
	"io/fs"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const cicdAssets = "skills-shared/smaqit.infrastructure-cicd-generate/assets"

// The generated full-mode deploy.yml must apply exactly the plan plan-guard.sh
// inspected: the guard keeps it with --out, and the apply step consumes that
// file. A bare `terraform apply` would plan again, unguarded.
func TestFullDeployTemplateAppliesTheGuardedPlan(t *testing.T) {
	content, err := skillFilesShared.ReadFile(cicdAssets + "/deploy.yml.full.template")
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		Jobs map[string]struct {
			Steps []struct {
				Name             string `yaml:"name"`
				Run              string `yaml:"run"`
				WorkingDirectory string `yaml:"working-directory"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		t.Fatalf("deploy.yml.full.template is not valid YAML: %v", err)
	}
	provision, ok := workflow.Jobs["provision"]
	if !ok {
		t.Fatal("deploy.yml.full.template has no provision job")
	}

	guard, apply := -1, -1
	for i, step := range provision.Steps {
		switch {
		case strings.Contains(step.Run, "plan-guard.sh"):
			if guard != -1 {
				t.Fatalf("more than one plan-guard step")
			}
			guard = i
			if !strings.Contains(step.Run, "--out deployment/terraform/tfplan") {
				t.Errorf("guard does not keep its plan: %q", step.Run)
			}
		case strings.Contains(step.Run, "terraform apply"):
			if apply != -1 {
				t.Fatalf("more than one terraform apply step")
			}
			apply = i
			fields := strings.Fields(step.Run)
			if fields[len(fields)-1] != "tfplan" || step.WorkingDirectory != "deployment/terraform" {
				t.Errorf("apply does not consume the guarded plan: run %q in %q", step.Run, step.WorkingDirectory)
			}
		}
	}
	if guard == -1 || apply == -1 || guard > apply {
		t.Fatalf("expected the plan guard before the apply, got guard step %d, apply step %d", guard, apply)
	}
}

func TestCicdTemplatesNeverAutoApprove(t *testing.T) {
	err := fs.WalkDir(skillFilesShared, cicdAssets, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := skillFilesShared.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), "-auto-approve") {
			t.Errorf("%s applies with -auto-approve instead of a guarded plan file", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// The k3s deploy.yml targets a machine-slug GitHub Environment, builds the
// image with its revision baked in, stamps the per-environment host into the
// committed Ingress before linting, and verifies /health.json with a bounded
// retry.
func TestK3sDeployTemplateContract(t *testing.T) {
	content, err := skillFilesShared.ReadFile(cicdAssets + "/deploy.yml.k3s.template")
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		On struct {
			WorkflowDispatch struct {
				Inputs map[string]struct {
					Type    string   `yaml:"type"`
					Options []string `yaml:"options"`
					Default string   `yaml:"default"`
				} `yaml:"inputs"`
			} `yaml:"workflow_dispatch"`
		} `yaml:"on"`
		Jobs map[string]struct {
			Environment string `yaml:"environment"`
			Steps       []struct {
				Name string `yaml:"name"`
				Run  string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		t.Fatalf("deploy.yml.k3s.template is not valid YAML: %v", err)
	}

	input := workflow.On.WorkflowDispatch.Inputs["environment"]
	if input.Type != "choice" || input.Default != "__PROD_MACHINE_SLUG__" ||
		strings.Join(input.Options, ",") != "__PROD_MACHINE_SLUG__,__TEST_MACHINE_SLUG__" {
		t.Errorf("environment input = %+v", input)
	}
	deploy := workflow.Jobs["deploy"]
	if deploy.Environment != "${{ inputs.environment || '__PROD_MACHINE_SLUG__' }}" {
		t.Errorf("deploy environment = %q, want the machine-slug expression", deploy.Environment)
	}

	build := ""
	for _, s := range workflow.Jobs["build"].Steps {
		build += s.Run
	}
	if !strings.Contains(build, `docker build --build-arg DEPLOY_SHA="${{ github.sha }}"`) {
		t.Errorf("build does not pass DEPLOY_SHA: %q", build)
	}

	index := map[string]int{}
	for i, s := range deploy.Steps {
		switch {
		case strings.Contains(s.Run, "s/__HOST__/"):
			index["stamp"] = i
		case strings.Contains(s.Run, "manifest-lint.sh"):
			index["lint"] = i
			if !strings.Contains(s.Run, "deployment/k3s/ingress.yaml") {
				t.Errorf("lint skips the committed Ingress: %q", s.Run)
			}
		case strings.Contains(s.Run, "kubectl apply") && strings.Contains(s.Run, "deployment/k3s/ingress.yaml"):
			index["apply"] = i
		case strings.Contains(s.Run, "health.json"):
			index["verify"] = i
			if !strings.Contains(s.Run, "seq 1 12") || !strings.Contains(s.Run, "sleep 5") {
				t.Errorf("verify has no bounded retry: %q", s.Run)
			}
		}
	}
	for _, step := range []string{"stamp", "lint", "apply", "verify"} {
		if _, ok := index[step]; !ok {
			t.Fatalf("deploy job has no %s step", step)
		}
	}
	if !(index["stamp"] < index["lint"] && index["lint"] < index["apply"] && index["apply"] < index["verify"]) {
		t.Errorf("step order = %v, want stamp < lint < apply < verify", index)
	}
	if strings.Contains(string(content), `/health"`) || strings.Contains(string(content), "environment: prod") {
		t.Error("template still uses the old /health path or a fixed prod Environment")
	}
}
