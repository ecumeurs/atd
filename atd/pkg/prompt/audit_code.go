package prompt

import (
	"encoding/json"
)

// AuditPersona represents the specialized role for the auditor.
type AuditPersona string

const (
	PersonaPM       AuditPersona = "PRODUCT_MANAGER"
	PersonaTechLead AuditPersona = "TECH_LEAD"
)

// CuratedAuditAtom is a stripped-down atom view for semantic auditing.
type CuratedAuditAtom struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Layer       string `json:"layer"`
	Intent      string `json:"intent"`
	Logic       string `json:"logic"`
	Expectation string `json:"expectation"`
}

// AuditCodeBuild constructs the code audit prompt in JSON format.
func AuditCodeBuild(persona AuditPersona, atom CuratedAuditAtom, codeContent string) string {
	systemObjective := "You are an ATD Auditor strictly verifying implementation against stated rules. You MUST NOT infer any information not explicitly present in the provided rule. If the documentation is missing a detail, the audit must fail or note the ambiguity."

	var instruction string
	switch persona {
	case PersonaPM:
		systemObjective = "You are a Synthetic Product Manager (PM). Your goal is to ensure that the code fulfills the BUSINESS INTENT and high-level logic. Ignore minor technical implementation details; focus on functional correctness and customer value."
		instruction = "Does this code implement the business intent and logic described? Focus on the 'What' and 'Why'."
	case PersonaTechLead:
		systemObjective = "You are a Synthetic Tech Lead. Your goal is to ensure technical rigor and interface compliance. Verify that the logic is correctly implemented at the code level, checking for edge cases defined in the rule."
		instruction = "Does this code perfectly satisfy the technical logic and expectations? Focus on the 'How' and the strict boundaries."
	}

	atomJSON, _ := json.MarshalIndent(atom, "", "  ")

	msg := map[string]interface{}{
		"system_objective": systemObjective,
		"persona":          string(persona),
		"rule_context":     string(atomJSON),
		"code":             codeContent,
		"instruction":      instruction,
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
