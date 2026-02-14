package observability

import (
	"encoding/json"
	"testing"
)

// --- test-local gate sets (not exported from the library) ---

func localKernelGates() []Gate {
	return []Gate{
		{Name: "Onboarding time < 15 min", Metric: "onboarding_time_min", Threshold: 15, Operator: "lt"},
		{Name: "Zero JSON editing", Metric: "manual_json_edits", Threshold: 0, Operator: "eq"},
		{Name: "Connector binding rate > 95%", Metric: "connector_binding_rate", Threshold: 95, Operator: "gt"},
		{Name: "Skill success rate > 98%", Metric: "skill_success_rate", Threshold: 98, Operator: "gt"},
		{Name: "Drift auto-resolve >= 90%", Metric: "drift_auto_resolve_rate", Threshold: 90, Operator: "gte"},
		{Name: "Skills built via Quick Builder >= 5", Metric: "skills_built", Threshold: 5, Operator: "gte"},
		{Name: "Non-eng users onboarded >= 2", Metric: "non_eng_users_onboarded", Threshold: 2, Operator: "gte"},
		{Name: "Weekly power users 10-20", Metric: "weekly_power_users", Threshold: 10, Operator: "gte"},
	}
}

func orgControlPlaneGates() []Gate {
	return []Gate{
		{Name: "Teams using shared skills >= 3", Metric: "teams_using_shared_skills", Threshold: 3, Operator: "gte"},
		{Name: "Org-wide rollout achieved >= 1", Metric: "org_wide_rollouts", Threshold: 1, Operator: "gte"},
		{Name: "Admin approval workflow functional", Metric: "approval_workflow_functional", Threshold: 1, Operator: "gte"},
		{Name: "Version enforcement stable", Metric: "version_enforcement_stable", Threshold: 1, Operator: "gte"},
		{Name: "Org-wide rollout time < 60 min", Metric: "rollout_time_min", Threshold: 60, Operator: "lt"},
		{Name: "Skill version adoption rate > 80%", Metric: "version_adoption_rate", Threshold: 80, Operator: "gt"},
		{Name: "Config support ticket reduction", Metric: "config_ticket_reduction_pct", Threshold: 0, Operator: "gt"},
		{Name: "First paid customers (Pro tier)", Metric: "paid_customers", Threshold: 1, Operator: "gte"},
	}
}

func platformGates() []Gate {
	return []Gate{
		{Name: "Policy enforcement coverage", Metric: "policy_enforcement_coverage", Threshold: 100, Operator: "eq"},
		{Name: "Audit trail completeness", Metric: "audit_trail_completeness", Threshold: 100, Operator: "eq"},
		{Name: "Enterprise adoption metrics tracked", Metric: "enterprise_adoption_tracked", Threshold: 1, Operator: "gte"},
		{Name: "External dev publishes verified skill", Metric: "external_verified_publishes", Threshold: 1, Operator: "gte"},
		{Name: "Org installs marketplace skill safely", Metric: "marketplace_safe_installs", Threshold: 1, Operator: "gte"},
		{Name: "Model routing cost tracking operational", Metric: "model_cost_tracking_operational", Threshold: 1, Operator: "gte"},
		{Name: "Analytics dashboard used by platform teams", Metric: "analytics_dashboard_usage", Threshold: 1, Operator: "gte"},
	}
}

// --- generic evaluator tests ---

func TestEvaluateGates_UnknownOperatorFails(t *testing.T) {
	gates := []Gate{
		{Name: "bad op", Metric: "x", Threshold: 10, Operator: "invalid"},
	}
	results := EvaluateGates(gates, map[string]float64{"x": 10})
	if results[0].Passed {
		t.Fatal("expected unknown operator to fail")
	}
}

func TestEvaluateGates_EmptyGates(t *testing.T) {
	results := EvaluateGates(nil, map[string]float64{"x": 1})
	if len(results) != 0 {
		t.Fatalf("expected empty results for nil gates, got %d", len(results))
	}
}

