package skillpackage_test

import (
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/skillpackage"
)

func TestValidate_EmptySkillDir(t *testing.T) {
	cmd := skillpackage.PackageSkillCommand{}.Normalized()
	if err := cmd.Validate(); err != skillpackage.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_WhitespaceSkillDir(t *testing.T) {
	cmd := skillpackage.PackageSkillCommand{SkillDir: "   "}.Normalized()
	if err := cmd.Validate(); err != skillpackage.ErrSkillDirRequired {
		t.Errorf("expected ErrSkillDirRequired, got %v", err)
	}
}

func TestValidate_ValidSkillDir(t *testing.T) {
	cmd := skillpackage.PackageSkillCommand{SkillDir: "/path/to/skill"}.Normalized()
	if err := cmd.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestArtifactName(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		version  string
		expected string
	}{
		{
			name:     "standard artifact",
			id:       "my-skill",
			version:  "1.0.0",
			expected: "my-skill-1.0.0.zip",
		},
		{
			name:     "with prerelease",
			id:       "analytics",
			version:  "0.2.0-beta",
			expected: "analytics-0.2.0-beta.zip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := skillpackage.ArtifactName(tt.id, tt.version)
			if got != tt.expected {
				t.Errorf("ArtifactName(%q, %q) = %q, want %q", tt.id, tt.version, got, tt.expected)
			}
		})
	}
}
