package components

import (
	"strings"
	"unicode/utf8"

	"github.com/mertbahardogan/escope/internal/ui/styles"
)

type Panel struct {
	border styles.Border
	title  string
}

func NewPanel(title string) *Panel {
	return &Panel{
		border: styles.Default,
		title:  title,
	}
}

func (p *Panel) Render(lines []string) string {
	width := p.calculateWidth(lines)

	var output strings.Builder

	output.WriteString(p.border.BuildLine(width))
	output.WriteString(p.border.BuildRow(p.title, width-2))
	output.WriteString(p.border.BuildLine(width))

	for _, line := range lines {
		output.WriteString(p.border.BuildRow(line, width-2))
	}

	output.WriteString(p.border.BuildLine(width))

	return output.String()
}

func (p *Panel) calculateWidth(lines []string) int {
	maxLen := utf8.RuneCountInString(p.title)
	for _, line := range lines {
		runeCount := utf8.RuneCountInString(line)
		if runeCount > maxLen {
			maxLen = runeCount
		}
	}
	return maxLen + 4
}
