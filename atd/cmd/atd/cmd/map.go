package cmd

// @spec-link [[service_atd_map]]

import (
	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"atd-tools/pkg/workspace"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// runMap is the core logic for atd map. Routes to confirm, propose, or discover path.
func runMap(filePath, atomID, docsDir string, isNew bool) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("--file is required")
	}

	if docsDir == "" {
		docsDir = config.DocsDir()
	}

	// filePath is expected relative to the active project root (the same
	// convention atd check uses), not the process cwd — under the MCP server
	// the two differ (e.g. cwd is a workspace umbrella root). Resolve before
	// reading so map accepts the same inputs as check.
	readPath := filePath
	if !filepath.IsAbs(readPath) {
		readPath = filepath.Join(config.ProjectRoot(), readPath)
	}

	content, err := os.ReadFile(readPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// Inject file type hint for non-atom .md files
	fileContent := string(content)
	if strings.HasSuffix(filePath, ".md") && !strings.HasSuffix(filePath, ".atom.md") {
		fileContent = "// [CONTEXT: This is a design document]\n" + fileContent
	}

	switch {
	case atomID != "":
		return runMapConfirm(filePath, fileContent, atomID)
	case isNew:
		return runMapPropose(filePath, fileContent)
	default:
		return runMapDiscover(filePath, fileContent, docsDir)
	}
}

// ReconMismatch is one itemized discrepancy between the atom's stated logic
// and the candidate code, as returned by the recon LLM call.
type ReconMismatch struct {
	Aspect   string `json:"aspect"`
	Expected string `json:"expected"`
	Found    string `json:"found"`
}

// ReconResult is the structured recon verdict produced by atd map --atom
// (confirm mode). See pkg/prompt/recon.go for the schema contract.
type ReconResult struct {
	Confidence int             `json:"Confidence"`
	Mismatches []ReconMismatch `json:"Mismatches"`
}

// isDegenerate reports whether mismatches carry no actual rationale (either
// the array is empty, or its entries are all blank placeholders).
func (r ReconResult) isDegenerate() bool {
	if len(r.Mismatches) == 0 {
		return true
	}
	for _, m := range r.Mismatches {
		if strings.TrimSpace(m.Aspect) != "" || strings.TrimSpace(m.Expected) != "" || strings.TrimSpace(m.Found) != "" {
			return false
		}
	}
	return true
}

// runMapConfirm validates whether filePath implements the given atom (recon path).
func runMapConfirm(filePath, fileContent, atomID string) (string, error) {
	atomPath, err := resolveAtomPath(atomID)
	if err != nil {
		return "", err
	}

	atomContent, err := os.ReadFile(atomPath)
	if err != nil {
		return "", fmt.Errorf("failed to read atom file %s: %v", atomPath, err)
	}

	requestPrompt := prompt.ReconBuild(string(atomContent), fileContent)

	resp, err := ollama.Query("code_analysis", requestPrompt, prompt.ReconFormat())
	if err == ollama.ErrIDEFallback {
		pipeline.WritePromptFile("map_confirm", requestPrompt)
		taskList, _ := pipeline.WriteTaskList("map --file "+filePath+" --atom "+atomID, []pipeline.PendingTask{
			{
				PromptFile:   "map_confirm.prompt",
				ResultFile:   "map_confirm.result",
				Instruction:  "validate if code implements the atom",
				OutputSchema: `{"Confidence": int, "Mismatches": [{"aspect": string, "expected": string, "found": string}]}`,
			},
		})
		return fmt.Sprintf("Task delegated to IDE Agent: %s", taskList), nil
	}
	if err != nil {
		return "", fmt.Errorf("ollama query failed: %v", err)
	}

	var result ReconResult
	if err := json.Unmarshal([]byte(resp.Response), &result); err != nil {
		// Model didn't honor the structured schema at all — surface the raw
		// response rather than pretending we parsed a verdict.
		return resp.Response, nil
	}

	if result.Confidence == 0 && result.isDegenerate() {
		return fmt.Sprintf(
			"Recon inconclusive: the model returned no usable rationale for %s against atom %s.\n"+
				"This can happen when the atom's logic lives partly behind a service seam in another "+
				"file this single-file review can't see — retry, or widen the review to include the "+
				"collaborating files.", filePath, atomID), nil
	}

	beauty, _ := json.MarshalIndent(result, "", "  ")
	return string(beauty), nil
}

