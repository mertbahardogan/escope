package ui

import (
	"fmt"
	"github.com/mertbahardogan/escope/internal/constants"
	"strings"
)

type GenericTableFormatter struct{}

func NewGenericTableFormatter() *GenericTableFormatter {
	return &GenericTableFormatter{}
}

type ReportSection struct {
	Title string
	Items []string
}

func (f *GenericTableFormatter) FormatTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return "No data found\n"
	}

	var output strings.Builder

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

	buildBorder := func() string {
		var b strings.Builder
		b.WriteString("+")
		for _, w := range widths {
			b.WriteString(strings.Repeat(constants.DashString, w+2))
			b.WriteString("+")
		}
		b.WriteString("\n")
		return b.String()
	}

	var fmtBuilder strings.Builder
	for i, w := range widths {
		_ = i
		fmtBuilder.WriteString("| %-")
		fmtBuilder.WriteString(fmt.Sprintf("%d", w))
		fmtBuilder.WriteString("s ")
	}
	fmtBuilder.WriteString("|\n")
	rowFmt := fmtBuilder.String()

	output.WriteString(buildBorder())

	headerVals := make([]interface{}, len(headers))
	for i := range headers {
		headerVals[i] = headers[i]
	}
	output.WriteString(fmt.Sprintf(rowFmt, headerVals...))
	output.WriteString(buildBorder())

	for _, row := range rows {
		if len(row) > 0 && strings.ToUpper(row[0]) == "TOTAL" {
			output.WriteString(buildBorder())
		}

		vals := make([]interface{}, len(row))
		for i := range row {
			vals[i] = row[i]
		}
		output.WriteString(fmt.Sprintf(rowFmt, vals...))
	}
	output.WriteString(buildBorder())

	return output.String()
}

func (f *GenericTableFormatter) FormatReport(title string, sections []ReportSection) string {
	var output strings.Builder

	var allLines []string
	allLines = append(allLines, title)

	for _, section := range sections {
		allLines = append(allLines, section.Title)
		for _, item := range section.Items {
			allLines = append(allLines, "• "+item)
		}
		if len(section.Items) > 0 {
			allLines = append(allLines, "")
		}
	}

	maxWidth := 0
	for _, line := range allLines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	if maxWidth < 80 {
		maxWidth = 80
	}

	buildBorder := func() string {
		return "+" + strings.Repeat(constants.DashString, maxWidth+2) + "+\n"
	}

	output.WriteString(buildBorder())

	titlePadding := (maxWidth - len(title)) / 2
	if titlePadding < 0 {
		titlePadding = 0
	}
	output.WriteString(fmt.Sprintf("| %*s%s%*s |\n", titlePadding, "", title, maxWidth-len(title)-titlePadding, ""))
	output.WriteString(buildBorder())

	for i, section := range sections {
		if len(section.Items) > 0 {
			if i > 0 {
				output.WriteString(buildBorder())
			}
			output.WriteString(fmt.Sprintf("| %-*s |\n", maxWidth, section.Title))
			output.WriteString(fmt.Sprintf("| %-*s |\n", maxWidth, ""))
			for _, item := range section.Items {
				output.WriteString(fmt.Sprintf("| %-*s |\n", maxWidth, "• "+item))
			}
		}
	}
	output.WriteString(buildBorder())

	return output.String()
}
