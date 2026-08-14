package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
	assertOpaqueCreamPNG(t, firstPNG)
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
	if _, err := ensureDesignTools(root); err != nil {
		t.Fatalf("expected corrupt runtime repair, got %v", err)
	}
	if err := validateMaterializedDesignTools(runtimeDir); err != nil {
		t.Fatalf("repaired runtime is invalid: %v", err)
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

func TestMaterializeDesignToolsIsConcurrentSafe(t *testing.T) {
	root := designTestProject(t)
	const callers = 4
	errs := make(chan error, callers)
	var group sync.WaitGroup
	for range callers {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := materializeDesignTools(root)
			errs <- err
		}()
	}
	group.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := validateMaterializedDesignTools(designRuntimePath(root)); err != nil {
		t.Fatalf("concurrent materialization did not leave a valid bundle: %v", err)
	}
}

func TestValidateDesignsReportsIndependentFailures(t *testing.T) {
	major, err := installedNodeMajor()
	if err != nil || major < minimumNodeMajor {
		t.Skipf("Node %d+ is the mandatory consumer prerequisite: %v", minimumNodeMajor, err)
	}
	root := designTestProject(t)
	t.Chdir(root)
	validDesign := filepath.Join(root, "docs", "designs", "business", "dsg-bus-login-use-case.md")
	brokenDesign := filepath.Join(root, "docs", "designs", "business", "broken.md")
	linkedSpec := filepath.Join(root, "specs", "business", "login.md")
	orphanSpec := filepath.Join(root, "specs", "business", "orphan.md")
	writeTestFile(t, validDesign, validDesignSource(`""`, `""`))
	writeTestFile(t, brokenDesign, "not a design\n")
	writeTestFile(t, linkedSpec, `---
id: BUS-LOGIN
status: draft
created: 2026-08-03
---

## Design References

- [Design](../../docs/designs/business/dsg-bus-login-use-case.md) · [Image](../../docs/designs/business/dsg-bus-login-use-case.png)

## Acceptance Criteria

- **BUS-LOGIN-001**: Login works.
`)
	writeTestFile(t, orphanSpec, `---
id: BUS-ORPHAN
status: draft
created: 2026-08-03
---

## Acceptance Criteria

- **BUS-ORPHAN-001**: Orphan is reported.
`)
	err = validateDesigns("")
	if err == nil {
		t.Fatal("expected aggregate validation failure")
	}
	for _, want := range []string{"broken.md:", "dsg-bus-login-use-case.md:", "active specification orphan.md:"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("aggregate diagnostics missing %q: %v", want, err)
		}
	}
}

func TestExtractDesignToolArchiveRejectsEscapingPaths(t *testing.T) {
	for _, archivePath := range []string{
		"../escape",
		"nested/../../escape",
		"./../escape",
		"/absolute/escape",
		"//network-share/escape",
	} {
		t.Run(archivePath, func(t *testing.T) {
			destination := t.TempDir()
			err := extractDesignToolArchive(singleFileArchive(t, archivePath), destination)
			if err == nil || !strings.Contains(err.Error(), "unsafe archive path") {
				t.Fatalf("expected unsafe archive path rejection, got %v", err)
			}
		})
	}
}

