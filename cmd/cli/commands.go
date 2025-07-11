package main

import (
	"fmt"
	"strings"

	"github.com/lerndmina/SSH-Tunnel/internal/config"
	"github.com/lerndmina/SSH-Tunnel/internal/interactive"
	"github.com/lerndmina/SSH-Tunnel/internal/tunnel"
	"github.com/spf13/cobra"
)

// newInteractiveCommand creates the interactive command
func newInteractiveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Start interactive tunnel management mode",
		Long:  `Start the interactive command-line interface for managing SSH tunnels`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return interactive.StartInteractiveMode()
		},
	}
	return cmd
}

// newSetupCommand creates the setup command
func newSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup a new SSH tunnel",
		Long:  `Interactive setup wizard for creating a new SSH tunnel configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return interactive.StartInteractiveMode()
		},
	}

	return cmd
}

// newListCommand creates the list command
func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured tunnels",
		Long:  `Display a list of all configured SSH tunnels with their status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager := config.GetManager()
			configs := configManager.ListConfigs()

			if len(configs) == 0 {
				fmt.Println("No tunnels configured. Run 'ssh-tunnel setup' to create one.")
				return nil
			}

			fmt.Printf("%-20s %-15s %-20s %-10s\n", "NAME", "LOCAL_PORT", "REMOTE_HOST", "STATUS")
			fmt.Println(strings.Repeat("-", 70))

			tunnelManager := tunnel.NewManager()
			for _, name := range configs {
				cfg, err := configManager.GetConfig(name)
				if err != nil {
					fmt.Printf("%-20s %-15s %-20s %-10s\n", name, "ERROR", "ERROR", "ERROR")
					continue
				}

				status := "stopped"
				if tunnelStatus, err := tunnelManager.GetStatus(name); err == nil && tunnelStatus != nil {
					status = tunnelStatus.Status.String()
				}

				fmt.Printf("%-20s %-15s %-20s %-10s\n", 
					name, 
					fmt.Sprintf("%d", cfg.LocalServer.ReversePort), 
					fmt.Sprintf("%s:%d", cfg.CloudServer.IP, cfg.CloudServer.Port),
					status)
			}

			return nil
		},
	}

	return cmd
}

// newStartCommand creates the start command
func newStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [tunnel-name]",
		Short: "Start SSH tunnel(s)",
		Long:  `Start one or more SSH tunnels by name, or all tunnels if no name provided`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tunnelManager := tunnel.NewManager()
			configManager := config.GetManager()
			
			all, _ := cmd.Flags().GetBool("all")
			
			if all || len(args) == 0 {
				// Start all tunnels
				configs := configManager.ListConfigs()
				if len(configs) == 0 {
					fmt.Println("No tunnels configured. Run 'ssh-tunnel setup' to create one.")
					return nil
				}
				
				var errors []string
				for _, name := range configs {
					if err := tunnelManager.Start(name); err != nil {
						errors = append(errors, fmt.Sprintf("%s: %v", name, err))
					} else {
						fmt.Printf("✓ Started tunnel: %s\n", name)
					}
				}
				
				if len(errors) > 0 {
					return fmt.Errorf("failed to start some tunnels:\n%s", strings.Join(errors, "\n"))
				}
				
				return nil
			}
			
			// Start specific tunnel
			tunnelName := args[0]
			if err := tunnelManager.Start(tunnelName); err != nil {
				return fmt.Errorf("failed to start tunnel '%s': %w", tunnelName, err)
			}
			
			fmt.Printf("✓ Started tunnel: %s\n", tunnelName)
			return nil
		},
	}

	cmd.Flags().Bool("all", false, "Start all configured tunnels")
	return cmd
}

