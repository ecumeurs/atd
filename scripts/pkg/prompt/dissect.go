package prompt

import "fmt"

// DissectBuild constructs the dissect prompt for a numbered document.
func DissectBuild(numberedContent string) string {
	return fmt.Sprintf(`
<System_Context>
You are an ATD Architect. The target document below has line numbers prepended (e.g., 001:). 
Identify atomic boundaries where a single architectural responsibility starts and ends.
</System_Context>

<Instruction>
1. Map each Atom to its exact line_range [start, end].
2. Identify the 'responsibility' as a deterministic skill definition.
3. Response must strictly follow the JSON schema.
</Instruction>

<Document>
%s
</Document>
`, numberedContent)
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
