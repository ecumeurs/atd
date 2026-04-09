package cmd

import (
	"atd-tools/config"
	"atd-tools/pkg/webui"
	"github.com/spf13/cobra"
)

var webuiCmd = &cobra.Command{
	Use:   "webui",
	Short: "Start the ATD WebUI",
	Long: `Start the ATD WebUI to explore and manage your atoms visually.
By default, it uses embedded static files. Use --dev for local development.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		devMode, _ := cmd.Flags().GetBool("dev")
		staticPath, _ := cmd.Flags().GetString("static-path")

		// If staticPath is empty and devMode is true, attempt to find local toolkit path
		if staticPath == "" && devMode {
			staticPath = config.ActiveConfig.WebUI.ToolkitPath
			if staticPath == "" {
				// Fallback to relative path if not configured
				staticPath = "../webui/dist" 
			}
		}

		s := webui.NewServer(devMode, staticPath)

		// Override config with flags if provided
		if cmd.Flags().Changed("host") {
			host, _ := cmd.Flags().GetString("host")
			config.ActiveConfig.WebUI.Host = host
		}
		if cmd.Flags().Changed("port") {
			port, _ := cmd.Flags().GetInt("port")
			config.ActiveConfig.WebUI.Port = port
		}

		return s.Start()
	},
}

func init() {
	rootCmd.AddCommand(webuiCmd)
	webuiCmd.Flags().Bool("dev", false, "Enable development mode (serve from filesystem instead of embed)")
	webuiCmd.Flags().String("static-path", "", "Override path to static files (useful for local development)")
	webuiCmd.Flags().String("host", "", "Host to listen on (default is all interfaces)")
	webuiCmd.Flags().IntP("port", "p", 0, "Port to listen on (default 8080)")
}