// runMapPropose extracts intent and returns a new atom skeleton.
func runMapPropose(filePath, fileContent string) (string, error) {
	intentPrompt := prompt.IntentExtractBuild(fileContent)

	resp, err := ollama.Query("text_analysis", intentPrompt, prompt.IntentExtractFormat())
	if err == ollama.ErrIDEFallback {
		pipeline.WritePromptFile("map_propose_intent", intentPrompt)
		taskList, _ := pipeline.WriteTaskList("map --file "+filePath+" --new", []pipeline.PendingTask{
			{
				PromptFile:   "map_propose_intent.prompt",
				ResultFile:   "map_propose_intent.result",
				Instruction:  "extract architectural intent from code to propose a new atom",
				OutputSchema: `{"intent": "string"}`,
			},
		})
		return fmt.Sprintf("Task delegated to IDE Agent: %s", taskList), nil
	}
	if err != nil {
		return "", fmt.Errorf("intent extraction failed: %v", err)
	}

	var intentResult struct {
		Intent string `json:"intent"`
	}
	json.Unmarshal([]byte(resp.Response), &intentResult)
	intent := intentResult.Intent

	// Derive proposed atom fields
	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	base = strings.ToLower(strings.ReplaceAll(base, "-", "_"))
	proposedID := base + "_spec"

	proposedType := "RULE"
	if strings.HasSuffix(filePath, ".md") {
		proposedType = "SPECIFICATION"
	}

	var sb strings.Builder
	sb.WriteString("### Proposed New Atom Skeleton\n\n")
	sb.WriteString("```yaml\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", proposedID))
	sb.WriteString(fmt.Sprintf("type: %s\n", proposedType))
	sb.WriteString("layer: IMPLEMENTATION\n")
	sb.WriteString("status: DRAFT\n")
	sb.WriteString("parents: []\n")
	sb.WriteString("dependents: []\n")
	sb.WriteString("```\n\n")
	sb.WriteString(fmt.Sprintf("**Intent:** %s\n\n", intent))
	sb.WriteString("**Logic:** TODO: define logic constraints\n")
	return sb.String(), nil
}

