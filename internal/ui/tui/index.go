package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui/styles"
	"github.com/mertbahardogan/escope/internal/util"
)

type dataPoint struct {
	value     float64
	timestamp time.Time
}

type IndexModel struct {
	indexName        string
	service          services.IndexService
	stats            *models.IndexDetailInfo
	err              error
	loading          bool
	searchHistory    []dataPoint
	indexHistory     []dataPoint
	queryTimeHistory []dataPoint
	indexTimeHistory []dataPoint
	startTime        time.Time
	border           styles.Border
	blueStyle        lipgloss.Style
	blueBoldStyle    lipgloss.Style
}

type indexTickMsg time.Time
type indexStatsMsg struct {
	stats *models.IndexDetailInfo
	err   error
}

func NewIndexModel(service services.IndexService, indexName string) IndexModel {
	return IndexModel{
		indexName:        indexName,
		service:          service,
		loading:          true,
		searchHistory:    make([]dataPoint, 0),
		indexHistory:     make([]dataPoint, 0),
		queryTimeHistory: make([]dataPoint, 0),
		indexTimeHistory: make([]dataPoint, 0),
		startTime:        time.Now(),
		border:           styles.Default,
		blueStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		blueBoldStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true),
	}
}

func (m IndexModel) Init() tea.Cmd {
	return tea.Batch(m.fetchStats(), m.tickCmd())
}

func (m IndexModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case indexTickMsg:
		return m, tea.Batch(m.fetchStats(), m.tickCmd())

	case indexStatsMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.stats = msg.stats
			m.err = nil
			m.addDataPoint(msg.stats)
		}
		return m, nil
	}

	return m, nil
}

func (m IndexModel) View() string {
	if m.loading && len(m.searchHistory) == 0 {
		return "Loading...\n"
	}
	return m.renderOutput()
}

func (m *IndexModel) addDataPoint(stats *models.IndexDetailInfo) {
	now := time.Now()

	searchRate := parseRate(stats.SearchRate)
	indexRate := parseRate(stats.IndexRate)
	queryTime := parseRate(stats.AvgQueryTime)
	indexTime := parseRate(stats.AvgIndexTime)

	m.searchHistory = append(m.searchHistory, dataPoint{value: searchRate, timestamp: now})
	m.indexHistory = append(m.indexHistory, dataPoint{value: indexRate, timestamp: now})
	m.queryTimeHistory = append(m.queryTimeHistory, dataPoint{value: queryTime, timestamp: now})
	m.indexTimeHistory = append(m.indexTimeHistory, dataPoint{value: indexTime, timestamp: now})

	maxPoints := 60
	if len(m.searchHistory) > maxPoints {
		m.searchHistory = m.searchHistory[len(m.searchHistory)-maxPoints:]
	}
	if len(m.indexHistory) > maxPoints {
		m.indexHistory = m.indexHistory[len(m.indexHistory)-maxPoints:]
	}
	if len(m.queryTimeHistory) > maxPoints {
		m.queryTimeHistory = m.queryTimeHistory[len(m.queryTimeHistory)-maxPoints:]
	}
	if len(m.indexTimeHistory) > maxPoints {
		m.indexTimeHistory = m.indexTimeHistory[len(m.indexTimeHistory)-maxPoints:]
	}
}

func parseRate(rate string) float64 {
	var value float64
	fmt.Sscanf(rate, "%f", &value)
	return value
}

func (m IndexModel) fetchStats() tea.Cmd {
	return func() tea.Msg {
		stats, err := util.ExecuteWithTimeout(func() (*models.IndexDetailInfo, error) {
			return m.service.GetIndexDetailInfo(context.Background(), m.indexName)
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return indexStatsMsg{err: fmt.Errorf("failed to get index stats: %s", constants.MsgTimeoutGeneric)}
			}
			return indexStatsMsg{err: fmt.Errorf("failed to get index stats: %v", err)}
		}
		return indexStatsMsg{stats: stats}
	}
}

func (m IndexModel) tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return indexTickMsg(t)
	})
}

