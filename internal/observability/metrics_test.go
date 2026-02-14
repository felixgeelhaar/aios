package observability

import "testing"

func TestMetricsCounter(t *testing.T) {
	m := NewMetrics()
	m.Inc("runs")
	m.Inc("runs")
	if m.Value("runs") != 2 {
		t.Fatalf("unexpected value: %d", m.Value("runs"))
	}
}

// AC1: Track skill usage counts per skill.
func TestMetrics_SkillUsageCounts(t *testing.T) {
	m := NewMetrics()
	m.Inc("skill:roadmap-reader:executions")
	m.Inc("skill:roadmap-reader:executions")
	m.Inc("skill:code-reviewer:executions")

	if m.Value("skill:roadmap-reader:executions") != 2 {
		t.Fatalf("expected 2 executions for roadmap-reader, got %d", m.Value("skill:roadmap-reader:executions"))
	}
	if m.Value("skill:code-reviewer:executions") != 1 {
		t.Fatalf("expected 1 execution for code-reviewer, got %d", m.Value("skill:code-reviewer:executions"))
	}
}

// AC1: Track usage counts per team.
func TestMetrics_TeamUsageCounts(t *testing.T) {
	m := NewMetrics()
	m.Inc("team:backend:executions")
	m.Inc("team:backend:executions")
	m.Inc("team:frontend:executions")

	if m.Value("team:backend:executions") != 2 {
		t.Fatalf("expected 2 for backend team, got %d", m.Value("team:backend:executions"))
	}
	if m.Value("team:frontend:executions") != 1 {
		t.Fatalf("expected 1 for frontend team, got %d", m.Value("team:frontend:executions"))
	}
}

// AC2: Track success and failure rates.
func TestMetrics_SuccessFailureRates(t *testing.T) {
	m := NewMetrics()
	m.Inc("executions:success")
	m.Inc("executions:success")
	m.Inc("executions:success")
	m.Inc("executions:failure")

	total := m.Value("executions:success") + m.Value("executions:failure")
	if total != 4 {
		t.Fatalf("expected 4 total executions, got %d", total)
	}
	if m.Value("executions:failure") != 1 {
		t.Fatalf("expected 1 failure, got %d", m.Value("executions:failure"))
	}
}

// AC7: Track drift incidents and auto-resolve rates.
func TestMetrics_DriftTracking(t *testing.T) {
	m := NewMetrics()
	m.Inc("drift:incidents")
	m.Inc("drift:incidents")
	m.Inc("drift:incidents")
	m.Inc("drift:auto_resolved")
	m.Inc("drift:auto_resolved")

	incidents := m.Value("drift:incidents")
	resolved := m.Value("drift:auto_resolved")
	if incidents != 3 {
		t.Fatalf("expected 3 drift incidents, got %d", incidents)
	}
	if resolved != 2 {
		t.Fatalf("expected 2 auto-resolved, got %d", resolved)
	}
}

// AC2: Untracked metric returns zero.
func TestMetrics_UntrackedReturnsZero(t *testing.T) {
	m := NewMetrics()
	if m.Value("nonexistent") != 0 {
		t.Fatalf("expected 0 for untracked metric, got %d", m.Value("nonexistent"))
	}
}
