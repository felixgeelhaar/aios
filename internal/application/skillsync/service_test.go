package skillsync

import (
	"context"
	"errors"
	"fmt"
	"testing"

	domain "github.com/felixgeelhaar/aios/internal/domain/skillsync"
)

type fakeSkillResolver struct {
	id  string
	err error
}

func (f fakeSkillResolver) ResolveSkillID(string) (string, error) {
	return f.id, f.err
}

type fakeInstaller struct {
	skillID string
	called  bool
	err     error
}

func (f *fakeInstaller) InstallSkillAcrossClients(_ context.Context, skillID string) error {
	f.called = true
	f.skillID = skillID
	return f.err
}

func TestServiceSyncSkill(t *testing.T) {
	installer := &fakeInstaller{}
	svc := NewService(fakeSkillResolver{id: "roadmap-reader"}, installer)
	id, err := svc.SyncSkill(context.Background(), domain.SyncSkillCommand{SkillDir: "testdata/skill"})
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if id != "roadmap-reader" {
		t.Fatalf("unexpected skill id: %q", id)
	}
	if installer.skillID != "roadmap-reader" {
		t.Fatalf("installer called with unexpected id: %q", installer.skillID)
	}
}

func TestServiceSyncSkillRequiresSkillDir(t *testing.T) {
	installer := &fakeInstaller{}
	svc := NewService(fakeSkillResolver{id: "roadmap-reader"}, installer)
	_, err := svc.SyncSkill(context.Background(), domain.SyncSkillCommand{})
	if !errors.Is(err, domain.ErrSkillDirRequired) {
		t.Fatalf("expected skill-dir required error, got %v", err)
	}
}

// AC1: Must validate skill.yaml before writing to any client directory.
func TestSyncValidatesBeforeInstall(t *testing.T) {
	validationErr := fmt.Errorf("invalid skill.yaml: missing required field 'id'")
	installer := &fakeInstaller{}
	svc := NewService(fakeSkillResolver{err: validationErr}, installer)
	_, err := svc.SyncSkill(context.Background(), domain.SyncSkillCommand{SkillDir: "/tmp/bad-skill"})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if err.Error() != validationErr.Error() {
		t.Fatalf("unexpected error: %v", err)
	}
	if installer.called {
		t.Fatal("installer must not be called when validation fails")
	}
}

// AC2 + AC3: Must validate JSON schemas and fail fast on schema errors.
func TestSyncFailsFastOnSchemaError(t *testing.T) {
	schemaErr := fmt.Errorf("schema validation failed: inputs.schema must have type=object")
	installer := &fakeInstaller{}
	svc := NewService(fakeSkillResolver{err: schemaErr}, installer)
	_, err := svc.SyncSkill(context.Background(), domain.SyncSkillCommand{SkillDir: "/tmp/bad-schema"})
	if err == nil {
		t.Fatal("expected schema error, got nil")
	}
	if err.Error() != schemaErr.Error() {
		t.Fatalf("unexpected error: %v", err)
	}
	if installer.called {
		t.Fatal("installer must not be called when schema validation fails")
	}
}

// AC8: Must install to all configured client adapters in a single sync invocation.
func TestSyncInstallsToAllClients(t *testing.T) {
	installer := &fakeInstaller{}
	svc := NewService(fakeSkillResolver{id: "multi-client-skill"}, installer)
	id, err := svc.SyncSkill(context.Background(), domain.SyncSkillCommand{SkillDir: "/tmp/skill"})
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if id != "multi-client-skill" {
		t.Fatalf("unexpected skill id: %q", id)
	}
	if !installer.called {
		t.Fatal("installer must be called for cross-client install")
	}
	if installer.skillID != "multi-client-skill" {
		t.Fatalf("installer called with wrong id: %q", installer.skillID)
	}
}