func (m IndexModel) renderOutput() string {
	searchRateTitle := "Search Rate:"
	if m.stats != nil {
		searchRateTitle = fmt.Sprintf("Search Rate: %s", m.stats.SearchRate)
	}
	searchChart := m.renderChartCompact(searchRateTitle, m.searchHistory)

	indexRateTitle := "Index Rate:"
	if m.stats != nil {
		indexRateTitle = fmt.Sprintf("Index Rate: %s", m.stats.IndexRate)
	}
	indexChart := m.renderChartCompact(indexRateTitle, m.indexHistory)

	queryTimeTitle := "Query Time:"
	if m.stats != nil {
		queryTimeTitle = fmt.Sprintf("Query Time: %s", m.stats.AvgQueryTime)
	}
	queryChart := m.renderChartCompact(queryTimeTitle, m.queryTimeHistory)

	indexTimeTitle := "Index Time:"
	if m.stats != nil {
		indexTimeTitle = fmt.Sprintf("Index Time: %s", m.stats.AvgIndexTime)
	}
	indexTimeChart := m.renderChartCompact(indexTimeTitle, m.indexTimeHistory)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, searchChart, "    ", indexChart)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, queryChart, "    ", indexTimeChart)

	return "\n" + topRow + "\n" + bottomRow
}

func (m IndexModel) renderMetricsTable() string {
	if m.stats == nil {
		return ""
	}

	searchLabel := m.blueStyle.Render("Search:")
	indexLabel := m.blueStyle.Render("Index:")
	queryLabel := m.blueStyle.Render("Query:")
	idxTimeLabel := m.blueStyle.Render("Index:")

	return fmt.Sprintf("%s %s    %s %s    %s %s    %s %s\n",
		searchLabel, m.stats.SearchRate,
		indexLabel, m.stats.IndexRate,
		queryLabel, m.stats.AvgQueryTime,
		idxTimeLabel, m.stats.AvgIndexTime)
}

func (m IndexModel) renderChartCompact(title string, data []dataPoint) string {
	var b strings.Builder

	b.WriteString(m.blueBoldStyle.Render(title))
	b.WriteString("\n")

	if len(data) < 2 {
		b.WriteString(m.blueStyle.Render("waiting for data..."))
		b.WriteString("\n")
		return b.String()
	}

	chartHeight := 6
	chartWidth := 40

	minVal, maxVal := m.getMinMax(data)
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	padding := (maxVal - minVal) * 0.1
	minVal = minVal - padding
	maxVal = maxVal + padding

	chart := make([][]rune, chartHeight)
	for i := range chart {
		chart[i] = make([]rune, chartWidth)
		for j := range chart[i] {
			chart[i][j] = ' '
		}
	}

	step := 1
	if len(data) > chartWidth {
		step = len(data) / chartWidth
	}

	for i := 0; i < len(data)-1 && i/step < chartWidth-1; i += step {
		idx := i / step
		y1 := int((data[i].value - minVal) / (maxVal - minVal) * float64(chartHeight-1))
		nextIdx := i + step
		if nextIdx >= len(data) {
			nextIdx = len(data) - 1
		}
		y2 := int((data[nextIdx].value - minVal) / (maxVal - minVal) * float64(chartHeight-1))

		if y1 < 0 {
			y1 = 0
		}
		if y1 >= chartHeight {
			y1 = chartHeight - 1
		}
		if y2 < 0 {
			y2 = 0
		}
		if y2 >= chartHeight {
			y2 = chartHeight - 1
		}

		y1 = chartHeight - 1 - y1
		y2 = chartHeight - 1 - y2

		if y1 == y2 {
			chart[y1][idx] = '─'
		} else if y1 < y2 {
			chart[y1][idx] = '╮'
			for y := y1 + 1; y < y2; y++ {
				chart[y][idx] = '│'
			}
			chart[y2][idx] = '╰'
		} else {
			chart[y1][idx] = '╯'
			for y := y2 + 1; y < y1; y++ {
				chart[y][idx] = '│'
			}
			chart[y2][idx] = '╭'
		}
	}

	for i := 0; i < chartHeight; i++ {
		val := maxVal - (float64(i)/float64(chartHeight-1))*(maxVal-minVal)
		label := fmt.Sprintf("%4.0f", val)
		b.WriteString(label)
		b.WriteString("│")
		b.WriteString(m.blueStyle.Render(string(chart[i])))
		b.WriteString("\n")
	}

	b.WriteString("    └")
	b.WriteString(strings.Repeat("─", chartWidth))
	b.WriteString("\n")

	b.WriteString(m.renderTimeAxisCompact(data, chartWidth))

	return b.String()
}

