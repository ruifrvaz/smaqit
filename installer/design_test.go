package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestParseDesignRejectsProseAndUnsafeIncludes(t *testing.T) {
	root := designTestProject(t)
	path := filepath.Join(root, "docs", "designs", "business", "dsg-bus-login-use-case.md")
	content := validDesignSource("[SOURCE_SHA256]", "[IMAGE_SHA256]")
	content = strings.Replace(content, "\n```plantuml\n", "\nThis prose is forbidden.\n\n```plantuml\n", 1)
	writeTestFile(t, path, content)
	if _, err := parseDesign(path); err == nil || !strings.Contains(err.Error(), "exactly one PlantUML fence") {
		t.Fatalf("expected prose rejection, got %v", err)
	}

	content = validDesignSource("[SOURCE_SHA256]", "[IMAGE_SHA256]")
	content = strings.Replace(content, "Alice -> Bob", "!include https://example.test/remote.puml\nAlice -> Bob", 1)
	writeTestFile(t, path, content)
	if _, err := parseDesign(path); err == nil || !strings.Contains(err.Error(), "includes/imports") {
		t.Fatalf("expected unsafe include rejection, got %v", err)
	}
}

func TestRenderAttestValidateDesign(t *testing.T) {
	major, err := installedNodeMajor()
	if err != nil || major < minimumNodeMajor {
		t.Skipf("Node %d+ is the mandatory consumer prerequisite: %v", minimumNodeMajor, err)
	}
	root := designTestProject(t)
	if _, err := materializeDesignTools(root); err != nil {
		t.Fatal(err)
	}
	designPath := filepath.Join(root, "docs", "designs", "business", "dsg-bus-login-use-case.md")
	specPath := filepath.Join(root, "specs", "business", "login.md")
	writeTestFile(t, designPath, validDesignSource(`""`, `""`))
	writeTestFile(t, specPath, `---
id: BUS-LOGIN
status: draft
created: 2026-08-03
---

# UC1-LOGIN: Login

## Design References

- [DSG-BUS-LOGIN-USE-CASE](../../docs/designs/business/dsg-bus-login-use-case.md) · [Image](../../docs/designs/business/dsg-bus-login-use-case.png)

## Acceptance Criteria

- **BUS-LOGIN-001**: A user can log in.
`)
	if err := renderDesign(designPath); err != nil {
		t.Fatal(err)
	}
	firstPNG, err := os.ReadFile(strings.TrimSuffix(designPath, ".md") + ".png")
	if err != nil {
		t.Fatal(err)
	}
	if err := renderDesign(designPath); err != nil {
		t.Fatal(err)
	}
	secondPNG, err := os.ReadFile(strings.TrimSuffix(designPath, ".md") + ".png")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstPNG, secondPNG) {
		t.Fatal("identical PlantUML source did not render a deterministic PNG")
	}
	if err := validateDesigns(designPath); err == nil || !strings.Contains(err.Error(), "visual attestation") {
		t.Fatalf("rendered design must remain pending visual review, got %v", err)
	}
	if err := attestDesign(designPath); err != nil {
		t.Fatal(err)
	}
	if err := validateDesigns(designPath); err != nil {
		t.Fatal(err)
	}
	ready, reason := specDesignReady(specPath, "business")
	if !ready {
		t.Fatalf("expected design-ready spec: %s", reason)
	}
	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, specPath, strings.Replace(string(content), "status: draft", "status: implemented", 1))
	design, err := parseDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDesignArtifact(design); err == nil || !strings.Contains(err.Error(), "DESIGN-ARTIFACT-STALE") {
		t.Fatalf("expected lifecycle-behind design to be stale, got %v", err)
	}
}

