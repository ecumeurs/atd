package prompt

import (
	"encoding/json"
)

// IntentExtractBuild constructs the intent extraction prompt in JSON format.
func IntentExtractBuild(codeContent string) string {
	msg := map[string]interface{}{
		"instruction": "Summarize the architectural and domain-level intent of this code in 2 sentences. Focus on what subsystem it belongs to and the core rules it enforces. Do not talk about specific variable names.",
		"code":        codeContent,
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// IntentExtractFormat returns the JSON schema for intent extraction results.
func IntentExtractFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"intent": map[string]string{"type": "string"},
		},
		"required": []string{"intent"},
	}
}
