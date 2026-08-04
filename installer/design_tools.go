package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tailscale/hujson"
)

const (
	designMCPServerName = "smaqit-plantuml"
	designBundleVersion = "plantuml-mcp-js-0.2.0_resvg-wasm-2.6.2_noto-sans-5.3.0"
	minimumNodeMajor    = 22
)

type designToolManifest struct {
	Schema            int               `json:"schema"`
	BundleVersion     string            `json:"bundle_version"`
	NodeMinimumMajor  int               `json:"node_minimum_major"`
	PlantUMLMCP       string            `json:"plantuml_mcp"`
	ResvgWASM         string            `json:"resvg_wasm"`
	NotoSans          string            `json:"noto_sans"`
	PackageLockSHA256 string            `json:"package_lock_sha256"`
	Files             map[string]string `json:"files"`
}

func preflightDesignTools() error {
	archive, err := readEmbeddedDesignToolArchive()
	if err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: embedded runtime is invalid: %w", err)
	}
	probeDir, err := os.MkdirTemp("", "smaqit-design-preflight-")
	if err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: cannot inspect embedded runtime: %w", err)
	}
	defer os.RemoveAll(probeDir)
	if err := extractDesignToolArchive(archive, probeDir); err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: embedded runtime is invalid: %w", err)
	}
	if err := validateMaterializedDesignTools(probeDir); err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: embedded runtime is invalid: %w", err)
	}
	major, err := installedNodeMajor()
	if err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: %w", err)
	}
	if major < minimumNodeMajor {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: Node %d or newer is required (found %d)", minimumNodeMajor, major)
	}
	return nil
}

func installedNodeMajor() (int, error) {
	out, err := exec.Command("node", "--version").Output()
	if err != nil {
		return 0, errors.New("Node is not installed or is not available on PATH")
	}
	version := strings.TrimSpace(string(out))
	version = strings.TrimPrefix(version, "v")
	majorText, _, _ := strings.Cut(version, ".")
	major, err := strconv.Atoi(majorText)
	if err != nil {
		return 0, fmt.Errorf("could not parse Node version %q", version)
	}
	return major, nil
}

func readEmbeddedDesignToolArchive() ([]byte, error) {
	b, err := designToolArchive.ReadFile("tools/plantuml-tools.tgz")
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, errors.New("embedded archive is empty")
	}
	return b, nil
}

func materializeDesignTools(projectRoot string) (string, error) {
	archive, err := readEmbeddedDesignToolArchive()
	if err != nil {
		return "", err
	}
	toolsRoot := filepath.Join(projectRoot, ".smaqit", "tools", "plantuml")
	destination := filepath.Join(toolsRoot, designBundleVersion)
	if validateMaterializedDesignTools(destination) == nil {
		return destination, nil
	}

	if err := os.MkdirAll(toolsRoot, 0o755); err != nil {
		return "", err
	}
	temporary := filepath.Join(toolsRoot, fmt.Sprintf(".%s.tmp-%d", designBundleVersion, time.Now().UnixNano()))
	if err := os.MkdirAll(temporary, 0o755); err != nil {
		return "", err
	}
	cleanupTemporary := true
	defer func() {
		if cleanupTemporary {
			_ = os.RemoveAll(temporary)
		}
	}()
	if err := extractDesignToolArchive(archive, temporary); err != nil {
		return "", err
	}
	if err := validateMaterializedDesignTools(temporary); err != nil {
		return "", err
	}

	backup := destination + ".previous"
	_ = os.RemoveAll(backup)
	if _, err := os.Stat(destination); err == nil {
		if err := os.Rename(destination, backup); err != nil {
			return "", err
		}
	}
	if err := os.Rename(temporary, destination); err != nil {
		if _, backupErr := os.Stat(backup); backupErr == nil {
			_ = os.Rename(backup, destination)
		}
		return "", err
	}
	cleanupTemporary = false
	_ = os.RemoveAll(backup)
	return destination, nil
}

