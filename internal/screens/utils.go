package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	warnStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500"))
	critStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000"))
	okStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))

	// Константы размеров
	gb = float64(1024 * 1024 * 1024)
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

func formatBytes(bytes uint64) float64 {
	return float64(bytes) / gb
}
