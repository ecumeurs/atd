package prompt

import "encoding/json"

// AuditBloatBuild constructs the bloat detection prompt in JSON format.
func AuditBloatBuild(role, text string, strictness float64) string {
	msg := map[string]interface{}{
		"role":       role,
		"text":       text,
		"strictness": strictness,
		"instruction": "Evaluate if the provided text is bloated (contains compound rules or broad statements). Respond in structured JSON.",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// AuditBloatFormat returns the JSON schema for bloat detection results.
func AuditBloatFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"is_bloated": map[string]string{"type": "boolean"},
			"reason":     map[string]string{"type": "string"},
		},
		"required": []string{"is_bloated", "reason"},
	}
}
