package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed templates/specs/*.md
var templateFiles embed.FS

//go:embed templates/designs/*.md
var designTemplateFiles embed.FS

//go:embed tools/plantuml-tools.tgz
var designToolArchive embed.FS

//go:embed templates/workflows/*.yml
var workflowFiles embed.FS

//go:embed agents-copilot/*.md
var agentFiles embed.FS

//go:embed agents-claude/*.md
var claudeAgentFiles embed.FS

//go:embed commands-claude/*.md
var claudeCommandFiles embed.FS

//go:embed skills-shared
var skillFilesShared embed.FS

//go:embed skills-claude
var skillFilesClaude embed.FS

//go:embed agents-codex/*.toml
var codexAgentFiles embed.FS

//go:embed templates/AGENTS.md.template
var agentsMdTemplate embed.FS

//go:embed templates/CLAUDE.md.template
var claudeMdTemplate embed.FS

// Version is set via ldflags during build: -X main.Version=$(VERSION)
var Version = "3.1.0"

func main() {
	if len(os.Args) < 2 {
		cmdHelp()
		return
	}

	switch os.Args[1] {
	case "init":
		targetDir := "."
		if len(os.Args) > 2 {
			targetDir = os.Args[2]
		}
		cmdInit(targetDir)
	case "--install-global":
		cmdInstallGlobal()
	case "plan":
		cmdPlan()
	case "status":
		cmdStatus()
	case "validate":
		cmdValidate()
	case "design":
		cmdDesign(os.Args[2:])
	case "slots":
		cmdSlots(os.Args[2:])
	case "spec":
		cmdSpec(os.Args[2:])
	case "mcp":
		cmdMCP(os.Args[2:])
	case "help", "--help", "-h":
		cmdHelp()
	case "uninstall":
		cmdUninstall()
	case "update":
		runUpdate()
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
  init [dir] Scaffold project-local smaqit directories and configuration
             Optional: specify target directory (default: current)
  plan       Show specs to process (for agents)
  status     Show project state and spec coverage
  validate   Verify project structure integrity
  design     Render, attest, and validate canonical PlantUML designs
  slots      Resolve and validate deployment slot bindings
  spec       Read structured Infrastructure spec data (environments)
  mcp        Run a bundled MCP server (normally invoked by an agent host)
  help       Show detailed command help
  uninstall  Remove smaqit from project
  update     Update smaqit to the latest release
  version    Show smaqit version`)
}

func cmdHelp() {
	fmt.Println("smaqit - Spec-Driven Agent Orchestration Kit")
	fmt.Printf("Version: %s\n\n", Version)

	fmt.Println("CLI Commands:")
	fmt.Println("  smaqit init [dir] Scaffold smaqit project structure")
	fmt.Println("                    Creates project-local templates, specifications, designs,")
	fmt.Println("                    and PlantUML configuration. Agents and skills are installed")
	fmt.Println("                    globally by the installer script.")
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
	fmt.Println("                    Checks project directory structure and")
	fmt.Println("                    validates spec template compliance")
	fmt.Println()
	fmt.Println("  smaqit design render <design.md>  Syntax-check and render its PNG")
	fmt.Println("  smaqit design attest <design.md>  Record a completed visual review")
	fmt.Println("  smaqit design validate [path]     Run all mandatory design gates")
	fmt.Println()
	fmt.Println("  smaqit slots plan --phase=deploy --mode=<mode> --env=<env>")
	fmt.Println("                    Resolve the ordered deployment slot plan from .smaqit/slots.yml")
	fmt.Println("                    and provisioning_mode defaults; --path=standard|existing-cicd-pr|ci-push|ci-bootstrap,")
	fmt.Println("                    --slot=<slot> for a single-slot (day-two) run, --json for agents")
	fmt.Println("  smaqit slots contract [--slot=<slot>] [--mode=<mode>] [--json]")
	fmt.Println("                    Print each slot's contract: inputs, handoff, completion signal, and what a")
	fmt.Println("                    mode requires; a filler that meets it replaces the default")
	fmt.Println("  smaqit slots validate             Validate .smaqit/slots.yml and its bound skills")
	fmt.Println("                    For --mode=existing-k3s, the plan is refused unless the")
	fmt.Println("                    Infrastructure spec's Environments row for --env is complete,")
	fmt.Println("                    and any slot that would deploy from this machine is refused")
	fmt.Println("                    (use --path=ci-bootstrap, ci-push or existing-cicd-pr)")
	fmt.Println()
	fmt.Println("  smaqit spec environments [--env=<env>] [--json]")
	fmt.Println("                    Validate and print the Infrastructure spec's Environments table")
	fmt.Println("                    (reachability, Ingress mode, host, health path, mesh access)")
	fmt.Println()
	fmt.Println("  smaqit mcp plantuml               Run the bundled PlantUML MCP server")
	fmt.Println("  smaqit mcp verify                 Verify the PlantUML MCP wrapper and tools")
	fmt.Println()
	fmt.Println("  smaqit help       Show this help message")
	fmt.Println()
	fmt.Println("  smaqit uninstall  Remove smaqit from project and its owned global payload")
	fmt.Println("                    Removes project state and only smaqit-owned global")
	fmt.Println("                    agents and skills; unrelated user content is preserved")
	fmt.Println()
	fmt.Println("  smaqit update     Update to the latest release")
	fmt.Println("                    Downloads the latest GitHub release for your platform,")
	fmt.Println("                    replaces the running binary, and re-initializes project")
	fmt.Println("                    assets if .smaqit/ exists")
	fmt.Println()
	fmt.Println("  smaqit version    Show smaqit version")
	fmt.Println()
	fmt.Println("Agents:")
	fmt.Println("  GitHub Copilot and Claude Code: use / with these names")
	fmt.Println("  Codex: ask Codex to spawn the named agent")
	fmt.Println("  /smaqit.development   Run Development implementation agent (build from specs)")
	fmt.Println("  /smaqit.deployment    Run Deployment implementation agent (deploy from specs)")
	fmt.Println("  /smaqit.validation    Run Validation implementation agent (test from specs)")
	fmt.Println("  /smaqit.business      Create business layer specifications")
	fmt.Println("  /smaqit.functional    Create functional layer specifications")
	fmt.Println("  /smaqit.stack         Create stack layer specifications")
	fmt.Println("  /smaqit.infrastructure Create infrastructure layer specifications")
	fmt.Println("  /smaqit.coverage      Create coverage layer specifications")
	fmt.Println("  /smaqit.qa            Answer questions about the smaqit framework")
	fmt.Println()
	fmt.Println("  Note: in Claude Code, only /smaqit.development, /smaqit.deployment,")
	fmt.Println("  /smaqit.validation, and /smaqit.qa are slash commands. The five spec")
	fmt.Println("  agents (business/functional/stack/infrastructure/coverage) are invoked")
	fmt.Println("  automatically by those phase commands — this matches Copilot's")
	fmt.Println("  user-invocable:false behavior for the same five agents.")
	fmt.Println("  In Codex, phase agents spawn the five specification agents as custom")
	fmt.Println("  subagents. Repository skills can be selected with /skills or mentioned with $.")
	fmt.Println()
	fmt.Println("Getting Started:")
	fmt.Println("  1. Run 'smaqit init' in your project directory")
	fmt.Println("  2. Open GitHub Copilot chat, Claude Code, or Codex")
	fmt.Println("  3. Invoke or ask to spawn smaqit.development for the Development phase")
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
		fmt.Println("================")
		fmt.Println()

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
	if err := validatePhaseDesignReadiness(getPhaseDesignGateSpecs(allSpecs, phase)); err != nil {
		fmt.Fprintf(os.Stderr, "Phase design readiness failed: %v\n", err)
		os.Exit(1)
	}
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
			fmt.Println("REGENERATION MODE: All specs marked for reprocessing")
			fmt.Println()
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
				if !spec.DesignReady {
					fmt.Printf("    design gate: %s\n", spec.DesignError)
				}
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
			fmt.Printf("Next: Run the smaqit.%s agent to process %d spec(s)\n",
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

// resolveGlobalDir returns a user-level destination for an owned payload kind.
// Platform configuration overrides apply only to their respective client roots;
// the shared skill tree read by both Copilot and Codex deliberately remains under the user's home.
func resolveGlobalDir(kind string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving user home: %w", err)
	}
	switch kind {
	case "copilot-agents":
		if root := os.Getenv("COPILOT_HOME"); root != "" {
			return filepath.Join(root, "agents"), nil
		}
		return filepath.Join(home, ".copilot", "agents"), nil
	case "claude-agents":
		if root := os.Getenv("CLAUDE_CONFIG_DIR"); root != "" {
			return filepath.Join(root, "agents"), nil
		}
		return filepath.Join(home, ".claude", "agents"), nil
	case "claude-commands":
		if root := os.Getenv("CLAUDE_CONFIG_DIR"); root != "" {
			return filepath.Join(root, "commands"), nil
		}
		return filepath.Join(home, ".claude", "commands"), nil
	case "claude-skills":
		if root := os.Getenv("CLAUDE_CONFIG_DIR"); root != "" {
			return filepath.Join(root, "skills"), nil
		}
		return filepath.Join(home, ".claude", "skills"), nil
	case "codex-agents":
		if root := os.Getenv("CODEX_HOME"); root != "" {
			return filepath.Join(root, "agents"), nil
		}
		return filepath.Join(home, ".codex", "agents"), nil
	case "shared-skills":
		return filepath.Join(home, ".agents", "skills"), nil
	default:
		return "", fmt.Errorf("unknown global installation kind %q", kind)
	}
}

func cmdInstallGlobal() {
	mappings := []struct {
		embeddedFS embed.FS
		srcDir     string
		kind       string
		label      string
		skills     bool
	}{
		{agentFiles, "agents-copilot", "copilot-agents", "Copilot agents", false},
		{skillFilesShared, "skills-shared", "shared-skills", "shared Copilot/Codex skills", true},
		{claudeAgentFiles, "agents-claude", "claude-agents", "Claude agents", false},
		{claudeCommandFiles, "commands-claude", "claude-commands", "Claude commands", false},
		{skillFilesClaude, "skills-claude", "claude-skills", "Claude skills", true},
		{codexAgentFiles, "agents-codex", "codex-agents", "Codex agents", false},
	}
	for _, mapping := range mappings {
		destination, err := resolveGlobalDir(mapping.kind)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving %s path: %v\n", mapping.label, err)
			os.Exit(1)
		}
		removed, err := installGlobalTree(mapping.embeddedFS, mapping.srcDir, destination, mapping.skills)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error installing %s: %v\n", mapping.label, err)
			os.Exit(1)
		}
		if removed > 0 {
			fmt.Printf("✓ Installed %s to %s (removed %d item(s) this release no longer ships)\n", mapping.label, destination, removed)
		} else {
			fmt.Printf("✓ Installed %s to %s\n", mapping.label, destination)
		}
	}
}

// detectConflicts checks which embedded files would conflict with existing files
func detectConflicts() []string {
	var conflicts []string

	// Define the file mappings that will be installed
	fileMappings := []struct {
		embeddedFS   embed.FS
		srcDir       string
		dstDir       string
		skipIfExists bool // Workflow files are never overwritten
	}{
		{templateFiles, "templates/specs", ".smaqit/templates/specs", false},
		{designTemplateFiles, "templates/designs", ".smaqit/templates/designs", false},
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

const managedToolsGitignoreRule = ".smaqit/tools/"

// ensureManagedToolsGitignore prevents the bundled, project-local PlantUML
// runtime from becoming consumer source control content. It appends only the
// exact managed rule, preserving all existing user-owned .gitignore content.
func ensureManagedToolsGitignore(projectRoot string) error {
	path := filepath.Join(projectRoot, ".gitignore")
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			if strings.TrimSpace(line) == managedToolsGitignoreRule {
				return nil
			}
		}
	}

	newline := "\n"
	if bytes.Contains(content, []byte("\r\n")) {
		newline = "\r\n"
	}
	updated := append([]byte(nil), content...)
	if len(updated) > 0 && !bytes.HasSuffix(updated, []byte("\n")) {
		updated = append(updated, newline...)
	}
	updated = append(updated, managedToolsGitignoreRule...)
	updated = append(updated, newline...)
	return os.WriteFile(path, updated, 0o644)
}

func cmdInit(targetDir string) {
	// The toolchain is mandatory. Fail before creating the target so an
	// unsupported host never receives a partial smaqit installation.
	if err := preflightDesignTools(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := preflightDesignMCPConfigs(targetDir); err != nil {
		fmt.Fprintf(os.Stderr, "DESIGN-TOOLCHAIN-UNAVAILABLE: %v\n", err)
		os.Exit(1)
	}

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
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error resolving project directory: %v\n", err)
		os.Exit(1)
	}

	// Handle both reinstallation and an initial install that would overwrite an exact
	// owned destination. Unrelated files in shared platform directories are not conflicts.
	_, existingErr := os.Stat(".smaqit")
	existingInstall := existingErr == nil
	conflicts := detectConflicts()
	if existingInstall || len(conflicts) > 0 {
		if existingInstall {
			fmt.Println("Existing smaqit installation detected.")
		} else {
			fmt.Println("Existing files conflict with the smaqit installation.")
		}
		fmt.Println()

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

	fmt.Printf("Initializing smaqit project in %s...\n", projectDir)

	// Create directory structure
	dirs := []string{
		".smaqit/templates/specs",
		".smaqit/templates/designs",
		".smaqit/reports",
		"specs/business",
		"specs/functional",
		"specs/stack",
		"specs/infrastructure",
		"specs/coverage",
		"docs/designs/business",
		"docs/designs/functional",
		"docs/designs/stack",
		"docs/designs/infrastructure",
		"docs/designs/coverage",
		"docs/designs/design-sequence",
		".vscode",
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
	if err := copyEmbeddedDir(designTemplateFiles, "templates/designs", ".smaqit/templates/designs"); err != nil {
		fmt.Printf("Error copying design templates: %v\n", err)
		os.Exit(1)
	}
	if _, err := materializeDesignTools("."); err != nil {
		fmt.Printf("Error installing mandatory PlantUML runtime: %v\n", err)
		os.Exit(1)
	}
	if err := ensureManagedToolsGitignore("."); err != nil {
		fmt.Printf("Error updating .gitignore for managed PlantUML runtime: %v\n", err)
		os.Exit(1)
	}
	if err := installDesignMCPConfigs("."); err != nil {
		fmt.Printf("Error installing mandatory PlantUML MCP configuration: %v\n", err)
		os.Exit(1)
	}

	// Copy workflow files
	if err := copyEmbeddedDir(workflowFiles, "templates/workflows", ".github/workflows"); err != nil {
		fmt.Printf("Error copying workflow files: %v\n", err)
		os.Exit(1)
	}

	// Install project-instructions files: AGENTS.md (read natively by GitHub Copilot and Codex)
	// with a thin CLAUDE.md pointing at it (Claude Code does not read AGENTS.md on its own).
	// Only smaqit's marked section is ever written: appended when absent, refreshed to the
	// current template when present, so updates reach existing projects. Content outside
	// the markers is never touched.
	agentsStatus := installInstructionsFileOrExit(agentsMdTemplate, "templates/AGENTS.md.template", "AGENTS.md")
	claudeStatus := installInstructionsFileOrExit(claudeMdTemplate, "templates/CLAUDE.md.template", "CLAUDE.md")
	// Project-owned deployment slot bindings: written once, never overwritten,
	// because update re-runs init.
	slotsStatus, err := installSlotsSkeleton(".")
	if err != nil {
		fmt.Printf("Error installing %s: %v\n", slotsConfigPath, err)
		os.Exit(1)
	}

	// This is an idempotent local gate. cmdInit also backs project updates, so
	// every initialization or update proves the generated MCP configuration and
	// its public stdio wrapper before reporting success. Client-side tool
	// discovery remains a separate authoring-session precondition.
	if err := verifyPlantUMLMCP("."); err != nil {
		fmt.Printf("Error verifying mandatory PlantUML MCP transport: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Created .smaqit/ directory structure")
	fmt.Println("✓ Copied templates")
	fmt.Println("✓ Installed bundled PlantUML MCP and PNG rendering runtime")
	fmt.Println("✓ Configured project-local PlantUML MCP discovery for VS Code, Claude Code, and Codex")
	fmt.Println("✓ Verified PlantUML MCP configuration and local stdio transport")
	fmt.Println("✓ Global agent and skill installation is managed by the installer script")
	fmt.Println("✓ Copied workflow files")
	fmt.Printf("✓ AGENTS.md %s, CLAUDE.md %s\n", agentsStatus, claudeStatus)
	fmt.Printf("✓ %s %s\n", slotsConfigPath, slotsStatus)
	fmt.Printf("✓ Initialized smaqit %s\n\n", Version)
	fmt.Println("Next steps:")
	fmt.Println("  1. Open or restart GitHub Copilot chat, Claude Code, or Codex in this project, then confirm smaqit-plantuml exposes its two tools to specification agents")
	fmt.Println("  2. Invoke or ask to spawn smaqit.development for the Development phase")
	fmt.Println("  3. Or type '/smaqit.business' to begin with business specifications only (GitHub Copilot only — see 'smaqit help')")
}

const (
	instructionsMarkerBegin = "<!-- smaqit:instructions:begin -->"
	instructionsMarkerEnd   = "<!-- smaqit:instructions:end -->"
)

// errMalformedInstructionsBlock reports a destination whose smaqit markers can't be
// trusted to delimit smaqit's section, so the file is left exactly as it is.
var errMalformedInstructionsBlock = errors.New("smaqit:instructions markers are malformed")

// installInstructionsFile installs a project-instructions file (AGENTS.md, CLAUDE.md):
//   - Destination absent: write the template as a new file.
//   - Destination exists, no smaqit marker found: append the template to the end (never
//     overwrite user content — the destination may already hold the user's own instructions).
//   - Destination exists with exactly one begin marker followed by one end marker: replace
//     everything from the begin marker through the end marker with the current template, so
//     template changes reach existing projects. Content outside the markers is preserved
//     byte for byte, and an already-current section is not rewritten.
//   - Any other marker layout (begin without end, reversed, duplicated): leave the file
//     untouched and return errMalformedInstructionsBlock — never guess where a section ends.
//
// Returns a short status word ("created", "appended", "updated", "up to date") for reporting.
func installInstructionsFile(embeddedFS embed.FS, srcPath, dstPath string) (string, error) {
	content, err := embeddedFS.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", srcPath, err)
	}

	existing, err := os.ReadFile(dstPath)
	if os.IsNotExist(err) {
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return "", fmt.Errorf("writing %s: %w", dstPath, err)
		}
		return "created", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", dstPath, err)
	}

	text := string(existing)
	begins := strings.Count(text, instructionsMarkerBegin)
	ends := strings.Count(text, instructionsMarkerEnd)
	if begins > 0 || ends > 0 {
		start := strings.Index(text, instructionsMarkerBegin)
		end := strings.Index(text, instructionsMarkerEnd)
		if begins != 1 || ends != 1 || end < start {
			return "", fmt.Errorf("%w in %s (expected one begin marker followed by one end marker)", errMalformedInstructionsBlock, dstPath)
		}
		block := strings.TrimRight(string(content), "\r\n")
		if !strings.HasPrefix(block, instructionsMarkerBegin) || !strings.HasSuffix(block, instructionsMarkerEnd) {
			return "", fmt.Errorf("%s must start with %s and end with %s", srcPath, instructionsMarkerBegin, instructionsMarkerEnd)
		}
		if strings.HasPrefix(text[start+len(instructionsMarkerBegin):], "\r\n") {
			block = strings.ReplaceAll(block, "\n", "\r\n")
		}
		refreshed := text[:start] + block + text[end+len(instructionsMarkerEnd):]
		if refreshed == text {
			return "up to date", nil
		}
		if err := os.WriteFile(dstPath, []byte(refreshed), 0644); err != nil {
			return "", fmt.Errorf("writing %s: %w", dstPath, err)
		}
		return "updated", nil
	}

	appended := text
	if !strings.HasSuffix(appended, "\n") {
		appended += "\n"
	}
	appended += "\n" + string(content)

	if err := os.WriteFile(dstPath, []byte(appended), 0644); err != nil {
		return "", fmt.Errorf("writing %s: %w", dstPath, err)
	}
	return "appended", nil
}

// installInstructionsFileOrExit runs installInstructionsFile for cmdInit. A malformed smaqit
// section is a warning, not a failure: the file is left as the user has it and init continues.
func installInstructionsFileOrExit(embeddedFS embed.FS, srcPath, dstPath string) string {
	status, err := installInstructionsFile(embeddedFS, srcPath, dstPath)
	if errors.Is(err, errMalformedInstructionsBlock) {
		fmt.Printf("⚠ %s left unchanged: %v. Fix the markers by hand, then run 'smaqit init' again.\n", dstPath, err)
		return "left unchanged"
	}
	if err != nil {
		fmt.Printf("Error installing %s: %v\n", dstPath, err)
		os.Exit(1)
	}
	return status
}

// copyEmbeddedDir copies files from an embedded FS to a target directory
// If dstDir contains "templates/specs", performs version substitution
func copyEmbeddedDir(embeddedFS fs.FS, srcDir, dstDir string) error {
	substituteVersion := strings.Contains(dstDir, "templates/specs")

	return fs.WalkDir(embeddedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Read embedded file
		content, err := fs.ReadFile(embeddedFS, path)
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

// removeEmbeddedFiles removes only files represented in an embedded source tree.
// It is used for shared Codex directories, where unrelated user or extension files
// must remain untouched.
func removeEmbeddedFiles(embeddedFS embed.FS, srcDir, dstDir string) (int, error) {
	removed := 0
	err := fs.WalkDir(embeddedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath := strings.TrimPrefix(path, srcDir+"/")
		dstPath := filepath.Join(dstDir, relPath)
		if err := os.Remove(dstPath); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("removing %s: %w", dstPath, err)
		}
		removed++
		return nil
	})
	return removed, err
}

// removeEmbeddedSkillDirs removes the exact files represented by embedded skill
// packages, then prunes only directories left empty. User-added files within a
// smaqit skill directory therefore survive uninstall.
func removeEmbeddedSkillDirs(embeddedFS embed.FS, srcDir, dstDir string) (int, error) {
	entries, err := embeddedFS.ReadDir(srcDir)
	if err != nil {
		return 0, err
	}

	presentPackages := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dstPath := filepath.Join(dstDir, entry.Name())
		if _, err := os.Stat(dstPath); err == nil {
			presentPackages++
		} else if !os.IsNotExist(err) {
			return presentPackages, fmt.Errorf("checking %s: %w", dstPath, err)
		}
	}

	if _, err := removeEmbeddedFiles(embeddedFS, srcDir, dstDir); err != nil {
		return presentPackages, err
	}

	var ownedDirs []string
	err = fs.WalkDir(embeddedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != srcDir {
			relPath := strings.TrimPrefix(path, srcDir+"/")
			ownedDirs = append(ownedDirs, filepath.Join(dstDir, filepath.FromSlash(relPath)))
		}
		return nil
	})
	if err != nil {
		return presentPackages, err
	}

	// Children must be considered before their parents.
	sort.Slice(ownedDirs, func(i, j int) bool {
		return strings.Count(ownedDirs[i], string(filepath.Separator)) >
			strings.Count(ownedDirs[j], string(filepath.Separator))
	})
	for _, dir := range ownedDirs {
		if _, err := removeDirIfEmpty(dir); err != nil {
			return presentPackages, fmt.Errorf("pruning %s: %w", dir, err)
		}
	}

	return presentPackages, nil
}

func removeDirIfEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if len(entries) != 0 {
		return false, nil
	}
	if err := os.Remove(path); err != nil {
		return false, err
	}
	return true, nil
}

func cmdUninstall() {
	// Check if .smaqit exists
	if _, err := os.Stat(".smaqit"); os.IsNotExist(err) {
		fmt.Println("No smaqit installation found in this directory")
		os.Exit(0)
	}

	// Prompt for confirmation
	fmt.Println("This will remove:")
	fmt.Println("  • .smaqit/ (except the project-owned .smaqit/slots.yml)")
	fmt.Println("  • smaqit-owned global agents, commands, and skills")
	fmt.Println("  • smaqit-owned PlantUML MCP registrations")
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

	for _, remove := range []func(string) error{
		removeVSCodeMCPConfig,
		removeClaudeMCPConfig,
		removeCodexMCPConfig,
	} {
		if err := remove("."); err != nil {
			fmt.Printf("Error removing smaqit PlantUML MCP configuration: %v\n", err)
			errors++
		}
	}
	if errors == 0 {
		fmt.Println("✓ Removed smaqit-owned PlantUML MCP registrations")
	}

	if kept, err := removeSmaqitDirPreservingSlots("."); err != nil {
		fmt.Printf("Error removing .smaqit/: %v\n", err)
		errors++
	} else if kept {
		fmt.Printf("✓ Removed .smaqit/ (kept project-owned %s)\n", slotsConfigPath)
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
	fmt.Println("✓ Kept docs/designs/ (user design artifacts)")

	globalMappings := []struct {
		embeddedFS embed.FS
		srcDir     string
		kind       string
		skills     bool
	}{
		{agentFiles, "agents-copilot", "copilot-agents", false},
		{skillFilesShared, "skills-shared", "shared-skills", true},
		{claudeAgentFiles, "agents-claude", "claude-agents", false},
		{claudeCommandFiles, "commands-claude", "claude-commands", false},
		{skillFilesClaude, "skills-claude", "claude-skills", true},
		{codexAgentFiles, "agents-codex", "codex-agents", false},
	}
	for _, mapping := range globalMappings {
		dir, err := resolveGlobalDir(mapping.kind)
		if err != nil {
			fmt.Printf("Error resolving global uninstall path: %v\n", err)
			errors++
			continue
		}
		if mapping.skills {
			_, err = removeEmbeddedSkillDirs(mapping.embeddedFS, mapping.srcDir, dir)
		} else {
			_, err = removeEmbeddedFiles(mapping.embeddedFS, mapping.srcDir, dir)
		}
		if err == nil {
			err = uninstallGlobalManifest(mapping.embeddedFS, mapping.srcDir, dir)
		}
		if err != nil {
			fmt.Printf("Error removing global smaqit payload from %s: %v\n", dir, err)
			errors++
			continue
		}
		if _, err := removeDirIfEmpty(dir); err != nil {
			fmt.Printf("Error pruning %s/: %v\n", dir, err)
			errors++
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
		".smaqit/templates/designs",
		"specs/business",
		"specs/functional",
		"specs/stack",
		"specs/infrastructure",
		"specs/coverage",
		"docs/designs/business",
		"docs/designs/functional",
		"docs/designs/stack",
		"docs/designs/infrastructure",
		"docs/designs/coverage",
		"docs/designs/design-sequence",
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
	if _, err := ensureDesignTools("."); err != nil {
		fmt.Printf("✗ %v\n", err)
		errors++
	} else if err := validateDesignMCPConfigs("."); err != nil {
		fmt.Printf("✗ DESIGN-TOOLCHAIN-UNAVAILABLE: %v\n", err)
		errors++
	} else {
		fmt.Println("✓ Mandatory PlantUML runtime and MCP configuration are valid")
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

	if err := validateDesigns(""); err != nil {
		fmt.Printf("✗ %v\n", err)
		errors++
	}

	if set, envErrs := loadInfraEnvironments("."); len(envErrs) > 0 || set != nil {
		if len(envErrs) == 0 {
			envErrs = validateInfraEnvironments(set, "")
		}
		for _, err := range envErrs {
			fmt.Printf("✗ %v\n", err)
		}
		errors += len(envErrs)
		if len(envErrs) == 0 {
			fmt.Printf("✓ %s Environments table is valid\n", set.Spec)
		}
	}

	if _, err := os.Stat(slotsConfigPath); err == nil {
		if slotErrs := validateSlotBindings("."); len(slotErrs) > 0 {
			for _, err := range slotErrs {
				fmt.Printf("✗ %v\n", err)
			}
			errors += len(slotErrs)
		} else {
			fmt.Printf("✓ %s is valid\n", slotsConfigPath)
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
	fmt.Println("=====================")
	fmt.Println()

	// Scan all specs
	allSpecs, err := scanSpecs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning specs: %v\n", err)
		os.Exit(1)
	}

	// Count specs by layer and status
	layerCounts := make(map[string]int)
	activeLayerCounts := make(map[string]int)
	layerStatusCounts := make(map[string]map[string]int)

	for layer, specs := range allSpecs {
		layerCounts[layer] = len(specs)
		layerStatusCounts[layer] = make(map[string]int)

		for _, spec := range specs {
			layerStatusCounts[layer][spec.Frontmatter.Status]++
			if spec.Frontmatter.Status != "deprecated" {
				activeLayerCounts[layer]++
			}
		}
	}

	totalSpecs := 0
	designBlocked := 0
	for _, count := range layerCounts {
		totalSpecs += count
	}
	for _, specs := range allSpecs {
		for _, spec := range specs {
			if spec.Frontmatter.Status != "deprecated" && !spec.DesignReady {
				designBlocked++
			}
		}
	}

	// Determine phase completion status
	// A phase is ONLY complete when:
	// 1. ALL required layers have at least one spec, AND
	// 2. ALL specs in those layers have reached the target status

	// Phase 1: Develop (requires business, functional, stack)
	developSpecs := getPhaseSpecs(allSpecs, "develop")
	developImplemented := 0
	for _, spec := range developSpecs {
		if spec.DesignReady && (spec.Frontmatter.Status == "implemented" ||
			spec.Frontmatter.Status == "deployed" ||
			spec.Frontmatter.Status == "validated") {
			developImplemented++
		}
	}
	// Require all three layers present
	hasAllDevelopLayers := activeLayerCounts["business"] > 0 &&
		activeLayerCounts["functional"] > 0 &&
		activeLayerCounts["stack"] > 0
	developCompleted := hasAllDevelopLayers &&
		len(developSpecs) > 0 &&
		developImplemented == len(developSpecs)

	// Phase 2: Deploy (requires infrastructure)
	deploySpecs := getPhaseSpecs(allSpecs, "deploy")
	deployDeployed := 0
	for _, spec := range deploySpecs {
		if spec.DesignReady && (spec.Frontmatter.Status == "deployed" ||
			spec.Frontmatter.Status == "validated") {
			deployDeployed++
		}
	}
	deployCompleted := activeLayerCounts["infrastructure"] > 0 &&
		len(deploySpecs) > 0 &&
		deployDeployed == len(deploySpecs)

	// Phase 3: Validate (requires coverage)
	validateSpecs := getPhaseSpecs(allSpecs, "validate")
	validateValidated := 0
	for _, spec := range validateSpecs {
		if spec.DesignReady && spec.Frontmatter.Status == "validated" {
			validateValidated++
		}
	}
	validateCompleted := activeLayerCounts["coverage"] > 0 &&
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
	if designBlocked > 0 {
		fmt.Printf("Design gates: %d active specification(s) missing a current validated design; run 'smaqit validate' for migration diagnostics\n", designBlocked)
	} else if totalSpecs > 0 {
		fmt.Println("Design gates: ✓ all active specifications have current validated designs")
	}

	// Next steps based on actual spec content and status
	fmt.Println("\nNext steps:")
	if designBlocked > 0 {
		fmt.Println("  • Migrate or repair the linked design pairs, visually review their PNGs, then run 'smaqit design validate'")
	}

	// Phase 1: Develop
	if !developCompleted {
		hasAnyPhase1 := activeLayerCounts["business"] > 0 || activeLayerCounts["functional"] > 0 || activeLayerCounts["stack"] > 0

		if !hasAnyPhase1 {
			fmt.Println("  • Run the smaqit.business agent to start with business specifications")
		} else {
			// Suggest missing layers first
			if activeLayerCounts["business"] == 0 {
				fmt.Println("  • Run the smaqit.business agent to add business specifications")
			} else if activeLayerCounts["functional"] == 0 {
				fmt.Println("  • Run the smaqit.functional agent to add functional specifications")
			} else if activeLayerCounts["stack"] == 0 {
				fmt.Println("  • Run the smaqit.stack agent to add technical stack specifications")
			} else {
				// All Phase 1 specs exist
				fmt.Println("  • Run 'smaqit plan --phase=develop' to see work plan")
				fmt.Println("  • Run the smaqit.development agent to implement from specifications")
			}
		}
	} else if !deployCompleted {
		// Phase 2: Deploy
		if activeLayerCounts["infrastructure"] == 0 {
			fmt.Println("  • Run the smaqit.infrastructure agent to define infrastructure specifications")
		} else {
			fmt.Println("  • Run 'smaqit plan --phase=deploy' to see work plan")
			fmt.Println("  • Run the smaqit.deployment agent to deploy the implementation")
		}
	} else if !validateCompleted {
		// Phase 3: Validate
		if activeLayerCounts["coverage"] == 0 {
			fmt.Println("  • Run the smaqit.coverage agent to define test coverage specifications")
		} else {
			fmt.Println("  • Run 'smaqit plan --phase=validate' to see work plan")
			fmt.Println("  • Run the smaqit.validation agent to validate the deployment")
		}
	} else {
		fmt.Println("  • All phases complete. Run the smaqit.development agent to iterate or extend.")
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
