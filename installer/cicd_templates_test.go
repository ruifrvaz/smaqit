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

// No workflow template interpolates a secret or an event value into a run:
// script, where it becomes part of the script text, and none passes a
// registry password on a command line.
func TestWorkflowTemplatesKeepSecretsOutOfScripts(t *testing.T) {
	checked := 0
	err := fs.WalkDir(skillFilesShared, "skills-shared", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.Contains(path, "/assets/") || !strings.HasSuffix(path, ".template") {
			return err
		}
		data, err := fs.ReadFile(skillFilesShared, path)
		if err != nil {
			return err
		}
		for _, banned := range []string{"--docker-password", "--password="} {
			if strings.Contains(string(data), banned) {
				t.Errorf("%s contains %q", path, banned)
			}
		}
		if !strings.Contains(string(data), "\njobs:") {
			return nil
		}
		var workflow struct {
			Jobs map[string]struct {
				Steps []struct {
					Name string `yaml:"name"`
					Run  string `yaml:"run"`
				} `yaml:"steps"`
			} `yaml:"jobs"`
		}
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			t.Errorf("%s is not valid YAML: %v", path, err)
			return nil
		}
		checked++
		for job, spec := range workflow.Jobs {
			for _, step := range spec.Steps {
				for _, banned := range []string{"${{ secrets.", "${{ secrets[", "${{ github.event."} {
					if strings.Contains(step.Run, banned) {
						t.Errorf("%s job %s step %q interpolates %s…}} into its script; pass it through env:", path, job, step.Name, banned)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 5 {
		t.Errorf("checked %d workflow templates, want at least 5", checked)
	}
}