// runMapDiscover extracts intent, searches the ATD index, and recommends atom links (discover path).
func runMapDiscover(filePath, fileContent, docsDir string) (string, error) {
	fmt.Println("Extracting architectural intent from source code...")
	intentPrompt := prompt.IntentExtractBuild(fileContent)

	var codeIntent string
	resp, err := ollama.Query("text_analysis", intentPrompt, prompt.IntentExtractFormat())
	if err == ollama.ErrIDEFallback {
		codeIntent = "IDE_FALLBACK_PENDING"
	} else if err != nil {
		return "", fmt.Errorf("failed to extract intent: %v", err)
	} else {
		var res struct {
			Intent string `json:"intent"`
		}
		json.Unmarshal([]byte(resp.Response), &res)
		codeIntent = res.Intent
	}

	// Semantic search — workspace-wide if a workspace is active, local otherwise.
	var registryBuilder strings.Builder

	if codeIntent != "IDE_FALLBACK_PENDING" {
		fmt.Println("Embedding code intent for semantic search...")
		dbPath := config.IndexDBPath(docsDir)
		results, err := exploration.Search(exploration.SearchOptions{
			Query:     codeIntent,
			DBPath:    dbPath,
			Limit:     3,
			Scope:     "docs",
			Workspace: config.ActiveConfig.Workspace != nil,
		})
		if err != nil {
			return "", fmt.Errorf("semantic search failed: %v (Did you run 'atd index'?)", err)
		}
		for _, r := range results {
			atomID := strings.TrimSuffix(filepath.Base(r.FilePath), ".atom.md")
			if r.Project != "" {
				atomID = r.Project + ":" + atomID
			}
			registryBuilder.WriteString(fmt.Sprintf("- [[%s]]: %s\n", atomID, r.ChunkText))
		}
	}

	requestPrompt := prompt.DiscoverLinksBuild(registryBuilder.String(), fileContent)

	if codeIntent == "IDE_FALLBACK_PENDING" {
		pipeline.WritePromptFile("map_discover_intent", intentPrompt)
		pipeline.WritePromptFile("map_discover_recommend", requestPrompt)
		taskList, _ := pipeline.WriteTaskList("map --file "+filePath, []pipeline.PendingTask{
			{
				PromptFile:   "map_discover_intent.prompt",
				ResultFile:   "map_discover_intent.result",
				Instruction:  "extract architectural intent from code",
				OutputSchema: `{"intent": "string"}`,
			},
			{
				PromptFile:   "map_discover_recommend.prompt",
				ResultFile:   "map_discover_recommend.result",
				Instruction:  "recommend atom links once intent is known",
				OutputSchema: `{"recommendations": ["atom_id"], "rationale": "string"}`,
			},
		})
		return fmt.Sprintf("Task delegated to IDE Agent (Multi-step): %s", taskList), nil
	}

	fmt.Println("Prompting LLM for final recommendation...")
	respRec, err := ollama.Query("text_analysis", requestPrompt, prompt.DiscoverLinksFormat())
	if err == ollama.ErrIDEFallback {
		pipeline.WritePromptFile("map_discover_recommend", requestPrompt)
		taskList, _ := pipeline.WriteTaskList("map --file "+filePath, []pipeline.PendingTask{
			{
				PromptFile:   "map_discover_recommend.prompt",
				ResultFile:   "map_discover_recommend.result",
				Instruction:  "recommend atom links based on semantic matches",
				OutputSchema: `{"recommendations": ["atom_id"], "rationale": "string"}`,
			},
		})
		return fmt.Sprintf("Recommendation task delegated to IDE Agent: %s", taskList), nil
	}
	if err != nil {
		return "", fmt.Errorf("recommendation query failed: %v", err)
	}

	var rec struct {
		Recommendations []string `json:"recommendations"`
		Rationale       string   `json:"rationale"`
	}
	json.Unmarshal([]byte(respRec.Response), &rec)

	var b strings.Builder
	b.WriteString("### Recommended Atom Links\n")
	for _, r := range rec.Recommendations {
		b.WriteString(fmt.Sprintf("- [[%s]]\n", r))
	}
	b.WriteString("\n### Rationale\n")
	b.WriteString(rec.Rationale)
	return b.String(), nil
}

// resolveAtomPath resolves an atom ID to its file path, workspace-aware.
func resolveAtomPath(atomID string) (string, error) {
	ws, err := workspace.LoadWorkspace(".")
	if err == nil {
		idx, err := ws.BuildIndex()
		if err != nil {
			return "", fmt.Errorf("failed to build workspace index: %v", err)
		}
		currentProj := ws.FindProjectByCWD(".")
		projName := ""
		if currentProj != nil {
			projName = currentProj.Name
		}
		resolver := workspace.NewResolver(ws, idx, projName)
		parsed, err := resolver.Resolve(atomID)
		if err == nil && parsed.Location != nil {
			return parsed.Location.Path, nil
		}
		return "", fmt.Errorf("atom ID %s not found in workspace", atomID)
	}

	// No workspace: try local docs
	docsDir := config.DocsDir()
	candidate := filepath.Join(docsDir, atomID+".atom.md")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	return "", fmt.Errorf("atom %s not found (no workspace and not in local docs)", atomID)
}

var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Map source files to ATD atoms (discover, confirm, or propose)",
	Long: `Extract architectural intent from a source file and recommend matching atoms,
confirm a specific atom match, or propose a new atom skeleton.

Usecases:
  1. Discover:  atd map --file src/foo.go          (find candidate atoms)
  2. Confirm:   atd map --file src/foo.go --atom rule_foo  (validate a specific match)
  3. Propose:   atd map --file src/foo.go --new    (propose a new atom)
  4. Doc file:  atd map --file design.md           (works with .md design docs too)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		atomID, _ := cmd.Flags().GetString("atom")
		docsDir, _ := cmd.Flags().GetString("docs")
		isNew, _ := cmd.Flags().GetBool("new")

		out, err := runMap(filePath, atomID, docsDir, isNew)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mapCmd)
	mapCmd.Flags().String("file", "", "Source file, test file, or .md doc to map (required)")
	mapCmd.Flags().String("atom", "", "Atom ID to confirm against (triggers confirm path)")
	mapCmd.Flags().Bool("new", false, "Skip search and propose a new atom skeleton")
	mapCmd.Flags().String("docs", "", "Override docs directory")
}
