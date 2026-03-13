package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"atd-tools/config"
	"github.com/spf13/cobra"
)

type ElementType string

const (
	TypeGeneric ElementType = "structure"
)

type RoadmapStatus string

const (
	StatusPending    RoadmapStatus = "PENDING"
	StatusDocumented RoadmapStatus = "DOCUMENTED"
)

type RoadmapItem struct {
	FilePath string        `json:"file_path"`
	Line     int           `json:"line"`
	Type     ElementType   `json:"type"`
	Name     string        `json:"name"`
	Keyword  string        `json:"keyword"`
	Status   RoadmapStatus `json:"status"`
	ATDLink  string        `json:"atd_link,omitempty"`
}

type Roadmap struct {
	Items []RoadmapItem `json:"items"`
}

var (
	genericRegex  = regexp.MustCompile(`^\s*(?:export\s+|public\s+|private\s+|protected\s+|static\s+|async\s+)*(class|struct|interface|func|function|def|fn|type|namespace|module|impl)\s+([a-zA-Z0-9_]+)`)
	specLinkRegex = regexp.MustCompile(`@spec-link\s+\[?\[?([^\]\s]+)\]?\]?`)
)

var roadmapCmd = &cobra.Command{
	Use:   "roadmap",
	Short: "Build a documentation roadmap from source code",
	Long: `Scan source code for structural elements and identify gaps 
where ATD documentation is missing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		scanDir, _ := cmd.Flags().GetString("dir")
		if scanDir == "" {
			scanDir = "."
		}
		outPath, _ := cmd.Flags().GetString("out")

		roadmap := Roadmap{
			Items: []RoadmapItem{},
		}

		validExts := config.ActiveConfig.SupportedExtensions

		gitCmd := exec.Command("git", "ls-files", "-c", "-o", "--exclude-standard")
		gitCmd.Dir = scanDir
		out, err := gitCmd.Output()
		if err != nil {
			fmt.Printf("Warning: Failed to run git ls-files in %s. Error: %v\n", scanDir, err)
			return nil
		}

		fileList := strings.Split(strings.TrimSpace(string(out)), "\n")

		for _, relPath := range fileList {
			if relPath == "" {
				continue
			}

			fullPath := filepath.Join(scanDir, relPath)
			ext := filepath.Ext(fullPath)
			if validExts[ext] {
				processGenericFile(fullPath, &roadmap)
			}
		}

		data, err := json.MarshalIndent(roadmap, "", "  ")
		if err != nil {
			return fmt.Errorf("error encoding roadmap: %v", err)
		}

		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("error writing roadmap to %s: %v", outPath, err)
		}

		fmt.Printf("Roadmap generated at %s with %d elements tracked.\n", outPath, len(roadmap.Items))
		return nil
	},
}

func processGenericFile(path string, roadmap *Roadmap) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	var pendingATD string

	for i, line := range lines {
		if match := specLinkRegex.FindStringSubmatch(line); len(match) > 1 {
			pendingATD = match[1]
			continue
		}

		if match := genericRegex.FindStringSubmatch(line); len(match) > 2 {
			keyword := match[1]
			name := match[2]

			if len(name) <= 2 && keyword != "fn" && keyword != "def" {
				continue
			}

			status := StatusPending
			if pendingATD != "" {
				status = StatusDocumented
			}

			roadmap.Items = append(roadmap.Items, RoadmapItem{
				FilePath: path,
				Line:     i + 1,
				Type:     TypeGeneric,
				Keyword:  keyword,
				Name:     name,
				Status:   status,
				ATDLink:  pendingATD,
			})

			pendingATD = ""
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "/*") && !strings.HasPrefix(trimmed, "*") {
			pendingATD = ""
		}
	}
}

func init() {
	rootCmd.AddCommand(roadmapCmd)
	roadmapCmd.Flags().String("dir", ".", "Directory to scan")
	roadmapCmd.Flags().String("out", "roadmap.json", "Output path for the roadmap JSON")
}
