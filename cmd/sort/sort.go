package sort

import (
	"context"
	"errors"
	"fmt"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/ui"
	"github.com/mertbahardogan/escope/internal/util"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type CommandConfig struct {
	Headers     []string
	FieldMap    map[string]string
	FetchFunc   func(string, string) (interface{}, error)
	DisplayFunc func(interface{})
}

func NewSortCommand(parentCmd *cobra.Command, commandType string) *cobra.Command {
	sortCmd := &cobra.Command{
		Use:   "sort [field] [direction]",
		Short: "Sort data by specified field",
		Long:  generateSortHelp(commandType),
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			field := args[0]
			direction := "desc"
			if len(args) > 1 {
				direction = args[1]
			}
			runSortCommand(commandType, field, direction)
		},
	}

	parentCmd.AddCommand(sortCmd)
	return sortCmd
}

func getCommandConfigs() map[string]CommandConfig {
	return map[string]CommandConfig{
		"shard": {
			Headers: []string{"Index", "Shard", "Pri/Rep", "State", "Node IP", "Node Name", "Store"},
			FieldMap: map[string]string{
				"index": "index",
				"shard": "shard",
				"type":  "prirep",
				"state": "state",
				"size":  "store",
			},
			FetchFunc:   fetchShardDataWithSort,
			DisplayFunc: displayShardTable,
		},
		"index": {
			Headers: []string{"Health", "Status", "Primary", "Replica", "Docs Count", "Store Size", "Alias", "Index"},
			FieldMap: map[string]string{
				"health":  "health",
				"status":  "status",
				"primary": "pri",
				"replica": "rep",
				"docs":    "docs.count",
				"size":    "store.size",
				"alias":   "alias",
				"index":   "index",
			},
			FetchFunc:   fetchIndexDataWithSort,
			DisplayFunc: displayIndexTable,
		},
	}
}

func generateSortHelp(commandType string) string {
	configs := getCommandConfigs()
	config, exists := configs[commandType]
	if !exists {
		return fmt.Sprintf("Sort command not implemented for: %s", commandType)
	}

	headers := config.Headers

	var help strings.Builder
	help.WriteString(fmt.Sprintf("Sort %s data by a specific field.\n\n", commandType))
	help.WriteString("Usage:\n")
	help.WriteString(fmt.Sprintf("  escope %s sort <field> [direction]\n\n", commandType))
	help.WriteString("Arguments:\n")
	help.WriteString("  field     Field to sort by (see available fields below)\n")
	help.WriteString("  direction Sorting direction: asc (ascending) or desc (descending) [default: desc]\n\n")
	help.WriteString("Available fields:\n")

	for _, header := range headers {
		help.WriteString(fmt.Sprintf("  • %s\n", header))
	}

	help.WriteString("\nExamples:\n")
	help.WriteString(fmt.Sprintf("  escope %s sort size        # Sort by size (descending)\n", commandType))
	help.WriteString(fmt.Sprintf("  escope %s sort size desc   # Sort by size (descending)\n", commandType))
	help.WriteString(fmt.Sprintf("  escope %s sort size asc    # Sort by size (ascending)\n", commandType))

	return help.String()
}

func runSortCommand(commandType, field, direction string) {
	configs := getCommandConfigs()
	config, exists := configs[commandType]
	if !exists {
		fmt.Printf("Data fetching not implemented for: %s\n", commandType)
		return
	}

	// Get ES field name
	esField, exists := config.FieldMap[strings.ToLower(field)]
	if !exists {
		fmt.Printf("Error: field '%s' not found\n", field)
		fmt.Println("Available fields:")
		for f := range config.FieldMap {
			fmt.Printf("  • %s\n", f)
		}
		return
	}

	data, err := config.FetchFunc(esField, direction)
	if err != nil {
		fmt.Printf("Data fetch failed: %v\n", err)
		return
	}

	fmt.Printf("\n%s sorted by %s (%s):\n",
		strings.Title(commandType), field, direction)

	config.DisplayFunc(data)
}

