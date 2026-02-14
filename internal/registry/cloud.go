package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/felixgeelhaar/aios/internal/agents"
)

type SkillVersion struct {
	ID                string
	Version           string
	CompatibleClients []string
	BadgeRequested    bool
	BadgeEvidence     string
}

type CloudRegistry struct {
	items       map[string][]string
	storagePath string
}

func NewCloudRegistry() *CloudRegistry {
	return &CloudRegistry{items: map[string][]string{}}
}

func NewCloudRegistryWithPath(path string) (*CloudRegistry, error) {
	r := &CloudRegistry{items: map[string][]string{}, storagePath: path}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *CloudRegistry) Publish(s SkillVersion) error {
	if err := validatePublishContract(s); err != nil {
		return err
	}
	r.items[s.ID] = append(r.items[s.ID], s.Version)
	if r.storagePath != "" {
		if err := r.persist(); err != nil {
			return err
		}
	}
	return nil
}

func validatePublishContract(s SkillVersion) error {
	if s.ID == "" || s.Version == "" {
		return fmt.Errorf("id and version are required")
	}
	if len(s.CompatibleClients) == 0 {
		return fmt.Errorf("at least one compatible client is required")
	}
	allowed, err := loadAllowedAgentNames()
	if err != nil {
		return fmt.Errorf("loading agent names: %w", err)
	}
	for _, client := range s.CompatibleClients {
		if !allowed[client] {
			return fmt.Errorf("unsupported compatible client %q", client)
		}
	}
	if s.BadgeRequested && s.BadgeEvidence == "" {
		return fmt.Errorf("badge evidence is required when badge is requested")
	}
	return nil
}

func loadAllowedAgentNames() (map[string]bool, error) {
	allAgents, err := agents.LoadAll()
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]bool, len(allAgents))
	for _, a := range allAgents {
		allowed[a.Name] = true
	}
	return allowed, nil
}

func (r *CloudRegistry) Versions(skillID string) []string {
	return r.items[skillID]
}

func (r *CloudRegistry) List() map[string][]string {
	out := map[string][]string{}
	for id, versions := range r.items {
		cp := make([]string, len(versions))
		copy(cp, versions)
		out[id] = cp
	}
	return out
}

func (r *CloudRegistry) persist() error {
	if err := os.MkdirAll(filepath.Dir(r.storagePath), 0o750); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}
	body, err := json.MarshalIndent(r.items, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	if err := os.WriteFile(r.storagePath, body, 0o600); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}
	return nil
}

func (r *CloudRegistry) load() error {
	if r.storagePath == "" {
		return nil
	}
	data, err := os.ReadFile(r.storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read registry: %w", err)
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, &r.items); err != nil {
		return fmt.Errorf("unmarshal registry: %w", err)
	}
	return nil
}
