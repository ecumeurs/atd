package cmd

import (
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

var congruenceCmd = &cobra.Command{
	Use:   "congruence",
	Short: "Audit logical consistency between a target atom and its related atoms",
	Long: `Audit logical consistency between a target atom and its related atoms.
Checks parents, dependents, and tag-siblings for contradictions in their INTENT and LOGIC sections.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetAtom, _ := cmd.Flags().GetString("target")
		if targetAtom == "" {
			return fmt.Errorf("--target is required")
		}

		docsDir, _ := cmd.Flags().GetString("docs")
		if docsDir == "" {
			docsDir = config.DocsDir()
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

		targetContent, ok := atomMap[targetAtom]
		if !ok {
			return fmt.Errorf("target atom '%s' not found in %s", targetAtom, docsDir)
		}

		// 2. Crawl relationships
		relatedAtoms := make(map[string]bool)
		relatedAtoms[targetAtom] = true

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
			if id == targetAtom {
				continue
			}

			// If this file links to the target
			remoteLinks := linkRegex.FindAllStringSubmatch(content, -1)
			for _, match := range remoteLinks {
				if len(match) > 1 && match[1] == targetAtom {
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

		requestPrompt := prompt.CongruenceBuild(targetAtom, specContents.String())

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

		fmt.Println(resp.Response)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(congruenceCmd)
	congruenceCmd.Flags().String("target", "", "The specific Atom ID to cross-audit")
	congruenceCmd.Flags().String("docs", "", "Path to docs directory")
}
