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

// runRecon validates whether a candidate source file implements a specific atom.
func runRecon(atomPath, candidatePath string) (string, error) {
	if atomPath == "" || candidatePath == "" {
		return "", fmt.Errorf("--atom and --candidate are required")
	}

	atomContent, err := os.ReadFile(atomPath)
	if err != nil {
		return "", fmt.Errorf("failed to read atom file: %v", err)
	}

	candidateContent, err := os.ReadFile(candidatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read candidate code file: %v", err)
	}

	requestPrompt := prompt.ReconBuild(string(atomContent), string(candidateContent))

	resp, err := ollama.Query("recon", requestPrompt, prompt.ReconFormat())
	if err == ollama.ErrIDEFallback {
		taskList, _ := pipeline.WriteTaskList("recon --atom "+atomPath, []pipeline.PendingTask{
			{
				PromptFile:   "recon.prompt",
				ResultFile:   "recon.result",
				Instruction:  "validate if code implements atom",
				OutputSchema: `{"Confidence": int, "Mismatches": string}`,
			},
		})
		pipeline.WritePromptFile("recon.prompt", requestPrompt)
		return fmt.Sprintf("Task delegated to IDE Agent: %s", taskList), nil
	}
	if err != nil {
		return "", fmt.Errorf("ollama query failed: %v", err)
	}

	var result interface{}
	if err := json.Unmarshal([]byte(resp.Response), &result); err != nil {
		return resp.Response, nil
	}
	beauty, _ := json.MarshalIndent(result, "", "  ")
	return string(beauty), nil
}

var reconCmd = &cobra.Command{
	Use:   "recon",
	Short: "Validate if code implements a specific atom",
	Long:  `The Semantic Archaeology Engine. Helps developers validate if existing unmapped source code is an implementation of a specific Atom constraint.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		atomPath, _ := cmd.Flags().GetString("atom")
		candidatePath, _ := cmd.Flags().GetString("candidate")
		out, err := runRecon(atomPath, candidatePath)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reconCmd)
	reconCmd.Flags().String("atom", "", "Path to the target atom file")
	reconCmd.Flags().String("candidate", "", "Path to the candidate source code file")
}
