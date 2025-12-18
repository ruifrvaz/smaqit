package main

import (
	"fmt"
	"os"
)

// Version is set via ldflags during build: -X main.Version=$(VERSION)
var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		cmdInit()
	case "develop":
		cmdDevelop()
	case "deploy":
		cmdDeploy()
	case "validate":
		cmdValidate()
	case "status":
		cmdStatus()
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
  init       Scaffold .smaqit/ and .github/agents/
  develop    Phase 1: business → functional → stack → build
  deploy     Phase 2: infrastructure → deploy
  validate   Phase 3: coverage → validate
  status     Show current state
  version    Show version`)
}

func cmdInit() {
	// TODO: Create .smaqit/ directory structure:
	//   .smaqit/
	//   ├── framework/
	//   ├── templates/
	//   └── specs/
	//       ├── business/
	//       ├── functional/
	//       ├── stack/
	//       ├── infrastructure/
	//       └── coverage/
	//
	// TODO: Copy framework/ → .smaqit/framework/ (entire directory)
	//   - SMAQIT.md, LAYERS.md, PHASES.md
	//   - TEMPLATES.md, AGENTS.md, ARTIFACTS.md
	//
	// TODO: Copy templates/ → .smaqit/templates/ (entire directory)
	//   - specs/*.template.md
	//   - agents/*.template.md
	//
	// TODO: Create .github/agents/ directory
	// TODO: Copy agents/*.agent.md → .github/agents/
	fmt.Println("init: scaffold .smaqit/ and .github/agents/")
}

func cmdDevelop() {
	// TODO: Parse flags (--spec, --layer, --build)
	// TODO: If --spec or default:
	//   - Invoke smaqit.business agent
	//   - Invoke smaqit.functional agent
	//   - Invoke smaqit.stack agent
	// TODO: If --build or default:
	//   - Invoke smaqit.development agent
	fmt.Println("develop: business → functional → stack → build")
}

func cmdDeploy() {
	// TODO: Parse flags (--spec, --run)
	// TODO: If --spec or default:
	//   - Invoke smaqit.infrastructure agent
	// TODO: If --run or default:
	//   - Invoke smaqit.deployment agent
	fmt.Println("deploy: infrastructure → deploy")
}

func cmdValidate() {
	// TODO: Parse flags (--spec, --run)
	// TODO: If --spec or default:
	//   - Invoke smaqit.coverage agent
	// TODO: If --run or default:
	//   - Invoke smaqit.validation agent
	fmt.Println("validate: coverage → validate")
}

func cmdStatus() {
	// TODO: Check .smaqit/specs/ for existing specs
	// TODO: Report which layers have specs
	// TODO: Report which phases have been run
	fmt.Println("status: show current state")
}
