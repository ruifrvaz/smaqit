package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

var (
	designIDPattern  = regexp.MustCompile(`^DSG-(BUS|FUN|STK|INF|COV)-[A-Z0-9]+(?:-[A-Z0-9]+)*$`)
	markdownLink     = regexp.MustCompile(`\[[^]]+\]\(([^)]+)\)`)
	unsafePlantUML   = regexp.MustCompile(`(?im)^\s*!(include|include_once|include_many|includeurl|import|pragma\s+include)\b`)
	requirementToken = regexp.MustCompile(`\b(?:BUS|FUN|STK|INF|COV)-[A-Z0-9]+(?:-[A-Z0-9]+)*\b`)
)

var designProfiles = map[string]map[string]bool{
	"business":       {"use-case": true},
	"functional":     {"system-sequence": true, "domain-model": true, "context-map": true, "state": true},
	"stack":          {"component": true},
	"infrastructure": {"deployment": true},
	"coverage":       {"requirement-trace": true},
}

var designLayerPrefix = map[string]string{
	"business": "BUS", "functional": "FUN", "stack": "STK", "infrastructure": "INF", "coverage": "COV",
}

var lifecycleRank = map[string]int{
	"draft": 0, "failed": 0, "implemented": 1, "deployed": 2, "validated": 3,
}

type visualValidation struct {
	Status       string `yaml:"status"`
	ValidatedAt  string `yaml:"validated_at"`
	SourceSHA256 string `yaml:"source_sha256"`
	ImageSHA256  string `yaml:"image_sha256"`
}

type designFrontmatter struct {
	ID               string           `yaml:"id"`
	Status           string           `yaml:"status"`
	Created          string           `yaml:"created"`
	Layer            string           `yaml:"layer"`
	DiagramType      string           `yaml:"diagram_type"`
	Notation         string           `yaml:"notation"`
	Specifications   []string         `yaml:"specifications"`
	Requirements     []string         `yaml:"requirements"`
	SourceSHA256     string           `yaml:"source_sha256"`
	ImageSHA256      string           `yaml:"image_sha256"`
	VisualValidation visualValidation `yaml:"visual_validation"`
}

type designArtifact struct {
	Path       string
	ImagePath  string
	Root       string
	Front      designFrontmatter
	FrontNode  yaml.Node
	Source     string
	SourceHash string
}

type plantUMLResult struct {
	Valid           bool     `json:"valid"`
	DiagramType     string   `json:"diagramType"`
	Warnings        []string `json:"warnings"`
	SVG             string   `json:"svg"`
	ErrorLineNumber int      `json:"errorLineNumber"`
	ErrorMessage    string   `json:"errorMessage"`
}

func cmdMCP(args []string) {
	root, err := findProjectRoot(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	switch {
	case len(args) == 1 && args[0] == "plantuml":
		err = runPlantUMLMCP(root)
	case len(args) == 1 && args[0] == "verify":
		err = verifyPlantUMLMCP(root)
	default:
		err = errors.New("Usage: smaqit mcp <plantuml|verify>")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// verifyPlantUMLMCP probes the same stdio wrapper declared in client config.
// It cannot prove a running host exposed the tools to an agent: Codex issue
// https://github.com/openai/codex/issues/30922 tracks that client-owned gap.
func verifyPlantUMLMCP(root string) error {
	if _, err := ensureDesignTools(root); err != nil {
		return err
	}
	if err := validateDesignMCPConfigs(root); err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: %w", err)
	}
	binary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: resolve smaqit executable: %w", err)
	}
	command := exec.Command(binary, "mcp", "plantuml")
	command.Dir = root
	client := mcp.NewClient(&mcp.Implementation{Name: "smaqit-mcp-verify", Version: Version}, nil)
	session, err := client.Connect(context.Background(), &mcp.CommandTransport{Command: command}, nil)
	if err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: PlantUML MCP wrapper did not initialize: %w", err)
	}
	defer session.Close()
	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: PlantUML MCP wrapper did not list tools: %w", err)
	}
	available := map[string]bool{}
	for _, tool := range tools.Tools {
		available[tool.Name] = true
	}
	for _, required := range []string{"check_syntax", "render_diagram"} {
		if !available[required] {
			return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: PlantUML MCP wrapper is missing required tool %q", required)
		}
	}
	if _, err := callPlantUML(session, "check_syntax", "@startuml\nAlice -> Bob: verify\n@enduml\n"); err != nil {
		return err
	}
	fmt.Println("✓ PlantUML MCP configuration and stdio transport are ready")
	fmt.Println("! Confirm the active client has loaded, trusted, and exposed the two MCP tools before authoring")
	return nil
}

