package screens

import (
	"fmt"
	"server-checker/internal/models"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type disksScreen struct {
	data models.SystemInfo
}

func NewDisksScreen() disksScreen {
	return disksScreen{}
}

func (d disksScreen) Init() tea.Cmd {
	return nil
}

func (d disksScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		d.data = models.SystemInfo(msg)
	}
	return d, nil
}

func (d disksScreen) View() string {
	if len(d.data.Disks) == 0 {
		return "Loading Disks data...\n\nPress ESC to return to menu."
	}

	var s strings.Builder
	s.WriteString("STORAGE INFORMATION" + "\n\n")

	for _, disk := range d.data.Disks {
		// Заголовок: Путь и Точка монтирования
		header := fmt.Sprintf("ID: %s  |  Mount: %s", disk.Device, disk.Mountpoint)
		s.WriteString(header + "\n")

		// Информация о железе
		modelInfo := fmt.Sprintf("  Model: %s [%s]", disk.Model, disk.Type)
		s.WriteString(modelInfo + "\n")

		// Прогресс-бар использования места
		usageLine := fmt.Sprintf("  Usage: %f / %f (%.1f%%)",
			formatBytes(disk.Used),
			formatBytes(disk.Total),
			disk.UsedPercent)
		s.WriteString(usageLine + "\n")
		s.WriteString("  " + renderProgressBar(disk.UsedPercent, 40) + "\n")

		// Износ (Wear) - Самое важное для Олега
		if disk.Wear >= 0 {
			wearText := fmt.Sprintf("%.0f%%", disk.Wear)
			wearStyleLocal := okStyle
			if disk.Wear > 80 {
				wearStyleLocal = critStyle
			} else if disk.Wear > 50 {
				wearStyleLocal = warnStyle
			}
			s.WriteString(fmt.Sprintf("  Wear Level: %s (Status: %s)\n",
				wearStyleLocal.Render(wearText),
				disk.Status))
		} else {
			s.WriteString("  Wear Level: N/A (Run as sudo or unsupported)" + "\n")
		}

		s.WriteString("\n")
	}

	s.WriteString("Press ESC to return to menu.")
	return s.String()
}