func fetchDataWithSort(sortBy, direction, dataType string) (interface{}, error) {
	client := elastic.NewClientWrapper(connection.GetClient())

	var data []map[string]interface{}
	var err error

	switch dataType {
	case "shard":
		data, err = util.ExecuteWithTimeout(func() ([]map[string]interface{}, error) {
			return client.GetShardsWithSort(context.Background(), sortBy, direction)
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return nil, fmt.Errorf("shard info fetch failed: %s", constants.MsgTimeoutGeneric)
			} else {
				return nil, fmt.Errorf("shard info fetch failed: %v", err)
			}
		}
		return processData(data, dataType)
	case "index":
		data, err = util.ExecuteWithTimeout(func() ([]map[string]interface{}, error) {
			return client.GetIndicesWithSort(context.Background(), sortBy, direction)
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return nil, fmt.Errorf("index info fetch failed: %s", constants.MsgTimeoutGeneric)
			} else {
				return nil, fmt.Errorf("index info fetch failed: %v", err)
			}
		}
		return processData(data, dataType)
	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}
}

func processData(data []map[string]interface{}, dataType string) (interface{}, error) {
	switch dataType {
	case "shard":
		return processShardData(data), nil
	case "index":
		return processIndexData(data), nil
	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}
}

func processShardData(shardData []map[string]interface{}) []models.ShardInfo {
	var filteredShards []models.ShardInfo
	for _, s := range shardData {
		indexName := util.GetStringField(s, "index")
		if !util.IsSystemIndex(indexName) {
			shardInfo := models.ShardInfo{
				Index:  indexName,
				Shard:  util.GetStringField(s, "shard"),
				Prirep: util.GetStringField(s, "prirep"),
				State:  util.GetStringField(s, "state"),
				IP:     util.GetStringField(s, "ip"),
				Node:   util.GetStringField(s, "node"),
				Store:  util.GetStringField(s, "store"),
			}
			filteredShards = append(filteredShards, shardInfo)
		}
	}
	return filteredShards
}

func processIndexData(indexData []map[string]interface{}) []models.IndexInfo {
	var filteredIndices []models.IndexInfo
	for _, idx := range indexData {
		indexName := util.GetStringField(idx, "index")
		if !util.IsSystemIndex(indexName) {
			indexInfo := models.IndexInfo{
				Name:      indexName,
				Health:    util.GetStringField(idx, "health"),
				Status:    util.GetStringField(idx, "status"),
				Primary:   util.GetStringField(idx, "pri"),
				Replica:   util.GetStringField(idx, "rep"),
				DocsCount: util.GetStringField(idx, "docs.count"),
				StoreSize: util.GetStringField(idx, "store.size"),
				Alias:     util.GetStringField(idx, "alias"),
			}
			filteredIndices = append(filteredIndices, indexInfo)
		}
	}
	return filteredIndices
}

func fetchShardDataWithSort(sortBy, direction string) (interface{}, error) {
	return fetchDataWithSort(sortBy, direction, "shard")
}

func fetchIndexDataWithSort(sortBy, direction string) (interface{}, error) {
	return fetchDataWithSort(sortBy, direction, "index")
}

func displayShardTable(data interface{}) {
	shards, ok := data.([]models.ShardInfo)
	if !ok {
		fmt.Println("Error: invalid data type for shard table display")
		return
	}

	headers := []string{"Shard", "Type", "State", "Size", "Node IP", "Index"}
	rows := make([][]string, 0, len(shards))

	for _, shard := range shards {
		shardType := "Primary"
		if shard.Prirep == "r" {
			shardType = "Replica"
		}

		row := []string{
			shard.Shard,
			shardType,
			shard.State,
			shard.Store,
			shard.IP,
			shard.Index,
		}
		rows = append(rows, row)
	}

	formatter := ui.NewGenericTableFormatter()
	fmt.Print(formatter.FormatTable(headers, rows))
	fmt.Printf("Total: %d shards\n", len(shards))
}

func displayIndexTable(data interface{}) {
	indices, ok := data.([]models.IndexInfo)
	if !ok {
		fmt.Println("Error: invalid data type for index table formatting")
		return
	}

	headers := []string{"Health", "Status", "Primary", "Replica", "Docs", "Size", "Alias", "Index"}
	rows := make([][]string, 0, len(indices))

	for _, index := range indices {
		docsCount := constants.DashString
		if index.DocsCount != "" {
			if count, err := strconv.ParseInt(index.DocsCount, 10, 64); err == nil {
				docsCount = util.FormatDocsCount(count)
			} else {
				docsCount = index.DocsCount
			}
		}

		row := []string{
			index.Health,
			index.Status,
			index.Primary,
			index.Replica,
			docsCount,
			index.StoreSize,
			index.Alias,
			index.Name,
		}
		rows = append(rows, row)
	}

	formatter := ui.NewGenericTableFormatter()
	fmt.Print(formatter.FormatTable(headers, rows))
	fmt.Printf("Total: %d indices\n", len(indices))
}
