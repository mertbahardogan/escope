package calculator

import (
	"context"
	"fmt"

	"github.com/mertbahardogan/escope/cmd/core"
	escalc "github.com/mertbahardogan/escope/internal/calculator"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui/tui"
	"github.com/mertbahardogan/escope/internal/util"
	"github.com/spf13/cobra"
)

var fromCluster bool

var calculatorCmd = &cobra.Command{
	Use:     "calculator",
	Aliases: []string{"calc"},
	Short:   "Interactive cluster sizing calculator (elastic-calculator style TUI)",
	Long: "Estimate stored size, per-shard load, and a simple shard/replica placement model.\n" +
		"Optional --from-cluster fills fields from the active Elasticsearch connection (approximate).",
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		var seed *escalc.Inputs
		if fromCluster {
			client := elastic.NewClientWrapper(connection.GetClient())
			svc := services.NewClusterService(client)
			stats, err := util.ExecuteWithTimeout(func() (*models.ClusterStats, error) {
				return svc.GetClusterStats(context.Background())
			})
			if err != nil {
				fmt.Printf("Failed to load cluster stats: %v\n", err)
				return
			}
			seed = clusterStatsToInputs(stats)
		}

		if err := tui.RunCalculatorTUI(seed); err != nil {
			fmt.Printf("Error running calculator TUI: %v\n", err)
		}
	},
}

func clusterStatsToInputs(st *models.ClusterStats) *escalc.Inputs {
	replicas := 0
	if st.PrimaryShards > 0 {
		replicas = st.TotalShards/st.PrimaryShards - 1
	}
	if replicas < 0 {
		replicas = 0
	}

	gbSize := 0
	if st.TotalShards > 0 && st.UsedDiskBytes > 0 {
		primaryBytes := float64(st.UsedDiskBytes) * float64(st.PrimaryShards) / float64(st.TotalShards)
		gbSize = int(primaryBytes/(1000*1000*1000) + 0.5)
	}

	dedicated := st.MasterNodes
	if dedicated > st.TotalNodes {
		dedicated = st.TotalNodes
	}

	return &escalc.Inputs{
		Nodes:            st.TotalNodes,
		DedicatedMasters: dedicated,
		Shards:           st.PrimaryShards,
		ReplicasPerShard: replicas,
		GBSize:           gbSize,
		Documents:        st.TotalDocuments,
		ReadRPM:          3000,
		WriteRPM:         500,
		Clusters:         1,
	}
}

func init() {
	calculatorCmd.Flags().BoolVar(&fromCluster, "from-cluster", false, "Pre-fill inputs from the connected cluster (rough estimate for replicas and primary data size)")
	core.RootCmd.AddCommand(calculatorCmd)
}
