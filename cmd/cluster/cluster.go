package cluster

import (
	"fmt"

	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui/tui"
	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:           "cluster",
	Short:         "Show detailed cluster health information",
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		client := elastic.NewClientWrapper(connection.GetClient())
		clusterService := services.NewClusterService(client)

		if err := tui.RunClusterTUI(clusterService); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
		}
	},
}

func init() {
	core.RootCmd.AddCommand(clusterCmd)
}
