package version

import (
	"fmt"

	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/mertbahardogan/escope/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the current version of escope",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("escope version %s\n", version.Version)
	},
}

func init() {
	core.RootCmd.AddCommand(versionCmd)
}
