package node

import (
	"context"
	"errors"
	"fmt"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui"
	"github.com/mertbahardogan/escope/internal/util"
	"github.com/spf13/cobra"
)

var nodeDistCmd = &cobra.Command{
	Use:           "dist",
	Short:         "Show node distribution analysis",
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		client := elastic.NewClientWrapper(connection.GetClient())
		nodeService := services.NewNodeService(client)

		shardService := services.NewShardService(client)
		shards, err := util.ExecuteWithTimeout(func() ([]models.ShardInfo, error) {
			return shardService.GetAllShardInfos(ctx)
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Shard info fetch failed: %s\n", constants.MsgTimeoutGeneric)
			} else {
				fmt.Printf("Shard info fetch failed: %v\n", err)
			}
			return
		}

		nodeNameToIP := make(map[string]string)
		for _, shard := range shards {
			if shard.Node != "" && shard.IP != "" && shard.IP != constants.DashString {
				nodeNameToIP[shard.Node] = shard.IP
			}
		}

		nodeStats, err := util.ExecuteWithTimeout(func() ([]models.NodeStat, error) {
			return nodeService.GetNodeStats(ctx)
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Node stats check failed: %s\n", constants.MsgTimeoutGeneric)
			} else {
				fmt.Printf("Node stats check failed: %v\n", err)
			}
			return
		}

		balance, err := util.ExecuteWithTimeout(func() (*models.BalanceAnalysis, error) {
			return nodeService.AnalyzeNodeBalance(ctx)
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Node balance analysis failed: %s\n", constants.MsgTimeoutGeneric)
			} else {
				fmt.Printf("Node balance analysis failed: %v\n", err)
			}
			return
		}
		if len(nodeStats) > 0 {
			headers := []string{"Primary", "Replica", "Total", "Indices", "IP", "Name"}
			rows := make([][]string, 0, len(nodeStats))

			for _, stat := range nodeStats {
				nodeName := stat.NodeIP
				nodeIP := nodeNameToIP[nodeName]

				row := []string{
					fmt.Sprintf("%d", stat.PrimaryShards),
					fmt.Sprintf("%d", stat.ReplicaShards),
					fmt.Sprintf("%d", stat.TotalShards),
					fmt.Sprintf("%d", stat.IndexCount),
					nodeIP,
					nodeName,
				}
				rows = append(rows, row)
			}

			formatter := ui.NewGenericTableFormatter()
			fmt.Print(formatter.FormatTable(headers, rows))
			fmt.Println()

			fmt.Printf("Balance Analysis:\n")

			mostLoadedNodeName := balance.MostLoadedNode
			leastLoadedNodeName := balance.LeastLoadedNode
			mostLoadedNodeIP := nodeNameToIP[mostLoadedNodeName]
			leastLoadedNodeIP := nodeNameToIP[leastLoadedNodeName]

			fmt.Printf("Most loaded node: %s - %s (%d shards)\n", mostLoadedNodeName, mostLoadedNodeIP, balance.MaxShards)
			fmt.Printf("Least loaded node: %s - %s (%d shards)\n", leastLoadedNodeName, leastLoadedNodeIP, balance.MinShards)
			fmt.Printf("Balance ratio: %.1f%%\n", balance.BalanceRatio*100)
			fmt.Printf("Status: %s\n", balance.Recommendation)
		} else {
			fmt.Printf("No node statistics available\n")
		}
	},
}
