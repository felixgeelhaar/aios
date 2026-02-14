package rollout

import "fmt"

type Bundle struct {
	Name   string
	Skills []string
}

type Plan struct {
	BundleName string
	Targets    []string
}

func BuildPlan(b Bundle, targets []string) (Plan, error) {
	if b.Name == "" {
		return Plan{}, fmt.Errorf("bundle name is required")
	}
	if len(b.Skills) == 0 {
		return Plan{}, fmt.Errorf("bundle must include skills")
	}
	if len(targets) == 0 {
		return Plan{}, fmt.Errorf("targets are required")
	}
	return Plan{BundleName: b.Name, Targets: targets}, nil
}
