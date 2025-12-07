package index

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/mertbahardogan/escope/cmd/sort"
	"github.com/mertbahardogan/escope/cmd/system"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui"
	"github.com/mertbahardogan/escope/internal/ui/tui"
	"github.com/mertbahardogan/escope/internal/util"
	"github.com/spf13/cobra"
)

var (
	indexName string
	topMode   bool
)

type IndexMetrics struct {
	SearchRate   string
	IndexRate    string
	AvgQueryTime string
	AvgIndexTime string
}

var indexCmd = &cobra.Command{
	Use:                "index",
	Short:              "Show index summary information",
	SilenceErrors:      true,
	DisableSuggestions: true,
	Args:               cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		indexName, _ := cmd.Flags().GetString("name")
		topMode, _ := cmd.Flags().GetBool("top")

		if indexName != "" {
			runIndexDetail(indexName, topMode)
			return
		}

		runIndexList()
	},
}

func getIndexMetricsWithSnapshot(indexService services.IndexService, indexName string) (*models.IndexDetailInfo, error) {
	_, err := util.ExecuteWithTimeout(func() (*models.IndexDetailInfo, error) {
		return indexService.GetIndexDetailInfo(context.Background(), indexName)
	})
	if err != nil {
		return nil, err
	}

	time.Sleep(2 * time.Second)

	return util.ExecuteWithTimeout(func() (*models.IndexDetailInfo, error) {
		return indexService.GetIndexDetailInfo(context.Background(), indexName)
	})
}

func runIndexList() {
	client := elastic.NewClientWrapper(connection.GetClient())
	indexService := services.NewIndexService(client)

	indices, err := util.ExecuteWithTimeout(func() ([]models.IndexInfo, error) {
		return indexService.GetAllIndexInfos(context.Background())
	})
	if util.HandleServiceErrorWithReturn(err, "Index info fetch") {
		return
	}

	var filteredIndices []models.IndexInfo
	for _, idx := range indices {
		if !util.IsSystemIndex(idx.Name) {
			filteredIndices = append(filteredIndices, idx)
		}
	}

	metricsMap := make(map[string]IndexMetrics)
	for _, idx := range filteredIndices {
		detail, err := getIndexMetricsWithSnapshot(indexService, idx.Name)
		if err == nil {
			metricsMap[idx.Name] = IndexMetrics{
				SearchRate:   detail.SearchRate,
				IndexRate:    detail.IndexRate,
				AvgQueryTime: detail.AvgQueryTime,
				AvgIndexTime: detail.AvgIndexTime,
			}
		} else {
			metricsMap[idx.Name] = IndexMetrics{
				SearchRate:   constants.DashString,
				IndexRate:    constants.DashString,
				AvgQueryTime: constants.DashString,
				AvgIndexTime: constants.DashString,
			}
		}
	}

	headers := []string{"Health", "Status", "Primary", "Replica", "Docs", "Size", "Alias", "Index", "Search Rate", "Index Rate", "Query Time", "Index Time"}
	rows := make([][]string, 0, len(filteredIndices))

	for _, index := range filteredIndices {
		docsCount := constants.DashString
		if index.DocsCount != "" {
			if count, err := strconv.ParseInt(index.DocsCount, 10, 64); err == nil {
				docsCount = util.FormatDocsCount(count)
			} else {
				docsCount = index.DocsCount
			}
		}

		metrics := metricsMap[index.Name]
		row := []string{
			index.Health,
			index.Status,
			index.Primary,
			index.Replica,
			docsCount,
			index.StoreSize,
			index.Alias,
			index.Name,
			metrics.SearchRate,
			metrics.IndexRate,
			metrics.AvgQueryTime,
			metrics.AvgIndexTime,
		}
		rows = append(rows, row)
	}

	formatter := ui.NewGenericTableFormatter()
	fmt.Print(formatter.FormatTable(headers, rows))
	fmt.Printf("Total: %d indices\n", len(filteredIndices))
}

func runIndexDetail(indexName string, topMode bool) {
	client := elastic.NewClientWrapper(connection.GetClient())
	indexService := services.NewIndexService(client)

	if topMode {
		if err := tui.RunIndexTUI(indexService, indexName); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
		}
	} else {
		formatter := ui.NewIndexDetailFormatter()
		detailInfo, err := getIndexMetricsWithSnapshot(indexService, indexName)
		if util.HandleServiceErrorWithReturn(err, "Index detail fetch") {
			return
		}

		formatterInfo := &ui.IndexDetailInfo{
			Name:         detailInfo.Name,
			SearchRate:   detailInfo.SearchRate,
			IndexRate:    detailInfo.IndexRate,
			AvgQueryTime: detailInfo.AvgQueryTime,
			AvgIndexTime: detailInfo.AvgIndexTime,
			CheckCount:   0,
		}

		fmt.Print(formatter.FormatIndexDetail(formatterInfo))
	}
}

func init() {
	core.RootCmd.AddCommand(indexCmd)

	indexCmd.Flags().StringVarP(&indexName, "name", "n", "", "Show detailed information for specific index")
	indexCmd.Flags().BoolVarP(&topMode, "top", "t", false, "Continuously monitor index (like top command)")

	system.NewSystemCommand(indexCmd, "index")
	sort.NewSortCommand(indexCmd, "index")
}
