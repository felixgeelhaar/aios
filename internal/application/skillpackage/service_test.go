package skillpackage

import (
	"context"
	"errors"
	"testing"

	domain "github.com/felixgeelhaar/aios/internal/domain/skillpackage"
)

type fakeMetadataResolver struct {
	id      string
	version string
	err     error
}

func (f fakeMetadataResolver) ResolveIDAndVersion(string) (string, string, error) {
	return f.id, f.version, f.err
}

type fakePackager struct {
	skillDir string
	outPath  string
	err      error
}

func (f *fakePackager) Package(skillDir string, outputPath string) error {
	f.skillDir = skillDir
	f.outPath = outputPath
	return f.err
}

func TestServicePackageSkill(t *testing.T) {
	p := &fakePackager{}
	svc := NewService(fakeMetadataResolver{id: "roadmap-reader", version: "0.1.0"}, p)
	res, err := svc.PackageSkill(context.Background(), domain.PackageSkillCommand{SkillDir: "/tmp/skill"})
	if err != nil {
		t.Fatalf("package failed: %v", err)
	}
	if res.ArtifactPath != "/tmp/roadmap-reader-0.1.0.zip" {
		t.Fatalf("unexpected artifact path: %q", res.ArtifactPath)
	}
	if p.skillDir != "/tmp/skill" {
		t.Fatalf("unexpected skill dir: %q", p.skillDir)
	}
}

func TestServicePackageSkillRequiresSkillDir(t *testing.T) {
	p := &fakePackager{}
	svc := NewService(fakeMetadataResolver{id: "roadmap-reader", version: "0.1.0"}, p)
	_, err := svc.PackageSkill(context.Background(), domain.PackageSkillCommand{})
	if !errors.Is(err, domain.ErrSkillDirRequired) {
		t.Fatalf("expected skill-dir required error, got %v", err)
	}
}
