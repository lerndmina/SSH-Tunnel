package main

import (
	"fmt"
	"os"

	"github.com/ssh-tunnel-manager/ssh-tunnel-manager/internal/config"
	"github.com/ssh-tunnel-manager/ssh-tunnel-manager/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	var configPath string
	var verbose bool

	rootCmd := &cobra.Command{
		Use:   "ssh-tunnel",
		Short: "Advanced SSH Tunnel Manager",
		Long: `A comprehensive cross-platform tool for managing persistent SSH tunnels.
		
Features:
- Multi-tunnel management
- Cross-platform service integration
- Real-time monitoring and analytics
- Configuration templates
- Backup and restore
- Interactive TUI interface`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize logger
			if verbose {
				logger.SetLevel(logger.DebugLevel)
			}
			
			// Load configuration
			if err := config.Initialize(configPath); err != nil {
				return fmt.Errorf("failed to initialize configuration: %w", err)
			}
			
			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(
		newSetupCommand(),
		newListCommand(),
		newStartCommand(),
		newStopCommand(),
		newRestartCommand(),
		newStatusCommand(),
		newLogsCommand(),
		newConfigCommand(),
		newBackupCommand(),
		newMonitorCommand(),
		newDiagnosticsCommand(),
		newTemplateCommand(),
	)

	return rootCmd
}
