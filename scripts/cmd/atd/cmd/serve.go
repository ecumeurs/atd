package cmd

import (
	"fmt"

	"atd-tools/config"
	"atd-tools/pkg/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the ATD MCP server",
	Long: `Start an MCP-compatible server exposing all ATD tools.

By default, the server uses stdio transport — suitable for VS Code, Claude Desktop,
and any MCP host that launches the server as a subprocess.

To use HTTP transport instead (Streamable HTTP, spec 2025-11-25):
  atd serve --http --port 7474

Example .mcp.json for VS Code (stdio, recommended):
  {
    "servers": {
      "atd": {
        "type": "stdio",
        "command": "/path/to/atd",
        "args": ["serve"]
      }
    }
  }

Example .mcp.json for VS Code (HTTP):
  {
    "servers": {
      "atd": {
        "type": "http",
        "url": "http://localhost:7474/mcp"
      }
    }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		useHTTP, _ := cmd.Flags().GetBool("http")
		port, _ := cmd.Flags().GetInt("port")

		// Config must be loaded so all tool handlers can call config.DocsDir() etc.
		if err := config.Load(); err != nil {
			return fmt.Errorf("config: %w", err)
		}

		r := mcp.NewRegistry()
		RegisterMCPTools(r)

		if useHTTP {
			addr := fmt.Sprintf(":%d", port)
			return r.ServeHTTP(addr)
		}

		return r.ServeStdio()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Bool("http", false, "Use Streamable HTTP transport instead of stdio")
	serveCmd.Flags().IntP("port", "p", 7474, "HTTP port (only used with --http)")
}
