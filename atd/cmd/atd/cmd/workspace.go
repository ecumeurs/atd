package cmd

import (
	"atd-tools/config"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage multi-project workspaces",
}

var workspaceInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			cwd, _ := os.Getwd()
			name = filepath.Base(cwd)
		}

		ws := config.WorkspaceConfig{
			WorkspaceName: name,
			WorkspaceRoot: ".",
			Projects:      []config.ProjectConfig{},
		}

		data, err := json.MarshalIndent(ws, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(".atd.workspace", data, 0644); err != nil {
			return err
		}

		fmt.Printf("Initialized workspace: %s\n", name)
		return nil
	},
}

var workspaceAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a project to the workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		path, _ := cmd.Flags().GetString("path")

		if name == "" || path == "" {
			return fmt.Errorf("name and path are required")
		}

		ws, err := config.LoadWorkspaceConfig(".")
		if err != nil {
			return fmt.Errorf("no workspace found: %v", err)
		}

		// Check if project already exists
		for _, p := range ws.Projects {
			if p.Name == name {
				return fmt.Errorf("project '%s' already exists", name)
			}
		}

		ws.Projects = append(ws.Projects, config.ProjectConfig{
			Name: name,
			Path: path,
		})

		data, err := json.MarshalIndent(ws, "", "  ")
		if err != nil {
			return err
		}

		workspacePath := filepath.Join(ws.LoadedFrom, ".atd.workspace")
		if err := os.WriteFile(workspacePath, data, 0644); err != nil {
			return err
		}

		fmt.Printf("Added project '%s' at '%s' to workspace\n", name, path)
		return nil
	},
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects in the workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := config.LoadWorkspaceConfig(".")
		if err != nil {
			return fmt.Errorf("no workspace found")
		}

		fmt.Printf("Workspace: %s (%s)\n", ws.WorkspaceName, ws.LoadedFrom)
		fmt.Println("Projects:")
		for _, p := range ws.Projects {
			active := ""
			if config.ActiveConfig.ActiveProject == p.Name {
				active = " (active)"
			}
			fmt.Printf("  - %s: %s%s\n", p.Name, p.Path, active)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
	workspaceCmd.AddCommand(workspaceInitCmd)
	workspaceCmd.AddCommand(workspaceAddCmd)
	workspaceCmd.AddCommand(workspaceListCmd)

	workspaceInitCmd.Flags().String("name", "", "Workspace name")
	workspaceAddCmd.Flags().String("name", "", "Project name")
	workspaceAddCmd.Flags().String("path", "", "Project path")
}
