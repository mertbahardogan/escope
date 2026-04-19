package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mertbahardogan/escope/internal/calculator"
	"github.com/mertbahardogan/escope/internal/ui/components"
	"github.com/mertbahardogan/escope/internal/util"
)

var calcFieldNames = []string{
	"Total nodes",
	"Dedicated master nodes",
	"Primary shards",
	"Replicas per shard",
	"Total primary data (GiB)",
	"Documents",
	"Read throughput (rpm)",
	"Write throughput (rpm)",
	"Cluster scenarios (copies)",
}

// CalculatorModel is an interactive sizing calculator (elastic-calculator port).
// Pointer receiver: bubbletea holds one model instance; value receivers made slice updates easy to get wrong.
type CalculatorModel struct {
	fields []string
	focus  int
	width  int
	height int
	scroll int // first visible line of body (fields + tables); PgUp/PgDn
}

func defaultCalculatorFields() []string {
	return []string{
		"3",
		"0",
		"3",
		"2",
		"90",
		"10000000",
		"3000",
		"500",
		"1",
	}
}

// NewCalculatorModel builds the TUI; optional seed overrides defaults (e.g. from cluster stats).
func NewCalculatorModel(seed *calculator.Inputs) *CalculatorModel {
	fields := defaultCalculatorFields()
	if seed != nil {
		fields[0] = itoaOrEmpty(seed.Nodes)
		fields[1] = itoaOrEmpty(seed.DedicatedMasters)
		fields[2] = itoaOrEmpty(seed.Shards)
		fields[3] = itoaOrEmpty(seed.ReplicasPerShard)
		fields[4] = itoaOrEmpty(seed.GBSize)
		fields[5] = formatInt64(seed.Documents)
		fields[6] = formatFloatTrim(seed.ReadRPM)
		fields[7] = formatFloatTrim(seed.WriteRPM)
		fields[8] = itoaOrEmpty(seed.Clusters)
	}
	return &CalculatorModel{fields: fields, focus: 0, width: 80, height: 24}
}

func itoaOrEmpty(v int) string {
	return strconv.Itoa(v)
}

func formatInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}

func formatFloatTrim(v float64) string {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	return s
}

func (m *CalculatorModel) Init() tea.Cmd {
	return nil
}

func (m *CalculatorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		if msg.Height > 0 {
			m.height = msg.Height
		}
		m.clampScroll(m.lineCount())
		return m, nil

	case tea.KeyMsg:
		fullLineCount := m.lineCount()
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "pgdown", "ctrl+d":
			m.scrollBy(8, fullLineCount)
			return m, nil
		case "pgup", "ctrl+u":
			m.scrollBy(-8, fullLineCount)
			return m, nil
		case "home", "ctrl+g":
			m.scroll = 0
			return m, nil
		case "end":
			m.scroll = m.maxScroll(fullLineCount)
			return m, nil
		case "tab":
			m.focus = (m.focus + 1) % len(m.fields)
			return m, nil
		case "shift+tab":
			m.focus = (m.focus - 1 + len(m.fields)) % len(m.fields)
			return m, nil
		case "up", "k":
			m.focus = (m.focus - 1 + len(m.fields)) % len(m.fields)
			return m, nil
		case "down", "j":
			m.focus = (m.focus + 1) % len(m.fields)
			return m, nil
		case "backspace":
			if len(m.fields[m.focus]) > 0 {
				m.fields[m.focus] = m.fields[m.focus][:len(m.fields[m.focus])-1]
			}
			m.clampScroll(m.lineCount())
			return m, nil
		default:
			if len(msg.Runes) == 0 {
				return m, nil
			}
			for _, r := range msg.Runes {
				if r >= '0' && r <= '9' || r == '.' {
					m.fields[m.focus] += string(r)
				}
			}
			m.clampScroll(m.lineCount())
			return m, nil
		}
	}
	return m, nil
}

func (m *CalculatorModel) maxScroll(lineCount int) int {
	viewport := m.viewportLines()
	if lineCount <= viewport || viewport < 1 {
		return 0
	}
	if lineCount-viewport < 0 {
		return 0
	}
	return lineCount - viewport
}

func (m *CalculatorModel) viewportLines() int {
	// One terminal row is reserved for the status line below the viewport.
	h := m.height
	if h <= 0 {
		h = 24
	}
	v := h - 1
	if v < 1 {
		v = 1
	}
	return v
}

func (m *CalculatorModel) scrollBy(delta, lineCount int) {
	m.scroll += delta
	if m.scroll < 0 {
		m.scroll = 0
	}
	ms := m.maxScroll(lineCount)
	if m.scroll > ms {
		m.scroll = ms
	}
}