func cmdDesign(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: smaqit design <render|attest|validate> [design.md]")
		os.Exit(1)
	}
	var err error
	switch args[0] {
	case "render":
		if len(args) != 2 {
			err = errors.New("Usage: smaqit design render <design.md>")
		} else {
			err = renderDesign(args[1])
		}
	case "attest":
		if len(args) != 2 {
			err = errors.New("Usage: smaqit design attest <design.md>")
		} else {
			err = attestDesign(args[1])
		}
	case "validate":
		if len(args) > 2 {
			err = errors.New("Usage: smaqit design validate [design.md]")
		} else {
			path := ""
			if len(args) == 2 {
				path = args[1]
			}
			err = validateDesigns(path)
		}
	default:
		err = fmt.Errorf("unknown design command %q", args[0])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseDesign(path string) (*designArtifact, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	root, err := findProjectRoot(filepath.Dir(absPath))
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("DESIGN-ARTIFACT-MISSING: %w", err)
	}
	if info, err := os.Lstat(absPath); err != nil || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("DESIGN-VISUAL-INVALID: design source must be a regular non-symlink file")
	}
	content := strings.ReplaceAll(string(b), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return nil, errors.New("DESIGN-VISUAL-INVALID: design must begin with YAML frontmatter")
	}
	closing := strings.Index(content[4:], "\n---\n")
	if closing < 0 {
		return nil, errors.New("DESIGN-VISUAL-INVALID: design frontmatter has no closing fence")
	}
	closing += 4
	frontText := content[4:closing]
	body := content[closing+5:]
	if !strings.HasPrefix(body, "\n```plantuml\n") || !strings.HasSuffix(body, "\n```\n") {
		return nil, errors.New("DESIGN-VISUAL-INVALID: design body must be exactly one PlantUML fence with no prose")
	}
	source := strings.TrimSuffix(strings.TrimPrefix(body, "\n```plantuml\n"), "\n```\n")
	if strings.Contains(source, "```") || strings.TrimSpace(source) == "" {
		return nil, errors.New("DESIGN-VISUAL-INVALID: design must contain exactly one non-empty PlantUML block")
	}
	source = normalizePlantUMLSource(source)
	if unsafePlantUML.MatchString(source) {
		return nil, errors.New("DESIGN-SYNTAX-INVALID: external PlantUML includes/imports are forbidden")
	}
	if strings.Count(strings.ToLower(source), "@start") != 1 || strings.Count(strings.ToLower(source), "@end") != 1 {
		return nil, errors.New("DESIGN-SYNTAX-INVALID: exactly one @start/@end diagram is required")
	}

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(frontText), &node); err != nil {
		return nil, fmt.Errorf("DESIGN-VISUAL-INVALID: invalid YAML: %w", err)
	}
	var front designFrontmatter
	if err := node.Decode(&front); err != nil {
		return nil, fmt.Errorf("DESIGN-VISUAL-INVALID: invalid metadata: %w", err)
	}
	if err := validateDesignMetadataKeys(&node); err != nil {
		return nil, err
	}
	artifact := &designArtifact{
		Path: absPath, ImagePath: strings.TrimSuffix(absPath, filepath.Ext(absPath)) + ".png",
		Root: root, Front: front, FrontNode: node, Source: source, SourceHash: hashBytes([]byte(source)),
	}
	if err := validateDesignMetadata(artifact); err != nil {
		return nil, err
	}
	return artifact, nil
}

