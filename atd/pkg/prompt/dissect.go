package prompt

import (
	"encoding/json"
)

// DissectBuild constructs the dissect prompt for a numbered document in JSON format.
func DissectBuild(numberedContent string) string {
	msg := map[string]interface{}{
		"system_context": "You are an ATD Architect. The target document below has line numbers prepended (e.g., 001:). Identify atomic boundaries where a single architectural responsibility starts and ends.",
		"instruction":    "Map each Atom to its exact line_range [start, end]. Identify the 'responsibility' as a deterministic skill definition. Response must strictly follow the JSON schema.",
		"document":       numberedContent,
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// DissectFormat returns the Ollama JSON format schema for dissect output.
func DissectFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"atoms": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":             map[string]string{"type": "string"},
						"responsibility": map[string]string{"type": "string"},
						"line_range": map[string]interface{}{
							"type":     "array",
							"items":    map[string]string{"type": "integer"},
							"minItems": 2,
							"maxItems": 2,
						},
					},
					"required": []string{"id", "responsibility", "line_range"},
				},
			},
		},
	}
}
