package prompt

import "fmt"

// AuditBloatBuild constructs the bloat detection prompt.
func AuditBloatBuild(role, text string, strictness float64) string {
	fullRole := fmt.Sprintf("%s\n\nStrictness Threshold (0.0 to 1.0): %.1f. At 1.0, aggressively fail broad statements. At lower thresholds, allow contextual grouping.", role, strictness)

	return fmt.Sprintf(`%s

Respond with exactly one word: YES (if it is bloated) or NO (if it is compliant). Do not add any extra text.

TEXT TO EVALUATE:
%s`, fullRole, text)
}

// AuditBloatFormat returns nil for freeform output.
func AuditBloatFormat() interface{} {
	return nil
}