func validateDesignMetadataKeys(node *yaml.Node) error {
	if len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode {
		return errors.New("DESIGN-VISUAL-INVALID: design frontmatter must be a mapping")
	}
	allowed := map[string]bool{
		"id": true, "status": true, "created": true, "layer": true, "diagram_type": true,
		"notation": true, "specifications": true, "requirements": true, "source_sha256": true,
		"image_sha256": true, "visual_validation": true,
	}
	mapping := node.Content[0]
	found := map[string]bool{}
	for i := 0; i < len(mapping.Content); i += 2 {
		key := mapping.Content[i].Value
		if !allowed[key] {
			return fmt.Errorf("DESIGN-VISUAL-INVALID: unsupported frontmatter field %q", key)
		}
		found[key] = true
		if key != "visual_validation" {
			continue
		}
		value := mapping.Content[i+1]
		if value.Kind != yaml.MappingNode {
			return errors.New("DESIGN-VISUAL-INVALID: visual_validation must be a mapping")
		}
		visualAllowed := map[string]bool{"status": true, "validated_at": true, "source_sha256": true, "image_sha256": true}
		visualFound := map[string]bool{}
		for j := 0; j < len(value.Content); j += 2 {
			visualKey := value.Content[j].Value
			if !visualAllowed[visualKey] {
				return fmt.Errorf("DESIGN-VISUAL-INVALID: unsupported visual_validation field %q", visualKey)
			}
			visualFound[visualKey] = true
		}
		for key := range visualAllowed {
			if !visualFound[key] {
				return fmt.Errorf("DESIGN-VISUAL-INVALID: missing visual_validation field %q", key)
			}
		}
	}
	for key := range allowed {
		if !found[key] {
			return fmt.Errorf("DESIGN-VISUAL-INVALID: missing frontmatter field %q", key)
		}
	}
	return nil
}

func validateDesignMetadata(d *designArtifact) error {
	f := d.Front
	if !designIDPattern.MatchString(f.ID) {
		return errors.New("DESIGN-VISUAL-INVALID: invalid design id")
	}
	if strings.ToLower(f.ID)+".md" != filepath.Base(d.Path) {
		return errors.New("DESIGN-VISUAL-INVALID: filename must be the lowercase design id")
	}
	prefix, layerExists := designLayerPrefix[f.Layer]
	if !layerExists || !strings.HasPrefix(f.ID, "DSG-"+prefix+"-") {
		return errors.New("DESIGN-VISUAL-INVALID: design id and layer disagree")
	}
	expectedDir := filepath.Join(d.Root, "docs", "designs", f.Layer)
	if filepath.Clean(filepath.Dir(d.Path)) != filepath.Clean(expectedDir) {
		return errors.New("DESIGN-VISUAL-INVALID: design is outside its declared layer directory")
	}
	if !designProfiles[f.Layer][f.DiagramType] {
		return fmt.Errorf("DESIGN-VISUAL-INVALID: diagram_type %q is not controlled for layer %q", f.DiagramType, f.Layer)
	}
	if f.DiagramType == "system-sequence" {
		if err := validateSystemSequenceProfile(d); err != nil {
			return err
		}
	}
	if f.Notation != "plantuml" || f.Created == "" {
		return errors.New("DESIGN-VISUAL-INVALID: required design metadata is missing")
	}
	if _, err := time.Parse(time.RFC3339, f.Created); err != nil {
		if _, dateErr := time.Parse("2006-01-02", f.Created); dateErr != nil {
			return errors.New("DESIGN-VISUAL-INVALID: created must be an ISO date or RFC3339 timestamp")
		}
	}
	if _, ok := lifecycleRank[f.Status]; !ok && f.Status != "deprecated" {
		return fmt.Errorf("DESIGN-VISUAL-INVALID: unsupported design status %q", f.Status)
	}
	if len(f.Specifications) == 0 || len(f.Requirements) == 0 {
		return errors.New("DESIGN-VISUAL-INVALID: specifications and requirements must be non-empty")
	}
	requirements := map[string]bool{}
	for _, requirement := range f.Requirements {
		if requirements[requirement] {
			return fmt.Errorf("DESIGN-VISUAL-INVALID: duplicate requirement reference %s", requirement)
		}
		requirements[requirement] = true
	}
	return nil
}

var (
	systemSequenceDecl     = regexp.MustCompile(`(?i)^(actor|participant|boundary|control|entity|database|collections|queue)\s+(.+)$`)
	systemSequenceAsAlias  = regexp.MustCompile(`(?i)^\s*as\s+("[^"]+"|[\w.]+)`)
	systemSequenceQuoted   = regexp.MustCompile(`^"([^"]*)"`)
	systemSequenceBareword = regexp.MustCompile(`^([\w.]+)`)
)

