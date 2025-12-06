package ui

import (
	"fmt"
	"strings"

	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/ui/components"
	"github.com/mertbahardogan/escope/internal/util"
)

type ClusterFormatter struct {
	table       *components.Table
	panel       *components.Panel
	progressBar *components.ProgressBar
}

func NewClusterFormatter() *ClusterFormatter {
	return &ClusterFormatter{
		table:       components.NewTable(),
		panel:       components.NewPanel("RESOURCES"),
		progressBar: components.NewProgressBar(),
	}
}

func (f *ClusterFormatter) FormatClusterStats(stats *models.ClusterStats) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("\nCluster: %s (%s)\n\n", stats.ClusterName, stats.Status))

	output.WriteString(f.formatResourcesPanel(stats))
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
	output.WriteString(f.table.Render(clusterHeaders, clusterRows))
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
	output.WriteString(f.table.Render(systemHeaders, systemRows))

	return output.String()
}

func (f *ClusterFormatter) formatResourcesPanel(stats *models.ClusterStats) string {
	usedDisk := stats.TotalDiskBytes - stats.AvailableDiskBytes

	lines := []string{
		f.formatResourceLine("DISK", stats.DiskUsagePercent, usedDisk, stats.TotalDiskBytes),
		f.formatResourceLine("HEAP", stats.HeapUsagePercent, stats.UsedHeapBytes, stats.TotalHeapBytes),
		f.formatResourceLine("MEMORY", stats.MemoryUsagePercent, stats.UsedMemoryBytes, stats.TotalMemoryBytes),
	}

	return f.panel.Render(lines)
}

func (f *ClusterFormatter) formatResourceLine(name string, percent float64, used, total int64) string {
	bar := f.progressBar.Render(percent)
	usedStr := util.FormatBytes(used)
	totalStr := util.FormatBytes(total)
	return fmt.Sprintf("%-8s %s  %5.1f%%    %8s / %-8s",
		name, bar, percent, usedStr, totalStr)
}
