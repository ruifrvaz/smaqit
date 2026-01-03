package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SpecFrontmatter represents the YAML frontmatter in specification files
type SpecFrontmatter struct {
	ID            string    `yaml:"id"`
	Status        string    `yaml:"status"`
	Created       time.Time `yaml:"created"`
	Implemented   time.Time `yaml:"implemented,omitempty"`
	Deployed      time.Time `yaml:"deployed,omitempty"`
	Validated     time.Time `yaml:"validated,omitempty"`
	PromptVersion string    `yaml:"prompt_version,omitempty"`
}

// Spec represents a specification file with its metadata
type Spec struct {
	Path        string
	Layer       string
	Frontmatter SpecFrontmatter
}

// scanSpecs scans all spec directories and parses frontmatter from markdown files
// Returns a map of layer -> specs
func scanSpecs() (map[string][]Spec, error) {
	layers := []string{"business", "functional", "stack", "infrastructure", "coverage"}
	result := make(map[string][]Spec)

	for _, layer := range layers {
		specDir := filepath.Join("specs", layer)

		// Check if directory exists
		if _, err := os.Stat(specDir); os.IsNotExist(err) {
			result[layer] = []Spec{}
			continue
		}

		entries, err := os.ReadDir(specDir)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", specDir, err)
		}

		layerSpecs := []Spec{}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			specPath := filepath.Join(specDir, entry.Name())
			frontmatter, err := parseSpecFrontmatter(specPath)
			if err != nil {
				// Warn but continue
				fmt.Fprintf(os.Stderr, "⚠ Warning: %s: %v (treating as draft)\n", specPath, err)
				// Create default frontmatter
				frontmatter = &SpecFrontmatter{
					ID:      "UNKNOWN",
					Status:  "draft",
					Created: time.Now(),
				}
			}

			layerSpecs = append(layerSpecs, Spec{
				Path:        specPath,
				Layer:       layer,
				Frontmatter: *frontmatter,
			})
		}

		result[layer] = layerSpecs
	}

	return result, nil
}

// parseSpecFrontmatter extracts and parses YAML frontmatter from a markdown file
func parseSpecFrontmatter(path string) (*SpecFrontmatter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Check for opening fence
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty file")
	}

	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "---" {
		return nil, fmt.Errorf("no frontmatter (expected '---' on first line)")
	}

	// Read frontmatter content until closing fence
	var frontmatterLines []string
	foundClosingFence := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundClosingFence = true
			break
		}
		frontmatterLines = append(frontmatterLines, line)
	}

	if !foundClosingFence {
		return nil, fmt.Errorf("no closing frontmatter fence")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Parse YAML
	frontmatterYAML := strings.Join(frontmatterLines, "\n")
	var frontmatter SpecFrontmatter

	if err := yaml.Unmarshal([]byte(frontmatterYAML), &frontmatter); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Validate required fields
	if frontmatter.ID == "" {
		return nil, fmt.Errorf("missing 'id' field")
	}
	if frontmatter.Status == "" {
		return nil, fmt.Errorf("missing 'status' field")
	}

	return &frontmatter, nil
}

// filterSpecsByStatus filters specs based on status for a given phase
// Returns list of specs to process
func filterSpecsByStatus(specs []Spec, phase string, regen bool) []Spec {
	if regen {
		// Regeneration mode: return all specs
		return specs
	}

	// Incremental mode: filter by status
	var toProcess []Spec

	for _, spec := range specs {
		switch phase {
		case "develop":
			// Process draft and failed, skip implemented/deployed/validated
			if spec.Frontmatter.Status == "draft" || spec.Frontmatter.Status == "failed" {
				toProcess = append(toProcess, spec)
			}
		case "deploy":
			// Process draft and failed, skip deployed/validated
			if spec.Frontmatter.Status == "draft" || spec.Frontmatter.Status == "failed" {
				toProcess = append(toProcess, spec)
			}
		case "validate":
			// Process draft and failed, skip validated
			if spec.Frontmatter.Status == "draft" || spec.Frontmatter.Status == "failed" {
				toProcess = append(toProcess, spec)
			}
		}
	}

	return toProcess
}

// getPhaseSpecs returns all specs for a given phase's layers
func getPhaseSpecs(allSpecs map[string][]Spec, phase string) []Spec {
	var specs []Spec

	switch phase {
	case "develop":
		specs = append(specs, allSpecs["business"]...)
		specs = append(specs, allSpecs["functional"]...)
		specs = append(specs, allSpecs["stack"]...)
	case "deploy":
		specs = allSpecs["infrastructure"]
	case "validate":
		specs = allSpecs["coverage"]
	}

	return specs
}
