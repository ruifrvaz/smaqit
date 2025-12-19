package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed framework/*.md
var frameworkFiles embed.FS

//go:embed templates/specs/*.md
var templateFiles embed.FS

//go:embed agents/*.md
var agentFiles embed.FS

// Version is set via ldflags during build: -X main.Version=$(VERSION)
var Version = "dev"

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
	case "version":
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
	fmt.Println("  /smaqit.develop       Run develop phase (business → functional → stack → build)")
	fmt.Println("  /smaqit.deploy        Run deploy phase (infrastructure → deploy)")
	fmt.Println("  /smaqit.validate      Run validate phase (coverage → validate)")
	fmt.Println("  /smaqit.business      Create business layer specifications")
	fmt.Println("  /smaqit.functional    Create functional layer specifications")
	fmt.Println("  /smaqit.stack         Create stack layer specifications")
	fmt.Println("  /smaqit.infrastructure Create infrastructure layer specifications")
	fmt.Println("  /smaqit.coverage      Create coverage layer specifications")
	fmt.Println()
	fmt.Println("Getting Started:")
	fmt.Println("  1. Run 'smaqit init' in your project directory")
	fmt.Println("  2. Open GitHub Copilot chat in VS Code")
	fmt.Println("  3. Type '/smaqit.develop' to start the development phase")
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
		".smaqit/framework",
		".smaqit/templates/specs",
		"specs/business",
		"specs/functional",
		"specs/stack",
		"specs/infrastructure",
		"specs/coverage",
		".github/agents",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Copy framework files
	if err := copyEmbeddedDir(frameworkFiles, "framework", ".smaqit/framework"); err != nil {
		fmt.Printf("Error copying framework files: %v\n", err)
		os.Exit(1)
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

	// Write version file
	versionFile := filepath.Join(".smaqit", "VERSION")
	if err := os.WriteFile(versionFile, []byte(Version+"\n"), 0644); err != nil {
		fmt.Printf("Error writing VERSION file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Created .smaqit/ directory structure")
	fmt.Println("✓ Copied framework files")
	fmt.Println("✓ Copied templates")
	fmt.Println("✓ Copied agent definitions")
	fmt.Println("✓ Initialized smaqit %s\n\n", Version)
	fmt.Println("Next steps:")
	fmt.Println("  1. Open GitHub Copilot chat in VS Code")
	fmt.Println("  2. Type '/smaqit.develop' to start the development phase")
	fmt.Println("  3. Or type '/smaqit.business' to create business specifications")
}

// copyEmbeddedDir copies files from an embedded FS to a target directory
func copyEmbeddedDir(embeddedFS embed.FS, srcDir, dstDir string) error {
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

func cmdUninstall() {
	// Check if .smaqit exists
	if _, err := os.Stat(".smaqit"); os.IsNotExist(err) {
		fmt.Println("No smaqit installation found in this directory")
		os.Exit(0)
	}

	// Prompt for confirmation
	fmt.Println("This will remove:")
	fmt.Println("  • .smaqit/")
	fmt.Println("  • specs/")
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

	// Remove directories
	errors := 0

	if err := os.RemoveAll(".smaqit"); err != nil {
		fmt.Printf("Error removing .smaqit/: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ Removed .smaqit/")
	}

	if err := os.RemoveAll("specs"); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error removing specs/: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ Removed specs/")
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

func cmdDevelop() {
	fmt.Println("The 'develop' command has been moved to GitHub Copilot prompts.")
	fmt.Println("Type '/smaqit.develop' in GitHub Copilot chat to run the develop phase.")
	os.Exit(1)
}

func cmdDeploy() {
	fmt.Println("The 'deploy' command has been moved to GitHub Copilot prompts.")
	fmt.Println("Type '/smaqit.deploy' in GitHub Copilot chat to run the deploy phase.")
	os.Exit(1)
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
		".smaqit/framework",
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

	// Check framework files
	frameworkFileNames := []string{
		"SMAQIT.md", "LAYERS.md", "PHASES.md",
		"TEMPLATES.md", "AGENTS.md", "ARTIFACTS.md",
	}

	for _, fileName := range frameworkFileNames {
		filePath := filepath.Join(".smaqit", "framework", fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("✗ Missing framework file: %s\n", fileName)
			errors++
		}
	}

	if errors == 0 {
		fmt.Println("✓ Framework files are present")
	}

	// Validate spec files (basic checks)
	layers := []string{"business", "functional", "stack", "infrastructure", "coverage"}
	for _, layer := range layers {
		specDir := filepath.Join(".smaqit", "specs", layer)
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

	// Read version
	versionFile := filepath.Join(".smaqit", "VERSION")
	versionBytes, err := os.ReadFile(versionFile)
	installedVersion := "unknown"
	if err == nil {
		installedVersion = strings.TrimSpace(string(versionBytes))
	}

	fmt.Printf("Version: %s", installedVersion)
	if installedVersion != Version {
		fmt.Printf(" (current: %s)", Version)
	}
	fmt.Println("\n")

	// Scan specs by layer
	layers := []string{"business", "functional", "stack", "infrastructure", "coverage"}
	layerCounts := make(map[string]int)
	totalSpecs := 0

	for _, layer := range layers {
		specDir := filepath.Join(".smaqit", "specs", layer)
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

	// Display layer coverage
	fmt.Println("Specification Coverage:")
	fmt.Printf("  Business:        %d spec(s)\n", layerCounts["business"])
	fmt.Printf("  Functional:      %d spec(s)\n", layerCounts["functional"])
	fmt.Printf("  Stack:           %d spec(s)\n", layerCounts["stack"])
	fmt.Printf("  Infrastructure:  %d spec(s)\n", layerCounts["infrastructure"])
	fmt.Printf("  Coverage:        %d spec(s)\n", layerCounts["coverage"])
	fmt.Printf("\nTotal: %d specification(s)\n\n", totalSpecs)

	// Phase completion status
	hasPhase1 := layerCounts["business"] > 0 || layerCounts["functional"] > 0 || layerCounts["stack"] > 0
	hasPhase2 := layerCounts["infrastructure"] > 0
	hasPhase3 := layerCounts["coverage"] > 0

	fmt.Println("Phase Status:")
	if hasPhase1 {
		fmt.Println("  ✓ Develop (Phase 1): In progress or complete")
	} else {
		fmt.Println("  ○ Develop (Phase 1): Not started")
	}

	if hasPhase2 {
		fmt.Println("  ✓ Deploy (Phase 2): In progress or complete")
	} else {
		fmt.Println("  ○ Deploy (Phase 2): Not started")
	}

	if hasPhase3 {
		fmt.Println("  ✓ Validate (Phase 3): In progress or complete")
	} else {
		fmt.Println("  ○ Validate (Phase 3): Not started")
	}

	fmt.Println("\nNext steps:")
	if !hasPhase1 {
		fmt.Println("  • Type '/develop' in GitHub Copilot chat to start Phase 1")
	} else if !hasPhase2 {
		fmt.Println("  • Type '/deploy' in GitHub Copilot chat to start Phase 2")
	} else if !hasPhase3 {
		fmt.Println("  • Type '/validate' in GitHub Copilot chat to start Phase 3")
	} else {
		fmt.Println("  • All phases have specs. Review and iterate as needed.")
	}
}
