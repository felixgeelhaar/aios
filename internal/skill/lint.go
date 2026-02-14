package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type LintResult struct {
	Valid  bool
	Issues []string
}

// credentialPatterns defines patterns that indicate embedded credentials
// or secrets in skill files. Skills must declare required connectors and
// rely on the runtime token store — never embed raw credentials.
var credentialPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bapi[_-]?key\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)\bsecret[_-]?key\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)\bclient[_-]?secret\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)\baccess[_-]?token\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)\bprivate[_-]?key\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)\bpassword\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9\-._~+/]+=*`),
	regexp.MustCompile(`(?i)\bsk-[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`(?i)\bghp_[A-Za-z0-9]{36,}`),
	regexp.MustCompile(`(?i)\bAIza[A-Za-z0-9\-_]{30,}`),
}

func LintSkillDir(skillDir string) (LintResult, error) {
	spec, err := LoadSkillSpec(filepath.Join(skillDir, "skill.yaml"))
	if err != nil {
		return LintResult{}, err
	}
	issues := []string{}
	if err := ValidateSkillSpec(skillDir, spec); err != nil {
		issues = append(issues, err.Error())
	}
	if _, err := os.Stat(filepath.Join(skillDir, "prompt.md")); err != nil {
		issues = append(issues, "missing prompt.md")
	}
	if _, err := os.Stat(filepath.Join(skillDir, "tests")); err != nil {
		issues = append(issues, "missing tests directory")
	} else {
		entries, err := os.ReadDir(filepath.Join(skillDir, "tests"))
		if err != nil {
			issues = append(issues, "cannot read tests directory")
		} else {
			fixtures := map[string]bool{}
			expecteds := map[string]bool{}
			for _, e := range entries {
				name := e.Name()
				switch {
				case strings.HasPrefix(name, "fixture_") && strings.HasSuffix(name, ".json"):
					fixtures[strings.TrimPrefix(name, "fixture_")] = true
				case strings.HasPrefix(name, "expected_") && strings.HasSuffix(name, ".json"):
					expecteds[strings.TrimPrefix(name, "expected_")] = true
				}
			}
			for suffix := range fixtures {
				if !expecteds[suffix] {
					issues = append(issues, "missing expected_"+suffix)
				}
			}
			for suffix := range expecteds {
				if !fixtures[suffix] {
					issues = append(issues, "missing fixture_"+suffix)
				}
			}
		}
	}

	// Scan all skill files for embedded credentials.
	issues = append(issues, scanForCredentials(skillDir)...)

	return LintResult{Valid: len(issues) == 0, Issues: issues}, nil
}

// scanForCredentials inspects all text files in the skill directory for
// credential patterns. Skills must never embed raw tokens, OAuth client
// secrets, or private keys — connectors provide credentials at runtime.
func scanForCredentials(skillDir string) []string {
	var issues []string
	scanFiles := []string{
		"skill.yaml",
		"prompt.md",
	}
	// Include schema files and any other yaml/json/md files at root level.
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return issues
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".md") {
			found := false
			for _, f := range scanFiles {
				if f == name {
					found = true
					break
				}
			}
			if !found {
				scanFiles = append(scanFiles, name)
			}
		}
	}
	// Also scan test fixtures.
	testsDir := filepath.Join(skillDir, "tests")
	if testEntries, err := os.ReadDir(testsDir); err == nil {
		for _, e := range testEntries {
			if !e.IsDir() {
				scanFiles = append(scanFiles, filepath.Join("tests", e.Name()))
			}
		}
	}

	for _, relPath := range scanFiles {
		absPath := filepath.Join(skillDir, relPath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}
		content := string(data)
		for _, pattern := range credentialPatterns {
			if pattern.MatchString(content) {
				issues = append(issues, fmt.Sprintf("embedded credential detected in %s: pattern %s", relPath, pattern.String()))
				break // One issue per file is sufficient.
			}
		}
	}
	return issues
}

func (r LintResult) Err() error {
	if r.Valid {
		return nil
	}
	return fmt.Errorf("lint failed: %d issue(s)", len(r.Issues))
}
