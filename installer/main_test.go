package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureManagedToolsGitignorePreservesExistingRules(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".gitignore")
	const existing = "node_modules/\r\nuser-owned-rule/\r\n"
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ensureManagedToolsGitignore(root); err != nil {
		t.Fatal(err)
	}
	if err := ensureManagedToolsGitignore(root); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := existing + managedToolsGitignoreRule + "\r\n"
	if string(content) != want {
		t.Fatalf(".gitignore = %q, want %q", content, want)
	}
	if strings.Count(string(content), managedToolsGitignoreRule) != 1 {
		t.Fatalf("managed rule was duplicated: %q", content)
	}
}

func TestManagedToolsGitignoreDoesNotIgnoreDesignArtifacts(t *testing.T) {
	root := t.TempDir()
	if output, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Fatalf("initialize test repository: %v: %s", err, output)
	}
	if err := ensureManagedToolsGitignore(root); err != nil {
		t.Fatal(err)
	}

	isIgnored := func(path string) bool {
		return exec.Command("git", "-C", root, "check-ignore", "-q", path).Run() == nil
	}
	if !isIgnored(".smaqit/tools/plantuml/runtime.js") {
		t.Fatal("managed runtime path is not ignored")
	}
	for _, designPath := range []string{
		"docs/designs/business/dsg-bus-login.md",
		"docs/designs/business/dsg-bus-login.png",
	} {
		if isIgnored(designPath) {
			t.Fatalf("canonical design artifact is ignored: %s", designPath)
		}
	}
}

func TestManagedMCPConfigurationsPreserveUnrelatedContent(t *testing.T) {
	root := t.TempDir()
	writeConfig := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeConfig(filepath.Join(root, ".vscode", "mcp.json"), "{\n  // keep this JSONC comment\n  \"servers\": {\"custom\": {\"command\": \"custom\"}}\n}\n")
	writeConfig(filepath.Join(root, ".mcp.json"), "{\n  \"mcpServers\": {\"custom\": {\"command\": \"custom\"}}\n}\n")
	writeConfig(filepath.Join(root, ".codex", "config.toml"), "custom_config = true\n\n[mcp_servers.custom]\ncommand = \"custom\"\n")

	if err := preflightDesignMCPConfigs(root); err != nil {
		t.Fatal(err)
	}
	if err := installDesignMCPConfigs(root); err != nil {
		t.Fatal(err)
	}
	if err := validateDesignMCPConfigs(root); err != nil {
		t.Fatal(err)
	}
	paths := []string{
		filepath.Join(root, ".vscode", "mcp.json"),
		filepath.Join(root, ".mcp.json"),
		filepath.Join(root, ".codex", "config.toml"),
	}
	first := make([][]byte, len(paths))
	for i, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		first[i] = content
		if !strings.Contains(string(content), "custom") || !strings.Contains(string(content), designMCPServerName) {
			t.Fatalf("%s did not retain custom configuration and add the managed server: %s", path, content)
		}
	}
	if err := installDesignMCPConfigs(root); err != nil {
		t.Fatal(err)
	}
	for i, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != string(first[i]) {
			t.Fatalf("reinstallation changed %s:\n%s", path, content)
		}
	}

	for _, remove := range []func(string) error{removeVSCodeMCPConfig, removeClaudeMCPConfig, removeCodexMCPConfig} {
		if err := remove(root); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "custom") || strings.Contains(string(content), designMCPServerName) {
			t.Fatalf("%s did not preserve only unrelated configuration: %s", path, content)
		}
	}
}

func TestPreflightDesignMCPConfigurationsRejectsConflicts(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
	}{
		{"VS Code", ".vscode/mcp.json", `{"servers":{"smaqit-plantuml":{"command":"other"}}}`},
		{"Claude Code", ".mcp.json", `{"mcpServers":{"smaqit-plantuml":{"command":"other"}}}`},
		{"Codex", ".codex/config.toml", "[mcp_servers.smaqit-plantuml]\ncommand = \"other\"\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, test.path)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(test.body), 0o644); err != nil {
				t.Fatal(err)
			}
			err := preflightDesignMCPConfigs(root)
			if err == nil || !strings.Contains(err.Error(), "already owned") {
				t.Fatalf("expected configuration conflict, got %v", err)
			}
		})
	}
}

