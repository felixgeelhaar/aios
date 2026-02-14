package observability

// Gate defines a measurable exit criterion with a threshold.
type Gate struct {
	Name      string  `json:"name"`
	Metric    string  `json:"metric"`
	Threshold float64 `json:"threshold"`
	Operator  string  `json:"operator"` // "gte", "lte", "gt", "lt", "eq"
}

// GateResult records whether a single gate passed or failed.
type GateResult struct {
	Gate    Gate    `json:"gate"`
	Value   float64 `json:"value"`
	Passed  bool    `json:"passed"`
	Message string  `json:"message"`
}

// EvaluateGates checks a set of gates against provided metric values
// and returns results for each gate. A gate with an unknown operator
// is treated as failing.
func EvaluateGates(gates []Gate, values map[string]float64) []GateResult {
	results := make([]GateResult, 0, len(gates))
	for _, g := range gates {
		v := values[g.Metric]
		passed := evaluate(g.Operator, v, g.Threshold)
		msg := "passed"
		if !passed {
			msg = "failed"
		}
		results = append(results, GateResult{
			Gate:    g,
			Value:   v,
			Passed:  passed,
			Message: msg,
		})
	}
	return results
}

// AllPassed returns true if every gate result passed.
func AllPassed(results []GateResult) bool {
	for _, r := range results {
		if !r.Passed {
			return false
		}
	}
	return true
}

func evaluate(op string, value, threshold float64) bool {
	switch op {
	case "gte":
		return value >= threshold
	case "lte":
		return value <= threshold
	case "gt":
		return value > threshold
	case "lt":
		return value < threshold
	case "eq":
		return value == threshold
	default:
		return false
	}
}
