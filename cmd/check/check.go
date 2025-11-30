package check

import (
	"context"
	"fmt"
	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/interfaces"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui"
	"github.com/mertbahardogan/escope/internal/util"
	"github.com/spf13/cobra"
	"time"
)

var (
	duration string
	interval string
)

var checkCmd = &cobra.Command{
	Use:           "check",
	Short:         "Check cluster health metrics",
	Long:          `Check various aspects of your Elasticsearch cluster health and performance`,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		client := elastic.NewClientWrapper(connection.GetClient())
		checkService := services.NewCheckService(client)
		formatter := ui.NewCheckFormatter()

		if duration != "" {
			runContinuousCheck(context.Background(), client, checkService, formatter)
			return
		}

		runSingleCheck(context.Background(), checkService, formatter)
	},
}

func runSingleCheck(ctx context.Context, checkService services.CheckService, formatter *ui.CheckFormatter) {
	clusterHealth, err := util.ExecuteWithTimeout(func() (*models.ClusterInfo, error) {
		return checkService.GetClusterHealthCheck(ctx)
	})
	util.HandleServiceError(err, "Cluster health check")

	nodeHealths, err := util.ExecuteWithTimeout(func() ([]models.CheckNodeHealth, error) {
		return checkService.GetNodeHealthCheck(ctx)
	})
	util.HandleServiceError(err, "Node health check")

	shardHealth, err := util.ExecuteWithTimeout(func() (*models.ShardHealth, error) {
		return checkService.GetShardHealthCheck(ctx)
	})
	util.HandleServiceError(err, "Shard health check")

	shardWarnings, err := util.ExecuteWithTimeout(func() (*models.ShardWarnings, error) {
		return checkService.GetShardWarningsCheck(ctx)
	})
	util.HandleServiceError(err, "Shard warnings check")

	indexHealths, err := util.ExecuteWithTimeout(func() ([]models.IndexHealth, error) {
		return checkService.GetIndexHealthCheck(ctx)
	})
	util.HandleServiceError(err, "Index health check")

	resourceUsage, err := util.ExecuteWithTimeout(func() (*models.ResourceUsage, error) {
		return checkService.GetResourceUsageCheck(ctx)
	})
	util.HandleServiceError(err, "Resource usage check")

	performance, err := util.ExecuteWithTimeout(func() (*models.Performance, error) {
		return checkService.GetPerformanceCheck(ctx)
	})
	util.HandleServiceError(err, "Performance check")

	nodeBreakdown, err := util.ExecuteWithTimeout(func() (*models.NodeBreakdown, error) {
		return checkService.GetNodeBreakdown(ctx)
	})
	util.HandleServiceError(err, "Node breakdown check")

	segmentWarnings, err := util.ExecuteWithTimeout(func() (*models.SegmentWarnings, error) {
		return checkService.GetSegmentWarningsCheck(ctx)
	})
	util.HandleServiceError(err, "Segment warnings check")

	scaleWarnings, err := util.ExecuteWithTimeout(func() (*models.ScaleWarnings, error) {
		return checkService.GetScaleWarningsCheck(ctx)
	})
	util.HandleServiceError(err, "Scale warnings check")

	output := formatter.FormatCheckReport(
		clusterHealth,
		nodeHealths,
		shardHealth,
		shardWarnings,
		indexHealths,
		resourceUsage,
		performance,
		nodeBreakdown,
		segmentWarnings,
		scaleWarnings,
	)
	fmt.Print(output)
}

func runContinuousCheck(ctx context.Context, client interfaces.ElasticClient, checkService services.CheckService, formatter *ui.CheckFormatter) {
	durationTime, err := time.ParseDuration(duration)
	if err != nil {
		fmt.Printf("Invalid duration format: %v\n", err)
		fmt.Println("Valid formats: 1m, 5m, 1h, etc.")
		return
	}

	intervalTime := time.Duration(constants.DefaultInterval) * time.Second
	if interval != "" {
		intervalTime, err = time.ParseDuration(interval)
		if err != nil {
			fmt.Printf("Invalid interval format: %v\n", err)
			fmt.Println("Valid formats: 5s, 10s, 1m, etc.")
			return
		}
	}

	monitoringService := services.NewMonitoringService(client)

	result, err := monitoringService.MonitorCluster(ctx, durationTime, intervalTime)
	if err != nil {
		fmt.Printf("Monitoring failed: %v\n", err)
		return
	}

	if result.SampleCount > 0 {
		runSingleCheck(ctx, checkService, formatter)
	} else {
		fmt.Println("No samples collected during monitoring period.")
	}
}

func init() {
	checkCmd.Flags().StringVarP(&duration, "duration", "d", "", "Duration for continuous monitoring (e.g., 1m, 5m, 1h)")
	checkCmd.Flags().StringVarP(&interval, "interval", "i", "",
		"Sampling interval for continuous monitoring (e.g., 5s, 10s, 1m, default: 2s)")

	core.RootCmd.AddCommand(checkCmd)
}
