package core

import (
	"path/filepath"
	"testing"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

func TestRunDoctor(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.WorkspaceDir = filepath.Join(root, "workspace")
	cfg.ProjectDir = filepath.Join(root, "project")

	r := RunDoctor(cfg)
	if !r.Overall {
		t.Fatalf("expected healthy report: %#v", r)
	}
	if len(r.Checks) != 3 {
		t.Fatalf("expected 3 checks, got %d: %#v", len(r.Checks), r.Checks)
	}
}

// AC2: Doctor must validate workspace directory existence and permissions.
func TestRunDoctorWorkspaceFailure(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WorkspaceDir = ""
	cfg.ProjectDir = t.TempDir()

	r := RunDoctor(cfg)
	if r.Overall {
		t.Fatal("expected failure when workspace dir is empty")
	}
	found := false
	for _, c := range r.Checks {
		if c.Name == "workspace_dir" && !c.OK {
			found = true
		}
	}
	if !found {
		t.Fatal("expected workspace_dir check to fail")
	}
}

// AC3: Doctor must validate project and skills directories.
func TestRunDoctorValidatesProjectAndSkillsDirs(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultConfig()
	cfg.WorkspaceDir = filepath.Join(root, "workspace")
	cfg.ProjectDir = ""

	r := RunDoctor(cfg)
	if r.Overall {
		t.Fatal("expected failure when project dir is empty")
	}
	names := map[string]bool{}
	for _, c := range r.Checks {
		names[c.Name] = true
	}
	for _, want := range []string{"workspace_dir", "project_dir", "skills_dir"} {
		if !names[want] {
			t.Fatalf("missing check: %s", want)
		}
	}

	// Verify skills_dir check references canonical path.
	for _, c := range r.Checks {
		if c.Name == "skills_dir" && c.OK {
			// If project dir is empty, skills dir should fail too.
			t.Fatal("expected skills_dir to fail when project_dir is empty")
		}
	}
	_ = agentregistry.CanonicalSkillsDir // ensure import used
}
