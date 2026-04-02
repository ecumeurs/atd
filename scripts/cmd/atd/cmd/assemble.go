package cmd

// @spec-link [[mechanic_atd_assemble]]

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/exploration"
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
		onlyParents, _ := cmd.Flags().GetBool("only-parents")
		onlyDependents, _ := cmd.Flags().GetBool("only-dependents")

		docsDir := config.DocsDir()

		text, err := runAssemble(starts, intent, length, structured, asJSON, onlyParents, onlyDependents, docsDir)
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
	assembleCmd.Flags().Bool("only-parents", false, "Restrict traversal to ancestry (upwards) only")
	assembleCmd.Flags().Bool("only-dependents", false, "Restrict traversal to descendants (downwards) only")
}

func runAssemble(starts, intent, length string, structured, asJSON, onlyParents, onlyDependents bool, docsDir string) (string, error) {
	if starts == "" {
		return "", fmt.Errorf("--starts parameter is required")
	}

	startIDs := strings.Split(starts, ",")
	for i := range startIDs {
		startIDs[i] = strings.TrimSpace(startIDs[i])
	}

	graph := &exploration.DependencyGraph{Atoms: make(map[string]*exploration.AtomNode)}
	if err := exploration.CrawlDocs(docsDir, graph); err != nil {
		return "", err
	}

	visited := make(map[string]bool)
	gatheredAtoms := []*exploration.AtomNode{}
	
	doUp := true
	doDown := true
	if onlyParents && !onlyDependents {
		doDown = false
	} else if onlyDependents && !onlyParents {
		doUp = false
	}

	for _, startID := range startIDs {
		startID = strings.TrimSpace(strings.Trim(startID, "[]"))
		
		if doUp {
			graph.WalkUp(startID, visited, func(id string) {
				if node, ok := graph.Atoms[id]; ok {
					gatheredAtoms = append(gatheredAtoms, node)
				}
			})
		}
		if doDown {
			graph.WalkDown(startID, visited, func(id string) {
				if node, ok := graph.Atoms[id]; ok {
					gatheredAtoms = append(gatheredAtoms, node)
				}
			})
		}
	}

	var assembledRaw string
	for _, a := range gatheredAtoms {
		content, err := os.ReadFile(a.FilePath)
		if err == nil {
			assembledRaw += string(content) + "\n\n"
		}
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