func (m IndexModel) renderTimeAxisCompact(data []dataPoint, width int) string {
	if len(data) < 2 {
		return ""
	}

	startTime := data[0].timestamp
	endTime := data[len(data)-1].timestamp

	startStr := startTime.Format("15:04:05")
	endStr := endTime.Format("15:04:05")

	spacing := width - len(startStr) - len(endStr) + 5
	if spacing < 1 {
		spacing = 1
	}

	return fmt.Sprintf("     %s%s%s\n", startStr, strings.Repeat(" ", spacing), endStr)
}

func (m IndexModel) renderChart(title string, data []dataPoint) string {
	var b strings.Builder

	b.WriteString(m.blueBoldStyle.Render(title))
	b.WriteString("\n")

	if len(data) < 2 {
		b.WriteString(m.blueStyle.Render("waiting for data..."))
		b.WriteString("\n")
		return b.String()
	}

	chartHeight := 5
	chartWidth := 72

	minVal, maxVal := m.getMinMax(data)
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	padding := (maxVal - minVal) * 0.1
	minVal = minVal - padding
	maxVal = maxVal + padding

	chart := make([][]rune, chartHeight)
	for i := range chart {
		chart[i] = make([]rune, chartWidth)
		for j := range chart[i] {
			chart[i][j] = ' '
		}
	}

	for i := 0; i < len(data)-1 && i < chartWidth-1; i++ {
		y1 := int((data[i].value - minVal) / (maxVal - minVal) * float64(chartHeight-1))
		y2 := int((data[i+1].value - minVal) / (maxVal - minVal) * float64(chartHeight-1))

		if y1 < 0 {
			y1 = 0
		}
		if y1 >= chartHeight {
			y1 = chartHeight - 1
		}
		if y2 < 0 {
			y2 = 0
		}
		if y2 >= chartHeight {
			y2 = chartHeight - 1
		}

		y1 = chartHeight - 1 - y1
		y2 = chartHeight - 1 - y2

		if y1 == y2 {
			chart[y1][i] = '─'
		} else if y1 < y2 {
			chart[y1][i] = '╮'
			for y := y1 + 1; y < y2; y++ {
				chart[y][i] = '│'
			}
			chart[y2][i] = '╰'
		} else {
			chart[y1][i] = '╯'
			for y := y2 + 1; y < y1; y++ {
				chart[y][i] = '│'
			}
			chart[y2][i] = '╭'
		}
	}

	for i := 0; i < chartHeight; i++ {
		val := maxVal - (float64(i)/float64(chartHeight-1))*(maxVal-minVal)
		label := fmt.Sprintf("%4.0f", val)
		b.WriteString(label)
		b.WriteString("│")
		b.WriteString(m.blueStyle.Render(string(chart[i])))
		b.WriteString("\n")
	}

	b.WriteString("    └")
	b.WriteString(strings.Repeat("─", chartWidth))
	b.WriteString("\n")

	b.WriteString(m.renderTimeAxis(data, chartWidth))

	return b.String()
}

func (m IndexModel) renderTimeAxis(data []dataPoint, width int) string {
	if len(data) < 2 {
		return ""
	}

	startTime := data[0].timestamp
	endTime := data[len(data)-1].timestamp

	startStr := startTime.Format("15:04:05")
	endStr := endTime.Format("15:04:05")

	midTime := startTime.Add(endTime.Sub(startTime) / 2)
	midStr := midTime.Format("15:04:05")

	spacing := width - len(startStr) - len(midStr) - len(endStr)
	leftSpace := spacing / 2
	rightSpace := spacing - leftSpace

	return fmt.Sprintf("     %s%s%s%s%s\n",
		startStr,
		strings.Repeat(" ", leftSpace),
		midStr,
		strings.Repeat(" ", rightSpace),
		endStr)
}

func (m IndexModel) getMinMax(data []dataPoint) (float64, float64) {
	if len(data) == 0 {
		return 0, 100
	}

	minVal := data[0].value
	maxVal := data[0].value

	for _, d := range data {
		if d.value < minVal {
			minVal = d.value
		}
		if d.value > maxVal {
			maxVal = d.value
		}
	}

	return minVal, maxVal
}

func RunIndexTUI(service services.IndexService, indexName string) error {
	p := tea.NewProgram(NewIndexModel(service, indexName))
	_, err := p.Run()
	return err
}
