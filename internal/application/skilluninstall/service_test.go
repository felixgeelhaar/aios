package skilluninstall

import (
	"context"
	"errors"
	"testing"

	domain "github.com/felixgeelhaar/aios/internal/domain/skilluninstall"
)

type fakeSkillIDResolver struct {
	skillID string
	err     error
}

func (f fakeSkillIDResolver) ResolveSkillID(string) (string, error) {
	return f.skillID, f.err
}

type fakeClientUninstaller struct {
	skillID string
	err     error
}

func (f *fakeClientUninstaller) UninstallAcrossClients(_ context.Context, skillID string) error {
	f.skillID = skillID
	return f.err
}

func TestServiceUninstallSkill(t *testing.T) {
	uninstaller := &fakeClientUninstaller{}
	svc := NewService(fakeSkillIDResolver{skillID: "roadmap-reader"}, uninstaller)
	id, err := svc.UninstallSkill(context.Background(), domain.UninstallSkillCommand{SkillDir: "/tmp/skill"})
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}
	if id != "roadmap-reader" {
		t.Fatalf("unexpected skill id: %q", id)
	}
	if uninstaller.skillID != "roadmap-reader" {
		t.Fatalf("unexpected uninstaller skill id: %q", uninstaller.skillID)
	}
}

func TestServiceUninstallSkillRequiresSkillDir(t *testing.T) {
	uninstaller := &fakeClientUninstaller{}
	svc := NewService(fakeSkillIDResolver{skillID: "roadmap-reader"}, uninstaller)
	_, err := svc.UninstallSkill(context.Background(), domain.UninstallSkillCommand{})
	if !errors.Is(err, domain.ErrSkillDirRequired) {
		t.Fatalf("expected skill-dir required error, got %v", err)
	}
}
