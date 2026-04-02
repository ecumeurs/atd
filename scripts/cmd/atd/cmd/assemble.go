package cmd

// @spec-link [[mechanic_atd_assemble]]

import (
	"encoding/json"
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

type AssembleMetadata struct {
	ID       string `json:"id"`
	Layer    string `json:"layer"`
	Type     string `json:"type"`
	Filepath string `json:"filepath"`
}

type AssembleJSON struct {
	CustomerLayer       string             `json:"customer_layer,omitempty"`
	ArchitectureLayer   string             `json:"architecture_layer,omitempty"`
	ImplementationLayer string             `json:"implementation_layer,omitempty"`
	Content             string             `json:"content"`
	Metadata            []AssembleMetadata `json:"metadata"`
}

var assembleCmd = &cobra.Command{
	Use:   "assemble",
	Short: "Stitch ATD atoms together into a cohesive document",
	Long: `Recursively gather ATD atoms starting from specified IDs and 
assemble their content into a single document. Supports structuring by layer and LLM summarization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		starts, _ := cmd.Flags().GetString("starts")
		intent, _ := cmd.Flags().GetString("intent")
		length, _ := cmd.Flags().GetString("length")
		structured, _ := cmd.Flags().GetBool("structured")
		asJSON, _ := cmd.Flags().GetBool("json")

		docsDir := config.DocsDir()

		text, err := runAssemble(starts, intent, length, structured, asJSON, docsDir)
		if err != nil {
			return err
		}
		if text != "" {
			fmt.Println(text)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(assembleCmd)
	assembleCmd.Flags().String("starts", "", "Comma-separated list of Root Atom IDs")
	assembleCmd.Flags().String("intent", "Executive Summary", "The intent the LLM should focus on (e.g., summarize, executive summary)")
	assembleCmd.Flags().String("length", "default", "Length constraint (short, default, extended, long)")
	assembleCmd.Flags().Bool("structured", false, "Group atoms by layer and perform multi-pass summarization")
	assembleCmd.Flags().Bool("json", false, "Output results as JSON")
}

func runAssemble(starts, intent, length string, structured, asJSON bool, docsDir string) (string, error) {
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
	gatheredAtoms := []atom.AtomData{}

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

		gatheredAtoms = append(gatheredAtoms, a)

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

	var metadata []AssembleMetadata
	for _, a := range gatheredAtoms {
		metadata = append(metadata, AssembleMetadata{
			ID:       a.ID,
			Layer:    a.Layer,
			Type:     a.Type,
			Filepath: a.FilePath,
		})
	}

	if structured {
		// Multi-pass LLM
		groupedRaw := map[string]string{
			"CUSTOMER":       "",
			"ARCHITECTURE":   "",
			"IMPLEMENTATION": "",
		}

		for _, a := range gatheredAtoms {
			content, _ := os.ReadFile(a.FilePath)
			layer := strings.ToUpper(a.Layer)
			if layer == "" {
				layer = "IMPLEMENTATION"
			}
			groupedRaw[layer] += string(content) + "\n\n"
		}

		useFallback := false

		// Helper to query LLM for layer
		queryLayer := func(layer string) string {
			if groupedRaw[layer] == "" || useFallback {
				return ""
			}
			promptLayer := prompt.LayerPassBuild(layer, length, groupedRaw[layer])
			resp, err := ollama.Query("assemble_layer_"+layer, promptLayer, nil)
			if err == ollama.ErrIDEFallback {
				useFallback = true
				return ""
			}
			if err == nil {
				return resp.Response
			}
			return ""
		}

		customerSummary := queryLayer("CUSTOMER")
		archSummary := queryLayer("ARCHITECTURE")
		implSummary := queryLayer("IMPLEMENTATION")

		if useFallback {
			// fallback without LLM
			finalText := groupedRaw["CUSTOMER"] + "\n" + groupedRaw["ARCHITECTURE"] + "\n" + groupedRaw["IMPLEMENTATION"]
			
			if asJSON {
				out := AssembleJSON{
					CustomerLayer:       groupedRaw["CUSTOMER"],
					ArchitectureLayer:   groupedRaw["ARCHITECTURE"],
					ImplementationLayer: groupedRaw["IMPLEMENTATION"],
					Content:             finalText,
					Metadata:            metadata,
				}
				b, _ := json.MarshalIndent(out, "", "  ")
				return string(b), nil
			}
			return finalText + renderMetadata(metadata), nil
		}

		// Final pass
		finalPrompt := prompt.FinalAssembleBuild(intent, length, customerSummary, archSummary, implSummary)
		finalResp, err := ollama.Query("assemble_final", finalPrompt, nil)
		var finalContent string
		if err == nil {
			finalContent = finalResp.Response
		}

		if asJSON {
			out := AssembleJSON{
				CustomerLayer:       customerSummary,
				ArchitectureLayer:   archSummary,
				ImplementationLayer: implSummary,
				Content:             finalContent,
				Metadata:            metadata,
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			return string(b), nil
		}

		return finalContent + renderMetadata(metadata), nil
	}

	// Unstructured mode (Single-pass)
	requestPrompt := prompt.AssembleBuild(intent, length, assembledRaw)
	resp, err := ollama.Query("assemble", requestPrompt, nil)
	
	if err == ollama.ErrIDEFallback {
		taskList, _ := pipeline.WriteTaskList("assemble", []pipeline.PendingTask{
			{
				PromptFile:   "assemble_" + startIDs[0] + ".prompt",
				ResultFile:   "assemble_" + startIDs[0] + ".result",
				Instruction:  "assemble and achieve intent: " + intent,
				OutputSchema: "markdown text",
			},
		})
		pipeline.WritePromptFile("assemble_"+startIDs[0]+".prompt", requestPrompt)
		
		msg := fmt.Sprintf("Task delegated to IDE Agent: %s", taskList)
		if asJSON {
			out := AssembleJSON{
				Content:  msg,
				Metadata: metadata,
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			return string(b), nil
		}
		return msg + renderMetadata(metadata), nil
	}
	
	if err != nil {
		return "", fmt.Errorf("LLM query failed: %v", err)
	}

	finalContent := resp.Response

	if asJSON {
		out := AssembleJSON{
			Content:  finalContent,
			Metadata: metadata,
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return string(b), nil
	}

	return finalContent + renderMetadata(metadata), nil
}

func renderMetadata(meta []AssembleMetadata) string {
	if len(meta) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\n\n---\n**Metadata Index:**\n")
	for _, m := range meta {
		out.WriteString(fmt.Sprintf("- `[[%s]]` (%s, %s)\n", m.ID, m.Layer, m.Type))
	}
	return out.String()
}
