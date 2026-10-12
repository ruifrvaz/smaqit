package main

import (
	"io/fs"
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
