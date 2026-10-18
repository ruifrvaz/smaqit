package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Global install writes a manifest of the files it placed into each global
// destination. The next install removes the manifest's files that the current
// release no longer ships, so a renamed or retired skill, agent, or command
// disappears instead of staying installed beside its replacement. The manifest
// lives inside the destination it describes, so it follows CLAUDE_CONFIG_DIR,
// COPILOT_HOME, and CODEX_HOME, and skill and agent scanners ignore it.
const globalManifestName = ".smaqit-installed.json"

// retiredSkills are skills a release installed globally before manifests
// existed. They are removed from a skill destination only when their SKILL.md
// still names them, so a user's own skill of another name is never touched.
// smaqit-extensions and smaqit-adk install into the same directories; none of
// these names is theirs.
var retiredSkills = []string{
	"smaqit.infrastructure-onboard-k3s-app", // v3.4.0–v3.5.0, removed by task 119
	"smaqit.new-greenfield-project",         // v3.0.0–v3.17.2, renamed smaqit.new-project
}

type globalManifest struct {
	SchemaVersion int      `json:"schemaVersion"`
	Version       string   `json:"version"`
	Files         []string `json:"files"`
}

// embeddedFiles lists the slash-separated paths of every file under srcDir,
// relative to it.
func embeddedFiles(embeddedFS fs.FS, srcDir string) ([]string, error) {
	var files []string
	err := fs.WalkDir(embeddedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, strings.TrimPrefix(path, srcDir+"/"))
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

// readGlobalManifest returns the files a previous install recorded in dstDir.
// A missing manifest is no error; an unreadable one is, so nothing is pruned
// on the strength of a list that can't be trusted.
func readGlobalManifest(dstDir string) ([]string, bool, error) {
	data, err := os.ReadFile(filepath.Join(dstDir, globalManifestName))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var manifest globalManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, false, fmt.Errorf("parsing %s: %w", filepath.Join(dstDir, globalManifestName), err)
	}
	if manifest.SchemaVersion != 1 {
		return nil, false, fmt.Errorf("%s has unsupported schemaVersion %d", filepath.Join(dstDir, globalManifestName), manifest.SchemaVersion)
	}
	return manifest.Files, true, nil
}

func writeGlobalManifest(dstDir string, files []string) error {
	data, err := json.MarshalIndent(globalManifest{SchemaVersion: 1, Version: Version, Files: files}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dstDir, globalManifestName), append(data, '\n'), 0644)
}

// safeManifestPath rejects a manifest entry that would leave dstDir.
func safeManifestPath(dstDir, rel string) (string, bool) {
	clean := filepath.Clean(filepath.FromSlash(rel))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.Join(dstDir, clean), true
}

// removeFilesAndEmptyParents removes the listed files under dstDir, then every
// parent directory left empty, never dstDir itself. It returns the number of
// files removed.
func removeFilesAndEmptyParents(dstDir string, rels []string) (int, error) {
	removed := 0
	dirs := map[string]bool{}
	for _, rel := range rels {
		path, ok := safeManifestPath(dstDir, rel)
		if !ok {
			return removed, fmt.Errorf("manifest entry %q leaves %s", rel, dstDir)
		}
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return removed, err
		}
		if info.IsDir() {
			continue
		}
		if err := os.Remove(path); err != nil {
			return removed, fmt.Errorf("removing %s: %w", path, err)
		}
		removed++
		for dir := filepath.Dir(path); dir != dstDir && strings.HasPrefix(dir, dstDir+string(filepath.Separator)); dir = filepath.Dir(dir) {
			dirs[dir] = true
		}
	}
	ordered := make([]string, 0, len(dirs))
	for dir := range dirs {
		ordered = append(ordered, dir)
	}
	// Children before their parents.
	sort.Slice(ordered, func(i, j int) bool {
		return strings.Count(ordered[i], string(filepath.Separator)) > strings.Count(ordered[j], string(filepath.Separator))
	})
	for _, dir := range ordered {
		if _, err := removeDirIfEmpty(dir); err != nil {
			return removed, fmt.Errorf("pruning %s: %w", dir, err)
		}
	}
	return removed, nil
}

// skillName reads the name: field of a SKILL.md frontmatter block.
func skillName(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return "", nil
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return "", nil
	}
	var frontmatter struct {
		Name string `yaml:"name"`
	}
	if err := yaml.Unmarshal([]byte(text[4:4+end]), &frontmatter); err != nil {
		return "", nil
	}
	return frontmatter.Name, nil
}

// removeRetiredSkills deletes each retired skill directory in dstDir whose
// SKILL.md still names that skill. It returns the names it removed.
func removeRetiredSkills(dstDir string, shipped map[string]bool) ([]string, error) {
	var removed []string
	for _, name := range retiredSkills {
		if shipped[name] {
			continue
		}
		dir := filepath.Join(dstDir, name)
		got, err := skillName(filepath.Join(dir, "SKILL.md"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return removed, err
		}
		if got != name {
			continue
		}
		if err := os.RemoveAll(dir); err != nil {
			return removed, fmt.Errorf("removing retired skill %s: %w", dir, err)
		}
		removed = append(removed, name)
	}
	return removed, nil
}

// installGlobalTree copies an embedded tree into dstDir, then removes what an
// earlier smaqit release installed there and this one no longer ships: the
// previous manifest's leftover files, and for skill trees the retired skills.
// It returns the number of stale files and retired skills removed.
func installGlobalTree(embeddedFS fs.FS, srcDir, dstDir string, skills bool) (int, error) {
	current, err := embeddedFiles(embeddedFS, srcDir)
	if err != nil {
		return 0, err
	}
	previous, _, err := readGlobalManifest(dstDir)
	if err != nil {
		return 0, err
	}
	if err := copyEmbeddedDir(embeddedFS, srcDir, dstDir); err != nil {
		return 0, err
	}
	shipped := map[string]bool{}
	for _, rel := range current {
		shipped[rel] = true
		shipped[strings.SplitN(rel, "/", 2)[0]] = true
	}
	var stale []string
	for _, rel := range previous {
		if !shipped[rel] {
			stale = append(stale, rel)
		}
	}
	removed, err := removeFilesAndEmptyParents(dstDir, stale)
	if err != nil {
		return removed, err
	}
	if skills {
		retired, err := removeRetiredSkills(dstDir, shipped)
		if err != nil {
			return removed, err
		}
		removed += len(retired)
	}
	return removed, writeGlobalManifest(dstDir, current)
}

// uninstallGlobalManifest removes the files a manifest lists that the current
// release does not ship (the current ones are removed by the caller), then the
// manifest itself.
func uninstallGlobalManifest(embeddedFS fs.FS, srcDir, dstDir string) error {
	previous, found, err := readGlobalManifest(dstDir)
	if err != nil || !found {
		return err
	}
	current, err := embeddedFiles(embeddedFS, srcDir)
	if err != nil {
		return err
	}
	shipped := map[string]bool{}
	for _, rel := range current {
		shipped[rel] = true
	}
	var stale []string
	for _, rel := range previous {
		if !shipped[rel] {
			stale = append(stale, rel)
		}
	}
	if _, err := removeFilesAndEmptyParents(dstDir, stale); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(dstDir, globalManifestName)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
