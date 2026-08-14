package main

import (
	"strings"
	"testing"
)

func TestValidatePhaseDesignReadiness(t *testing.T) {
	ready := Spec{Path: "specs/business/ready.md", DesignReady: true}
	blocked := Spec{
		Path:        "specs/functional/blocked.md",
		DesignReady: false,
		DesignError: "DESIGN-ARTIFACT-STALE: current source/image hashes do not match",
	}

	if err := validatePhaseDesignReadiness([]Spec{ready}); err != nil {
		t.Fatalf("ready phase was blocked: %v", err)
	}

	err := validatePhaseDesignReadiness([]Spec{ready, blocked})
	if err == nil {
		t.Fatal("phase with a failed design gate was accepted")
	}
	if !strings.Contains(err.Error(), blocked.Path) || !strings.Contains(err.Error(), "DESIGN-ARTIFACT-STALE") {
		t.Fatalf("phase gate lost actionable context: %v", err)
	}
}

func TestGetPhaseDesignGateSpecsIncludesConsumedUpstreamLayers(t *testing.T) {
	draft := func(layer string) Spec {
		return Spec{Layer: layer, Frontmatter: SpecFrontmatter{Status: "draft"}}
	}
	allSpecs := map[string][]Spec{
		"business":       {draft("business")},
		"functional":     {draft("functional")},
		"stack":          {draft("stack")},
		"infrastructure": {draft("infrastructure")},
		"coverage":       {draft("coverage")},
	}

	assertLayers := func(phase string, want []string) {
		t.Helper()
		gotSpecs := getPhaseDesignGateSpecs(allSpecs, phase)
		if len(gotSpecs) != len(want) {
			t.Fatalf("%s gate returned %d specs, want %d", phase, len(gotSpecs), len(want))
		}
		for i, spec := range gotSpecs {
			if spec.Layer != want[i] {
				t.Fatalf("%s gate layer %d = %s, want %s", phase, i, spec.Layer, want[i])
			}
		}
	}

	assertLayers("develop", []string{"business", "functional", "stack"})
	assertLayers("deploy", []string{"stack", "infrastructure"})
	assertLayers("validate", []string{"business", "functional", "stack", "infrastructure", "coverage"})
}

func TestGetPhaseDesignGateSpecsScopesToCurrentCycle(t *testing.T) {
	// An old, unrelated, un-paired spec (status implemented) sits alongside a
	// feature's new draft spec in the same layer. The gate must scope to the
	// current cycle so the legacy spec never blocks an unrelated feature's
	// phase plan — the exact regression reported in task 109.
	legacyUnpaired := Spec{
		Path:        "specs/business/legacy-unrelated.md",
		Layer:       "business",
		Frontmatter: SpecFrontmatter{Status: "implemented"},
		DesignReady: false,
		DesignError: "DESIGN-REFERENCE-MISSING: no ## Design References section",
	}
	deployedReady := Spec{
		Path:        "specs/business/deployed-ready.md",
		Layer:       "business",
		Frontmatter: SpecFrontmatter{Status: "deployed"},
		DesignReady: true,
	}
	deprecated := Spec{
		Path:        "specs/business/deprecated.md",
		Layer:       "business",
		Frontmatter: SpecFrontmatter{Status: "deprecated"},
		DesignReady: false,
		DesignError: "DESIGN-REFERENCE-MISSING: no ## Design References section",
	}
	failedRetry := Spec{
		Path:        "specs/business/failed-retry.md",
		Layer:       "business",
		Frontmatter: SpecFrontmatter{Status: "failed"},
		DesignReady: true,
	}
	newDraft := Spec{
		Path:        "specs/business/new-feature.md",
		Layer:       "business",
		Frontmatter: SpecFrontmatter{Status: "draft"},
		DesignReady: true,
	}
	allSpecs := map[string][]Spec{
		"business": {legacyUnpaired, deployedReady, deprecated, failedRetry, newDraft},
	}

	gotSpecs := getPhaseDesignGateSpecs(allSpecs, "develop")
	if len(gotSpecs) != 2 {
		t.Fatalf("gate returned %d specs, want 2 (failed-retry, new-feature): %+v", len(gotSpecs), gotSpecs)
	}
	for _, spec := range gotSpecs {
		if spec.Path == legacyUnpaired.Path {
			t.Fatalf("gate included legacy unrelated spec %s — scoping regression", spec.Path)
		}
		if spec.Path == deployedReady.Path {
			t.Fatalf("gate included already-deployed spec %s — scoping regression", spec.Path)
		}
		if spec.Path == deprecated.Path {
			t.Fatalf("gate included deprecated spec %s", spec.Path)
		}
	}

	// The scoped set must pass the gate cleanly — the legacy spec's missing
	// design reference never even reaches validatePhaseDesignReadiness.
	if err := validatePhaseDesignReadiness(gotSpecs); err != nil {
		t.Fatalf("scoped gate set unexpectedly blocked: %v", err)
	}
}

func TestValidatePhaseDesignReadinessAggregatesAllFailures(t *testing.T) {
	blockedA := Spec{
		Path:        "specs/business/a.md",
		DesignReady: false,
		DesignError: "DESIGN-REFERENCE-MISSING: no ## Design References section",
	}
	blockedB := Spec{
		Path:        "specs/functional/b.md",
		DesignReady: false,
		DesignError: "DESIGN-ARTIFACT-STALE: current source/image hashes do not match",
	}
	ready := Spec{Path: "specs/stack/c.md", DesignReady: true}

	err := validatePhaseDesignReadiness([]Spec{blockedA, ready, blockedB})
	if err == nil {
		t.Fatal("phase with two failed design gates was accepted")
	}
	msg := err.Error()
	for _, want := range []string{blockedA.Path, blockedB.Path, "DESIGN-REFERENCE-MISSING", "DESIGN-ARTIFACT-STALE"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("aggregated error missing %q: %v", want, err)
		}
	}
	if strings.Contains(msg, ready.Path) {
		t.Fatalf("aggregated error mentioned a ready spec: %v", err)
	}
}
