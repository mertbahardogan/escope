package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mertbahardogan/escope/internal/calculator"
	"github.com/mertbahardogan/escope/internal/calculatorsession"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/ui/components"
	"github.com/mertbahardogan/escope/internal/util"
)

var calcFieldNames = []string{
	"Data nodes",
	"Dedicated master nodes",
	"Primary shards",
	"Replicas per shard",
	"Total primary data (GiB)",
	"Documents",
	"Read throughput (rps)",
	"Write throughput (rps)",
	"RAM per data node (GiB)",
	"Disk per data node (GiB)",
}

type CalculatorHeader struct {
	HostAlias string
	Status    string
}

type CalculatorModel struct {
	fields   []string
	focus    int
	width    int
	height   int
	scroll   int
	saveHint string
	header   *CalculatorHeader
}

func defaultCalculatorFields() []string {
	return []string{
		"3",
		"0",
		"3",
		"2",
		"90",
		"10000000",
		"50",
		"50",
		"64",
		"2000",
	}
}

func NewCalculatorModel(seed *calculator.Inputs, fromSnapshot bool, header *CalculatorHeader) *CalculatorModel {
	fields := defaultCalculatorFields()
	focus, scroll := 0, 0

	if seed != nil {
		fields[0] = itoaOrEmpty(seed.DataNodes)
		fields[1] = itoaOrEmpty(seed.DedicatedMasters)
		fields[2] = itoaOrEmpty(seed.Shards)
		fields[3] = itoaOrEmpty(seed.ReplicasPerShard)
		fields[4] = itoaOrEmpty(seed.GBSize)
		fields[5] = formatInt64(seed.Documents)
		fields[6] = formatFloatTrim(seed.ReadRPS)
		fields[7] = formatFloatTrim(seed.WriteRPS)
		fields[8] = formatFloatTrim(seed.RAMGiBPerDataNode)
		fields[9] = formatFloatTrim(seed.DiskGiBPerDataNode)
	} else if fromSnapshot {
		st, ok := calculatorsession.ReadState()
		if !ok {
			return &CalculatorModel{fields: fields, focus: focus, scroll: scroll, width: 80, height: 24, header: header}
		}
		fields = append([]string(nil), st.Fields...)
		focus = st.Focus
		scroll = st.Scroll
	} else if st, ok := calculatorsession.ReadState(); ok {
		fields = append([]string(nil), st.Fields...)
		focus = st.Focus
		scroll = st.Scroll
	}

	return &CalculatorModel{fields: fields, focus: focus, scroll: scroll, width: 80, height: 24, header: header}
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
		key := msg.String()
		if key != "ctrl+s" {
			m.saveHint = ""
		}
		fullLineCount := m.lineCount()
		switch key {
		case "ctrl+s":
			err := calculatorsession.Write(&calculatorsession.State{
				Fields: append([]string(nil), m.fields...),
				Focus:  m.focus,
				Scroll: m.scroll,
			})
			if err != nil {
				m.saveHint = "Save failed: " + err.Error()
			} else {
				m.saveHint = constants.CalculatorMsgSaved
			}
			return m, nil
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

func renderCalculatorStatusBadge(status string) string {
	s := strings.TrimSpace(status)
	if s == "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("—")
	}
	label := strings.ToUpper(s)
	var bg, fg string
	switch strings.ToLower(s) {
	case "green":
		bg, fg = "42", "230"
	case "yellow":
		bg, fg = "220", "235"
	case "red":
		bg, fg = "196", "230"
	default:
		bg, fg = "240", "252"
	}
	return lipgloss.NewStyle().Bold(true).Background(lipgloss.Color(bg)).Foreground(lipgloss.Color(fg)).Padding(0, 1).Render(label)
}

func (m *CalculatorModel) renderFullContent() string {
	in, parseErr := parseCalculatorFields(m.fields)
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Render("Cluster Scale Calculator")
	if m.header != nil && m.header.HostAlias != "" {
		clusterPart := lipgloss.NewStyle().Bold(true).Render("Cluster: " + m.header.HostAlias)
		badge := renderCalculatorStatusBadge(m.header.Status)
		b.WriteString(title + "  " + clusterPart + "  " + badge)
	} else {
		b.WriteString(title)
	}
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		"tab/shift+tab · j/k fields · PgUp/PgDn scroll · home/end · ctrl+s save · q quit",
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
	statusLine := statusText
	if m.saveHint != "" {
		statusLine = m.saveHint + " | " + statusLine
	}
	status := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(statusLine)

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

	if in.DataNodes, err = parseInt(0); err != nil {
		return in, "invalid data nodes"
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
	if in.ReadRPS, err = parseF(6); err != nil {
		return in, "invalid read rps"
	}
	if in.WriteRPS, err = parseF(7); err != nil {
		return in, "invalid write rps"
	}
	if in.RAMGiBPerDataNode, err = parseF(8); err != nil {
		return in, "invalid RAM per data node"
	}
	if in.DiskGiBPerDataNode, err = parseF(9); err != nil {
		return in, "invalid disk per data node"
	}

	if in.RAMGiBPerDataNode < 1 {
		in.RAMGiBPerDataNode = 1
	}
	if in.DiskGiBPerDataNode < 1 {
		in.DiskGiBPerDataNode = 1
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
		{"Est. total stored (primaries + replicas)", util.FormatBytes(res.ClusterBytes)},
		{"Avg primary shard size", fmt.Sprintf("%.2f GiB", res.GBPerPrimaryShard)},
		{"Read load per piece (shard or replica)", fmt.Sprintf("%.2f rps", res.ReadPerPiece)},
		{"Write load per primary shard", fmt.Sprintf("%.2f rps", res.WritePerShard)},
		{"Shard size guidance", warnCell},
		{"Allocation viable", allocation},
	}
	out.WriteString(tbl.Render(summaryHeaders, summaryRows))

	out.WriteString("\n")
	out.WriteString(renderHealthLines(in, res))

	out.WriteString(renderRAMBreakdown(in, res))

	return out.String()
}

func renderHealthLines(in calculator.Inputs, res calculator.Result) string {
	errSt := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warnSt := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	okSt := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	var lines []string

	switch res.SizeWarning {
	case calculator.SizeWarningHighDanger:
		lines = append(lines, errSt.Render("Error: average primary shard size > 50 GiB — oversize-shard risk vs Elasticsearch guidance."))
	case calculator.SizeWarningLowDanger:
		lines = append(lines, errSt.Render("Error: average primary shard < 8 GiB with multiple primaries — many undersized shards (wasted overhead)."))
	case calculator.SizeWarningLowWarn:
		lines = append(lines, warnSt.Render("Warning: average primary shard < 13 GiB — many relatively small shards."))
	}

	if in.Shards > 0 && !res.HasExpectedNodes {
		lines = append(lines, warnSt.Render("Warning: replica placement or total node count does not satisfy this allocation model."))
	}
	for _, msg := range res.Messages {
		lines = append(lines, warnSt.Render("Warning: "+msg))
	}

	healthy := in.Shards > 0 && res.HasExpectedNodes && res.SizeWarning == calculator.SizeWarningNone && len(res.Messages) == 0
	if healthy {
		lines = append(lines, okSt.Render("Good! Data node count and primary shard size fall within typical Elasticsearch guidance for this model."))
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderRAMBreakdown(in calculator.Inputs, res calculator.Result) string {
	if in.Shards == 0 || res.DataNodes == 0 {
		return ""
	}
	nodeRows := calculator.NodeSummaries(in, res.Allocation)
	views := calculator.NodeResourceViews(in, nodeRows)
	if len(views) == 0 {
		return ""
	}
	jvmOrange := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	osTeal := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	detailOrange := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	subtle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	pctRed := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	hd := lipgloss.NewStyle().Bold(true).MarginTop(1).MarginBottom(1)

	const barW = 24

	v := views[0]
	var sumCover float64
	for _, x := range views {
		sumCover += x.PageCacheCoversHotPct
	}
	avgCover := sumCover / float64(len(views))

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(hd.Render(fmt.Sprintf("RAM breakdown (cluster · %d data nodes · %.1f GiB RAM per node)",
		len(views), v.RAMGiB)) + "\n")
	b.WriteString(subtle.Render("JVM heap cap = min(50% RAM, 31 GiB); remainder = OS page cache budget (same assumption for every data node)") + "\n")

	if v.RAMGiB > 0 {
		jvmBars := int(math.Round(v.HeapCapGiB / v.RAMGiB * float64(barW)))
		if jvmBars > barW {
			jvmBars = barW
		}
		osBars := barW - jvmBars
		b.WriteString(fmt.Sprintf("  %-22s %s  %6.1f GiB\n",
			"JVM heap", jvmOrange.Render(strings.Repeat("█", jvmBars)), v.HeapCapGiB))
		b.WriteString(fmt.Sprintf("  %-22s %s  %6.1f GiB\n",
			"OS page cache", osTeal.Render(strings.Repeat("█", osBars)), v.OSPageCacheGiB))
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).MarginTop(1).Render("JVM heap detail") + "\n")
	heapCap := math.Max(v.HeapCapGiB, 1e-9)
	heapRow := func(label string, g float64) {
		n := 0
		if g > 0 {
			n = int(math.Round(g / heapCap * float64(barW)))
		}
		if n > barW {
			n = barW
		}
		b.WriteString(fmt.Sprintf("  %-22s %s  %6.1f GiB\n",
			label, detailOrange.Render(strings.Repeat("█", n)), g))
	}
	heapRow("Field data cache", v.FieldDataGiB)
	heapRow("Query buffer", v.QueryBufferGiB)
	heapRow("Indexing buffer", v.IndexBufferGiB)
	heapRow("Available", v.HeapAvailGiB)

	b.WriteString("  Page cache covers hot data: ")
	b.WriteString(pctRed.Render(fmt.Sprintf("%.0f%%", avgCover)))
	b.WriteString(subtle.Render("  (mean across data nodes: OS cache budget vs data on node)") + "\n")
	return b.String()
}

func describeSizeWarning(sw calculator.SizeWarning) string {
	switch sw {
	case calculator.SizeWarningHighDanger:
		return ">50 GiB / primary — very large shard"
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
	views := calculator.NodeResourceViews(in, nodeRows)
	tbl := components.NewTable()
	h := lipgloss.NewStyle().Bold(true).MarginTop(1).MarginBottom(1)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(h.Render("Data nodes (model allocation)") + "\n")

	headers := []string{"Node", "Pri", "Rep", "Read rps", "Write rps", "Docs (pri)", "Data GiB", "Disk %", "Heap % RAM", "Cache fit*"}
	rows := make([][]string, len(nodeRows))
	for i := range nodeRows {
		r := nodeRows[i]
		v := views[i]
		rows[i] = []string{
			fmt.Sprintf("%d", r.NodeIndex),
			fmt.Sprintf("%d", r.Primaries),
			fmt.Sprintf("%d", r.Replicas),
			fmt.Sprintf("%.1f", r.ReadRPS),
			fmt.Sprintf("%.1f", r.WriteRPS),
			fmt.Sprintf("%.0f", r.Docs),
			fmt.Sprintf("%.1f", v.DataGiB),
			fmt.Sprintf("%.0f%%", v.DiskUsePct),
			fmt.Sprintf("%.0f%%", v.HeapOfRAMPct),
			fmt.Sprintf("%.0f%%", v.PageCacheCoversHotPct),
		}
	}
	b.WriteString(tbl.Render(headers, rows))
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		"*Cache fit: OS page-cache budget (RAM − JVM heap cap) vs indexed data on node (model).") + "\n")
	return b.String()
}

func RunCalculatorTUI(seed *calculator.Inputs, fromSnapshot bool, header *CalculatorHeader) error {
	if fromSnapshot {
		if _, ok := calculatorsession.ReadState(); !ok {
			return fmt.Errorf("%s", constants.CalculatorErrSnapshotMissing)
		}
	}
	p := tea.NewProgram(NewCalculatorModel(seed, fromSnapshot, header), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