// systemSequenceIdentifier extracts a normalized participant identifier from
// a PlantUML token: a quoted label or a bareword.
func systemSequenceIdentifier(token string) (string, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}
	if m := systemSequenceQuoted.FindStringSubmatch(token); m != nil {
		return m[1], true
	}
	if m := systemSequenceBareword.FindStringSubmatch(token); m != nil {
		return m[1], true
	}
	return "", false
}

// systemSequenceDeclIdentifier extracts the effective identifier of an
// actor/participant-family declaration: the `as <alias>` alias if present,
// otherwise the quoted label or bareword name itself. Any trailing
// decoration (color, stereotype, `order N`) is ignored since it is only
// ever searched for within the tail after the label/bareword, never inside
// a quoted label itself.
func systemSequenceDeclIdentifier(rest string) string {
	rest = strings.TrimSpace(rest)
	var label, tail string
	if m := systemSequenceQuoted.FindStringSubmatch(rest); m != nil {
		label = m[1]
		tail = rest[len(m[0]):]
	} else if m := systemSequenceBareword.FindStringSubmatch(rest); m != nil {
		label = m[1]
		tail = rest[len(m[0]):]
	} else {
		return ""
	}
	if am := systemSequenceAsAlias.FindStringSubmatch(tail); am != nil {
		if alias, ok := systemSequenceIdentifier(am[1]); ok {
			return alias
		}
	}
	return label
}

// validateSystemSequenceProfile enforces the black-box System Sequence
// Diagram convention the `system-sequence` profile requires: exactly one
// actor and exactly one system-side participant, the latter always
// identified as `System` (e.g. `participant "<domain name>" as System`). It
// is a source-level scan over explicit actor/participant-family
// declarations only — no PlantUML grammar parser exists in this codebase.
// PlantUML also auto-creates participants purely from arrow references that
// are never explicitly declared; this check does not see those, so it is
// not adversarial-proof. It exists to catch the honest/default-authoring
// case: extra system-side collaborators declared outright, more than one
// actor, or a missing/misnamed System participant.
func validateSystemSequenceProfile(d *designArtifact) error {
	var actors, systemSide []string
	seenActor := map[string]bool{}
	seenSystem := map[string]bool{}

	inNote, inTitle, inLegend := false, false, false
	for _, raw := range strings.Split(d.Source, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "'") || strings.HasPrefix(line, "/'") {
			continue
		}
		if inNote {
			if strings.EqualFold(line, "end note") {
				inNote = false
			}
			continue
		}
		if inTitle {
			if strings.EqualFold(line, "end title") {
				inTitle = false
			}
			continue
		}
		if inLegend {
			if strings.EqualFold(line, "endlegend") || strings.EqualFold(line, "end legend") {
				inLegend = false
			}
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "note") {
			if !strings.Contains(line, ":") {
				inNote = true
			}
			continue
		}
		if lower == "title" {
			inTitle = true
			continue
		}
		if strings.HasPrefix(lower, "title ") {
			continue
		}
		if strings.HasPrefix(lower, "legend") {
			inLegend = true
			continue
		}

		m := systemSequenceDecl.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		keyword := strings.ToLower(m[1])
		name := systemSequenceDeclIdentifier(m[2])
		if name == "" {
			continue
		}
		if keyword == "actor" {
			if !seenActor[name] {
				seenActor[name] = true
				actors = append(actors, name)
			}
			continue
		}
		if !seenSystem[name] {
			seenSystem[name] = true
			systemSide = append(systemSide, name)
		}
	}

	switch len(actors) {
	case 0:
		return errors.New("DESIGN-VISUAL-INVALID: system-sequence diagrams must declare exactly one actor, found none")
	case 1:
	default:
		return fmt.Errorf(
			"DESIGN-VISUAL-INVALID: system-sequence diagrams must model exactly one actor, found %d (%s) — author one system-sequence design per actor instead",
			len(actors), strings.Join(actors, ", "),
		)
	}

	switch len(systemSide) {
	case 0:
		return errors.New(`DESIGN-VISUAL-INVALID: system-sequence diagrams must declare exactly one system participant identified as "System" (e.g. participant "<name>" as System), found none`)
	case 1:
	default:
		return fmt.Errorf(
			`DESIGN-VISUAL-INVALID: system-sequence diagrams must model exactly one system participant, found %d (%s) — split multi-flow behavior into separate per-flow system-sequence designs instead`,
			len(systemSide), strings.Join(systemSide, ", "),
		)
	}

	if systemSide[0] != "System" {
		return fmt.Errorf(
			`DESIGN-VISUAL-INVALID: system-sequence diagrams must identify their system participant as "System" (e.g. participant "<name>" as System), found %q instead`,
			systemSide[0],
		)
	}
	return nil
}

