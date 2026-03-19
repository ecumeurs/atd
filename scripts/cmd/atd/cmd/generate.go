package cmd
// @spec-link [[atd_generate]]

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

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate atom boundaries from a dissect output file",
	Long: `Generate atom boundaries from a dissect output file.

Takes the prompt output from 'atd dissect' and sends it to the LLM
for structured JSON extraction of atom boundaries.

Example:
  atd dissect --file ruler.go > /tmp/dissect_output.txt
  atd generate --dissect /tmp/dissect_output.txt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dissectFile, _ := cmd.Flags().GetString("dissect")

		if dissectFile == "" {
			return fmt.Errorf("--dissect parameter is required")
		}

		promptData, err := os.ReadFile(dissectFile)
		if err != nil {
			return fmt.Errorf("failed to read %s: %v", dissectFile, err)
		}

		p := string(promptData)

		resp, err := ollama.Query("dissect", p, prompt.DissectFormat())
		if err == ollama.ErrIDEFallback {
			basename := strings.TrimSuffix(filepath.Base(dissectFile), filepath.Ext(dissectFile))
			promptName := "generate_" + basename
			resultFile := promptName + ".result"

			promptPath, writeErr := pipeline.WritePromptFile(promptName, p)
			if writeErr != nil {
				return fmt.Errorf("failed to write prompt file: %v", writeErr)
			}

			tasks := []pipeline.PendingTask{
				{
					PromptFile:   promptPath,
					ResultFile:   resultFile,
					Instruction:  "extract atom boundaries from the prompt, output JSON",
					OutputSchema: `{"atoms": [{"id": "string", "responsibility": "string", "line_range": [int, int]}]}`,
				},
			}

			taskListPath, writeErr := pipeline.WriteTaskList("atd generate", tasks)
			if writeErr != nil {
				return fmt.Errorf("failed to write task list: %v", writeErr)
			}

			fmt.Printf("Task delegated to IDE Agent.\nSee: %s\n", taskListPath)
			config.Log("atd-generate", "Delegated generate to IDE Agent")
			return nil
		}
		if err != nil {
			return fmt.Errorf("LLM query failed: %v", err)
		}

		out, _ := json.MarshalIndent(json.RawMessage(resp.Response), "", "  ")
		fmt.Println(string(out))
		config.Log("atd-generate", "Generate complete via LLM")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("dissect", "d", "", "Path to the atd dissect output file (prompt)")
}
