package cmd

import (
	"errors"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/llmservice"
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

		resp, err := ollama.Query("code_analysis", p, prompt.DissectFormat())
		if err != nil {
			if errors.Is(err, ollama.ErrIDEFallback) {
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

				msg, delegateErr := llmservice.HandleIDEFallback(err, "atd generate", tasks, "")
				if delegateErr != nil {
					return delegateErr
				}
				fmt.Println(msg)
				config.Log("atd-generate", "Delegated generate to IDE Agent")
				return nil
			}
			return fmt.Errorf("LLM query failed: %v", err)
		}

		out, err := json.MarshalIndent(json.RawMessage(resp.Response), "", "  ")
		if err != nil {
			// Surgical fix, same class as cmd/atd/cmd/dissect.go (WP-6/S12,
			// test_atd_07_26.md §3.5): don't discard the error and print an
			// empty line as if generate succeeded — salvage the raw text.
			config.Log("atd-generate", fmt.Sprintf("LLM response was not valid JSON (%v); printing raw text", err))
			fmt.Println(resp.Response)
			return nil
		}
		fmt.Println(string(out))
		config.Log("atd-generate", "Generate complete via LLM")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("dissect", "d", "", "Path to the atd dissect output file (prompt)")
}