func TestPreflightDesignMCPConfigurationsRejectsMalformedConfig(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
	}{
		{"VS Code", ".vscode/mcp.json", `{"servers":`},
		{"Claude Code", ".mcp.json", `{"mcpServers":`},
		{"Codex", ".codex/config.toml", "[mcp_servers\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, test.path)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(test.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := preflightDesignMCPConfigs(root); err == nil {
				t.Fatal("expected malformed configuration to fail preflight")
			}
		})
	}
}

func TestRemoveEmbeddedFilesPreservesUnownedCodexAgents(t *testing.T) {
	dstDir := filepath.Join(t.TempDir(), ".codex", "agents")
	if err := copyEmbeddedDir(codexAgentFiles, "agents-codex", dstDir); err != nil {
		t.Fatalf("copying embedded Codex agents: %v", err)
	}

	customPath := filepath.Join(dstDir, "custom-agent.toml")
	customContent := []byte("name = \"custom-agent\"\n")
	if err := os.WriteFile(customPath, customContent, 0644); err != nil {
		t.Fatalf("writing custom agent: %v", err)
	}

	removed, err := removeEmbeddedFiles(codexAgentFiles, "agents-codex", dstDir)
	if err != nil {
		t.Fatalf("removing embedded Codex agents: %v", err)
	}
	if removed != 9 {
		t.Fatalf("removed %d Codex agents, want 9", removed)
	}

	got, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("custom agent was not preserved: %v", err)
	}
	if string(got) != string(customContent) {
		t.Fatalf("custom agent changed: got %q, want %q", got, customContent)
	}
}

func TestRemoveEmbeddedSkillDirsPreservesUnownedCodexContent(t *testing.T) {
	dstDir := filepath.Join(t.TempDir(), ".agents", "skills")
	if err := copyEmbeddedDir(skillFilesCodex, "skills-codex", dstDir); err != nil {
		t.Fatalf("copying embedded Codex skills: %v", err)
	}

	customPath := filepath.Join(dstDir, "custom-skill", "SKILL.md")
	customContent := []byte("---\nname: custom-skill\ndescription: Custom.\n---\n")
	if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
		t.Fatalf("creating custom skill directory: %v", err)
	}
	if err := os.WriteFile(customPath, customContent, 0644); err != nil {
		t.Fatalf("writing custom skill: %v", err)
	}
	nestedCustomPath := filepath.Join(dstDir, "smaqit.input-business", "custom-note.txt")
	nestedCustomContent := []byte("keep this neighboring file\n")
	if err := os.WriteFile(nestedCustomPath, nestedCustomContent, 0644); err != nil {
		t.Fatalf("writing custom content inside owned skill directory: %v", err)
	}

	removed, err := removeEmbeddedSkillDirs(skillFilesCodex, "skills-codex", dstDir)
	if err != nil {
		t.Fatalf("removing embedded Codex skills: %v", err)
	}
	if removed != 26 {
		t.Fatalf("removed %d Codex skills, want 26", removed)
	}

	got, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("custom skill was not preserved: %v", err)
	}
	if string(got) != string(customContent) {
		t.Fatalf("custom skill changed: got %q, want %q", got, customContent)
	}
	nestedGot, err := os.ReadFile(nestedCustomPath)
	if err != nil {
		t.Fatalf("custom content inside owned skill directory was not preserved: %v", err)
	}
	if string(nestedGot) != string(nestedCustomContent) {
		t.Fatalf("nested custom content changed: got %q, want %q", nestedGot, nestedCustomContent)
	}
}

func TestRemoveDirIfEmpty(t *testing.T) {
	emptyDir := filepath.Join(t.TempDir(), "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatalf("creating empty directory: %v", err)
	}

	removed, err := removeDirIfEmpty(emptyDir)
	if err != nil {
		t.Fatalf("removing empty directory: %v", err)
	}
	if !removed {
		t.Fatal("empty directory was not removed")
	}
	if _, err := os.Stat(emptyDir); !os.IsNotExist(err) {
		t.Fatalf("empty directory still exists: %v", err)
	}
}
