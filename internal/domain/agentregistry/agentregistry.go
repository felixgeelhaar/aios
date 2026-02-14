// Package agentregistry defines the domain model for AI coding agent
// definitions. An agent represents an AI coding tool (e.g. OpenCode, Cursor,
// Claude Code) that can consume skills installed in project directories.
//
// Agents are categorized as universal or non-universal:
//   - Universal agents share a canonical .agents/skills/ directory.
//   - Non-universal agents have their own skills directory and receive
//     symlinks pointing back to the canonical location.
package agentregistry

import (
	"fmt"
	"strings"
)

// AgentDefinition is a value object describing an AI coding agent and its
// skill directory conventions. It is loaded from a data-driven configuration
// and must remain free of I/O or infrastructure concerns.
type AgentDefinition struct {
	// Name is the machine-readable identifier (e.g. "opencode", "cursor").
	Name string

	// DisplayName is the human-readable label (e.g. "OpenCode", "Cursor").
	DisplayName string

	// SkillsDir is the project-relative skill directory
	// (e.g. ".cursor/skills" or ".agents/skills" for universal agents).
	SkillsDir string

	// AltSkillsDirs lists additional directories the agent reads from
	// (e.g. ".opencode/skills" as an alternative to ".agents/skills").
	AltSkillsDirs []string

	// GlobalSkillsDir is the global (user-level) skill directory path
	// (may contain ~ or $ENV_VAR placeholders for infrastructure to expand).
	GlobalSkillsDir string

	// DetectPaths lists paths to check for agent presence on the system
	// (may contain ~ or $ENV_VAR placeholders for infrastructure to expand).
	DetectPaths []string

	// Universal indicates whether this agent reads from the shared
	// .agents/skills/ directory. Universal agents do not need symlinks.
	Universal bool
}

// CanonicalSkillsDir is the shared directory used by all universal agents.
const CanonicalSkillsDir = ".agents/skills"

// Validate checks that the agent definition has all required fields.
func (a AgentDefinition) Validate() error {
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("agent name is required")
	}
	if strings.TrimSpace(a.DisplayName) == "" {
		return fmt.Errorf("agent display name is required for %q", a.Name)
	}
	if strings.TrimSpace(a.SkillsDir) == "" {
		return fmt.Errorf("agent skills dir is required for %q", a.Name)
	}
	if a.Universal && a.SkillsDir != CanonicalSkillsDir {
		return fmt.Errorf("universal agent %q must use %s as skills dir, got %q", a.Name, CanonicalSkillsDir, a.SkillsDir)
	}
	return nil
}

// IsUniversal returns whether this agent uses the shared canonical skills directory.
func (a AgentDefinition) IsUniversal() bool {
	return a.Universal
}

// FilterUniversal returns only the universal agents from the given slice.
func FilterUniversal(agents []AgentDefinition) []AgentDefinition {
	var result []AgentDefinition
	for _, a := range agents {
		if a.Universal {
			result = append(result, a)
		}
	}
	return result
}

// FilterNonUniversal returns only the non-universal agents from the given slice.
func FilterNonUniversal(agents []AgentDefinition) []AgentDefinition {
	var result []AgentDefinition
	for _, a := range agents {
		if !a.Universal {
			result = append(result, a)
		}
	}
	return result
}

// ResolveByNames returns agent definitions matching the given names.
// Returns an error if any name does not match a known agent.
func ResolveByNames(agents []AgentDefinition, names []string) ([]AgentDefinition, error) {
	index := make(map[string]AgentDefinition, len(agents))
	for _, a := range agents {
		index[a.Name] = a
	}

	var resolved []AgentDefinition
	for _, name := range names {
		agent, ok := index[name]
		if !ok {
			var valid []string
			for _, a := range agents {
				valid = append(valid, a.Name)
			}
			return nil, fmt.Errorf("unknown agent %q; available: %s", name, strings.Join(valid, ", "))
		}
		resolved = append(resolved, agent)
	}
	return resolved, nil
}

// AllSkillsDirs returns all unique skill directory paths from all agent
// definitions (both primary SkillsDir and AltSkillsDirs), with the canonical
// directory listed first.
func AllSkillsDirs(agents []AgentDefinition) []string {
	seen := make(map[string]bool)
	dirs := []string{CanonicalSkillsDir}
	seen[CanonicalSkillsDir] = true

	for _, agent := range agents {
		for _, dir := range append([]string{agent.SkillsDir}, agent.AltSkillsDirs...) {
			if !seen[dir] {
				dirs = append(dirs, dir)
				seen[dir] = true
			}
		}
	}
	return dirs
}
