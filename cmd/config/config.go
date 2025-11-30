package config

import (
	"fmt"
	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/mertbahardogan/escope/internal/config"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/spf13/cobra"
	"strconv"
)

var (
	cfgHost     string
	cfgUsername string
	cfgPassword string
	cfgSecure   bool
	cfgAlias    string
	clearConfig bool
)

var configCmd = &cobra.Command{
	Use:           "config",
	Short:         "Set and save connection info for escope, or clear saved config",
	SilenceErrors: true,
	Run: func(cmdx *cobra.Command, args []string) {
		configService := services.NewConfigService()

		if clearConfig {
			if err := configService.ClearConfig(); err != nil {
				fmt.Printf("Failed to clear config: %v\n", err)
				return
			}
			fmt.Println("Connection config cleared.")
			return
		}

		if cfgAlias == "" {
			fmt.Println("Error: Host alias is required. Use alias name to specify a name for this host configuration.")
			fmt.Println("Example: escope config --alias myhost --host http://localhost:9200")
			return
		}

		c := config.ConnectionConfig{
			Host:     cfgHost,
			Username: cfgUsername,
			Password: cfgPassword,
			Secure:   cfgSecure,
		}

		fmt.Println(constants.MsgConnectionTesting)
		if err := configService.SaveHost(cfgAlias, c); err != nil {
			fmt.Printf("Error: %s\n", constants.ErrConnectionTestFailed)
			return
		}
		fmt.Println(constants.MsgConnectionTestPassed)

		aliases, _ := configService.ListHosts()
		if len(aliases) == 1 {
			fmt.Printf("Host '%s' saved successfully and set as active host.\n", cfgAlias)
		} else {
			fmt.Printf("Host '%s' saved successfully.\n", cfgAlias)
		}
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [alias]",
	Short: "Show saved connection configuration for a specific host",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configService := services.NewConfigService()

		var alias string
		if len(args) > 0 {
			alias = args[0]
		} else {
			aliases, err := configService.ListHosts()
			if err != nil {
				fmt.Printf("Failed to list hosts: %v\n", err)
				return
			}

			if len(aliases) == 0 {
				fmt.Println("No hosts configured.")
				return
			}

			if len(aliases) == 1 {
				alias = aliases[0]
			} else {
				fmt.Println("Multiple hosts configured. Please specify an alias:")
				for _, a := range aliases {
					fmt.Printf("  - %s\n", a)
				}
				return
			}
		}

		savedConfig, err := configService.LoadHost(alias)
		if err != nil {
			fmt.Printf("Error: Host '%s' not found. Available hosts:\n", alias)
			aliases, listErr := configService.ListHosts()
			if listErr == nil && len(aliases) > 0 {
				for _, a := range aliases {
					fmt.Printf("  - %s\n", a)
				}
			} else {
				fmt.Println("No hosts configured. Use 'escope config --alias <name> --host <url>' to add a host.")
			}
			return
		}

		fmt.Printf("Configuration for host '%s':\n", alias)
		fmt.Printf(constants.MsgHostLabel+"\n", savedConfig.Host)

		if savedConfig.Secure {
			fmt.Printf(constants.MsgUsernameLabel+"\n", savedConfig.Username)
			if savedConfig.Password != "" {
				fmt.Printf(constants.MsgPasswordLabel+"\n", constants.MsgPasswordHidden)
			} else {
				fmt.Printf(constants.MsgPasswordLabel+"\n", constants.MsgPasswordNotSet)
			}
		}

		fmt.Printf(constants.MsgSecureLabel+"\n", savedConfig.Secure)
	},
}

var configClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear saved connection configuration",
	Run: func(cmd *cobra.Command, args []string) {
		configService := services.NewConfigService()

		if err := configService.ClearConfig(); err != nil {
			fmt.Printf("Error: Failed to clear config: %v\n", err)
			return
		}
		fmt.Println("All configurations cleared.")
	},
}

var configSwitchCmd = &cobra.Command{
	Use:   "switch <alias>",
	Short: "Switch to a specific host alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configService := services.NewConfigService()
		alias := args[0]

		if err := configService.SetActiveHost(alias); err != nil {
			fmt.Printf("Error: Failed to switch to host '%s': %v\n", alias, err)
			return
		}

		fmt.Printf("Switched to host '%s'.\n", alias)
	},
}

var configCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show currently active host",
	Run: func(cmd *cobra.Command, args []string) {
		configService := services.NewConfigService()

		activeHost, err := configService.GetActiveHost()
		if err != nil {
			fmt.Printf("Error: Failed to get active host: %v\n", err)
			return
		}

		if activeHost == "" {
			fmt.Println("No active host set.")
			fmt.Println("Use 'escope config switch <alias>' to set an active host.")
			return
		}

		fmt.Printf("Active host alias: %s\n", activeHost)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured hosts",
	Run: func(cmd *cobra.Command, args []string) {
		configService := services.NewConfigService()

		aliases, err := configService.ListHosts()
		if err != nil {
			fmt.Printf("Error: Failed to list hosts: %v\n", err)
			return
		}

		if len(aliases) == 0 {
			fmt.Println("No hosts configured.")
			fmt.Println("Use 'escope config --alias <name> --host <url>' to add a host.")
			return
		}

		fmt.Println("Configured hosts:")
		for _, alias := range aliases {
			fmt.Printf("  - %s\n", alias)
		}
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete <alias>",
	Short: "Delete a specific host by alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configService := services.NewConfigService()
		alias := args[0]

		if err := configService.DeleteHost(alias); err != nil {
			fmt.Printf("Error: Failed to delete host '%s': %v\n", alias, err)
			return
		}
		fmt.Printf("Host '%s' deleted successfully.\n", alias)
	},
}

var configTimeoutCmd = &cobra.Command{
	Use:   "timeout [seconds]",
	Short: "Set or get connection timeout",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configService := services.NewConfigService()

		if len(args) == 0 {
			timeout, err := configService.GetConnectionTimeout()
			if err != nil {
				fmt.Printf("Error: Failed to get connection timeout: %v\n", err)
				return
			}
			fmt.Printf("Current connection timeout: %d seconds\n", timeout)
			return
		}

		timeoutStr := args[0]
		timeout, err := strconv.Atoi(timeoutStr)
		if err != nil {
			fmt.Printf("Error: Invalid timeout value '%s'. Must be a number.\n", timeoutStr)
			return
		}

		if timeout <= 0 {
			fmt.Printf("Error: Timeout must be greater than 0.\n")
			return
		}

		if err := configService.SetConnectionTimeout(timeout); err != nil {
			fmt.Printf("Error: Failed to set connection timeout: %v\n", err)
			return
		}

		fmt.Printf("Connection timeout set to %d seconds.\n", timeout)
	},
}

func init() {
	configCmd.Flags().StringVar(&cfgHost, "host", "", "Elasticsearch host address (required)")
	configCmd.Flags().StringVar(&cfgUsername, "username", "", "Username (required in secure mode)")
	configCmd.Flags().StringVar(&cfgPassword, "password", "", "Password (required in secure mode)")
	configCmd.Flags().BoolVar(&cfgSecure, "secure", false, "Connect with username and password (default: false)")
	configCmd.Flags().StringVar(&cfgAlias, "alias", "", "Host alias name (required)")
	configCmd.Flags().BoolVar(&clearConfig, "clear", false, "Clear saved connection config")

	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configClearCmd)
	configCmd.AddCommand(configSwitchCmd)
	configCmd.AddCommand(configCurrentCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configTimeoutCmd)
	core.RootCmd.AddCommand(configCmd)
}
