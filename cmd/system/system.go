package system

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

type SystemCommand struct {
	parentCmd   *cobra.Command
	commandType string
}

func NewSystemCommand(parentCmd *cobra.Command, commandType string) *cobra.Command {
	systemCmd := &cobra.Command{
		Use:   "system",
		Short: "Show system information (including system indices)",
		Run: func(cmd *cobra.Command, args []string) {
			runSystemCommand(commandType)
		},
	}

	parentCmd.AddCommand(systemCmd)
	return systemCmd
}

func runSystemCommand(commandType string) {
	client := elastic.NewClientWrapper(connection.GetClient())
	systemService := services.NewSystemService(client)

	switch commandType {
	case "index":
		runSystemIndex(context.Background(), systemService)
	case "shard":
		runSystemShard(context.Background(), systemService)
	default:
		fmt.Printf("System command not implemented for: %s\n", commandType)
	}
}

func runSystemIndex(ctx context.Context, systemService services.SystemService) {
	indices, err := util.ExecuteWithTimeout(func() ([]models.IndexInfo, error) {
		return systemService.GetSystemIndices(ctx)
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Index info fetch failed: %s\n", constants.MsgTimeoutGeneric)
		} else {
			fmt.Printf("Index info fetch failed: %v\n", err)
		}
		return
	}

	headers := []string{"Health", "Status", "Docs", "Size", "Primary", "Repl", "Alias", "Index"}
	rows := make([][]string, 0, len(indices))

	for _, idx := range indices {
		if !util.IsSystemIndex(idx.Name) {
			continue
		}

		alias := constants.DashString
		if idx.Alias != "" {
			alias = idx.Alias
		}

		health := idx.Health
		if len(health) > 6 {
			health = health[:6]
		}

		status := idx.Status
		if len(status) > 6 {
			status = status[:6]
		}

		docs := idx.DocsCount
		if len(docs) > 6 {
			docs = docs[:6]
		}

		size := idx.StoreSize
		if len(size) > 8 {
			size = size[:8]
		}

		primary := idx.Primary
		if len(primary) > 6 {
			primary = primary[:6]
		}

		replica := idx.Replica
		if len(replica) > 6 {
			replica = replica[:6]
		}

		if len(alias) > 6 {
			alias = alias[:6]
		}

		indexName := idx.Name
		if len(indexName) > 18 {
			indexName = indexName[:18]
		}

		row := []string{
			health,
			status,
			docs,
			size,
			primary,
			replica,
			alias,
			indexName,
		}
		rows = append(rows, row)
	}

	formatter := ui.NewGenericTableFormatter()
	fmt.Print(formatter.FormatTable(headers, rows))
}

func runSystemShard(ctx context.Context, systemService services.SystemService) {
	shards, err := util.ExecuteWithTimeout(func() ([]models.ShardInfo, error) {
		return systemService.GetSystemShards(ctx)
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Shard info fetch failed: %s\n", constants.MsgTimeoutGeneric)
		} else {
			fmt.Printf("Shard info fetch failed: %v\n", err)
		}
		return
	}

	headers := []string{"Shard", "Type", "State", "Size", "Node IP", "Index"}
	rows := make([][]string, 0, len(shards))

	for _, s := range shards {
		if !util.IsSystemIndex(s.Index) {
			continue
		}

		typeStr := util.ConvertShardName(s.Prirep)
		row := []string{
			s.Shard,
			typeStr,
			s.State,
			s.Store,
			s.IP,
			s.Index,
		}
		rows = append(rows, row)
	}

	formatter := ui.NewGenericTableFormatter()
	fmt.Print(formatter.FormatTable(headers, rows))
}
