package rollout

import (
	"strings"
	"testing"
)

func TestBuildPlan(t *testing.T) {
	_, err := BuildPlan(Bundle{Name: "support", Skills: []string{"roadmap-reader"}}, []string{"team-a"})
	if err != nil {
		t.Fatalf("build plan failed: %v", err)
	}
}

// AC1: Bundles support multiple skills for department targeting.
func TestBuildPlan_MultipleSkillsBundle(t *testing.T) {
	bundle := Bundle{
		Name:   "engineering-tools",
		Skills: []string{"code-reviewer", "test-runner", "deploy-helper"},
	}
	plan, err := BuildPlan(bundle, []string{"team-backend", "team-frontend"})
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if plan.BundleName != "engineering-tools" {
		t.Fatalf("expected bundle name 'engineering-tools', got %q", plan.BundleName)
	}
	if len(plan.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(plan.Targets))
	}
}

// AC1: Bundle requires name.
func TestBuildPlan_RejectsEmptyBundleName(t *testing.T) {
	_, err := BuildPlan(Bundle{Skills: []string{"x"}}, []string{"team-a"})
	if err == nil {
		t.Fatal("expected error for empty bundle name")
	}
	if !strings.Contains(err.Error(), "bundle name is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AC1: Bundle requires at least one skill.
func TestBuildPlan_RejectsEmptySkills(t *testing.T) {
	_, err := BuildPlan(Bundle{Name: "empty"}, []string{"team-a"})
	if err == nil {
		t.Fatal("expected error for empty skills")
	}
	if !strings.Contains(err.Error(), "bundle must include skills") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AC2: Rollout plan targets can represent staged rollout groups.
func TestBuildPlan_StagedTargets(t *testing.T) {
	bundle := Bundle{Name: "gradual-rollout", Skills: []string{"analytics"}}
	// Targets represent percentage-based groups.
	targets := []string{"canary-10pct", "early-adopters-25pct", "general-50pct", "all-100pct"}
	plan, err := BuildPlan(bundle, targets)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if len(plan.Targets) != 4 {
		t.Fatalf("expected 4 staged targets, got %d", len(plan.Targets))
	}
}

// AC2: Rollout requires at least one target.
func TestBuildPlan_RejectsEmptyTargets(t *testing.T) {
	_, err := BuildPlan(Bundle{Name: "x", Skills: []string{"y"}}, nil)
	if err == nil {
		t.Fatal("expected error for empty targets")
	}
	if !strings.Contains(err.Error(), "targets are required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AC3: Plan preserves enough data for rollback (bundle name + targets are recorded).
func TestBuildPlan_PreservesRollbackData(t *testing.T) {
	bundle := Bundle{Name: "support-bundle", Skills: []string{"help-desk", "faq-bot"}}
	plan, err := BuildPlan(bundle, []string{"team-support"})
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	// Rollback requires knowing which bundle was deployed to which targets.
	if plan.BundleName == "" {
		t.Fatal("plan must record bundle name for rollback")
	}
	if len(plan.Targets) == 0 {
		t.Fatal("plan must record targets for rollback")
	}
}
