package workspaceorchestration

import "testing"

func TestRecommendAction(t *testing.T) {
	tests := []struct {
		name       string
		status     LinkStatus
		wantKind   ActionKind
		wantReason string
	}{
		{"ok returns skip", LinkStatusOK, ActionSkip, "already healthy"},
		{"missing returns create", LinkStatusMissing, ActionCreate, "link missing"},
		{"broken returns repair", LinkStatusBroken, ActionRepair, "link target mismatch"},
		{"conflict returns skip", LinkStatusConflict, ActionSkip, "non-symlink conflict at link path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lr := LinkReport{
				ProjectID:   "p1",
				ProjectPath: "/repo",
				LinkPath:    "/links/p1",
				Status:      tt.status,
			}
			action := lr.RecommendAction()
			if action.Kind != tt.wantKind {
				t.Fatalf("kind = %s, want %s", action.Kind, tt.wantKind)
			}
			if action.Reason != tt.wantReason {
				t.Fatalf("reason = %q, want %q", action.Reason, tt.wantReason)
			}
			if action.ProjectID != "p1" {
				t.Fatalf("project_id = %s, want p1", action.ProjectID)
			}
		})
	}
}

func TestIsApplicable(t *testing.T) {
	tests := []struct {
		name string
		kind ActionKind
		want bool
	}{
		{"create is applicable", ActionCreate, true},
		{"repair is applicable", ActionRepair, true},
		{"skip is not applicable", ActionSkip, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := PlanAction{Kind: tt.kind}
			if got := pa.IsApplicable(); got != tt.want {
				t.Fatalf("IsApplicable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeHealthy(t *testing.T) {
	tests := []struct {
		name  string
		links []LinkReport
		want  bool
	}{
		{
			"all ok",
			[]LinkReport{
				{Status: LinkStatusOK},
				{Status: LinkStatusOK},
			},
			true,
		},
		{
			"one broken",
			[]LinkReport{
				{Status: LinkStatusOK},
				{Status: LinkStatusBroken},
			},
			false,
		},
		{"empty", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComputeHealthy(tt.links); got != tt.want {
				t.Fatalf("ComputeHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}