// newStopCommand creates the stop command
func newStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [tunnel-name]",
		Short: "Stop SSH tunnel(s)",
		Long:  `Stop one or more SSH tunnels by name, or all tunnels if no name provided`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tunnelManager := tunnel.NewManager()
			configManager := config.GetManager()
			
			all, _ := cmd.Flags().GetBool("all")
			
			if all || len(args) == 0 {
				// Stop all tunnels
				configs := configManager.ListConfigs()
				if len(configs) == 0 {
					fmt.Println("No tunnels configured.")
					return nil
				}
				
				var errors []string
				for _, name := range configs {
					if err := tunnelManager.Stop(name); err != nil {
						errors = append(errors, fmt.Sprintf("%s: %v", name, err))
					} else {
						fmt.Printf("✓ Stopped tunnel: %s\n", name)
					}
				}
				
				if len(errors) > 0 {
					return fmt.Errorf("failed to stop some tunnels:\n%s", strings.Join(errors, "\n"))
				}
				
				return nil
			}
			
			// Stop specific tunnel
			tunnelName := args[0]
			if err := tunnelManager.Stop(tunnelName); err != nil {
				return fmt.Errorf("failed to stop tunnel '%s': %w", tunnelName, err)
			}
			
			fmt.Printf("✓ Stopped tunnel: %s\n", tunnelName)
			return nil
		},
	}

	cmd.Flags().Bool("all", false, "Stop all configured tunnels")
	return cmd
}

// newRestartCommand creates the restart command
func newRestartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart [tunnel-name]",
		Short: "Restart SSH tunnel(s)",
		Long:  `Restart one or more SSH tunnels by name, or all tunnels if no name provided`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("restart command not yet implemented")
		},
	}

	cmd.Flags().Bool("all", false, "Restart all configured tunnels")
	return cmd
}

// newStatusCommand creates the status command
func newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [tunnel-name]",
		Short: "Show tunnel status",
		Long:  `Display the status of one or more SSH tunnels`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tunnelManager := tunnel.NewManager()
			configManager := config.GetManager()
			
			all, _ := cmd.Flags().GetBool("all")
			
			if all || len(args) == 0 {
				// Show status for all tunnels
				configs := configManager.ListConfigs()
				if len(configs) == 0 {
					fmt.Println("No tunnels configured.")
					return nil
				}
				
				fmt.Printf("%-20s %-15s %-15s %-20s\n", "NAME", "STATUS", "UPTIME", "DETAILS")
				fmt.Println(strings.Repeat("-", 75))
				
				for _, name := range configs {
					status, err := tunnelManager.GetStatus(name)
					if err != nil {
						fmt.Printf("%-20s %-15s %-15s %-20s\n", name, "ERROR", "-", err.Error())
						continue
					}
					
					uptime := "-"
					if status != nil && !status.StartTime.IsZero() {
						uptime = status.StartTime.Format("15:04:05")
					}
					
					details := "-"
					if status != nil && status.Error != nil {
						details = status.Error.Error()
					}
					
					statusStr := "stopped"
					if status != nil {
						statusStr = status.Status.String()
					}
					
					fmt.Printf("%-20s %-15s %-15s %-20s\n", name, statusStr, uptime, details)
				}
				
				return nil
			}
			
			// Show status for specific tunnel
			tunnelName := args[0]
			status, err := tunnelManager.GetStatus(tunnelName)
			if err != nil {
				return fmt.Errorf("failed to get status for tunnel '%s': %w", tunnelName, err)
			}
			
			fmt.Printf("Tunnel: %s\n", tunnelName)
			if status != nil {
				fmt.Printf("Status: %s\n", status.Status.String())
				if !status.StartTime.IsZero() {
					fmt.Printf("Started: %s\n", status.StartTime.Format("2006-01-02 15:04:05"))
				}
				if !status.LastHealthCheck.IsZero() {
					fmt.Printf("Last Health Check: %s\n", status.LastHealthCheck.Format("2006-01-02 15:04:05"))
				}
				if status.Error != nil {
					fmt.Printf("Error: %s\n", status.Error.Error())
				}
			} else {
				fmt.Println("Status: stopped")
			}
			
			return nil
		},
	}

	cmd.Flags().Bool("all", false, "Show status for all tunnels")
	cmd.Flags().Bool("watch", false, "Watch status continuously")
	return cmd
}

