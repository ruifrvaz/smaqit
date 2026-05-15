package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/specs/*.md
var templateFiles embed.FS

//go:embed templates/workflows/*.yml
var workflowFiles embed.FS

//go:embed agents/*.md
var agentFiles embed.FS

//go:embed skills/**/*.md
var skillFiles embed.FS

// Version is set via ldflags during build: -X main.Version=$(VERSION)
var Version = "1.0.0"

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
	case "plan":
		cmdPlan()
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
  plan       Show specs to process (for agents)
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
	fmt.Println("                    spec templates, skills, and agent definitions")
	fmt.Println("                    Optional: specify target directory (created if needed)")
	fmt.Println()
	fmt.Println("  smaqit plan       Show work plan for current phase")
	fmt.Println("    --phase=X       Specify phase: develop, deploy, or validate")
	fmt.Println("    --verbose       Show detailed information")
	fmt.Println("    --regen         Mark all specs for regeneration")
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
	fmt.Println("                    Removes .smaqit/, .github/agents/, .github/skills/")
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

func cmdPlan() {
	// Check if .smaqit exists
	if _, err := os.Stat(".smaqit"); os.IsNotExist(err) {
		fmt.Println("Error: .smaqit/ directory not found")
		fmt.Println("Run 'smaqit init' to initialize smaqit in this project")
		os.Exit(1)
	}

	// Parse flags
	phase := ""
	verbose := false
	regen := false

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--phase=") {
			phase = strings.TrimPrefix(arg, "--phase=")
		} else if arg == "--verbose" {
			verbose = true
		} else if arg == "--regen" {
			regen = true
		}
	}

	// Scan all specs
	allSpecs, err := scanSpecs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning specs: %v\n", err)
		os.Exit(1)
	}

	// If no phase specified, show summary of all phases
	if phase == "" {
		fmt.Println("smaqit Work Plan")
		fmt.Println("================\n")

		phases := []struct {
			name   string
			layers []string
		}{
			{"develop", []string{"business", "functional", "stack"}},
			{"deploy", []string{"infrastructure"}},
			{"validate", []string{"coverage"}},
		}

		for i, p := range phases {
			specs := getPhaseSpecs(allSpecs, p.name)
			toProcess := filterSpecsByStatus(specs, p.name, false)
			completed := len(specs) - len(toProcess)

			fmt.Printf("Phase %d (%s): %d pending, %d completed\n",
				i+1, strings.Title(p.name), len(toProcess), completed)
		}

		fmt.Println("\nRun `smaqit plan --phase=<phase>` for details")
		return
	}

	// Validate phase
	validPhases := map[string]bool{"develop": true, "deploy": true, "validate": true}
	if !validPhases[phase] {
		fmt.Fprintf(os.Stderr, "Error: invalid phase '%s' (must be: develop, deploy, or validate)\n", phase)
		os.Exit(1)
	}

	// Get specs for the specified phase
	phaseSpecs := getPhaseSpecs(allSpecs, phase)
	toProcess := filterSpecsByStatus(phaseSpecs, phase, regen)
	completed := []Spec{}

	for _, spec := range phaseSpecs {
		isProcessing := false
		for _, tp := range toProcess {
			if tp.Path == spec.Path {
				isProcessing = true
				break
			}
		}
		if !isProcessing {
			completed = append(completed, spec)
		}
	}

	// Verbose output (human-readable)
	if verbose {
		if regen {
			fmt.Println("REGENERATION MODE: All specs marked for reprocessing\n")
		}

		fmt.Printf("Phase: %s\n\n", strings.Title(phase))

		if len(toProcess) > 0 {
			fmt.Printf("To Process (%d):\n", len(toProcess))
			for _, spec := range toProcess {
				if regen && spec.Frontmatter.Status != "draft" {
					fmt.Printf("  [%s→regen] %s", spec.Frontmatter.Status, spec.Path)
				} else {
					fmt.Printf("  [%s] %s", spec.Frontmatter.Status, spec.Path)
				}
				if spec.Frontmatter.ID != "" {
					fmt.Printf(" (%s)", spec.Frontmatter.ID)
				}
				fmt.Println()
			}
			fmt.Println()
		}

		if len(completed) > 0 {
			fmt.Printf("Completed (%d):\n", len(completed))
			for _, spec := range completed {
				fmt.Printf("  [%s] %s", spec.Frontmatter.Status, spec.Path)
				if spec.Frontmatter.ID != "" {
					fmt.Printf(" (%s)", spec.Frontmatter.ID)
				}
				fmt.Println()
			}
			fmt.Println()
		}

		if len(toProcess) == 0 {
			fmt.Println("All specs already processed.")
			if !regen {
				fmt.Printf("Use `smaqit plan --phase=%s --regen` to regenerate.\n", phase)
			}
		} else {
			fmt.Printf("Next: Run /smaqit.%s to process %d spec(s)\n",
				getAgentName(phase), len(toProcess))
		}

		return
	}

	// Default output (agent-friendly): just paths, one per line
	if len(toProcess) == 0 {
		// Empty output means nothing to process
		os.Exit(0)
	}

	for _, spec := range toProcess {
		fmt.Println(spec.Path)
	}
}

