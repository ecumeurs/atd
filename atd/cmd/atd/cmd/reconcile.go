package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"atd-tools/pkg/llmservice"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"github.com/spf13/cobra"
)

var reconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Match new inbound doc logic against the existing library",
	Long: `Match new inbound doc logic against the existing library, finding redundancies, overlaps, and conflicts.
Returns a JSON mapping of proposed IDs to their relationship with existing atoms.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		newPath, _ := cmd.Flags().GetString("new")
		storePath, _ := cmd.Flags().GetString("store")

		if newPath == "" || storePath == "" {
			return fmt.Errorf("--new and --store are required")
		}

		inboundContent, err := os.ReadFile(newPath)
		if err != nil {
			return fmt.Errorf("failed to read inbound file: %v", err)
		}

		storeContent, err := os.ReadFile(storePath)
		if err != nil {
			return fmt.Errorf("failed to read store file: %v", err)
		}

		requestPrompt := prompt.ReconcileBuild(string(storeContent), string(inboundContent))

		resp, err := ollama.Query("code_analysis", requestPrompt, prompt.ReconcileFormat())
		if err != nil {
			if errors.Is(err, ollama.ErrIDEFallback) {
				pipeline.WritePromptFile("reconcile.prompt", requestPrompt)
				tasks := []pipeline.PendingTask{
					{
						PromptFile:   "reconcile.prompt",
						ResultFile:   "reconcile.result",
						Instruction:  "reconcile inbound edits with existing store",
						OutputSchema: `[{"proposed_id": string, "relationship": "UPDATE|CONFLICT|NEW", "change_context": string}]`,
					},
				}
				msg, delegateErr := llmservice.HandleIDEFallback(err, "reconcile --new "+newPath, tasks, "")
				if delegateErr != nil {
					return delegateErr
				}
				fmt.Println(msg)
				return nil
			}
			return fmt.Errorf("ollama query failed: %v", err)
		}

		var result struct {
			Diffs []interface{} `json:"diffs"`
		}
		// A syntactically-valid JSON *object* that simply lacks a "diffs"
		// key (or carries "diffs": null) unmarshals into result with
		// Diffs == nil and no error -- printing that would be the literal
		// string "null", a zero-value success that neither surfaces the
		// model's actual (unusable) response nor signals that
		// reconciliation produced nothing parseable. Treat it the same as
		// the malformed-JSON case below: salvage the raw response so the
		// operator sees exactly what the model said.
		if err := json.Unmarshal([]byte(resp.Response), &result); err == nil && result.Diffs != nil {
			beauty, _ := json.MarshalIndent(result.Diffs, "", "  ")
			fmt.Println(string(beauty))
		} else {
			var rawSlice []interface{}
			if errS := json.Unmarshal([]byte(resp.Response), &rawSlice); errS == nil {
				beauty, _ := json.MarshalIndent(rawSlice, "", "  ")
				fmt.Println(string(beauty))
			} else {
				fmt.Println(resp.Response)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reconcileCmd)
	reconcileCmd.Flags().String("new", "", "Path to file containing inbound markdown edits")
	reconcileCmd.Flags().String("store", "", "Path to file containing existing atom blocks")
}
