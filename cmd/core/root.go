package core

import (
	"github.com/mertbahardogan/escope/internal/connection"
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
		(cmd.Parent() != nil && cmd.Parent().Name() == "config") ||
		(cmd.Name() == "record") ||
		(cmd.Parent() != nil && cmd.Parent().Name() == "record") {
		return nil
	}
	return connection.ApplyPersistentConnection(cmd)
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
