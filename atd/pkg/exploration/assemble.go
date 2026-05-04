package exploration

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"atd-tools/pkg/atom"
)

type AssembleMetadata struct {
	ID       string `json:"id"`
	Layer    string `json:"layer"`
	Type     string `json:"type"`
	Filepath string `json:"filepath"`
	Project  string `json:"project,omitempty"` // NEW
}

type AssembleJSON struct {
	CustomerLayer       string             `json:"customer_layer,omitempty"`
	ArchitectureLayer   string             `json:"architecture_layer,omitempty"`
	ImplementationLayer string             `json:"implementation_layer,omitempty"`
	Content             string             `json:"content"`
	Metadata            []AssembleMetadata `json:"metadata"`
}

type AssembleOptions struct {
	Starts         string
	Intent         string
	Length         string
	Structured     bool
	AsJSON         bool
	OnlyParents    bool
	OnlyDependents bool
	DocsDir        string
	Workspace      bool // NEW: Allow cross-project assembly
}

func Assemble(opts AssembleOptions) (string, error) {
	if opts.Starts == "" {
		return "", fmt.Errorf("starts parameter is required")
	}

	startIDs := strings.Split(opts.Starts, ",")
	for i := range startIDs {
		startIDs[i] = strings.TrimSpace(startIDs[i])
	}

	graph := &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
	
	// Determine how to load atoms
	if opts.Workspace {
		if err := CrawlWorkspaceDocs(graph); err != nil {
			return "", err
		}
	} else {
		if err := CrawlDocs(opts.DocsDir, graph); err != nil {
			return "", err
		}
	}

	// Build reverse relationships (dependents) for graph traversal
	parentToDependents := make(map[string][]string)
	for id, atom := range graph.Atoms {
		for _, parent := range atom.Parents {
			parentToDependents[parent] = append(parentToDependents[parent], id)
		}
	}
	for id, atom := range graph.Atoms {
		atom.Dependents = parentToDependents[id]
	}

	visited := make(map[string]bool)
	gatheredAtoms := []*atom.AtomData{}
	
	doUp := true
	doDown := true
	if opts.OnlyParents && !opts.OnlyDependents {
		doDown = false
	} else if opts.OnlyDependents && !opts.OnlyParents {
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
			Project:  a.GetProject(),
		})
	}

	if opts.Structured {
		// Multi-pass LLM
		groupedRaw := map[string]string{
			"BUSINESS":       "",
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
			promptLayer := prompt.LayerPassBuild(layer, opts.Length, groupedRaw[layer])
			resp, err := ollama.Query("text_generation", promptLayer, prompt.AssembleFormat())
			if err == ollama.ErrIDEFallback {
				useFallback = true
				return ""
			}
			if err == nil {
				var res struct {
					Document string `json:"document"`
				}
				json.Unmarshal([]byte(resp.Response), &res)
				return res.Document
			}
			return ""
		}

		customerSummary := queryLayer("BUSINESS")
		archSummary := queryLayer("ARCHITECTURE")
		implSummary := queryLayer("IMPLEMENTATION")

		if useFallback {
			finalText := groupedRaw["BUSINESS"] + "\n" + groupedRaw["ARCHITECTURE"] + "\n" + groupedRaw["IMPLEMENTATION"]
			
			if opts.AsJSON {
				out := AssembleJSON{
					CustomerLayer:       groupedRaw["BUSINESS"],
					ArchitectureLayer:   groupedRaw["ARCHITECTURE"],
					ImplementationLayer: groupedRaw["IMPLEMENTATION"],
					Content:             finalText,
					Metadata:            metadata,
				}
				b, _ := json.MarshalIndent(out, "", "  ")
				return string(b), nil
			}
			return finalText + RenderMetadata(metadata), nil
		}

		// Final pass
		finalPrompt := prompt.FinalAssembleBuild(opts.Intent, opts.Length, customerSummary, archSummary, implSummary)
		finalResp, err := ollama.Query("text_generation", finalPrompt, prompt.AssembleFormat())
		var finalContent string
		if err == nil {
			var res struct {
				Document string `json:"document"`
			}
			json.Unmarshal([]byte(finalResp.Response), &res)
			finalContent = res.Document
		}

		if opts.AsJSON {
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

		return finalContent + RenderMetadata(metadata), nil
	}

	// Unstructured mode (Single-pass)
	requestPrompt := prompt.AssembleBuild(opts.Intent, opts.Length, assembledRaw)
	resp, err := ollama.Query("text_generation", requestPrompt, prompt.AssembleFormat())
	
	if err == ollama.ErrIDEFallback {
		taskList, _ := pipeline.WriteTaskList("assemble", []pipeline.PendingTask{
			{
				PromptFile:   "assemble_" + startIDs[0] + ".prompt",
				ResultFile:   "assemble_" + startIDs[0] + ".result",
				Instruction:  "assemble and achieve intent: " + opts.Intent,
				OutputSchema: "markdown text",
			},
		})
		pipeline.WritePromptFile("assemble_"+startIDs[0]+".prompt", requestPrompt)
		
		msg := fmt.Sprintf("Task delegated to IDE Agent: %s", taskList)
		if opts.AsJSON {
			out := AssembleJSON{
				Content:  msg,
				Metadata: metadata,
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			return string(b), nil
		}
		return msg + RenderMetadata(metadata), nil
	}
	
	if err != nil {
		return "", fmt.Errorf("LLM query failed: %v", err)
	}

	var res struct {
		Document string `json:"document"`
	}
	json.Unmarshal([]byte(resp.Response), &res)
	finalContent := res.Document

	if opts.AsJSON {
		out := AssembleJSON{
			Content:  finalContent,
			Metadata: metadata,
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return string(b), nil
	}

	return finalContent + RenderMetadata(metadata), nil
}

func RenderMetadata(meta []AssembleMetadata) string {
	if len(meta) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\n\n---\n**Metadata Index:**\n")
	for _, m := range meta {
		projectInfo := ""
		if m.Project != "" {
			projectInfo = fmt.Sprintf("[%s] ", m.Project)
		}
		out.WriteString(fmt.Sprintf("- %s`[[%s]]` (%s, %s)\n", projectInfo, m.ID, m.Layer, m.Type))
	}
	return out.String()
}
