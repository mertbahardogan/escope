package upgrade

import (
	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/mertbahardogan/escope/internal/upgrade"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade escope to the latest version",
	Long:  "Check for updates and upgrade escope to the latest version using go install",
	Run: func(cmd *cobra.Command, args []string) {
		upgrade.CheckAndUpgrade()
	},
}

func init() {
	core.RootCmd.AddCommand(upgradeCmd)
}
