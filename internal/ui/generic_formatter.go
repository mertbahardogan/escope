package ui

import (
	"fmt"
	"strings"

	"github.com/mertbahardogan/escope/internal/ui/components"
	"github.com/mertbahardogan/escope/internal/ui/styles"
)

type GenericTableFormatter struct {
	table  *components.Table
	border styles.Border
}

func NewGenericTableFormatter() *GenericTableFormatter {
	return &GenericTableFormatter{
		table:  components.NewTable(),
		border: styles.Default,
	}
}

type ReportSection struct {
	Title string
	Items []string
}

func (f *GenericTableFormatter) FormatTable(headers []string, rows [][]string) string {
	return f.table.Render(headers, rows)
}

func (f *GenericTableFormatter) FormatReport(title string, sections []ReportSection) string {
	var output strings.Builder

	var allLines []string
	allLines = append(allLines, title)

	for _, section := range sections {
		allLines = append(allLines, section.Title)
		for _, item := range section.Items {
			allLines = append(allLines, item)
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

	output.WriteString(f.border.BuildLine(maxWidth + 2))

	titlePadding := (maxWidth - len(title)) / 2
	if titlePadding < 0 {
		titlePadding = 0
	}
	paddedTitle := fmt.Sprintf("%*s%s%*s", titlePadding, "", title, maxWidth-len(title)-titlePadding, "")
	output.WriteString(f.border.BuildRow(paddedTitle, maxWidth))
	output.WriteString(f.border.BuildLine(maxWidth + 2))

	for i, section := range sections {
		if len(section.Items) > 0 {
			if i > 0 {
				output.WriteString(f.border.BuildLine(maxWidth + 2))
			}
			output.WriteString(f.border.BuildRow(section.Title, maxWidth))
			output.WriteString(f.border.BuildRow("", maxWidth))
			for _, item := range section.Items {
				output.WriteString(f.border.BuildRow(item, maxWidth))
			}
		}
	}
	output.WriteString(f.border.BuildLine(maxWidth + 2))

	return output.String()
}
