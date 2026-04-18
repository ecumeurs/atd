package prompt

import (
	"encoding/json"
)

// CongruenceBuild constructs the congruence audit prompt in JSON format.
func CongruenceBuild(targetAtom, specContents string) string {
	msg := map[string]interface{}{
		"system_objective": "You are the ATD Lead Architect Meta-Auditor. Your absolute priority is System Congruence: verifying that all documentation rules mathematically and logically align perfectly with one another BEFORE any implementation begins.",
		"target_atom":       targetAtom,
		"spec_contents":     specContents,
		"instruction":       "Seek out logical contradictions, missing state resolutions, and mismatched properties strictly between the Target and its siblings.",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// CongruenceFormat returns the JSON schema for congruence results.
func CongruenceFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"is_congruent": map[string]string{"type": "boolean"},
			"audit_report": map[string]string{"type": "string"},
		},
		"required": []string{"is_congruent", "audit_report"},
	}
}
