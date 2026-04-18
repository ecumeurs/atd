package prompt

import (
	"encoding/json"
)

// ReconcileBuild constructs the reconciliation prompt in JSON format.
func ReconcileBuild(storeContent, inboundContent string) string {
	msg := map[string]interface{}{
		"system_objective": "You are an ATD Reconciler managing conflict detection. Analyze the Inbound changes against the Existing knowledge store. Return a Semantic Diff mapping identifying contradictions or overlaps.",
		"existing_store":    storeContent,
		"inbound_edits":     inboundContent,
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// ReconcileFormat returns the JSON schema for reconciliation results.
func ReconcileFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"diffs": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"proposed_id":    map[string]string{"type": "string"},
						"relationship":   map[string]interface{}{"type": "string", "enum": []string{"UPDATE", "CONFLICT", "NEW"}},
						"change_context": map[string]string{"type": "string"},
					},
					"required": []string{"proposed_id", "relationship", "change_context"},
				},
			},
		},
		"required": []string{"diffs"},
	}
}
