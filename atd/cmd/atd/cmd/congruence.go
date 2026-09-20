package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/llmservice"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"github.com/spf13/cobra"
)

// congruenceFinding is one structured contradiction entry from a
// congruent-false LLM verdict -- see pkg/prompt.CongruenceFormat.
type congruenceFinding struct {
	AtomID        string `json:"atom_id"`
	Section       string `json:"section"`
	Contradiction string `json:"contradiction"`
}

// congruenceResult is the parsed shape of resp.Response, matching
// pkg/prompt.CongruenceFormat's schema.
type congruenceResult struct {
	IsCongruent bool                `json:"is_congruent"`
	AuditReport string              `json:"audit_report"`
	Findings    []congruenceFinding `json:"findings"`
}

// resolveCongruenceTarget determines which docs directory to load atoms
// from and the bare (unqualified) atom id to look up within it, given the
// raw --target value, the --docs override, and the --workspace flag.
//
// A workspace-qualified target ("project:atom_id") is resolved against that
// project's own docs directory via config.ActiveConfig.Workspace (already
// populated by config.LoadFromDir's automatic upward .atd.workspace scan --
// see config.go's LoadFromDirLegacy), following the same
// project-path/docs-path resolution pkg/exploration/crawl.go's
// CrawlWorkspaceDocsWithConfig already uses, rather than a fresh
// workspace.LoadWorkspace(".") call: that literal "." is relative to the
// process's actual working directory, not config's loaded project root, and
// so does not survive testutil.Sandbox-style tests that load config from a
// directory without also os.Chdir-ing the process there.
func resolveCongruenceTarget(targetAtom, docsDirFlag string, workspaceFlag bool) (docsDir string, bareID string, err error) {
	bareID = targetAtom
	project := ""
	if idx := strings.Index(targetAtom, ":"); idx > 0 {
		project = targetAtom[:idx]
		bareID = targetAtom[idx+1:]
	}

	if project == "" && !workspaceFlag {
		if docsDirFlag == "" {
			docsDirFlag = config.DocsDir()
		}
		return docsDirFlag, bareID, nil
	}

	ws := config.ActiveConfig.Workspace
	if ws == nil {
		return "", "", fmt.Errorf("no workspace found but --workspace flag used")
	}

	if project == "" {
		// --workspace was set but the target isn't qualified; fall back to
		// the active project's own docs dir (same as the non-workspace path).
		if docsDirFlag == "" {
			docsDirFlag = config.DocsDir()
		}
		return docsDirFlag, bareID, nil
	}

	for _, p := range ws.Projects {
		if p.Name != project {
			continue
		}
		absProjPath := p.Path
		if !filepath.IsAbs(absProjPath) {
			absProjPath = filepath.Join(ws.LoadedFrom, p.Path)
		}
		pDocs := p.DocsPath
		if pDocs == "" {
			pDocs = "docs/"
		}
		if !filepath.IsAbs(pDocs) {
			pDocs = filepath.Join(absProjPath, pDocs)
		}
		return pDocs, bareID, nil
	}

	return "", "", fmt.Errorf("workspace project %q not found (target %q)", project, targetAtom)
}

