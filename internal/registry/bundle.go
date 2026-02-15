package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Bundle struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Skills      []BundleSkill     `json:"skills"`
	Version     string            `json:"version"`
	CreatedAt   time.Time         `json:"created_at"`
	Metadata    map[string]string `json:"metadata"`
}

type BundleSkill struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type BundleRegistry struct {
	bundles     map[string]*Bundle
	storagePath string
}

func NewBundleRegistry() *BundleRegistry {
	return &BundleRegistry{bundles: map[string]*Bundle{}}
}

func NewBundleRegistryWithPath(path string) (*BundleRegistry, error) {
	r := &BundleRegistry{bundles: map[string]*Bundle{}, storagePath: path}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *BundleRegistry) Create(bundle Bundle) error {
	if bundle.ID == "" || bundle.Name == "" {
		return fmt.Errorf("id and name are required")
	}
	if len(bundle.Skills) == 0 {
		return fmt.Errorf("at least one skill is required")
	}
	bundle.CreatedAt = time.Now()
	r.bundles[bundle.ID] = &bundle
	if r.storagePath != "" {
		if err := r.persist(); err != nil {
			return err
		}
	}
	return nil
}

func (r *BundleRegistry) Get(id string) (*Bundle, bool) {
	b, ok := r.bundles[id]
	return b, ok
}

func (r *BundleRegistry) List() []*Bundle {
	out := make([]*Bundle, 0, len(r.bundles))
	for _, b := range r.bundles {
		out = append(out, b)
	}
	return out
}

func (r *BundleRegistry) Delete(id string) error {
	if _, ok := r.bundles[id]; !ok {
		return fmt.Errorf("bundle not found: %s", id)
	}
	delete(r.bundles, id)
	if r.storagePath != "" {
		if err := r.persist(); err != nil {
			return err
		}
	}
	return nil
}

func (r *BundleRegistry) AddSkill(bundleID string, skill BundleSkill) error {
	bundle, ok := r.bundles[bundleID]
	if !ok {
		return fmt.Errorf("bundle not found: %s", bundleID)
	}
	if skill.ID == "" || skill.Version == "" {
		return fmt.Errorf("skill id and version are required")
	}
	bundle.Skills = append(bundle.Skills, skill)
	if r.storagePath != "" {
		if err := r.persist(); err != nil {
			return err
		}
	}
	return nil
}

func (r *BundleRegistry) persist() error {
	if err := os.MkdirAll(filepath.Dir(r.storagePath), 0o750); err != nil {
		return fmt.Errorf("create bundle registry dir: %w", err)
	}
	type bundleList struct {
		Version int                `json:"version"`
		Bundles map[string]*Bundle `json:"bundles"`
	}
	data := bundleList{Version: 1, Bundles: r.bundles}
	body, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal bundle registry: %w", err)
	}
	if err := os.WriteFile(r.storagePath, body, 0o600); err != nil {
		return fmt.Errorf("write bundle registry: %w", err)
	}
	return nil
}

func (r *BundleRegistry) load() error {
	if r.storagePath == "" {
		return nil
	}
	data, err := os.ReadFile(r.storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read bundle registry: %w", err)
	}
	if len(data) == 0 {
		return nil
	}
	type bundleList struct {
		Version int                `json:"version"`
		Bundles map[string]*Bundle `json:"bundles"`
	}
	var dataStruct bundleList
	if err := json.Unmarshal(data, &dataStruct); err != nil {
		return fmt.Errorf("unmarshal bundle registry: %w", err)
	}
	if dataStruct.Bundles != nil {
		r.bundles = dataStruct.Bundles
	}
	return nil
}
