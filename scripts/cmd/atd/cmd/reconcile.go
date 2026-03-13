package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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

		resp, err := ollama.Query("reconcile", requestPrompt, prompt.ReconcileFormat())
		if err == ollama.ErrIDEFallback {
			taskList, _ := pipeline.WriteTaskList("reconcile --new "+newPath, []pipeline.PendingTask{
				{
					PromptFile:   "reconcile.prompt",
					ResultFile:   "reconcile.result",
					Instruction:  "reconcile inbound edits with existing store",
					OutputSchema: `[{"proposed_id": string, "relationship": "UPDATE|CONFLICT|NEW", "change_context": string}]`,
				},
			})
			pipeline.WritePromptFile("reconcile.prompt", requestPrompt)
			fmt.Printf("Task delegated to IDE Agent: %s\n", taskList)
			return nil
		}
		if err != nil {
			return fmt.Errorf("ollama query failed: %v", err)
		}

		// Validate JSON output if possible
		var results []interface{}
		if err := json.Unmarshal([]byte(resp.Response), &results); err != nil {
			// If not valid JSON, just print the response
			fmt.Println(resp.Response)
		} else {
			beauty, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(beauty))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reconcileCmd)
	reconcileCmd.Flags().String("new", "", "Path to file containing inbound markdown edits")
	reconcileCmd.Flags().String("store", "", "Path to file containing existing atom blocks")
}