func getAgentName(phase string) string {
	switch phase {
	case "develop":
		return "development"
	case "deploy":
		return "deployment"
	case "validate":
		return "validation"
	default:
		return phase
	}
}

// detectConflicts checks which embedded files would conflict with existing files
func detectConflicts() []string {
	var conflicts []string

	// Define the file mappings that will be installed
	fileMappings := []struct {
		embeddedFS       embed.FS
		srcDir           string
		dstDir           string
		skipIfExists     bool // Workflow files are never overwritten
	}{
		{templateFiles, "templates/specs", ".smaqit/templates/specs", false},
		{agentFiles, "agents", ".github/agents", false},
		{skillFiles, "skills", ".github/skills", false},
		{workflowFiles, "templates/workflows", ".github/workflows", true},
	}

	// Check each file mapping for conflicts
	for _, mapping := range fileMappings {
		err := fs.WalkDir(mapping.embeddedFS, mapping.srcDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			// Calculate destination path (handle cross-platform path separators)
			relPath := strings.TrimPrefix(path, mapping.srcDir)
			relPath = strings.TrimPrefix(relPath, "/") // Remove leading slash if present
			dstPath := filepath.Join(mapping.dstDir, relPath)

			// Skip files that are never overwritten (e.g., workflows)
			if mapping.skipIfExists {
				if _, err := os.Stat(dstPath); err == nil {
					// File exists, would be skipped anyway
					return nil
				}
			}

			// Check if file exists
			if _, err := os.Stat(dstPath); err == nil {
				conflicts = append(conflicts, dstPath)
			}

			return nil
		})

		if err != nil {
			// Continue checking other mappings even if one fails
			continue
		}
	}

	return conflicts
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

	// Check if .smaqit already exists and handle reinstallation
	if _, err := os.Stat(".smaqit"); err == nil {
		fmt.Println("Existing smaqit installation detected.")
		fmt.Println()

		// Check for conflicts
		conflicts := detectConflicts()
		
		if len(conflicts) == 0 {
			fmt.Println("No conflicts detected. Proceeding with installation...")
		} else {
			fmt.Println("The following files will be overwritten:")
			for _, file := range conflicts {
				fmt.Printf("  • %s\n", file)
			}
			fmt.Println()
			fmt.Print("Continue with installation and overwrite these files? [y/N]: ")

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response != "y" && response != "yes" {
				fmt.Println("Installation cancelled")
				os.Exit(0)
			}
		}
		fmt.Println()
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
		".github/skills",
		".github/workflows",
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

	// Copy skill files
	if err := copyEmbeddedDir(skillFiles, "skills", ".github/skills"); err != nil {
		fmt.Printf("Error copying skill files: %v\n", err)
		os.Exit(1)
	}

	// Copy workflow files
	if err := copyEmbeddedDir(workflowFiles, "templates/workflows", ".github/workflows"); err != nil {
		fmt.Printf("Error copying workflow files: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Created .smaqit/ directory structure")
	fmt.Println("✓ Copied templates")
	fmt.Println("✓ Copied agent definitions")
	fmt.Println("✓ Copied skill files")
	fmt.Println("✓ Copied workflow files")
	fmt.Printf("✓ Initialized smaqit %s\n\n", Version)
	fmt.Println("Next steps:")
	fmt.Println("  1. Open GitHub Copilot chat in VS Code")
	fmt.Println("  2. Type '/smaqit.development' to orchestrate the entire Development phase")
	fmt.Println("  3. Or type '/smaqit.business' to begin with business specifications only")
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

		// Skip if file already exists (don't overwrite workflows)
		if strings.Contains(dstDir, "workflows") {
			if _, err := os.Stat(dstPath); err == nil {
				// File exists, skip it
				return nil
			}
		}

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
	fmt.Println("  • .github/skills/")
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

	if err := os.RemoveAll(filepath.Join(".github", "skills")); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error removing .github/skills/: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ Removed .github/skills/")
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
		".github/skills",
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

func cmdStatus() {
	// Check if .smaqit exists
	if _, err := os.Stat(".smaqit"); os.IsNotExist(err) {
		fmt.Println("Error: .smaqit/ directory not found")
		fmt.Println("Run 'smaqit init' to initialize smaqit in this project")
		os.Exit(1)
	}

	fmt.Println("smaqit Project Status")
	fmt.Println("=====================\n")

	// Scan all specs
	allSpecs, err := scanSpecs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning specs: %v\n", err)
		os.Exit(1)
	}

	// Count specs by layer and status
	layerCounts := make(map[string]int)
	layerStatusCounts := make(map[string]map[string]int)

	for layer, specs := range allSpecs {
		layerCounts[layer] = len(specs)
		layerStatusCounts[layer] = make(map[string]int)

		for _, spec := range specs {
			layerStatusCounts[layer][spec.Frontmatter.Status]++
		}
	}

	totalSpecs := 0
	for _, count := range layerCounts {
		totalSpecs += count
	}

	// Determine phase completion status
	// A phase is ONLY complete when:
	// 1. ALL required layers have at least one spec, AND
	// 2. ALL specs in those layers have reached the target status

	// Phase 1: Develop (requires business, functional, stack)
	developSpecs := getPhaseSpecs(allSpecs, "develop")
	developImplemented := 0
	for _, spec := range developSpecs {
		if spec.Frontmatter.Status == "implemented" ||
			spec.Frontmatter.Status == "deployed" ||
			spec.Frontmatter.Status == "validated" {
			developImplemented++
		}
	}
	// Require all three layers present
	hasAllDevelopLayers := layerCounts["business"] > 0 &&
		layerCounts["functional"] > 0 &&
		layerCounts["stack"] > 0
	developCompleted := hasAllDevelopLayers &&
		len(developSpecs) > 0 &&
		developImplemented == len(developSpecs)

	// Phase 2: Deploy (requires infrastructure)
	deploySpecs := getPhaseSpecs(allSpecs, "deploy")
	deployDeployed := 0
	for _, spec := range deploySpecs {
		if spec.Frontmatter.Status == "deployed" ||
			spec.Frontmatter.Status == "validated" {
			deployDeployed++
		}
	}
	deployCompleted := layerCounts["infrastructure"] > 0 &&
		len(deploySpecs) > 0 &&
		deployDeployed == len(deploySpecs)

	// Phase 3: Validate (requires coverage)
	validateSpecs := getPhaseSpecs(allSpecs, "validate")
	validateValidated := 0
	for _, spec := range validateSpecs {
		if spec.Frontmatter.Status == "validated" {
			validateValidated++
		}
	}
	validateCompleted := layerCounts["coverage"] > 0 &&
		len(validateSpecs) > 0 &&
		validateValidated == len(validateSpecs)

	// Calculate pending counts for in-progress display
	developPending := len(filterSpecsByStatus(developSpecs, "develop", false))
	deployPending := len(filterSpecsByStatus(deploySpecs, "deploy", false))
	validatePending := len(filterSpecsByStatus(validateSpecs, "validate", false))

	// Display phases with nested layers
	// Phase 1: Develop
	fmt.Print("Phase 1 (Develop): ")
	if developCompleted {
		fmt.Println("✓ Complete")
	} else if len(developSpecs) > 0 {
		fmt.Printf("⚙ In progress (%d pending)\n", developPending)
	} else {
		fmt.Println("✗ Not started")
	}

	// Show status breakdown for each layer
	businessStatusSummary := getStatusSummary(layerStatusCounts["business"], layerCounts["business"])
	functionalStatusSummary := getStatusSummary(layerStatusCounts["functional"], layerCounts["functional"])
	stackStatusSummary := getStatusSummary(layerStatusCounts["stack"], layerCounts["stack"])

	fmt.Printf("  Business:        %d spec(s)%s\n", layerCounts["business"], businessStatusSummary)
	fmt.Printf("  Functional:      %d spec(s)%s\n", layerCounts["functional"], functionalStatusSummary)
	fmt.Printf("  Stack:           %d spec(s)%s\n", layerCounts["stack"], stackStatusSummary)
	fmt.Println()

	// Phase 2: Deploy
	fmt.Print("Phase 2 (Deploy): ")
	if deployCompleted {
		fmt.Println("✓ Complete")
	} else if len(deploySpecs) > 0 {
		fmt.Printf("⚙ In progress (%d pending)\n", deployPending)
	} else {
		fmt.Println("✗ Not started")
	}

	infraStatusSummary := getStatusSummary(layerStatusCounts["infrastructure"], layerCounts["infrastructure"])
	fmt.Printf("  Infrastructure:  %d spec(s)%s\n", layerCounts["infrastructure"], infraStatusSummary)
	fmt.Println()

	// Phase 3: Validate
	fmt.Print("Phase 3 (Validate): ")
	if validateCompleted {
		fmt.Println("✓ Complete")
	} else if len(validateSpecs) > 0 {
		fmt.Printf("⚙ In progress (%d pending)\n", validatePending)
	} else {
		fmt.Println("✗ Not started")
	}

	coverageStatusSummary := getStatusSummary(layerStatusCounts["coverage"], layerCounts["coverage"])
	fmt.Printf("  Coverage:        %d spec(s)%s\n", layerCounts["coverage"], coverageStatusSummary)

	// Display total
	fmt.Printf("\nTotal: %d specification(s)\n", totalSpecs)

	// Next steps based on actual spec content and status
	fmt.Println("\nNext steps:")

	// Phase 1: Develop
	if !developCompleted {
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
				fmt.Println("  • Run 'smaqit plan --phase=develop' to see work plan")
				fmt.Println("  • Type '/smaqit.development' to implement from specifications")
			}
		}
	} else if !deployCompleted {
		// Phase 2: Deploy
		if layerCounts["infrastructure"] == 0 {
			fmt.Println("  • Type '/smaqit.infrastructure' to define infrastructure specifications")
		} else {
			fmt.Println("  • Run 'smaqit plan --phase=deploy' to see work plan")
			fmt.Println("  • Type '/smaqit.deployment' to deploy the implementation")
		}
	} else if !validateCompleted {
		// Phase 3: Validate
		if layerCounts["coverage"] == 0 {
			fmt.Println("  • Type '/smaqit.coverage' to define test coverage specifications")
		} else {
			fmt.Println("  • Run 'smaqit plan --phase=validate' to see work plan")
			fmt.Println("  • Type '/smaqit.validation' to validate the deployment")
		}
	} else {
		fmt.Println("  • All phases complete. Use '/smaqit.development --regen' to iterate or extend.")
	}
}

// getStatusSummary returns a formatted string showing status breakdown
func getStatusSummary(statusCounts map[string]int, total int) string {
	if total == 0 {
		return ""
	}

	parts := []string{}
	if draft := statusCounts["draft"]; draft > 0 {
		parts = append(parts, fmt.Sprintf("%d draft", draft))
	}
	if failed := statusCounts["failed"]; failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	if impl := statusCounts["implemented"]; impl > 0 {
		parts = append(parts, fmt.Sprintf("%d implemented", impl))
	}
	if deployed := statusCounts["deployed"]; deployed > 0 {
		parts = append(parts, fmt.Sprintf("%d deployed", deployed))
	}
	if validated := statusCounts["validated"]; validated > 0 {
		parts = append(parts, fmt.Sprintf("%d validated", validated))
	}

	if len(parts) == 0 {
		return ""
	}

	return " (" + strings.Join(parts, ", ") + ")"
}