var congruenceCmd = &cobra.Command{
	Use:   "congruence",
	Short: "Audit logical consistency between a target atom and its related atoms",
	Long: `Audit logical consistency between a target atom and its related atoms.
Checks parents, dependents, and tag-siblings for contradictions in their INTENT and LOGIC sections.

A verdict of is_congruent:false requires a non-empty structured findings list
(atom id, section, contradiction) -- a bare title-only response is rejected
with a non-zero exit instead of being printed as a complete result.

--target may be workspace-qualified ("project:atom_id"); with --workspace set
(or a "project:" prefix on --target) the atom is resolved against that
project's own docs directory instead of the current project's --docs/default
docs path.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetAtom, _ := cmd.Flags().GetString("target")
		if targetAtom == "" {
			return fmt.Errorf("--target is required")
		}

		docsDirFlag, _ := cmd.Flags().GetString("docs")
		workspaceFlag, _ := cmd.Flags().GetBool("workspace")

		docsDir, bareTarget, err := resolveCongruenceTarget(targetAtom, docsDirFlag, workspaceFlag)
		if err != nil {
			return err
		}

		// 1. Load all ATDs
		files, err := os.ReadDir(docsDir)
		if err != nil {
			return fmt.Errorf("error reading docs directory: %v", err)
		}

		atomMap := make(map[string]string)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".atom.md") {
				path := filepath.Join(docsDir, f.Name())
				content, err := os.ReadFile(path)
				if err == nil {
					id := strings.TrimSuffix(f.Name(), ".atom.md")
					atomMap[id] = string(content)
				}
			}
		}

		targetContent, ok := atomMap[bareTarget]
		if !ok {
			return fmt.Errorf("target atom '%s' not found in %s", targetAtom, docsDir)
		}

		// 2. Crawl relationships
		relatedAtoms := make(map[string]bool)
		relatedAtoms[bareTarget] = true

		linkRegex := regexp.MustCompile(`\[\[(.*?)\]\]`)
		tagRegex := regexp.MustCompile(`tags:\s*\[(.*?)\]`)

		// Extract links from target
		links := linkRegex.FindAllStringSubmatch(targetContent, -1)
		for _, match := range links {
			if len(match) > 1 {
				relatedAtoms[match[1]] = true
			}
		}

		// Extract tags from target
		var targetTags []string
		tagMatch := tagRegex.FindStringSubmatch(targetContent)
		if len(tagMatch) > 1 {
			rawTags := strings.Split(tagMatch[1], ",")
			for _, t := range rawTags {
				targetTags = append(targetTags, strings.TrimSpace(t))
			}
		}

		// Find siblings sharing tags or linking to target
		for id, content := range atomMap {
			if id == bareTarget {
				continue
			}

			// If this file links to the target
			remoteLinks := linkRegex.FindAllStringSubmatch(content, -1)
			for _, match := range remoteLinks {
				if len(match) > 1 && match[1] == bareTarget {
					relatedAtoms[id] = true
				}
			}

			// Or if it shares tags
			if len(targetTags) > 0 {
				rmTagMatch := tagRegex.FindStringSubmatch(content)
				if len(rmTagMatch) > 1 {
					rawTags := strings.Split(rmTagMatch[1], ",")
					for _, t := range rawTags {
						for _, tgtTag := range targetTags {
							if strings.TrimSpace(t) == tgtTag {
								relatedAtoms[id] = true
								break
							}
						}
					}
				}
			}
		}

		// 3. Build Prompt
		var specContents strings.Builder
		for id := range relatedAtoms {
			content, ok := atomMap[id]
			if ok {
				specContents.WriteString(fmt.Sprintf("--- ATOM: %s.atom.md ---\n%s\n\n", id, content))
			}
		}

		requestPrompt := prompt.CongruenceBuild(bareTarget, specContents.String())

		// 4. Resolve and Query
		resp, err := ollama.Query("code_analysis", requestPrompt, prompt.CongruenceFormat())
		if err != nil {
			if errors.Is(err, ollama.ErrIDEFallback) {
				pipeline.WritePromptFile("congruence_"+targetAtom+".prompt", requestPrompt)
				tasks := []pipeline.PendingTask{
					{
						PromptFile:   "congruence_" + targetAtom + ".prompt",
						ResultFile:   "congruence_" + targetAtom + ".result",
						Instruction:  "audit logical consistency between atoms",
						OutputSchema: "markdown table",
					},
				}
				msg, delegateErr := llmservice.HandleIDEFallback(err, "congruence --target "+targetAtom, tasks, "")
				if delegateErr != nil {
					return delegateErr
				}
				fmt.Println(msg)
				return nil
			}
			return fmt.Errorf("ollama query failed: %v", err)
		}

		// 5. Validate: an is_congruent:false verdict with no structured
		// findings is a bare title, not an actionable result -- reject it
		// loudly instead of printing it and returning nil (see
		// failures/20260917_atd_congruence_empty_verdict_and_no_workspace_resolution.md
		// and cmd/atd/cmd/congruence_findings_test.go /
		// congruence_known_defects_test.go).
		var result congruenceResult
		if jsonErr := json.Unmarshal([]byte(resp.Response), &result); jsonErr != nil {
			return fmt.Errorf("congruence: could not parse LLM response as JSON: %v\nraw response: %s", jsonErr, resp.Response)
		}
		if !result.IsCongruent && len(result.Findings) == 0 {
			return fmt.Errorf("congruence: target atom '%s' reported as incongruent (is_congruent: false) but the response carried no structured findings -- nothing to act on\naudit_report: %s", targetAtom, result.AuditReport)
		}

		fmt.Println(resp.Response)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(congruenceCmd)
	congruenceCmd.Flags().String("target", "", "The specific Atom ID to cross-audit")
	congruenceCmd.Flags().String("docs", "", "Path to docs directory")
	congruenceCmd.Flags().Bool("workspace", false, "Resolve a workspace-qualified target (project:atom_id) against that project's docs")
}
