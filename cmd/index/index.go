package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/mertbahardogan/escope/cmd/sort"
	"github.com/mertbahardogan/escope/cmd/system"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui"
	"github.com/mertbahardogan/escope/internal/util"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	indexName string
	topMode   bool
)

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

func runIndexList() {
	client := elastic.NewClientWrapper(connection.GetClient())
	indexService := services.NewIndexService(client)

	indices, err := util.ExecuteWithTimeout(func() ([]models.IndexInfo, error) {
		return indexService.GetAllIndexInfos(context.Background())
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Index info fetch failed: %s\n", constants.MsgTimeoutGeneric)
		} else {
			fmt.Printf("Index info fetch failed: %v\n", err)
		}
		return
	}

	var filteredIndices []models.IndexInfo
	for _, idx := range indices {
		if !util.IsSystemIndex(idx.Name) {
			filteredIndices = append(filteredIndices, idx)
		}
	}

	headers := []string{"Health", "Status", "Primary", "Replica", "Docs", "Size", "Alias", "Index"}
	rows := make([][]string, 0, len(filteredIndices))

	for _, index := range filteredIndices {
		docsCount := "-"
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
	fmt.Printf("Total: %d indices\n", len(filteredIndices))
}

func runIndexDetail(indexName string, topMode bool) {
	client := elastic.NewClientWrapper(connection.GetClient())
	indexService := services.NewIndexService(client)
	formatter := ui.NewIndexDetailFormatter()

	if topMode {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		checkCount := 0

		displayIndexDetail(indexService, formatter, indexName, &checkCount)

		for {
			select {
			case <-c:
				return
			case <-ticker.C:
				displayIndexDetail(indexService, formatter, indexName, &checkCount)
			}
		}
	} else {
		detailInfo, err := util.ExecuteWithTimeout(func() (*models.IndexDetailInfo, error) {
			return indexService.GetIndexDetailInfo(context.Background(), indexName)
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Index detail fetch failed: %s\n", constants.MsgTimeoutGeneric)
			} else {
				fmt.Printf("Index detail fetch failed: %v\n", err)
			}
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

func displayIndexDetail(indexService services.IndexService, formatter *ui.IndexDetailFormatter, indexName string, checkCount *int) {
	*checkCount++

	if *checkCount == constants.FirstCheckCount {
		fmt.Print(constants.ANSIClearScreen)
	} else {
		fmt.Printf(constants.ANSIMoveUpFormat, constants.IndexDetailLineCount)
	}

	detailInfo, err := util.ExecuteWithTimeout(func() (*models.IndexDetailInfo, error) {
		return indexService.GetIndexDetailInfo(context.Background(), indexName)
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Index detail fetch failed: %s\n", constants.MsgTimeoutGeneric)
		} else {
			fmt.Printf("Index detail fetch failed: %v\n", err)
		}
		return
	}

	formatterInfo := &ui.IndexDetailInfo{
		Name:         detailInfo.Name,
		SearchRate:   detailInfo.SearchRate,
		IndexRate:    detailInfo.IndexRate,
		AvgQueryTime: detailInfo.AvgQueryTime,
		AvgIndexTime: detailInfo.AvgIndexTime,
		CheckCount:   *checkCount,
	}

	fmt.Print(formatter.FormatIndexDetail(formatterInfo))
}

func init() {
	core.RootCmd.AddCommand(indexCmd)

	indexCmd.Flags().StringVarP(&indexName, "name", "n", "", "Show detailed information for specific index")
	indexCmd.Flags().BoolVarP(&topMode, "top", "t", false, "Continuously monitor index (like top command)")

	system.NewSystemCommand(indexCmd, "index")
	sort.NewSortCommand(indexCmd, "index")
}
