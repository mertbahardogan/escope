package components

import (
	"github.com/mertbahardogan/escope/internal/ui/styles"
)

type ProgressBar struct {
	progress styles.Progress
}

func NewProgressBar() *ProgressBar {
	return &ProgressBar{
		progress: styles.DefaultProgress,
	}
}

func NewProgressBarWithWidth(width int) *ProgressBar {
	p := styles.DefaultProgress
	p.Width = width
	return &ProgressBar{
		progress: p,
	}
}

func (p *ProgressBar) Render(percent float64) string {
	return p.progress.Render(percent)
}

func (p *ProgressBar) Width() int {
	return p.progress.Width
}
