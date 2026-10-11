package main

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const agentsTemplatePath = "templates/AGENTS.md.template"

// agentsTemplateBlock is the embedded AGENTS.md template as installInstructionsFile
// splices it between existing content: without its trailing newline.
func agentsTemplateBlock(t *testing.T) string {
	t.Helper()
	content, err := agentsMdTemplate.ReadFile(agentsTemplatePath)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimRight(string(content), "\r\n")
}

func installInstructionsAt(t *testing.T, existing *string) (path, status string, err error) {
	t.Helper()
	path = filepath.Join(t.TempDir(), "AGENTS.md")
	if existing != nil {
		if err := os.WriteFile(path, []byte(*existing), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	status, err = installInstructionsFile(agentsMdTemplate, agentsTemplatePath, path)
	return path, status, err
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestInstallInstructionsFileCreatesMissingFile(t *testing.T) {
	path, status, err := installInstructionsAt(t, nil)
	if err != nil || status != "created" {
		t.Fatalf("status = %q, err = %v; want created", status, err)
	}
	if got := readFileString(t, path); got != agentsTemplateBlock(t)+"\n" {
		t.Fatalf("created file = %q, want the template", got)
	}
}

func TestInstallInstructionsFileAppendsToUserFile(t *testing.T) {
	existing := "# My project\n\nUser rules."
	path, status, err := installInstructionsAt(t, &existing)
	if err != nil || status != "appended" {
		t.Fatalf("status = %q, err = %v; want appended", status, err)
	}
	if got, want := readFileString(t, path), existing+"\n\n"+agentsTemplateBlock(t)+"\n"; got != want {
		t.Fatalf("appended file = %q, want %q", got, want)
	}
}

func TestInstallInstructionsFileRefreshesOnlyTheMarkedSection(t *testing.T) {
	before := "# My project\n\nUser rules above.\n\n"
	after := "\n\n## Mine\n\nUser rules below.\n"
	stale := instructionsMarkerBegin + "\n# smaqit\n\nOutdated section.\n" + instructionsMarkerEnd
	existing := before + stale + after
	path, status, err := installInstructionsAt(t, &existing)
	if err != nil || status != "updated" {
		t.Fatalf("status = %q, err = %v; want updated", status, err)
	}
	if got, want := readFileString(t, path), before+agentsTemplateBlock(t)+after; got != want {
		t.Fatalf("refreshed file = %q, want %q", got, want)
	}
}

func TestInstallInstructionsFileLeavesCurrentSectionUnwritten(t *testing.T) {
	existing := "# My project\n\n" + agentsTemplateBlock(t) + "\n\nUser rules below.\n"
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}
	status, err := installInstructionsFile(agentsMdTemplate, agentsTemplatePath, path)
	if err != nil || status != "up to date" {
		t.Fatalf("status = %q, err = %v; want up to date", status, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(past) {
		t.Fatalf("current section was rewritten: mtime %v, want %v", info.ModTime(), past)
	}
	if got := readFileString(t, path); got != existing {
		t.Fatalf("file changed: %q", got)
	}
}

func TestInstallInstructionsFilePreservesCRLFLineEndings(t *testing.T) {
	before := "# My project\r\n\r\n"
	after := "\r\nUser rules below.\r\n"
	existing := before + instructionsMarkerBegin + "\r\nOutdated.\r\n" + instructionsMarkerEnd + after
	path, status, err := installInstructionsAt(t, &existing)
	if err != nil || status != "updated" {
		t.Fatalf("status = %q, err = %v; want updated", status, err)
	}
	want := before + strings.ReplaceAll(agentsTemplateBlock(t), "\n", "\r\n") + after
	if got := readFileString(t, path); got != want {
		t.Fatalf("refreshed CRLF file = %q, want %q", got, want)
	}
}

func TestInstallInstructionsFileLeavesMalformedSectionUntouched(t *testing.T) {
	cases := map[string]string{
		"begin without end": "# Mine\n" + instructionsMarkerBegin + "\nsmaqit\n",
		"end without begin": "# Mine\nsmaqit\n" + instructionsMarkerEnd + "\n",
		"end before begin":  instructionsMarkerEnd + "\n# Mine\n" + instructionsMarkerBegin + "\n",
		"duplicate begin":   instructionsMarkerBegin + "\n" + instructionsMarkerBegin + "\nsmaqit\n" + instructionsMarkerEnd + "\n",
	}
	for name, existing := range cases {
		t.Run(name, func(t *testing.T) {
			path, status, err := installInstructionsAt(t, &existing)
			if !errors.Is(err, errMalformedInstructionsBlock) {
				t.Fatalf("status = %q, err = %v; want errMalformedInstructionsBlock", status, err)
			}
			if got := readFileString(t, path); got != existing {
				t.Fatalf("malformed file was modified: %q", got)
			}
		})
	}
}

func TestResolveGlobalDirHonorsOverrides(t *testing.T) {
	t.Setenv("COPILOT_HOME", "/tmp/copilot")
	t.Setenv("CLAUDE_CONFIG_DIR", "/tmp/claude")
	t.Setenv("CODEX_HOME", "/tmp/codex")
	tests := map[string]string{
		"copilot-agents":  "/tmp/copilot/agents",
		"claude-agents":   "/tmp/claude/agents",
		"claude-commands": "/tmp/claude/commands",
		"claude-skills":   "/tmp/claude/skills",
		"codex-agents":    "/tmp/codex/agents",
	}
	for kind, want := range tests {
		got, err := resolveGlobalDir(kind)
		if err != nil {
			t.Fatalf("resolveGlobalDir(%q): %v", kind, err)
		}
		if got != want {
			t.Errorf("resolveGlobalDir(%q) = %q, want %q", kind, got, want)
		}
	}
	shared, err := resolveGlobalDir("shared-skills")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(shared, filepath.Join(".agents", "skills")) {
		t.Errorf("shared skills path = %q", shared)
	}
}

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

func TestRemoveEmbeddedSkillDirsPreservesUnownedSharedContent(t *testing.T) {
	dstDir := filepath.Join(t.TempDir(), ".agents", "skills")
	if err := copyEmbeddedDir(skillFilesShared, "skills-shared", dstDir); err != nil {
		t.Fatalf("copying embedded shared skills: %v", err)
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

	removed, err := removeEmbeddedSkillDirs(skillFilesShared, "skills-shared", dstDir)
	if err != nil {
		t.Fatalf("removing embedded shared skills: %v", err)
	}
	if removed != 30 {
		t.Fatalf("removed %d shared skills, want 30", removed)
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

// TestSharedSkillsServeCopilotAndCodex is the merge invariant test for task 107: the
// skills-shared embed must serve as the single global skill tree read by both GitHub
// Copilot and Codex. It asserts the expected skill count, that no [SMAQIT_SKILLS_DIR]
// placeholder was left unresolved, and that every resolved self-reference points at the
// shared ${HOME}/.agents/skills path — never a Claude-specific or legacy project path.
func TestSharedSkillsServeCopilotAndCodex(t *testing.T) {
	entries, err := skillFilesShared.ReadDir("skills-shared")
	if err != nil {
		t.Fatalf("reading skills-shared: %v", err)
	}
	topLevelDirs := 0
	for _, entry := range entries {
		if entry.IsDir() {
			topLevelDirs++
		}
	}
	if topLevelDirs != 30 {
		t.Fatalf("skills-shared has %d top-level skill directories, want 30", topLevelDirs)
	}

	err = fs.WalkDir(skillFilesShared, "skills-shared", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		content, err := skillFilesShared.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(content)
		if strings.Contains(text, "[SMAQIT_SKILLS_DIR]") {
			t.Errorf("%s: unresolved [SMAQIT_SKILLS_DIR] placeholder", path)
		}
		if strings.Contains(text, filepath.Join(".github", "skills")) {
			t.Errorf("%s: contains legacy .github/skills path", path)
		}
		if strings.Contains(text, "CLAUDE_CONFIG_DIR") || strings.Contains(text, filepath.Join(".claude", "skills")) {
			t.Errorf("%s: contains Claude-specific path in shared skills tree", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking skills-shared: %v", err)
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
