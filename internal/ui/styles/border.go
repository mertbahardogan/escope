package styles

import (
	"strings"
	"unicode/utf8"
)

type Border struct {
	Corner     string
	Horizontal string
	Vertical   string
}

var Default = Border{
	Corner:     "+",
	Horizontal: "-",
	Vertical:   "|",
}

func (b Border) BuildLine(width int) string {
	return b.Corner + strings.Repeat(b.Horizontal, width) + b.Corner + "\n"
}

func (b Border) BuildRow(content string, width int) string {
	runeCount := utf8.RuneCountInString(content)
	padding := width - runeCount
	if padding < 0 {
		padding = 0
	}
	return b.Vertical + " " + content + strings.Repeat(" ", padding) + " " + b.Vertical + "\n"
}

type Progress struct {
	Filled string
	Empty  string
	Width  int
}

var DefaultProgress = Progress{
	Filled: "\u2588",
	Empty:  "\u2591",
	Width:  20,
}

func (p Progress) Render(percent float64) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := int(percent / 100.0 * float64(p.Width))
	empty := p.Width - filled
	return strings.Repeat(p.Filled, filled) + strings.Repeat(p.Empty, empty)
}
