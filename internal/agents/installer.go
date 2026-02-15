package agents

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

var sanitizeRegexp = regexp.MustCompile(`[^a-zA-Z0-9-]`)

// SkillInstaller handles skill installation into the canonical .agents/skills/
// directory and creates symlinks for non-universal agents.
type SkillInstaller struct {
	agents []agentregistry.AgentDefinition
}

// NewSkillInstaller creates a SkillInstaller with the given agent definitions.
func NewSkillInstaller(agents []agentregistry.AgentDefinition) *SkillInstaller {
	return &SkillInstaller{agents: agents}
}

// InstallOptions configures a skill installation.
type InstallOptions struct {
	// ProjectDir is the project root where skills will be installed.
	ProjectDir string

	// TargetAgents is the explicit list of agents to install for.
	// If nil, defaults to all agents.
	TargetAgents []agentregistry.AgentDefinition

	// SkillContent is the full SKILL.md content to write. When non-empty,
	// this replaces the default stub marker. Typically composed from
	// skill.yaml metadata and prompt.md body by the caller.
	SkillContent string
}

// InstallResult represents the outcome of a skill installation.
type InstallResult struct {
	// SkillID is the identifier of the installed skill.
	SkillID string

	// CanonicalPath is the path where the skill was installed.
	CanonicalPath string

	// Agents lists the display names of agents that received the skill.
	Agents []string
}

// InstallSkill installs a skill to the canonical location and creates symlinks
// for non-universal agents. The skillID is used as the directory name (sanitized).
func (si *SkillInstaller) InstallSkill(skillID string, opts InstallOptions) (*InstallResult, error) {
	if skillID == "" {
		return nil, fmt.Errorf("skill id is required")
	}
	if opts.ProjectDir == "" {
		return nil, fmt.Errorf("project directory is required")
	}

	sanitized := SanitizeName(skillID)
	canonicalDir := filepath.Join(opts.ProjectDir, agentregistry.CanonicalSkillsDir, sanitized)

	// Create canonical directory.
	if err := os.MkdirAll(canonicalDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating canonical dir: %w", err)
	}

	// Write skill marker file. If SkillContent is provided, always write it
	// (overwriting any previous stub). Otherwise, write a default stub only
	// when no SKILL.md exists yet.
	markerPath := filepath.Join(canonicalDir, "SKILL.md")
	if opts.SkillContent != "" {
		if err := os.WriteFile(markerPath, []byte(opts.SkillContent), 0o644); err != nil {
			return nil, fmt.Errorf("writing SKILL.md: %w", err)
		}
	} else if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		content := fmt.Sprintf("---\nname: %s\ndescription: \"\"\n---\n", skillID)
		if err := os.WriteFile(markerPath, []byte(content), 0o644); err != nil {
			return nil, fmt.Errorf("writing SKILL.md: %w", err)
		}
	}

	// Determine target agents.
	targets := si.agents
	if len(opts.TargetAgents) > 0 {
		targets = opts.TargetAgents
	}

	// Create symlinks for non-universal agents.
	var installedAgents []string
	for _, agent := range targets {
		if agent.Universal {
			installedAgents = append(installedAgents, agent.DisplayName)
			continue
		}

		agentSkillDir := filepath.Join(opts.ProjectDir, agent.SkillsDir)
		if err := os.MkdirAll(agentSkillDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating agent dir for %s: %w", agent.DisplayName, err)
		}

		linkPath := filepath.Join(agentSkillDir, sanitized)

		// Remove existing link/dir if present.
		_ = os.RemoveAll(linkPath)

		// Create relative symlink from agent dir to canonical location.
		rel, err := filepath.Rel(agentSkillDir, canonicalDir)
		if err != nil {
			return nil, fmt.Errorf("computing relative path for %s: %w", agent.DisplayName, err)
		}

		if err := os.Symlink(rel, linkPath); err != nil {
			// Fall back to copy if symlink fails (e.g., Windows without privileges).
			if copyErr := copyDirectory(canonicalDir, linkPath); copyErr != nil {
				return nil, fmt.Errorf("symlink and copy both failed for %s: symlink: %w, copy: %v",
					agent.DisplayName, err, copyErr)
			}
		}

		installedAgents = append(installedAgents, agent.DisplayName)
	}

	return &InstallResult{
		SkillID:       skillID,
		CanonicalPath: canonicalDir,
		Agents:        installedAgents,
	}, nil
}

// UninstallSkill removes a skill from the canonical location and removes
// symlinks/copies from all agent skill directories.
func (si *SkillInstaller) UninstallSkill(skillID string, projectDir string) error {
	if skillID == "" {
		return fmt.Errorf("skill id is required")
	}
	if projectDir == "" {
		return fmt.Errorf("project directory is required")
	}

	sanitized := SanitizeName(skillID)

	// Remove symlinks from non-universal agents first.
	for _, agent := range si.agents {
		if agent.Universal {
			continue
		}
		linkPath := filepath.Join(projectDir, agent.SkillsDir, sanitized)
		_ = os.RemoveAll(linkPath)
	}

	// Remove canonical directory.
	canonicalDir := filepath.Join(projectDir, agentregistry.CanonicalSkillsDir, sanitized)
	if err := os.RemoveAll(canonicalDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing canonical dir: %w", err)
	}

	return nil
}

// PlanWriteTargets returns the list of paths that would be written to
// when installing a skill. Used for dry-run planning.
func (si *SkillInstaller) PlanWriteTargets(skillID string, projectDir string) []string {
	sanitized := SanitizeName(skillID)
	targets := []string{
		filepath.Join(projectDir, agentregistry.CanonicalSkillsDir, sanitized),
	}
	for _, agent := range si.agents {
		if agent.Universal {
			continue
		}
		targets = append(targets,
			filepath.Join(projectDir, agent.SkillsDir, sanitized))
	}
	return targets
}

// CollectInstalledSkills scans all agent skill directories in a project
// and returns a deduplicated sorted list of installed skill IDs.
func (si *SkillInstaller) CollectInstalledSkills(projectDir string) ([]string, error) {
	seen := make(map[string]struct{})

	// Scan canonical directory first.
	canonicalDir := filepath.Join(projectDir, agentregistry.CanonicalSkillsDir)
	if entries, err := os.ReadDir(canonicalDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				seen[e.Name()] = struct{}{}
			}
		}
	}

	// Also scan non-universal agent directories for skills not yet in canonical.
	for _, agent := range si.agents {
		if agent.Universal {
			continue
		}
		agentDir := filepath.Join(projectDir, agent.SkillsDir)
		entries, err := os.ReadDir(agentDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if name != "" {
				seen[name] = struct{}{}
			}
		}
	}

	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sortStrings(out)
	return out, nil
}

// SanitizeName normalizes a skill name for use as a directory name.
func SanitizeName(name string) string {
	name = strings.ToLower(name)
	name = sanitizeRegexp.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-.")
	if len(name) > 255 {
		name = name[:255]
	}
	if name == "" {
		name = "unnamed-skill"
	}
	return name
}

// copyDirectory copies the contents of src to dst.
func copyDirectory(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if strings.HasPrefix(filepath.Base(path), ".") && path != src {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dstPath := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// sortStrings sorts a string slice in place.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}