func (m *CalculatorModel) clampScroll(lineCount int) {
	ms := m.maxScroll(lineCount)
	if m.scroll > ms {
		m.scroll = ms
	}
	if m.scroll < 0 {
		m.scroll = 0
	}
}

func (m *CalculatorModel) lineCount() int {
	return len(splitContentLines(m.renderFullContent()))
}

func splitContentLines(s string) []string {
	s = strings.TrimRight(s, "\n\r")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func (m *CalculatorModel) renderFullContent() string {
	in, parseErr := parseCalculatorFields(m.fields)
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Cluster Calculator"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		"tab/shift+tab · j/k fields · PgUp/PgDn scroll · home/end · q quit",
	))
	b.WriteString("\n\n")

	focusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	normal := lipgloss.NewStyle()

	maxName := 0
	for _, n := range calcFieldNames {
		if len(n) > maxName {
			maxName = len(n)
		}
	}

	for i := range m.fields {
		name := calcFieldNames[i]
		padding := strings.Repeat(" ", maxName-len(name))
		line := fmt.Sprintf("%s:%s %s", name, padding, m.fields[i])
		if i == m.focus {
			b.WriteString(focusStyle.Render("› "+line) + "\n")
		} else {
			b.WriteString(normal.Render("  "+line) + "\n")
		}
	}

	b.WriteString("\n")
	if parseErr != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(parseErr) + "\n\n")
	}

	res := calculator.Compute(in)
	b.WriteString(renderCalculatorSummary(res, in))

	nodeBlock := renderNodeSummaries(in, res)
	if nodeBlock != "" {
		b.WriteString("\n")
		b.WriteString(nodeBlock)
	}

	return b.String()
}

func (m *CalculatorModel) View() string {
	lines := splitContentLines(m.renderFullContent())
	n := len(lines)
	vp := m.viewportLines()

	maxScr := m.maxScroll(n)
	scroll := m.scroll
	if scroll > maxScr {
		scroll = maxScr
	}
	if scroll < 0 {
		scroll = 0
	}
	end := scroll + vp
	if end > n {
		end = n
	}

	var body string
	if scroll < n && end > scroll {
		body = strings.Join(lines[scroll:end], "\n")
	}

	startLine, lastLine := 0, 0
	if n > 0 && end > scroll {
		startLine = scroll + 1
		lastLine = end
	}

	statusText := fmt.Sprintf("lines %d-%d of %d — PgUp/PgDn · home/end", startLine, lastLine, n)
	switch {
	case n == 0:
		statusText = "empty"
	case n <= vp:
		statusText = fmt.Sprintf("all %d lines visible — PgUp/PgDn when output grows", n)
	}
	status := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(statusText)

	if n == 0 {
		return status
	}
	if body == "" {
		body = " "
	}
	return body + "\n" + status
}

func parseCalculatorFields(s []string) (calculator.Inputs, string) {
	var in calculator.Inputs
	var err error

	parseInt := func(i int) (int, error) {
		v := strings.TrimSpace(s[i])
		if v == "" {
			return 0, nil
		}
		return strconv.Atoi(v)
	}
	parseI64 := func(i int) (int64, error) {
		v := strings.TrimSpace(s[i])
		if v == "" {
			return 0, nil
		}
		return strconv.ParseInt(v, 10, 64)
	}
	parseF := func(i int) (float64, error) {
		v := strings.TrimSpace(s[i])
		if v == "" {
			return 0, nil
		}
		return strconv.ParseFloat(v, 64)
	}

	if in.Nodes, err = parseInt(0); err != nil {
		return in, "invalid total nodes"
	}
	if in.DedicatedMasters, err = parseInt(1); err != nil {
		return in, "invalid dedicated masters"
	}
	if in.Shards, err = parseInt(2); err != nil {
		return in, "invalid primary shards"
	}
	if in.ReplicasPerShard, err = parseInt(3); err != nil {
		return in, "invalid replicas per shard"
	}
	if in.GBSize, err = parseInt(4); err != nil {
		return in, "invalid GiB size"
	}
	if in.Documents, err = parseI64(5); err != nil {
		return in, "invalid documents"
	}
	if in.ReadRPM, err = parseF(6); err != nil {
		return in, "invalid read rpm"
	}
	if in.WriteRPM, err = parseF(7); err != nil {
		return in, "invalid write rpm"
	}
	if in.Clusters, err = parseInt(8); err != nil {
		return in, "invalid cluster scenarios"
	}

	if in.Clusters < 1 {
		in.Clusters = 1
	}
	return in, ""
}