func TestEvaluateGates_MissingMetricDefaultsToZero(t *testing.T) {
	gates := []Gate{
		{Name: "test", Metric: "missing_key", Threshold: 0, Operator: "eq"},
	}
	results := EvaluateGates(gates, map[string]float64{})
	if !results[0].Passed {
		t.Fatal("expected missing metric (zero) to match eq 0")
	}
	if results[0].Value != 0 {
		t.Fatalf("expected value 0 for missing metric, got %f", results[0].Value)
	}
}

func TestAllPassed_SingleFailure(t *testing.T) {
	gates := localKernelGates()
	results := EvaluateGates(gates, map[string]float64{})
	if AllPassed(results) {
		t.Fatal("expected AllPassed to return false with zero metrics")
	}
}

func TestGateResults_JSONExportable(t *testing.T) {
	gates := localKernelGates()
	values := map[string]float64{
		"onboarding_time_min":     10,
		"manual_json_edits":       0,
		"connector_binding_rate":  97,
		"skill_success_rate":      99,
		"drift_auto_resolve_rate": 92,
		"skills_built":            7,
		"non_eng_users_onboarded": 3,
		"weekly_power_users":      15,
	}
	results := EvaluateGates(gates, values)
	data, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("gate results not JSON-serializable: %v", err)
	}
	var parsed []GateResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("cannot unmarshal gate results: %v", err)
	}
	if len(parsed) != len(results) {
		t.Fatalf("expected %d results, got %d", len(results), len(parsed))
	}
}

// --- local kernel exit criteria ---

func TestGate_OnboardingTimeUnder15Min(t *testing.T) {
	gates := []Gate{
		{Name: "Onboarding time", Metric: "onboarding_time_min", Threshold: 15, Operator: "lt"},
	}
	results := EvaluateGates(gates, map[string]float64{"onboarding_time_min": 10})
	if !results[0].Passed {
		t.Fatal("expected onboarding time gate to pass with 10 min")
	}
	results = EvaluateGates(gates, map[string]float64{"onboarding_time_min": 15})
	if results[0].Passed {
		t.Fatal("expected onboarding time gate to fail at exactly 15 min")
	}
}

func TestGate_ZeroJSONEditing(t *testing.T) {
	gates := []Gate{
		{Name: "Zero JSON editing", Metric: "manual_json_edits", Threshold: 0, Operator: "eq"},
	}
	results := EvaluateGates(gates, map[string]float64{"manual_json_edits": 0})
	if !results[0].Passed {
		t.Fatal("expected zero JSON editing gate to pass")
	}
	results = EvaluateGates(gates, map[string]float64{"manual_json_edits": 1})
	if results[0].Passed {
		t.Fatal("expected zero JSON editing gate to fail with 1 edit")
	}
}

func TestGate_ConnectorBindingRate(t *testing.T) {
	gates := []Gate{
		{Name: "Connector binding rate", Metric: "connector_binding_rate", Threshold: 95, Operator: "gt"},
	}
	results := EvaluateGates(gates, map[string]float64{"connector_binding_rate": 97})
	if !results[0].Passed {
		t.Fatal("expected connector binding gate to pass at 97%")
	}
	results = EvaluateGates(gates, map[string]float64{"connector_binding_rate": 95})
	if results[0].Passed {
		t.Fatal("expected connector binding gate to fail at exactly 95%")
	}
}

func TestGate_SkillSuccessRate(t *testing.T) {
	gates := []Gate{
		{Name: "Skill success rate", Metric: "skill_success_rate", Threshold: 98, Operator: "gt"},
	}
	results := EvaluateGates(gates, map[string]float64{"skill_success_rate": 99.5})
	if !results[0].Passed {
		t.Fatal("expected skill success rate gate to pass at 99.5%")
	}
	results = EvaluateGates(gates, map[string]float64{"skill_success_rate": 98})
	if results[0].Passed {
		t.Fatal("expected skill success rate gate to fail at exactly 98%")
	}
}

func TestGate_DriftAutoResolveRate(t *testing.T) {
	gates := []Gate{
		{Name: "Drift auto-resolve", Metric: "drift_auto_resolve_rate", Threshold: 90, Operator: "gte"},
	}
	results := EvaluateGates(gates, map[string]float64{"drift_auto_resolve_rate": 90})
	if !results[0].Passed {
		t.Fatal("expected drift auto-resolve gate to pass at exactly 90%")
	}
	results = EvaluateGates(gates, map[string]float64{"drift_auto_resolve_rate": 89.9})
	if results[0].Passed {
		t.Fatal("expected drift auto-resolve gate to fail at 89.9%")
	}
}

