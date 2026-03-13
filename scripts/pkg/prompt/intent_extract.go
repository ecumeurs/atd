package prompt

import "fmt"

// IntentExtractBuild constructs the intent extraction prompt.
func IntentExtractBuild(codeContent string) string {
	return fmt.Sprintf(`Summarize the architectural and domain-level intent of this code in 2 sentences. Focus on what subsystem it belongs to and the core rules it enforces. Do not talk about specific variable names.:
%s`, codeContent)
}

// IntentExtractFormat returns nil for freeform output.
func IntentExtractFormat() interface{} {
	return nil
}
