package model

import "fmt"

type RouteRequest struct {
	UseCase    string
	Budget     string
	PolicyPack string
}

type PolicyPack struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RouteDecision struct {
	PolicyPack string `json:"policy_pack"`
	Model      string `json:"model"`
	Reason     string `json:"reason"`
}

type Router struct {
	packs []PolicyPack
}

func NewRouter() *Router {
	return &Router{
		packs: []PolicyPack{
			{Name: "cost-first", Description: "Prefer lower-cost models for routine tasks."},
			{Name: "quality-first", Description: "Prefer strongest quality model for critical output."},
			{Name: "balanced", Description: "Balance cost and quality by budget/use case."},
		},
	}
}

func (r *Router) Packs() []PolicyPack {
	out := make([]PolicyPack, len(r.packs))
	copy(out, r.packs)
	return out
}

func (r *Router) Select(req RouteRequest) (string, error) {
	decision, err := r.Decide(req)
	if err != nil {
		return "", err
	}
	return decision.Model, nil
}

func (r *Router) Decide(req RouteRequest) (RouteDecision, error) {
	if req.UseCase == "" {
		return RouteDecision{}, fmt.Errorf("use case is required")
	}
	pack := req.PolicyPack
	if pack == "" {
		pack = "balanced"
	}
	switch pack {
	case "cost-first":
		return RouteDecision{
			PolicyPack: pack,
			Model:      "gpt-4.1-mini",
			Reason:     "cost-first pack prefers lower-cost model",
		}, nil
	case "quality-first":
		return RouteDecision{
			PolicyPack: pack,
			Model:      "gpt-4.1",
			Reason:     "quality-first pack prefers strongest quality model",
		}, nil
	case "balanced":
		if req.Budget == "low" {
			return RouteDecision{
				PolicyPack: pack,
				Model:      "gpt-4.1-mini",
				Reason:     "balanced pack selected lower-cost model for low budget",
			}, nil
		}
		return RouteDecision{
			PolicyPack: pack,
			Model:      "gpt-4.1",
			Reason:     "balanced pack selected higher-quality model for normal budget",
		}, nil
	default:
		return RouteDecision{}, fmt.Errorf("unknown policy pack %q", pack)
	}
}
