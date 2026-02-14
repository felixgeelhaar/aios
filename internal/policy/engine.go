package policy

import "strings"

type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

type RuntimeTelemetry struct {
	Violations []string `json:"violations"`
	Redactions int      `json:"redactions"`
	Blocked    bool     `json:"blocked"`
}

func (e *Engine) Evaluate(text string) []string {
	var violations []string
	if strings.Contains(strings.ToLower(text), "api_key") {
		violations = append(violations, "contains_secret")
	}
	if strings.Contains(strings.ToLower(text), "ignore previous instructions") {
		violations = append(violations, "prompt_injection")
	}
	return violations
}

func (e *Engine) ApplyRuntimeHooks(input map[string]any) (map[string]any, RuntimeTelemetry) {
	telemetry := RuntimeTelemetry{
		Violations: []string{},
	}
	sanitized := sanitizeMap(input, &telemetry)
	if contains(telemetry.Violations, "prompt_injection") {
		telemetry.Blocked = true
	}
	return sanitized, telemetry
}

func sanitizeMap(in map[string]any, telemetry *RuntimeTelemetry) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		out[k] = sanitizeValue(v, telemetry)
	}
	return out
}

func sanitizeValue(v any, telemetry *RuntimeTelemetry) any {
	switch t := v.(type) {
	case string:
		return sanitizeString(t, telemetry)
	case map[string]any:
		return sanitizeMap(t, telemetry)
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = sanitizeValue(t[i], telemetry)
		}
		return out
	default:
		return v
	}
}

func sanitizeString(value string, telemetry *RuntimeTelemetry) string {
	lower := strings.ToLower(value)
	if strings.Contains(lower, "ignore previous instructions") {
		appendViolation(telemetry, "prompt_injection")
	}
	if strings.Contains(lower, "api_key") {
		appendViolation(telemetry, "contains_secret")
		telemetry.Redactions++
		return "[REDACTED_SECRET]"
	}
	return value
}

func appendViolation(telemetry *RuntimeTelemetry, violation string) {
	if !contains(telemetry.Violations, violation) {
		telemetry.Violations = append(telemetry.Violations, violation)
	}
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
