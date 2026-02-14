package model

import (
	"strings"
	"testing"
)

func TestSelectModel(t *testing.T) {
	r := NewRouter()
	model, err := r.Select(RouteRequest{UseCase: "summary", Budget: "low"})
	if err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if model != "gpt-4.1-mini" {
		t.Fatalf("unexpected model: %s", model)
	}
}

func TestPolicyPacks(t *testing.T) {
	r := NewRouter()
	packs := r.Packs()
	if len(packs) != 3 {
		t.Fatalf("expected three packs, got %d", len(packs))
	}
}

func TestSelectQualityFirst(t *testing.T) {
	r := NewRouter()
	model, err := r.Select(RouteRequest{UseCase: "decision", PolicyPack: "quality-first"})
	if err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if model != "gpt-4.1" {
		t.Fatalf("unexpected model: %s", model)
	}
}

// AC1: All three policy packs are supported and selectable.
func TestAllPolicyPacksSupported(t *testing.T) {
	r := NewRouter()
	packs := []string{"cost-first", "quality-first", "balanced"}
	for _, pack := range packs {
		t.Run(pack, func(t *testing.T) {
			_, err := r.Select(RouteRequest{UseCase: "test", PolicyPack: pack})
			if err != nil {
				t.Fatalf("policy pack %q should be supported: %v", pack, err)
			}
		})
	}
}

// AC1: Each policy pack returns a distinct routing decision.
func TestPolicyPacksReturnDistinctDecisions(t *testing.T) {
	r := NewRouter()
	costModel, _ := r.Select(RouteRequest{UseCase: "test", PolicyPack: "cost-first"})
	qualityModel, _ := r.Select(RouteRequest{UseCase: "test", PolicyPack: "quality-first"})
	if costModel == qualityModel {
		t.Fatalf("cost-first and quality-first should route to different models, both got %q", costModel)
	}
}

// AC1: Packs() returns defensive copy with name and description.
func TestPacksReturnsCompleteMetadata(t *testing.T) {
	r := NewRouter()
	packs := r.Packs()
	for _, p := range packs {
		if p.Name == "" {
			t.Fatal("pack missing name")
		}
		if p.Description == "" {
			t.Fatalf("pack %q missing description", p.Name)
		}
	}
	// Mutate returned slice and verify no effect.
	packs[0].Name = "hacked"
	fresh := r.Packs()
	if fresh[0].Name == "hacked" {
		t.Fatal("Packs() did not return a defensive copy")
	}
}

// AC2: Routing is transparent — Select returns model name string, caller doesn't need to know internals.
func TestRoutingIsTransparent(t *testing.T) {
	r := NewRouter()
	model, err := r.Select(RouteRequest{UseCase: "analyze", PolicyPack: "balanced"})
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	// Caller gets a model identifier, not internal routing details.
	if model == "" {
		t.Fatal("routing should return a model name")
	}
}

// AC2: Decide returns structured decision with reason for observability.
func TestDecideReturnsStructuredDecision(t *testing.T) {
	r := NewRouter()
	decision, err := r.Decide(RouteRequest{UseCase: "review", PolicyPack: "cost-first"})
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if decision.PolicyPack != "cost-first" {
		t.Fatalf("expected policy pack 'cost-first', got %q", decision.PolicyPack)
	}
	if decision.Model == "" {
		t.Fatal("decision missing model")
	}
	if decision.Reason == "" {
		t.Fatal("decision missing reason")
	}
}

// AC1: Default policy pack is balanced when not specified.
func TestDefaultPolicyPack(t *testing.T) {
	r := NewRouter()
	decision, err := r.Decide(RouteRequest{UseCase: "test"})
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if decision.PolicyPack != "balanced" {
		t.Fatalf("expected default policy pack 'balanced', got %q", decision.PolicyPack)
	}
}

// AC1: Unknown policy pack returns error.
func TestUnknownPolicyPackReturnsError(t *testing.T) {
	r := NewRouter()
	_, err := r.Select(RouteRequest{UseCase: "test", PolicyPack: "turbo"})
	if err == nil {
		t.Fatal("expected error for unknown policy pack")
	}
	if !strings.Contains(err.Error(), "unknown policy pack") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AC2: Empty use case is rejected.
func TestSelectRejectsEmptyUseCase(t *testing.T) {
	r := NewRouter()
	_, err := r.Select(RouteRequest{PolicyPack: "balanced"})
	if err == nil {
		t.Fatal("expected error for empty use case")
	}
}

// AC1: Balanced pack respects budget parameter.
func TestBalancedPackBudgetSensitivity(t *testing.T) {
	r := NewRouter()
	lowBudget, _ := r.Select(RouteRequest{UseCase: "test", PolicyPack: "balanced", Budget: "low"})
	normalBudget, _ := r.Select(RouteRequest{UseCase: "test", PolicyPack: "balanced", Budget: "normal"})
	if lowBudget == normalBudget {
		t.Fatalf("balanced pack should differentiate by budget, both returned %q", lowBudget)
	}
}
