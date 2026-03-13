package prompt

import "fmt"

// FixSplitBuild constructs the audit fix split prompt.
func FixSplitBuild(content string) string {
	return fmt.Sprintf(`You are an ATD (Atomic Traceable Documentation) architect.
The following ATD atom VIOLATES the "Minimum Atomic Scale" rule — its RULE/LOGIC section describes more than one distinct state-changing rule.

Your task: split it into N focused child atoms, each with EXACTLY ONE rule in its logic section.
The original atom will be rewritten as a MODULE parent that aggregates its children.

Rules:
- Each split atom must have a single, non-compound intent sentence
- Each split atom must have no more than one distinct rule in its logic section
- Output ONLY valid JSON. No markdown. No explanation text.

JSON schema:
{
  "parent_logic": "<one-sentence summary of what the children share>",
  "splits": [
    {"id_suffix": "<short_snake_case>", "human_name": "<Human Name>", "intent": "<single purpose sentence>", "logic": "<single rule text>"}
  ]
}

Atom to split:
%s`, content)
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
