package prompt

import (
	"encoding/json"
	"fmt"
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
//
// The target atom must be unambiguous and primary in the resulting narrative.
// Earlier revisions gave the target, ancestry, and dependents roughly equal
// billing in the prompt, which let the model drift into summarizing the
// parent instead of the atom actually under assessment. To prevent that, the
// target's id/name/intent are named directly in the system objective, and the
// ancestry/dependents are explicitly labeled as context only.
func TraceSummaryBuild(ctx TraceSummaryContext) string {
	ctxJSON, _ := json.MarshalIndent(ctx, "", "  ")

	targetLabel := ctx.Target.ID
	if ctx.Target.HumanName != "" {
		targetLabel = fmt.Sprintf("%s (%s)", ctx.Target.ID, ctx.Target.HumanName)
	}
	targetIntent := ctx.Target.Intent
	if targetIntent == "" {
		targetIntent = "(no intent recorded)"
	}

	msg := map[string]interface{}{
		"system_objective": fmt.Sprintf(
			"You are an ATD Architect performing an impact assessment. "+
				"The atom under assessment is `%s` — %s. "+
				"Your goal is to provide a narrative summary of THIS atom's role in the vertical graph slice. "+
				"The `ancestry` and `dependents` arrays in context_data are supporting context ONLY — they are "+
				"not the subject of the summary and must never be substituted for the target.",
			targetLabel, targetIntent,
		),
		"context_data": string(ctxJSON),
		"instruction": fmt.Sprintf(
			"Analyze the provided Trace Context and generate a ~500-word summary about `%s` organized as follows. "+
				"The summary MUST be about the target atom `%s` and must name it explicitly in the opening sentence:\n",
			ctx.Target.ID, ctx.Target.ID,
		) +
			"1. Upper Context: Summarize the intent and logic of the ancestry (context only). What high-level business goal does the target atom fulfil?\n" +
			"2. Current Objective: Explain the intent and responsibility of the TARGET atom itself — this is the main subject of the summary.\n" +
			"3. Downward Impact: Look at the dependents (context only). How does the target atom's logic constrain or enable these descendants?\n\n" +
			"Favor the current objective section. If ancestry is thin, treat the target as a new root. If dependents are missing, treat it as a leaf-node. " +
			"Do not summarize the ancestry or a dependent as if it were the target.",
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