func normalizePlantUMLSource(source string) string {
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	return strings.TrimRight(source, "\n") + "\n"
}

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func resolvedProjectPath(root, origin, reference string) (string, error) {
	if reference == "" || filepath.IsAbs(reference) {
		return "", errors.New("path must be a non-empty relative path")
	}
	clean := filepath.Clean(filepath.Join(filepath.Dir(origin), filepath.FromSlash(reference)))
	rel, err := filepath.Rel(root, clean)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes project: %s", reference)
	}
	for current := clean; current != root; current = filepath.Dir(current) {
		if info, err := os.Lstat(current); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlink paths are forbidden: %s", reference)
		}
	}
	return clean, nil
}

func validateDesignArtifact(d *designArtifact) error {
	imageBytes, err := os.ReadFile(d.ImagePath)
	if err != nil {
		return fmt.Errorf("DESIGN-ARTIFACT-MISSING: PNG for %s is missing", filepath.Base(d.Path))
	}
	if info, err := os.Lstat(d.ImagePath); err != nil || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("DESIGN-VISUAL-INVALID: PNG for %s must be a regular non-symlink file", filepath.Base(d.Path))
	}
	config, err := png.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil || config.Width < 640 || config.Width > 4096 || config.Height < 1 || config.Height > 16384 {
		return fmt.Errorf("DESIGN-VISUAL-INVALID: invalid PNG for %s", filepath.Base(d.Path))
	}
	imageHash := hashBytes(imageBytes)
	if d.Front.SourceSHA256 != d.SourceHash || d.Front.ImageSHA256 != imageHash {
		return fmt.Errorf("DESIGN-ARTIFACT-STALE: current source/image hashes do not match %s", filepath.Base(d.Path))
	}
	v := d.Front.VisualValidation
	if v.Status != "passed" || v.ValidatedAt == "" || v.SourceSHA256 != d.SourceHash || v.ImageSHA256 != imageHash {
		return fmt.Errorf("DESIGN-VISUAL-INVALID: current PNG has no passing visual attestation for %s", filepath.Base(d.Path))
	}
	if _, err := time.Parse(time.RFC3339, v.ValidatedAt); err != nil {
		return fmt.Errorf("DESIGN-VISUAL-INVALID: visual attestation timestamp is invalid for %s", filepath.Base(d.Path))
	}
	return validateDesignReferences(d)
}

