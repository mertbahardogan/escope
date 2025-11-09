package core

import (
	"fmt"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/spf13/cobra"
)

var (
	host     string
	username string
	password string
	secure   bool
	alias    string
)

var RootCmd = &cobra.Command{
	Use:                "escope",
	Short:              "escope: Elasticsearch auto diagnostics",
	SilenceErrors:      true,
	SilenceUsage:       true,
	DisableSuggestions: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return validateConfig(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func validateConfig(cmd *cobra.Command) error {
	if cmd.Name() == "escope" || cmd.Name() == "config" || cmd.Name() == "clear" ||
		(cmd.Parent() != nil && cmd.Parent().Name() == "config") {
		return nil
	}

	if alias != "" {
		savedConfig := connection.GetSavedConfig(alias)
		if savedConfig.Host == "" {
			fmt.Printf("Error: Host alias '%s' not found. Available aliases:\n", alias)
			aliases, err := connection.ListSavedConfigs()
			if err == nil && len(aliases) > 0 {
				for _, a := range aliases {
					fmt.Printf("  - %s\n", a)
				}
			} else {
				fmt.Println("No hosts configured. Use 'escope config --help' to set up hosts.")
			}
			return fmt.Errorf("host alias '%s' not found", alias)
		}
		connection.SetConfig(savedConfig)
		return nil
	}

	if host != "" {
		connection.SetConfig(connection.Config{
			Host:     host,
			Username: username,
			Password: password,
			Secure:   secure,
		})
		return nil
	}

	aliases, err := connection.ListSavedConfigs()
	if err != nil || len(aliases) == 0 {
		fmt.Println(constants.ErrNoConfigurationFound)
		fmt.Println(constants.MsgPleaseSetConfiguration)
		fmt.Println(constants.MsgConfigSetExample)
		fmt.Println("")
		fmt.Println(constants.MsgExampleHeader)
		fmt.Println(constants.MsgConfigSetLocalhost)
		fmt.Println(constants.MsgConfigSetSecure)
		fmt.Println("")
		fmt.Println(constants.MsgUseFlagsDirectly)
		fmt.Println(constants.MsgUseFlagsExample)
		return fmt.Errorf("no configuration found")
	}

	activeHost, err := connection.GetActiveHost()
	if err != nil || activeHost == "" {
		fmt.Println("Error: No active host set.")
		fmt.Println("Available hosts:")
		for _, a := range aliases {
			fmt.Printf("  - %s\n", a)
		}
		fmt.Println("")
		fmt.Println("Use 'escope config switch <alias>' to set an active host.")
		return fmt.Errorf("no active host set")
	}

	savedConfig := connection.GetSavedConfig(activeHost)
	if savedConfig.Host == "" {
		fmt.Printf("Error: Active host '%s' not found. Available hosts:\n", activeHost)
		for _, a := range aliases {
			fmt.Printf("  - %s\n", a)
		}
		fmt.Println("")
		fmt.Println("Use 'escope config switch <alias>' to set an active host.")
		return fmt.Errorf("active host '%s' not found", activeHost)
	}
	connection.SetConfig(savedConfig)
	return nil
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "Elasticsearch host address (required for most commands)")
	RootCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Username (required in secure mode)")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password (required in secure mode)")
	RootCmd.PersistentFlags().BoolVar(&secure, "secure", false, "Connect with username and password (default: false)")
	RootCmd.PersistentFlags().StringVarP(&alias, "alias", "a", "", "Use a saved host alias instead of specifying connection details")

}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
	}
}
