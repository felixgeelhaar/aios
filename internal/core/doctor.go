package core

import (
	"os"
	"path/filepath"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

type DoctorCheck struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

type DoctorReport struct {
	Overall bool          `json:"overall"`
	Checks  []DoctorCheck `json:"checks"`
}

func RunDoctor(cfg Config) DoctorReport {
	projectCheck := dirCheck("project_dir", cfg.ProjectDir)
	skillsCheck := DoctorCheck{Name: "skills_dir", OK: false, Detail: "project directory not set"}
	if cfg.ProjectDir != "" {
		skillsCheck = dirCheck("skills_dir", filepath.Join(cfg.ProjectDir, agentregistry.CanonicalSkillsDir))
	}
	checks := []DoctorCheck{
		dirCheck("workspace_dir", cfg.WorkspaceDir),
		projectCheck,
		skillsCheck,
	}
	overall := true
	for _, c := range checks {
		if !c.OK {
			overall = false
			break
		}
	}
	return DoctorReport{Overall: overall, Checks: checks}
}

func dirCheck(name, path string) DoctorCheck {
	if path == "" {
		return DoctorCheck{Name: name, OK: false, Detail: "path is empty"}
	}
	if err := os.MkdirAll(path, 0o750); err != nil {
		return DoctorCheck{Name: name, OK: false, Detail: err.Error()}
	}
	return DoctorCheck{Name: name, OK: true, Detail: path}
}