func renderCalculatorSummary(res calculator.Result, in calculator.Inputs) string {
	tbl := components.NewTable()
	h := lipgloss.NewStyle().Bold(true).MarginBottom(1)

	var out strings.Builder
	out.WriteString(h.Render("Summary") + "\n")

	warn := describeSizeWarning(res.SizeWarning)
	warnCell := warn
	if warnCell == "" {
		warnCell = "—"
	}

	allocation := "no"
	if res.HasExpectedNodes {
		allocation = "yes"
	}

	summaryHeaders := []string{"Metric", "Value"}
	summaryRows := [][]string{
		{"Est. total size (incl. replicas × scenarios)", util.FormatBytes(res.ClusterBytes)},
		{"Avg primary shard size", fmt.Sprintf("%.2f GiB", res.GBPerPrimaryShard)},
		{"Read load per piece (shard or replica)", fmt.Sprintf("%.2f rpm", res.ReadPerPiece)},
		{"Write load per primary shard", fmt.Sprintf("%.2f rpm", res.WritePerShard)},
		{"Shard size guidance", warnCell},
		{"Allocation viable", allocation},
	}
	out.WriteString(tbl.Render(summaryHeaders, summaryRows))

	if !res.HasExpectedNodes {
		out.WriteString(
			lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("Warning: allocation or node count does not match the model.") + "\n",
		)
	}

	if len(res.Messages) > 0 {
		out.WriteString("\n")
		out.WriteString(lipgloss.NewStyle().Bold(true).MarginTop(1).MarginBottom(1).Render("Notices") + "\n")
		msgRows := make([][]string, len(res.Messages))
		for i, msg := range res.Messages {
			msgRows[i] = []string{msg}
		}
		out.WriteString(tbl.Render([]string{"Message"}, msgRows))
	}

	shards := calculator.ShardSummaries(in)
	if len(shards) > 32 {
		out.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			"(Per-shard table omitted: more than 32 primaries.)") + "\n")
	} else if len(shards) > 0 {
		out.WriteString("\n")
		out.WriteString(lipgloss.NewStyle().Bold(true).MarginTop(1).MarginBottom(1).Render("Per primary shard (rpm uses ceil)") + "\n")
		shardHeaders := []string{"#", "Read rpm", "Write rpm", "Docs", "Size"}
		shardRows := make([][]string, len(shards))
		for i, row := range shards {
			shardRows[i] = []string{
				fmt.Sprintf("%d", row.Index),
				fmt.Sprintf("%.0f", row.ReadRPM),
				fmt.Sprintf("%.0f", row.WriteRPM),
				fmt.Sprintf("%.0f", row.Docs),
				util.FormatBytes(row.Bytes),
			}
		}
		out.WriteString(tbl.Render(shardHeaders, shardRows))
	}

	return out.String()
}

func describeSizeWarning(sw calculator.SizeWarning) string {
	switch sw {
	case calculator.SizeWarningHighDanger:
		return ">32 GiB / primary — very large shard"
	case calculator.SizeWarningHighWarn:
		return ">28 GiB / primary — consider smaller shards"
	case calculator.SizeWarningLowDanger:
		return "<8 GiB / primary — may be wasteful (many small shards)"
	case calculator.SizeWarningLowWarn:
		return "<13 GiB / primary — many small shards"
	default:
		return ""
	}
}

func renderNodeSummaries(in calculator.Inputs, res calculator.Result) string {
	if in.Shards == 0 || res.DataNodes == 0 {
		return ""
	}
	nodeRows := calculator.NodeSummaries(in, res.Allocation)
	if len(nodeRows) == 0 {
		return ""
	}
	tbl := components.NewTable()
	h := lipgloss.NewStyle().Bold(true).MarginTop(1).MarginBottom(1)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(h.Render("Data nodes (model allocation)") + "\n")

	headers := []string{"Node", "Pri", "Rep", "Read rpm", "Write rpm", "Docs (pri)", "Bytes"}
	rows := make([][]string, len(nodeRows))
	for i, r := range nodeRows {
		rows[i] = []string{
			fmt.Sprintf("%d", r.NodeIndex),
			fmt.Sprintf("%d", r.Primaries),
			fmt.Sprintf("%d", r.Replicas),
			fmt.Sprintf("%.1f", r.ReadRPM),
			fmt.Sprintf("%.1f", r.WriteRPM),
			fmt.Sprintf("%.0f", r.Docs),
			util.FormatBytes(r.BytesAll),
		}
	}
	b.WriteString(tbl.Render(headers, rows))
	return b.String()
}

// RunCalculatorTUI launches the interactive calculator; seed may be nil for defaults.
func RunCalculatorTUI(seed *calculator.Inputs) error {
	p := tea.NewProgram(NewCalculatorModel(seed), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
