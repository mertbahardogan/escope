package components

import (
	"fmt"
	"strings"

	"github.com/mertbahardogan/escope/internal/ui/styles"
)

type Table struct {
	border styles.Border
}

func NewTable() *Table {
	return &Table{
		border: styles.Default,
	}
}

func (t *Table) Render(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return "No data found\n"
	}

	widths := t.calculateWidths(headers, rows)
	rowFormat := t.buildRowFormat(widths)

	var output strings.Builder

	output.WriteString(t.buildBorder(widths))
	output.WriteString(t.formatRow(rowFormat, headers))
	output.WriteString(t.buildBorder(widths))

	for _, row := range rows {
		if len(row) > 0 && strings.ToUpper(row[0]) == "TOTAL" {
			output.WriteString(t.buildBorder(widths))
		}
		output.WriteString(t.formatRow(rowFormat, row))
	}

	output.WriteString(t.buildBorder(widths))

	return output.String()
}

func (t *Table) calculateWidths(headers []string, rows [][]string) []int {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, col := range row {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}
	return widths
}

func (t *Table) buildRowFormat(widths []int) string {
	var builder strings.Builder
	for _, w := range widths {
		builder.WriteString(t.border.Vertical + " %-")
		builder.WriteString(fmt.Sprintf("%d", w))
		builder.WriteString("s ")
	}
	builder.WriteString(t.border.Vertical + "\n")
	return builder.String()
}

func (t *Table) buildBorder(widths []int) string {
	var builder strings.Builder
	builder.WriteString(t.border.Corner)
	for _, w := range widths {
		builder.WriteString(strings.Repeat(t.border.Horizontal, w+2))
		builder.WriteString(t.border.Corner)
	}
	builder.WriteString("\n")
	return builder.String()
}

func (t *Table) formatRow(format string, values []string) string {
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = v
	}
	return fmt.Sprintf(format, args...)
}
