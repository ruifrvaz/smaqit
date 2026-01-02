package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed templates/specs/*.md
var templateFiles embed.FS

//go:embed agents/*.md
var agentFiles embed.FS

//go:embed prompts/*.md
var promptFiles embed.FS

// Version is set via ldflags during build: -X main.Version=$(VERSION)
var Version = "dev"

// PhaseState represents the completion state of a phase
type PhaseState struct {
	Completed bool   `json:"completed"`
	Timestamp string `json:"timestamp,omitempty"`
}

// Phases represents the ordered phases structure
type Phases struct {
	Develop  PhaseState `json:"develop"`
	Deploy   PhaseState `json:"deploy"`
	Validate PhaseState `json:"validate"`
}

// StateFile represents the .smaqit/state.json structure
type StateFile struct {
	Version string `json:"version"`
	Phases  Phases `json:"phases"`
}

// initStateFile creates a new state.json with all phases marked incomplete
func initStateFile() StateFile {
	return StateFile{
		Version: "1.0",
		Phases: Phases{
			Develop:  PhaseState{Completed: false},
			Deploy:   PhaseState{Completed: false},
			Validate: PhaseState{Completed: false},
		},
	}
}

// readStateFile reads and validates state.json, returning default state on error
func readStateFile(path string) StateFile {
	content, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist or can't be read - return default
		return initStateFile()
	}

	var state StateFile
	if err := json.Unmarshal(content, &state); err != nil {
		// Corrupted JSON - warn and return default
		fmt.Println("⚠ Warning: state.json is corrupted, using default state")
		return initStateFile()
	}

	// Validate schema
	if state.Version == "" {
		fmt.Println("⚠ Warning: state.json has invalid schema, using default state")
		return initStateFile()
	}

	return state
}

// writeStateFile writes state.json using atomic write pattern (temp + rename)
func writeStateFile(path string, state StateFile) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	// Write to temporary file
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath) // Clean up temp file on failure
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		targetDir := "."
		if len(os.Args) > 2 {
			targetDir = os.Args[2]
		}
		cmdInit(targetDir)
	case "status":
		cmdStatus()
	case "validate":
		cmdValidate()
	case "help", "--help", "-h":
		cmdHelp()
	case "uninstall":
		cmdUninstall()
	case "version", "--version", "-v":
		fmt.Printf("smaqit %s\n", Version)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`smaqit - Spec-Driven Agent Orchestration Kit

Usage: smaqit <command>

Commands:
  init [dir] Scaffold .smaqit/ and .github/ directories
             Optional: specify target directory (default: current)
  status     Show project state and spec coverage
  validate   Verify project structure integrity
  help       Show detailed command help
  uninstall  Remove smaqit from project
  version    Show smaqit version`)
}

func cmdHelp() {
	fmt.Println("smaqit - Spec-Driven Agent Orchestration Kit")
	fmt.Printf("Version: %s\n\n", Version)

	fmt.Println("CLI Commands:")
	fmt.Println("  smaqit init [dir] Scaffold smaqit project structure")
	fmt.Println("                    Creates .smaqit/ and .github/ directories with")
	fmt.Println("                    framework files, templates, and agent definitions")
	fmt.Println("                    Optional: specify target directory (created if needed)")
	fmt.Println()
	fmt.Println("  smaqit status     Show project state and spec coverage")
	fmt.Println("                    Reports number of specs per layer and phase status")
	fmt.Println()
	fmt.Println("  smaqit validate   Verify project structure integrity")
	fmt.Println("                    Checks directory structure, framework files, and")
	fmt.Println("                    validates spec template compliance")
	fmt.Println()
	fmt.Println("  smaqit help       Show this help message")
	fmt.Println()
	fmt.Println("  smaqit uninstall  Remove smaqit from project")
	fmt.Println("                    Removes .smaqit/, .github/agents/, .github/prompts/")
	fmt.Println()
	fmt.Println("  smaqit version    Show smaqit version")
	fmt.Println()
	fmt.Println("Copilot Prompts (use in GitHub Copilot chat with /):")
	fmt.Println("  /smaqit.development   Run Development implementation agent (build from specs)")
	fmt.Println("  /smaqit.deployment    Run Deployment implementation agent (deploy from specs)")
	fmt.Println("  /smaqit.validation    Run Validation implementation agent (test from specs)")
	fmt.Println("  /smaqit.business      Create business layer specifications")
	fmt.Println("  /smaqit.functional    Create functional layer specifications")
	fmt.Println("  /smaqit.stack         Create stack layer specifications")
	fmt.Println("  /smaqit.infrastructure Create infrastructure layer specifications")
	fmt.Println("  /smaqit.coverage      Create coverage layer specifications")
	fmt.Println()
	fmt.Println("Getting Started:")
	fmt.Println("  1. Run 'smaqit init' in your project directory")
	fmt.Println("  2. Open GitHub Copilot chat in VS Code")
	fmt.Println("  3. Type '/smaqit.development' to run the Development implementation step")
	fmt.Println()
	fmt.Println("Documentation: https://github.com/ruifrvaz/smaqit")
}

func cmdInit(targetDir string) {
	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", targetDir, err)
		os.Exit(1)
	}

	// Change to target directory
	if err := os.Chdir(targetDir); err != nil {
		fmt.Printf("Error changing to directory %s: %v\n", targetDir, err)
		os.Exit(1)
	}

	// Check if .smaqit already exists
	if _, err := os.Stat(".smaqit"); err == nil {
		fmt.Println("Error: .smaqit/ directory already exists")
		fmt.Println("Run 'smaqit uninstall' first to remove existing installation")
		os.Exit(1)
	}

	fmt.Printf("Initializing smaqit project in %s...\n", targetDir)

	// Create directory structure
	dirs := []string{
		".smaqit/templates/specs",
		".smaqit/reports",
		"specs/business",
		"specs/functional",
		"specs/stack",
		"specs/infrastructure",
		"specs/coverage",
		".github/agents",
		".github/prompts",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Copy spec templates
	if err := copyEmbeddedDir(templateFiles, "templates/specs", ".smaqit/templates/specs"); err != nil {
		fmt.Printf("Error copying spec templates: %v\n", err)
		os.Exit(1)
	}

	// Copy agent files
	if err := copyEmbeddedDir(agentFiles, "agents", ".github/agents"); err != nil {
		fmt.Printf("Error copying agent files: %v\n", err)
		os.Exit(1)
	}

	// Copy prompt files
	if err := copyEmbeddedDir(promptFiles, "prompts", ".github/prompts"); err != nil {
		fmt.Printf("Error copying prompt files: %v\n", err)
		os.Exit(1)
	}

	// Initialize state.json
	stateFilePath := filepath.Join(".smaqit", "state.json")
	initialState := initStateFile()
	if err := writeStateFile(stateFilePath, initialState); err != nil {
		fmt.Printf("Error writing state.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Created .smaqit/ directory structure")
	fmt.Println("✓ Copied templates")
	fmt.Println("✓ Copied agent definitions")
	fmt.Println("✓ Copied prompt files")
	fmt.Printf("✓ Initialized smaqit %s\n\n", Version)
	fmt.Println("Next steps:")
	fmt.Println("  1. Open GitHub Copilot chat in VS Code")
	fmt.Println("  2. Type '/smaqit.development' to run the Development implementation step")
	fmt.Println("  3. Or type '/smaqit.business' to create business specifications")
}

// copyEmbeddedDir copies files from an embedded FS to a target directory
// If dstDir contains "templates/specs", performs version substitution
func copyEmbeddedDir(embeddedFS embed.FS, srcDir, dstDir string) error {
	substituteVersion := strings.Contains(dstDir, "templates/specs")
	
	return fs.WalkDir(embeddedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Read embedded file
		content, err := embeddedFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		// Perform version substitution for spec templates
		if substituteVersion {
			contentStr := string(content)
			contentStr = strings.ReplaceAll(contentStr, "[SMAQIT_VERSION]", Version)
			content = []byte(contentStr)
		}

		// Calculate destination path
		relPath := strings.TrimPrefix(path, srcDir+"/")
		dstPath := filepath.Join(dstDir, relPath)

		// Ensure destination directory exists
		dstFileDir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstFileDir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dstFileDir, err)
		}

		// Write file
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", dstPath, err)
		}

		return nil
	})
}

// validateStateFile performs comprehensive validation of state.json structure
// Returns the count of errors found
func validateStateFile(path string) int {
	errors := 0

	// Check file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("✗ Missing file: %s\n", path)
		fmt.Println("  → Run 'smaqit init' to create state.json")
		return 1
	}

	// Read and parse JSON
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("✗ Cannot read %s: %v\n", path, err)
		fmt.Println("  → Check file permissions")
		return 1
	}

	// Parse JSON structure
	var rawState map[string]interface{}
	if err := json.Unmarshal(content, &rawState); err != nil {
		fmt.Printf("✗ Invalid JSON in %s: %v\n", path, err)
		fmt.Println("  → Fix JSON syntax or regenerate with 'smaqit init' (after backup)")
		return 1
	}

	// Validate version field
	version, versionExists := rawState["version"]
	if !versionExists {
		fmt.Printf("✗ Missing 'version' field in %s\n", path)
		fmt.Println("  → Add: \"version\": \"1.0\"")
		errors++
	} else if versionStr, ok := version.(string); !ok {
		fmt.Printf("✗ Invalid 'version' field type in %s (expected string)\n", path)
		fmt.Println("  → Change to: \"version\": \"1.0\"")
		errors++
	} else if versionStr == "" {
		fmt.Printf("✗ Empty 'version' field in %s\n", path)
		fmt.Println("  → Set to: \"version\": \"1.0\"")
		errors++
	}

	// Validate phases map exists
	phasesRaw, phasesExists := rawState["phases"]
	if !phasesExists {
		fmt.Printf("✗ Missing 'phases' object in %s\n", path)
		fmt.Println("  → Add phases object with develop, deploy, validate keys")
		return errors + 1
	}

	phasesMap, ok := phasesRaw.(map[string]interface{})
	if !ok {
		fmt.Printf("✗ Invalid 'phases' field type in %s (expected object)\n", path)
		fmt.Println("  → Change to object containing phase states")
		return errors + 1
	}

	// Check phase keys exist and are ordered correctly
	requiredPhases := []string{"develop", "deploy", "validate"}
	
	// Check each required phase exists
	for _, phase := range requiredPhases {
		phaseData, exists := phasesMap[phase]
		if !exists {
			fmt.Printf("✗ Missing phase '%s' in %s\n", phase, path)
			fmt.Printf("  → Add: \"%s\": {\"completed\": false}\n", phase)
			errors++
			continue
		}

		// Validate phase object structure
		phaseObj, ok := phaseData.(map[string]interface{})
		if !ok {
			fmt.Printf("✗ Phase '%s' is not an object in %s\n", phase, path)
			fmt.Printf("  → Change to: \"%s\": {\"completed\": false}\n", phase)
			errors++
			continue
		}

		// Check 'completed' field exists
		completed, completedExists := phaseObj["completed"]
		if !completedExists {
			fmt.Printf("✗ Phase '%s' missing 'completed' field in %s\n", phase, path)
			fmt.Printf("  → Add 'completed' boolean to %s phase\n", phase)
			errors++
		} else if _, ok := completed.(bool); !ok {
			fmt.Printf("✗ Phase '%s' has invalid 'completed' type in %s (expected boolean)\n", phase, path)
			fmt.Printf("  → Change 'completed' to true or false in %s phase\n", phase)
			errors++
		}
	}

	// Check phase ordering in JSON
	// Parse JSON preserving order by examining raw bytes
	if err := validatePhaseOrdering(content, path); err != nil {
		fmt.Printf("⚠ Phase ordering in %s: %v\n", path, err)
		fmt.Println("  → Expected order: develop, deploy, validate")
		fmt.Println("  → This is a semantic issue; functionality is not affected")
		// Warning only, not an error
	}

	if errors == 0 {
		fmt.Printf("✓ State file %s is valid\n", path)
	}

	return errors
}

// validatePhaseOrdering checks if phases appear in correct order in JSON text
// This is a semantic check for readability, not a functional requirement
func validatePhaseOrdering(jsonContent []byte, path string) error {
	content := string(jsonContent)
	
	// Find position of each phase key in JSON text
	developPos := strings.Index(content, "\"develop\"")
	deployPos := strings.Index(content, "\"deploy\"")
	validatePos := strings.Index(content, "\"validate\"")
	
	// If any phase is missing, ordering check is irrelevant (will be caught by earlier checks)
	if developPos == -1 || deployPos == -1 || validatePos == -1 {
		return nil
	}
	
	// Check ordering: develop < deploy < validate
	if developPos > deployPos || deployPos > validatePos {
		return fmt.Errorf("phases not in workflow order (develop → deploy → validate)")
	}
	
	return nil
}

func cmdUninstall() {
	// Check if .smaqit exists
	if _, err := os.Stat(".smaqit"); os.IsNotExist(err) {
		fmt.Println("No smaqit installation found in this directory")
		os.Exit(0)
	}

	// Prompt for confirmation
	fmt.Println("This will remove:")
	fmt.Println("  • .smaqit/")
	fmt.Println("  • .github/agents/")
	fmt.Println("  • .github/prompts/")
	fmt.Print("\nContinue? [y/N]: ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != "yes" {
		fmt.Println("Uninstall cancelled")
		os.Exit(0)
	}

	// Ask about specs directory
	removeSpecs := false
	if _, err := os.Stat("specs"); err == nil {
		fmt.Print("\nAlso remove specs/ directory (contains your specifications)? [y/N]: ")
		var specsResponse string
		fmt.Scanln(&specsResponse)
		specsResponse = strings.ToLower(strings.TrimSpace(specsResponse))
		removeSpecs = (specsResponse == "y" || specsResponse == "yes")
	}

	// Remove directories
	errors := 0

	if err := os.RemoveAll(".smaqit"); err != nil {
		fmt.Printf("Error removing .smaqit/: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ Removed .smaqit/")
	}

	if removeSpecs {
		if err := os.RemoveAll("specs"); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Error removing specs/: %v\n", err)
			errors++
		} else {
			fmt.Println("✓ Removed specs/")
		}
	} else {
		fmt.Println("✓ Kept specs/ (user specifications)")
	}

	if err := os.RemoveAll(filepath.Join(".github", "agents")); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error removing .github/agents/: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ Removed .github/agents/")
	}

	if err := os.RemoveAll(filepath.Join(".github", "prompts")); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error removing .github/prompts/: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ Removed .github/prompts/")
	}

	// Check if .github is empty and remove it
	entries, err := os.ReadDir(".github")
	if err == nil && len(entries) == 0 {
		if err := os.Remove(".github"); err == nil {
			fmt.Println("✓ Removed empty .github/")
		}
	}

	if errors > 0 {
		fmt.Printf("\nUninstall completed with %d error(s)\n", errors)
		os.Exit(1)
	} else {
		fmt.Println("\n✓ Uninstall complete")
	}
}

func cmdValidate() {
	// Check if .smaqit exists
	if _, err := os.Stat(".smaqit"); os.IsNotExist(err) {
		fmt.Println("Error: .smaqit/ directory not found")
		fmt.Println("Run 'smaqit init' to initialize smaqit in this project")
		os.Exit(1)
	}

	fmt.Println("Validating smaqit project structure...")
	errors := 0

	// Check directory structure
	requiredDirs := []string{
		".smaqit/templates/specs",
		"specs/business",
		"specs/functional",
		"specs/stack",
		"specs/infrastructure",
		"specs/coverage",
		".github/agents",
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Printf("✗ Missing directory: %s\n", dir)
			errors++
		}
	}

	if errors == 0 {
		fmt.Println("✓ Directory structure is complete")
	}

	// Validate state.json
	stateFilePath := filepath.Join(".smaqit", "state.json")
	stateErrors := validateStateFile(stateFilePath)
	errors += stateErrors

	// Validate spec files (basic checks)
	layers := []string{"business", "functional", "stack", "infrastructure", "coverage"}
	for _, layer := range layers {
		specDir := filepath.Join("specs", layer)
		entries, err := os.ReadDir(specDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			specPath := filepath.Join(specDir, entry.Name())
			content, err := os.ReadFile(specPath)
			if err != nil {
				fmt.Printf("✗ Cannot read spec: %s\n", specPath)
				errors++
				continue
			}

			// Check for placeholder text
			if strings.Contains(string(content), "[PLACEHOLDER]") {
				fmt.Printf("✗ Spec contains placeholder text: %s\n", specPath)
				errors++
			}

			// Check for requirement ID format (basic check)
			layerPrefixes := map[string]string{
				"business":       "BUS",
				"functional":     "FUN",
				"stack":          "STK",
				"infrastructure": "INF",
				"coverage":       "COV",
			}
			prefix := layerPrefixes[layer]

			// Look for at least one requirement ID with correct prefix
			hasCorrectID := strings.Contains(string(content), prefix+"-")
			hasAcceptanceCriteria := strings.Contains(string(content), "## Acceptance Criteria") ||
				strings.Contains(string(content), "# Acceptance Criteria")

			if hasAcceptanceCriteria && !hasCorrectID {
				fmt.Printf("⚠ Spec may have malformed requirement IDs: %s (expected %s-*)\n", specPath, prefix)
				// Warning, not error
			}
		}
	}

	fmt.Println()
	if errors == 0 {
		fmt.Println("✓ Validation passed - project structure is valid")
		os.Exit(0)
	} else {
		fmt.Printf("✗ Validation failed with %d error(s)\n", errors)
		os.Exit(1)
	}
}

// printPhaseStatus outputs the phase completion status with optional timestamp
// hasSpecs indicates if there are any specs in the phase's layers
func printPhaseStatus(phase PhaseState, hasSpecs bool) {
	if phase.Completed {
		timestamp := phase.Timestamp
		if timestamp != "" {
			// Parse and format timestamp
			if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
				timestamp = t.Format("2006-01-02")
			}
			fmt.Printf("✓ Complete (%s)\n", timestamp)
		} else {
			fmt.Println("✓ Complete")
		}
	} else if hasSpecs {
		fmt.Println("In progress")
	} else {
		fmt.Println("Not started")
	}
}

func cmdStatus() {
	// Check if .smaqit exists
	if _, err := os.Stat(".smaqit"); os.IsNotExist(err) {
		fmt.Println("Error: .smaqit/ directory not found")
		fmt.Println("Run 'smaqit init' to initialize smaqit in this project")
		os.Exit(1)
	}

	fmt.Println("smaqit Project Status")
	fmt.Println("=====================\n")

	// Read state.json
	stateFilePath := filepath.Join(".smaqit", "state.json")
	state := readStateFile(stateFilePath)

	// Scan specs by layer
	layers := []string{"business", "functional", "stack", "infrastructure", "coverage"}
	layerCounts := make(map[string]int)
	totalSpecs := 0

	for _, layer := range layers {
		specDir := filepath.Join("specs", layer)
		entries, err := os.ReadDir(specDir)
		if err != nil {
			layerCounts[layer] = 0
			continue
		}

		count := 0
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				count++
			}
		}
		layerCounts[layer] = count
		totalSpecs += count
	}

	// Display phases with nested layers
	// Phase 1: Develop
	developPhase := state.Phases.Develop
	developHasSpecs := layerCounts["business"] > 0 || layerCounts["functional"] > 0 || layerCounts["stack"] > 0
	fmt.Print("Phase 1 (Develop): ")
	printPhaseStatus(developPhase, developHasSpecs)
	fmt.Printf("  Business:        %d spec(s)\n", layerCounts["business"])
	fmt.Printf("  Functional:      %d spec(s)\n", layerCounts["functional"])
	fmt.Printf("  Stack:           %d spec(s)\n", layerCounts["stack"])
	fmt.Println()

	// Phase 2: Deploy
	deployPhase := state.Phases.Deploy
	deployHasSpecs := layerCounts["infrastructure"] > 0
	fmt.Print("Phase 2 (Deploy): ")
	printPhaseStatus(deployPhase, deployHasSpecs)
	fmt.Printf("  Infrastructure:  %d spec(s)\n", layerCounts["infrastructure"])
	fmt.Println()

	// Phase 3: Validate
	validatePhase := state.Phases.Validate
	validateHasSpecs := layerCounts["coverage"] > 0
	fmt.Print("Phase 3 (Validate): ")
	printPhaseStatus(validatePhase, validateHasSpecs)
	fmt.Printf("  Coverage:        %d spec(s)\n", layerCounts["coverage"])
	
	// Display total
	fmt.Printf("\nTotal: %d specification(s)\n", totalSpecs)

	// Next steps based on actual spec content and phase state
	fmt.Println("\nNext steps:")

	// Phase 1: Develop
	if !developPhase.Completed {
		hasAnyPhase1 := layerCounts["business"] > 0 || layerCounts["functional"] > 0 || layerCounts["stack"] > 0

		if !hasAnyPhase1 {
			fmt.Println("  • Type '/smaqit.business' to start with business specifications")
		} else {
			// Suggest missing layers first
			if layerCounts["business"] == 0 {
				fmt.Println("  • Type '/smaqit.business' to add business specifications")
			} else if layerCounts["functional"] == 0 {
				fmt.Println("  • Type '/smaqit.functional' to add functional specifications")
			} else if layerCounts["stack"] == 0 {
				fmt.Println("  • Type '/smaqit.stack' to add technical stack specifications")
			} else {
				// All Phase 1 specs exist
				fmt.Println("  • Type '/smaqit.development' to implement from specifications")
			}
		}
	} else if !deployPhase.Completed {
		// Phase 2: Deploy
		if layerCounts["infrastructure"] == 0 {
			fmt.Println("  • Type '/smaqit.infrastructure' to define infrastructure specifications")
		} else {
			fmt.Println("  • Type '/smaqit.deployment' to deploy the implementation")
		}
	} else if !validatePhase.Completed {
		// Phase 3: Validate
		if layerCounts["coverage"] == 0 {
			fmt.Println("  • Type '/smaqit.coverage' to define test coverage specifications")
		} else {
			fmt.Println("  • Type '/smaqit.validation' to validate the deployment")
		}
	} else {
		fmt.Println("  • All phases complete. Run '/smaqit.orchestrate' to iterate or extend.")
	}
}