func validateDesignReferences(d *designArtifact) error {
	seen := map[string]bool{}
	minimumLinkedRank := 4
	for _, ref := range d.Front.Specifications {
		specPath, err := resolvedProjectPath(d.Root, d.Path, ref)
		if err != nil {
			return fmt.Errorf("DESIGN-VISUAL-INVALID: unsafe specification reference: %w", err)
		}
		expected := filepath.Join(d.Root, "specs", d.Front.Layer)
		if filepath.Dir(specPath) != expected || filepath.Ext(specPath) != ".md" {
			return errors.New("DESIGN-VISUAL-INVALID: design references a specification outside its layer")
		}
		if seen[specPath] {
			return errors.New("DESIGN-VISUAL-INVALID: duplicate specification reference")
		}
		seen[specPath] = true
		content, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("DESIGN-ARTIFACT-MISSING: referenced specification %s", ref)
		}
		if !specReferencesDesign(d.Root, specPath, content, d.Path, d.ImagePath) {
			return fmt.Errorf("DESIGN-VISUAL-INVALID: specification %s does not link the design pair", filepath.Base(specPath))
		}
		text := string(content)
		specFront, err := parseSpecFrontmatter(specPath)
		if err != nil {
			return fmt.Errorf("DESIGN-VISUAL-INVALID: invalid linked specification metadata: %w", err)
		}
		if specFront.Status != "deprecated" {
			rank, exists := lifecycleRank[specFront.Status]
			if !exists {
				return fmt.Errorf("DESIGN-VISUAL-INVALID: unsupported linked specification status %q", specFront.Status)
			}
			if rank < minimumLinkedRank {
				minimumLinkedRank = rank
			}
		}
		for _, requirement := range d.Front.Requirements {
			if !strings.Contains(text, requirement) {
				continue
			}
			seen[requirement] = true
		}
	}
	for _, requirement := range d.Front.Requirements {
		if !requirementToken.MatchString(requirement) || !strings.HasPrefix(requirement, designLayerPrefix[d.Front.Layer]+"-") || !seen[requirement] {
			return fmt.Errorf("DESIGN-VISUAL-INVALID: requirement %s does not exist in a linked specification", requirement)
		}
	}
	if d.Front.Status != "deprecated" && minimumLinkedRank < 4 {
		designRank := lifecycleRank[d.Front.Status]
		if designRank != minimumLinkedRank {
			return fmt.Errorf("DESIGN-ARTIFACT-STALE: design lifecycle rank %d must equal least-advanced linked specification rank %d", designRank, minimumLinkedRank)
		}
	}
	return nil
}

func specReferencesDesign(root, specPath string, content []byte, designPath, imagePath string) bool {
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	marker := strings.Index(text, "## Design References\n")
	if marker < 0 {
		return false
	}
	section := text[marker+len("## Design References\n"):]
	if next := strings.Index(section, "\n## "); next >= 0 {
		section = section[:next]
	}
	wantDesign, _ := filepath.Rel(filepath.Dir(specPath), designPath)
	wantImage, _ := filepath.Rel(filepath.Dir(specPath), imagePath)
	foundDesign, foundImage := false, false
	for _, match := range markdownLink.FindAllStringSubmatch(section, -1) {
		resolved, err := resolvedProjectPath(root, specPath, match[1])
		if err != nil {
			continue
		}
		foundDesign = foundDesign || resolved == filepath.Clean(designPath) || filepath.Clean(match[1]) == filepath.Clean(wantDesign)
		foundImage = foundImage || resolved == filepath.Clean(imagePath) || filepath.Clean(match[1]) == filepath.Clean(wantImage)
	}
	return foundDesign && foundImage
}

func specDesignReady(specPath, layer string) (bool, string) {
	absSpecPath, err := filepath.Abs(specPath)
	if err != nil {
		return false, err.Error()
	}
	specPath = absSpecPath
	root, err := findProjectRoot(filepath.Dir(specPath))
	if err != nil {
		return false, err.Error()
	}
	content, err := os.ReadFile(specPath)
	if err != nil {
		return false, err.Error()
	}
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	marker := strings.Index(text, "## Design References\n")
	if marker < 0 {
		return false, "DESIGN-ARTIFACT-MISSING: no Design References section"
	}
	section := text[marker+len("## Design References\n"):]
	if next := strings.Index(section, "\n## "); next >= 0 {
		section = section[:next]
	}
	links := markdownLink.FindAllStringSubmatch(section, -1)
	if len(links) == 0 {
		return false, "DESIGN-ARTIFACT-MISSING: no linked design"
	}
	seenPair := false
	for _, link := range links {
		if filepath.Ext(link[1]) != ".md" {
			continue
		}
		designPath, err := resolvedProjectPath(root, specPath, link[1])
		if err != nil {
			return false, "DESIGN-VISUAL-INVALID: " + err.Error()
		}
		d, err := parseDesign(designPath)
		if err != nil {
			return false, err.Error()
		}
		if d.Front.Layer != layer {
			return false, "DESIGN-VISUAL-INVALID: linked design belongs to another layer"
		}
		if d.Front.Status == "deprecated" {
			return false, "DESIGN-ARTIFACT-MISSING: active specification links only a deprecated design"
		}
		if err := validateDesignArtifact(d); err != nil {
			return false, err.Error()
		}
		seenPair = true
	}
	if !seenPair {
		return false, "DESIGN-ARTIFACT-MISSING: no canonical Markdown/PNG design pair"
	}
	return true, ""
}