func TestGate_AdoptionMetrics_AllPass(t *testing.T) {
	gates := localKernelGates()
	values := map[string]float64{
		"onboarding_time_min":     10,
		"manual_json_edits":       0,
		"connector_binding_rate":  97,
		"skill_success_rate":      99,
		"drift_auto_resolve_rate": 92,
		"skills_built":            7,
		"non_eng_users_onboarded": 3,
		"weekly_power_users":      15,
	}
	results := EvaluateGates(gates, values)
	if !AllPassed(results) {
		for _, r := range results {
			if !r.Passed {
				t.Errorf("gate %q failed: value=%f threshold=%f op=%s",
					r.Gate.Name, r.Value, r.Gate.Threshold, r.Gate.Operator)
			}
		}
		t.Fatal("expected all local kernel gates to pass")
	}
}

func TestGate_LocalKernelGateCount(t *testing.T) {
	gates := localKernelGates()
	if len(gates) != 8 {
		t.Fatalf("expected 8 local kernel gates, got %d", len(gates))
	}
	assertUniqueGateNames(t, gates)
}

// --- org control plane exit criteria ---

func TestGate_TeamSharedSkills(t *testing.T) {
	gates := []Gate{
		{Name: "Teams sharing", Metric: "teams_using_shared_skills", Threshold: 3, Operator: "gte"},
	}
	results := EvaluateGates(gates, map[string]float64{"teams_using_shared_skills": 3})
	if !results[0].Passed {
		t.Fatal("expected gate to pass at exactly 3 teams")
	}
	results = EvaluateGates(gates, map[string]float64{"teams_using_shared_skills": 2})
	if results[0].Passed {
		t.Fatal("expected gate to fail with 2 teams")
	}
}

func TestGate_OrgRolloutAchieved(t *testing.T) {
	gates := []Gate{
		{Name: "Org rollout", Metric: "org_wide_rollouts", Threshold: 1, Operator: "gte"},
	}
	results := EvaluateGates(gates, map[string]float64{"org_wide_rollouts": 1})
	if !results[0].Passed {
		t.Fatal("expected gate to pass with 1 rollout")
	}
	results = EvaluateGates(gates, map[string]float64{"org_wide_rollouts": 0})
	if results[0].Passed {
		t.Fatal("expected gate to fail with 0 rollouts")
	}
}

func TestGate_RolloutTimeBound(t *testing.T) {
	gates := []Gate{
		{Name: "Rollout time", Metric: "rollout_time_min", Threshold: 60, Operator: "lt"},
	}
	results := EvaluateGates(gates, map[string]float64{"rollout_time_min": 45})
	if !results[0].Passed {
		t.Fatal("expected gate to pass at 45 min")
	}
	results = EvaluateGates(gates, map[string]float64{"rollout_time_min": 60})
	if results[0].Passed {
		t.Fatal("expected gate to fail at exactly 60 min")
	}
}

func TestGate_VersionAdoptionRate(t *testing.T) {
	gates := []Gate{
		{Name: "Adoption rate", Metric: "version_adoption_rate", Threshold: 80, Operator: "gt"},
	}
	results := EvaluateGates(gates, map[string]float64{"version_adoption_rate": 85})
	if !results[0].Passed {
		t.Fatal("expected gate to pass at 85%")
	}
	results = EvaluateGates(gates, map[string]float64{"version_adoption_rate": 80})
	if results[0].Passed {
		t.Fatal("expected gate to fail at exactly 80%")
	}
}

