package ui

import (
	"fmt"
	"github.com/mertbahardogan/escope/internal/models"
)

type CheckFormatter struct{}

func NewCheckFormatter() *CheckFormatter {
	return &CheckFormatter{}
}

func (f *CheckFormatter) FormatCheckReport(
	clusterHealth *models.ClusterInfo,
	nodeHealths []models.CheckNodeHealth,
	shardHealth *models.ShardHealth,
	shardWarnings *models.ShardWarnings,
	indexHealths []models.IndexHealth,
	resourceUsage *models.ResourceUsage,
	performance *models.Performance,
	nodeBreakdown *models.NodeBreakdown,
	segmentWarnings *models.SegmentWarnings,
) string {
	title := "ESCOPE CLUSTER CHECK ANALYSIS"

	var sections []ReportSection

	criticalIssues := f.getCriticalIssues(clusterHealth, shardHealth, shardWarnings, resourceUsage)
	if len(criticalIssues) > 0 {
		sections = append(sections, ReportSection{
			Title: "CRITICAL ISSUES: " + fmt.Sprintf("%d", len(criticalIssues)),
			Items: criticalIssues,
		})
	}

	warningIssues := f.getWarningIssues(clusterHealth, shardHealth, shardWarnings, indexHealths, nodeHealths, resourceUsage, segmentWarnings)
	if len(warningIssues) > 0 {
		sections = append(sections, ReportSection{
			Title: "WARNING ISSUES: " + fmt.Sprintf("%d", len(warningIssues)),
			Items: warningIssues,
		})
	}

	// CPU Performance Metrics
	var cpuItems []string
	if resourceUsage != nil {
		cpuItems = append(cpuItems, "Average CPU Usage: "+fmt.Sprintf("%.1f%%", resourceUsage.CPUUsage))
		cpuItems = append(cpuItems, "CPU Usage Min: "+fmt.Sprintf("%.1f%%", resourceUsage.CPUUsageMin))
		cpuItems = append(cpuItems, "CPU Usage Max: "+fmt.Sprintf("%.1f%%", resourceUsage.CPUUsageMax))
	}
	if len(cpuItems) > 0 {
		sections = append(sections, ReportSection{
			Title: "CPU METRICS",
			Items: cpuItems,
		})
	}

	// Memory Performance Metrics
	var memoryItems []string
	if resourceUsage != nil {
		memoryItems = append(memoryItems, "Average Memory Usage: "+fmt.Sprintf("%.1f%%", resourceUsage.HeapUsage))
		memoryItems = append(memoryItems, "Memory Usage Min: "+fmt.Sprintf("%.1f%%", resourceUsage.HeapUsageMin))
		memoryItems = append(memoryItems, "Memory Usage Max: "+fmt.Sprintf("%.1f%%", resourceUsage.HeapUsageMax))
	}
	if len(memoryItems) > 0 {
		sections = append(sections, ReportSection{
			Title: "MEMORY METRICS",
			Items: memoryItems,
		})
	}

	// General Performance Metrics (Index/Query times)
	var performanceItems []string
	if performance != nil && performance.IndexTotal > 0 {
		avgIndexTime := float64(performance.IndexTimeInMillis) / float64(performance.IndexTotal)
		performanceItems = append(performanceItems, "Average Index Time: "+fmt.Sprintf("%.1fms", avgIndexTime))
	}
	if performance != nil && performance.QueryTotal > 0 {
		avgQueryTime := float64(performance.QueryTimeInMillis) / float64(performance.QueryTotal)
		performanceItems = append(performanceItems, "Average Query Time: "+fmt.Sprintf("%.1fms", avgQueryTime))
	}
	if len(performanceItems) > 0 {
		sections = append(sections, ReportSection{
			Title: "PERFORMANCE METRICS",
			Items: performanceItems,
		})
	}

	recommendations := f.getCategorizedRecommendations(clusterHealth, shardHealth, shardWarnings, indexHealths, nodeHealths, resourceUsage, segmentWarnings)

	if len(recommendations["SHARD"]) > 0 {
		sections = append(sections, ReportSection{
			Title: "SHARD:",
			Items: recommendations["SHARD"],
		})
	}

	if len(recommendations["INDEX"]) > 0 {
		sections = append(sections, ReportSection{
			Title: "INDEX",
			Items: recommendations["INDEX"],
		})
	}

	if len(recommendations["NODE"]) > 0 {
		sections = append(sections, ReportSection{
			Title: "NODE:",
			Items: recommendations["NODE"],
		})
	}

	if len(recommendations["GENERAL"]) > 0 {
		sections = append(sections, ReportSection{
			Title: "GENERAL",
			Items: recommendations["GENERAL"],
		})
	}

	formatter := NewGenericTableFormatter()
	return formatter.FormatReport(title, sections)
}

