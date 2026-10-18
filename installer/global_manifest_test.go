package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func skillFile(name string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte("---\nname: " + name + "\ndescription: test\n---\n\n# " + name + "\n")}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func manifestFiles(t *testing.T, dir string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, globalManifestName))
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}
	var m globalManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m.SchemaVersion != 1 || m.Version != Version {
		t.Errorf("manifest header = %+v", m)
	}
	return m.Files
}

// A later release that drops a skill removes exactly the files the earlier
// release installed. User-added files and skills smaqit never installed (such
// as smaqit-extensions') survive.
func TestInstallGlobalTreePrunesWhatTheNextReleaseDrops(t *testing.T) {
	dst := t.TempDir()
	releaseA := fstest.MapFS{
		"skills/smaqit.keep/SKILL.md":           skillFile("smaqit.keep"),
		"skills/smaqit.drop/SKILL.md":           skillFile("smaqit.drop"),
		"skills/smaqit.drop/references/a.md":    {Data: []byte("a")},
		"skills/smaqit.drop/scripts/run.sh":     {Data: []byte("#!/bin/sh\n")},
		"skills/smaqit.dropped-whole/SKILL.md":  skillFile("smaqit.dropped-whole"),
		"skills/smaqit.dropped-whole/extra.txt": {Data: []byte("x")},
	}
	if removed, err := installGlobalTree(releaseA, "skills", dst, true); err != nil || removed != 0 {
		t.Fatalf("release A: removed %d, %v", removed, err)
	}
	if got := strings.Join(manifestFiles(t, dst), ","); got != "smaqit.drop/SKILL.md,smaqit.drop/references/a.md,smaqit.drop/scripts/run.sh,smaqit.dropped-whole/SKILL.md,smaqit.dropped-whole/extra.txt,smaqit.keep/SKILL.md" {
		t.Errorf("release A manifest = %s", got)
	}

	writeTestFile(t, filepath.Join(dst, "smaqit.drop", "my-notes.md"), "user file")
	writeTestFile(t, filepath.Join(dst, "smaqit.task-start", "SKILL.md"), "---\nname: smaqit.task-start\n---\n")

	releaseB := fstest.MapFS{
		"skills/smaqit.keep/SKILL.md":         skillFile("smaqit.keep"),
		"skills/smaqit.drop/SKILL.md":         skillFile("smaqit.drop"),
		"skills/smaqit.added/SKILL.md":        skillFile("smaqit.added"),
		"skills/smaqit.added/references/b.md": {Data: []byte("b")},
	}
	removed, err := installGlobalTree(releaseB, "skills", dst, true)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 4 {
		t.Errorf("release B removed %d files, want 4", removed)
	}
	for _, gone := range []string{"smaqit.drop/references", "smaqit.drop/scripts", "smaqit.dropped-whole"} {
		if exists(filepath.Join(dst, gone)) {
			t.Errorf("%s survived though release B no longer ships it", gone)
		}
	}
	for _, kept := range []string{"smaqit.keep/SKILL.md", "smaqit.drop/SKILL.md", "smaqit.drop/my-notes.md", "smaqit.added/references/b.md", "smaqit.task-start/SKILL.md"} {
		if !exists(filepath.Join(dst, kept)) {
			t.Errorf("%s was removed", kept)
		}
	}
	if got := strings.Join(manifestFiles(t, dst), ","); got != "smaqit.added/SKILL.md,smaqit.added/references/b.md,smaqit.drop/SKILL.md,smaqit.keep/SKILL.md" {
		t.Errorf("release B manifest = %s", got)
	}
}

// Skills installed before manifests existed are removed by name, and only when
// their SKILL.md still carries that name and this release doesn't ship it.
func TestInstallGlobalTreeRemovesRetiredSkillsByName(t *testing.T) {
	release := fstest.MapFS{"skills/smaqit.new-project/SKILL.md": skillFile("smaqit.new-project")}

	dst := t.TempDir()
	writeTestFile(t, filepath.Join(dst, "smaqit.new-greenfield-project", "SKILL.md"), string(skillFile("smaqit.new-greenfield-project").Data))
	writeTestFile(t, filepath.Join(dst, "smaqit.new-greenfield-project", "references", "diagrams.md"), "old")
	writeTestFile(t, filepath.Join(dst, "smaqit.infrastructure-onboard-k3s-app", "SKILL.md"), "---\nname: my-own-onboarding\n---\n")
	removed, err := installGlobalTree(release, "skills", dst, true)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 || exists(filepath.Join(dst, "smaqit.new-greenfield-project")) {
		t.Errorf("retired skill not removed (removed %d)", removed)
	}
	if !exists(filepath.Join(dst, "smaqit.infrastructure-onboard-k3s-app", "SKILL.md")) {
		t.Error("a directory whose SKILL.md names a different skill was removed")
	}
	if !exists(filepath.Join(dst, "smaqit.new-project", "SKILL.md")) {
		t.Error("the shipped skill was not installed")
	}

	// A non-skill tree (agents, commands) never applies the skill seed.
	agents := t.TempDir()
	writeTestFile(t, filepath.Join(agents, "smaqit.new-greenfield-project", "SKILL.md"), string(skillFile("smaqit.new-greenfield-project").Data))
	if _, err := installGlobalTree(fstest.MapFS{"agents/smaqit.qa.md": {Data: []byte("qa")}}, "agents", agents, false); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(agents, "smaqit.new-greenfield-project", "SKILL.md")) {
		t.Error("the skill seed ran on a non-skill destination")
	}

	// A release that still ships a seeded name keeps it.
	again := t.TempDir()
	if _, err := installGlobalTree(fstest.MapFS{"skills/smaqit.new-greenfield-project/SKILL.md": skillFile("smaqit.new-greenfield-project")}, "skills", again, true); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(again, "smaqit.new-greenfield-project", "SKILL.md")) {
		t.Error("a seeded name this release still ships was removed")
	}
}