func TestExtractDesignToolArchiveKeepsLocalPathContained(t *testing.T) {
	destination := t.TempDir()
	if err := extractDesignToolArchive(singleFileArchive(t, "nested/runtime.txt"), destination); err != nil {
		t.Fatalf("extract local entry: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(destination, "nested", "runtime.txt"))
	if err != nil {
		t.Fatalf("read extracted local entry: %v", err)
	}
	if string(content) != "x" {
		t.Fatalf("unexpected extracted content %q", content)
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

func TestSystemSequenceProfileEnforcesBlackBox(t *testing.T) {
	root := designTestProject(t)
	path := filepath.Join(root, "docs", "designs", "functional", "dsg-fun-test-flow-system-sequence.md")

	blackBox := `@startuml
actor Customer
participant "SYSTEM" as SYSTEM
hide footbox

Customer -> SYSTEM: GET /activate?token
activate SYSTEM
SYSTEM --> Customer: activation form (rendered unconditionally)
deactivate SYSTEM

Customer -> SYSTEM: POST /activate (token, password)
activate SYSTEM
alt token valid and unexpired
  SYSTEM --> Customer: activated
else token invalid, expired, or already used
  SYSTEM --> Customer: generic rejection, form re-rendered
end
deactivate SYSTEM

note over SYSTEM
  RegistrationToken is issued earlier, at customer
  creation, out of scope for this diagram.
end note
@enduml`

	decoratedAlias := `@startuml
actor Customer
participant "System" as System #LightBlue
hide footbox
Customer -> System: GET /activate?token
System --> Customer: 200
@enduml`

	multiLineTitle := `@startuml
title
Describes the Customer -> System handoff
end title
actor Customer
participant "System" as System
hide footbox
Customer -> System: GET /activate?token
System --> Customer: 200
@enduml`

	multiParticipant := `@startuml
actor Customer
participant "VisitorConvertHandler / CustomerNewHandler" as CreateHandler
participant CustomerActivationHandler as ActivateHandler
participant CustomerActivationService as ActSvc
hide footbox

== Token Issuance ==
Customer -> CreateHandler: POST /visitors/{id}/convert
activate CreateHandler
CreateHandler -> ActivateHandler: forward
CreateHandler --> Customer: 200 (customer created)
deactivate CreateHandler

== Submit Activation ==
Customer -> ActivateHandler: POST /activate
activate ActivateHandler
ActivateHandler -> ActSvc: activate(token, password)
activate ActSvc
ActSvc --> ActivateHandler: activated
ActivateHandler --> Customer: 200
deactivate ActSvc
deactivate ActivateHandler
@enduml`

	twoActors := `@startuml
actor Admin
actor Customer
participant "System" as System
hide footbox
Admin -> System: seed data
Customer -> System: GET /activate?token
System --> Customer: 200
@enduml`

	misnamedSystem := `@startuml
actor Customer
participant "System" as CRM
hide footbox
Customer -> CRM: GET /activate?token
CRM --> Customer: 200
@enduml`

	noExplicitSystem := `@startuml
actor Customer
hide footbox
Customer -> System: GET /activate?token
System --> Customer: 200
@enduml`

	noActor := `@startuml
participant "System" as System
hide footbox
System --> System: noop
@enduml`

	missingFootbox := `@startuml
actor Customer
participant "System" as System
Customer -> System: request
System --> Customer: response
@enduml`

	footer := `@startuml
actor Customer
participant "System" as System
hide footbox
footer generated externally
Customer -> System: request
System --> Customer: response
@enduml`

	inferredEndpoint := `@startuml
actor Customer
participant "System" as System
hide footbox
Customer -> Other: request
Other --> Customer: response
@enduml`

	tests := []struct {
		name     string
		plantUML string
		wantErr  bool
		wantMsg  string
	}{
		{"case-insensitive visible System passes", blackBox, false, ""},
		{"decorated canonical System passes", decoratedAlias, false, ""},
		{"multi-line title body is not parsed as content", multiLineTitle, false, ""},
		{"two actors fails", twoActors, true, "exactly one actor"},
		{"explicit multi participant fails", multiParticipant, true, "exactly one system participant"},
		{"system participant alias not named System fails", misnamedSystem, true, `participant "System" as System`},
		{"system participant never explicitly declared fails", noExplicitSystem, true, "found none"},
		{"no actor declared fails", noActor, true, "exactly one actor"},
		{"missing hide footbox fails", missingFootbox, true, "hide footbox"},
		{"footer fails", footer, true, "footer directive"},
		{"inferred endpoint fails", inferredEndpoint, true, "message endpoints"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			writeTestFile(t, path, validFunctionalSystemSequenceSource(tc.plantUML))
			_, err := parseDesign(path)
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "DESIGN-VISUAL-INVALID") || !strings.Contains(err.Error(), tc.wantMsg) {
					t.Fatalf("expected system-sequence profile failure containing %q, got %v", tc.wantMsg, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected success, got %v", err)
			}
		})
	}
}

func TestSystemSequenceProfileAcceptsShippedTemplateAndReference(t *testing.T) {
	block := func(path string) string {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		matches := regexp.MustCompile("(?s)```plantuml\\n(.*?)```").FindSubmatch(content)
		if matches == nil {
			t.Fatalf("no PlantUML fence found in %s", path)
		}
		return string(matches[1])
	}
	for _, path := range []string{
		filepath.Join("..", "templates", "designs", "functional.template.md"),
	} {
		source := normalizePlantUMLSource(block(path))
		if err := validateSystemSequenceProfile(&designArtifact{Source: source}); err != nil {
			t.Fatalf("%s: expected shipped reference to satisfy the black-box profile, got %v", path, err)
		}
	}
}

func TestSystemSequenceProfileSupportsMultipleDesignsPerSpec(t *testing.T) {
	major, err := installedNodeMajor()
	if err != nil || major < minimumNodeMajor {
		t.Skipf("Node %d+ is the mandatory consumer prerequisite: %v", minimumNodeMajor, err)
	}
	root := designTestProject(t)
	for _, dir := range []string{"docs/designs/functional", "specs/functional"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	registrationPath := filepath.Join(root, "docs", "designs", "functional", "dsg-fun-test-flow-registration-system-sequence.md")
	activationPath := filepath.Join(root, "docs", "designs", "functional", "dsg-fun-test-flow-activation-system-sequence.md")
	specPath := filepath.Join(root, "specs", "functional", "test-flow.md")

	writeTestFile(t, registrationPath, strings.Replace(
		validFunctionalSystemSequenceSource(`@startuml
actor Admin
participant "System" as System
hide footbox
Admin -> System: create customer
System --> Admin: 200
@enduml`),
		"DSG-FUN-TEST-FLOW-SYSTEM-SEQUENCE", "DSG-FUN-TEST-FLOW-REGISTRATION-SYSTEM-SEQUENCE", 1))
	writeTestFile(t, activationPath, strings.Replace(
		validFunctionalSystemSequenceSource(`@startuml
actor Customer
participant "System" as System
hide footbox
Customer -> System: GET /activate?token
System --> Customer: 200
@enduml`),
		"DSG-FUN-TEST-FLOW-SYSTEM-SEQUENCE", "DSG-FUN-TEST-FLOW-ACTIVATION-SYSTEM-SEQUENCE", 1))
	writeTestFile(t, specPath, `---
id: FUN-TEST-FLOW
status: draft
created: 2026-08-07
---

# Test Flow

## Design References

- [DSG-FUN-TEST-FLOW-REGISTRATION-SYSTEM-SEQUENCE](../../docs/designs/functional/dsg-fun-test-flow-registration-system-sequence.md) · [Image](../../docs/designs/functional/dsg-fun-test-flow-registration-system-sequence.png)
- [DSG-FUN-TEST-FLOW-ACTIVATION-SYSTEM-SEQUENCE](../../docs/designs/functional/dsg-fun-test-flow-activation-system-sequence.md) · [Image](../../docs/designs/functional/dsg-fun-test-flow-activation-system-sequence.png)

## Acceptance Criteria

- **FUN-TEST-FLOW-001**: A customer can activate their account.
`)

	for _, path := range []string{registrationPath, activationPath} {
		if err := renderDesign(path); err != nil {
			t.Fatalf("render %s: %v", path, err)
		}
		if err := attestDesign(path); err != nil {
			t.Fatalf("attest %s: %v", path, err)
		}
	}
	ready, reason := specDesignReady(specPath, "functional")
	if !ready {
		t.Fatalf("expected spec with two linked black-box system-sequence designs to be ready: %s", reason)
	}
}

func validFunctionalSystemSequenceSource(plantUML string) string {
	return fmt.Sprintf(`---
id: DSG-FUN-TEST-FLOW-SYSTEM-SEQUENCE
status: draft
created: 2026-08-07
layer: functional
diagram_type: system-sequence
notation: plantuml
specifications:
  - ../../../specs/functional/test-flow.md
requirements:
  - FUN-TEST-FLOW-001
source_sha256: "[SOURCE_SHA256]"
image_sha256: "[IMAGE_SHA256]"
visual_validation:
  status: pending
  validated_at: null
  source_sha256: null
  image_sha256: null
---

`+"```plantuml"+`
%s
`+"```"+`
`, plantUML)
}

func designTestProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{
		".smaqit", "docs/designs/business", "docs/designs/functional", "docs/designs/design-sequence",
		"specs/business", "specs/functional",
	} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// validSystemSequenceSource builds a minimal DSG-FUN-ORDER-SYSTEM-SEQUENCE
// fixture with one labeled actor->system arrow per operation.
func validSystemSequenceSource(sourceHash, imageHash string, operations []string) string {
	lines := []string{`actor Customer`, `participant "System" as System`, `hide footbox`}
	for _, op := range operations {
		lines = append(lines, fmt.Sprintf("Customer -> System: %s", op))
	}
	return fmt.Sprintf(`---
id: DSG-FUN-ORDER-SYSTEM-SEQUENCE
status: implemented
created: 2026-08-07
layer: functional
diagram_type: system-sequence
notation: plantuml
specifications:
  - ../../../specs/functional/order.md
requirements:
  - FUN-ORDER-001
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
%s
@enduml
`+"```"+`
`, sourceHash, imageHash, strings.Join(lines, "\n"))
}

// validDesignSequenceSource builds a DSG-DSD-ORDER-DESIGN-SEQUENCE fixture
// realizing the system-sequence fixture above, one internal-collaborator
// arrow per operation, each optionally followed by a `' impl:` citation
// (citations[i] == "" omits the citation for that operation).
func validDesignSequenceSource(sourceHash, imageHash, realizes string, operations, citations []string) string {
	lines := []string{`participant "OrderHandler" as Handler`, `participant "OrderService" as Service`}
	for i, op := range operations {
		lines = append(lines, fmt.Sprintf("Handler -> Service: %s", op))
		if i < len(citations) && citations[i] != "" {
			lines = append(lines, "' impl: "+citations[i])
		}
	}
	return fmt.Sprintf(`---
id: DSG-DSD-ORDER-DESIGN-SEQUENCE
status: implemented
created: 2026-08-07
layer: design-sequence
diagram_type: design-sequence
notation: plantuml
realizes: %s
specifications:
  - ../../../specs/functional/order.md
requirements:
  - FUN-ORDER-001
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
%s
@enduml
`+"```"+`
`, realizes, sourceHash, imageHash, strings.Join(lines, "\n"))
}

// minimalValidPNG returns PNG bytes satisfying validateDesignArtifact's
// signature/dimension checks (width 640-4096, height 1-16384) without going
// through the PlantUML render pipeline, so attestation-gating tests don't
// need Node/MCP.
func minimalValidPNG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 640, 1))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// writeAttestableDesign writes a design-sequence markdown fixture whose
// source/image hashes already match a freshly written PNG, so attestDesign
// can be exercised directly without renderDesign/Node.
func writeAttestableDesign(t *testing.T, root, designPath, realizes string, operations, citations []string) {
	t.Helper()
	imageBytes := minimalValidPNG(t)
	imageHash := hashBytes(imageBytes)
	if err := os.WriteFile(strings.TrimSuffix(designPath, ".md")+".png", imageBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	source := validDesignSequenceSource("PLACEHOLDER", imageHash, realizes, operations, citations)
	writeTestFile(t, designPath, source)
	d, err := parseDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	final := validDesignSequenceSource(d.SourceHash, imageHash, realizes, operations, citations)
	writeTestFile(t, designPath, final)
}

func TestValidateDesignSequenceGroundingRejectsUngroundedCitations(t *testing.T) {
	root := designTestProject(t)
	srcPath := filepath.Join(root, "src", "order_service.go")
	writeTestFile(t, srcPath, "package order\n\nfunc CreateOrder() {}\nfunc CancelOrder() {}\n")
	designPath := filepath.Join(root, "docs", "designs", "design-sequence", "dsg-dsd-order-design-sequence.md")

	writeTestFile(t, designPath, validDesignSequenceSource(`""`, `""`, "DSG-FUN-ORDER-SYSTEM-SEQUENCE",
		[]string{"CreateOrder"}, nil))
	d, err := parseDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDesignSequenceGrounding(d); err == nil || !strings.Contains(err.Error(), "no `' impl:") {
		t.Fatalf("expected missing-citation rejection, got %v", err)
	}

	writeTestFile(t, designPath, validDesignSequenceSource(`""`, `""`, "DSG-FUN-ORDER-SYSTEM-SEQUENCE",
		[]string{"CreateOrder"}, []string{"src/missing.go:1"}))
	d, err = parseDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDesignSequenceGrounding(d); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing-file rejection, got %v", err)
	}

	writeTestFile(t, designPath, validDesignSequenceSource(`""`, `""`, "DSG-FUN-ORDER-SYSTEM-SEQUENCE",
		[]string{"CreateOrder"}, []string{"src/order_service.go:999"}))
	d, err = parseDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDesignSequenceGrounding(d); err == nil || !strings.Contains(err.Error(), "exceeds file length") {
		t.Fatalf("expected out-of-range line rejection, got %v", err)
	}

	writeTestFile(t, designPath, validDesignSequenceSource(`""`, `""`, "DSG-FUN-ORDER-SYSTEM-SEQUENCE",
		[]string{"CreateOrder"}, []string{"src/order_service.go:3"}))
	d, err = parseDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDesignSequenceGrounding(d); err != nil {
		t.Fatalf("expected valid citation to pass grounding, got %v", err)
	}
}

func TestValidateDesignSequenceCompletenessRejectsMissingOperations(t *testing.T) {
	root := designTestProject(t)
	ssdPath := filepath.Join(root, "docs", "designs", "functional", "dsg-fun-order-system-sequence.md")
	writeTestFile(t, ssdPath, validSystemSequenceSource(`""`, `""`, []string{"CreateOrder", "CancelOrder"}))
	designPath := filepath.Join(root, "docs", "designs", "design-sequence", "dsg-dsd-order-design-sequence.md")

	writeTestFile(t, designPath, validDesignSequenceSource(`""`, `""`, "DSG-FUN-ORDER-SYSTEM-SEQUENCE",
		[]string{"CreateOrder"}, nil))
	d, err := parseDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	err = validateDesignSequenceCompleteness(d)
	if err == nil || !strings.Contains(err.Error(), "CancelOrder") {
		t.Fatalf("expected missing-operation rejection naming CancelOrder, got %v", err)
	}

	writeTestFile(t, designPath, validDesignSequenceSource(`""`, `""`, "DSG-FUN-ORDER-SYSTEM-SEQUENCE",
		[]string{"CreateOrder", "CancelOrder"}, nil))
	d, err = parseDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDesignSequenceCompleteness(d); err != nil {
		t.Fatalf("expected full operation coverage to pass, got %v", err)
	}
}

func TestAttestDesignSequenceEnforcesGroundingAndCompleteness(t *testing.T) {
	root := designTestProject(t)
	srcPath := filepath.Join(root, "src", "order_service.go")
	writeTestFile(t, srcPath, "package order\n\nfunc CreateOrder() {}\nfunc CancelOrder() {}\n")
	ssdPath := filepath.Join(root, "docs", "designs", "functional", "dsg-fun-order-system-sequence.md")
	writeTestFile(t, ssdPath, validSystemSequenceSource(`""`, `""`, []string{"CreateOrder", "CancelOrder"}))
	designPath := filepath.Join(root, "docs", "designs", "design-sequence", "dsg-dsd-order-design-sequence.md")

	// Ungrounded: covers both operations but cites nothing.
	writeAttestableDesign(t, root, designPath, "DSG-FUN-ORDER-SYSTEM-SEQUENCE",
		[]string{"CreateOrder", "CancelOrder"}, nil)
	if err := attestDesign(designPath); err == nil || !strings.Contains(err.Error(), "no `' impl:") {
		t.Fatalf("expected attest to refuse an ungrounded diagram, got %v", err)
	}

	// Grounded but incomplete: only cites/covers CreateOrder.
	writeAttestableDesign(t, root, designPath, "DSG-FUN-ORDER-SYSTEM-SEQUENCE",
		[]string{"CreateOrder"}, []string{"src/order_service.go:3"})
	if err := attestDesign(designPath); err == nil || !strings.Contains(err.Error(), "CancelOrder") {
		t.Fatalf("expected attest to refuse an incomplete diagram, got %v", err)
	}

	// Grounded and complete: attestation must succeed.
	writeAttestableDesign(t, root, designPath, "DSG-FUN-ORDER-SYSTEM-SEQUENCE",
		[]string{"CreateOrder", "CancelOrder"}, []string{"src/order_service.go:3", "src/order_service.go:4"})
	if err := attestDesign(designPath); err != nil {
		t.Fatalf("expected grounded, complete diagram to attest cleanly, got %v", err)
	}
	attested, err := parseDesign(designPath)
	if err != nil {
		t.Fatal(err)
	}
	if attested.Front.VisualValidation.Status != "passed" {
		t.Fatalf("expected passed attestation, got %q", attested.Front.VisualValidation.Status)
	}
}

func TestRenderAttestValidateDesignSequenceEndToEnd(t *testing.T) {
	major, err := installedNodeMajor()
	if err != nil || major < minimumNodeMajor {
		t.Skipf("Node %d+ is the mandatory consumer prerequisite: %v", minimumNodeMajor, err)
	}
	root := designTestProject(t)
	srcPath := filepath.Join(root, "src", "order_service.go")
	writeTestFile(t, srcPath, "package order\n\nfunc CreateOrder() {}\n")
	specPath := filepath.Join(root, "specs", "functional", "order.md")
	ssdPath := filepath.Join(root, "docs", "designs", "functional", "dsg-fun-order-system-sequence.md")
	dsdPath := filepath.Join(root, "docs", "designs", "design-sequence", "dsg-dsd-order-design-sequence.md")

	writeTestFile(t, specPath, `---
id: FUN-ORDER
status: implemented
created: 2026-08-07
---

# FUN-ORDER: Order Processing

## Design References

- [DSG-FUN-ORDER-SYSTEM-SEQUENCE](../../docs/designs/functional/dsg-fun-order-system-sequence.md) · [Image](../../docs/designs/functional/dsg-fun-order-system-sequence.png)
- [DSG-DSD-ORDER-DESIGN-SEQUENCE](../../docs/designs/design-sequence/dsg-dsd-order-design-sequence.md) · [Image](../../docs/designs/design-sequence/dsg-dsd-order-design-sequence.png)

## Acceptance Criteria

- **FUN-ORDER-001**: A customer can create an order.
`)
	writeTestFile(t, ssdPath, validSystemSequenceSource(`""`, `""`, []string{"CreateOrder"}))
	writeTestFile(t, dsdPath, validDesignSequenceSource(`""`, `""`, "DSG-FUN-ORDER-SYSTEM-SEQUENCE",
		[]string{"CreateOrder"}, []string{"src/order_service.go:3"}))

	if err := renderDesign(ssdPath); err != nil {
		t.Fatalf("render SSD: %v", err)
	}
	if err := attestDesign(ssdPath); err != nil {
		t.Fatalf("attest SSD: %v", err)
	}
	if err := renderDesign(dsdPath); err != nil {
		t.Fatalf("render DSD: %v", err)
	}
	if err := attestDesign(dsdPath); err != nil {
		t.Fatalf("attest DSD: %v", err)
	}
	if err := validateDesigns(dsdPath); err != nil {
		t.Fatalf("validate DSD: %v", err)
	}
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

func assertOpaqueCreamPNG(t *testing.T, content []byte) {
	t.Helper()
	decoded, err := png.Decode(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("decode rendered PNG: %v", err)
	}
	red, green, blue, alpha := decoded.At(decoded.Bounds().Min.X, decoded.Bounds().Min.Y).RGBA()
	if red != 0xffff || green != 0xf9f9 || blue != 0xf0f0 || alpha != 0xffff {
		t.Fatalf("rendered PNG canvas = (%#x, %#x, %#x, %#x), want opaque #FFF9F0", red, green, blue, alpha)
	}
	for y := decoded.Bounds().Min.Y; y < decoded.Bounds().Max.Y; y++ {
		for x := decoded.Bounds().Min.X; x < decoded.Bounds().Max.X; x++ {
			_, _, _, alpha := decoded.At(x, y).RGBA()
			if alpha != 0xffff {
				t.Fatalf("rendered PNG has non-opaque pixel at (%d, %d): alpha=%#x", x, y, alpha)
			}
		}
	}
}

func singleFileArchive(t *testing.T, path string) []byte {
	t.Helper()
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: path, Mode: 0o644, Size: 1}); err != nil {
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
	return archive.Bytes()
}
