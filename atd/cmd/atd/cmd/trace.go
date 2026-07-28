package cmd

import (
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/llmservice"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var traceCmd = &cobra.Command{
	Use:   "trace <atom_id>",
	Short: "Trace an atom's dependencies and source implementations",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir, _ := cmd.Flags().GetString("docs")
		srcPath, _ := cmd.Flags().GetString("src")
		summary, _ := cmd.Flags().GetBool("summary")

		out, err := runTrace(args[0], docsDir, srcPath, summary)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(traceCmd)
	traceCmd.Flags().String("docs", "", "Override docs directory")
	traceCmd.Flags().String("src", "", "Override source directory (defaults to project root)")
	traceCmd.Flags().Bool("summary", false, "Generate a narrative contextual summary via LLM")
}

func runTrace(targetID, docsDir, srcPath string, summary bool) (string, error) {
	explorer := exploration.NewExplorer(srcPath, docsDir)
	if err := explorer.Load(false); err != nil {
		return "", err
	}

	snap, err := explorer.Trace(targetID)
	if err != nil {
		return "", err
	}

	if summary {
		// Curate the context. IMPORTANT: snap.Context is keyed by the CANONICAL
		// atom id (snap.TargetID), which Trace() resolves internally — it may
		// differ from the raw, possibly bare/aliased `targetID` argument passed
		// in above. Looking this up by the raw targetID silently misses and
		// leaves ctx.Target as a zero-value AtomBrief, which pushes the LLM to
		// narrate the ancestry/dependents (i.e. the parent) as if it were the
		// target. Always look up by snap.TargetID.
		targetBrief, ok := snap.Context[snap.TargetID]
		if !ok {
			// Defensive fallback: Trace() always seeds its own target into
			// Context, so this should not happen. Rather than silently handing
			// the prompt an empty target, rebuild the brief directly or fail
			// loudly.
			node, resolveErr := explorer.ResolveAtom(snap.TargetID)
			if resolveErr != nil {
				return "", fmt.Errorf("cannot build trace summary: no brief available for target '%s': %v", snap.TargetID, resolveErr)
			}
			targetBrief = exploration.AtomBrief{
				ID:        node.ID,
				HumanName: node.HumanName,
				Type:      node.Type,
				Layer:     node.Layer,
				Intent:    node.Intent,
				Logic:     node.Logic,
				FilePath:  node.FilePath,
			}
		}

		ctx := prompt.TraceSummaryContext{
			Target:     prompt.AtomBrief(targetBrief),
			Ancestry:   []prompt.AtomBrief{},
			Dependents: []prompt.AtomBrief{},
		}
		for _, p := range snap.GraphSlice.Parents {
			if brief, ok := snap.Context[p]; ok {
				ctx.Ancestry = append(ctx.Ancestry, prompt.AtomBrief(brief))
			}
		}
		for _, d := range snap.GraphSlice.Dependents {
			if brief, ok := snap.Context[d]; ok {
				ctx.Dependents = append(ctx.Dependents, prompt.AtomBrief(brief))
			}
		}

		tracePrompt := prompt.TraceSummaryBuild(ctx)

		resp, queryErr := ollama.Query("text_generation", tracePrompt, prompt.TraceSummaryFormat())
		if queryErr != nil {
			if errors.Is(queryErr, ollama.ErrIDEFallback) {
				promptFile := "trace_summary_" + targetID
				pipeline.WritePromptFile(promptFile, tracePrompt)
				tasks := []pipeline.PendingTask{
					{
						PromptFile:   promptFile + ".prompt",
						ResultFile:   promptFile + ".result",
						Instruction:  "Generate a narrative trace summary",
						OutputSchema: `{"summary": "string"}`,
					},
				}
				msg, delegateErr := llmservice.HandleIDEFallback(queryErr, "trace --summary "+targetID, tasks, "")
				if delegateErr != nil {
					return "", delegateErr
				}
				// Add the summary field to the snapshot
				snap.Summary = msg
				out, _ := json.MarshalIndent(snap, "", "  ")
				return string(out), nil
			}
			return "", queryErr
		}

		var result struct {
			Summary string `json:"summary"`
		}
		if err := json.Unmarshal([]byte(resp.Response), &result); err != nil {
			// Fallback to raw response
			snap.Summary = resp.Response
		} else {
			snap.Summary = result.Summary
		}
		out, _ := json.MarshalIndent(snap, "", "  ")
		return string(out), nil
	}

	out, _ := json.MarshalIndent(snap, "", "  ")
	return string(out), nil
}

