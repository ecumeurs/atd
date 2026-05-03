package prompt

import (
	"encoding/json"
)

type AtomBrief struct {
	ID        string `json:"id"`
	HumanName string `json:"human_name"`
	Type      string `json:"type"`
	Layer     string `json:"layer"`
	Intent    string `json:"intent"`
	Logic     string `json:"logic"`
}

// TraceSummaryContext is a curated view of an atom's vertical slice for LLM summarization.
type TraceSummaryContext struct {
	Target     AtomBrief   `json:"target"`
	Ancestry   []AtomBrief `json:"ancestry"`
	Dependents []AtomBrief `json:"dependents"`
}

// TraceSummaryBuild constructs the contextual trace summary prompt.
func TraceSummaryBuild(ctx TraceSummaryContext) string {
	ctxJSON, _ := json.MarshalIndent(ctx, "", "  ")
	msg := map[string]interface{}{
		"system_objective": "You are an ATD Architect performing an impact assessment. Your goal is to provide a narrative summary of an atom's role in the vertical graph slice.",
		"context_data":     string(ctxJSON),
		"instruction": "Analyze the provided Trace Context and generate a ~500-word summary organized as follows:\n" +
			"1. Upper Context: Summarize the intent and logic of the ancestry. What high-level business goal are we fulfilling?\n" +
			"2. Current Objective: Explain the intent of the target atom. Highlight its specific responsibility.\n" +
			"3. Downward Impact: Look at the dependents. How does this atom's logic constrain or enable these descendants?\n\n" +
			"Favor the current objective. If ancestry is thin, focus on the atom as a new root. If descendants are missing, treat it as a leaf-node.",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// TraceSummaryFormat returns the JSON schema for trace summary results.
func TraceSummaryFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"summary": map[string]string{"type": "string"},
		},
		"required": []string{"summary"},
	}
}