func (f *CheckFormatter) getCriticalIssues(clusterHealth *models.ClusterInfo, shardHealth *models.ShardHealth, shardWarnings *models.ShardWarnings, resourceUsage *models.ResourceUsage) []string {
	var issues []string

	if clusterHealth.UnassignedShards > 0 {
		issues = append(issues, fmt.Sprintf("Unassigned Shards: %d", clusterHealth.UnassignedShards))
	}

	if shardHealth.UnassignedShards > 0 {
		issues = append(issues, fmt.Sprintf("Unassigned Shards: %d", shardHealth.UnassignedShards))
	}

	if clusterHealth.DelayedUnassignedShards > 0 {
		issues = append(issues, fmt.Sprintf("Delayed Unassigned Shards: %d", clusterHealth.DelayedUnassignedShards))
	}

	if resourceUsage != nil && resourceUsage.DiskTotal > 0 {
		diskUsed := resourceUsage.DiskTotal - resourceUsage.DiskAvailable
		diskPercent := float64(diskUsed) * 100 / float64(resourceUsage.DiskTotal)
		if diskPercent > 85 {
			issues = append(issues, fmt.Sprintf("High Disk Usage: %.1f%%", diskPercent))
		}
	}

	if clusterHealth.TimedOut {
		issues = append(issues, "Cluster health check timed out")
	}

	if shardWarnings != nil {
		issues = append(issues, shardWarnings.CriticalIssues...)
	}

	return issues
}

func (f *CheckFormatter) getWarningIssues(clusterHealth *models.ClusterInfo, shardHealth *models.ShardHealth, shardWarnings *models.ShardWarnings, indexHealths []models.IndexHealth, nodeHealths []models.CheckNodeHealth, resourceUsage *models.ResourceUsage, segmentWarnings *models.SegmentWarnings) []string {
	var issues []string

	if clusterHealth.RelocatingShards > 0 {
		issues = append(issues, fmt.Sprintf("Relocating Shards: %d", clusterHealth.RelocatingShards))
	}

	if clusterHealth.InitializingShards > 0 {
		issues = append(issues, fmt.Sprintf("Initializing Shards: %d", clusterHealth.InitializingShards))
	}

	if shardHealth.RelocatingShards > 0 {
		issues = append(issues, fmt.Sprintf("Relocating Shards: %d", shardHealth.RelocatingShards))
	}

	if clusterHealth.NumberOfPendingTasks > 0 {
		issues = append(issues, fmt.Sprintf("Pending Tasks: %d", clusterHealth.NumberOfPendingTasks))
	}

	if clusterHealth.NumberOfInFlightFetch > 0 {
		issues = append(issues, fmt.Sprintf("In Flight Fetch: %d", clusterHealth.NumberOfInFlightFetch))
	}

	if clusterHealth.TaskMaxWaitingInQueueMillis > 1000 {
		issues = append(issues, fmt.Sprintf("Task Max Waiting: %dms", clusterHealth.TaskMaxWaitingInQueueMillis))
	}

	if clusterHealth.ActiveShardsPercentAsNumber < 100.0 {
		issues = append(issues, fmt.Sprintf("Active Shards Percent: %.1f%%", clusterHealth.ActiveShardsPercentAsNumber))
	}

	yellowIndices := 0
	for _, idx := range indexHealths {
		if idx.Health == "yellow" {
			yellowIndices++
		}
	}
	if yellowIndices > 0 {
		issues = append(issues, fmt.Sprintf("Yellow Indices: %d", yellowIndices))
	}

	for _, node := range nodeHealths {
		if node.HeapUsage > 75 {
			issues = append(issues, fmt.Sprintf("High Heap Usage: %s (%.1f%%)", node.Name, node.HeapUsage))
		}
	}

	if resourceUsage != nil && resourceUsage.HeapUsage > 75 {
		issues = append(issues, fmt.Sprintf("High Heap Usage: %.1f%%", resourceUsage.HeapUsage))
	}

	if shardWarnings != nil {
		issues = append(issues, shardWarnings.WarningIssues...)
	}

	if segmentWarnings != nil {
		if segmentWarnings.HighSegmentIndices > 0 {
			issues = append(issues, fmt.Sprintf("High Segment Count: %d indices with high segment counts (threshold varies by cluster size)", segmentWarnings.HighSegmentIndices))
		}
		if segmentWarnings.SmallSegmentIndices > 0 {
			issues = append(issues, fmt.Sprintf("Small Segments: %d indices with avg segment size <1MB", segmentWarnings.SmallSegmentIndices))
		}
	}

	return issues
}