func TestCanonicalDesignFailureCodes(t *testing.T) {
	root := designTestProject(t)
	designPath := filepath.Join(root, "docs", "designs", "business", "dsg-bus-login-use-case.md")
	writeTestFile(t, designPath, validDesignSource(`""`, `""`))
	design, err := parseDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDesignArtifact(design); err == nil || !strings.Contains(err.Error(), "DESIGN-ARTIFACT-MISSING") {
		t.Fatalf("expected missing artifact code, got %v", err)
	}

	for _, name := range []string{"business", "functional", "stack", "infrastructure", "coverage"} {
		content, err := os.ReadFile(filepath.Join("..", "agents", name+".md"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(content, []byte("DESIGN-VISION-UNAVAILABLE")) {
			t.Fatalf("%s agent has no mandatory vision-unavailable failure", name)
		}
	}
	for _, name := range []string{"development", "deployment", "validation"} {
		content, err := os.ReadFile(filepath.Join("..", "agents", name+".md"))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(content, []byte("DESIGN-VISION-UNAVAILABLE")) || !bytes.Contains(content, []byte("linked PlantUML source")) {
			t.Fatalf("%s agent does not enforce source-only design consumption", name)
		}
	}
}

func TestCorruptRuntimeAndArchiveAreRejected(t *testing.T) {
	root := designTestProject(t)
	runtimeDir, err := materializeDesignTools(root)
	if err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(runtimeDir, "package-lock.json")
	if err := os.WriteFile(lockPath, []byte("corrupt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateMaterializedDesignTools(runtimeDir); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("expected corrupt runtime rejection, got %v", err)
	}
	if err := runPlantUMLMCP(root); err == nil || !strings.Contains(err.Error(), "DESIGN-TOOLCHAIN-UNAVAILABLE") {
		t.Fatalf("expected mandatory toolchain failure code, got %v", err)
	}

	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "../escape", Mode: 0o644, Size: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := extractDesignToolArchive(archive.Bytes(), t.TempDir()); err == nil || !strings.Contains(err.Error(), "unsafe archive path") {
		t.Fatalf("expected archive traversal rejection, got %v", err)
	}
}

func TestRenderRejectsInvalidPlantUMLSyntax(t *testing.T) {
	major, err := installedNodeMajor()
	if err != nil || major < minimumNodeMajor {
		t.Skipf("Node %d+ is the mandatory consumer prerequisite: %v", minimumNodeMajor, err)
	}
	root := designTestProject(t)
	if _, err := materializeDesignTools(root); err != nil {
		t.Fatal(err)
	}
	designPath := filepath.Join(root, "docs", "designs", "business", "dsg-bus-login-use-case.md")
	content := strings.Replace(validDesignSource(`""`, `""`), "Alice -> Bob", "!definitely_not_a_plantuml_directive", 1)
	writeTestFile(t, designPath, content)
	if err := renderDesign(designPath); err == nil || !strings.Contains(err.Error(), "DESIGN-SYNTAX-INVALID") {
		t.Fatalf("expected PlantUML syntax failure code, got %v", err)
	}
}

func TestGreenfieldReferencePlantUMLIsValid(t *testing.T) {
	major, err := installedNodeMajor()
	if err != nil || major < minimumNodeMajor {
		t.Skipf("Node %d+ is the mandatory consumer prerequisite: %v", minimumNodeMajor, err)
	}
	root := designTestProject(t)
	if _, err := materializeDesignTools(root); err != nil {
		t.Fatal(err)
	}
	session, err := openPlantUMLSession(root)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	content, err := os.ReadFile(filepath.Join("..", "skills", "smaqit.new-greenfield-project", "references", "diagrams.md"))
	if err != nil {
		t.Fatal(err)
	}
	blocks := regexp.MustCompile("(?s)```plantuml\\n(.*?)```").FindAllSubmatch(content, -1)
	if len(blocks) != 5 {
		t.Fatalf("expected five PlantUML reference diagrams, found %d", len(blocks))
	}
	for index, block := range blocks {
		if _, err := callPlantUML(session, "check_syntax", normalizePlantUMLSource(string(block[1]))); err != nil {
			t.Fatalf("reference diagram %d: %v", index+1, err)
		}
	}
}

func designTestProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{
		".smaqit", "docs/designs/business", "specs/business",
	} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func validDesignSource(sourceHash, imageHash string) string {
	return fmt.Sprintf(`---
id: DSG-BUS-LOGIN-USE-CASE
status: draft
created: 2026-08-03
layer: business
diagram_type: use-case
notation: plantuml
specifications:
  - ../../../specs/business/login.md
requirements:
  - BUS-LOGIN-001
source_sha256: %s
image_sha256: %s
visual_validation:
  status: pending
  validated_at: null
  source_sha256: null
  image_sha256: null
---

`+"```plantuml"+`
@startuml
Alice -> Bob
@enduml
`+"```"+`
`, sourceHash, imageHash)
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
