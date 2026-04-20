package calculator

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mertbahardogan/escope/cmd/core"
	escalc "github.com/mertbahardogan/escope/internal/calculator"
	"github.com/mertbahardogan/escope/internal/calculatorsession"
	"github.com/mertbahardogan/escope/internal/config"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/indexsession"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui/tui"
	"github.com/mertbahardogan/escope/internal/util"
	"github.com/spf13/cobra"
)

var fromCluster bool
var fromSnapshot bool
var clearSession bool

var calculatorCmd = &cobra.Command{
	Use:     "calculator",
	Aliases: []string{"calc"},
	Short:   "Interactive cluster sizing calculator; save with ctrl+s",
	Long: "Interactive sizing calculator: edit numeric inputs, summary tables, shard/replica placement,\n" +
		"per-data-node RAM chart (JVM vs OS cache, heap detail), and utilization hints.\n" +
		"Throughput is in rps (requests/s); the first field is data nodes (total = data + dedicated masters).\n" +
		"Page-cache vs data coverage is derived from RAM/disk per node and data on each node.\n\n" +
		"Persistence: field values, focus, and scroll are stored under `sessions` for the active host\n" +
		"(same host key as `escope index use` default index).\n" +
		"Nothing is written automatically: press ctrl+s to save. With no flags, the last saved snapshot for this host is restored when present; otherwise built-in default numbers are used.\n" +
		"--from-cluster pre-fills from the live cluster (and refines with 'escope index use' default index when set).\n" +
		"--snapshot loads only from a stored snapshot and exits with an error if none exists for this host.\n" +
		"--clear (or the single argument clear) removes the calculator snapshot for the current host (or the host entry if no default index).\n\n" +
		"Requires a configured Elasticsearch host like other escope commands, except --help and --clear.",
	Example: `  escope calculator --help

  escope calculator

  escope calculator --from-cluster
  escope calculator from-cluster

  escope calculator --snapshot

  escope calculator --clear
  escope calculator clear`,
	Args:          cobra.MaximumNArgs(1),
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		fc, fs, clr := parseCalculatorRunFlags(cmd, args)
		if clr && fc {
			fmt.Println("Error: do not combine --clear with --from-cluster")
			return
		}
		if clr && fs {
			fmt.Println("Error: do not combine --clear with --snapshot")
			return
		}
		if fc && fs {
			fmt.Println("Error: do not combine --from-cluster with --snapshot")
			return
		}
		if clr {
			if err := calculatorsession.Clear(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("Cleared saved snapshot.")
			return
		}

		var seed *escalc.Inputs
		var liveStats *models.ClusterStats
		if fc {
			in, st, err := loadSeedFromLiveCluster()
			if err != nil {
				fmt.Printf("Failed to load cluster stats: %v\n", err)
				return
			}
			seed, liveStats = in, st
		}

		header := buildCalculatorHeader(liveStats)
		if err := tui.RunCalculatorTUI(seed, fs, header); err != nil {
			fmt.Printf("Error running calculator: %v\n", err)
		}
	},
}

func parseCalculatorRunFlags(cmd *cobra.Command, args []string) (fromCluster, fromSnapshot, clearSession bool) {
	fromCluster, _ = cmd.Flags().GetBool("from-cluster")
	fromSnapshot, _ = cmd.Flags().GetBool("snapshot")
	clearSession, _ = cmd.Flags().GetBool("clear")
	if len(args) == 1 {
		switch strings.ToLower(strings.TrimSpace(args[0])) {
		case "clear":
			clearSession = true
		case "from-cluster":
			fromCluster = true
		}
	}
	return fromCluster, fromSnapshot, clearSession
}

func loadSeedFromLiveCluster() (*escalc.Inputs, *models.ClusterStats, error) {
	es := connection.GetClient()
	if es == nil {
		return nil, nil, fmt.Errorf("no elasticsearch client")
	}
	client := elastic.NewClientWrapper(es)
	svc := services.NewClusterService(client)
	stats, err := util.ExecuteWithTimeout(func() (*models.ClusterStats, error) {
		return svc.GetClusterStats(context.Background())
	})
	if err != nil {
		return nil, nil, err
	}
	seed := clusterStatsToInputs(stats)
	if idx, ok := indexsession.ReadSelectedIndex(); ok {
		idxSvc := services.NewIndexService(client)
		if err := idxSvc.MergeCalculatorInputsFromIndex(context.Background(), idx, seed); err != nil {
			fmt.Fprintf(os.Stderr, "calculator: could not apply default index %q (using cluster-wide stats): %v\n", idx, err)
		}
	}
	return seed, stats, nil
}

func buildCalculatorHeader(st *models.ClusterStats) *tui.CalculatorHeader {
	alias, err := config.GetActiveHost()
	if err != nil || alias == "" {
		return nil
	}
	status := ""
	if st != nil && st.Status != "" {
		status = st.Status
	} else {
		client := elastic.NewClientWrapper(connection.GetClient())
		svc := services.NewClusterService(client)
		h, err := util.ExecuteWithTimeout(func() (*models.ClusterInfo, error) {
			return svc.GetClusterHealth(context.Background())
		})
		if err == nil && h != nil {
			status = h.Status
		}
	}
	return &tui.CalculatorHeader{HostAlias: alias, Status: status}
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

	dataNodes := st.TotalNodes - dedicated
	if dataNodes < 0 {
		dataNodes = 0
	}

	nodes := st.TotalNodes
	if nodes <= 0 {
		nodes = 1
	}

	ramPer := 64.0
	if st.TotalMemoryBytes > 0 {
		ramPer = float64(st.TotalMemoryBytes) / float64(nodes) / 1e9
	}
	diskPer := 2000.0
	if st.TotalDiskBytes > 0 {
		diskPer = float64(st.TotalDiskBytes) / float64(nodes) / 1e9
	}

	return &escalc.Inputs{
		DataNodes:          dataNodes,
		DedicatedMasters:   dedicated,
		Shards:             st.PrimaryShards,
		ReplicasPerShard:   replicas,
		GBSize:             gbSize,
		Documents:          st.TotalDocuments,
		ReadRPS:            3000.0 / 60.0,
		WriteRPS:           500.0 / 60.0,
		RAMGiBPerDataNode:  ramPer,
		DiskGiBPerDataNode: diskPer,
	}
}

func init() {
	calculatorCmd.Flags().BoolVar(&fromCluster, "from-cluster", false,
		"Pre-fill from live cluster stats; optional default index from 'escope index use' refines shards, size, and docs")
	calculatorCmd.Flags().BoolVar(&fromSnapshot, "snapshot", false,
		"Load inputs only from stored snapshot for this host; fails if none exists")
	calculatorCmd.Flags().BoolVar(&clearSession, "clear", false,
		"Remove calculator snapshot for the current host; does not start the interactive session")
	core.RootCmd.AddCommand(calculatorCmd)
}
