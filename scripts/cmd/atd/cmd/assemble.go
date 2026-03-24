package cmd
// @spec-link [[mechanic_atd_assemble]]

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"github.com/spf13/cobra"
)

var assembleCmd = &cobra.Command{
	Use:   "assemble",
	Short: "Stitch ATD atoms together into a cohesive document",
	Long: `Recursively gather ATD atoms starting from specified IDs and 
assemble their content into a single document. Supports narrative snapshots.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		starts, _ := cmd.Flags().GetString("starts")
		purpose, _ := cmd.Flags().GetString("purpose")
		snapshot, _ := cmd.Flags().GetBool("snapshot")
		theme, _ := cmd.Flags().GetString("theme")
		docsDir, _ := cmd.Flags().GetString("docs")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		text, err := runAssemble(starts, purpose, snapshot, theme, docsDir)
		if err != nil {
			return err
		}
		if text != "" {
			fmt.Println(text)
		}
		return nil
	},
}

func runAssemble(starts, purpose string, snapshot bool, theme, docsDir string) (string, error) {
	if starts == "" {
		return "", fmt.Errorf("--starts parameter is required")
	}

	startIDs := strings.Split(starts, ",")
	for i := range startIDs {
		startIDs[i] = strings.TrimSpace(startIDs[i])
	}

	atoms := make(map[string]atom.AtomData)
	filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".atom.md") {
			a, err := atom.Parse(path)
			if err == nil {
				atoms[a.ID] = a
			}
		}
		return nil
	})

	visited := make(map[string]bool)
	var gather func(id string) string
	gather = func(id string) string {
		id = strings.TrimSpace(strings.Trim(id, "[]"))
		if visited[id] {
			return ""
		}
		visited[id] = true
		a, ok := atoms[id]
		if !ok {
			return ""
		}

		// Read full file content safely
		content, _ := os.ReadFile(a.FilePath)
		res := string(content) + "\n\n"
		for _, dep := range a.Dependents {
			res += gather(dep)
		}
		return res
	}

	var assembledRaw string
	for _, startID := range startIDs {
		assembledRaw += gather(startID)
	}

	if snapshot {
		requestPrompt := prompt.SnapshotBuild(theme, assembledRaw)
		resp, err := ollama.Query("snapshot", requestPrompt, nil)
		if err == ollama.ErrIDEFallback {
			taskList, _ := pipeline.WriteTaskList("assemble --snapshot", []pipeline.PendingTask{
				{
					PromptFile:   "snapshot_" + startIDs[0] + ".prompt",
					ResultFile:   "snapshot_" + startIDs[0] + ".result",
					Instruction:  "generate narrative snapshot: " + theme,
					OutputSchema: "markdown narrative",
				},
			})
			pipeline.WritePromptFile("snapshot_"+startIDs[0]+".prompt", requestPrompt)
			return fmt.Sprintf("Task delegated to IDE Agent: %s", taskList), nil
		}
		if err != nil {
			return "", fmt.Errorf("failed to generate snapshot: %v", err)
		}
		return resp.Response, nil
	}

	if purpose != "" {
		return fmt.Sprintf("<System Objective>\nYou are an ATD Assembler. Purpose: %s\n</System Objective>\n\n%s", purpose, assembledRaw), nil
	}

	return assembledRaw, nil
}

func init() {
	rootCmd.AddCommand(assembleCmd)
	assembleCmd.Flags().String("starts", "", "Comma-separated list of Root Atom IDs")
	assembleCmd.Flags().String("purpose", "", "Purpose of the assembly")
	assembleCmd.Flags().Bool("snapshot", false, "Generate a narrative snapshot prompt")
	assembleCmd.Flags().String("theme", "Executive Summary", "Theme for the snapshot")
	assembleCmd.Flags().String("docs", "", "Override docs directory")
}
