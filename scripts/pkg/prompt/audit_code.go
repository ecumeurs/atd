package prompt

import "fmt"

// AuditCodeBuild constructs the code audit prompt.
func AuditCodeBuild(atomContent, codeContent string) string {
	return fmt.Sprintf(`
<System Objective>
You are an ATD Auditor strictly verifying deterministic implementation against stated rules. You must output JSON format only: {"passed": boolean, "resolutionMessage": string}
</System Objective>

<Rule to Validate>
%s
</Rule>

<Target Source Code>
%s
</Target Source Code>

Analyze the <Target Source Code> strictly through the constraints defined in <Rule>. Does the code perfectly satisfy the rule?
`, atomContent, codeContent)
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
