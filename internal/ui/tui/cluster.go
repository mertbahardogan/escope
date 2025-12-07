package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui/components"
	"github.com/mertbahardogan/escope/internal/util"
)

type ClusterModel struct {
	stats       *models.ClusterStats
	service     services.ClusterService
	err         error
	loading     bool
	table       *components.Table
	panel       *components.Panel
	progressBar *components.ProgressBar
}

type statsMsg struct {
	stats *models.ClusterStats
	err   error
}

func NewClusterModel(service services.ClusterService) ClusterModel {
	return ClusterModel{
		service:     service,
		loading:     true,
		table:       components.NewTable(),
		panel:       components.NewPanel("RESOURCES"),
		progressBar: components.NewProgressBar(),
	}
}

func (m ClusterModel) Init() tea.Cmd {
	return m.fetchStats()
}

func (m ClusterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case statsMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.stats = msg.stats
			m.err = nil
		}
		return m, tea.Sequence(tea.Println(m.renderOutput()), tea.Quit)
	}

	return m, nil
}

func (m ClusterModel) View() string {
	if m.loading {
		return "Loading...\n"
	}
	return ""
}

func (m ClusterModel) renderOutput() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		return b.String()
	}

	if m.stats == nil {
		return b.String()
	}

	b.WriteString(m.renderResources())
	b.WriteString("\n")
	b.WriteString(m.renderMetrics())
	b.WriteString("\n")
	b.WriteString(m.renderSystem())

	return b.String()
}

func (m ClusterModel) fetchStats() tea.Cmd {
	return func() tea.Msg {
		stats, err := util.ExecuteWithTimeout(func() (*models.ClusterStats, error) {
			return m.service.GetClusterStats(context.Background())
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return statsMsg{err: fmt.Errorf("failed to get cluster health: %s", constants.MsgTimeoutGeneric)}
			}
			return statsMsg{err: fmt.Errorf("failed to get cluster health: %v", err)}
		}
		return statsMsg{stats: stats}
	}
}

func (m ClusterModel) renderHeader() string {
	if m.stats == nil {
		return "Cluster: loading..."
	}

	var badgeStyle lipgloss.Style
	switch m.stats.Status {
	case "green":
		badgeStyle = lipgloss.NewStyle().Background(lipgloss.Color("42")).Foreground(lipgloss.Color("0"))
	case "yellow":
		badgeStyle = lipgloss.NewStyle().Background(lipgloss.Color("226")).Foreground(lipgloss.Color("0"))
	case "red":
		badgeStyle = lipgloss.NewStyle().Background(lipgloss.Color("196")).Foreground(lipgloss.Color("15"))
	default:
		badgeStyle = lipgloss.NewStyle()
	}

	nameStyle := lipgloss.NewStyle().Bold(true)
	name := nameStyle.Render(m.stats.ClusterName)
	badge := badgeStyle.Render(" " + strings.ToUpper(m.stats.Status) + " ")
	return fmt.Sprintf("Cluster: %s  %s", name, badge)
}

func (m ClusterModel) renderResources() string {
	if m.stats == nil {
		return ""
	}

	usedDisk := m.stats.TotalDiskBytes - m.stats.AvailableDiskBytes

	lines := []string{
		m.renderResourceLine("DISK", m.stats.DiskUsagePercent, usedDisk, m.stats.TotalDiskBytes),
		m.renderResourceLine("HEAP", m.stats.HeapUsagePercent, m.stats.UsedHeapBytes, m.stats.TotalHeapBytes),
		m.renderResourceLine("MEMORY", m.stats.MemoryUsagePercent, m.stats.UsedMemoryBytes, m.stats.TotalMemoryBytes),
	}

	return m.panel.Render(lines)
}

func (m ClusterModel) renderResourceLine(name string, percent float64, used, total int64) string {
	bar := m.progressBar.Render(percent)
	usedStr := util.FormatBytes(used)
	totalStr := util.FormatBytes(total)
	return fmt.Sprintf("%-8s %s  %5.1f%%    %8s / %-8s", name, bar, percent, usedStr, totalStr)
}

func (m ClusterModel) renderMetrics() string {
	if m.stats == nil {
		return ""
	}

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Nodes", fmt.Sprintf("%d (%s)", m.stats.TotalNodes, m.stats.GetNodeBreakdown())},
		{"Indices", fmt.Sprintf("%d", m.stats.TotalIndices)},
		{"Documents", util.FormatDocsCount(m.stats.TotalDocuments)},
		{"Primary Shards", fmt.Sprintf("%d", m.stats.PrimaryShards)},
		{"Total Shards", fmt.Sprintf("%d", m.stats.TotalShards)},
		{"Avg Shard Size", fmt.Sprintf("%.2f GB", m.stats.AvgShardSizeGB)},
	}

	return m.table.Render(headers, rows)
}

func (m ClusterModel) renderSystem() string {
	if m.stats == nil {
		return ""
	}

	jvmVersions := "N/A"
	if len(m.stats.JVMVersions) > 0 {
		jvmVersions = strings.Join(m.stats.JVMVersions, ", ")
	}

	headers := []string{"System Info", "Value"}
	rows := [][]string{
		{"ES Version", m.stats.ESVersion},
		{"JVM Versions", jvmVersions},
	}

	return m.table.Render(headers, rows)
}

func RunClusterTUI(service services.ClusterService) error {
	p := tea.NewProgram(NewClusterModel(service))
	_, err := p.Run()
	return err
}
