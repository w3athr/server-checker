package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	warnStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500"))
	critStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000"))
	okStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
)

func renderProgressBar(percent float64, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(float64(width) * (percent / 100))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	color := "#00FF00" // Green
	if percent > 85 {
		color = "#FF0000" // Red
	} else if percent > 70 {
		color = "#FFFF00" // Yellow
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(bar)
}

func formatBytes(bytes uint64) string {
	const (
		kb uint64 = 1024
		mb uint64 = kb * 1024
		gb uint64 = mb * 1024
	)

	if bytes < gb {
		return fmt.Sprintf("%.2fMB", float64(bytes)/float64(mb))
	}

	return fmt.Sprintf("%.2fGB", float64(bytes)/float64(gb))
}

func gbToBytesDec(gb int) uint64 {
	return uint64(gb) * 1000 * 1000 * 1000
}

func bytesToGBDec(b uint64) float64 {
	return float64(b) / 1000 / 1000 / 1000
}