func extractDesignToolArchive(archive []byte, destination string) error {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		clean := filepath.Clean(filepath.FromSlash(header.Name))
		if clean == "." {
			continue
		}
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("unsafe archive path %q", header.Name)
		}
		target := filepath.Join(destination, clean)
		rel, err := filepath.Rel(destination, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("archive path escapes destination: %q", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(f, tr)
			closeErr := f.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		default:
			return fmt.Errorf("unsupported archive entry %q (type %d)", header.Name, header.Typeflag)
		}
	}
}

func validateMaterializedDesignTools(runtimeDir string) error {
	manifestBytes, err := os.ReadFile(filepath.Join(runtimeDir, "manifest.json"))
	if err != nil {
		return err
	}
	var manifest designToolManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return err
	}
	if manifest.Schema != 1 || manifest.BundleVersion != designBundleVersion || manifest.NodeMinimumMajor != minimumNodeMajor || manifest.PlantUMLMCP != "0.2.0" || manifest.ResvgWASM != "2.6.2" || manifest.NotoSans != "5.3.0" {
		return errors.New("runtime manifest does not match the smaqit build")
	}
	lock, err := os.ReadFile(filepath.Join(runtimeDir, "package-lock.json"))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(lock)
	if hex.EncodeToString(sum[:]) != manifest.PackageLockSHA256 {
		return errors.New("runtime package-lock hash mismatch")
	}
	if len(manifest.Files) == 0 {
		return errors.New("runtime manifest has no file integrity inventory")
	}
	for relative, expectedHash := range manifest.Files {
		content, err := os.ReadFile(filepath.Join(runtimeDir, filepath.FromSlash(relative)))
		if err != nil {
			return fmt.Errorf("runtime file missing: %s", relative)
		}
		if hashBytes(content) != expectedHash {
			return fmt.Errorf("runtime file hash mismatch: %s", relative)
		}
	}
	seenFiles := 0
	if err := filepath.WalkDir(runtimeDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("runtime symlink is forbidden: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(runtimeDir, path)
		if err != nil {
			return err
		}
		if filepath.ToSlash(relative) == "manifest.json" {
			return nil
		}
		seenFiles++
		if _, exists := manifest.Files[filepath.ToSlash(relative)]; !exists {
			return fmt.Errorf("unexpected runtime file: %s", relative)
		}
		return nil
	}); err != nil {
		return err
	}
	if seenFiles != len(manifest.Files) {
		return errors.New("runtime file inventory count mismatch")
	}
	for _, required := range []string{
		"node_modules/@plantuml/mcp-js/server.js",
		"node_modules/@plantuml/mcp-js/engine.js",
		"node_modules/@resvg/resvg-wasm/index_bg.wasm",
		"node_modules/@fontsource/noto-sans/files/noto-sans-latin-400-normal.woff2",
		"render-png.mjs",
		"THIRD_PARTY_NOTICES.md",
		"MPL-2.0.txt",
	} {
		if info, err := os.Stat(filepath.Join(runtimeDir, filepath.FromSlash(required))); err != nil || info.IsDir() {
			return fmt.Errorf("runtime file missing: %s", required)
		}
	}
	return nil
}

func findProjectRoot(start string) (string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(current, ".smaqit")); err == nil && info.IsDir() {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("not inside an initialized smaqit project")
		}
		current = parent
	}
}

func runPlantUMLMCP(projectRoot string) error {
	runtimeDir := filepath.Join(projectRoot, ".smaqit", "tools", "plantuml", designBundleVersion)
	if err := preflightDesignTools(); err != nil {
		return err
	}
	if err := validateMaterializedDesignTools(runtimeDir); err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: %w", err)
	}
	cmd := exec.Command("node", filepath.Join(runtimeDir, "node_modules", "@plantuml", "mcp-js", "server.js"))
	cmd.Dir = projectRoot
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func vscodeMCPDefinition() map[string]any {
	return map[string]any{
		"type":    "stdio",
		"command": "smaqit",
		"args":    []any{"mcp", "plantuml"},
	}
}

