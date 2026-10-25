package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// rawVault matches a vault CLI invocation a reader could run, and the tool
// grant that would allow one. Skill text reaches Vault only through the
// vault-loader and repo-config scripts, which may call vault themselves.
var rawVault = regexp.MustCompile("(^|[\\s`$(|;&])vault\\s+(kv|login|operator|secrets|server|write|read|token|status|policy|auth|delete|list|unseal)(\\s|`|$)|Bash\\(vault")

func TestSkillTextNeverRunsVaultDirectly(t *testing.T) {
	checked := 0
	err := fs.WalkDir(skillFilesShared, "skills-shared", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" || strings.Contains(path, "/scripts/") {
			return err
		}
		content, err := skillFilesShared.ReadFile(path)
		if err != nil {
			return err
		}
		checked++
		for n, line := range strings.Split(string(content), "\n") {
			if rawVault.MatchString(line) {
				t.Errorf("%s:%d runs vault directly instead of through a script: %s", path, n+1, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatal("no skill documents found in the embedded tree")
	}
}

// A definition carrying skill-body markers is the authoring source for its
// shipped skill, applied by hand: its frontmatter block and body must equal the
// skill's, so a change made in only one of them fails here.
func TestSkillDefinitionsMatchShippedSkills(t *testing.T) {
	definitions, err := filepath.Glob(filepath.Join("..", ".smaqit", "definitions", "skills", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	frontmatter := regexp.MustCompile("(?s)## Frontmatter\n\n```yaml\n(.*?)\n```")
	body := regexp.MustCompile("(?s)<!-- skill-body:begin -->\n(.*?)\n?<!-- skill-body:end -->")
	skill := regexp.MustCompile("(?s)^---\n(.*?)\n---\n\n?(.*)$")
	checked := 0
	for _, path := range definitions {
		def, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		b := body.FindSubmatch(def)
		if b == nil {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		shipped, err := os.ReadFile(filepath.Join("..", "skills", name, "SKILL.md"))
		if err != nil {
			t.Errorf("%s: no shipped skill: %v", name, err)
			continue
		}
		checked++
		s := skill.FindSubmatch(shipped)
		f := frontmatter.FindSubmatch(def)
		if s == nil || f == nil {
			t.Errorf("%s: definition or SKILL.md has no frontmatter block", name)
			continue
		}
		if string(f[1]) != string(s[1]) {
			t.Errorf("%s: definition frontmatter differs from SKILL.md's", name)
		}
		if strings.TrimRight(string(b[1]), "\n") != strings.TrimRight(string(s[2]), "\n") {
			t.Errorf("%s: definition body differs from SKILL.md's; apply the change to both", name)
		}
	}
	if checked == 0 {
		t.Fatal("no definition with skill-body markers found")
	}
}

// The onboarding request opens its PR with the operator's own gh login: the
// skill loads no stored token, so it never needs Vault.
func TestK3sOnboardingLoadsNoStoredCredential(t *testing.T) {
	content, err := skillFilesShared.ReadFile("skills-shared/smaqit.infrastructure-request-k3s-onboarding/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, banned := range []string{"credential-file.sh", "GH_TOKEN", "platform-repo", "load-credentials.sh"} {
		if strings.Contains(string(content), banned) {
			t.Errorf("request-k3s-onboarding mentions %q; it must use the operator's own gh login", banned)
		}
	}
	if !strings.Contains(string(content), ".permissions.push") {
		t.Error("request-k3s-onboarding no longer preflights push access with the operator's gh login")
	}
}
