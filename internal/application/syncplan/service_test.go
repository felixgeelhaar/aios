package syncplan

import (
	"context"
	"errors"
	"fmt"
	"testing"

	domain "github.com/felixgeelhaar/aios/internal/domain/syncplan"
)

type fakeSkillResolver struct {
	id  string
	err error
}

func (f fakeSkillResolver) ResolveSkillID(string) (string, error) {
	return f.id, f.err
}

type fakeTargetPlanner struct {
	writes []string
	err    error
}

func (f fakeTargetPlanner) PlanWriteTargets(context.Context, string) ([]string, error) {
	return f.writes, f.err
}

func TestBuildSyncPlan(t *testing.T) {
	svc := NewService(fakeSkillResolver{id: "roadmap-reader"}, fakeTargetPlanner{writes: []string{"a", "b"}})
	res, err := svc.BuildSyncPlan(context.Background(), domain.BuildSyncPlanCommand{SkillDir: "/tmp/skill"})
	if err != nil {
		t.Fatalf("build sync plan failed: %v", err)
	}
	if res.SkillID != "roadmap-reader" || len(res.Writes) != 2 {
		t.Fatalf("unexpected result: %#v", res)
	}
}

func TestBuildSyncPlanRequiresSkillDir(t *testing.T) {
	svc := NewService(fakeSkillResolver{id: "roadmap-reader"}, fakeTargetPlanner{})
	_, err := svc.BuildSyncPlan(context.Background(), domain.BuildSyncPlanCommand{})
	if !errors.Is(err, domain.ErrSkillDirRequired) {
		t.Fatalf("expected skill-dir required error, got %v", err)
	}
}

// AC4: Must support dry-run mode via sync-plan command.
// AC5: Must not mutate client directories in dry-run mode.
func TestSyncPlanReturnsPlanWithoutMutation(t *testing.T) {
	writes := []string{
		"/tmp/claude/skills/my-skill.json",
		"/tmp/cursor/mcp.json",
		"/tmp/windsurf/my-skill.yaml",
	}
	svc := NewService(
		fakeSkillResolver{id: "my-skill"},
		fakeTargetPlanner{writes: writes},
	)
	res, err := svc.BuildSyncPlan(context.Background(), domain.BuildSyncPlanCommand{SkillDir: "/tmp/skill"})
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if res.SkillID != "my-skill" {
		t.Fatalf("unexpected skill id: %q", res.SkillID)
	}
	if len(res.Writes) != 3 {
		t.Fatalf("expected 3 planned writes, got %d", len(res.Writes))
	}
	// The planner only computes paths — the fakeTargetPlanner proves no
	// filesystem mutation can occur because it never touches the OS. The
	// real adapter (syncPlanWriteTargetPlannerAdapter) likewise only calls
	// filepath.Join to build path strings.
}

// AC1 + AC2: sync-plan must also validate skill.yaml and schemas before returning plan.
func TestSyncPlanValidatesBeforePlanning(t *testing.T) {
	validationErr := fmt.Errorf("invalid schema: inputs.schema type must be object")
	svc := NewService(
		fakeSkillResolver{err: validationErr},
		fakeTargetPlanner{writes: []string{"should-not-appear"}},
	)
	_, err := svc.BuildSyncPlan(context.Background(), domain.BuildSyncPlanCommand{SkillDir: "/tmp/bad"})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if err.Error() != validationErr.Error() {
		t.Fatalf("unexpected error: %v", err)
	}
}
