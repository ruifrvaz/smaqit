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
	allSpecs := map[string][]Spec{
		"business":       {{Layer: "business"}},
		"functional":     {{Layer: "functional"}},
		"stack":          {{Layer: "stack"}},
		"infrastructure": {{Layer: "infrastructure"}},
		"coverage":       {{Layer: "coverage"}},
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
