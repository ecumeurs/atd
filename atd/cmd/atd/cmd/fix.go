package cmd
// @spec-link [[mechanic_atd_fix]]

import (
	"bufio"
	"database/sql"
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
	_ "github.com/mattn/go-sqlite3"
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Fix bloated atoms by splitting them",
	Long: `Fix parses an audit report and automatically splits atoms marked as [BLOATED].

It uses an LLM to decompose compound rules into focused child atoms,
rewriting the original atom as a MODULE parent.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		auditPath, _ := cmd.Flags().GetString("audit")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		docsDir, _ := cmd.Flags().GetString("docs")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		if auditPath == "" {
			return fmt.Errorf("--audit report path is required")
		}

		return runFix(auditPath, docsDir, dryRun)
	},
}

func parseBloatedFiles(reportPath string) ([]string, error) {
	f, err := os.Open(reportPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var bloated []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[BLOATED]") {
			parts := strings.SplitN(line, "Auditing:", 2)
			if len(parts) == 2 {
				rest := strings.TrimSpace(parts[1])
				filename := strings.Fields(rest)[0]
				bloated = append(bloated, filename)
			}
		}
	}
	return bloated, scanner.Err()
}

type splitResult struct {
	ParentLogic string `json:"parent_logic"`
	Splits      []struct {
		IDSuffix  string `json:"id_suffix"`
		HumanName string `json:"human_name"`
		Intent    string `json:"intent"`
		Logic     string `json:"logic"`
	} `json:"splits"`
}

func runFix(auditPath, docsDir string, dryRun bool) error {
	bloated, err := parseBloatedFiles(auditPath)
	if err != nil {
		return fmt.Errorf("failed to parse audit report: %v", err)
	}

	if len(bloated) == 0 {
		fmt.Println("No bloated atoms found in report.")
		return nil
	}

	fmt.Printf("Found %d bloated atoms to fix.\n\n", len(bloated))

	dbPath := filepath.Join(docsDir, ".atd_audit.db")
	db, _ := sql.Open("sqlite3", dbPath)
	if db != nil {
		defer db.Close()
	}

	for _, filename := range bloated {
		path := filepath.Join(docsDir, filename)
		data, err := atom.Parse(path)
		if err != nil {
			fmt.Printf("[SKIP] %s: %v\n", filename, err)
			continue
		}

		fmt.Printf("── Processing: %s (%s) ──\n", filename, data.ID)

		content, _ := os.ReadFile(path)
		requestPrompt := prompt.FixSplitBuild(string(content))

		resp, err := ollama.Query("fix_split", requestPrompt, prompt.FixSplitFormat())
		if err == ollama.ErrIDEFallback {
			basename := strings.TrimSuffix(filename, filepath.Ext(filename))
			promptName := "fix_split_" + basename
			pipeline.WritePromptFile(promptName, requestPrompt)
			fmt.Printf("  → Task delegated to IDE Agent (IDE Fallback)\n")
			continue
		}
		if err != nil {
			fmt.Printf("  [ERROR] LLM failed: %v\n", err)
			continue
		}

		var sr splitResult
		if err := json.Unmarshal([]byte(resp.Response), &sr); err != nil {
			fmt.Printf("  [ERROR] Failed to parse LLM JSON: %v\n", err)
			continue
		}

		if len(sr.Splits) == 0 {
			fmt.Println("  [WARN] LLM returned no splits.")
			continue
		}

		fmt.Printf("  → %d split(s) proposed:\n", len(sr.Splits))
		for _, s := range sr.Splits {
			newID := data.ID + "_" + s.IDSuffix
			newFilename := newID + ".atom.md"
			fmt.Printf("    • %s  →  %s\n", s.HumanName, newFilename)

			if !dryRun {
				childData := atom.AtomData{
					ID:        newID,
					HumanName: s.HumanName,
					Type:      data.Type,
					Status:    "DRAFT",
					Priority:  "CORE",
					Parents:   []string{data.ID},
					Intent:    s.Intent,
					Logic:     s.Logic,
				}
				content := atom.BuildContent(childData)
				os.WriteFile(filepath.Join(docsDir, newFilename), []byte(content), 0644)
			}
		}

		// Rewrite original as MODULE
		if !dryRun {
			parentIntent := fmt.Sprintf("To aggregate the constituent rules of %s.", data.HumanName)
			content := atom.BuildParentContent(data.ID, data.HumanName, parentIntent, sr.ParentLogic)
			os.WriteFile(path, []byte(content), 0644)
			fmt.Printf("  ✓ Rewrote %s as MODULE parent\n", filename)

			if db != nil {
				db.Exec("DELETE FROM audit_cache WHERE id = ?", filename)
				db.Exec("DELETE FROM collision_cache WHERE file_a = ? OR file_b = ?", filename, filename)
			}
		} else {
			fmt.Printf("  [DRY-RUN] Would rewrite %s as MODULE parent\n", filename)
		}
		fmt.Println()
	}

	return nil
}

func init() {
	rootCmd.AddCommand(fixCmd)
	fixCmd.Flags().String("audit", "", "Path to audit report")
	fixCmd.Flags().Bool("dry-run", false, "Preview changes without writing")
	fixCmd.Flags().String("docs", "", "Path to docs directory")
}
