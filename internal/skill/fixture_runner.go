package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FixtureResult struct {
	Name   string
	Passed bool
	Error  string
}

func RunFixtureSuite(skillDir string) ([]FixtureResult, error) {
	spec, err := LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
	if err != nil {
		return nil, err
	}
	if err := ValidateSkillSpec(skillDir, spec); err != nil {
		return nil, err
	}

	testsDir := filepath.Join(skillDir, "tests")
	entries, err := os.ReadDir(testsDir)
	if err != nil {
		return nil, fmt.Errorf("read tests dir: %w", err)
	}

	exec := NewExecutor()
	var results []FixtureResult
	for _, e := range entries {
		name := e.Name()
		if filepath.Ext(name) != ".json" || filepath.Base(name) == "" {
			continue
		}
		if len(name) < len("fixture_") || name[:8] != "fixture_" {
			continue
		}
		fixturePath := filepath.Join(testsDir, name)
		expectedPath := filepath.Join(testsDir, "expected_"+name[len("fixture_"):])

		input, err := readJSONMap(fixturePath)
		if err != nil {
			results = append(results, FixtureResult{Name: name, Passed: false, Error: err.Error()})
			continue
		}
		expected, err := readJSONMap(expectedPath)
		if err != nil {
			results = append(results, FixtureResult{Name: name, Passed: false, Error: err.Error()})
			continue
		}

		out, err := exec.Execute(Artifact{
			ID:           spec.ID,
			Version:      spec.Version,
			InputSchema:  spec.Inputs.Schema,
			OutputSchema: spec.Outputs.Schema,
		}, input)
		if err != nil {
			results = append(results, FixtureResult{Name: name, Passed: false, Error: err.Error()})
			continue
		}

		passed := true
		for k, v := range expected {
			if out[k] != v {
				passed = false
				break
			}
		}
		res := FixtureResult{Name: name, Passed: passed}
		if !passed {
			res.Error = "output mismatch"
		}
		results = append(results, res)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no fixture files found in %s", testsDir)
	}
	return results, nil
}

func readJSONMap(path string) (map[string]any, error) {
	path = filepath.Clean(path)
	// #nosec G304 -- path is derived from validated tests directory entries.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return out, nil
}
