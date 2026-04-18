package prompt

import (
	"encoding/json"
)

// AuditCodeBuild constructs the code audit prompt in JSON format.
func AuditCodeBuild(atomContent, codeContent string) string {
	msg := map[string]interface{}{
		"system_objective": "You are an ATD Auditor strictly verifying deterministic implementation against stated rules. You must output JSON format only: {\"passed\": boolean, \"resolutionMessage\": string}",
		"rule":             atomContent,
		"code":             codeContent,
		"instruction":      "Analyze the code strictly through the constraints defined in the rule. Does the code perfectly satisfy the rule?",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// AuditCodeFormat returns the JSON schema for code audit results.
func AuditCodeFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"passed":            map[string]string{"type": "boolean"},
			"resolutionMessage": map[string]string{"type": "string"},
		},
		"required": []string{"passed", "resolutionMessage"},
	}
}
