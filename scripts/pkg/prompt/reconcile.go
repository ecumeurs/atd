package prompt

import "fmt"

// ReconcileBuild constructs the reconciliation prompt.
func ReconcileBuild(storeContent, inboundContent string) string {
	return fmt.Sprintf(`
<System Objective>
You are an ATD Reconciler managing conflict detection. Analyze the Inbound changes against the Existing knowledge store. Return a Semantic Diff mapping identifying contradictions or overlaps: [{"proposed_id": string, "relationship": "UPDATE|CONFLICT|NEW", "change_context": string}] 
</System Objective>

<Existing Store>
%s
</Existing Store>

<Inbound Edits>
%s
</Inbound Edits>
`, storeContent, inboundContent)
}

// ReconcileFormat returns the JSON schema for reconciliation results.
func ReconcileFormat() interface{} {
	return map[string]interface{}{
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
	}
}
