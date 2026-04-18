package prompt

import (
	"encoding/json"
)

// FixSplitBuild constructs the audit fix split prompt in JSON format.
func FixSplitBuild(content string) string {
	msg := map[string]interface{}{
		"system_objective": "You are an ATD architect. The following ATD atom VIOLATES the 'Minimum Atomic Scale' rule — its RULE/LOGIC section describes more than one distinct state-changing rule.",
		"instruction":      "Split it into N focused child atoms, each with EXACTLY ONE rule in its logic section. Output as structured JSON.",
		"atom_to_split":    content,
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// FixSplitFormat returns the JSON schema for split results.
func FixSplitFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"parent_logic": map[string]string{"type": "string"},
			"splits": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id_suffix":  map[string]string{"type": "string"},
						"human_name": map[string]string{"type": "string"},
						"intent":     map[string]string{"type": "string"},
						"logic":      map[string]string{"type": "string"},
					},
					"required": []string{"id_suffix", "human_name", "intent", "logic"},
				},
			},
		},
		"required": []string{"parent_logic", "splits"},
	}
}