func validateDesigns(path string) error {
	rootStart := "."
	if path != "" {
		rootStart = filepath.Dir(path)
	}
	root, err := findProjectRoot(rootStart)
	if err != nil {
		return err
	}
	var paths []string
	if path != "" {
		paths = []string{path}
	} else {
		err = filepath.WalkDir(filepath.Join(root, "docs", "designs"), func(p string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
				paths = append(paths, p)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	sort.Strings(paths)
	ids := map[string]bool{}
	var syntaxSession *mcp.ClientSession
	if len(paths) > 0 {
		syntaxSession, err = openPlantUMLSession(root)
		if err != nil {
			return err
		}
		defer syntaxSession.Close()
	}
	var diagnostics []string
	for _, p := range paths {
		d, err := parseDesign(p)
		if err != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("%s: %v", filepath.Base(p), err))
			continue
		}
		if ids[d.Front.ID] {
			diagnostics = append(diagnostics, fmt.Sprintf("%s: DESIGN-VISUAL-INVALID: duplicate design id %s", filepath.Base(p), d.Front.ID))
		} else {
			ids[d.Front.ID] = true
		}
		if d.Front.Status == "deprecated" {
			continue
		}
		if _, err := callPlantUML(syntaxSession, "check_syntax", d.Source); err != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("%s: %v", filepath.Base(p), err))
		}
		if err := validateDesignArtifact(d); err != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("%s: %v", filepath.Base(p), err))
		}
	}
	if path == "" {
		diagnostics = append(diagnostics, collectActiveSpecDesignDiagnostics(root)...)
	}
	if len(diagnostics) > 0 {
		return fmt.Errorf("DESIGN-VALIDATION-FAILED: %d issue(s):\n%s", len(diagnostics), strings.Join(diagnostics, "\n"))
	}
	fmt.Printf("✓ Validated %d canonical design(s)\n", len(paths))
	return nil
}

func collectActiveSpecDesignDiagnostics(root string) []string {
	var diagnostics []string
	for _, layer := range []string{"business", "functional", "stack", "infrastructure", "coverage"} {
		entries, err := os.ReadDir(filepath.Join(root, "specs", layer))
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
				continue
			}
			path := filepath.Join(root, "specs", layer, entry.Name())
			front, err := parseSpecFrontmatter(path)
			if err != nil {
				diagnostics = append(diagnostics, fmt.Sprintf("%s: DESIGN-VISUAL-INVALID: invalid linked specification metadata: %v", filepath.Base(path), err))
				continue
			}
			if front.Status == "deprecated" {
				continue
			}
			ready, reason := specDesignReady(path, layer)
			if !ready {
				diagnostics = append(diagnostics, fmt.Sprintf("active specification %s: %s", filepath.Base(path), reason))
			}
		}
	}
	return diagnostics
}

func openPlantUMLSession(root string) (*mcp.ClientSession, error) {
	runtimeDir, err := ensureDesignTools(root)
	if err != nil {
		return nil, err
	}
	command := exec.Command(nodeCommandName(), filepath.Join(runtimeDir, "node_modules", "@plantuml", "mcp-js", "server.js"))
	command.Dir = root
	client := mcp.NewClient(&mcp.Implementation{Name: "smaqit", Version: Version}, nil)
	session, err := client.Connect(context.Background(), &mcp.CommandTransport{Command: command}, nil)
	if err != nil {
		return nil, fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: %w", err)
	}
	return session, nil
}

func callPlantUML(session *mcp.ClientSession, tool, source string) (*plantUMLResult, error) {
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: map[string]any{"source": source}})
	if err != nil {
		return nil, err
	}
	var text string
	for _, content := range result.Content {
		if value, ok := content.(*mcp.TextContent); ok {
			text += value.Text
		}
	}
	var parsed plantUMLResult
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("PlantUML MCP returned invalid JSON: %w", err)
	}
	if result.IsError || !parsed.Valid {
		return nil, fmt.Errorf("DESIGN-SYNTAX-INVALID: line %d: %s", parsed.ErrorLineNumber, parsed.ErrorMessage)
	}
	return &parsed, nil
}