// An unreadable or tampered manifest stops the install before anything is
// removed, rather than pruning on a list that can't be trusted.
func TestInstallGlobalTreeRefusesAnUntrustedManifest(t *testing.T) {
	release := fstest.MapFS{"skills/smaqit.a/SKILL.md": skillFile("smaqit.a")}
	for name, manifest := range map[string]string{
		"corrupt":        "{not json",
		"unknown schema": `{"schemaVersion":2,"files":[]}`,
		"path escape":    `{"schemaVersion":1,"files":["../outside.txt"]}`,
		"absolute path":  `{"schemaVersion":1,"files":["/etc/hosts"]}`,
	} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			dst := filepath.Join(parent, "skills")
			writeTestFile(t, filepath.Join(parent, "outside.txt"), "must survive")
			writeTestFile(t, filepath.Join(dst, globalManifestName), manifest)
			if _, err := installGlobalTree(release, "skills", dst, true); err == nil {
				t.Fatal("install accepted an untrusted manifest")
			}
			if !exists(filepath.Join(parent, "outside.txt")) {
				t.Error("a file outside the destination was removed")
			}
		})
	}
}

// Uninstall removes the files a manifest lists beyond the current release, then
// the manifest itself; files the manifest never listed stay.
func TestUninstallGlobalManifestRemovesLeftovers(t *testing.T) {
	dst := t.TempDir()
	old := fstest.MapFS{
		"skills/smaqit.a/SKILL.md":   skillFile("smaqit.a"),
		"skills/smaqit.old/SKILL.md": skillFile("smaqit.old"),
	}
	if _, err := installGlobalTree(old, "skills", dst, true); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(dst, "smaqit.task-start", "SKILL.md"), "extension")
	current := fstest.MapFS{"skills/smaqit.a/SKILL.md": skillFile("smaqit.a")}
	if err := uninstallGlobalManifest(current, "skills", dst); err != nil {
		t.Fatal(err)
	}
	if exists(filepath.Join(dst, "smaqit.old")) || exists(filepath.Join(dst, globalManifestName)) {
		t.Error("uninstall left the manifest's leftovers or the manifest")
	}
	if !exists(filepath.Join(dst, "smaqit.task-start", "SKILL.md")) {
		t.Error("uninstall removed a file the manifest never listed")
	}
	// No manifest: nothing to do, no error.
	if err := uninstallGlobalManifest(current, "skills", t.TempDir()); err != nil {
		t.Errorf("no manifest: %v", err)
	}
}

// The retired-skill seed must never name a skill smaqit still ships, or one its
// sibling repositories install into the same directories.
func TestRetiredSkillsAreNotShippedNames(t *testing.T) {
	shipped := map[string]bool{}
	entries, err := fs.ReadDir(skillFilesShared, "skills-shared")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		shipped[e.Name()] = true
	}
	for _, name := range retiredSkills {
		if shipped[name] {
			t.Errorf("retired skill %s is still shipped", name)
		}
		if !strings.HasPrefix(name, "smaqit.") {
			t.Errorf("retired skill %s is not smaqit-prefixed; a user's own skill could share it", name)
		}
	}
}

// No shipped file refers to the skill's old name.
func TestShippedFilesNeverNameNewGreenfieldProject(t *testing.T) {
	for root, embedded := range map[string]fs.FS{
		"skills-shared": skillFilesShared, "skills-claude": skillFilesClaude,
		"agents-copilot": agentFiles, "agents-claude": claudeAgentFiles, "agents-codex": codexAgentFiles,
		"commands-claude": claudeCommandFiles,
	} {
		err := fs.WalkDir(embedded, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, err := fs.ReadFile(embedded, path)
			if err != nil {
				return err
			}
			if strings.Contains(string(data), "new-greenfield-project") {
				t.Errorf("%s still names new-greenfield-project", path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{"../templates/AGENTS.md.template", "../README.md", "../framework/AGENTS.md"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "new-greenfield-project") {
			t.Errorf("%s still names new-greenfield-project", path)
		}
	}
}