func installVSCodeMCPConfig(projectRoot string) error {
	path := filepath.Join(projectRoot, ".vscode", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	definition := vscodeMCPDefinition()
	input, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		content := map[string]any{"servers": map[string]any{designMCPServerName: definition}}
		b, marshalErr := json.MarshalIndent(content, "", "  ")
		if marshalErr != nil {
			return marshalErr
		}
		return os.WriteFile(path, append(b, '\n'), 0o644)
	}
	if err != nil {
		return err
	}
	standard, err := hujson.Standardize(input)
	if err != nil {
		return fmt.Errorf("cannot parse %s: %w", path, err)
	}
	var document map[string]any
	if err := json.Unmarshal(standard, &document); err != nil {
		return err
	}
	servers, exists := document["servers"].(map[string]any)
	if !exists && document["servers"] != nil {
		return errors.New(".vscode/mcp.json field 'servers' must be an object")
	}
	if servers != nil {
		if existing, exists := servers[designMCPServerName]; exists {
			if sameJSON(existing, definition) {
				return nil
			}
			return fmt.Errorf("MCP server name %q is already owned by another configuration", designMCPServerName)
		}
	}

	value, err := hujson.Parse(input)
	if err != nil {
		return err
	}
	var patch []byte
	if servers == nil {
		patch, _ = json.Marshal([]map[string]any{{"op": "add", "path": "/servers", "value": map[string]any{designMCPServerName: definition}}})
	} else {
		patch, _ = json.Marshal([]map[string]any{{"op": "add", "path": "/servers/" + designMCPServerName, "value": definition}})
	}
	if err := value.Patch(patch); err != nil {
		return err
	}
	value.Format()
	return os.WriteFile(path, value.Pack(), 0o644)
}

func preflightVSCodeMCPConfig(projectRoot string) error {
	path := filepath.Join(projectRoot, ".vscode", "mcp.json")
	input, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	standard, err := hujson.Standardize(input)
	if err != nil {
		return fmt.Errorf("cannot parse %s: %w", path, err)
	}
	var document map[string]any
	if err := json.Unmarshal(standard, &document); err != nil {
		return err
	}
	servers, exists := document["servers"].(map[string]any)
	if !exists && document["servers"] != nil {
		return errors.New(".vscode/mcp.json field 'servers' must be an object")
	}
	if existing, exists := servers[designMCPServerName]; exists && !sameJSON(existing, vscodeMCPDefinition()) {
		return fmt.Errorf("MCP server name %q is already owned by another configuration", designMCPServerName)
	}
	return nil
}

func removeVSCodeMCPConfig(projectRoot string) error {
	path := filepath.Join(projectRoot, ".vscode", "mcp.json")
	input, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	standard, err := hujson.Standardize(input)
	if err != nil {
		return err
	}
	var document map[string]any
	if err := json.Unmarshal(standard, &document); err != nil {
		return err
	}
	servers, _ := document["servers"].(map[string]any)
	existing, exists := servers[designMCPServerName]
	if !exists || !sameJSON(existing, vscodeMCPDefinition()) {
		return nil
	}
	if len(servers) == 1 && len(document) == 1 {
		return os.Remove(path)
	}
	value, err := hujson.Parse(input)
	if err != nil {
		return err
	}
	patch, _ := json.Marshal([]map[string]any{{"op": "remove", "path": "/servers/" + designMCPServerName}})
	if err := value.Patch(patch); err != nil {
		return err
	}
	value.Format()
	return os.WriteFile(path, value.Pack(), 0o644)
}

func validateVSCodeMCPConfig(projectRoot string) error {
	path := filepath.Join(projectRoot, ".vscode", "mcp.json")
	input, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	standard, err := hujson.Standardize(input)
	if err != nil {
		return err
	}
	var document map[string]any
	if err := json.Unmarshal(standard, &document); err != nil {
		return err
	}
	servers, ok := document["servers"].(map[string]any)
	if !ok || !sameJSON(servers[designMCPServerName], vscodeMCPDefinition()) {
		return fmt.Errorf("mandatory MCP server %q is missing or incompatible", designMCPServerName)
	}
	return nil
}

func sameJSON(a, b any) bool {
	aa, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	return bytes.Equal(aa, bb)
}

func designRuntimePath(projectRoot string, parts ...string) string {
	all := append([]string{projectRoot, ".smaqit", "tools", "plantuml", designBundleVersion}, parts...)
	return filepath.Join(all...)
}

func nodeCommandName() string {
	if runtime.GOOS == "windows" {
		return "node.exe"
	}
	return "node"
}