func (f *CheckFormatter) getCategorizedRecommendations(clusterHealth *models.ClusterInfo, shardHealth *models.ShardHealth, shardWarnings *models.ShardWarnings, indexHealths []models.IndexHealth, nodeHealths []models.CheckNodeHealth, resourceUsage *models.ResourceUsage, segmentWarnings *models.SegmentWarnings) map[string][]string {
	recommendations := make(map[string][]string)
	recommendations["SHARD"] = []string{}
	recommendations["INDEX"] = []string{}
	recommendations["NODE"] = []string{}
	recommendations["GENERAL"] = []string{}

	if clusterHealth.UnassignedShards > 0 {
		recommendations["SHARD"] = append(recommendations["SHARD"], fmt.Sprintf("Investigate unassigned shards (%d) - check cluster allocation settings", clusterHealth.UnassignedShards))
	}
	if clusterHealth.RelocatingShards > 0 {
		recommendations["SHARD"] = append(recommendations["SHARD"], fmt.Sprintf("Monitor shard relocation progress (%d) - ensure completion", clusterHealth.RelocatingShards))
	}

	yellowIndices := 0
	for _, idx := range indexHealths {
		if idx.Health == "yellow" {
			yellowIndices++
		}
	}
	if yellowIndices > 0 {
		recommendations["INDEX"] = append(recommendations["INDEX"], fmt.Sprintf("Review yellow indices (%d) - check replica settings and node availability", yellowIndices))
	}
	if resourceUsage != nil && resourceUsage.DiskTotal > 0 {
		diskUsed := resourceUsage.DiskTotal - resourceUsage.DiskAvailable
		diskPercent := float64(diskUsed) * 100 / float64(resourceUsage.DiskTotal)
		if diskPercent > 85 {
			recommendations["INDEX"] = append(recommendations["INDEX"], "Consider index lifecycle management for disk usage optimization")
		}
	}

	if segmentWarnings != nil {
		if segmentWarnings.HighSegmentIndices > 0 {
			recommendations["INDEX"] = append(recommendations["INDEX"], fmt.Sprintf("Consider force merge for %d indices with high segment counts (threshold varies by cluster size)", segmentWarnings.HighSegmentIndices))
		}
		if segmentWarnings.SmallSegmentIndices > 0 {
			recommendations["INDEX"] = append(recommendations["INDEX"], fmt.Sprintf("Run force merge on %d indices with small segments (<1MB avg) to improve query performance", segmentWarnings.SmallSegmentIndices))
		}
		if segmentWarnings.LargeSegmentIndices > 0 {
			recommendations["INDEX"] = append(recommendations["INDEX"], fmt.Sprintf("%d indices have large segments (>1GB) - good for performance", segmentWarnings.LargeSegmentIndices))
		}
	}

	for _, node := range nodeHealths {
		if node.HeapUsage > 75 {
			recommendations["NODE"] = append(recommendations["NODE"], fmt.Sprintf("Consider heap tuning for %s - current usage %.1f%% (threshold: 75%%)", node.Name, node.HeapUsage))
		}
	}
	if resourceUsage != nil && resourceUsage.HeapUsage > 75 {
		recommendations["NODE"] = append(recommendations["NODE"], fmt.Sprintf("Monitor cluster heap usage - current %.1f%% (threshold: 75%%)", resourceUsage.HeapUsage))
	}

	recommendations["GENERAL"] = append(recommendations["GENERAL"], "Performance metrics are within acceptable ranges.")
	recommendations["GENERAL"] = append(recommendations["GENERAL"], "Consider implementing monitoring alerts for thresholds.")

	if shardWarnings != nil {
		for _, rec := range shardWarnings.Recommendations {
			recommendations["GENERAL"] = append(recommendations["GENERAL"], rec)
		}
	}

	return recommendations
}