// newLogsCommand creates the logs command
func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs [tunnel-name]",
		Short: "Show tunnel logs",
		Long:  `Display logs for one or more SSH tunnels`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("logs command not yet implemented")
		},
	}

	cmd.Flags().BoolP("follow", "f", false, "Follow log output")
	cmd.Flags().IntP("lines", "n", 50, "Number of lines to show")
	return cmd
}

// newConfigCommand creates the config command
func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `Commands for managing SSH tunnel configurations`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List configurations",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("config list not yet implemented")
			},
		},
		&cobra.Command{
			Use:   "show [tunnel-name]",
			Short: "Show configuration",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("config show not yet implemented")
			},
		},
		&cobra.Command{
			Use:   "edit [tunnel-name]",
			Short: "Edit configuration",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("config edit not yet implemented")
			},
		},
		&cobra.Command{
			Use:   "delete [tunnel-name]",
			Short: "Delete configuration",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("config delete not yet implemented")
			},
		},
	)

	return cmd
}

// newBackupCommand creates the backup command
func newBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup and restore operations",
		Long:  `Commands for backing up and restoring SSH tunnel configurations`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "create",
			Short: "Create backup",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("backup create not yet implemented")
			},
		},
		&cobra.Command{
			Use:   "restore [backup-file]",
			Short: "Restore from backup",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("backup restore not yet implemented")
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List backups",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("backup list not yet implemented")
			},
		},
	)

	return cmd
}

// newMonitorCommand creates the monitor command
func newMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Real-time monitoring dashboard",
		Long:  `Interactive terminal dashboard for monitoring SSH tunnels`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("monitor command not yet implemented")
		},
	}

	cmd.Flags().Int("refresh", 5, "Refresh interval in seconds")
	return cmd
}

// newDiagnosticsCommand creates the diagnostics command
func newDiagnosticsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnostics [tunnel-name]",
		Short: "Run diagnostics on tunnels",
		Long:  `Run comprehensive diagnostics on SSH tunnels to identify issues`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("diagnostics command not yet implemented")
		},
	}

	cmd.Flags().Bool("performance", false, "Include performance tests")
	cmd.Flags().Bool("connectivity", false, "Test connectivity only")
	return cmd
}

// newRemoteSetupCommand creates the remote setup command
func newRemoteSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote-setup [flags] <host>",
		Short: "Setup a remote server for SSH tunneling",
		Long: `Deploy and run the remote setup script on a cloud server to prepare it for SSH tunneling.

This command will:
- Copy the remote setup script to the target server
- Run the script to install required packages
- Configure SSH daemon for secure tunneling
- Setup firewall rules and security measures
- Create a dedicated tunnel user

Examples:
  ssh-tunnel remote-setup 1.2.3.4
  ssh-tunnel remote-setup --user ubuntu --key ~/.ssh/id_rsa.pub server.example.com`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("remote-setup command not yet implemented")
		},
	}

	cmd.Flags().StringP("user", "u", "root", "SSH user for connecting to remote server")
	cmd.Flags().StringP("key", "k", "", "Path to public key file to deploy")
	cmd.Flags().StringP("tunnel-user", "t", "tunneluser", "Username to create for tunnel connections")
	cmd.Flags().IntP("port", "p", 22, "SSH port on remote server")
	cmd.Flags().Bool("dry-run", false, "Show what would be done without executing")

	return cmd
}

// newTemplateCommand creates the template command
func newTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage configuration templates",
		Long:  `Commands for managing and using configuration templates`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List available templates",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("template list not yet implemented")
			},
		},
		&cobra.Command{
			Use:   "show [template-name]",
			Short: "Show template details",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("template show not yet implemented")
			},
		},
		&cobra.Command{
			Use:   "apply [template-name] [tunnel-name]",
			Short: "Apply template to create new tunnel",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("template apply not yet implemented")
			},
		},
	)

	return cmd
}
