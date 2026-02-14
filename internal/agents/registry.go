// Package agents provides the infrastructure implementation for the agent
// registry. It loads agent definitions from an embedded JSON file and provides
// system-level operations such as agent detection and path resolution.
package agents

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/felixgeelhaar/aios/internal/domain/agentregistry"
)

//go:embed agents.json
var embeddedAgentsJSON []byte

// agentJSON mirrors the JSON schema for deserialization into domain objects.
type agentJSON struct {
	Name            string   `json:"name"`
	DisplayName     string   `json:"displayName"`
	SkillsDir       string   `json:"skillsDir"`
	AltSkillsDirs   []string `json:"altSkillsDirs,omitempty"`
	GlobalSkillsDir string   `json:"globalSkillsDir"`
	DetectPaths     []string `json:"detectPaths"`
	Universal       bool     `json:"universal"`
}

// LoadAll parses the embedded agent definitions and returns validated domain
// objects. Returns an error if the JSON is malformed or any agent fails
// validation.
func LoadAll() ([]agentregistry.AgentDefinition, error) {
	return loadFrom(embeddedAgentsJSON)
}

// loadFrom parses agent definitions from the given JSON bytes.
func loadFrom(data []byte) ([]agentregistry.AgentDefinition, error) {
	var raw []agentJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing agent definitions: %w", err)
	}
	agents := make([]agentregistry.AgentDefinition, 0, len(raw))
	for _, r := range raw {
		a := agentregistry.AgentDefinition{
			Name:            r.Name,
			DisplayName:     r.DisplayName,
			SkillsDir:       r.SkillsDir,
			AltSkillsDirs:   r.AltSkillsDirs,
			GlobalSkillsDir: r.GlobalSkillsDir,
			DetectPaths:     r.DetectPaths,
			Universal:       r.Universal,
		}
		if err := a.Validate(); err != nil {
			return nil, fmt.Errorf("invalid agent definition: %w", err)
		}
		agents = append(agents, a)
	}
	return agents, nil
}

// DetectInstalled returns agent definitions for agents detected on the system
// by checking whether any of their detection paths exist as directories.
func DetectInstalled(agents []agentregistry.AgentDefinition) []agentregistry.AgentDefinition {
	var detected []agentregistry.AgentDefinition
	for _, agent := range agents {
		if isDetected(agent) {
			detected = append(detected, agent)
		}
	}
	return detected
}

// DetectInFolder returns agents detected for a specific project folder.
// It checks both project-local skill directories and global detection paths.
func DetectInFolder(agents []agentregistry.AgentDefinition, folderPath string) []agentregistry.AgentDefinition {
	var detected []agentregistry.AgentDefinition
	for _, agent := range agents {
		skillDir := filepath.Join(folderPath, agent.SkillsDir)
		if dirExists(skillDir) {
			detected = append(detected, agent)
			continue
		}
		for _, alt := range agent.AltSkillsDirs {
			if dirExists(filepath.Join(folderPath, alt)) {
				detected = append(detected, agent)
				break
			}
		}
		if isDetected(agent) {
			detected = append(detected, agent)
		}
	}
	return detected
}

// ResolveGlobalSkillsDir expands environment variables and ~ in the agent's
// global skills directory path.
func ResolveGlobalSkillsDir(agent agentregistry.AgentDefinition) string {
	return expandPath(agent.GlobalSkillsDir)
}

// ResolveProjectSkillsDir returns the absolute path to an agent's skill
// directory within a project folder.
func ResolveProjectSkillsDir(agent agentregistry.AgentDefinition, projectDir string) string {
	return filepath.Join(projectDir, agent.SkillsDir)
}

func isDetected(agent agentregistry.AgentDefinition) bool {
	for _, p := range agent.DetectPaths {
		expanded := expandPath(p)
		if dirExists(expanded) {
			return true
		}
	}
	return false
}

// expandPath expands ~ to home directory and $VAR to environment variable values.
func expandPath(p string) string {
	if strings.Contains(p, "$XDG_CONFIG") {
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			home, _ := os.UserHomeDir()
			xdgConfig = filepath.Join(home, ".config")
		}
		p = strings.ReplaceAll(p, "$XDG_CONFIG", xdgConfig)
	}

	if strings.Contains(p, "$") {
		p = os.Expand(p, func(key string) string {
			if key == "XDG_CONFIG" {
				return ""
			}
			return os.Getenv(key)
		})
	}

	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		p = filepath.Join(home, p[2:])
	} else if p == "~" {
		home, _ := os.UserHomeDir()
		p = home
	}

	return p
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