func renderDesign(path string) error {
	d, err := parseDesign(path)
	if err != nil {
		return err
	}
	session, err := openPlantUMLSession(d.Root)
	if err != nil {
		return err
	}
	defer session.Close()
	if _, err := callPlantUML(session, "check_syntax", d.Source); err != nil {
		return err
	}
	rendered, err := callPlantUML(session, "render_diagram", d.Source)
	if err != nil {
		return err
	}
	if rendered.SVG == "" {
		return errors.New("DESIGN-TOOLCHAIN-UNAVAILABLE: PlantUML MCP returned no SVG")
	}
	tempDir, err := os.MkdirTemp(filepath.Dir(d.ImagePath), ".smaqit-render-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	svgPath := filepath.Join(tempDir, "diagram.svg")
	pngPath := filepath.Join(tempDir, "diagram.png")
	if err := os.WriteFile(svgPath, []byte(rendered.SVG), 0o600); err != nil {
		return err
	}
	command := exec.Command(nodeCommandName(), designRuntimePath(d.Root, "render-png.mjs"), svgPath, pngPath, "1800")
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: SVG-to-PNG conversion failed: %v: %s", err, strings.TrimSpace(string(output)))
	}
	imageBytes, err := os.ReadFile(pngPath)
	if err != nil {
		return err
	}
	if _, err := png.DecodeConfig(bytes.NewReader(imageBytes)); err != nil {
		return fmt.Errorf("DESIGN-TOOLCHAIN-UNAVAILABLE: converter produced invalid PNG: %w", err)
	}
	if err := atomicWriteFile(d.ImagePath, imageBytes, 0o644); err != nil {
		return err
	}
	d.Front.SourceSHA256 = d.SourceHash
	d.Front.ImageSHA256 = hashBytes(imageBytes)
	d.Front.VisualValidation = visualValidation{Status: "pending"}
	if err := writeDesignFrontmatter(d); err != nil {
		return err
	}
	fmt.Printf("✓ Rendered %s\n", d.ImagePath)
	return nil
}

func attestDesign(path string) error {
	d, err := parseDesign(path)
	if err != nil {
		return err
	}
	imageBytes, err := os.ReadFile(d.ImagePath)
	if err != nil {
		return fmt.Errorf("DESIGN-ARTIFACT-MISSING: open the rendered PNG before attesting: %w", err)
	}
	if _, err := png.DecodeConfig(bytes.NewReader(imageBytes)); err != nil {
		return fmt.Errorf("DESIGN-VISUAL-INVALID: %w", err)
	}
	imageHash := hashBytes(imageBytes)
	if d.Front.SourceSHA256 != d.SourceHash || d.Front.ImageSHA256 != imageHash {
		return errors.New("DESIGN-ARTIFACT-STALE: render the current design before attesting")
	}
	d.Front.VisualValidation = visualValidation{
		Status: "passed", ValidatedAt: time.Now().UTC().Format(time.RFC3339), SourceSHA256: d.SourceHash, ImageSHA256: imageHash,
	}
	if err := writeDesignFrontmatter(d); err != nil {
		return err
	}
	fmt.Printf("✓ Recorded visual attestation for %s\n", d.Path)
	return nil
}

func writeDesignFrontmatter(d *designArtifact) error {
	front, err := yaml.Marshal(d.Front)
	if err != nil {
		return err
	}
	content := append([]byte("---\n"), front...)
	content = append(content, []byte("---\n\n```plantuml\n"+d.Source+"```\n")...)
	return atomicWriteFile(d.Path, content, 0o644)
}

func atomicWriteFile(path string, content []byte, mode os.FileMode) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".smaqit-write-")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(mode); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(content); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Rename(tempName, path)
	} else if err != nil {
		return err
	}
	backupFile, err := os.CreateTemp(filepath.Dir(path), ".smaqit-backup-")
	if err != nil {
		return err
	}
	backupName := backupFile.Name()
	defer os.Remove(backupName)
	if err := backupFile.Close(); err != nil {
		return err
	}
	if err := os.Remove(backupName); err != nil {
		return err
	}
	if err := os.Rename(path, backupName); err != nil {
		return err
	}
	if err := os.Rename(tempName, path); err != nil {
		_ = os.Rename(backupName, path)
		return err
	}
	return os.Remove(backupName)
}
