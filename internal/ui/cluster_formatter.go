package ui

import (
	"fmt"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/util"
	"strings"
)

type ClusterFormatter struct {
	genericFormatter *GenericTableFormatter
}

func NewClusterFormatter() *ClusterFormatter {
	return &ClusterFormatter{
		genericFormatter: NewGenericTableFormatter(),
	}
}

func (f *ClusterFormatter) FormatClusterStats(stats *models.ClusterStats) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("\nCluster: %s (%s)\n\n", stats.ClusterName, stats.Status))

	usedDisk := stats.TotalDiskBytes - stats.AvailableDiskBytes
	resourceHeaders := []string{"Resource", "Used", "Total", "Usage %"}
	resourceRows := [][]string{
		{"Disk Storage", util.FormatBytes(usedDisk), util.FormatBytes(stats.TotalDiskBytes), fmt.Sprintf("%.1f%%", stats.DiskUsagePercent)},
		{"Heap Memory", util.FormatBytes(stats.UsedHeapBytes), util.FormatBytes(stats.TotalHeapBytes), fmt.Sprintf("%.1f%%", stats.HeapUsagePercent)},
		{"System Memory", util.FormatBytes(stats.UsedMemoryBytes), util.FormatBytes(stats.TotalMemoryBytes), fmt.Sprintf("%.1f%%", stats.MemoryUsagePercent)},
	}
	output.WriteString(f.genericFormatter.FormatTable(resourceHeaders, resourceRows))
	output.WriteString("\n")

	clusterHeaders := []string{"Metric", "Value"}
	clusterRows := [][]string{
		{"Nodes", fmt.Sprintf("%d (%s)", stats.TotalNodes, stats.GetNodeBreakdown())},
		{"Indices", fmt.Sprintf("%d", stats.TotalIndices)},
		{"Documents", util.FormatDocsCount(stats.TotalDocuments)},
		{"Primary Shards", fmt.Sprintf("%d", stats.PrimaryShards)},
		{"Total Shards", fmt.Sprintf("%d", stats.TotalShards)},
		{"Avg Shard Size", fmt.Sprintf("%.2f GB", stats.AvgShardSizeGB)},
	}
	output.WriteString(f.genericFormatter.FormatTable(clusterHeaders, clusterRows))
	output.WriteString("\n")

	jvmVersions := "N/A"
	if len(stats.JVMVersions) > 0 {
		jvmVersions = strings.Join(stats.JVMVersions, ", ")
	}
	systemHeaders := []string{"System Info", "Value"}
	systemRows := [][]string{
		{"ES Version", stats.ESVersion},
		{"JVM Versions", jvmVersions},
	}
	output.WriteString(f.genericFormatter.FormatTable(systemHeaders, systemRows))

	return output.String()
}
