package marketplace

import (
	"fmt"

	"github.com/felixgeelhaar/aios/internal/agents"
)

type Listing struct {
	SkillID           string
	Version           string
	Verified          bool
	Publisher         string
	CompatibleClients []string
	BadgeEvidence     string
}

type Catalog struct {
	listings []Listing
}

func NewCatalog() *Catalog { return &Catalog{} }

func (c *Catalog) Add(l Listing) error {
	if l.SkillID == "" || l.Version == "" {
		return fmt.Errorf("skill id and version are required")
	}
	if err := validateInstallContract(l); err != nil {
		return err
	}
	c.listings = append(c.listings, l)
	return nil
}

func (c *Catalog) All() []Listing {
	out := make([]Listing, len(c.listings))
	copy(out, c.listings)
	return out
}

func validateInstallContract(l Listing) error {
	if len(l.CompatibleClients) == 0 {
		return fmt.Errorf("listing must define compatible clients")
	}
	allAgents, err := agents.LoadAll()
	if err != nil {
		return fmt.Errorf("loading agent names: %w", err)
	}
	allowed := make(map[string]bool, len(allAgents))
	for _, a := range allAgents {
		allowed[a.Name] = true
	}
	for _, client := range l.CompatibleClients {
		if !allowed[client] {
			return fmt.Errorf("unsupported compatible client %q", client)
		}
	}
	if l.Verified && l.BadgeEvidence == "" {
		return fmt.Errorf("verified listings require badge evidence")
	}
	return nil
}
