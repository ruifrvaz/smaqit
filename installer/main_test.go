package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
