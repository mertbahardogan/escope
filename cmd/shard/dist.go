package shard

import (
	"context"
	"errors"
	"escope/internal/connection"
	"escope/internal/constants"
	"escope/internal/elastic"
	"escope/internal/models"
	"escope/internal/services"
	"escope/internal/ui"
	"escope/internal/util"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var distributionCmd = &cobra.Command{
	Use:                "dist",
	Short:              "Show shard distribution analysis across nodes",
	SilenceErrors:      true,
	DisableSuggestions: true,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		client := elastic.NewClientWrapper(connection.GetClient())
		shardService := services.NewShardService(client)
		nodeService := services.NewNodeService(client)

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

		nodesInfo, err := util.ExecuteWithTimeout(func() ([]models.NodeInfo, error) {
			return nodeService.GetNodesInfo(ctx)
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Node info fetch failed: %s\n", constants.MsgTimeoutGeneric)
			} else {
				fmt.Printf("Node info fetch failed: %v\n", err)
			}
			return
		}
		ipToName := make(map[string]string)
		for _, node := range nodesInfo {
			ipToName[node.IP] = node.Name
		}
		nodeShards := make(map[string][]models.ShardInfo)
		for _, s := range shards {
			if s.State != "STARTED" || s.IP == "-" || util.IsSystemIndex(s.Index) {
				continue
			}
			nodeShards[s.IP] = append(nodeShards[s.IP], s)
		}

		var nodeIPs []string
		for ip := range nodeShards {
			nodeIPs = append(nodeIPs, ip)
		}
		util.SortStrings(nodeIPs, util.Ascending)

		headers := []string{"Shards", "Index", "IP", "Node Name"}
		formatter := ui.NewGenericTableFormatter()

		var allShards []models.ShardInfo
		for _, nodeIP := range nodeIPs {
			shardsForNode := nodeShards[nodeIP]
			util.SortShardsByTypeAndIndex(shardsForNode)
			allShards = append(allShards, shardsForNode...)
		}

		var tableRows [][]string
		var lastIP string
		for _, s := range allShards {
			if lastIP != "" && lastIP != s.IP {
				separatorRow := []string{"---", "---", "---", "---"}
				tableRows = append(tableRows, separatorRow)
			}
			lastIP = s.IP

			shardType := "Primary-" + s.Shard
			if s.Prirep == "r" {
				shardType = "Replica-" + s.Shard
			}

			indexName := s.Index
			if strings.HasPrefix(indexName, "search_") {
				indexName = strings.TrimPrefix(indexName, "search_")
			}

			nodeName := ipToName[s.IP]
			if nodeName == "" {
				nodeName = "unknown"
			}
			row := []string{
				shardType,
				indexName,
				s.IP,
				nodeName,
			}
			tableRows = append(tableRows, row)
		}
		fmt.Print(formatter.FormatTable(headers, tableRows))
	},
}

func init() {
	shardCmd.AddCommand(distributionCmd)
}
