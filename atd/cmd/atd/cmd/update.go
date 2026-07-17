package cmd
// @spec-link [[service_atd_update]]

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an ATD atom file's frontmatter or body sections",
	Long: `Update an ATD atom file's frontmatter or body sections.
Supports setting individual frontmatter keys, updating content sections (Intent, Rule/Logic, Technical Interface),
and prepending @spec-link tags to source files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		filterArg, _ := cmd.Flags().GetString("filter")
		intentArg, _ := cmd.Flags().GetString("intent")
		logicArg, _ := cmd.Flags().GetString("logic")
		interfaceArg, _ := cmd.Flags().GetString("interface")
		expectationArg, _ := cmd.Flags().GetString("expectation")
		setArgs, _ := cmd.Flags().GetStringSlice("set")
		specLinkArgs, _ := cmd.Flags().GetStringSlice("spec-link")
		force, _ := cmd.Flags().GetBool("force")

		var specLinkID, specLinkFile string
		if len(specLinkArgs) == 2 {
			specLinkID = specLinkArgs[0]
			specLinkFile = specLinkArgs[1]
		} else if len(specLinkArgs) > 0 {
			return fmt.Errorf("--spec-link requires exactly two arguments: <id> <file>")
		}

		if filterArg != "" {
			if filePath != "" {
				return fmt.Errorf("cannot use both --file and --filter")
			}
			text, err := runBatchUpdate(filterArg, setArgs, intentArg, logicArg, interfaceArg, expectationArg, specLinkID, specLinkFile, force)
			if err != nil {
				return err
			}
			fmt.Println(text)
			return nil
		}

		text, err := runUpdate(filePath, setArgs, intentArg, logicArg, interfaceArg, expectationArg, specLinkID, specLinkFile, force)
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

// force is variadic so existing callers that don't yet pass an explicit
// confirmation flag keep compiling and default to no-force (guard enforced).
func runBatchUpdate(filter string, setArgs []string, intentArg, logicArg, interfaceArg, expectationArg, specLinkID, specLinkFile string, force ...bool) (string, error) {
	filters := make(map[string]string)
	for _, part := range strings.Split(filter, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			filters[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	if len(filters) == 0 {
		return "", fmt.Errorf("invalid filter format, expected key=value,key2=value2")
	}

	docsDir := config.DocsDir()
	yamlRegex := regexp.MustCompile(`(?s)^---[\r\n]+(.*?)[\r\n]+---`)
	var targetFiles []string

	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".atom.md") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		yamlMatch := yamlRegex.FindStringSubmatch(string(content))
		if len(yamlMatch) > 1 {
			frontmatter := yamlMatch[1]
			matchesAll := true
			for k, v := range filters {
				fieldRegex := regexp.MustCompile(fmt.Sprintf(`(?m)^%s:\s*\[?(.*?)\]?$`, regexp.QuoteMeta(k)))
				fieldMatch := fieldRegex.FindStringSubmatch(frontmatter)
				if len(fieldMatch) > 1 {
					if strings.ToLower(strings.TrimSpace(fieldMatch[1])) != strings.ToLower(v) {
						matchesAll = false
						break
					}
				} else {
					matchesAll = false
					break
				}
			}
			if matchesAll {
				targetFiles = append(targetFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error finding matching files: %v", err)
	}

	if len(targetFiles) == 0 {
		return "No atoms matched the given filter.", nil
	}

	var results []string
	successCount := 0
	errorCount := 0

	for _, file := range targetFiles {
		res, err := runUpdate(file, setArgs, intentArg, logicArg, interfaceArg, expectationArg, specLinkID, specLinkFile, force...)
		if err != nil {
			results = append(results, fmt.Sprintf("Error updating %s: %v", file, err))
			errorCount++
		} else {
			results = append(results, res)
			successCount++
		}
	}

	summary := fmt.Sprintf("\nBatch complete: %d updated, %d errors out of %d matching atoms.", successCount, errorCount, len(targetFiles))
	return strings.Join(results, "\n") + summary, nil
}

// force is variadic so existing callers that don't yet pass an explicit
// confirmation flag keep compiling and default to no-force (guard enforced).
func runUpdate(filePath string, setArgs []string, intentArg, logicArg, interfaceArg, expectationArg, specLinkID, specLinkFile string, force ...bool) (string, error) {
	forceFlag := false
	if len(force) > 0 {
		forceFlag = force[0]
	}
	opts := atom.UpdateOptions{
		FilePath:       filePath,
		SetArgs:        setArgs,
		Intent:         intentArg,
		Logic:          logicArg,
		Interface:      interfaceArg,
		Expectation:    expectationArg,
		SpecLinkID:     specLinkID,
		SpecLinkFile:   specLinkFile,
		Force:          forceFlag,
	}
	return atom.Update(opts)
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("file", "f", "", "Path to the ATD file")
	updateCmd.Flags().String("filter", "", "Filter atoms to update instead of a single file (e.g. 'status=DRAFT,type=RULE')")
	updateCmd.Flags().String("intent", "", "New content for ## INTENT section (use '-' for stdin)")
	updateCmd.Flags().String("logic", "", "New content for ## THE RULE / LOGIC section (use '-' for stdin)")
	updateCmd.Flags().String("interface", "", "New content for ## TECHNICAL INTERFACE section (use '-' for stdin)")
	updateCmd.Flags().String("expectation", "", "New content for ## EXPECTATION section (use '-' for stdin)")
	updateCmd.Flags().StringSlice("set", []string{}, "Set frontmatter key=value (can be used multiple times)")
	updateCmd.Flags().StringSlice("spec-link", []string{}, "Inbound ID and source file to tag: --spec-link <id> <file>")
	updateCmd.Flags().Bool("force", false, "Override the STABLE+BUSINESS governance guard and confirm the modification")
}
