package cmd
// @spec-link [[mechanic_atd_dissect]]

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"github.com/spf13/cobra"
)

// runDissect dissects target file into atomic boundaries.
// If useLLM is false, returns the raw prompt (passthrough mode for IDE Agent).
// If useLLM is true, routes through the tiered Ollama provider.
func runDissect(target string, useLLM bool) (string, error) {
	if target == "" {
		return "", fmt.Errorf("--file parameter is required")
	}

	content, err := os.ReadFile(target)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %v", target, err)
	}

	// Prepend line numbers
	lines := strings.Split(string(content), "\n")
	var numbered strings.Builder
	for i, line := range lines {
		numbered.WriteString(fmt.Sprintf("%03d: %s\n", i+1, line))
	}

	p := prompt.DissectBuild(numbered.String())

	if !useLLM {
		config.Log("atd-dissect", fmt.Sprintf("Prompt for %s printed to stdout", filepath.Base(target)))
		return p, nil
	}

	// --llm: route through tiered provider
	resp, err := ollama.Query("code_analysis", p, prompt.DissectFormat())
	if err == ollama.ErrIDEFallback {
		basename := strings.TrimSuffix(filepath.Base(target), filepath.Ext(target))
		promptName := "dissect_" + basename
		resultFile := promptName + ".result"

		promptPath, writeErr := pipeline.WritePromptFile(promptName, p)
		if writeErr != nil {
			return "", fmt.Errorf("failed to write prompt file: %v", writeErr)
		}

		tasks := []pipeline.PendingTask{
			{
				PromptFile:   promptPath,
				ResultFile:   resultFile,
				Instruction:  "analyze the document, output JSON atom boundaries",
				OutputSchema: `{"atoms": [{"id": "string", "responsibility": "string", "line_range": [int, int]}]}`,
			},
		}

		taskListPath, writeErr := pipeline.WriteTaskList("atd dissect", tasks)
		if writeErr != nil {
			return "", fmt.Errorf("failed to write task list: %v", writeErr)
		}

		msg := fmt.Sprintf("Task delegated to IDE Agent.\nSee: %s", taskListPath)
		config.Log("atd-dissect", fmt.Sprintf("Delegated dissect of %s to IDE Agent", filepath.Base(target)))
		return msg, nil
	}
	if err != nil {
		return "", fmt.Errorf("LLM query failed: %v", err)
	}

	out, _ := json.MarshalIndent(json.RawMessage(resp.Response), "", "  ")
	config.Log("atd-dissect", fmt.Sprintf("Dissect of %s complete via LLM", filepath.Base(target)))
	return string(out), nil
}

var dissectCmd = &cobra.Command{
	Use:   "dissect",
	Short: "Dissect a file into atomic boundaries",
	Long: `Dissect a document or source code file into atomic boundaries.

By default, outputs the analysis prompt to stdout for the IDE Agent to process.
Use --llm to route through the tiered LLM provider (Ollama).

Examples:
  atd dissect --file ruler.go              # Print prompt to stdout
  atd dissect --file ruler.go --llm        # Use Ollama for extraction
  atd dissect --file commerce.md --llm     # Dissect documentation`,
	RunE: func(cmd *cobra.Command, args []string) error {
		target, _ := cmd.Flags().GetString("file")
		useLLM, _ := cmd.Flags().GetBool("llm")
		out, err := runDissect(target, useLLM)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dissectCmd)
	dissectCmd.Flags().StringP("file", "f", "", "Target file to dissect")
	dissectCmd.Flags().Bool("llm", false, "Route through tiered LLM provider instead of stdout")
}