func TestGate_OrgControlPlane_AllPass(t *testing.T) {
	gates := orgControlPlaneGates()
	values := map[string]float64{
		"teams_using_shared_skills":    4,
		"org_wide_rollouts":            2,
		"approval_workflow_functional": 1,
		"version_enforcement_stable":   1,
		"rollout_time_min":             30,
		"version_adoption_rate":        90,
		"config_ticket_reduction_pct":  25,
		"paid_customers":               3,
	}
	results := EvaluateGates(gates, values)
	if !AllPassed(results) {
		for _, r := range results {
			if !r.Passed {
				t.Errorf("gate %q failed: value=%f threshold=%f op=%s",
					r.Gate.Name, r.Value, r.Gate.Threshold, r.Gate.Operator)
			}
		}
		t.Fatal("expected all org control plane gates to pass")
	}
}

func TestGate_OrgControlPlaneGateCount(t *testing.T) {
	gates := orgControlPlaneGates()
	if len(gates) != 8 {
		t.Fatalf("expected 8 org control plane gates, got %d", len(gates))
	}
	assertUniqueGateNames(t, gates)
}

// --- platform exit criteria ---

func TestGate_PolicyEnforcementCoverage(t *testing.T) {
	gates := []Gate{
		{Name: "Policy coverage", Metric: "policy_enforcement_coverage", Threshold: 100, Operator: "eq"},
	}
	results := EvaluateGates(gates, map[string]float64{"policy_enforcement_coverage": 100})
	if !results[0].Passed {
		t.Fatal("expected gate to pass at 100%")
	}
	results = EvaluateGates(gates, map[string]float64{"policy_enforcement_coverage": 99})
	if results[0].Passed {
		t.Fatal("expected gate to fail at 99%")
	}
}

func TestGate_AuditTrailCompleteness(t *testing.T) {
	gates := []Gate{
		{Name: "Audit completeness", Metric: "audit_trail_completeness", Threshold: 100, Operator: "eq"},
	}
	results := EvaluateGates(gates, map[string]float64{"audit_trail_completeness": 100})
	if !results[0].Passed {
		t.Fatal("expected gate to pass at 100%")
	}
	results = EvaluateGates(gates, map[string]float64{"audit_trail_completeness": 95})
	if results[0].Passed {
		t.Fatal("expected gate to fail at 95%")
	}
}

func TestGate_ExternalVerifiedPublish(t *testing.T) {
	gates := []Gate{
		{Name: "External publish", Metric: "external_verified_publishes", Threshold: 1, Operator: "gte"},
	}
	results := EvaluateGates(gates, map[string]float64{"external_verified_publishes": 1})
	if !results[0].Passed {
		t.Fatal("expected gate to pass with 1 publish")
	}
	results = EvaluateGates(gates, map[string]float64{"external_verified_publishes": 0})
	if results[0].Passed {
		t.Fatal("expected gate to fail with 0 publishes")
	}
}

func TestGate_Platform_AllPass(t *testing.T) {
	gates := platformGates()
	values := map[string]float64{
		"policy_enforcement_coverage":     100,
		"audit_trail_completeness":        100,
		"enterprise_adoption_tracked":     1,
		"external_verified_publishes":     2,
		"marketplace_safe_installs":       5,
		"model_cost_tracking_operational": 1,
		"analytics_dashboard_usage":       1,
	}
	results := EvaluateGates(gates, values)
	if !AllPassed(results) {
		for _, r := range results {
			if !r.Passed {
				t.Errorf("gate %q failed: value=%f threshold=%f op=%s",
					r.Gate.Name, r.Value, r.Gate.Threshold, r.Gate.Operator)
			}
		}
		t.Fatal("expected all platform gates to pass")
	}
}

func TestGate_PlatformGateCount(t *testing.T) {
	gates := platformGates()
	if len(gates) != 7 {
		t.Fatalf("expected 7 platform gates, got %d", len(gates))
	}
	assertUniqueGateNames(t, gates)
}

// --- helper ---

func assertUniqueGateNames(t *testing.T, gates []Gate) {
	t.Helper()
	names := make(map[string]bool)
	for _, g := range gates {
		if g.Name == "" {
			t.Fatal("gate has empty name")
		}
		if g.Metric == "" {
			t.Fatal("gate has empty metric key")
		}
		if g.Operator == "" {
			t.Fatal("gate has empty operator")
		}
		names[g.Name] = true
	}
	if len(names) != len(gates) {
		t.Fatalf("expected %d unique gate names, got %d", len(gates), len(names))
	}
}
